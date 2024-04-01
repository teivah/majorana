package mvm3

import (
	"github.com/teivah/majorana/risc"
)

type simpleBranchUnit struct {
	toCheck     bool
	expectation int32
}

func (bu *simpleBranchUnit) assert(runner risc.InstructionRunnerPc) {
	instructionType := runner.Runner.InstructionType()
	if risc.IsJump(instructionType) {
		bu.toCheck = true
		// Not implemented
		bu.expectation = -1
	} else if risc.IsConditionalBranching(instructionType) {
		// Assuming next instruction
		bu.toCheck = true
		bu.expectation = runner.Pc + 4
	}
}

func (bu *simpleBranchUnit) shouldFlushPipeline(pc int32) bool {
	if !bu.toCheck {
		return false
	}
	bu.toCheck = false

	// If the expectation doesn't correspond to the current pc, we made a wrong
	// assumption; therefore, we should flush
	return bu.expectation != pc
}
