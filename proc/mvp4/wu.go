package mvp4

import (
	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type writeUnit struct {
	pendingMemoryWrite bool
	cycles             int
}

func (wu *writeUnit) cycle(ctx *risc.Context, inBus *comp.SimpleBus[risc.ExecutionContext]) {
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
	if execution.Execution.RegisterChange {
		ctx.WriteRegister(execution.Execution)
		ctx.DeletePendingWriteRegisters(execution.WriteRegisters)
	} else if execution.Execution.MemoryChange {
		// TODO Do after
		wu.pendingMemoryWrite = true
		wu.cycles = latency.MemoryAccess
		ctx.WriteMemory(execution.Execution)
	}
}

func (wu *writeUnit) isEmpty() bool {
	return !wu.pendingMemoryWrite
}
