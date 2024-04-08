package mvp4

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type fetchUnit struct {
	pc                 int32
	mmu                *memoryManagementUnit
	remainingCycles    int
	complete           bool
	processing         bool
	cyclesMemoryAccess int
}

func newFetchUnit(mmu *memoryManagementUnit, cyclesMemoryAccess int) *fetchUnit {
	return &fetchUnit{
		mmu:                mmu,
		cyclesMemoryAccess: cyclesMemoryAccess,
	}
}

func (fu *fetchUnit) reset(pc int32) {
	fu.complete = false
	fu.pc = pc
}

func (fu *fetchUnit) cycle(app risc.Application, ctx *risc.Context, outBus *comp.SimpleBus[int32]) {
	if fu.complete {
		return
	}

	if !fu.processing {
		fu.processing = true
		if _, exists := fu.mmu.getFromL1I([]int32{fu.pc}); exists {
			fu.remainingCycles = 1
		} else {
			fu.remainingCycles = fu.cyclesMemoryAccess
			fu.mmu.pushLineToL1I(fu.pc, make([]int8, l1ICacheLineSize))
		}
	}

	fu.remainingCycles -= 1.0
	if fu.remainingCycles == 0.0 {
		if !outBus.CanAdd() {
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
		outBus.Add(currentPC)
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
