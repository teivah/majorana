package proc

import "github.com/teivah/majorana/risc"

type virtualMachine interface {
	Run(application risc.Application) (int, error)
	Context() *risc.Context
	Stats() map[string]any
}
