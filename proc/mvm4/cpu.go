package mvm4

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
	branchUnit  *btbBranchUnit

	counterFlush int
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	fu := newFetchUnit(l1ICacheLineSizeInBytes, cyclesMemoryAccess)
	du := &decodeUnit{}
	bu := newBTBBranchUnit(4, fu, du)
	return &CPU{
		ctx:         risc.NewContext(debug, memoryBytes),
		fetchUnit:   fu,
		decodeBus:   &comp.SimpleBus[int32]{},
		decodeUnit:  du,
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
		m.decodeUnit.cycle(app, m.ctx, m.decodeBus, m.executeBus)

		// Execute
		flush, pc, ret, err := m.executeUnit.cycle(m.ctx, app, m.executeBus, m.writeBus)
		if err != nil {
			return 0, err
		}

		// Write back
		m.writeUnit.cycle(m.ctx, m.writeBus)
		if m.ctx.Debug {
			fmt.Printf("\tRegisters: %v\n", m.ctx.Registers)
		}

		if ret {
			return cycle, nil
		}
		if flush {
			if m.ctx.Debug {
				fmt.Printf("\tFlush to %d\n", pc/4)
			}
			m.flush(pc)
			m.counterFlush++
			cycle += flushCycles
			continue
		}

		if m.isComplete() {
			if m.ctx.Registers[risc.Ra] != 0 {
				m.ctx.Registers[risc.Ra] = 0
				m.fetchUnit.reset(m.ctx.Registers[risc.Ra], false)
				continue
			}
			break
		}
	}
	return cycle, nil
}

func (m *CPU) Stats() map[string]any {
	return map[string]any{
		"flush": m.counterFlush,
	}
}

func (m *CPU) flush(pc int32) {
	m.fetchUnit.flush(pc)
	m.decodeUnit.flush()
	m.executeUnit.flush()
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
