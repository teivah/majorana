package mvm4

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct {
	pendingBranchResolution int
}

func (du *decodeUnit) cycle(currentCycle int, app risc.Application, inBus comp.Bus[int], outBus comp.Bus[risc.InstructionRunner]) {
	if du.pendingBranchResolution > 0 {
		du.pendingBranchResolution--
		return
	}
	inBus.Connect(currentCycle)
	if !inBus.IsElementInQueue() || outBus.IsBufferFull() {
		return
	}

	idx := inBus.Get()
	runner := app.Instructions[idx]
	if risc.IsJump(runner.InstructionType()) {
		du.pendingBranchResolution = 2
	}
	outBus.Add(runner, currentCycle)
}

func (du *decodeUnit) flush() {}

func (du *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
