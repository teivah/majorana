package mvp6_2

import (
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type wuReq struct {
	ctx        *risc.Context
	sequenceID int32
}

type writeUnit struct {
	ctx *risc.Context
	co.Coroutine[wuReq, error]
	memoryWrite risc.ExecutionContext
	inBus       *comp.BufferedBus[risc.ExecutionContext]
}

func newWriteUnit(ctx *risc.Context, inBus *comp.BufferedBus[risc.ExecutionContext]) *writeUnit {
	wu := &writeUnit{
		ctx:   ctx,
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
	if r.sequenceID != -1 && execution.SequenceID > r.sequenceID {
		return nil
	}
	if execution.Execution.RegisterChange {
		r.ctx.TransactionWriteRegister(execution.Execution, execution.SequenceID)
		r.ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		log.Infoi(r.ctx, "WU", execution.InstructionType, execution.SequenceID, "write to register")
	} else if execution.Execution.MemoryChange {
		remainingCycle := cyclesMemoryAccess
		log.Infoi(r.ctx, "WU", execution.InstructionType, execution.SequenceID, "pending memory write")

		u.Checkpoint(func(r wuReq) error {
			if remainingCycle > 0 {
				log.Infoi(r.ctx, "WU", u.memoryWrite.InstructionType, execution.SequenceID, "pending memory write")
				remainingCycle--
				return nil
			}
			u.Reset()
			r.ctx.WriteMemory(u.memoryWrite.Execution)
			r.ctx.DeletePendingRegisters(u.memoryWrite.ReadRegisters, u.memoryWrite.WriteRegisters)
			log.Infoi(r.ctx, "WU", u.memoryWrite.InstructionType, execution.SequenceID, "write to memory")
			return nil
		})

		u.memoryWrite = execution
	} else {
		r.ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		log.Infoi(r.ctx, "WU", execution.InstructionType, execution.SequenceID, "cleaning")
	}
	return nil
}

func (u *writeUnit) commit() {
	u.ctx.Commit()
}

func (u *writeUnit) rollback(sequenceID int32) {
	u.ctx.Rollback(sequenceID)
}

func (u *writeUnit) isEmpty() bool {
	return u.IsStart()
}
