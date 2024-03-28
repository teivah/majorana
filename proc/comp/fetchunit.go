package comp

import (
	"fmt"

	"github.com/teivah/ettore/risc"
)

type FetchUnit struct {
	pc                 int32
	l1i                L1i
	remainingCycles    float32
	complete           bool
	processing         bool
	cyclesMemoryAccess float32
	reset              bool
}

func NewFetchUnit(l1iCacheLineSizeInBytes int32, cyclesMemoryAccess float32) *FetchUnit {
	return &FetchUnit{
		l1i:                NewL1I(l1iCacheLineSizeInBytes),
		cyclesMemoryAccess: cyclesMemoryAccess,
	}
}

// TODO Ugly
func (fu *FetchUnit) Reset(pc int32) {
	fu.complete = false
	fu.pc = pc
	fu.reset = true
}

func (fu *FetchUnit) Cycle(currentCycle float32, app risc.Application, outBus Bus[int]) {
	// TODO Explain better why
	// In case of a reset, we need to delete the last element in the bus
	if fu.reset {
		if app.Debug {
			fmt.Printf("\tFU: Delete latest element from the queue\n")
		}
		outBus.DeleteLast()
		fu.reset = false
	}
	if fu.complete {
		return
	}

	if !fu.processing {
		fu.processing = true
		if fu.l1i.Present(fu.pc) {
			fu.remainingCycles = 1
		} else {
			fu.remainingCycles = fu.cyclesMemoryAccess
			// Should be done after the processing of the 50 cycles
			fu.l1i.Fetch(fu.pc)
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
		if app.Debug {
			fmt.Printf("\tFU: Pushing new element from pc %d\n", currentPC/4)
		}
		outBus.Add(int(currentPC/4), currentCycle)
	}
}

func (fu *FetchUnit) Flush(pc int32) {
	fu.processing = false
	fu.complete = false
	fu.pc = pc
}

func (fu *FetchUnit) IsEmpty() bool {
	return fu.complete
}
