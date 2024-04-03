package mvp1

import (
	"fmt"

	"github.com/teivah/majorana/risc"
)

const (
	cyclesMemoryAccess   = 50
	cyclesRegisterAccess = 1
	cyclesDecode         = 1
)

type CPU struct {
	ctx   *risc.Context
	cycle int
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	return &CPU{
		ctx: risc.NewContext(debug, memoryBytes),
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
		exe, ins, err := m.execute(app, r, nextPc)
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
			m.cycle += cyclesRegisterAccess
		} else if exe.MemoryChange {
			m.ctx.WriteMemory(exe)
			m.cycle += cyclesMemoryAccess
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
	m.cycle += cyclesMemoryAccess
	return pc
}

func (m *CPU) decode(app risc.Application, pc int32) risc.InstructionRunner {
	r := app.Instructions[pc/4]
	m.cycle += cyclesDecode
	return r
}

func (m *CPU) execute(app risc.Application, r risc.InstructionRunner, pc int32) (risc.Execution, risc.InstructionType, error) {
	exe, err := r.Run(m.ctx, app.Labels, pc)
	if err != nil {
		return risc.Execution{}, 0, err
	}
	m.cycle += r.InstructionType().Cycles()
	return exe, r.InstructionType(), nil
}
