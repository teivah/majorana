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
	cycle   int
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

		if risc.IsWriteBack(ins) {
			m.ctx.Write(exe)
			m.cycle += cyclesRegisterAccess
		}
	}
	if m.ctx.Registers[risc.Ra] != 0 {
		pc = m.ctx.Registers[risc.Ra]
		m.ctx.Registers[risc.Ra] = 0
		goto loop
	}

	return m.cycle, nil
}

func (m *CPU) fetchInstruction(pc int32) int32 {
	if m.isPresentInL1i(pc) {
		m.cycle += cyclesL1Access
	} else {
		m.fetchL1i(pc)
	}

	return pc
}

func (m *CPU) isPresentInL1i(pc int32) bool {
	return pc >= m.li1From && pc <= m.li1To
}

func (m *CPU) fetchL1i(pc int32) {
	m.cycle += cyclesMemoryAccess
	m.li1From = pc
	m.li1To = pc + l1iSize
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
	m.cycle += risc.CyclesPerInstruction(r.InstructionType())
	return exe, r.InstructionType(), nil
}
