package mvm5

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	pending         bool
	remainingCycles int
	runner          risc.InstructionRunnerPc
	bu              *btbBranchUnit
	inBus           *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus          *comp.BufferedBus[comp.ExecutionContext]
}

func newExecuteUnit(bu *btbBranchUnit, inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[comp.ExecutionContext]) *executeUnit {
	return &executeUnit{bu: bu, inBus: inBus, outBus: outBus}
}

func (u *executeUnit) cycle(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, bool, error) {
	if !u.pending {
		runner, exists := u.inBus.Get()
		if !exists {
			return false, 0, false, nil
		}
		u.runner = runner
		u.remainingCycles = risc.CyclesPerInstruction(runner.Runner.InstructionType())
		u.pending = true
	}

	u.remainingCycles--
	if u.remainingCycles != 0 {
		return false, 0, false, nil
	}

	if !u.outBus.CanAdd() {
		u.remainingCycles = 1
		return false, 0, false, nil
	}

	runner := u.runner
	// Create the branch unit assertions
	u.bu.assert(runner)

	// To avoid writeback hazard, if the pipeline contains read registers not
	// written yet, we wait for it
	if ctx.ContainWrittenRegisters(runner.Runner.ReadRegisters()) {
		u.remainingCycles = 1
		return false, 0, false, nil
	}

	if ctx.Debug {
		fmt.Printf("\tEU: Executing instruction %d\n", u.runner.Pc/4)
	}
	execution, err := runner.Runner.Run(ctx, app.Labels, u.runner.Pc)
	if err != nil {
		return false, 0, false, err
	}
	if execution.Return {
		return false, 0, true, nil
	}
	defer func() {
		u.runner = risc.InstructionRunnerPc{}
	}()

	u.outBus.Add(comp.ExecutionContext{
		Execution:       execution,
		InstructionType: runner.Runner.InstructionType(),
		WriteRegisters:  runner.Runner.WriteRegisters(),
	}, cycle)
	ctx.AddWriteRegisters(runner.Runner.WriteRegisters())
	u.pending = false

	if risc.IsJump(runner.Runner.InstructionType()) {
		if ctx.Debug {
			fmt.Printf("\tEU: Notify jump address resolved from %d to %d\n", u.runner.Pc/4, execution.NextPc/4)
		}
		u.bu.notifyJumpAddressResolved(u.runner.Pc, execution.NextPc)
	}

	if execution.PcChange && u.bu.shouldFlushPipeline(execution.NextPc) {
		return true, execution.NextPc, false, nil
	}

	return false, 0, false, nil
}

func (u *executeUnit) flush() {
	u.pending = false
	u.remainingCycles = 0
}

func (u *executeUnit) isEmpty() bool {
	return !u.pending
}
