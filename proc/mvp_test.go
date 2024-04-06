package proc

import (
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teivah/majorana/proc/mvp1"
	"github.com/teivah/majorana/proc/mvp2"
	"github.com/teivah/majorana/proc/mvp3"
	"github.com/teivah/majorana/proc/mvp4"
	"github.com/teivah/majorana/proc/mvp5-0"
	"github.com/teivah/majorana/proc/mvp5-1"
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
	if n == 2 {
		return true
	}

	for i := 2; i <= n/2+1; i++ {
		if n%i == 0 {
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

func TestMvp5_0Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5_0.NewCPU(false, memory)
	}
	testPrime(t, factory, testFrom, testTo, false)
}

func TestMvp5_1Prime(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5_1.NewCPU(false, memory)
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

func TestMvp5_0Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5_0.NewCPU(false, memory)
	}
	testSums(t, factory, testFrom, testTo, false)
}

func TestMvp5_1Sums(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5_1.NewCPU(false, memory)
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
	testStringCopy(t, factory, testTo, true)
}

func TestMvp2StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp2.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo, true)
}

func TestMvp3StringCopy(t *testing.T) {
	// FIXME
	t.SkipNow()
	factory := func() virtualMachine {
		return mvp3.NewCPU(true, testTo*2)
	}
	testStringCopy(t, factory, testTo, true)
}

func TestMvp4StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp4.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo, true)
}

func TestMvp5_0StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5_0.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo, true)
}

func TestMvp5_1StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp5_1.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo, true)
}

func TestMvp6StringCopy(t *testing.T) {
	factory := func() virtualMachine {
		return mvp6.NewCPU(false, testTo*2)
	}
	testStringCopy(t, factory, testTo, true)
}

func testStringCopy(t *testing.T, factory func() virtualMachine, length int, stats bool) {
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
	cycle, err := vm.Run(app)
	require.NoError(t, err)
	for _, v := range vm.Context().Memory {
		assert.Equal(t, int8('1'), v)
	}

	if stats {
		t.Logf("Cycle: %d", cycle)
		for k, v := range vm.Stats() {
			t.Log(k, v)
		}
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
	tableRow := map[string]int{
		"MVP-1":   0,
		"MVP-2":   1,
		"MVP-3":   2,
		"MVP-4":   3,
		"MVP-5.0": 4,
		"MVP-5.1": 5,
		"MVP-6":   6,
	}

	vms := map[string]func(m int) virtualMachine{
		"MVP-1": func(m int) virtualMachine {
			return mvp1.NewCPU(false, m)
		},
		"MVP-2": func(m int) virtualMachine {
			return mvp2.NewCPU(false, m)
		},
		"MVP-3": func(m int) virtualMachine {
			return mvp3.NewCPU(false, m)
		},
		"MVP-4": func(m int) virtualMachine {
			return mvp4.NewCPU(false, m)
		},
		"MVP-5.0": func(m int) virtualMachine {
			return mvp5_0.NewCPU(false, m)
		},
		"MVP-5.1": func(m int) virtualMachine {
			return mvp5_1.NewCPU(false, m)
		},
		"MVP-6": func(m int) virtualMachine {
			return mvp6.NewCPU(false, m)
		},
	}

	primeOutput := make([]string, len(tableRow))
	prime := map[string]int{
		"MVP-1":   13120071,
		"MVP-2":   851454,
		"MVP-3":   450790,
		"MVP-4":   400717,
		"MVP-5.0": 400721,
		"MVP-5.1": 400716,
		"MVP-6":   400719,
	}
	t.Run("Prime", func(t *testing.T) {
		for name, factory := range vms {
			t.Run(name, func(t *testing.T) {
				vm := factory(5)
				cycles, err := execute(t, vm, fmt.Sprintf(test.ReadFile(t, "../res/prime-number-var-no-memory.asm"), benchPrimeNumber))
				require.NoError(t, err)
				assert.Equal(t, prime[name], cycles)
				primeOutput[tableRow[name]] = primeStats(cycles)
			})
		}
	})

	sumsOutput := make([]string, len(tableRow))
	sums := map[string]int{
		"MVP-1":   1921287,
		"MVP-2":   520260,
		"MVP-3":   454716,
		"MVP-4":   450621,
		"MVP-5.0": 462913,
		"MVP-5.1": 450625,
		"MVP-6":   62562,
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
				sumsOutput[tableRow[name]] = sumStats(cycles)
			})
		}
	})

	cpyOutput := make([]string, len(tableRow))
	cpy := map[string]int{
		"MVP-1":   5826769,
		"MVP-2":   1833023,
		"MVP-4":   1628381,
		"MVP-5.0": 1167466,
		"MVP-5.1": 1146986,
		"MVP-6":   341352,
	}
	t.Run("String copy", func(t *testing.T) {
		for name, factory := range vms {
			t.Run(name, func(t *testing.T) {
				switch name {
				case "MVP-3":
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
				cpyOutput[tableRow[name]] = sumStringCopy(cycles)
			})
		}
	})

	output := `| Machine | Prime number | Sum of array | String copy |
|:------:|:-----:|:-----:|:-----:|
`
	output += fmt.Sprintf("| Apple M1 | %.1f ns | %.1f ns | %.1f ns |\n", m1PrimeExecutionTime, m1SumsExecutionTime, m1StringCopyExecutionTime)
	var keys []string
	for k := range tableRow {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, mvp := range keys {
		idx := tableRow[mvp]
		output += fmt.Sprintf("| %s | %s | %s | %s |\n", mvp, primeOutput[idx], sumsOutput[idx], cpyOutput[idx])
	}
	fmt.Println(output)
}
