package proc

import "github.com/teivah/ettore/risc"

type virtualMachine interface {
	Run(application risc.Application) (float32, error)
	Context() *risc.Context
}
