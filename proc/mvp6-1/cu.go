package mvp6_1

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/common/obs"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	pendingLength = 10
)

type controlUnit struct {
	inBus                        *comp.BufferedBus[risc.InstructionRunnerPc]
	outBuses                     []*comp.BufferedBus[*risc.InstructionRunnerPc]
	pendings                     *comp.Queue[risc.InstructionRunnerPc]
	pushedRunnersInPreviousCycle map[*risc.InstructionRunnerPc]int
	pushedRunnersInCurrentCycle  map[*risc.InstructionRunnerPc]int

	pushed            *obs.Gauge
	pending           *obs.Gauge
	pendingRead       *obs.Gauge
	blocked           *obs.Gauge
	forwarding        int
	total             int
	cantAdd           int
	blockedBranch     int
	blockedDataHazard int
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBuses []*comp.BufferedBus[*risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:                        inBus,
		outBuses:                     outBuses,
		pendings:                     comp.NewQueue[risc.InstructionRunnerPc](pendingLength),
		pushed:                       &obs.Gauge{},
		pending:                      &obs.Gauge{},
		pendingRead:                  &obs.Gauge{},
		blocked:                      &obs.Gauge{},
		pushedRunnersInCurrentCycle:  make(map[*risc.InstructionRunnerPc]int),
		pushedRunnersInPreviousCycle: make(map[*risc.InstructionRunnerPc]int),
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	pushedCount := 0
	u.pushedRunnersInCurrentCycle = make(map[*risc.InstructionRunnerPc]int)
	defer func() {
		u.pushed.Push(pushedCount)
		u.pending.Push(u.pendings.Length())
		u.pushedRunnersInPreviousCycle = u.pushedRunnersInCurrentCycle
	}()
	u.pendingRead.Push(u.inBus.PendingRead())
	if u.inBus.CanGet() {
		u.blocked.Push(1)
	} else {
		u.blocked.Push(0)
	}
	u.total++

	if !u.outBuses[0].CanAdd() && !u.outBuses[1].CanAdd() {
		u.cantAdd++
		log.Infou(ctx, "CU", "can't add")
		return
	}

	//remaining := u.outBus.RemainingToAdd()
	for elem := range u.pendings.Iterator() {
		runner := u.pendings.Value(elem)

		push, stop := u.handleRunner(ctx, cycle, pushedCount, &runner)
		if push {
			u.pendings.Remove(elem)
			pushedCount++
		}
		if stop {
			return
		}
	}

	for !u.pendings.IsFull() {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}

		push, stop := u.handleRunner(ctx, cycle, pushedCount, &runner)
		if push {
			pushedCount++
		} else {
			u.pendings.Push(runner)
		}
		if stop {
			return
		}
	}
}

func (u *controlUnit) handleRunner(ctx *risc.Context, cycle int, pushedCount int, runner *risc.InstructionRunnerPc) (push, stop bool) {
	if pushedCount > 0 && runner.Runner.InstructionType().IsBranch() {
		u.blockedBranch++
		return false, true
	}

	hazard, reason := ctx.IsDataHazard(runner.Runner)
	if !hazard {
		pushed, id := u.pushRunner(ctx, cycle, runner)
		if !pushed {
			// TODO Return?
			return false, true
		}
		u.pushedRunnersInCurrentCycle[runner] = id
		return true, false
	} else {
		if should, previousRunner, id, register := u.shouldUseForwarding(runner); should {
			ch := make(chan int32, 1)
			previousRunner.Forwarder = ch
			previousRunner.ForwardRegister = register
			runner.Receiver = ch
			runner.ForwardRegister = register

			pushed := u.pushRunnerToBus(ctx, cycle, runner, id)
			if !pushed {
				// TODO Return?
				return false, true
			}
			u.pushedRunnersInCurrentCycle[runner] = id
			log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "forward runner on %s (source %d)", register, previousRunner.Pc/4)
			u.forwarding++
			// TODO Return?
			return true, true
		} else {
			log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%s", reason)
			u.blockedDataHazard++
			return false, true
		}
	}
}

func (u *controlUnit) shouldUseForwarding(runner *risc.InstructionRunnerPc) (bool, *risc.InstructionRunnerPc, int, risc.RegisterType) {
	// If there's a hazard with an instruction pushed in the same cycle
	for previousRunner := range u.pushedRunnersInCurrentCycle {
		for _, writeRegister := range previousRunner.Runner.WriteRegisters() {
			for _, readRegister := range runner.Runner.ReadRegisters() {
				if readRegister == writeRegister {
					return false, nil, 0, risc.Zero
				}
			}
		}
	}

	// Can we use forwarding with an instruction pushed in the previous cycle
	for previousRunner, id := range u.pushedRunnersInPreviousCycle {
		for _, writeRegister := range previousRunner.Runner.WriteRegisters() {
			for _, readRegister := range runner.Runner.ReadRegisters() {
				if readRegister == writeRegister {
					return true, previousRunner, id, readRegister
				}
			}
		}
	}
	return false, nil, 0, risc.Zero
}

func (u *controlUnit) pushRunner(ctx *risc.Context, cycle int, runner *risc.InstructionRunnerPc) (bool, int) {
	if u.pushRunnerToBus(ctx, cycle, runner, 0) {
		return true, 0
	}
	if u.pushRunnerToBus(ctx, cycle, runner, 1) {
		return true, 1
	}
	return false, 0
}

func (u *controlUnit) pushRunnerToBus(ctx *risc.Context, cycle int, runner *risc.InstructionRunnerPc, id int) bool {
	if !u.outBuses[id].CanAdd() {
		return false
	}
	u.outBuses[id].Add(runner, cycle)
	ctx.AddPendingRegisters(runner.Runner)
	log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner to %d", id)
	return true
}

func (u *controlUnit) flush() {
	u.pendings = comp.NewQueue[risc.InstructionRunnerPc](pendingLength)
	u.pushedRunnersInPreviousCycle = nil
}

func (u *controlUnit) isEmpty() bool {
	return u.pendings.Length() == 0
}
