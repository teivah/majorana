package mvp6_4

import (
	co "github.com/teivah/majorana/common/coroutine"
	"github.com/teivah/majorana/common/latency"
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type fuReq struct {
	cycle int
	app   risc.Application
}

type fetchUnit struct {
	ctx *risc.Context
	co.Coroutine[fuReq, error]
	pc              int32
	toCleanPending  bool
	outBus          *comp.BufferedBus[int32]
	complete        bool
	mmu             *memoryManagementUnit
	remainingCycles int
	l1i             *comp.LRUCache
}

func newFetchUnit(ctx *risc.Context, outBus *comp.BufferedBus[int32]) *fetchUnit {
	fu := &fetchUnit{
		ctx:    ctx,
		outBus: outBus,
		l1i:    comp.NewLRUCache(l1ICacheLineSize, l1ICacheSize),
	}
	fu.Coroutine = co.New(fu.start)
	fu.Coroutine.Pre(func(r fuReq) bool {
		if fu.toCleanPending {
			// The fetch unit may have sent to the bus wrong instruction, we make sure
			// this is not the case by cleaning it
			log.Infou(ctx, "FU", "cleaning output bus")
			fu.outBus.Clean()
			fu.toCleanPending = false
		}
		return false
	})
	return fu
}

func (u *fetchUnit) start(r fuReq) error {
	for i := 0; i < u.outBus.OutLength(); i++ {
		if !u.outBus.CanAdd() {
			log.Infou(u.ctx, "FU", "can't add")
			return nil
		}

		if _, exists := u.getFromL1I([]int32{u.pc}); !exists {
			u.remainingCycles = latency.MemoryAccess - 1
			u.Checkpoint(u.memoryAccess)
			return nil
		}

		currentPc := u.pc
		u.pc += 4
		if u.pc/4 >= int32(len(r.app.Instructions)) {
			u.Checkpoint(func(fuReq) error { return nil })
			u.complete = true
		}
		log.Infou(u.ctx, "FU", "pushing new element from pc %d", currentPc/4)
		u.outBus.Add(currentPc, r.cycle)
	}
	return nil
}

func (u *fetchUnit) memoryAccess(r fuReq) error {
	if u.remainingCycles != 0 {
		log.Infou(u.ctx, "FU", "pending memory access")
		u.remainingCycles--
		return nil
	}
	u.Reset()
	u.pushLineToL1I(u.pc, make([]int8, l1ICacheLineSize))

	currentPc := u.pc
	u.pc += 4
	if u.pc/4 >= int32(len(r.app.Instructions)) {
		u.Checkpoint(func(fuReq) error { return nil })
		u.complete = true
	}
	log.Infou(u.ctx, "FU", "pushing new element from pc %d", currentPc/4)
	u.outBus.Add(currentPc, r.cycle)
	return nil

}

func (u *fetchUnit) reset(pc int32, cleanPending bool) {
	u.ctx.IncSequenceID()
	u.Reset()
	u.pc = pc
	u.toCleanPending = cleanPending
}

func (u *fetchUnit) flush(pc int32) {
	u.ctx.IncSequenceID()
	u.Reset()
	u.complete = false
	u.pc = pc
}

func (u *fetchUnit) isEmpty() bool {
	return u.complete
}

func (u *fetchUnit) getFromL1I(addrs []int32) ([]int8, bool) {
	memory := make([]int8, 0, len(addrs))
	for _, addr := range addrs {
		v, exists := u.l1i.Get(addr)
		if !exists {
			return nil, false
		}
		memory = append(memory, v)
	}
	return memory, true
}

func (u *fetchUnit) pushLineToL1I(addr int32, line []int8) {
	u.l1i.PushLine(addr, line)
}
