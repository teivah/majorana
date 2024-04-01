package mvm3

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	processing      bool
	remainingCycles int
	runner          risc.InstructionRunner
}

func (eu *executeUnit) cycle(currentCycle int, ctx *risc.Context, app risc.Application, inBus *comp.SimpleBus[risc.InstructionRunner], outBus *comp.SimpleBus[comp.ExecutionContext]) error {
	if !eu.processing {
		runner, exists := inBus.Get()
		if !exists {
			return nil
		}
		eu.runner = runner
		eu.remainingCycles = risc.CyclesPerInstruction[runner.InstructionType()]
		eu.processing = true
	}

	eu.remainingCycles--
	if eu.remainingCycles != 0 {
		return nil
	}

	if !outBus.CanAdd() {
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

	ctx.Pc = execution.Pc
	outBus.Add(comp.ExecutionContext{
		Execution:       execution,
		InstructionType: runner.InstructionType(),
		WriteRegisters:  runner.WriteRegisters(),
	})
	ctx.AddWriteRegisters(runner.WriteRegisters())
	eu.runner = nil
	eu.processing = false

	return nil
}

func (eu *executeUnit) isEmpty() bool {
	return !eu.processing
}
