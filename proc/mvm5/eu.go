package mvm5

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	processing      bool
	remainingCycles int
	runner          risc.InstructionRunnerPc
	bu              *btbBranchUnit
	inBus           *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus          *comp.BufferedBus[comp.ExecutionContext]
}

func newExecuteUnit(bu *btbBranchUnit, inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[comp.ExecutionContext]) *executeUnit {
	return &executeUnit{bu: bu, inBus: inBus, outBus: outBus}
}

func (eu *executeUnit) cycle(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, bool, error) {
	if !eu.processing {
		runner, exists := eu.inBus.Get()
		if !exists {
			return false, 0, false, nil
		}
		eu.runner = runner
		eu.remainingCycles = risc.CyclesPerInstruction(runner.Runner.InstructionType())
		eu.processing = true
	}

	eu.remainingCycles--
	if eu.remainingCycles != 0 {
		return false, 0, false, nil
	}

	if !eu.outBus.CanAdd() {
		eu.remainingCycles = 1
		return false, 0, false, nil
	}

	runner := eu.runner
	// Create the branch unit assertions
	eu.bu.assert(runner)

	// To avoid writeback hazard, if the pipeline contains read registers not
	// written yet, we wait for it
	if ctx.ContainWrittenRegisters(runner.Runner.ReadRegisters()) {
		eu.remainingCycles = 1
		return false, 0, false, nil
	}

	if ctx.Debug {
		fmt.Printf("\tEU: Executing instruction %d\n", eu.runner.Pc/4)
	}
	execution, err := runner.Runner.Run(ctx, app.Labels, eu.runner.Pc)
	if err != nil {
		return false, 0, false, err
	}
	if execution.Return {
		return false, 0, true, nil
	}
	defer func() {
		eu.runner = risc.InstructionRunnerPc{}
	}()

	eu.outBus.Add(comp.ExecutionContext{
		Execution:       execution,
		InstructionType: runner.Runner.InstructionType(),
		WriteRegisters:  runner.Runner.WriteRegisters(),
	}, cycle)
	ctx.AddWriteRegisters(runner.Runner.WriteRegisters())
	eu.processing = false

	if risc.IsJump(runner.Runner.InstructionType()) {
		eu.bu.notifyJumpAddressResolved(eu.runner.Pc, execution.NextPc)
	}

	if execution.PcChange && eu.bu.shouldFlushPipeline(execution.NextPc) {
		return true, execution.NextPc, false, nil
	}

	return false, 0, false, nil
}

func (eu *executeUnit) flush() {
	eu.processing = false
	eu.remainingCycles = 0
}

func (eu *executeUnit) isEmpty() bool {
	return !eu.processing
}
