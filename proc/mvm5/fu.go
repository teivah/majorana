package mvm5

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type fetchUnit struct {
	pc                 int32
	l1i                l1i
	remainingCycles    int
	complete           bool
	processing         bool
	pendingMemoryFetch bool
	toCleanPending     bool
	outBus             *comp.BufferedBus[int32]
}

func newFetchUnit(l1iCacheLineSizeInBytes int32, outBus *comp.BufferedBus[int32]) *fetchUnit {
	return &fetchUnit{
		l1i:    newL1I(l1iCacheLineSizeInBytes),
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
		if ctx.Debug {
			fmt.Printf("\tFU: Cleaning output bus\n")
		}
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

	if u.l1i.present(u.pc) {
		u.remainingCycles = 1
	} else {
		u.pendingMemoryFetch = true
		u.remainingCycles = cyclesMemoryAccess
		// Should be done after the processing of the 50 cycles
		u.l1i.fetch(u.pc)
		return
	}

	for i := 0; i < u.outBus.OutLength(); i++ {
		if !u.outBus.CanAdd() {
			return
		}
		u.processing = false
		currentPc := u.pc
		u.pc += 4
		if u.pc/4 >= int32(len(app.Instructions)) {
			u.complete = true
		}
		if ctx.Debug {
			fmt.Printf("\tFU: Pushing new element from pc %d\n", currentPc/4)
		}
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
