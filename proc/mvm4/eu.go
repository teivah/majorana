package mvm4

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
}

func newExecuteUnit(bu *btbBranchUnit) *executeUnit {
	return &executeUnit{
		bu: bu,
	}
}

func (eu *executeUnit) cycle(ctx *risc.Context, app risc.Application, inBus *comp.SimpleBus[risc.InstructionRunnerPc], outBus *comp.SimpleBus[comp.ExecutionContext]) (bool, int32, error) {
	if !eu.processing {
		runner, exists := inBus.Get()
		if !exists {
			return false, 0, nil
		}
		eu.runner = runner
		eu.remainingCycles = risc.CyclesPerInstruction[runner.Runner.InstructionType()]
		eu.processing = true
	}

	eu.remainingCycles--
	if eu.remainingCycles != 0 {
		return false, 0, nil
	}

	if !outBus.CanAdd() {
		eu.remainingCycles = 1
		return false, 0, nil
	}

	runner := eu.runner
	// Create the branch unit assertions
	eu.bu.assert(runner)

	// To avoid writeback hazard, if the pipeline contains read registers not
	// written yet, we wait for it
	if ctx.ContainWrittenRegisters(runner.Runner.ReadRegisters()) {
		eu.remainingCycles = 1
		return false, 0, nil
	}

	if ctx.Debug {
		fmt.Printf("\tEU: Executing instruction %d\n", eu.runner.Pc/4)
	}
	execution, err := runner.Runner.Run(ctx, app.Labels, eu.runner.Pc)
	if err != nil {
		return false, 0, err
	}

	outBus.Add(comp.ExecutionContext{
		Execution:       execution,
		InstructionType: runner.Runner.InstructionType(),
		WriteRegisters:  runner.Runner.WriteRegisters(),
	})
	ctx.AddWriteRegisters(runner.Runner.WriteRegisters())
	eu.runner = risc.InstructionRunnerPc{}
	eu.processing = false

	if risc.IsJump(runner.Runner.InstructionType()) {
		eu.bu.notifyJumpAddressResolved(eu.runner.Pc, execution.Pc)
	}

	if execution.PcChange && eu.bu.shouldFlushPipeline(execution.Pc) {
		return true, execution.Pc, nil
	}

	return false, 0, nil
}

func (eu *executeUnit) flush() {
	eu.processing = false
	eu.remainingCycles = 0
}

func (eu *executeUnit) isEmpty() bool {
	return !eu.processing
}
