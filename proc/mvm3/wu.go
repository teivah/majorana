package mvm3

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type writeUnit struct{}

func (wu *writeUnit) cycle(ctx *risc.Context, inBus *comp.SimpleBus[comp.ExecutionContext]) {
	execution, exists := inBus.Get()
	if !exists {
		return
	}
	// TODO If write to memory +50
	if risc.IsWriteBack(execution.InstructionType) {
		ctx.Write(execution.Execution)
		ctx.DeleteWriteRegisters(execution.WriteRegisters)
	}
}

func (wu *writeUnit) isEmpty() bool {
	return true
}
