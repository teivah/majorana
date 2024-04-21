package mvp6_0

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	bytes     = 1
	kilobytes = 1024

	cyclesMemoryAccess = 50
	cycleL1DAccess     = 1
	flushCycles        = 1

	l1ICacheLineSize = 64 * bytes
	liICacheSize     = 1 * kilobytes
	l1DCacheLineSize = 64 * bytes
	liDCacheSize     = 1 * kilobytes
)

type CPU struct {
	ctx                  *risc.Context
	fetchUnit            *fetchUnit
	decodeBus            *comp.BufferedBus[int32]
	decodeUnit           *decodeUnit
	controlBus           *comp.BufferedBus[risc.InstructionRunnerPc]
	controlUnit          *controlUnit
	executeBus           *comp.BufferedBus[*risc.InstructionRunnerPc]
	executeUnits         []*executeUnit
	writeBus             *comp.BufferedBus[risc.ExecutionContext]
	writeUnits           []*writeUnit
	branchUnit           *btbBranchUnit
	memoryManagementUnit *memoryManagementUnit

	counterFlush int
}

func NewCPU(debug bool, memoryBytes int, eu, wu int) *CPU {
	busSize := 2
	multiplier := 1
	decodeBus := comp.NewBufferedBus[int32](busSize*multiplier, busSize*multiplier)
	controlBus := comp.NewBufferedBus[risc.InstructionRunnerPc](busSize*multiplier, busSize*multiplier)
	executeBus := comp.NewBufferedBus[*risc.InstructionRunnerPc](busSize, busSize)
	writeBus := comp.NewBufferedBus[risc.ExecutionContext](busSize, busSize)

	ctx := risc.NewContext(debug, memoryBytes, false)
	mmu := newMemoryManagementUnit(ctx)
	fu := newFetchUnit(mmu, decodeBus)
	du := newDecodeUnit(decodeBus, controlBus)
	bu := newBTBBranchUnit(4, fu, du)

	eus := make([]*executeUnit, 0, eu)
	for i := 0; i < eu; i++ {
		eus = append(eus, newExecuteUnit(bu, executeBus, writeBus, mmu))
	}

	wus := make([]*writeUnit, 0, wu)
	for i := 0; i < wu; i++ {
		wus = append(wus, newWriteUnit(writeBus))
	}

	return &CPU{
		ctx:                  ctx,
		fetchUnit:            fu,
		decodeBus:            decodeBus,
		decodeUnit:           du,
		controlBus:           controlBus,
		controlUnit:          newControlUnit(controlBus, executeBus),
		executeBus:           executeBus,
		executeUnits:         eus,
		writeBus:             writeBus,
		writeUnits:           wus,
		branchUnit:           bu,
		memoryManagementUnit: mmu,
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
	defer func() {
		log.Infou(m.ctx, "L1d", m.memoryManagementUnit.l1d.String())
	}()
	cycle := 0
	for {
		cycle++
		log.Info(m.ctx, "Cycle %d", cycle)
		m.decodeBus.Connect(cycle)
		m.controlBus.Connect(cycle)
		m.executeBus.Connect(cycle)
		m.writeBus.Connect(cycle)

		// Fetch
		m.fetchUnit.cycle(cycle, app, m.ctx)

		// Decode
		m.decodeUnit.cycle(cycle, app, m.ctx)

		// Control
		m.controlUnit.cycle(cycle, m.ctx)

		// Execute
		var (
			flush bool
			from  int32
			pc    int32
			ret   bool
		)
		for _, eu := range m.executeUnits {
			f, fp, p, r, err := eu.cycle(cycle, m.ctx, app)
			if err != nil {
				return 0, err
			}
			if f {
				from = fp
			}
			flush = flush || f
			pc = max(pc, p)
			ret = ret || r
		}

		// Write back
		for _, wu := range m.writeUnits {
			wu.cycle(m.ctx, -1)
		}
		log.Info(m.ctx, "\tRegisters: %v", m.ctx.Registers)

		if ret {
			log.Info(m.ctx, "\t🛑 Return")
			m.counterFlush++
			cycle++
			m.writeBus.Connect(cycle)
			for !m.areWriteUnitsEmpty() || !m.writeBus.IsEmpty() {
				for _, wu := range m.writeUnits {
					wu.cycle(m.ctx, -1)
				}
				cycle++
				m.writeBus.Connect(cycle)
			}
			break
		}
		if flush {
			// TODO Same checks as in MVP 6.1
			m.writeBus.Connect(cycle + 1)
			for _, wu := range m.writeUnits {
				for !wu.isEmpty() || !m.writeBus.IsEmpty() {
					cycle++
					wu.cycle(m.ctx, from)
				}
			}

			log.Info(m.ctx, "\t️⚠️ Flush to %d", pc/4)
			m.flush(pc)
			cycle += flushCycles
			log.Info(m.ctx, "\tRegisters: %v", m.ctx.Registers)
			continue
		}

		if m.isEmpty() {
			//if m.ctx.Registers[risc.Ra] != 0 {
			//	m.ctx.Registers[risc.Ra] = 0
			//	m.fetchUnit.reset(m.ctx.Registers[risc.Ra], false)
			//	continue
			//}
			break
		}
	}
	cycle += m.memoryManagementUnit.flush()
	return cycle, nil
}

func (m *CPU) Stats() map[string]any {
	return map[string]any{
		"flush":                  m.counterFlush,
		"du_pending_read":        m.decodeUnit.pendingRead.Stats(),
		"du_blocked":             m.decodeUnit.blocked.Stats(),
		"du_pushed":              m.decodeUnit.pushed.Stats(),
		"cu_push":                m.controlUnit.pushed.Stats(),
		"cu_pending":             m.controlUnit.pending.Stats(),
		"cu_pending_read":        m.controlUnit.pendingRead.Stats(),
		"cu_blocked":             m.controlUnit.blocked.Stats(),
		"cu_total":               m.controlUnit.total,
		"cu_cant_add":            m.controlUnit.cantAdd,
		"cu_blocked_branch":      m.controlUnit.blockedBranch,
		"cu_blocked_data_hazard": m.controlUnit.blockedDataHazard,
	}
}

func (m *CPU) flush(pc int32) {
	m.fetchUnit.flush(pc)
	m.decodeUnit.flush()
	m.controlUnit.flush()
	for _, executeUnit := range m.executeUnits {
		executeUnit.flush()
	}
	m.decodeBus.Clean()
	m.controlBus.Clean()
	m.executeBus.Clean()
	m.writeBus.Clean()
	m.ctx.Flush()
}

func (m *CPU) isEmpty() bool {
	empty := m.fetchUnit.isEmpty() &&
		m.decodeUnit.isEmpty() &&
		m.controlUnit.isEmpty() &&
		m.areWriteUnitsEmpty() &&
		m.decodeBus.IsEmpty() &&
		m.controlBus.IsEmpty() &&
		m.executeBus.IsEmpty() &&
		m.writeBus.IsEmpty()
	if !empty {
		return false
	}
	for _, eu := range m.executeUnits {
		if !eu.isEmpty() {
			return false
		}
	}
	return true
}

func (m *CPU) areWriteUnitsEmpty() bool {
	for _, wu := range m.writeUnits {
		if !wu.isEmpty() {
			return false
		}
	}
	return true
}
