package mvm4

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct {
	pendingBranchResolution int
}

func (du *decodeUnit) cycle(app risc.Application, inBus *comp.SimpleBus[int32], outBus *comp.SimpleBus[risc.InstructionRunnerPc]) {
	if du.pendingBranchResolution > 0 {
		du.pendingBranchResolution--
		return
	}
	if !outBus.CanAdd() {
		return
	}

	pc, exists := inBus.Get()
	if !exists {
		return
	}
	runner := app.Instructions[pc/4]
	if risc.IsJump(runner.InstructionType()) {
		du.pendingBranchResolution = 1
	}
	outBus.Add(risc.InstructionRunnerPc{
		Runner: runner,
		Pc:     pc,
	})
}

func (du *decodeUnit) flush() {}

func (du *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
