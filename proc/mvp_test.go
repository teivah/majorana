package proc

import (
	"fmt"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teivah/majorana/proc/mvp1"
	"github.com/teivah/majorana/proc/mvp2"
	"github.com/teivah/majorana/proc/mvp3"
	"github.com/teivah/majorana/proc/mvp4"
	"github.com/teivah/majorana/proc/mvp5"
	mvp6_0 "github.com/teivah/majorana/proc/mvp6-0"
	mvp6_1 "github.com/teivah/majorana/proc/mvp6-1"
	mvp6_2 "github.com/teivah/majorana/proc/mvp6-2"
	mvp6_3 "github.com/teivah/majorana/proc/mvp6-3"
	"github.com/teivah/majorana/risc"
	"github.com/teivah/majorana/test"
)

const (
	memory            = benchSums * 4
	benchPrimeNumber  = 100151
	benchSums         = 4096
	benchStringCopy   = 10 * 1024 // 10 KB
	benchStringLength = 10 * 1024 // 10 KB
	benchBubSort      = 200
	testFrom          = 2
	testTo            = 200
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

func strncpy(dst, src []byte, n int) {
	var i int
	for i = 0; i < n && i < len(src); i++ {
		dst[i] = src[i]
	}
	for ; i < n; i++ {
		dst[i] = 0
	}
}

func strlen(bytes []byte) int {
	res := 0
	for i := 0; i < len(bytes); i++ {

	}
	return res
}

func bubsort(list []int32, size int) {
	swapped := true
	for swapped {
		swapped = false
		for i := 1; i < size; i++ {
			if list[i-1] > list[i] {
				// Swap elements
				list[i-1], list[i] = list[i], list[i-1]
				swapped = true
			}
		}
	}
}

func TestMvp1(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp1.NewCPU(false, memory)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp2.NewCPU(false, memory)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp3.NewCPU(false, memory)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp4(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp4.NewCPU(false, memory)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp5(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp5.NewCPU(false, memory)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp6_0_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_0.NewCPU(false, memory, 2, 2)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp6_0_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_0.NewCPU(false, memory, 3, 3)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp6_1_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_1.NewCPU(false, memory, 2, 2)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp6_1_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_1.NewCPU(false, memory, 3, 3)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	// Not passing, fixed in MVP 6.3
	//testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	// Not passing, fixed in MVP 6.3
	//testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp6_2_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_2.NewCPU(false, memory, 2, 2)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp6_2_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_2.NewCPU(false, memory, 3, 3)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	// Not passing, fixed in MVP 6.3
	//testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp6_3_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_3.NewCPU(false, memory, 2, 2)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp6_3_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_3.NewCPU(false, memory, 3, 3)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func testPrime(t *testing.T, factory func(int) virtualMachine, memory, from, to int, stats bool) {
	cache := make(map[int]bool, to-from+1)
	for i := from; i < to; i++ {
		cache[i] = isPrime(i)
	}

	for i := from; i < to; i++ {
		t.Run(fmt.Sprintf("Prime - nominal - %d", i), func(t *testing.T) {
			vm := factory(memory)
			instructions := test.ReadFile(t, "../res/prime-number.asm")
			app, err := risc.Parse(instructions)
			bytes := risc.BytesFromLowBits(int32(i))
			vm.Context().Memory[0] = bytes[0]
			vm.Context().Memory[1] = bytes[1]
			vm.Context().Memory[2] = bytes[2]
			vm.Context().Memory[3] = bytes[3]
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

		t.Run(fmt.Sprintf("Prime - with extra memory - %d", i), func(t *testing.T) {
			vm := factory(memory)
			instructions := test.ReadFile(t, "../res/prime-number-2.asm")
			app, err := risc.Parse(instructions)
			bytes := risc.BytesFromLowBits(int32(i))
			vm.Context().Memory[0] = bytes[0]
			vm.Context().Memory[1] = bytes[1]
			vm.Context().Memory[2] = bytes[2]
			vm.Context().Memory[3] = bytes[3]
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

func testSums(t *testing.T, factory func(int) virtualMachine, memory, from, to int, stats bool) {
	for i := from; i < to; i++ {
		t.Run(fmt.Sprintf("Sums - %d", i), func(t *testing.T) {
			vm := factory(memory)
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

func testStringLength(t *testing.T, factory func(int) virtualMachine, memory int, length int, stats bool) {
	t.Run("String length", func(t *testing.T) {
		vm := factory(memory)
		for i := 0; i < length; i++ {
			vm.Context().Memory[i] = '1'
		}
		vm.Context().Registers[risc.A0] = int32(0)

		instructions := test.ReadFile(t, "../res/string-length.asm")
		app, err := risc.Parse(instructions)
		require.NoError(t, err)
		cycle, err := vm.Run(app)
		require.NoError(t, err)

		got := risc.I32FromBytes(vm.Context().Memory[0], vm.Context().Memory[1], vm.Context().Memory[2], vm.Context().Memory[3])
		assert.Equal(t, int32(length), got)

		if stats {
			t.Logf("Cycle: %d", cycle)
			for k, v := range vm.Stats() {
				t.Log(k, v)
			}
		}
	})
}

func testStringCopy(t *testing.T, factory func(int) virtualMachine, memory int, length int, stats bool) {
	t.Run("String copy", func(t *testing.T) {
		vm := factory(memory)
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
	})
}

func testBubbleSort(t *testing.T, factory func(int) virtualMachine, stats bool) {
	t.Run("Bubble sort", func(t *testing.T) {
		//data := 100
		data := 151
		vm := factory(data * 4)
		for i := 0; i < data; i++ {
			bytes := risc.BytesFromLowBits(int32(data - i))
			vm.Context().Memory[4*i+0] = bytes[0]
			vm.Context().Memory[4*i+1] = bytes[1]
			vm.Context().Memory[4*i+2] = bytes[2]
			vm.Context().Memory[4*i+3] = bytes[3]
		}
		vm.Context().Registers[risc.A0] = 0
		vm.Context().Registers[risc.A1] = int32(data)

		instructions := test.ReadFile(t, "../res/bubble-sort.asm")
		app, err := risc.Parse(instructions)
		require.NoError(t, err)
		cycle, err := vm.Run(app)
		require.NoError(t, err)

		for i := 0; i < data; i++ {
			n := risc.I32FromBytes(vm.Context().Memory[4*i], vm.Context().Memory[4*i+1], vm.Context().Memory[4*i+2], vm.Context().Memory[4*i+3])
			require.Equal(t, int32(i+1), n)
		}

		if stats {
			t.Logf("Cycle: %d", cycle)
			for k, v := range vm.Stats() {
				t.Log(k, v)
			}
		}
	})
}

func testConditionalBranch(t *testing.T, factory func(int) virtualMachine, stats bool) {
	t.Run("Conditional branch 1", func(t *testing.T) {
		vm := factory(40)
		instructions := test.ReadFile(t, "../res/conditional-branch-1.asm")
		app, err := risc.Parse(instructions)
		require.NoError(t, err)
		_, err = vm.Run(app)
		require.NoError(t, err)
		assert.Equal(t, int32(0), vm.Context().Registers[risc.T1])
		assert.Equal(t, int32(2), vm.Context().Registers[risc.T2])
	})
	t.Run("Conditional branch 2", func(t *testing.T) {
		vm := factory(40)
		bytes := risc.BytesFromLowBits(int32(2))
		vm.Context().Memory[0] = bytes[0]
		vm.Context().Memory[1] = bytes[1]
		vm.Context().Memory[2] = bytes[2]
		vm.Context().Memory[3] = bytes[3]
		instructions := test.ReadFile(t, "../res/conditional-branch-2.asm")
		app, err := risc.Parse(instructions)
		require.NoError(t, err)
		_, err = vm.Run(app)
		require.NoError(t, err)
		assert.Equal(t, int32(2), vm.Context().Registers[risc.T0])
		assert.Equal(t, int32(1), vm.Context().Registers[risc.T1])
		assert.Equal(t, int32(2), vm.Context().Registers[risc.T2])
	})
}

func testSpectre(t *testing.T, factory func(int) virtualMachine, stats bool) {
	t.Run("Spectre", func(t *testing.T) {
		vm := factory(40)
		secret := 42
		data := []int{3, 1, 2, 3, 0, 0, 0, 0, 0, secret}
		for idx, i := range data {
			bytes := risc.BytesFromLowBits(int32(i))
			vm.Context().Memory[4*idx+0] = bytes[0]
			vm.Context().Memory[4*idx+1] = bytes[1]
			vm.Context().Memory[4*idx+2] = bytes[2]
			vm.Context().Memory[4*idx+3] = bytes[3]
		}
		instructions := test.ReadFile(t, "../res/spectre.asm")
		app, err := risc.Parse(instructions)
		require.NoError(t, err)
		_, err = vm.Run(app)
		require.NoError(t, err)
		got := risc.I32FromBytes(vm.Context().Memory[0], vm.Context().Memory[1], vm.Context().Memory[2], vm.Context().Memory[3])
		assert.NotEqual(t, int32(secret), got)
	})
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
	t.Parallel()

	wg := sync.WaitGroup{}
	versions := []string{
		"MVP-1",
		"MVP-2",
		"MVP-3",
		"MVP-4",
		"MVP-5",
		"MVP-6.0",
		"MVP-6.1",
		"MVP-6.2",
		"MVP-6.3",
	}
	const (
		versionMVP1 = iota
		versionMVP2
		versionMVP3
		versionMVP4
		versionMVP5
		versionMVP6_0
		versionMVP6_1
		versionMVP6_2
		versionMVP6_3
		totalVersions
	)

	expected := map[string][]int{
		"Prime": {
			versionMVP1:   77969626,
			versionMVP2:   1353346,
			versionMVP3:   1353352,
			versionMVP4:   451973,
			versionMVP5:   401900,
			versionMVP6_0: 401860,
			versionMVP6_1: 351784,
			versionMVP6_2: 351784,
			versionMVP6_3: 301710,
		},
		"Sum": {
			versionMVP1:   10409494,
			versionMVP2:   1634638,
			versionMVP3:   465310,
			versionMVP4:   345743,
			versionMVP5:   341648,
			versionMVP6_0: 153241,
			versionMVP6_1: 140952,
			versionMVP6_2: 140952,
			versionMVP6_3: 140952,
		},
		"String copy": {
			versionMVP1:   32349405,
			versionMVP2:   7280967,
			versionMVP3:   4201911,
			versionMVP4:   3883985,
			versionMVP5:   3853268,
			versionMVP6_0: 3382236,
			versionMVP6_1: 3361139,
			versionMVP6_2: 3361139,
			versionMVP6_3: 1722900,
		},
		"String length": {
			versionMVP1:   19622376,
			versionMVP2:   3953646,
			versionMVP3:   874593,
			versionMVP4:   679223,
			versionMVP5:   668984,
			versionMVP6_0: 198216,
			versionMVP6_1: 167490,
			versionMVP6_2: 167490,
			versionMVP6_3: 167490,
		},
		"Bubble sort": {
			versionMVP1:   158852511,
			versionMVP2:   42909111,
			versionMVP3:   6380745,
			versionMVP4:   4786503,
			versionMVP5:   4746704,
			versionMVP6_0: 867211,
			versionMVP6_1: 807309,
			versionMVP6_2: 807309,
			versionMVP6_3: 807309,
		},
	}

	vms := []func(m int) virtualMachine{
		versionMVP1: func(m int) virtualMachine {
			return mvp1.NewCPU(false, m)
		},
		versionMVP2: func(m int) virtualMachine {
			return mvp2.NewCPU(false, m)
		},
		versionMVP3: func(m int) virtualMachine {
			return mvp3.NewCPU(false, m)
		},
		versionMVP4: func(m int) virtualMachine {
			return mvp4.NewCPU(false, m)
		},
		versionMVP5: func(m int) virtualMachine {
			return mvp5.NewCPU(false, m)
		},
		versionMVP6_0: func(m int) virtualMachine {
			return mvp6_0.NewCPU(false, m, 2, 2)
		},
		versionMVP6_1: func(m int) virtualMachine {
			return mvp6_1.NewCPU(false, m, 2, 2)
		},
		versionMVP6_2: func(m int) virtualMachine {
			return mvp6_2.NewCPU(false, m, 2, 2)
		},
		versionMVP6_3: func(m int) virtualMachine {
			return mvp6_3.NewCPU(false, m, 2, 2)
		},
	}

	primeOutput := make([]string, totalVersions)
	wg.Add(1)
	t.Run("Prime", func(t *testing.T) {
		t.Parallel()
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				v := expected["Prime"][idx]
				if v == -1 {
					t.SkipNow()
				}

				vm := factory(5)
				bytes := risc.BytesFromLowBits(int32(benchPrimeNumber))
				vm.Context().Memory[0] = bytes[0]
				vm.Context().Memory[1] = bytes[1]
				vm.Context().Memory[2] = bytes[2]
				vm.Context().Memory[3] = bytes[3]

				cycles, err := execute(t, vm, test.ReadFile(t, "../res/prime-number.asm"))
				require.NoError(t, err)

				want := isPrime(benchPrimeNumber)
				if want {
					assert.Equal(t, int8(1), vm.Context().Memory[4])
				} else {
					assert.Equal(t, int8(0), vm.Context().Memory[4])
				}
				assert.Equal(t, v, cycles)
				primeOutput[idx] = primeStats(cycles)
			})
		}
	})

	sumsOutput := make([]string, totalVersions)
	wg.Add(1)
	t.Run("Sum", func(t *testing.T) {
		t.Parallel()
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				v := expected["Sum"][idx]
				if v == -1 {
					t.SkipNow()
				}

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

				s := make([]int, 0, n)
				for i := 0; i < n; i++ {
					s = append(s, i)
				}
				assert.Equal(t, int32(sumArray(s)), vm.Context().Registers[risc.A0])

				assert.Equal(t, v, cycles)
				sumsOutput[idx] = sumStats(cycles)
			})
		}
	})

	cpyOutput := make([]string, totalVersions)
	wg.Add(1)
	t.Run("String copy", func(t *testing.T) {
		t.Parallel()
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				v := expected["String copy"][idx]
				if v == -1 {
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
				assert.Equal(t, v, cycles)
				cpyOutput[idx] = stringCopyStats(cycles)
			})
		}
	})

	lengthOutput := make([]string, totalVersions)
	wg.Add(1)
	t.Run("String length", func(t *testing.T) {
		t.Parallel()
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				v := expected["String length"][idx]
				if v == -1 {
					t.SkipNow()
				}

				length := benchStringLength
				vm := factory(benchStringLength * 2)
				for i := 0; i < length; i++ {
					vm.Context().Memory[i] = '1'
				}
				vm.Context().Registers[risc.A0] = int32(0)

				instructions := test.ReadFile(t, "../res/string-length.asm")
				app, err := risc.Parse(instructions)
				require.NoError(t, err)
				cycles, err := vm.Run(app)
				require.NoError(t, err)

				got := risc.I32FromBytes(vm.Context().Memory[0], vm.Context().Memory[1], vm.Context().Memory[2], vm.Context().Memory[3])
				assert.Equal(t, int32(length), got)

				assert.Equal(t, v, cycles)
				lengthOutput[idx] = stringLengthStats(cycles)
			})
		}
	})

	bubbleOutput := make([]string, totalVersions)
	wg.Add(1)
	t.Run("Bubble sort", func(t *testing.T) {
		t.Parallel()
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				v := expected["Bubble sort"][idx]
				if v == -1 {
					t.SkipNow()
				}

				data := benchBubSort
				vm := factory(data * 4)
				for i := 0; i < data; i++ {
					bytes := risc.BytesFromLowBits(int32(data - i))
					vm.Context().Memory[4*i+0] = bytes[0]
					vm.Context().Memory[4*i+1] = bytes[1]
					vm.Context().Memory[4*i+2] = bytes[2]
					vm.Context().Memory[4*i+3] = bytes[3]
				}
				vm.Context().Registers[risc.A0] = 0
				vm.Context().Registers[risc.A1] = int32(data)

				instructions := test.ReadFile(t, "../res/bubble-sort.asm")
				app, err := risc.Parse(instructions)
				require.NoError(t, err)
				cycles, err := vm.Run(app)
				require.NoError(t, err)

				for i := 0; i < data; i++ {
					n := risc.I32FromBytes(vm.Context().Memory[4*i], vm.Context().Memory[4*i+1], vm.Context().Memory[4*i+2], vm.Context().Memory[4*i+3])
					require.Equal(t, int32(i+1), n)
				}

				assert.Equal(t, v, cycles)
				bubbleOutput[idx] = bubbleSortStats(cycles)
			})
		}
	})

	output := `| Machine | Prime number | Sum of array | String copy | String length | Bubble sort |
|:------:|:-----:|:-----:|:-----:|:-----:|:-----:|
`
	output += fmt.Sprintf("| Apple M1 | %.1f ns | %.1f ns | %.1f ns | %.1f ns | %.1f ns |\n", m1PrimeExecutionTime, m1SumsExecutionTime, m1StringCopyExecutionTime, m1StringLengthExecutionTime, m1BubbleSortExecutionTime)
	var keys []string
	for _, k := range versions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for idx := range keys {
		output += fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n", versions[idx], primeOutput[idx], sumsOutput[idx], cpyOutput[idx], lengthOutput[idx], bubbleOutput[idx])
	}
	fmt.Println(output)
}
