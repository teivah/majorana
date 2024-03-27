package proc

import (
	"github.com/teivah/ettore/proc/comp"
	"github.com/teivah/ettore/risc"
)

const (
	mvm3CyclesMemoryAccess      float32 = 50. + 1. // +1 cycle to get from l1
	mvm3L1ICacheLineSizeInBytes int32   = 64
)

type decodeUnit struct{}

func (du *decodeUnit) Cycle(currentCycle float32, app risc.Application, inBus comp.Bus[int], outBus comp.Bus[risc.InstructionRunner]) {
	if !inBus.IsElementInQueue() || outBus.IsBufferFull() {
		return
	}
	idx := inBus.Get()
	runner := app.Instructions[idx]
	outBus.Add(runner, currentCycle)
}

func (du *decodeUnit) flush() {}

func (du *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}

type executionContext struct {
	execution       risc.Execution
	instructionType risc.InstructionType
	writeRegisters  []risc.RegisterType
}

type executeUnit struct {
	Processing      bool
	RemainingCycles float32
	Runner          risc.InstructionRunner
}

func (eu *executeUnit) cycle(currentCycle float32, ctx *risc.Context, application risc.Application, inBus comp.Bus[risc.InstructionRunner], outBus comp.Bus[executionContext]) error {
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
	outBus.Add(executionContext{
		execution:       execution,
		instructionType: runner.InstructionType(),
		writeRegisters:  runner.WriteRegisters(),
	}, currentCycle)
	ctx.AddWriteRegisters(runner.WriteRegisters())
	eu.Runner = nil
	eu.Processing = false
	return nil
}

func (eu *executeUnit) isEmpty() bool {
	return !eu.Processing
}

type writeUnit struct{}

func (wu *writeUnit) cycle(ctx *risc.Context, writeBus comp.Bus[executionContext]) {
	if !writeBus.IsElementInQueue() {
		return
	}

	execution := writeBus.Get()
	if risc.WriteBack(execution.instructionType) {
		ctx.Write(execution.execution)
		ctx.DeleteWriteRegisters(execution.writeRegisters)
	}
}

func (wu *writeUnit) isEmpty() bool {
	return true
}

type branchUnit struct {
	conditionBranchingExpected *int32
	jumpCondition              bool
}

func (bu *branchUnit) assert(ctx *risc.Context, executeBus comp.Bus[risc.InstructionRunner]) {
	if executeBus.IsElementInQueue() {
		runner := executeBus.Peek()
		instructionType := runner.InstructionType()
		if risc.Jump(instructionType) {
			bu.jumpCondition = true
		} else if risc.ConditionalBranching(instructionType) {
			bu.ConditionalBranching(ctx.Pc + 4)
		}
	}
}

func (bu *branchUnit) Jump() {
	bu.jumpCondition = true
}

func (bu *branchUnit) ConditionalBranching(expected int32) {
	bu.conditionBranchingExpected = &expected
}

func (bu *branchUnit) pipelineToBeFlushed(ctx *risc.Context, writeBus comp.Bus[executionContext]) bool {
	if !writeBus.IsElementInBuffer() {
		return false
	}

	conditionalBranching := false
	if bu.conditionBranchingExpected != nil {
		conditionalBranching = *bu.conditionBranchingExpected != ctx.Pc
	}
	assert := conditionalBranching || bu.jumpCondition
	bu.conditionBranchingExpected = nil
	bu.jumpCondition = false
	return assert
}

type mvm3 struct {
	ctx         *risc.Context
	fetchUnit   *comp.FetchUnit
	decodeBus   comp.Bus[int]
	decodeUnit  *decodeUnit
	executeBus  comp.Bus[risc.InstructionRunner]
	executeUnit *executeUnit
	writeBus    comp.Bus[executionContext]
	writeUnit   *writeUnit
	branchUnit  *branchUnit
}

func newMvm3(memoryBytes int) *mvm3 {
	return &mvm3{
		ctx:         risc.NewContext(memoryBytes),
		fetchUnit:   comp.NewFetchUnit(mvm3L1ICacheLineSizeInBytes, mvm3CyclesMemoryAccess),
		decodeBus:   comp.NewBufferedBus[int](1, 1),
		decodeUnit:  &decodeUnit{},
		executeBus:  comp.NewBufferedBus[risc.InstructionRunner](1, 1),
		executeUnit: &executeUnit{},
		writeBus:    comp.NewBufferedBus[executionContext](1, 1),
		writeUnit:   &writeUnit{},
		branchUnit:  &branchUnit{},
	}
}

func (m *mvm3) context() *risc.Context {
	return m.ctx
}

func (m *mvm3) run(app risc.Application) (float32, error) {
	var cycles float32 = 0
	for {
		cycles += 1

		// Fetch
		m.fetchUnit.Cycle(cycles, app, m.decodeBus)

		// Decode
		m.decodeBus.Connect(cycles)
		m.decodeUnit.Cycle(cycles, app, m.decodeBus, m.executeBus)

		// Execute
		m.executeBus.Connect(cycles)

		// Create branch unit assertions
		m.branchUnit.assert(m.ctx, m.executeBus)

		// Execute
		err := m.executeUnit.cycle(cycles, m.ctx, app, m.executeBus, m.writeBus)
		if err != nil {
			return 0, err
		}

		// Branch unit assertions check
		flush := false
		if m.branchUnit.pipelineToBeFlushed(m.ctx, m.writeBus) {
			flush = true
		}

		// Write back
		m.writeBus.Connect(cycles)
		m.writeUnit.cycle(m.ctx, m.writeBus)

		if flush {
			if m.writeBus.IsElementInBuffer() {
				// We need to waste a cycle to write the element in the queue buffer
				cycles++
				m.writeBus.Connect(cycles)
				m.writeUnit.cycle(m.ctx, m.writeBus)
			}
			m.flush(m.ctx.Pc)
		}
		if m.isComplete() {
			break
		}
	}
	return cycles, nil
}

func (m *mvm3) flush(pc int32) {
	m.fetchUnit.Flush(pc)
	m.decodeUnit.flush()
	m.decodeBus.Flush()
	m.executeBus.Flush()
	m.writeBus.Flush()
}

func (m *mvm3) isComplete() bool {
	return m.fetchUnit.IsEmpty() &&
		m.decodeUnit.isEmpty() &&
		m.executeUnit.isEmpty() &&
		m.writeUnit.isEmpty() &&
		m.decodeBus.IsEmpty() &&
		m.executeBus.IsEmpty() &&
		m.writeBus.IsEmpty()
}
