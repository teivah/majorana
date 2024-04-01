package mvm5

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct {
	pendingBranchResolution bool
	inBus                   *comp.BufferedBus[int32]
	outBus                  *comp.BufferedBus[risc.InstructionRunnerPc]
}

func newDecodeUnit(inBus *comp.BufferedBus[int32], outBus *comp.BufferedBus[risc.InstructionRunnerPc]) *decodeUnit {
	return &decodeUnit{inBus: inBus, outBus: outBus}
}

func (du *decodeUnit) cycle(cycle int, app risc.Application, ctx *risc.Context) {
	if du.pendingBranchResolution {
		return
	}

	for i := 0; i < du.outBus.InLength(); i++ {
		if !du.outBus.CanAdd() {
			return
		}
		pc, exists := du.inBus.Get()
		if !exists {
			return
		}
		if ctx.Debug {
			fmt.Printf("\tDU: Decoding instruction %d\n", pc/4)
		}
		runner := app.Instructions[pc/4]
		if risc.IsJump(runner.InstructionType()) {
			du.pendingBranchResolution = true
		}
		du.outBus.Add(risc.InstructionRunnerPc{
			Runner: runner,
			Pc:     pc,
		}, cycle)
	}
}

func (du *decodeUnit) notifyBranchResolved() {
	du.pendingBranchResolution = false
}

func (du *decodeUnit) flush() {
	du.pendingBranchResolution = false
}

func (du *decodeUnit) isEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
