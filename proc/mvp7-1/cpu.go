package mvp7_1

import (
	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

const (
	bytes     = 1
	kilobytes = 1024

	l1ICacheLineSize = 64 * bytes
	l1ICacheSize     = 1 * kilobytes
	l1DCacheLineSize = 64 * bytes
	l1DCacheSize     = 1 * kilobytes
	l3CacheLineSize  = 64 * bytes
	l3CacheSize      = 1 * kilobytes
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
	cacheControllers     []*cacheController
}

func NewCPU(debug bool, memoryBytes int, parallelism int) *CPU {
	busSize := 2
	multiplier := 1
	decodeBus := comp.NewBufferedBus[int32](busSize*multiplier, busSize*multiplier)
	controlBus := comp.NewBufferedBus[risc.InstructionRunnerPc](busSize*multiplier, busSize*multiplier)
	executeBus := comp.NewBufferedBus[*risc.InstructionRunnerPc](busSize, busSize)
	writeBus := comp.NewBufferedBus[risc.ExecutionContext](busSize, busSize)

	ctx := risc.NewContext(debug, memoryBytes, true)

	mmu := newMemoryManagementUnit(ctx)
	fu := newFetchUnit(ctx, decodeBus)
	du := newDecodeUnit(ctx, decodeBus, controlBus)
	cu := newControlUnit(ctx, controlBus, executeBus)
	bu := newBTBBranchUnit(ctx, 4, fu, du, cu)

	eus := make([]*executeUnit, 0, parallelism)
	wus := make([]*writeUnit, 0, parallelism)
	ccs := make([]*cacheController, 0, parallelism)
	lock := newMSI()
	for i := 0; i < parallelism; i++ {
		cc := newCacheController(i, ctx, mmu, lock)
		ccs = append(ccs, cc)
		eus = append(eus, newExecuteUnit(ctx, bu, executeBus, writeBus, mmu, cc))
		wus = append(wus, newWriteUnit(ctx, writeBus))
	}

	return &CPU{
		ctx:                  ctx,
		fetchUnit:            fu,
		decodeBus:            decodeBus,
		decodeUnit:           du,
		controlBus:           controlBus,
		controlUnit:          cu,
		executeBus:           executeBus,
		executeUnits:         eus,
		writeBus:             writeBus,
		writeUnits:           wus,
		branchUnit:           bu,
		memoryManagementUnit: mmu,
		cacheControllers:     ccs,
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
	m.ctx.InitRAT()
	cycle := 0
	for {
		cycle++
		log.Info(m.ctx, "Cycle %d", cycle)
		m.decodeBus.Connect(cycle)
		m.controlBus.Connect(cycle)
		m.executeBus.Connect(cycle)
		m.writeBus.Connect(cycle)

		// Fetch
		_ = m.fetchUnit.Cycle(fuReq{cycle, app})

		// Decode
		m.decodeUnit.cycle(cycle, app)

		// Control
		m.controlUnit.cycle(cycle)

		for _, cc := range m.cacheControllers {
			cc.snoop.Cycle(struct{}{})
		}

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
			resp := eu.Cycle(euReq{cycle, app})
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

		// Write-back
		for _, wu := range m.writeUnits {
			if flush {
				// In case of a flush, we shouldn't write pending-write instructions.
				_ = wu.Cycle(wuReq{sequenceID})
			} else {
				_ = wu.Cycle(wuReq{-1})
			}
		}
		log.Info(m.ctx, "\tRegisters: %v", m.ctx.Registers)

		if ret {
			log.Info(m.ctx, "\t🛑 Return")
			cycle++
			m.writeBus.Connect(cycle)
			for !m.areWriteUnitsEmpty() || !m.writeBus.IsEmpty() {
				for _, wu := range m.writeUnits {
					_ = wu.Cycle(wuReq{-1})
				}
				cycle++
				m.writeBus.Connect(cycle)
			}
			break
		}
		if flush {
			// Execute pending instructions up to sequenceID.
			log.Info(m.ctx, "\t️⚠️ Executing previous unit cycles")

			for _, eu := range m.executeUnits {
				eu.sequenceID = sequenceID
			}
			fromCycle := cycle

			for {
				isEmpty := true
				cycle++

				for _, cc := range m.cacheControllers {
					cc.snoop.Cycle(struct{}{})
				}

				for _, eu := range m.executeUnits {
					if !eu.isEmpty() {
						isEmpty = false
						resp := eu.Cycle(euReq{fromCycle, app})
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
				}
				m.writeBus.Connect(cycle + 1)
				for _, wu := range m.writeUnits {
					for !wu.isEmpty() || !m.writeBus.IsEmpty() {
						_ = wu.Cycle(wuReq{sequenceID})
					}
				}
				if isEmpty {
					break
				}
			}

			log.Info(m.ctx, "\t️⚠️ Flush to %d", pc/4)
			m.flush(pc)
			cycle += latency.Flush
			log.Info(m.ctx, "\tRegisters: %v", m.ctx.Registers)
			continue
		}

		if m.isEmpty() {
			break
		}
	}

	for {
		cycle++
		empty := true
		for _, cc := range m.cacheControllers {
			if !cc.snoop.IsStart() {
				empty = false
			}
			cc.snoop.Cycle(struct{}{})
		}
		for i, eu := range m.executeUnits {
			if eu.isEmpty() && m.cacheControllers[i].read.IsStart() && m.cacheControllers[i].write.IsStart() {
				continue
			}
			empty = false
			eu.Cycle(euReq{cycle, app})
		}
		if empty {
			break
		}
	}

	for _, cc := range m.cacheControllers {
		cycle += cc.export()
	}

	m.ctx.RATCommit()
	m.ctx.RATFlush()
	log.Info(m.ctx, "Registers: %v", m.ctx.Registers)
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
