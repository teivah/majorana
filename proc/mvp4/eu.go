package mvp4

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	processing        bool
	remainingCycles   int
	pendingMemoryRead bool
	addrs             []int32
	runner            risc.InstructionRunnerPc
	bu                *btbBranchUnit
}

func newExecuteUnit(bu *btbBranchUnit) *executeUnit {
	return &executeUnit{
		bu: bu,
	}
}

func (eu *executeUnit) cycle(ctx *risc.Context, app risc.Application, inBus *comp.SimpleBus[risc.InstructionRunnerPc], outBus *comp.SimpleBus[risc.ExecutionContext]) (bool, int32, bool, error) {
	if eu.pendingMemoryRead {
		eu.remainingCycles--
		if eu.remainingCycles != 0 {
			return false, 0, false, nil
		}
		eu.pendingMemoryRead = false
		defer func() {
			eu.runner = risc.InstructionRunnerPc{}
		}()
		var memory []int8
		for _, addr := range eu.addrs {
			memory = append(memory, ctx.Memory[addr])
		}
		return eu.run(ctx, app, outBus, memory)
	}

	if !eu.processing {
		runner, exists := inBus.Get()
		if !exists {
			return false, 0, false, nil
		}
		eu.runner = runner
		eu.remainingCycles = runner.Runner.InstructionType().Cycles()
		eu.processing = true
	}

	eu.remainingCycles--
	if eu.remainingCycles != 0 {
		return false, 0, false, nil
	}

	if !outBus.CanAdd() {
		eu.remainingCycles = 1
		return false, 0, false, nil
	}

	runner := eu.runner
	// Create the branch unit assertions
	eu.bu.assert(runner)

	// To avoid writeback hazard, if the pipeline contains read registers not
	// written yet, we wait for it
	if ctx.IsWriteDataHazard(runner.Runner.ReadRegisters()) {
		eu.remainingCycles = 1
		return false, 0, false, nil
	}

	if ctx.Debug {
		fmt.Printf("\tEU: Executing instruction %d\n", eu.runner.Pc/4)
	}

	addrs := runner.Runner.MemoryRead(ctx)
	if len(addrs) != 0 {
		eu.addrs = addrs
		eu.pendingMemoryRead = true
		eu.remainingCycles = cyclesMemoryAccess
		return false, 0, false, nil
	}

	defer func() {
		eu.runner = risc.InstructionRunnerPc{}
	}()
	return eu.run(ctx, app, outBus, nil)
}

func (eu *executeUnit) run(ctx *risc.Context, app risc.Application, outBus *comp.SimpleBus[risc.ExecutionContext], memory []int8) (bool, int32, bool, error) {
	execution, err := eu.runner.Runner.Run(ctx, app.Labels, eu.runner.Pc, memory)
	if err != nil {
		return false, 0, false, err
	}
	if execution.Return {
		return false, 0, true, nil
	}

	outBus.Add(risc.ExecutionContext{
		Execution:       execution,
		InstructionType: eu.runner.Runner.InstructionType(),
		WriteRegisters:  eu.runner.Runner.WriteRegisters(),
	})
	ctx.AddPendingWriteRegisters(eu.runner.Runner.WriteRegisters())
	eu.processing = false

	if eu.runner.Runner.InstructionType().IsUnconditionalBranch() {
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
