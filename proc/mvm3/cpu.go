package mvm3

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	cyclesMemoryAccess            = 50
	flushCycles                   = 1
	l1ICacheLineSizeInBytes int32 = 64
)

type CPU struct {
	ctx         *risc.Context
	fetchUnit   *fetchUnit
	decodeBus   *comp.SimpleBus[int32]
	decodeUnit  *decodeUnit
	executeBus  *comp.SimpleBus[risc.InstructionRunnerPc]
	executeUnit *executeUnit
	writeBus    *comp.SimpleBus[comp.ExecutionContext]
	writeUnit   *writeUnit
	branchUnit  *simpleBranchUnit
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	bu := &simpleBranchUnit{}
	return &CPU{
		ctx:         risc.NewContext(debug, memoryBytes),
		fetchUnit:   newFetchUnit(l1ICacheLineSizeInBytes, cyclesMemoryAccess),
		decodeBus:   &comp.SimpleBus[int32]{},
		decodeUnit:  &decodeUnit{},
		executeBus:  &comp.SimpleBus[risc.InstructionRunnerPc]{},
		executeUnit: newExecuteUnit(bu),
		writeBus:    &comp.SimpleBus[comp.ExecutionContext]{},
		writeUnit:   &writeUnit{},
		branchUnit:  bu,
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
	cycle := 0
	for {
		cycle += 1
		if m.ctx.Debug {
			fmt.Printf("%d\n", int32(cycle))
		}

		// Fetch
		m.fetchUnit.cycle(app, m.ctx, m.decodeBus)

		// Decode
		m.decodeUnit.cycle(app, m.decodeBus, m.executeBus)

		// Create branch unit assertions
		m.branchUnit.assert(m.ctx, m.executeBus)

		// Execute
		flush, pc, err := m.executeUnit.cycle(m.ctx, app, m.executeBus, m.writeBus)
		if err != nil {
			return 0, err
		}

		// Write back
		m.writeUnit.cycle(m.ctx, m.writeBus)

		if flush {
			m.flush(pc)
			cycle += flushCycles
			continue
		}

		if m.isComplete() {
			if m.ctx.Registers[risc.Ra] != 0 {
				m.ctx.Registers[risc.Ra] = 0
				m.fetchUnit.reset(m.ctx.Registers[risc.Ra])
				continue
			}
			break
		}
	}
	return cycle, nil
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
