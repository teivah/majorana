package mvm4

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type writeUnit struct {
	pendingMemoryWrite bool
	cycles             int
}

func (wu *writeUnit) cycle(ctx *risc.Context, inBus *comp.SimpleBus[comp.ExecutionContext]) {
	if wu.pendingMemoryWrite {
		wu.cycles--
		if wu.cycles == 0 {
			wu.pendingMemoryWrite = false
		}
		return
	}

	execution, exists := inBus.Get()
	if !exists {
		return
	}
	if risc.IsWriteBack(execution.InstructionType) {
		ctx.Write(execution.Execution)
		ctx.DeleteWriteRegisters(execution.WriteRegisters)
	} else {
		wu.pendingMemoryWrite = true
		wu.cycles = cyclesMemoryAccess
	}
}

func (wu *writeUnit) isEmpty() bool {
	return true
}
