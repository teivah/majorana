package mvm3

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type simpleBranchUnit struct {
	toCheck     bool
	expectation int32
}

func (bu *simpleBranchUnit) assert(ctx *risc.Context, executeBus *comp.SimpleBus[risc.InstructionRunner]) {
	runner, exists := executeBus.Peek()
	if !exists {
		return
	}
	instructionType := runner.InstructionType()
	if risc.IsJump(instructionType) {
		bu.toCheck = true
		// Not implemented
		bu.expectation = -1
	} else if risc.IsConditionalBranching(instructionType) {
		bu.toCheck = true
		// Next instruction
		bu.expectation = ctx.Pc + 4
	}
}

func (bu *simpleBranchUnit) shouldFlushPipeline(ctx *risc.Context) bool {
	if !bu.toCheck {
		return false
	}
	bu.toCheck = false

	// If the expectation doesn't correspond to the current pc, we made a wrong
	// assumption; therefore, we should flush
	return bu.expectation != ctx.Pc
}
