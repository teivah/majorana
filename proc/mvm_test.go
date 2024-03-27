package proc

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teivah/ettore/proc/mvm1"
	"github.com/teivah/ettore/proc/mvm2"
	"github.com/teivah/ettore/proc/mvm3"
	"github.com/teivah/ettore/risc"
	"github.com/teivah/ettore/test"
)

func execute(t *testing.T, vm virtualMachine, instructions string) (float32, error) {
	app, err := risc.Parse(instructions)
	require.NoError(t, err)
	cycles, err := vm.Run(app)
	require.NoError(t, err)
	require.Equal(t, int8(1), vm.Context().Memory[4])
	return cycles, nil
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
	vm := mvm1.NewCPU(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"))
	require.NoError(t, err)
	require.Equal(t, float32(147432), cycles)
	stats(cycles)
}

func TestMvm2(t *testing.T) {
	vm := mvm2.NewCPU(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"))
	require.NoError(t, err)
	require.Equal(t, float32(11361), cycles)
	stats(cycles)
}

func TestMvm3(t *testing.T) {
	vm := mvm3.NewCPU(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"))
	require.NoError(t, err)
	require.Equal(t, float32(6918), cycles)
	stats(cycles)
}
