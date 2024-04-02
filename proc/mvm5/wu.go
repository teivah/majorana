package mvm5

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type writeUnit struct {
	pendingMemoryWrite bool
	memoryWrite        comp.ExecutionContext
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
			ctx.WriteMemory(u.memoryWrite.Execution)
			ctx.DeletePendingRegisters(u.memoryWrite.ReadRegisters, u.memoryWrite.WriteRegisters)
			if risc.IsBranch(u.memoryWrite.InstructionType) {
				ctx.DeletePendingBranch()
			}
			logi(ctx, "WU", u.memoryWrite.InstructionType, -1, "write to memory")
		}
		return
	}

	execution, exists := u.inBus.Get()
	if !exists {
		return
	}
	if execution.Execution.RegisterChange {
		ctx.WriteRegister(execution.Execution)
		ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		if risc.IsBranch(execution.InstructionType) {
			ctx.DeletePendingBranch()
		}
		logi(ctx, "WU", execution.InstructionType, -1, "write to register")
	} else if execution.Execution.MemoryChange {
		u.pendingMemoryWrite = true
		u.cycles = cyclesMemoryAccess
		u.memoryWrite = execution
	} else {
		ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		if risc.IsBranch(execution.InstructionType) {
			ctx.DeletePendingBranch()
		}
		logi(ctx, "WU", execution.InstructionType, -1, "cleaning")
	}
}

func (u *writeUnit) isEmpty() bool {
	return !u.pendingMemoryWrite
}
