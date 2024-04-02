package mvm5

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type writeUnit struct {
	pendingMemoryWrite bool
	memoryWrite        risc.Execution
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
			ctx.WriteMemory(u.memoryWrite)
		}
		return
	}

	execution, exists := u.inBus.Get()
	if !exists {
		return
	}
	if execution.Execution.RegisterChange {
		ctx.WriteRegister(execution.Execution)
		ctx.DeleteWriteRegisters(execution.WriteRegisters)
	} else if execution.Execution.MemoryChange {
		// TODO Do after
		u.pendingMemoryWrite = true
		u.cycles = cyclesMemoryAccess
		u.memoryWrite = execution.Execution
	}
}

func (u *writeUnit) isEmpty() bool {
	return true
}
