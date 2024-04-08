package mvp6_0

import (
	"github.com/teivah/majorana/common/log"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type fetchUnit struct {
	pc                 int32
	mmu                *memoryManagementUnit
	remainingCycles    int
	complete           bool
	processing         bool
	pendingMemoryFetch bool
	toCleanPending     bool
	outBus             *comp.BufferedBus[int32]
}

func newFetchUnit(mmu *memoryManagementUnit, outBus *comp.BufferedBus[int32]) *fetchUnit {
	return &fetchUnit{
		mmu:    mmu,
		outBus: outBus,
	}
}

func (u *fetchUnit) reset(pc int32, cleanPending bool) {
	u.complete = false
	u.pc = pc
	u.toCleanPending = cleanPending
}

func (u *fetchUnit) cycle(cycle int, app risc.Application, ctx *risc.Context) {
	if u.toCleanPending {
		// The fetch unit may have sent to the bus wrong instruction, we make sure
		// this is not the case by cleaning it
		log.Infou(ctx, "FU", "cleaning output bus")
		u.outBus.Clean()
		u.toCleanPending = false
	}
	if u.complete {
		return
	}

	if u.pendingMemoryFetch {
		u.remainingCycles--
		if u.remainingCycles == 0 {
			u.pendingMemoryFetch = false
		} else {
			return
		}
	}

	if !u.processing {
		u.processing = true
		if _, exists := u.mmu.getFromL1I([]int32{u.pc}); exists {
			u.remainingCycles = 1
		} else {
			u.remainingCycles = cyclesMemoryAccess
			u.mmu.pushLineToL1I(u.pc, make([]int8, l1ICacheLineSize))
		}
	}

	for i := 0; i < u.outBus.OutLength(); i++ {
		if !u.outBus.CanAdd() {
			log.Infou(ctx, "FU", "can't add")
			return
		}
		u.processing = false
		currentPc := u.pc
		u.pc += 4
		if u.pc/4 >= int32(len(app.Instructions)) {
			u.complete = true
		}
		log.Infou(ctx, "FU", "pushing new element from pc %d", currentPc/4)
		u.outBus.Add(currentPc, cycle)
	}
}

func (u *fetchUnit) flush(pc int32) {
	u.processing = false
	u.complete = false
	u.pc = pc
}

func (u *fetchUnit) isEmpty() bool {
	return u.complete
}
