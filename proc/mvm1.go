package proc

import "github.com/teivah/ettore/risc"

const (
	cyclesMemoryAccess   float32 = 50.
	cyclesRegisterAccess float32 = 1.
	cyclesDecode         float32 = 1.
)

type mvm1 struct {
	ctx    *risc.Context
	cycles float32
}

func newMvm1(memoryBytes int) *mvm1 {
	return &mvm1{
		ctx: risc.NewContext(memoryBytes),
	}
}

func (m *mvm1) run(app risc.Application) (float32, error) {
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
			m.cycles += cyclesRegisterAccess
		}
	}
	return m.cycles, nil
}

func (m *mvm1) fetchInstruction() int {
	m.cycles += cyclesMemoryAccess
	return int(m.ctx.Pc / 4)
}

func (m *mvm1) decode(app risc.Application, i int) risc.InstructionRunner {
	r := app.Instructions[i]
	m.cycles += cyclesDecode
	return r
}

func (m *mvm1) execute(app risc.Application, r risc.InstructionRunner) (risc.Execution, risc.InstructionType, error) {
	exe, err := r.Run(m.ctx, app.Labels)
	if err != nil {
		return risc.Execution{}, 0, err
	}
	m.cycles += risc.CyclesPerInstruction[r.InstructionType()]
	return exe, r.InstructionType(), nil
}
