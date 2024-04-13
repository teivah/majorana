package mvp6_0

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
	pendingRead       *obs.Gauge
	blocked           *obs.Gauge
	total             int
	cantAdd           int
	blockedBranch     int
	blockedDataHazard int
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[*risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:       inBus,
		outBus:      outBus,
		pendings:    comp.NewQueue[risc.InstructionRunnerPc](pendingLength),
		pushed:      &obs.Gauge{},
		pending:     &obs.Gauge{},
		pendingRead: &obs.Gauge{},
		blocked:     &obs.Gauge{},
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	pushed := 0
	defer func() {
		u.pushed.Push(pushed)
		u.pending.Push(u.pendings.Length())
	}()
	u.pendingRead.Push(u.inBus.PendingRead())
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
	for elem := range u.pendings.Iterator() {
		runner := u.pendings.Value(elem)

		push, stop := u.handleRunner(ctx, cycle, pushed, runner)
		if push {
			u.pendings.Remove(elem)
			remaining--
			pushed++
		} else {
			u.blockedBranch++
		}
		if stop {
			return
		}
	}

	for remaining > 0 && !u.pendings.IsFull() {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}

		push, stop := u.handleRunner(ctx, cycle, pushed, runner)
		if push {
			remaining--
			pushed++
		} else {
			u.pendings.Push(runner)
			u.blockedBranch++
		}
		if stop {
			return
		}
	}
}

func (u *controlUnit) handleRunner(ctx *risc.Context, cycle int, pushed int, runner risc.InstructionRunnerPc) (push, stop bool) {
	if pushed > 0 && runner.Runner.InstructionType().IsBranch() {
		u.blockedBranch++
		return false, true
	}

	hazards, _ := ctx.IsDataHazard3(runner.Runner)
	if len(hazards) == 0 {
		u.pushRunner(ctx, cycle, &runner)
		return true, false
	} else {
		log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%s", hazards)
		u.blockedDataHazard++
		return false, true
	}
}

func (u *controlUnit) pushRunner(ctx *risc.Context, cycle int, runner *risc.InstructionRunnerPc) {
	u.outBus.Add(runner, cycle)
	ctx.AddPendingRegisters(runner.Runner)
	log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner")
}

func (u *controlUnit) flush() {
	u.pendings = comp.NewQueue[risc.InstructionRunnerPc](pendingLength)
}

func (u *controlUnit) isEmpty() bool {
	return u.pendings.Length() == 0
}
