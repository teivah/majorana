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
	cyclesMemoryAccess int
	toCleanPending     bool
	outBus             *comp.BufferedBus[int32]
}

func newFetchUnit(l1iCacheLineSizeInBytes int32, cyclesMemoryAccess int, outBus *comp.BufferedBus[int32]) *fetchUnit {
	return &fetchUnit{
		l1i:                newL1I(l1iCacheLineSizeInBytes),
		cyclesMemoryAccess: cyclesMemoryAccess,
		outBus:             outBus,
	}
}

func (fu *fetchUnit) reset(pc int32, cleanPending bool) {
	fu.complete = false
	fu.pc = pc
	fu.toCleanPending = cleanPending
}

func (fu *fetchUnit) cycle(cycle int, app risc.Application, ctx *risc.Context) {
	if fu.toCleanPending {
		// The fetch unit may have sent to the bus wrong instruction, we make sure
		// this is not the case by cleaning it
		if ctx.Debug {
			fmt.Printf("\tFU: Cleaning output bus\n")
		}
		fu.outBus.Clean()
		fu.toCleanPending = false
	}
	if fu.complete {
		return
	}

	if fu.pendingMemoryFetch {
		fu.remainingCycles--
		if fu.remainingCycles == 0 {
			fu.pendingMemoryFetch = false
		} else {
			return
		}
	}

	if fu.l1i.present(fu.pc) {
		fu.remainingCycles = 1
	} else {
		fu.pendingMemoryFetch = true
		fu.remainingCycles = fu.cyclesMemoryAccess
		// Should be done after the processing of the 50 cycles
		fu.l1i.fetch(fu.pc)
		return
	}

	for i := 0; i < fu.outBus.OutLength(); i++ {
		if !fu.outBus.CanAdd() {
			return
		}
		fu.processing = false
		currentPc := fu.pc
		fu.pc += 4
		if fu.pc/4 >= int32(len(app.Instructions)) {
			fu.complete = true
		}
		if ctx.Debug {
			fmt.Printf("\tFU: Pushing new element from pc %d\n", currentPc/4)
		}
		fu.outBus.Add(currentPc, cycle)
	}
}

func (fu *fetchUnit) flush(pc int32) {
	fu.processing = false
	fu.complete = false
	fu.pc = pc
}

func (fu *fetchUnit) isEmpty() bool {
	return fu.complete
}
