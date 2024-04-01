package mvm3

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct{}

func (du *decodeUnit) cycle(app risc.Application, inBus *comp.SimpleBus[int], outBus *comp.SimpleBus[risc.InstructionRunner]) {
	if !outBus.CanAdd() {
		return
	}

	idx, exists := inBus.Get()
	if !exists {
		return
	}
	runner := app.Instructions[idx]
	outBus.Add(runner)
}

func (du *decodeUnit) flush() {}

func (du *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
