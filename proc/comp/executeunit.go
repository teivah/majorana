package comp

import "github.com/teivah/ettore/risc"

type ExecuteUnit struct {
	Processing      bool
	RemainingCycles float32
	Runner          risc.InstructionRunner
}

func (eu *ExecuteUnit) Cycle(currentCycle float32, ctx *risc.Context, application risc.Application, inBus Bus[risc.InstructionRunner], outBus Bus[ExecutionContext]) error {
	if !eu.Processing {
		if !inBus.IsElementInQueue() {
			return nil
		}
		runner := inBus.Get()
		eu.Runner = runner
		eu.RemainingCycles = risc.CyclesPerInstruction[runner.InstructionType()]
		eu.Processing = true
	}

	eu.RemainingCycles--
	if eu.RemainingCycles != 0 {
		return nil
	}

	if outBus.IsBufferFull() {
		eu.RemainingCycles = 1
		return nil
	}

	runner := eu.Runner

	// To avoid writeback hazard, if the pipeline contains read registers not written yet, we wait for it.
	if ctx.ContainWrittenRegisters(runner.ReadRegisters()) {
		eu.RemainingCycles = 1
		return nil
	}

	execution, err := runner.Run(ctx, application.Labels)
	if err != nil {
		return err
	}

	ctx.Pc = execution.Pc
	outBus.Add(ExecutionContext{
		Execution:       execution,
		InstructionType: runner.InstructionType(),
		WriteRegisters:  runner.WriteRegisters(),
	}, currentCycle)
	ctx.AddWriteRegisters(runner.WriteRegisters())
	eu.Runner = nil
	eu.Processing = false
	return nil
}

func (eu *ExecuteUnit) IsEmpty() bool {
	return !eu.Processing
}
