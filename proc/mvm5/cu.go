package mvm5

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type controlUnit struct {
	inBus  *comp.BufferedBus[risc.InstructionRunnerPc]
	outBus *comp.BufferedBus[risc.InstructionRunnerPc]
}

func newControlUnit(inBus *comp.BufferedBus[risc.InstructionRunnerPc], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *controlUnit {
	return &controlUnit{
		inBus:  inBus,
		outBus: outBus,
	}
}

func (u *controlUnit) cycle(cycle int, ctx *risc.Context) {
	for u.outBus.CanAdd() {
		runner, exists := u.inBus.Get()
		if !exists {
			return
		}
		u.outBus.Add(runner, cycle)
		if ctx.Debug {
			fmt.Printf("\tCU: Pushing element %v\n", runner.Pc)
		}
	}
}

func (u *controlUnit) flush() {
}

func (u *controlUnit) isEmpty() bool {
	return true
}
