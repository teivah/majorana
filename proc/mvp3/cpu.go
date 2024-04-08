package mvp3

import (
	"fmt"

	"github.com/teivah/majorana/risc"
)

const (
	cyclesL1Access       = 1
	cyclesMemoryAccess   = 50 + cyclesL1Access
	cyclesRegisterAccess = 1
	cyclesDecode         = 1

	bytes            = 1
	kilobytes        = 1024
	l1ICacheLineSize = 64 * bytes
	liICacheSize     = 1 * kilobytes
	l1DCacheLineSize = 64 * bytes
	liDCacheSize     = 1 * kilobytes
)

type CPU struct {
	ctx   *risc.Context
	cycle int
	mmu   *memoryManagementUnit
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	ctx := risc.NewContext(debug, memoryBytes)
	return &CPU{
		ctx: ctx,
		mmu: newMemoryManagementUnit(ctx),
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
	var pc int32
	for pc/4 < int32(len(app.Instructions)) {
		nextPc := m.fetchInstruction(pc)
		r := m.decode(app, nextPc)
		exe, ins, err := m.execute(app, r, pc)
		if err != nil {
			return 0, err
		}
		if exe.Return {
			break
		}
		if exe.PcChange {
			pc = exe.NextPc
		} else {
			pc += 4
		}

		if exe.RegisterChange {
			m.ctx.WriteRegister(exe)
			if m.ctx.Debug {
				fmt.Println(ins, m.ctx.Registers)
			}
			m.cycle += cyclesRegisterAccess
		} else if exe.MemoryChange {
			if m.mmu.doesExecutionMemoryChangesExistsInL1D(exe) {
				m.mmu.writeExecutionMemoryChangesToL1D(exe)
				m.cycle += cyclesL1Access
			} else {
				m.ctx.WriteMemory(exe)
				m.cycle += cyclesMemoryAccess
			}
		}
	}
	//if m.ctx.Registers[risc.Ra] != 0 {
	//	pc = m.ctx.Registers[risc.Ra]
	//	m.ctx.Registers[risc.Ra] = 0
	//	goto loop
	//}
	m.cycle += m.mmu.flush()
	return m.cycle, nil
}

func (m *CPU) Stats() map[string]any {
	return nil
}

func (m *CPU) fetchInstruction(pc int32) int32 {
	if _, exists := m.mmu.getFromL1I([]int32{pc}); exists {
		m.cycle += cyclesL1Access
	} else {
		m.cycle += cyclesMemoryAccess
		m.mmu.pushLineToL1I(pc, make([]int8, l1ICacheLineSize))
	}

	return pc
}

func (m *CPU) decode(app risc.Application, pc int32) risc.InstructionRunner {
	r := app.Instructions[pc/4]
	m.cycle += cyclesDecode
	return r
}

func (m *CPU) execute(app risc.Application, r risc.InstructionRunner, pc int32) (risc.Execution, risc.InstructionType, error) {
	addrs := r.MemoryRead(m.ctx)
	var memory []int8
	if len(addrs) != 0 {
		m.cycle += cyclesL1Access
		if mem, exists := m.mmu.getFromL1D(addrs); exists {
			memory = mem
		} else {
			m.cycle += cyclesMemoryAccess
			line := m.mmu.fetchCacheLine(addrs[0])
			m.mmu.pushLineToL1D(addrs[0], line)
			mem, exists := m.mmu.getFromL1D(addrs)
			if !exists {
				panic("cache line doesn't exist")
			}
			memory = mem
		}
	}

	exe, err := r.Run(m.ctx, app.Labels, pc, memory)
	if err != nil {
		return risc.Execution{}, 0, err
	}
	m.cycle += r.InstructionType().Cycles()
	return exe, r.InstructionType(), nil
}
