package mvm5

import (
	"slices"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type controlUnit struct {
	inBus    *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus   *comp.BufferedBus[risc.InstructionRunnerPc]
	pendings []risc.InstructionRunnerPc
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:  inBus,
		outBus: outBus,
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	if !u.outBus.CanAdd() {
		logu(ctx, "CU", "can't add")
		return
	}
	if ctx.IsControlHazard() {
		logu(ctx, "CU", "control hazard")
		return
	}

	pushed := 0
	remaining := u.outBus.RemainingToAdd()
	for i := 0; i < len(u.pendings) && remaining > 0; i++ {
		pending := u.pendings[i]
		if pushed > 0 && risc.IsBranch(pending.Runner.InstructionType()) {
			return
		}

		hazard, reason := ctx.IsDataHazard(pending.Runner)
		if !hazard {
			u.outBus.Add(pending, cycle)
			ctx.AddPendingRegisters(pending.Runner)
			if risc.IsBranch(pending.Runner.InstructionType()) {
				ctx.SetPendingBranch()
			}
			logi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "pushing runner")
			u.pendings = slices.Delete(u.pendings, i, i+1)
			remaining--
			pushed++
		} else {
			logi(ctx, "CU", pending.Runner.InstructionType(), pending.Pc, "data hazard: reason=%s", reason)
			return
		}
	}

	for remaining > 0 {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}
		if pushed > 0 && risc.IsBranch(runner.Runner.InstructionType()) {
			u.pendings = append(u.pendings, runner)
			return
		}

		hazard, reason := ctx.IsDataHazard(runner.Runner)
		if !hazard {
			u.outBus.Add(runner, cycle)
			ctx.AddPendingRegisters(runner.Runner)
			if risc.IsBranch(runner.Runner.InstructionType()) {
				ctx.SetPendingBranch()
			}
			logi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "pushing runner")
			remaining--
			pushed++
		} else {
			u.pendings = append(u.pendings, runner)
			logi(ctx, "CU", runner.Runner.InstructionType(), runner.Pc, "data hazard: reason=%s", reason)
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
