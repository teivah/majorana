package mvm5

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type writeUnit struct {
	pendingMemoryWrite bool
	cycles             int
	inBus              *comp.BufferedBus[comp.ExecutionContext]
}

func newWriteUnit(inBus *comp.BufferedBus[comp.ExecutionContext]) *writeUnit {
	return &writeUnit{inBus: inBus}
}

func (wu *writeUnit) cycle(ctx *risc.Context) {
	if wu.pendingMemoryWrite {
		wu.cycles--
		if wu.cycles == 0 {
			wu.pendingMemoryWrite = false
		}
		return
	}

	execution, exists := wu.inBus.Get()
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
