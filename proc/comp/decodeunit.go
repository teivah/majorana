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
