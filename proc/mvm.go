package proc

import "github.com/teivah/ettore/risc"

type virtualMachine interface {
	run(application risc.Application) (float32, error)
	context() *risc.Context
}
