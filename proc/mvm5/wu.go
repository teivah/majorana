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

func (u *writeUnit) cycle(ctx *risc.Context) {
	if u.pendingMemoryWrite {
		u.cycles--
		if u.cycles == 0 {
			u.pendingMemoryWrite = false
		}
		return
	}

	execution, exists := u.inBus.Get()
	if !exists {
		return
	}
	if risc.IsWriteBack(execution.InstructionType) {
		ctx.Write(execution.Execution)
		ctx.DeleteWriteRegisters(execution.WriteRegisters)
	} else {
		u.pendingMemoryWrite = true
		u.cycles = cyclesMemoryAccess
	}
}

func (u *writeUnit) isEmpty() bool {
	return true
}
