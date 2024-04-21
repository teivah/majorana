package mvp2

import (
	"fmt"

	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/risc"
)

const (
	cyclesDecode       = 1
	l1iSize      int32 = 64
)

type CPU struct {
	ctx     *risc.Context
	cycle   int
	l1iFrom int32
	l1iTo   int32
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	return &CPU{
		ctx:     risc.NewContext(debug, memoryBytes, false),
		l1iFrom: -1,
		l1iTo:   -1,
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (int, error) {
loop:
	var pc int32
	for pc/4 < int32(len(app.Instructions)) {
		nextPc := m.fetchInstruction(pc)
		r := m.decode(app, nextPc)
		exe, ins, err := m.execute(app, r, pc)
		if err != nil {
			return 0, err
		}
		if exe.Return {
			return m.cycle, nil
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
			m.cycle += latency.RegisterAccess
		} else if exe.MemoryChange {
			m.ctx.WriteMemory(exe)
			m.cycle += latency.MemoryAccess
		}
	}
	if m.ctx.Registers[risc.Ra] != 0 {
		pc = m.ctx.Registers[risc.Ra]
		m.ctx.Registers[risc.Ra] = 0
		goto loop
	}

	return m.cycle, nil
}

func (m *CPU) Stats() map[string]any {
	return nil
}

func (m *CPU) fetchInstruction(pc int32) int32 {
	if m.isPresentInL1i(pc) {
		m.cycle += latency.L1Access
	} else {
		m.fetchL1i(pc)
	}

	return pc
}

func (m *CPU) isPresentInL1i(pc int32) bool {
	return pc >= m.l1iFrom && pc <= m.l1iTo
}

func (m *CPU) fetchL1i(pc int32) {
	m.cycle += latency.MemoryAccess
	m.l1iFrom = pc
	m.l1iTo = pc + l1iSize
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
		for _, addr := range addrs {
			memory = append(memory, m.ctx.Memory[addr])
		}
		m.cycle += latency.MemoryAccess
	}

	exe, err := r.Run(m.ctx, app.Labels, pc, memory)
	if err != nil {
		return risc.Execution{}, 0, err
	}
	m.cycle += r.InstructionType().Cycles()
	return exe, r.InstructionType(), nil
}
