package mvm4

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	cyclesMemoryAccess            = 50 + 1 // +1 cycle to get from l1
	l1ICacheLineSizeInBytes int32 = 64
)

type CPU struct {
	ctx         *risc.Context
	fetchUnit   *fetchUnit
	decodeBus   comp.Bus[int]
	decodeUnit  *decodeUnit
	executeBus  comp.Bus[risc.InstructionRunner]
	executeUnit *alu
	writeBus    comp.Bus[comp.ExecutionContext]
	writeUnit   *writeUnit
	branchUnit  *btbBranchUnit
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	fu := NewFetchUnit(l1ICacheLineSizeInBytes, cyclesMemoryAccess)
	bu := newBTBBranchUnit(4, fu)
	return &CPU{
		ctx:         risc.NewContext(debug, memoryBytes),
		fetchUnit:   fu,
		decodeBus:   comp.NewBufferedBus[int](1, 1),
		decodeUnit:  newDecodeUnitWithBranchPredictor(bu),
		executeBus:  comp.NewBufferedBus[risc.InstructionRunner](1, 1),
		executeUnit: NewALU(bu),
		writeBus:    comp.NewBufferedBus[comp.ExecutionContext](1, 1),
		writeUnit:   &writeUnit{},
		branchUnit:  bu,
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
	var cycles int = 0
	for {
		cycles += 1
		if m.ctx.Debug {
			fmt.Printf("%d\n", int32(cycles))
		}

		// Fetch
		m.fetchUnit.cycle(cycles, app, m.ctx, m.decodeBus)

		// Decode
		m.decodeUnit.cycle(cycles, app, m.decodeBus, m.executeBus)

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
		if m.branchUnit.shouldFlushPipeline(m.ctx, m.writeBus) {
			if m.ctx.Debug {
				fmt.Println("\tFlush")
			}
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
			//if m.ctx.Registers[risc.Ra] != 0 {
			//	m.ctx.Pc = m.ctx.Registers[risc.Ra]
			//	m.ctx.Registers[risc.Ra] = 0
			//	m.fetchUnit.Reset(m.ctx.Pc)
			//	continue
			//}
			break
		}
	}
	return cycles, nil
}

func (m *CPU) flush(pc int32) {
	m.fetchUnit.flush(pc)
	m.decodeUnit.flush()
	m.decodeBus.Flush()
	m.executeBus.Flush()
	m.writeBus.Flush()
}

func (m *CPU) isComplete() bool {
	return m.fetchUnit.isEmpty() &&
		m.decodeUnit.isEmpty() &&
		m.executeUnit.isEmpty() &&
		m.writeUnit.isEmpty() &&
		m.decodeBus.IsEmpty() &&
		m.executeBus.IsEmpty() &&
		m.writeBus.IsEmpty()
}
