package mvm1

import "github.com/teivah/majorana/risc"

const (
	cyclesMemoryAccess   float32 = 50.
	cyclesRegisterAccess float32 = 1.
	cyclesDecode         float32 = 1.
)

type CPU struct {
	ctx    *risc.Context
	cycles float32
}

func NewCPU(memoryBytes int) *CPU {
	return &CPU{
		ctx: risc.NewContext(memoryBytes),
	}
}

func (m *CPU) Context() *risc.Context {
	return m.ctx
}

func (m *CPU) Run(app risc.Application) (float32, error) {
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
			m.cycles += cyclesRegisterAccess
		}
	}
	return m.cycles, nil
}

func (m *CPU) fetchInstruction() int {
	m.cycles += cyclesMemoryAccess
	return int(m.ctx.Pc / 4)
}

func (m *CPU) decode(app risc.Application, i int) risc.InstructionRunner {
	r := app.Instructions[i]
	m.cycles += cyclesDecode
	return r
}

func (m *CPU) execute(app risc.Application, r risc.InstructionRunner) (risc.Execution, risc.InstructionType, error) {
	exe, err := r.Run(m.ctx, app.Labels)
	if err != nil {
		return risc.Execution{}, 0, err
	}
	m.cycles += risc.CyclesPerInstruction[r.InstructionType()]
	return exe, r.InstructionType(), nil
}
