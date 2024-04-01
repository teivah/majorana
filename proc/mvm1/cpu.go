package mvm1

import "github.com/teivah/majorana/risc"

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
	for m.ctx.Pc/4 < int32(len(app.Instructions)) {
		idx := m.fetchInstruction()
		r := m.decode(app, idx)
		exe, ins, err := m.execute(app, r)
		if err != nil {
			return 0, err
		}
		m.ctx.Pc = exe.Pc
		if risc.IsWriteBack(ins) {
			m.ctx.Write(exe)
			m.cycle += cyclesRegisterAccess
		}
	}
	if m.ctx.Registers[risc.Ra] != 0 {
		m.ctx.Pc = m.ctx.Registers[risc.Ra]
		m.ctx.Registers[risc.Ra] = 0
		goto loop
	}
	return m.cycle, nil
}

func (m *CPU) fetchInstruction() int {
	m.cycle += cyclesMemoryAccess
	return int(m.ctx.Pc / 4)
}

func (m *CPU) decode(app risc.Application, i int) risc.InstructionRunner {
	r := app.Instructions[i]
	m.cycle += cyclesDecode
	return r
}

func (m *CPU) execute(app risc.Application, r risc.InstructionRunner) (risc.Execution, risc.InstructionType, error) {
	exe, err := r.Run(m.ctx, app.Labels)
	if err != nil {
		return risc.Execution{}, 0, err
	}
	m.cycle += risc.CyclesPerInstruction[r.InstructionType()]
	return exe, r.InstructionType(), nil
}
