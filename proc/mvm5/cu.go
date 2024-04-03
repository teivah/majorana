package mvm5

import (
	"slices"

	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/common/obs"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type controlUnit struct {
	inBus    *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus   *comp.BufferedBus[risc.InstructionRunnerPc]
	pendings []risc.InstructionRunnerPc

	pushed            *obs.Gauge
	blocked           *obs.Gauge
	total             int
	cantAdd           int
	blockedBranch     int
	blockedDataHazard int
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:   inBus,
		outBus:  outBus,
		pushed:  &obs.Gauge{},
		blocked: &obs.Gauge{},
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	pushed := 0
	defer func() {
		u.pushed.Push(pushed)
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
	for i := 0; i < len(u.pendings) && remaining > 0; i++ {
		pending := u.pendings[i]
		if pushed > 0 && pending.Runner.InstructionType().IsBranch() {
			u.blockedBranch++
			return
		}

		hazard, reason := ctx.IsDataHazard(pending.Runner)
		if !hazard {
			u.outBus.Add(pending, cycle)
			ctx.AddPendingRegisters(pending.Runner)
			log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "pushing runner")
			// TODO Delete when i=1 and 0 was deleted?
			u.pendings = slices.Delete(u.pendings, i, i+1)
			remaining--
			pushed++
		} else {
			log.Infoi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "data hazard: reason=%s", reason)
			u.blockedDataHazard++
			return
		}
	}

	for remaining > 0 {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}
		if pushed > 0 && runner.Runner.InstructionType().IsBranch() {
			u.pendings = append(u.pendings, runner)
			u.blockedBranch++
			return
		}

		hazard, reason := ctx.IsDataHazard(runner.Runner)
		if !hazard {
			u.outBus.Add(runner, cycle)
			ctx.AddPendingRegisters(runner.Runner)
			log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner")
			remaining--
			pushed++
		} else {
			u.pendings = append(u.pendings, runner)
			log.Infoi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%s", reason)
			u.blockedDataHazard++
			return
		}
	}
}

func (u *controlUnit) flush() {
	u.pendings = nil
}

func (u *controlUnit) isEmpty() bool {
	return len(u.pendings) == 0
}
