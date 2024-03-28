package mvm4

import (
	"fmt"

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
	decodeUnit  *comp.DecodeUnitWithBranchPredictor
	executeBus  comp.Bus[risc.InstructionRunner]
	executeUnit *comp.ExecuteUnit
	writeBus    comp.Bus[comp.ExecutionContext]
	writeUnit   *comp.WriteUnit
	branchUnit  *comp.BTBBranchUnit
}

func NewCPU(memoryBytes int) *CPU {
	fu := comp.NewFetchUnit(l1ICacheLineSizeInBytes, cyclesMemoryAccess)
	bu := comp.NewBTBBranchUnit(4, fu)
	return &CPU{
		ctx:         risc.NewContext(memoryBytes),
		fetchUnit:   fu,
		decodeBus:   comp.NewBufferedBus[int](1, 1),
		decodeUnit:  comp.NewDecodeUnitWithBranchPredictor(bu),
		executeBus:  comp.NewBufferedBus[risc.InstructionRunner](1, 1),
		executeUnit: comp.NewExecuteUnitWithBu(bu),
		writeBus:    comp.NewBufferedBus[comp.ExecutionContext](1, 1),
		writeUnit:   &comp.WriteUnit{},
		branchUnit:  bu,
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (float32, error) {
	var cycles float32 = 0
	for {
		cycles += 1
		if app.Debug {
			fmt.Printf("Cycle %d\n", int32(cycles))
		}

		// Fetch
		m.fetchUnit.Cycle(cycles, app, m.decodeBus)

		// Decode
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
		if m.branchUnit.ShouldFlushPipeline(m.ctx, m.writeBus) {
			if app.Debug {
				fmt.Println("\tFlush")
			}
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
