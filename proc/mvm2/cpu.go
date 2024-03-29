package mvm2

import "github.com/teivah/majorana/risc"

const (
	cyclesL1Access             = 1
	cyclesMemoryAccess         = 50 + cyclesL1Access
	cyclesRegisterAccess       = 1
	cyclesDecode               = 1
	l1iSize              int32 = 64
)

type CPU struct {
	ctx     *risc.Context
	cycles  int
	li1From int32
	li1To   int32
}

func NewCPU(debug bool, memoryBytes int) *CPU {
	return &CPU{
		ctx:     risc.NewContext(debug, memoryBytes),
		li1From: -1,
		li1To:   -1,
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
			m.cycles += cyclesRegisterAccess
		}
	}
	if m.ctx.Registers[risc.Ra] != 0 {
		m.ctx.Pc = m.ctx.Registers[risc.Ra]
		m.ctx.Registers[risc.Ra] = 0
		goto loop
	}

	return m.cycles, nil
}

func (m *CPU) fetchInstruction() int {
	if m.isPresentInL1i() {
		m.cycles += cyclesL1Access
	} else {
		m.fetchL1i()
	}

	return int(m.ctx.Pc / 4)
}

func (m *CPU) isPresentInL1i() bool {
	return m.ctx.Pc >= m.li1From && m.ctx.Pc <= m.li1To
}

func (m *CPU) fetchL1i() {
	m.cycles += cyclesMemoryAccess
	m.li1From = m.ctx.Pc
	m.li1To = m.ctx.Pc + l1iSize
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
