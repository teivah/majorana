package mvp6_0

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type executeUnit struct {
	pending           bool
	remainingCycles   int
	pendingMemoryRead bool
	addrs             []int32
	memory            []int8
	runner            risc.InstructionRunnerPc
	bu                *btbBranchUnit
	inBus             *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus            *comp.BufferedBus[risc.ExecutionContext]
	mmu               *memoryManagementUnit
}

func newExecuteUnit(bu *btbBranchUnit, inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.ExecutionContext], mmu *memoryManagementUnit) *executeUnit {
	return &executeUnit{bu: bu, inBus: inBus, outBus: outBus, mmu: mmu}
}

func (eu *executeUnit) cycle(cycle int, ctx *risc.Context, app risc.Application) (bool, int32, int32, bool, error) {
	if eu.pendingMemoryRead {
		eu.remainingCycles--
		if eu.remainingCycles != 0 {
			return false, 0, 0, false, nil
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
			eu.mmu.pushLineToL1D(eu.addrs[0], line)
			m, exists := eu.mmu.getFromL1D(eu.addrs)
			if !exists {
				panic("cache line doesn't exist")
			}
			memory = m
		}
		eu.memory = nil
		return eu.run(cycle, ctx, app, memory)
	}

	if !eu.pending {
		runner, exists := eu.inBus.Get()
		if !exists {
			return false, 0, 0, false, nil
		}
		eu.runner = runner
		eu.remainingCycles = runner.Runner.InstructionType().Cycles()
		eu.pending = true
	}

	eu.remainingCycles--
	if eu.remainingCycles != 0 {
		return false, 0, 0, false, nil
	}

	if !eu.outBus.CanAdd() {
		eu.remainingCycles = 1
		log.Infou(ctx, "EU", "can't add")
		return false, 0, 0, false, nil
	}

	runner := eu.runner
	// Create the branch unit assertions
	eu.bu.assert(runner)

	log.Infoi(ctx, "EU", runner.Runner.InstructionType(), runner.Pc, "executing")

	addrs := runner.Runner.MemoryRead(ctx)
	if len(addrs) != 0 {
		if m, exists := eu.mmu.getFromL1D(addrs); exists {
			eu.memory = m
			eu.pendingMemoryRead = true
			eu.remainingCycles = cyclesL1Access
		} else {
			eu.addrs = addrs
			eu.pendingMemoryRead = true
			eu.remainingCycles = cyclesMemoryAccess
		}
		return false, 0, 0, false, nil
	}

	defer func() {
		eu.runner = risc.InstructionRunnerPc{}
	}()
	return eu.run(cycle, ctx, app, nil)
}

func (eu *executeUnit) run(cycle int, ctx *risc.Context, app risc.Application, memory []int8) (bool, int32, int32, bool, error) {
	execution, err := eu.runner.Runner.Run(ctx, app.Labels, eu.runner.Pc, memory)
	if err != nil {
		return false, 0, 0, false, err
	}
	if execution.Return {
		return false, 0, 0, true, nil
	}

	eu.pending = false
	if execution.MemoryChange && eu.mmu.doesExecutionMemoryChangesExistsInL1D(execution) {
		eu.mmu.writeExecutionMemoryChangesToL1D(execution)
		return false, 0, 0, false, nil
	}

	eu.outBus.Add(risc.ExecutionContext{
		Pc:              eu.runner.Pc,
		Execution:       execution,
		InstructionType: eu.runner.Runner.InstructionType(),
		WriteRegisters:  eu.runner.Runner.WriteRegisters(),
		ReadRegisters:   eu.runner.Runner.ReadRegisters(),
	}, cycle)

	if eu.runner.Runner.InstructionType().IsUnconditionalBranch() {
		log.Infoi(ctx, "EU", eu.runner.Runner.InstructionType(), eu.runner.Pc,
			"notify jump address resolved from %d to %d", eu.runner.Pc/4, execution.NextPc/4)
		eu.bu.notifyJumpAddressResolved(eu.runner.Pc, execution.NextPc)
	}

	if execution.PcChange && eu.bu.shouldFlushPipeline(execution.NextPc) {
		log.Infoi(ctx, "EU", eu.runner.Runner.InstructionType(), eu.runner.Pc,
			"should be a flush")
		return true, eu.runner.Pc, execution.NextPc, false, nil
	}

	return false, 0, 0, false, nil
}

func (eu *executeUnit) flush() {
	eu.pending = false
	eu.remainingCycles = 0
}

func (eu *executeUnit) isEmpty() bool {
	return !eu.pending
}
