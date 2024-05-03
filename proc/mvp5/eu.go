package mvp5

import (
	"fmt"

	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	processing        bool
	remainingCycles   int
	pendingMemoryRead bool
	addrs             []int32
	memory            []int8
	runner            risc.InstructionRunnerPc
	bu                *btbBranchUnit
	mmu               *memoryManagementUnit
}

func newExecuteUnit(bu *btbBranchUnit, mmu *memoryManagementUnit) *executeUnit {
	return &executeUnit{
		bu:  bu,
		mmu: mmu,
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
		if eu.memory != nil {
			memory = eu.memory
		} else {
			line := eu.mmu.fetchCacheLine(eu.addrs[0])
			eu.mmu.pushLineToL1D(comp.AlignedAddress(eu.addrs[0]), line)
			m, exists := eu.mmu.getFromL1D(eu.addrs)
			if !exists {
				panic("cache line doesn't exist")
			}
			memory = m
		}
		eu.memory = nil
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

	addrs := runner.Runner.MemoryRead(ctx, 0)
	if len(addrs) != 0 {
		if m, exists := eu.mmu.getFromL1D(addrs); exists {
			eu.memory = m
			eu.pendingMemoryRead = true
			eu.remainingCycles = latency.L1Access
		} else {
			eu.addrs = addrs
			eu.pendingMemoryRead = true
			eu.remainingCycles = latency.MemoryAccess
		}
		return false, 0, false, nil
	}

	defer func() {
		eu.runner = risc.InstructionRunnerPc{}
	}()
	return eu.run(ctx, app, outBus, nil)
}

func (eu *executeUnit) run(ctx *risc.Context, app risc.Application, outBus *comp.SimpleBus[risc.ExecutionContext], memory []int8) (bool, int32, bool, error) {
	execution, err := eu.runner.Runner.Run(ctx, app.Labels, eu.runner.Pc, memory, 0)
	if err != nil {
		return false, 0, false, err
	}
	if execution.Return {
		return false, 0, true, nil
	}

	eu.processing = false
	if execution.MemoryChange && eu.mmu.doesExecutionMemoryChangesExistsInL1D(execution) {
		eu.mmu.writeExecutionMemoryChangesToL1D(execution)
		return false, 0, false, nil
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
