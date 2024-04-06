package mvp6

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type fetchUnit struct {
	pc             int32
	l1i            *comp.LRUCache
	toCleanPending bool
	outBus         *comp.BufferedBus[int32]
	// Pending
	coroutine       func(cycle int, app risc.Application, ctx *risc.Context)
	remainingCycles int
}

func newFetchUnit(outBus *comp.BufferedBus[int32]) *fetchUnit {
	return &fetchUnit{
		l1i:    comp.NewLRUCache(l1ICacheLineSizeInBytes, liICacheSizeInBytes),
		outBus: outBus,
	}
}

func (u *fetchUnit) cycle(cycle int, app risc.Application, ctx *risc.Context) {
	if u.toCleanPending {
		// The fetch unit may have sent to the bus wrong instruction, we make sure
		// this is not the case by cleaning it
		log.Infou(ctx, "FU", "cleaning output bus")
		u.outBus.Clean()
		u.toCleanPending = false
	}
	if u.coroutine != nil {
		u.coroutine(cycle, app, ctx)
		return
	}

	// TODO In the loop
	if _, exists := u.l1i.Get(u.pc); !exists {
		u.remainingCycles = cyclesMemoryAccess - 1
		u.coroutine = func(cycle int, app risc.Application, ctx *risc.Context) {
			if u.remainingCycles != 0 {
				u.remainingCycles--
				return
			}
			u.l1i.PushLine(u.pc, make([]int8, l1ICacheLineSizeInBytes))
			u.coFetch(cycle, app, ctx)
			return
		}
	}
	u.coFetch(cycle, app, ctx)
}

func (u *fetchUnit) coFetch(cycle int, app risc.Application, ctx *risc.Context) {
	u.coroutine = nil
	for i := 0; i < u.outBus.OutLength(); i++ {
		if !u.outBus.CanAdd() {
			log.Infou(ctx, "FU", "can't add")
			return
		}
		currentPc := u.pc
		u.pc += 4
		if u.pc/4 >= int32(len(app.Instructions)) {
			u.coroutine = func(cycle int, app risc.Application, ctx *risc.Context) {}
		}
		log.Infou(ctx, "FU", "pushing new element from pc %d", currentPc/4)
		u.outBus.Add(currentPc, cycle)
	}
}

func (u *fetchUnit) reset(pc int32, cleanPending bool) {
	u.coroutine = nil
	u.pc = pc
	u.toCleanPending = cleanPending
}

func (u *fetchUnit) flush(pc int32) {
	u.coroutine = nil
	u.pc = pc
}

func (u *fetchUnit) isEmpty() bool {
	return u.coroutine != nil
}
