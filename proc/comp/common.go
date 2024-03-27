package comp

import "github.com/teivah/ettore/risc"

type ExecutionContext struct {
	Execution       risc.Execution
	InstructionType risc.InstructionType
	WriteRegisters  []risc.RegisterType
}
