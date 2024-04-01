package mvm3

import (
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

type simpleBranchUnit struct {
	toCheck     bool
	expectation int32
}

func (bu *simpleBranchUnit) assert(ctx *risc.Context, executeBus comp.Bus[risc.InstructionRunner]) {
	if !executeBus.IsElementInQueue() {
		return
	}
	runner := executeBus.Peek()
	instructionType := runner.InstructionType()
	if risc.IsJump(instructionType) {
		bu.toCheck = true
		// Not implemented
		bu.expectation = -1
	} else if risc.IsConditionalBranching(instructionType) {
		bu.toCheck = true
		bu.expectation = ctx.Pc + 4 // Next instruction
	}
}

func (bu *simpleBranchUnit) shouldFlushPipeline(ctx *risc.Context) bool {
	if !bu.toCheck {
		return false
	}

	defer func() {
		bu.toCheck = false
		bu.expectation = 0
	}()

	// If the expectation doesn't correspond to the current pc, we made a wrong
	// assumption; therefore, we should flush
	return bu.expectation != ctx.Pc
}
