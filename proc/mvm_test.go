package proc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teivah/majorana/proc/mvm1"
	"github.com/teivah/majorana/proc/mvm2"
	"github.com/teivah/majorana/proc/mvm3"
	"github.com/teivah/majorana/proc/mvm4"
	"github.com/teivah/majorana/risc"
	"github.com/teivah/majorana/test"
)

func execute(t *testing.T, vm virtualMachine, instructions string, debug bool) (float32, error) {
	app, err := risc.Parse(instructions, debug)
	require.NoError(t, err)
	cycles, err := vm.Run(app)
	require.NoError(t, err)
	return cycles, nil
}

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}

	if n <= 3 {
		return true
	}

	if n%2 == 0 || n%3 == 0 {
		return false
	}

	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}

	return true
}

func TestPrimeNumber(t *testing.T) {
	bits := risc.BytesFromLowBits(1109)
	test.RunAssert(t, map[risc.RegisterType]int32{}, 5,
		map[int]int8{0: bits[0], 1: bits[1], 2: bits[2], 3: bits[3]},
		test.ReadFile(t, "../res/prime-number.asm"),
		map[risc.RegisterType]int32{risc.A0: 4},
		map[int]int8{4: 1})
}

func TestMvms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		factory func() virtualMachine
	}{
		{
			name: "mvm1",
			factory: func() virtualMachine {
				return mvm1.NewCPU(5)
			},
		},
		{
			name: "mvm2",
			factory: func() virtualMachine {
				return mvm2.NewCPU(5)
			},
		},
		{
			name: "mvm3",
			factory: func() virtualMachine {
				return mvm3.NewCPU(5)
			},
		},
		{
			name: "mvm4",
			factory: func() virtualMachine {
				return mvm4.NewCPU(5)
			},
		},
	}

	for _, tc := range cases {
		for i := 2; i < 4096; i++ {
			t.Run(fmt.Sprintf("%s - %d", tc.name, i), func(t *testing.T) {
				vm := tc.factory()
				instructions := fmt.Sprintf(test.ReadFile(t, "../res/prime-number-fix.asm"), i)
				app, err := risc.Parse(instructions, false)
				require.NoError(t, err)
				_, err = vm.Run(app)
				require.NoError(t, err)

				want := isPrime(i)
				if want {
					assert.Equal(t, int8(1), vm.Context().Memory[4])
				} else {
					assert.Equal(t, int8(0), vm.Context().Memory[4])
				}
			})
		}
	}
}

func TestMvm1Execution(t *testing.T) {
	vm := mvm1.NewCPU(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"), false)
	require.NoError(t, err)
	require.Equal(t, float32(147432), cycles)
	stats(cycles)
}

func TestMvm2(t *testing.T) {
	vm := mvm2.NewCPU(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"), false)
	require.NoError(t, err)
	require.Equal(t, float32(11361), cycles)
	stats(cycles)
}

func TestMvm3(t *testing.T) {
	vm := mvm3.NewCPU(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"), false)
	require.NoError(t, err)
	require.Equal(t, float32(6918), cycles)
	stats(cycles)
}

func TestMvm4(t *testing.T) {
	vm := mvm4.NewCPU(5)
	cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number-1109.asm"), false)
	require.NoError(t, err)
	require.Equal(t, float32(6364), cycles)
	stats(cycles)
}

// FIXME
//func TestMvmr(t *testing.T) {
//	vm := mvm4.NewCPU(5)
//	c, err := execute(t, vm, `start:
//  jal zero, func
//  addi t1, t0, 3
//func:
//  addi t0, zero, 2`, false)
//	fmt.Println(c)
//	require.NoError(t, err)
//	assert.Equal(t, int32(5), vm.Context().Registers[risc.T1])
//}
