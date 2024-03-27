package comp

import "github.com/teivah/ettore/risc"

type WriteUnit struct{}

func (wu *WriteUnit) Cycle(ctx *risc.Context, writeBus Bus[ExecutionContext]) {
	if !writeBus.IsElementInQueue() {
		return
	}

	execution := writeBus.Get()
	if risc.IsWriteBack(execution.InstructionType) {
		ctx.Write(execution.Execution)
		ctx.DeleteWriteRegisters(execution.WriteRegisters)
	}
}

func (wu *WriteUnit) IsEmpty() bool {
	return true
}
