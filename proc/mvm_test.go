package proc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teivah/ettore/risc"
	"github.com/teivah/ettore/test"
)

func execute(t *testing.T, vm virtualMachine, instructions string) (float32, error) {
	app, err := risc.Parse(instructions)
	require.NoError(t, err)
	return vm.run(app)
}

func TestMvm1(t *testing.T) {
	vm := newMvm1(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"))
	require.NoError(t, err)
	stats("mvm1 - prime number", cycles)
}
