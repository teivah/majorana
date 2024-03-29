package mvm4

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type alu struct {
	processing      bool
	remainingCycles int
	runner          risc.InstructionRunner
	bu              *btbBranchUnit
}

func NewALU(bu *btbBranchUnit) *alu {
	return &alu{
		bu: bu,
	}
}

func (eu *alu) cycle(currentCycle int, ctx *risc.Context, app risc.Application, inBus comp.Bus[risc.InstructionRunner], outBus comp.Bus[comp.ExecutionContext]) error {
	if !eu.processing {
		if !inBus.IsElementInQueue() {
			return nil
		}
		runner := inBus.Get()
		eu.runner = runner
		eu.remainingCycles = risc.CyclesPerInstruction[runner.InstructionType()]
		eu.processing = true
	}

	eu.remainingCycles--
	if eu.remainingCycles != 0 {
		return nil
	}

	if outBus.IsBufferFull() {
		eu.remainingCycles = 1
		return nil
	}

	runner := eu.runner

	// To avoid writeback hazard, if the pipeline contains read registers not written yet, we wait for it.
	if ctx.ContainWrittenRegisters(runner.ReadRegisters()) {
		eu.remainingCycles = 1
		return nil
	}

	execution, err := runner.Run(ctx, app.Labels)
	if err != nil {
		return err
	}

	pc := ctx.Pc
	ctx.Pc = execution.Pc
	outBus.Add(comp.ExecutionContext{
		Execution:       execution,
		InstructionType: runner.InstructionType(),
		WriteRegisters:  runner.WriteRegisters(),
	}, currentCycle)
	ctx.AddWriteRegisters(runner.WriteRegisters())
	eu.runner = nil
	eu.processing = false

	if eu.bu != nil {
		if risc.IsJump(runner.InstructionType()) {
			if ctx.Debug {
				fmt.Printf("\tEU: Branch notify, from %d to %d\n", pc/4, execution.Pc/4)
			}
			eu.bu.branchNotify(pc, execution.Pc)
		}
	}

	return nil
}

func (eu *alu) isEmpty() bool {
	return !eu.processing
}
