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

func execute(t *testing.T, vm virtualMachine, instructions string) (int, error) {
	app, err := risc.Parse(instructions)
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

func TestMvms(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		factory func() virtualMachine
	}{
		//{
		//	name: "mvm1",
		//	factory: func() virtualMachine {
		//		return mvm1.NewCPU(false, 5)
		//	},
		//},
		//{
		//	name: "mvm2",
		//	factory: func() virtualMachine {
		//		return mvm2.NewCPU(false, 5)
		//	},
		//},
		//{
		//	name: "mvm3",
		//	factory: func() virtualMachine {
		//		return mvm3.NewCPU(false, 5)
		//	},
		//},
		{
			name: "mvm4",
			factory: func() virtualMachine {
				return mvm4.NewCPU(true, 5)
			},
		},
	}

	for _, tc := range cases {
		//from := 5
		//to := 4096
		from := 10
		to := 11
		cache := make(map[int]bool, to-from+1)
		for i := from; i < to; i++ {
			cache[i] = isPrime(i)
		}
		for i := from; i < to; i++ {
			t.Run(fmt.Sprintf("Prime: %s - %d", tc.name, i), func(t *testing.T) {
				vm := tc.factory()
				instructions := fmt.Sprintf(test.ReadFile(t, "../res/prime-number-var.asm"), i)
				app, err := risc.Parse(instructions)
				require.NoError(t, err)
				_, err = vm.Run(app)
				require.NoError(t, err)

				want := cache[i]
				if want {
					assert.Equal(t, int8(1), vm.Context().Memory[4])
				} else {
					assert.Equal(t, int8(0), vm.Context().Memory[4])
				}
			})
		}
		//		t.Run(fmt.Sprintf("Jal: %s", tc.name), func(t *testing.T) {
		//			vm := tc.factory()
		//			_, err := execute(t, vm, `start:
		// jal zero, func
		// addi t1, t0, 3
		//func:
		// addi t0, zero, 2`)
		//			require.NoError(t, err)
		//			assert.Equal(t, int32(5), vm.Context().Registers[risc.T1])
		//		})
	}
}

func TestMvm1(t *testing.T) {
	vm := mvm1.NewCPU(false, 5)
	cycles, err := execute(t, vm, fmt.Sprintf(test.ReadFile(t, "../res/prime-number-var.asm"), 1109))
	require.NoError(t, err)
	require.Equal(t, 147485, cycles)
	stats(cycles)
}

func TestMvm2(t *testing.T) {
	vm := mvm2.NewCPU(false, 5)
	cycles, err := execute(t, vm, fmt.Sprintf(test.ReadFile(t, "../res/prime-number-var.asm"), 1109))
	require.NoError(t, err)
	require.Equal(t, 11365, cycles)
	stats(cycles)
}

func TestMvm3(t *testing.T) {
	vm := mvm3.NewCPU(false, 5)
	cycles, err := execute(t, vm, fmt.Sprintf(test.ReadFile(t, "../res/prime-number-var.asm"), 1109))
	require.NoError(t, err)
	require.Equal(t, 5853, cycles)
	stats(cycles)
}

func TestMvm4(t *testing.T) {
	vm := mvm4.NewCPU(false, 5)
	cycles, err := execute(t, vm, fmt.Sprintf(test.ReadFile(t, "../res/prime-number-var.asm"), 1109))
	require.NoError(t, err)
	require.Equal(t, 6406, cycles)
	stats(cycles)
}
