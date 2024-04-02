package proc

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teivah/majorana/proc/mvm1"
	"github.com/teivah/majorana/proc/mvm2"
	"github.com/teivah/majorana/proc/mvm3"
	"github.com/teivah/majorana/proc/mvm4"
	"github.com/teivah/majorana/proc/mvm5"
	"github.com/teivah/majorana/risc"
	"github.com/teivah/majorana/test"
)

const (
	memory           = benchSums * 4
	benchPrimeNumber = 100151
	benchSums        = 4096
	testFrom         = 2
	testTo           = 200
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

func sumArray(s []int) int {
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += i
	}
	return sum
}

func TestMvm1Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvm1.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvm2Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvm2.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvm3Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvm3.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvm4Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvm4.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvm5Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvm5.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func testPrime(t *testing.T, factory func() virtualMachine, from, to int, stats bool) {
	cache := make(map[int]bool, to-from+1)
	for i := from; i < to; i++ {
		cache[i] = isPrime(i)
	}

	for i := from; i < to; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			vm := factory()
			instructions := fmt.Sprintf(test.ReadFile(t, "../res/prime-number-var.asm"), i)
			app, err := risc.Parse(instructions)
			require.NoError(t, err)
			cycle, err := vm.Run(app)
			require.NoError(t, err)

			want := cache[i]
			if want {
				assert.Equal(t, int8(1), vm.Context().Memory[4])
			} else {
				assert.Equal(t, int8(0), vm.Context().Memory[4])
			}

			if stats {
				t.Logf("Cycle: %d", cycle)
				for k, v := range vm.Stats() {
					t.Log(k, v)
				}
			}
		})
	}
}

func TestMvm1Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvm1.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvm2Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvm2.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvm3Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvm3.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvm4Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvm4.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvm5Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvm5.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func testSums(t *testing.T, factory func() virtualMachine, from, to int, stats bool) {
	for i := from; i < to; i++ {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			vm := factory()
			n := i
			for i := 0; i < n; i++ {
				bytes := risc.BytesFromLowBits(int32(i))
				vm.Context().Memory[4*i+0] = bytes[0]
				vm.Context().Memory[4*i+1] = bytes[1]
				vm.Context().Memory[4*i+2] = bytes[2]
				vm.Context().Memory[4*i+3] = bytes[3]
			}
			vm.Context().Registers[risc.A1] = int32(n)

			instructions := fmt.Sprintf(test.ReadFile(t, "../res/array-sum.asm"), "")
			app, err := risc.Parse(instructions)
			require.NoError(t, err)
			cycle, err := vm.Run(app)
			require.NoError(t, err)

			s := make([]int, 0, n)
			for i := 0; i < n; i++ {
				s = append(s, i)
			}
			assert.Equal(t, int32(sumArray(s)), vm.Context().Registers[risc.A0])

			if stats {
				t.Logf("Cycle: %d", cycle)
				for k, v := range vm.Stats() {
					t.Log(k, v)
				}
			}
		})
	}
}

//func TestMvm1Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvm1.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func TestMvm2Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvm2.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func TestMvm3Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvm3.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func TestMvm4Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvm4.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func TestMvm5Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvm5.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}

func testJal(t *testing.T, factory func() virtualMachine) {
	vm := factory()
	_, err := execute(t, vm, `start:
	jal zero, func
	addi t1, t0, 3
	func:
	addi t0, zero, 2`)
	require.NoError(t, err)
	assert.Equal(t, int32(5), vm.Context().Registers[risc.T1])
}

func TestBenchmarks(t *testing.T) {
	vms := map[string]func(m int) virtualMachine{
		"mvm1": func(m int) virtualMachine {
			return mvm1.NewCPU(false, m)
		},
		"mvm2": func(m int) virtualMachine {
			return mvm2.NewCPU(false, m)
		},
		"mvm3": func(m int) virtualMachine {
			return mvm3.NewCPU(false, m)
		},
		"mvm4": func(m int) virtualMachine {
			return mvm4.NewCPU(false, m)
		},
		"mvm5": func(m int) virtualMachine {
			return mvm5.NewCPU(false, m)
		},
	}

	prime := map[string]int{
		"mvm1": 13170146,
		"mvm2": 901529,
		"mvm3": 450790,
		"mvm4": 400717,
		"mvm5": 400721,
	}
	t.Run("Prime", func(t *testing.T) {
		for name, factory := range vms {
			t.Run(name, func(t *testing.T) {
				vm := factory(5)
				cycles, err := execute(t, vm, fmt.Sprintf(test.ReadFile(t, "../res/prime-number-var-no-memory.asm"), benchPrimeNumber))
				require.NoError(t, err)
				assert.Equal(t, prime[name], cycles)
				primeStats(t, cycles)
			})
		}
	})

	sums := map[string]int{
		"mvm1": 1716487,
		"mvm2": 311364,
		"mvm3": 249916,
		"mvm4": 245821,
		"mvm5": 258113,
	}
	t.Run("Sum", func(t *testing.T) {
		for name, factory := range vms {
			t.Run(name, func(t *testing.T) {
				vm := factory(memory)
				n := benchSums
				for i := 0; i < n; i++ {
					bytes := risc.BytesFromLowBits(int32(i))
					vm.Context().Memory[4*i+0] = bytes[0]
					vm.Context().Memory[4*i+1] = bytes[1]
					vm.Context().Memory[4*i+2] = bytes[2]
					vm.Context().Memory[4*i+3] = bytes[3]
				}
				vm.Context().Registers[risc.A1] = int32(n)

				cycles, err := execute(t, vm, fmt.Sprintf(test.ReadFile(t, "../res/array-sum.asm"), benchSums))
				require.NoError(t, err)
				assert.Equal(t, sums[name], cycles)
				sumStats(t, cycles)
			})
		}
	})
}
