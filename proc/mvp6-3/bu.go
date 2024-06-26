package mvp6_3

import (
	"github.com/teivah/majorana/risc"
)

type btbBranchUnit struct {
	btb         *branchTargetBuffer
	fu          *fetchUnit
	du          *decodeUnit
	cu          *controlUnit
	wu          *writeUnit
	toCheck     bool
	expectation int32
}

func newBTBBranchUnit(btbSize int, fu *fetchUnit, du *decodeUnit, cu *controlUnit, wu *writeUnit) *btbBranchUnit {
	return &btbBranchUnit{
		btb: newBranchTargetBuffer(btbSize),
		fu:  fu,
		du:  du,
		cu:  cu,
		wu:  wu,
	}
}

func (u *btbBranchUnit) assert(runner risc.InstructionRunnerPc) {
	instructionType := runner.Runner.InstructionType()
	if instructionType.IsUnconditionalBranch() {
		nextPc, exists := u.btb.get(runner.Pc)
		if !exists {
			// Unknown branch, it will lead to a pipeline flush
			u.toCheck = true
			u.expectation = -1
		} else {
			// Known branch, no need to check
			u.toCheck = false
			u.fu.reset(nextPc, true)
		}
	} else if instructionType.IsConditionalBranch() {
		// Assuming next instruction
		u.toCheck = true
		u.expectation = runner.Pc + 4
	} else {
		u.toCheck = false
	}
}

func (u *btbBranchUnit) shouldFlushPipeline(pc int32) bool {
	if !u.toCheck {
		return false
	}
	u.toCheck = false

	// If the expectation doesn't correspond to the current pc, we made a wrong
	// assumption; therefore, we should flush
	return u.expectation != pc
}

func (u *btbBranchUnit) notifyConditionalBranchTaken(sequenceID int32) {
	u.cu.notifyConditionalBranch()
	u.wu.rollback(sequenceID)
}

func (u *btbBranchUnit) notifyConditionalBranchNotTaken() {
	u.cu.notifyConditionalBranch()
	u.wu.commit()
}

func (u *btbBranchUnit) notifyUnconditionalJumpAddressResolved(pc, pcTo int32) {
	u.btb.add(pc, pcTo)
	u.fu.reset(pcTo, true)
	u.du.notifyBranchResolved()
}
