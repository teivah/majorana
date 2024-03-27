package comp

import "github.com/teivah/ettore/risc"

type BranchUnit struct {
	conditionBranchingExpected *int32
	jumpCondition              bool
}

func (bu *BranchUnit) Assert(ctx *risc.Context, executeBus Bus[risc.InstructionRunner]) {
	if executeBus.IsElementInQueue() {
		runner := executeBus.Peek()
		instructionType := runner.InstructionType()
		if risc.Jump(instructionType) {
			bu.jumpCondition = true
		} else if risc.ConditionalBranching(instructionType) {
			bu.ConditionalBranching(ctx.Pc + 4)
		}
	}
}

func (bu *BranchUnit) Jump() {
	bu.jumpCondition = true
}

func (bu *BranchUnit) ConditionalBranching(expected int32) {
	bu.conditionBranchingExpected = &expected
}

func (bu *BranchUnit) PipelineToBeFlushed(ctx *risc.Context, writeBus Bus[ExecutionContext]) bool {
	if !writeBus.IsElementInBuffer() {
		return false
	}

	conditionalBranching := false
	if bu.conditionBranchingExpected != nil {
		conditionalBranching = *bu.conditionBranchingExpected != ctx.Pc
	}
	assert := conditionalBranching || bu.jumpCondition
	bu.conditionBranchingExpected = nil
	bu.jumpCondition = false
	return assert
}
