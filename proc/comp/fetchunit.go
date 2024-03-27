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

// TODO Ugly
func (fu *FetchUnit) Reset(pc int32) {
	fu.complete = false
	fu.pc = pc
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

type FetchUnitWithBranchPredictor struct {
	pc                 int32
	l1i                L1i
	remainingCycles    float32
	complete           bool
	waiting            bool
	cyclesMemoryAccess float32
	bu                 *BTBBranchUnit
}

func NewFetchUnitWithBranchPredictor(l1iCacheLineSizeInBytes int32, cyclesMemoryAccess float32, bu *BTBBranchUnit) *FetchUnitWithBranchPredictor {
	return &FetchUnitWithBranchPredictor{
		l1i:                NewL1I(l1iCacheLineSizeInBytes),
		cyclesMemoryAccess: cyclesMemoryAccess,
		bu:                 bu,
	}
}

func (fu *FetchUnitWithBranchPredictor) Cycle(currentCycle float32, application risc.Application, outBus Bus[int]) {
	if fu.complete {
		return
	}

	if !fu.waiting {
		fu.waiting = true
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

		fu.waiting = false
		currentPC := fu.pc
		fu.pc += 4
		if fu.pc/4 >= int32(len(application.Instructions)) {
			fu.complete = true
		}
		outBus.Add(int(currentPC/4), currentCycle)
	}
}

func (fu *FetchUnitWithBranchPredictor) Flush(pc int32) {
	fu.waiting = false
	fu.complete = false
	fu.pc = pc
}

func (fu *FetchUnitWithBranchPredictor) IsEmpty() bool {
	return fu.complete
}
