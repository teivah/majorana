package proc

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teivah/majorana/proc/mvp1"
	"github.com/teivah/majorana/proc/mvp2"
	"github.com/teivah/majorana/proc/mvp3"
	"github.com/teivah/majorana/proc/mvp4"
	"github.com/teivah/majorana/proc/mvp5"
	"github.com/teivah/majorana/proc/mvp6"
	"github.com/teivah/majorana/risc"
	"github.com/teivah/majorana/test"
)

const (
	memory           = benchSums * 4
	benchPrimeNumber = 100151
	benchSums        = 4096
	benchStringCopy  = 10 * 1024 // 10 KB
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

func TestMvp1Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvp1.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvp2Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvp2.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvp3Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvp3.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvp4Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvp4.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvp5Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvp6Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvp6.NewCPU(false, memory)
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

func TestMvp1Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvp1.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvp2Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvp2.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvp3Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvp3.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvp4Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvp4.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvp5Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvp6Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvp6.NewCPU(false, memory)
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

func strncpy(dst, src []byte, n int) {
	var i int
	for i = 0; i < n && i < len(src); i++ {
		dst[i] = src[i]
	}
	for ; i < n; i++ {
		dst[i] = 0
	}
}

func TestMvp1StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp1.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo)
}

func TestMvp2StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp2.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo)
}

func TestMvp5StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo)
}

func TestMvp6StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp6.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo)
}

func testStringCopy(t *testing.T, factory func() virtualMachine, length int) {
	vm := factory()
	for i := 0; i < length; i++ {
		vm.Context().Memory[i] = '1'
	}
	vm.Context().Registers[risc.A1] = int32(0)
	vm.Context().Registers[risc.A0] = int32(length)
	vm.Context().Registers[risc.A2] = int32(length)

	instructions := test.ReadFile(t, "../res/string-copy.asm")
	app, err := risc.Parse(instructions)
	require.NoError(t, err)
	_, err = vm.Run(app)
	require.NoError(t, err)
	for _, v := range vm.Context().Memory {
		assert.Equal(t, int8('1'), v)
	}
}

//func TestMvp1Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvp1.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func TestMvp2Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvp2.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func TestMvp3Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvp3.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func TestMvp4Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvp4.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func TestMvp5Jal(t *testing.T) {
//	factory := func() virtualMachine {
//		return mvp5.NewCPU(false, memory)
//	}
//	testJal(t, factory)
//}
//
//func testJal(t *testing.T, factory func() virtualMachine) {
//	vm := factory()
//	_, err := execute(t, vm, `start:
//	jal zero, func
//	addi t1, t0, 3
//	func:
//	addi t0, zero, 2`)
//	require.NoError(t, err)
//	assert.Equal(t, int32(5), vm.Context().Registers[risc.T1])
//}

func TestBenchmarks(t *testing.T) {
	vms := map[string]func(m int) virtualMachine{
		"mvp1": func(m int) virtualMachine {
			return mvp1.NewCPU(false, m)
		},
		"mvp2": func(m int) virtualMachine {
			return mvp2.NewCPU(false, m)
		},
		"mvp3": func(m int) virtualMachine {
			return mvp3.NewCPU(false, m)
		},
		"mvp4": func(m int) virtualMachine {
			return mvp4.NewCPU(false, m)
		},
		"mvp5": func(m int) virtualMachine {
			return mvp5.NewCPU(false, m)
		},
		"mvp6": func(m int) virtualMachine {
			return mvp6.NewCPU(false, m)
		},
	}

	prime := map[string]int{
		"mvp1": 13170146,
		"mvp2": 901529,
		"mvp3": 450790,
		"mvp4": 400717,
		"mvp5": 400721,
		"mvp6": 400716,
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
		"mvp1": 1716487,
		"mvp2": 311364,
		"mvp3": 249916,
		"mvp4": 245821,
		"mvp5": 258113,
		"mvp6": 245825,
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

	cpy := map[string]int{
		"mvp1": 5314769,
		"mvp2": 1310783,
		"mvp3": 249916,
		"mvp4": 245821,
		"mvp5": 655466,
		"mvp6": 634986,
	}
	t.Run("String copy", func(t *testing.T) {
		for name, factory := range vms {
			t.Run(name, func(t *testing.T) {
				switch name {
				case "mvp3", "mvp4":
					t.SkipNow()
				}

				length := benchStringCopy
				vm := factory(2 * length)
				for i := 0; i < length; i++ {
					vm.Context().Memory[i] = '1'
				}
				vm.Context().Registers[risc.A1] = int32(0)
				vm.Context().Registers[risc.A0] = int32(length)
				vm.Context().Registers[risc.A2] = int32(length)

				instructions := test.ReadFile(t, "../res/string-copy.asm")
				app, err := risc.Parse(instructions)
				require.NoError(t, err)
				cycles, err := vm.Run(app)
				require.NoError(t, err)
				for _, v := range vm.Context().Memory {
					assert.Equal(t, int8('1'), v)
				}
				require.NoError(t, err)
				assert.Equal(t, cpy[name], cycles)
				sumStringCopy(t, cycles)
			})
		}
	})
}
