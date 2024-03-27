package mvm3

import (
	"github.com/teivah/ettore/proc/comp"
	"github.com/teivah/ettore/risc"
)

const (
	cyclesMemoryAccess      float32 = 50. + 1. // +1 cycle to get from l1
	l1ICacheLineSizeInBytes int32   = 64
)

type CPU struct {
	ctx         *risc.Context
	fetchUnit   *comp.FetchUnit
	decodeBus   comp.Bus[int]
	decodeUnit  *comp.DecodeUnit
	executeBus  comp.Bus[risc.InstructionRunner]
	executeUnit *comp.ExecuteUnit
	writeBus    comp.Bus[comp.ExecutionContext]
	writeUnit   *comp.WriteUnit
	branchUnit  *comp.BranchUnit
}

func NewCPU(memoryBytes int) *CPU {
	return &CPU{
		ctx:         risc.NewContext(memoryBytes),
		fetchUnit:   comp.NewFetchUnit(l1ICacheLineSizeInBytes, cyclesMemoryAccess),
		decodeBus:   comp.NewBufferedBus[int](1, 1),
		decodeUnit:  &comp.DecodeUnit{},
		executeBus:  comp.NewBufferedBus[risc.InstructionRunner](1, 1),
		executeUnit: &comp.ExecuteUnit{},
		writeBus:    comp.NewBufferedBus[comp.ExecutionContext](1, 1),
		writeUnit:   &comp.WriteUnit{},
		branchUnit:  &comp.BranchUnit{},
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (float32, error) {
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
		m.branchUnit.Assert(m.ctx, m.executeBus)

		// Execute
		err := m.executeUnit.Cycle(cycles, m.ctx, app, m.executeBus, m.writeBus)
		if err != nil {
			return 0, err
		}

		// Branch unit assertions check
		flush := false
		if m.branchUnit.PipelineToBeFlushed(m.ctx, m.writeBus) {
			flush = true
		}

		// Write back
		m.writeBus.Connect(cycles)
		m.writeUnit.Cycle(m.ctx, m.writeBus)

		if flush {
			if m.writeBus.IsElementInBuffer() {
				// We need to waste a cycle to write the element in the queue buffer
				cycles++
				m.writeBus.Connect(cycles)
				m.writeUnit.Cycle(m.ctx, m.writeBus)
			}
			m.flush(m.ctx.Pc)
		}
		if m.isComplete() {
			break
		}
	}
	return cycles, nil
}

func (m *CPU) flush(pc int32) {
	m.fetchUnit.Flush(pc)
	m.decodeUnit.Flush()
	m.decodeBus.Flush()
	m.executeBus.Flush()
	m.writeBus.Flush()
}

func (m *CPU) isComplete() bool {
	return m.fetchUnit.IsEmpty() &&
		m.decodeUnit.IsEmpty() &&
		m.executeUnit.IsEmpty() &&
		m.writeUnit.IsEmpty() &&
		m.decodeBus.IsEmpty() &&
		m.executeBus.IsEmpty() &&
		m.writeBus.IsEmpty()
}
