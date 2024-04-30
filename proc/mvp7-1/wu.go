package mvp7_1

import (
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type wuReq struct {
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
		u.ctx.TransactionRATWrite(execution.Execution, execution.SequenceID)
		u.ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		log.Infoi(u.ctx, "WU", execution.InstructionType, execution.SequenceID, "write to register")
	} else if execution.Execution.MemoryChange {
		panic("From MVP 6.4, memory changes are written via L1 cache eviction solely")
	} else {
		u.ctx.DeletePendingRegisters(execution.ReadRegisters, execution.WriteRegisters)
		log.Infoi(u.ctx, "WU", execution.InstructionType, execution.SequenceID, "cleaning")
	}
	return nil
}

func (u *writeUnit) isEmpty() bool {
	return u.IsStart()
}
