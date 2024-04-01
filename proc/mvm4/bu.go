package mvm4

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type btbBranchUnit struct {
	btb         *branchTargetBuffer
	fu          *fetchUnit
	toCheck     bool
	expectation int32
}

func newBTBBranchUnit(btbSize int, fu *fetchUnit) *btbBranchUnit {
	return &btbBranchUnit{
		btb: newBranchTargetBuffer(btbSize),
		fu:  fu,
	}
}

func (bu *btbBranchUnit) assert(ctx *risc.Context, executeBus *comp.SimpleBus[risc.InstructionRunnerPc]) {
	runner, exists := executeBus.Peek()
	if !exists {
		return
	}
	instructionType := runner.Runner.InstructionType()
	if risc.IsJump(instructionType) {
		nextPc, exists := bu.btb.get(runner.Pc)
		if !exists {
			// Unknown branch, it will lead to a pipeline flush
			bu.toCheck = true
			bu.expectation = -1
		} else {
			//Known branch, no need to check
			bu.fu.reset(nextPc, true)
		}
	} else if risc.IsConditionalBranching(instructionType) {
		bu.toCheck = true
		// Next instruction
		bu.expectation = runner.Pc
	}
}

func (bu *btbBranchUnit) shouldFlushPipeline(pc int32) bool {
	if !bu.toCheck {
		return false
	}
	bu.toCheck = false

	// If the expectation doesn't correspond to the current pc, we made a wrong
	// assumption; therefore, we should flush
	return bu.expectation != pc
}

func (bu *btbBranchUnit) branchNotify(pc, pcTo int32) {
	bu.btb.add(pc, pcTo)
	bu.fu.reset(pcTo, true)
}
