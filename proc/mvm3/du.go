package mvm3

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct{}

func (du *decodeUnit) cycle(currentCycle float32, app risc.Application, inBus comp.Bus[int], outBus comp.Bus[risc.InstructionRunner]) {
	if !inBus.IsElementInQueue() || outBus.IsBufferFull() {
		return
	}
	idx := inBus.Get()
	runner := app.Instructions[idx]
	outBus.Add(runner, currentCycle)
}

func (du *decodeUnit) flush() {}

func (du *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
