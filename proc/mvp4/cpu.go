package mvp4

import (
	"fmt"

	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	bytes            = 1
	kilobytes        = 1024
	l1ICacheLineSize = 64 * bytes
	liICacheSize     = 1 * kilobytes
	l1DCacheLineSize = 64 * bytes
	liDCacheSize     = 1 * kilobytes
)

type CPU struct {
	ctx                  *risc.Context
	fetchUnit            *fetchUnit
	decodeBus            *comp.SimpleBus[int32]
	decodeUnit           *decodeUnit
	executeBus           *comp.SimpleBus[risc.InstructionRunnerPc]
	executeUnit          *executeUnit
	writeBus             *comp.SimpleBus[risc.ExecutionContext]
	writeUnit            *writeUnit
	branchUnit           *simpleBranchUnit
	memoryManagementUnit *memoryManagementUnit
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	bu := &simpleBranchUnit{}
	ctx := risc.NewContext(debug, memoryBytes, false)
	mmu := newMemoryManagementUnit(ctx)
	return &CPU{
		ctx:                  ctx,
		fetchUnit:            newFetchUnit(mmu, latency.MemoryAccess),
		decodeBus:            &comp.SimpleBus[int32]{},
		decodeUnit:           &decodeUnit{},
		executeBus:           &comp.SimpleBus[risc.InstructionRunnerPc]{},
		executeUnit:          newExecuteUnit(bu, mmu),
		writeBus:             &comp.SimpleBus[risc.ExecutionContext]{},
		writeUnit:            &writeUnit{},
		branchUnit:           bu,
		memoryManagementUnit: mmu,
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
	cycle := 0
	for {
		cycle++
		if m.ctx.Debug {
			fmt.Printf("%d\n", int32(cycle))
		}

		// Fetch
		m.fetchUnit.cycle(app, m.ctx, m.decodeBus)

		// Decode
		m.decodeUnit.cycle(app, m.decodeBus, m.executeBus)

		// Create branch unit assertions

		// Execute
		flush, pc, ret, err := m.executeUnit.cycle(m.ctx, app, m.executeBus, m.writeBus)
		if err != nil {
			return 0, err
		}

		// Write back
		m.writeUnit.cycle(m.ctx, m.writeBus)

		if ret {
			break
		}
		if flush {
			for !m.writeUnit.isEmpty() || !m.writeBus.IsEmpty() {
				cycle++
				m.writeUnit.cycle(m.ctx, m.writeBus)
			}
			m.flush(pc)
			continue
		}

		if m.isComplete() {
			//if m.ctx.Registers[risc.Ra] != 0 {
			//	m.ctx.Registers[risc.Ra] = 0
			//	m.fetchUnit.reset(m.ctx.Registers[risc.Ra])
			//	continue
			//}
			break
		}
	}
	cycle += m.memoryManagementUnit.flush()
	return cycle, nil
}

func (m *CPU) Stats() map[string]any {
	return nil
}

func (m *CPU) flush(pc int32) {
	m.fetchUnit.flush(pc)
	m.decodeUnit.flush()
	m.decodeBus.Flush()
	m.executeBus.Flush()
	m.writeBus.Flush()
	m.ctx.Flush()
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
