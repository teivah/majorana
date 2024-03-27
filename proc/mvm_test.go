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

func TestPrimeNumber(t *testing.T) {
	bits := risc.BytesFromLowBits(1109)
	test.RunAssert(t, map[risc.RegisterType]int32{}, 5,
		map[int]int8{0: bits[0], 1: bits[1], 2: bits[2], 3: bits[3]},
		test.ReadFile(t, "../res/prime-number.asm"),
		map[risc.RegisterType]int32{risc.A0: 4},
		map[int]int8{4: 1})
}

func TestMvm1(t *testing.T) {
	vm := newMvm1(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"))
	require.NoError(t, err)
	stats("mvm1 - prime number", cycles)
}

func TestMvm2(t *testing.T) {
	vm := newMvm2(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"))
	require.NoError(t, err)
	stats("mvm1 - prime number", cycles)
}

func TestMvm3(t *testing.T) {
	vm := newMvm3(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"))
	require.NoError(t, err)
	stats("mvm1 - prime number", cycles)
}
