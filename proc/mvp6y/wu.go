package mvp6y

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type writeUnit struct {
	memoryWrite risc.ExecutionContext
	inBus       *comp.BufferedBus[risc.ExecutionContext]

	// Pending
	coroutine func(ctx *risc.Context)
}

func newWriteUnit(inBus *comp.BufferedBus[risc.ExecutionContext]) *writeUnit {
	return &writeUnit{inBus: inBus}
}

func (u *writeUnit) cycle(ctx *risc.Context) {
	if u.coroutine != nil {
		u.coroutine(ctx)
		return
	}

	execution, exists := u.inBus.Get()
	if !exists {
		return
	}
	if execution.Execution.RegisterChange {
		ctx.WriteRegister(execution.Execution)
		ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		log.Infoi(ctx, "WU", execution.InstructionType, -1, "write to register")
	} else if execution.Execution.MemoryChange {
		remainingCycle := cyclesMemoryAccess
		log.Infoi(ctx, "WU", execution.InstructionType, -1, "pending memory write")

		u.coroutine = func(ctx *risc.Context) {
			if remainingCycle > 0 {
				log.Infoi(ctx, "WU", u.memoryWrite.InstructionType, -1, "pending memory write")
				remainingCycle--
				return
			}
			u.coroutine = nil
			ctx.WriteMemory(u.memoryWrite.Execution)
			ctx.DeletePendingRegisters(u.memoryWrite.ReadRegisters, u.memoryWrite.WriteRegisters)
			log.Infoi(ctx, "WU", u.memoryWrite.InstructionType, -1, "write to memory")
		}

		u.memoryWrite = execution
	} else {
		ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		log.Infoi(ctx, "WU", execution.InstructionType, -1, "cleaning")
	}
}

func (u *writeUnit) isEmpty() bool {
	return u.coroutine == nil
}
