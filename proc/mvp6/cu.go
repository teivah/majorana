package mvp6

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
	inBus    *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus   *comp.BufferedBus[*risc.InstructionRunnerPc]
	pendings *comp.Queue[risc.InstructionRunnerPc]

	pushed            *obs.Gauge
	pending           *obs.Gauge
	blocked           *obs.Gauge
	forwarding        int
	total             int
	cantAdd           int
	blockedBranch     int
	blockedDataHazard int
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[*risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:    inBus,
		outBus:   outBus,
		pendings: comp.NewQueue[risc.InstructionRunnerPc](pendingLength),
		pushed:   &obs.Gauge{},
		pending:  &obs.Gauge{},
		blocked:  &obs.Gauge{},
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	pushed := 0
	defer func() {
		u.pushed.Push(pushed)
		u.pending.Push(u.pendings.Length())
	}()
	if u.inBus.CanGet() {
		u.blocked.Push(1)
	} else {
		u.blocked.Push(0)
	}
	u.total++

	if !u.outBus.CanAdd() {
		u.cantAdd++
		log.Infou(ctx, "CU", "can't add")
		return
	}

	remaining := u.outBus.RemainingToAdd()
	var pushedRunners []*risc.InstructionRunnerPc

	defer func() {
		for _, runner := range pushedRunners {
			ctx.AddPendingRegisters(runner.Runner)
		}
	}()

	for elem := range u.pendings.Iterator() {
		pending := u.pendings.Value(elem)
		if pushed > 0 && pending.Runner.InstructionType().IsBranch() {
			u.blockedBranch++
			return
		}

		hazard, reason := ctx.IsDataHazard(pending.Runner)
		if !hazard {
			hazard, conflictRunner, register := isHazardWithPushedRunners(pushedRunners, pending.Runner)
			if hazard {
				ch := make(chan int32, 1)
				conflictRunner.Forwarder = ch
				conflictRunner.ForwardRegister = register
				pending.Receiver = ch
				pending.ForwardRegister = register
				log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "pushing forward runner")
				u.forwarding++
			} else {
				log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "pushing runner")
			}

			u.outBus.Add(&pending, cycle)
			u.pendings.Remove(elem)
			pushedRunners = append(pushedRunners, &pending)
			remaining--
			pushed++
		} else {
			log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "data hazard: reason=%s", reason)
			u.blockedDataHazard++
			return
		}
	}

	for remaining > 0 && !u.pendings.IsFull() {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}
		if pushed > 0 && runner.Runner.InstructionType().IsBranch() {
			u.pendings.Push(runner)
			u.blockedBranch++
			return
		}

		hazard, reason := ctx.IsDataHazard(runner.Runner)
		if !hazard {
			hazard, conflictRunner, register := isHazardWithPushedRunners(pushedRunners, runner.Runner)
			if hazard {
				ch := make(chan int32, 1)
				conflictRunner.Forwarder = ch
				conflictRunner.ForwardRegister = register
				runner.Receiver = ch
				runner.ForwardRegister = register
				log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing forward runner")
				u.forwarding++
			} else {
				log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner")
			}

			u.outBus.Add(&runner, cycle)
			pushedRunners = append(pushedRunners, &runner)
			remaining--
			pushed++
		} else {
			u.pendings.Push(runner)
			log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%s", reason)
			u.blockedDataHazard++
		}
	}
}

func isHazardWithPushedRunners(pushedRunners []*risc.InstructionRunnerPc, runner risc.InstructionRunner) (bool, *risc.InstructionRunnerPc, risc.RegisterType) {
	pendingWriteRegisters := make(map[risc.RegisterType]*risc.InstructionRunnerPc)
	for _, runner := range pushedRunners {
		for _, register := range runner.Runner.WriteRegisters() {
			if register == risc.Zero {
				continue
			}
			pendingWriteRegisters[register] = runner
		}
	}

	for _, register := range runner.ReadRegisters() {
		if register == risc.Zero {
			continue
		}
		if v, exists := pendingWriteRegisters[register]; exists {
			// An instruction needs to read from a register that was updated
			return true, v, register
		}
	}
	return false, nil, risc.Zero
}

func (u *controlUnit) flush() {
	u.pendings = comp.NewQueue[risc.InstructionRunnerPc](pendingLength)
}

func (u *controlUnit) isEmpty() bool {
	return u.pendings.Length() == 0
}
