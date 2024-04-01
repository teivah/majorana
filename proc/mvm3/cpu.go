package mvm3

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	cyclesMemoryAccess            = 50 + 1 // +1 cycle to get from l1
	l1ICacheLineSizeInBytes int32 = 64
)

type CPU struct {
	ctx        *risc.Context
	fetchUnit  *fetchUnit
	decodeBus  *comp.SimpleBus[int]
	decodeUnit *decodeUnit
	executeBus *comp.SimpleBus[risc.InstructionRunner]
	eu         *executeUnit
	writeBus   *comp.SimpleBus[comp.ExecutionContext]
	writeUnit  *writeUnit
	branchUnit *simpleBranchUnit
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	return &CPU{
		ctx:        risc.NewContext(debug, memoryBytes),
		fetchUnit:  newFetchUnit(l1ICacheLineSizeInBytes, cyclesMemoryAccess),
		decodeBus:  &comp.SimpleBus[int]{},
		decodeUnit: &decodeUnit{},
		executeBus: &comp.SimpleBus[risc.InstructionRunner]{},
		eu:         &executeUnit{},
		writeBus:   &comp.SimpleBus[comp.ExecutionContext]{},
		writeUnit:  &writeUnit{},
		branchUnit: &simpleBranchUnit{},
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
	cycles := 0
	for {
		cycles += 1

		// Fetch
		m.fetchUnit.cycle(app, m.ctx, m.decodeBus)

		// Decode
		m.decodeUnit.cycle(app, m.decodeBus, m.executeBus)

		// Create branch unit assertions
		m.branchUnit.assert(m.ctx, m.executeBus)

		// Execute
		err := m.eu.cycle(m.ctx, app, m.executeBus, m.writeBus)
		if err != nil {
			return 0, err
		}

		// Branch unit assertions check
		flush := false
		if m.branchUnit.shouldFlushPipeline(m.ctx) {
			flush = true
		}

		// Write back
		m.writeBus.Connect()
		m.writeUnit.cycle(m.ctx, m.writeBus)

		if flush {
			m.flush(m.ctx.Pc)
			continue
		}

		if m.isComplete() {
			if m.ctx.Registers[risc.Ra] != 0 {
				m.ctx.Pc = m.ctx.Registers[risc.Ra]
				m.ctx.Registers[risc.Ra] = 0
				m.fetchUnit.reset(m.ctx.Pc)
				continue
			}
			break
		}
	}
	return cycles, nil
}

func (m *CPU) flush(pc int32) {
	m.fetchUnit.Flush(pc)
	m.decodeUnit.flush()
	m.decodeBus.Flush()
	m.executeBus.Flush()
	m.writeBus.Flush()
}

func (m *CPU) isComplete() bool {
	return m.fetchUnit.IsEmpty() &&
		m.decodeUnit.isEmpty() &&
		m.eu.isEmpty() &&
		m.writeUnit.isEmpty() &&
		m.decodeBus.IsEmpty() &&
		m.executeBus.IsEmpty() &&
		m.writeBus.IsEmpty()
}
