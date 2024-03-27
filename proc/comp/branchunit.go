package comp

import (
	"github.com/teivah/ettore/risc"
)

type BranchUnit interface {
	Assert(ctx *risc.Context, executeBus Bus[risc.InstructionRunner])
	ShouldFlushPipeline(ctx *risc.Context, writeBus Bus[ExecutionContext]) bool
}

type SimpleBranchUnit struct {
	conditionBranchingExpected *int32
	isJump                     bool
}

func (bu *SimpleBranchUnit) Assert(ctx *risc.Context, executeBus Bus[risc.InstructionRunner]) {
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

func (bu *SimpleBranchUnit) conditionalBranching(expected int32) {
	bu.conditionBranchingExpected = &expected
}

func (bu *SimpleBranchUnit) ShouldFlushPipeline(ctx *risc.Context, writeBus Bus[ExecutionContext]) bool {
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
	return bu.isJump
}
