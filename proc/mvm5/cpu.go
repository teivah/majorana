package mvm5

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
	decodeBus   *comp.BufferedBus[int32]
	decodeUnit  *decodeUnit
	executeBus  *comp.BufferedBus[risc.InstructionRunnerPc]
	executeUnit *executeUnit
	writeBus    *comp.BufferedBus[comp.ExecutionContext]
	writeUnit   *writeUnit
	branchUnit  *btbBranchUnit
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	fu := newFetchUnit(l1ICacheLineSizeInBytes, cyclesMemoryAccess)
	du := &decodeUnit{}
	bu := newBTBBranchUnit(4, fu, du)
	return &CPU{
		ctx:         risc.NewContext(debug, memoryBytes),
		fetchUnit:   fu,
		decodeBus:   comp.NewBufferedBus[int32](2, 2),
		decodeUnit:  du,
		executeBus:  comp.NewBufferedBus[risc.InstructionRunnerPc](2, 2),
		executeUnit: newExecuteUnit(bu),
		writeBus:    comp.NewBufferedBus[comp.ExecutionContext](2, 2),
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
		m.decodeBus.Connect(cycle)
		m.executeBus.Connect(cycle)
		m.writeBus.Connect(cycle)

		// Fetch
		m.fetchUnit.cycle(cycle, app, m.ctx, m.decodeBus)

		// Decode
		m.decodeUnit.cycle(cycle, app, m.ctx, m.decodeBus, m.executeBus)

		// Execute
		flush, pc, ret, err := m.executeUnit.cycle(cycle, m.ctx, app, m.executeBus, m.writeBus)
		if err != nil {
			return 0, err
		}

		// Write back
		m.writeUnit.cycle(m.ctx, m.writeBus)
		if m.ctx.Debug {
			fmt.Printf("\tMemory: %v\n", m.ctx.Registers)
		}

		if ret {
			return cycle, nil
		}
		if flush {
			m.flush(pc)
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

func (m *CPU) flush(pc int32) {
	m.fetchUnit.flush(pc)
	m.decodeUnit.flush()
	m.executeUnit.flush()
	m.decodeBus.Clean()
	m.executeBus.Clean()
	m.writeBus.Clean()
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
