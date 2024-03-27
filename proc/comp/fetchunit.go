package comp

import "github.com/teivah/ettore/risc"

type FetchUnit struct {
	pc                 int32
	l1i                L1i
	remainingCycles    float32
	complete           bool
	processing         bool
	cyclesMemoryAccess float32
}

func NewFetchUnit(l1iCacheLineSizeInBytes int32, cyclesMemoryAccess float32) *FetchUnit {
	return &FetchUnit{
		l1i:                NewL1I(l1iCacheLineSizeInBytes),
		cyclesMemoryAccess: cyclesMemoryAccess,
	}
}

func (fu *FetchUnit) Cycle(currentCycle float32, application risc.Application, outBus Bus[int]) {
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
		if fu.pc/4 >= int32(len(application.Instructions)) {
			fu.complete = true
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
