package proc

import (
	"github.com/teivah/ettore/risc"
)

const (
	mvm3CyclesL1Access     float32 = 1.
	mvm3CyclesMemoryAccess         = 50. + mvm3CyclesL1Access
	mvm3L1ICacheLine       int32   = 64 * 8
)

type bus[T any] struct {
	// Before buffer
	entry  []T
	buffer []T
	queue  []T
	length int
}

func newBus[T any](length int) *bus[T] {
	return &bus[T]{
		entry:  make([]T, 0),
		buffer: make([]T, 0),
		queue:  make([]T, 0),
		length: length,
	}
}

func (bus *bus[T]) flush() {
	bus.entry = make([]T, 0)
	bus.buffer = make([]T, 0)
	bus.queue = make([]T, 0)
}

func (bus *bus[T]) add(t T) {
	bus.entry = append(bus.entry, t)
}

func (bus *bus[T]) get() T {
	elem := bus.queue[0]
	bus.queue = bus.queue[1:]
	return elem
}

func (bus *bus[T]) peek() T {
	return bus.queue[0]
}

func (bus *bus[T]) isFull() bool {
	return len(bus.queue) == bus.length || len(bus.entry) == bus.length
}

func (bus *bus[T]) isEmpty() bool {
	return len(bus.queue) == 0 && len(bus.buffer) == 0 && len(bus.entry) == 0
}

func (bus *bus[T]) containsElementInBuffer() bool {
	return len(bus.buffer) != 0
}

func (bus *bus[T]) containsElementInQueue() bool {
	return len(bus.queue) != 0
}

func (bus *bus[T]) containsElementInEntry() bool {
	return len(bus.entry) != 0
}

func (bus *bus[T]) connect() {
	if len(bus.queue) == bus.length {
		return
	}

	for _, list := range bus.buffer {
		bus.queue = append(bus.queue, list)
	}
	bus.buffer = make([]T, 0)

	for _, list := range bus.entry {
		bus.buffer = append(bus.buffer, list)
	}
	bus.entry = make([]T, 0)
}

type l1i struct {
	boundary [2]int32
}

func (l1i *l1i) present(pc int32) bool {
	return pc >= l1i.boundary[0] && pc <= l1i.boundary[1]
}

func (l1i *l1i) fetch(pc int32) {
	l1i.boundary = [2]int32{pc, pc + mvm3L1ICacheLine}
}

type fetchUnit struct {
	pc              int32
	l1i             l1i
	remainingCycles float32
	complete        bool
	processing      bool
}

func newFetchUnit() *fetchUnit {
	return &fetchUnit{
		l1i: l1i{boundary: [2]int32{-1, -1}},
	}
}

func (fu *fetchUnit) cycle(application risc.Application, outBus *bus[int]) {
	if fu.complete {
		return
	}

	if !fu.processing {
		fu.processing = true
		if fu.l1i.present(fu.pc) {
			fu.remainingCycles = mvm3CyclesL1Access
		} else {
			fu.remainingCycles = mvm3CyclesMemoryAccess
			// Should be done after the processing of the 50 cycles
			fu.l1i.fetch(fu.pc)
		}
	}

	fu.remainingCycles -= 1.0
	if fu.remainingCycles == 0.0 {
		if outBus.isFull() {
			fu.remainingCycles = 1.0
			return
		}

		fu.processing = false
		currentPC := fu.pc
		fu.pc += 4
		if fu.pc/4 >= int32(len(application.Instructions)) {
			fu.complete = true
		}
		outBus.add(int(currentPC / 4))
	}
}

func (fu *fetchUnit) flush(pc int32) {
	fu.processing = false
	fu.complete = false
	fu.pc = pc
}

func (fu *fetchUnit) isEmpty() bool {
	return fu.complete
}

type decodeUnit struct{}

func (du *decodeUnit) Cycle(app risc.Application, inBus *bus[int], outBus *bus[risc.InstructionRunner]) {
	if !inBus.containsElementInQueue() || outBus.isFull() {
		return
	}
	idx := inBus.get()
	runner := app.Instructions[idx]
	outBus.add(runner)
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

func (eu *executeUnit) cycle(ctx *risc.Context, application risc.Application, inBus *bus[risc.InstructionRunner], outBus *bus[executionContext]) error {
	if !eu.Processing {
		if !inBus.containsElementInQueue() {
			return nil
		}
		runner := inBus.get()
		eu.Runner = runner
		eu.RemainingCycles = risc.CyclesPerInstruction[runner.InstructionType()]
		eu.Processing = true
	}

	eu.RemainingCycles--
	if eu.RemainingCycles != 0 {
		return nil
	}

	if outBus.isFull() {
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
	outBus.add(executionContext{
		execution:       execution,
		instructionType: runner.InstructionType(),
		writeRegisters:  runner.WriteRegisters(),
	})
	ctx.AddWriteRegisters(runner.WriteRegisters())
	eu.Runner = nil
	eu.Processing = false
	return nil
}

func (eu *executeUnit) isEmpty() bool {
	return !eu.Processing
}

type writeUnit struct{}

func (wu *writeUnit) cycle(ctx *risc.Context, writeBus *bus[executionContext]) {
	if !writeBus.containsElementInQueue() {
		return
	}

	execution := writeBus.get()
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

func (bu *branchUnit) assert(ctx *risc.Context, executeBus *bus[risc.InstructionRunner]) {
	if executeBus.containsElementInQueue() {
		runner := executeBus.peek()
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

func (bu *branchUnit) pipelineToBeFlushed(ctx *risc.Context, writeBus *bus[executionContext]) bool {
	if !writeBus.containsElementInEntry() {
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
	fetchUnit   *fetchUnit
	decodeBus   *bus[int]
	decodeUnit  *decodeUnit
	executeBus  *bus[risc.InstructionRunner]
	executeUnit *executeUnit
	writeBus    *bus[executionContext]
	writeUnit   *writeUnit
	branchUnit  *branchUnit
}

func newMvm3(memoryBytes int) *mvm3 {
	return &mvm3{
		ctx:         risc.NewContext(memoryBytes),
		fetchUnit:   newFetchUnit(),
		decodeBus:   newBus[int](1),
		decodeUnit:  &decodeUnit{},
		executeBus:  newBus[risc.InstructionRunner](1),
		executeUnit: &executeUnit{},
		writeBus:    newBus[executionContext](1),
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
		m.fetchUnit.cycle(app, m.decodeBus)

		// Decode
		m.decodeBus.connect()
		m.decodeUnit.Cycle(app, m.decodeBus, m.executeBus)

		// Execute
		m.executeBus.connect()

		// Create branch unit assertions
		m.branchUnit.assert(m.ctx, m.executeBus)

		// Execute
		err := m.executeUnit.cycle(m.ctx, app, m.executeBus, m.writeBus)
		if err != nil {
			return 0, err
		}

		// Branch unit assertions check
		flush := false
		if m.branchUnit.pipelineToBeFlushed(m.ctx, m.writeBus) {
			flush = true
		}

		// Write back
		m.writeBus.connect()
		m.writeUnit.cycle(m.ctx, m.writeBus)

		if flush {
			if m.writeBus.containsElementInBuffer() {
				// We need to waste a cycle to write the element in the queue buffer
				cycles++
				m.writeBus.connect()
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
	m.fetchUnit.flush(pc)
	m.decodeUnit.flush()
	m.decodeBus.flush()
	m.executeBus.flush()
	m.writeBus.flush()
}

func (m *mvm3) isComplete() bool {
	return m.fetchUnit.isEmpty() &&
		m.decodeUnit.isEmpty() &&
		m.executeUnit.isEmpty() &&
		m.writeUnit.isEmpty() &&
		m.decodeBus.isEmpty() &&
		m.executeBus.isEmpty() &&
		m.writeBus.isEmpty()
}
