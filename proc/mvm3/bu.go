package mvm3

import (
	"github.com/teivah/ettore/proc/comp"
	"github.com/teivah/ettore/risc"
)

type simpleBranchUnit struct {
	conditionBranchingExpected *int32
	isJump                     bool
}

func (bu *simpleBranchUnit) assert(ctx *risc.Context, executeBus comp.Bus[risc.InstructionRunner]) {
	if !executeBus.IsElementInQueue() {
		return
	}
	runner := executeBus.Peek()
	instructionType := runner.InstructionType()
	if risc.IsJump(instructionType) {
		bu.isJump = true
	} else if risc.IsConditionalBranching(instructionType) {
		// Move to the next instruction
		bu.conditionalBranching(ctx.Pc + 4)
	}
}

func (bu *simpleBranchUnit) conditionalBranching(expected int32) {
	bu.conditionBranchingExpected = &expected
}

func (bu *simpleBranchUnit) shouldFlushPipeline(ctx *risc.Context, writeBus comp.Bus[comp.ExecutionContext]) bool {
	if !writeBus.IsElementInBuffer() {
		return false
	}

	defer func() {
		bu.conditionBranchingExpected = nil
		bu.isJump = false
	}()

	if bu.conditionBranchingExpected != nil {
		return *bu.conditionBranchingExpected != ctx.Pc
	}
	// In case of a non-conditional jump, we need to flush the pipeline as the CPU
	// already fetches the next instructions, assuming sequential execution
	return bu.isJump
}
