package proc

import "github.com/teivah/majorana/risc"

type virtualMachine interface {
	Run(application risc.Application) (float32, error)
	Context() *risc.Context
}
