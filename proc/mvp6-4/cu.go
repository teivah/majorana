package mvp6_3

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
	ctx                          *risc.Context
	inBus                        *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus                       *comp.BufferedBus[*risc.InstructionRunnerPc]
	pendings                     *comp.Queue[risc.InstructionRunnerPc]
	pushedRunnersInPreviousCycle map[*risc.InstructionRunnerPc]bool
	pushedRunnersInCurrentCycle  map[*risc.InstructionRunnerPc]bool
	skippedInCurrentCycle        []risc.InstructionRunnerPc
	pushedBranchInCurrentCycle   bool
	pendingConditionalBranch     bool

	// Monitoring
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

func newControlUnit(ctx *risc.Context, inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[*risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		ctx:                          ctx,
		inBus:                        inBus,
		outBus:                       outBus,
		pendings:                     comp.NewQueue[risc.InstructionRunnerPc](pendingLength),
		pushed:                       &obs.Gauge{},
		pending:                      &obs.Gauge{},
		pendingRead:                  &obs.Gauge{},
		blocked:                      &obs.Gauge{},
		pushedRunnersInCurrentCycle:  make(map[*risc.InstructionRunnerPc]bool),
		pushedRunnersInPreviousCycle: make(map[*risc.InstructionRunnerPc]bool),
	}
}

func (u *controlUnit) cycle(cycle int) {
	u.pushedRunnersInCurrentCycle = make(map[*risc.InstructionRunnerPc]bool)
	defer func() {
		u.pushed.Push(len(u.pushedRunnersInCurrentCycle))
		u.pending.Push(u.pendings.Length())
		u.pushedRunnersInPreviousCycle = u.pushedRunnersInCurrentCycle
	}()
	u.skippedInCurrentCycle = nil
	u.pushedBranchInCurrentCycle = false
	u.pendingRead.Push(u.inBus.PendingRead())
	if u.inBus.CanGet() {
		u.blocked.Push(1)
	} else {
		u.blocked.Push(0)
	}
	u.total++

	if !u.outBus.CanAdd() {
		u.cantAdd++
		log.Infou(u.ctx, "CU", "can't add")
		return
	}

	for elem := range u.pendings.Iterator() {
		runner := u.pendings.Value(elem)

		push, stop := u.handleRunner(u.ctx, cycle, &runner)
		if push {
			u.pushedRunnersInCurrentCycle[&runner] = true
			u.pendings.Remove(elem)
			if runner.Runner.InstructionType().IsBranch() {
				u.pushedBranchInCurrentCycle = true
			}
			if runner.Runner.InstructionType().IsConditionalBranch() {
				u.pendingConditionalBranch = true
			}
		} else {
			u.skippedInCurrentCycle = append(u.skippedInCurrentCycle, runner)
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

		push, stop := u.handleRunner(u.ctx, cycle, &runner)
		if push {
			u.pushedRunnersInCurrentCycle[&runner] = true
			if runner.Runner.InstructionType().IsBranch() {
				u.pushedBranchInCurrentCycle = true
			}
			if runner.Runner.InstructionType().IsConditionalBranch() {
				u.pendingConditionalBranch = true
			}
		} else {
			u.pendings.Push(runner)
			u.skippedInCurrentCycle = append(u.skippedInCurrentCycle, runner)
		}
		if stop {
			return
		}
	}
}

func (u *controlUnit) handleRunner(ctx *risc.Context, cycle int, runner *risc.InstructionRunnerPc) (push, stop bool) {
	if runner.Runner.InstructionType().IsBranch() && u.pushedBranchInCurrentCycle {
		return false, true
	}

	if runner.Runner.InstructionType() == risc.Ret && (!u.outBus.IsEmpty() || u.pendingConditionalBranch) {
		return false, true
	}

	if u.isDataHazardWithSkippedRunners(runner) {
		log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "hazard with skipped runner")
		return false, false
	}

	hazards, hazardTypes := ctx.IsDataHazard3(runner.Runner)
	if len(hazards) == 0 {
		pushed := u.pushRunner(ctx, cycle, runner)
		if !pushed {
			return false, true
		}
		return true, false
	}

	if should, previousRunner, register := u.shouldUseForwarding(runner, hazards, hazardTypes); should {
		ch := make(chan int32, 1)
		previousRunner.Forwarder = ch
		previousRunner.ForwardRegister = register
		runner.Receiver = ch
		runner.ForwardRegister = register

		pushed := u.pushRunner(ctx, cycle, runner)
		if !pushed {
			return false, true
		}
		log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "forward runner on %s (source %d)", register, previousRunner.Pc/4)
		u.forwarding++
		return true, true
	}

	if u.shouldUseRenaming(hazards, hazardTypes) {
		pushed := u.pushRunner(ctx, cycle, runner)
		if !pushed {
			return false, true
		}
		log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "renaming")
		return true, false
	}

	log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%+v, types=%+v", hazards, hazardTypes)
	u.blockedDataHazard++

	// We have to stop here, otherwise we could fall into the case where an
	// instruction is executed even if a branch shouldn't be taken.
	return false, true
}

func (u *controlUnit) isDataHazardWithSkippedRunners(runner *risc.InstructionRunnerPc) bool {
	for _, skippedRunner := range u.skippedInCurrentCycle {
		for _, register := range runner.Runner.ReadRegisters() {
			if register == risc.Zero {
				continue
			}
			for _, skippedRegister := range skippedRunner.Runner.WriteRegisters() {
				if register == skippedRegister {
					// Read after write
					return true
				}
			}
		}

		for _, register := range runner.Runner.WriteRegisters() {
			if register == risc.Zero {
				continue
			}
			for _, skippedRegister := range skippedRunner.Runner.WriteRegisters() {
				if register == skippedRegister {
					// Write after write
					return true
				}
			}
			for _, skippedRegister := range skippedRunner.Runner.ReadRegisters() {
				if register == skippedRegister {
					// Write after read
					return true
				}
			}
		}
	}

	return false
}

func (u *controlUnit) shouldUseForwarding(runner *risc.InstructionRunnerPc, hazards []risc.Hazard, hazardTypes map[risc.HazardType]bool) (bool, *risc.InstructionRunnerPc, risc.RegisterType) {
	if len(hazardTypes) > 1 || !hazardTypes[risc.ReadAfterWrite] || len(hazards) > 1 {
		return false, nil, risc.Zero
	}

	// Can we use forwarding with an instruction pushed in the previous cycle
	for previousRunner := range u.pushedRunnersInPreviousCycle {
		for _, writeRegister := range previousRunner.Runner.WriteRegisters() {
			for _, readRegister := range runner.Runner.ReadRegisters() {
				if readRegister == risc.Zero {
					continue
				}
				if readRegister == writeRegister {
					return true, previousRunner, readRegister
				}
			}
		}
	}
	return false, nil, risc.Zero
}

func (u *controlUnit) shouldUseRenaming(hazards []risc.Hazard, hazardTypes map[risc.HazardType]bool) bool {
	if len(hazards) > 1 {
		return false
	}
	if hazardTypes[risc.ReadAfterWrite] {
		return false
	}
	return true
}

func (u *controlUnit) notifyConditionalBranch() {
	u.pendingConditionalBranch = false
}

func (u *controlUnit) pushRunner(ctx *risc.Context, cycle int, runner *risc.InstructionRunnerPc) bool {
	if !u.outBus.CanAdd() {
		return false
	}

	u.outBus.Add(runner, cycle)
	ctx.AddPendingRegisters(runner.Runner)
	log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner")
	return true
}

func (u *controlUnit) flush() {
	u.pendings = comp.NewQueue[risc.InstructionRunnerPc](pendingLength)
	u.pushedRunnersInPreviousCycle = nil
	u.pendingConditionalBranch = false
}

func (u *controlUnit) isEmpty() bool {
	return u.pendings.Length() == 0
}
