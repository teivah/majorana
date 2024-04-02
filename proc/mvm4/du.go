package mvm4

import (
	"fmt"

	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type decodeUnit struct {
	pendingBranchResolution bool
}

func (du *decodeUnit) cycle(app risc.Application, ctx *risc.Context, inBus *comp.SimpleBus[int32], outBus *comp.SimpleBus[risc.InstructionRunnerPc]) {
	if du.pendingBranchResolution {
		return
	}
	if !outBus.CanAdd() {
		return
	}

	pc, exists := inBus.Get()
	if !exists {
		return
	}
	if ctx.Debug {
		fmt.Printf("\tDU: Decoding instruction %d\n", pc/4)
	}
	runner := app.Instructions[pc/4]
	if risc.IsUnconditionalBranch(runner.InstructionType()) {
		du.pendingBranchResolution = true
	}
	outBus.Add(risc.InstructionRunnerPc{
		Runner: runner,
		Pc:     pc,
	})
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
