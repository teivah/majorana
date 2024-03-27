package comp

import "github.com/teivah/ettore/risc"

type ExecuteUnit struct {
	processing      bool
	remainingCycles float32
	runner          risc.InstructionRunner
	bu              *BTBBranchUnit
}

func (eu *ExecuteUnit) Cycle(currentCycle float32, ctx *risc.Context, application risc.Application, inBus Bus[risc.InstructionRunner], outBus Bus[ExecutionContext]) error {
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

	execution, err := runner.Run(ctx, application.Labels)
	if err != nil {
		return err
	}

	pc := ctx.Pc
	ctx.Pc = execution.Pc
	outBus.Add(ExecutionContext{
		Execution:       execution,
		InstructionType: runner.InstructionType(),
		WriteRegisters:  runner.WriteRegisters(),
	}, currentCycle)
	ctx.AddWriteRegisters(runner.WriteRegisters())
	eu.runner = nil
	eu.processing = false

	if eu.bu != nil {
		if risc.IsJump(runner.InstructionType()) {
			eu.bu.BranchNotify(pc, execution.Pc)
		}
	}

	return nil
}

func (eu *ExecuteUnit) IsEmpty() bool {
	return !eu.processing
}
