package proc

import "github.com/teivah/ettore/risc"

const (
	mvm2CyclesL1Access       float32 = 1.
	mvm2CyclesMemoryAccess   float32 = 50. + mvm2CyclesL1Access
	mvm2CyclesRegisterAccess float32 = 1.
	mvm2CyclesDecode         float32 = 1.
	mvm2L1iSize              int32   = 64
)

type mvm2 struct {
	ctx     *risc.Context
	cycles  float32
	li1From int32
	li1To   int32
}

func newMvm2(memoryBytes int) *mvm2 {
	return &mvm2{
		ctx:     risc.NewContext(memoryBytes),
		li1From: -1,
		li1To:   -1,
	}
}

func (m *mvm2) run(app risc.Application) (float32, error) {
	for m.ctx.Pc/4 < int32(len(app.Instructions)) {
		idx := m.fetchInstruction()
		r := m.decode(app, idx)
		exe, ins, err := m.execute(app, r)
		if err != nil {
			return 0, err
		}
		m.ctx.Pc = exe.Pc
		if risc.WriteBack(ins) {
			m.ctx.Write(exe)
			m.cycles += mvm2CyclesRegisterAccess
		}
	}
	return m.cycles, nil
}

func (m *mvm2) fetchInstruction() int {
	if m.isPresentInL1i() {
		m.cycles += mvm2CyclesL1Access
	} else {
		m.fetchL1i()
	}

	return int(m.ctx.Pc / 4)
}

func (m *mvm2) isPresentInL1i() bool {
	return m.ctx.Pc >= m.li1From && m.ctx.Pc <= m.li1To
}

func (m *mvm2) fetchL1i() {
	m.cycles += mvm2CyclesMemoryAccess
	m.li1From = m.ctx.Pc
	m.li1To = m.ctx.Pc + mvm2L1iSize
}

func (m *mvm2) decode(app risc.Application, i int) risc.InstructionRunner {
	r := app.Instructions[i]
	m.cycles += mvm2CyclesDecode
	return r
}

func (m *mvm2) execute(app risc.Application, r risc.InstructionRunner) (risc.Execution, risc.InstructionType, error) {
	exe, err := r.Run(m.ctx, app.Labels)
	if err != nil {
		return risc.Execution{}, 0, err
	}
	m.cycles += risc.CyclesPerInstruction[r.InstructionType()]
	return exe, r.InstructionType(), nil
}
