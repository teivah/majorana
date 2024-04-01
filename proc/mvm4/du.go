package mvm4

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct {
	pendingBranchResolution int
}

func (du *decodeUnit) cycle(currentCycle int, app risc.Application, inBus *comp.SimpleBus[int], outBus *comp.SimpleBus[risc.InstructionRunner]) {
	if du.pendingBranchResolution > 0 {
		du.pendingBranchResolution--
		return
	}
	if !outBus.CanAdd() {
		return
	}

	idx, exists := inBus.Get()
	if !exists {
		return
	}
	runner := app.Instructions[idx]
	if risc.IsJump(runner.InstructionType()) {
		du.pendingBranchResolution = 2
	}
	outBus.Add(runner)
}

func (du *decodeUnit) flush() {}

func (du *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
