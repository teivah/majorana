package mvp6_1

import (
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type wuReq struct {
	ctx    *risc.Context
	before int32
}

type writeUnit struct {
	co.Coroutine[wuReq, error]
	memoryWrite risc.ExecutionContext
	inBus       *comp.BufferedBus[risc.ExecutionContext]
}

func newWriteUnit(inBus *comp.BufferedBus[risc.ExecutionContext]) *writeUnit {
	wu := &writeUnit{
		inBus: inBus,
	}
	wu.Coroutine = co.New(wu.start)
	return wu
}

func (u *writeUnit) start(r wuReq) error {
	execution, exists := u.inBus.Get()
	if !exists {
		return nil
	}
	if r.before != -1 && execution.Pc > r.before {
		return nil
	}
	if execution.Execution.RegisterChange {
		r.ctx.WriteRegister(execution.Execution)
		r.ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		log.Infoi(r.ctx, "WU", execution.InstructionType, execution.Pc, "write to register")
	} else if execution.Execution.MemoryChange {
		remainingCycle := cyclesMemoryAccess
		log.Infoi(r.ctx, "WU", execution.InstructionType, execution.Pc, "pending memory write")

		u.Checkpoint(func(r wuReq) error {
			if remainingCycle > 0 {
				log.Infoi(r.ctx, "WU", u.memoryWrite.InstructionType, execution.Pc, "pending memory write")
				remainingCycle--
				return nil
			}
			u.Reset()
			r.ctx.WriteMemory(u.memoryWrite.Execution)
			r.ctx.DeletePendingRegisters(u.memoryWrite.ReadRegisters, u.memoryWrite.WriteRegisters)
			log.Infoi(r.ctx, "WU", u.memoryWrite.InstructionType, execution.Pc, "write to memory")
			return nil
		})

		u.memoryWrite = execution
	} else {
		r.ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		log.Infoi(r.ctx, "WU", execution.InstructionType, -1, "cleaning")
	}
	return nil
}

func (u *writeUnit) isEmpty() bool {
	return u.IsStart()
}
