package mvm3

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
	cyclesMemoryAccess int
	toReset            bool
}

func newFetchUnit(l1iCacheLineSizeInBytes int32, cyclesMemoryAccess int) *fetchUnit {
	return &fetchUnit{
		l1i:                newL1I(l1iCacheLineSizeInBytes),
		cyclesMemoryAccess: cyclesMemoryAccess,
	}
}

func (fu *fetchUnit) reset(pc int32) {
	fu.complete = false
	fu.pc = pc
	fu.toReset = true
}

func (fu *fetchUnit) cycle(currentCycle int, app risc.Application, ctx *risc.Context, outBus comp.Bus[int]) {
	if fu.complete {
		return
	}

	if !fu.processing {
		fu.processing = true
		if fu.l1i.present(fu.pc) {
			fu.remainingCycles = 1
		} else {
			fu.remainingCycles = fu.cyclesMemoryAccess
			// Should be done after the processing of the 50 cycles
			fu.l1i.fetch(fu.pc)
		}
	}

	fu.remainingCycles -= 1.0
	if fu.remainingCycles == 0.0 {
		if outBus.IsBufferFull() {
			fu.remainingCycles = 1.0
			return
		}

		fu.processing = false
		currentPC := fu.pc
		fu.pc += 4
		if fu.pc/4 >= int32(len(app.Instructions)) {
			fu.complete = true
		}
		if ctx.Debug {
			fmt.Printf("\tFU: Pushing new element from pc %d\n", currentPC/4)
		}
		outBus.Add(int(currentPC/4), currentCycle)
	}
}

func (fu *fetchUnit) Flush(pc int32) {
	fu.processing = false
	fu.complete = false
	fu.pc = pc
}

func (fu *fetchUnit) IsEmpty() bool {
	return fu.complete
}
