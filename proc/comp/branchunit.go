package comp

import (
	"github.com/teivah/ettore/risc"
)

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
	// In case of a non-conditional jump, we need to flush the pipeline as the CPU
	// already fetches the next instructions, assuming sequential execution
	return bu.isJump
}

type BTBBranchUnit struct {
	SimpleBranchUnit
	btb *BranchTargetBuffer
}

func NewBTBBranchUnit(btbSize int) *BTBBranchUnit {
	return &BTBBranchUnit{
		btb: NewBranchTargetBuffer(btbSize),
	}
}

func (bu *BTBBranchUnit) BranchNotify(pc, pcTo int32) {
	bu.btb.Add(pc, pcTo)
}
