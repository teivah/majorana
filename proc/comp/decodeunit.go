package comp

import "github.com/teivah/ettore/risc"

type DecodeUnit struct{}

func (du *DecodeUnit) Cycle(currentCycle float32, app risc.Application, inBus Bus[int], outBus Bus[risc.InstructionRunner]) {
	if !inBus.IsElementInQueue() || outBus.IsBufferFull() {
		return
	}
	idx := inBus.Get()
	runner := app.Instructions[idx]
	outBus.Add(runner, currentCycle)
}

func (du *DecodeUnit) Flush() {}

func (du *DecodeUnit) IsEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}

type DecodeUnitWithBranchPredictor struct {
	bu                      *BTBBranchUnit
	pendingBranchResolution int
}

func NewDecodeUnitWithBranchPredictor(bu *BTBBranchUnit) *DecodeUnitWithBranchPredictor {
	return &DecodeUnitWithBranchPredictor{bu: bu}
}

func (du *DecodeUnitWithBranchPredictor) Cycle(currentCycle float32, app risc.Application, inBus Bus[int], outBus Bus[risc.InstructionRunner]) {
	if du.pendingBranchResolution > 0 {
		du.pendingBranchResolution--
		return
	}
	inBus.Connect(currentCycle)
	if !inBus.IsElementInQueue() || outBus.IsBufferFull() {
		return
	}

	idx := inBus.Get()
	runner := app.Instructions[idx]
	if risc.IsJump(runner.InstructionType()) {
		du.pendingBranchResolution = 2
	}
	outBus.Add(runner, currentCycle)
}

func (du *DecodeUnitWithBranchPredictor) Flush() {}

func (du *DecodeUnitWithBranchPredictor) IsEmpty() bool {
	// As the decode unit takes only one cycle, it is considered as empty by default
	return true
}
