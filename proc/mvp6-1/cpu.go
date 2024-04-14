package mvp6_1

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
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	busSize := 2
	multiplier := 1
	decodeBus := comp.NewBufferedBus[int32](busSize*multiplier, busSize*multiplier)
	controlBus := comp.NewBufferedBus[risc.InstructionRunnerPc](busSize*multiplier, busSize*multiplier)
	executeBus := comp.NewBufferedBus[*risc.InstructionRunnerPc](busSize, busSize)
	writeBus := comp.NewBufferedBus[risc.ExecutionContext](busSize, busSize)

	ctx := risc.NewContext(debug, memoryBytes)
	mmu := newMemoryManagementUnit(ctx)
	fu := newFetchUnit(ctx, mmu, decodeBus)
	du := newDecodeUnit(decodeBus, controlBus)
	bu := newBTBBranchUnit(4, fu, du)
	return &CPU{
		ctx:         ctx,
		fetchUnit:   fu,
		decodeBus:   decodeBus,
		decodeUnit:  du,
		controlBus:  controlBus,
		controlUnit: newControlUnit(controlBus, executeBus),
		executeBus:  executeBus,
		executeUnits: []*executeUnit{
			newExecuteUnit(bu, executeBus, writeBus, mmu),
			newExecuteUnit(bu, executeBus, writeBus, mmu),
			newExecuteUnit(bu, executeBus, writeBus, mmu),
		},
		writeBus: writeBus,
		writeUnits: []*writeUnit{
			newWriteUnit(writeBus),
			newWriteUnit(writeBus),
			//newWriteUnit(writeBus),
		},
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
		//if cycle > 1000 {
		//	return 0, nil
		//}
		log.Info(m.ctx, "Cycle %d", cycle)
		m.decodeBus.Connect(cycle)
		m.controlBus.Connect(cycle)
		m.executeBus.Connect(cycle)
		m.writeBus.Connect(cycle)

		// Fetch
		_ = m.fetchUnit.Cycle(fuReq{cycle, app, m.ctx})

		// Decode
		m.decodeUnit.cycle(cycle, app, m.ctx)

		// Control
		m.controlUnit.cycle(cycle, m.ctx)

		// Execute
		var (
			flush      bool
			sequenceID int32
			pc         int32
			ret        bool
		)
		for i, eu := range m.executeUnits {
			log.Infou(m.ctx, "EU", "Execute unit %d", i)
			eu.sequenceID = sequenceID
			resp := eu.Cycle(euReq{cycle, m.ctx, app})
			if resp.err != nil {
				return 0, resp.err
			}
			if resp.flush {
				sequenceID = resp.sequenceID
			}
			flush = flush || resp.flush
			pc = max(pc, resp.pc)
			ret = ret || resp.isReturn
		}

		// Write back
		for _, wu := range m.writeUnits {
			_ = wu.Cycle(wuReq{m.ctx, -1})
		}
		log.Info(m.ctx, "\tRegisters: %v", m.ctx.Registers)

		if ret {
			log.Info(m.ctx, "\t🛑 Return")
			cycle++
			m.writeBus.Connect(cycle)
			for !m.areWriteUnitsEmpty() || !m.writeBus.IsEmpty() {
				for _, wu := range m.writeUnits {
					_ = wu.Cycle(wuReq{m.ctx, -1})
				}
				cycle++
				m.writeBus.Connect(cycle)
			}
			break
		}
		if flush {
			// TODO Comment the rationale for this code
			log.Info(m.ctx, "\t️⚠️ Executing previous unit cycles")

			for _, eu := range m.executeUnits {
				eu.sequenceID = sequenceID
			}
			fromCycle := cycle

			for {
				isEmpty := true
				for _, eu := range m.executeUnits {
					if !eu.isEmpty() {
						isEmpty = false
						resp := eu.Cycle(euReq{fromCycle, m.ctx, app})
						if resp.err != nil {
							return 0, nil
						}
						if resp.flush {
							log.Info(m.ctx, "\t️⚠️️⚠️ Proposition of an inner flush")
							sequenceID = resp.sequenceID
							flush = resp.flush
							pc = resp.pc
							ret = resp.isReturn
						}
					}
					cycle++
				}
				m.writeBus.Connect(cycle + 1)
				for _, wu := range m.writeUnits {
					for !wu.isEmpty() || !m.writeBus.IsEmpty() {
						cycle++
						_ = wu.Cycle(wuReq{m.ctx, sequenceID})
					}
				}
				if isEmpty {
					break
				}
			}

			//m.writeBus.Connect(cycle + 1)
			//for _, wu := range m.writeUnits {
			//	for !wu.isEmpty() || !m.writeBus.IsEmpty() {
			//		cycle++
			//		_ = wu.Cycle(wuReq{m.ctx, from})
			//	}
			//}

			log.Info(m.ctx, "\t️⚠️ Flush to %d", pc/4)
			m.flush(pc)
			cycle += flushCycles
			log.Info(m.ctx, "\tRegisters: %v", m.ctx.Registers)
			continue
		}

		if m.isEmpty() {
			break
		}
	}
	cycle += m.memoryManagementUnit.flush()
	return cycle, nil
}

func (m *CPU) Stats() map[string]any {
	return map[string]any{
		"du_pending_read":        m.decodeUnit.pendingRead.Stats(),
		"du_blocked":             m.decodeUnit.blocked.Stats(),
		"du_pushed":              m.decodeUnit.pushed.Stats(),
		"cu_push":                m.controlUnit.pushed.Stats(),
		"cu_pending":             m.controlUnit.pending.Stats(),
		"cu_pending_read":        m.controlUnit.pendingRead.Stats(),
		"cu_blocked":             m.controlUnit.blocked.Stats(),
		"cu_forward":             m.controlUnit.forwarding,
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
