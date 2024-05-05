package proc

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teivah/majorana/common/bytes"
	"github.com/teivah/majorana/proc/mvp1"
	"github.com/teivah/majorana/proc/mvp2"
	"github.com/teivah/majorana/proc/mvp3"
	"github.com/teivah/majorana/proc/mvp4"
	"github.com/teivah/majorana/proc/mvp5"
	"github.com/teivah/majorana/proc/mvp6-0"
	"github.com/teivah/majorana/proc/mvp6-1"
	"github.com/teivah/majorana/proc/mvp6-2"
	"github.com/teivah/majorana/proc/mvp6-3"
	"github.com/teivah/majorana/proc/mvp7-0"
	mvp7_1 "github.com/teivah/majorana/proc/mvp7-1"
	mvp7_2 "github.com/teivah/majorana/proc/mvp7-2"
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
	testBubSort       = 100
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
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
	// Not passing, fixed from MVP 6.3
	//testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, testBubSort, factory, false)
	// Not passing, fixed from MVP 6.2
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
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
	testBubbleSort(t, testBubSort, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp7_0_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp7_0.NewCPU(false, memory, 2)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, testBubSort, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp7_0_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp7_0.NewCPU(false, memory, 3)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, testBubSort, factory, true)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp7_1_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp7_1.NewCPU(false, memory, 2)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, testBubSort, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp7_1_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp7_1.NewCPU(false, memory, 3)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, testBubSort, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp7_2_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp7_2.NewCPU(false, memory, 2)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, testBubSort, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}

func TestMvp7_2_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp7_2.NewCPU(false, memory, 3)
	}
	testPrime(t, factory, memory, testFrom, testTo, false)
	testSums(t, factory, memory, testFrom, testTo, false)
	testStringLength(t, factory, 1024, testTo, false)
	testStringCopy(t, factory, testTo*2, testTo, false)
	testBubbleSort(t, testBubSort, factory, false)
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
			t.Parallel()
			vm := factory(memory)
			instructions := test.ReadFile(t, "../res/prime-number.asm")
			app, err := risc.Parse(instructions)
			b := bytes.BytesFromLowBits(int32(i))
			vm.Context().Memory[0] = b[0]
			vm.Context().Memory[1] = b[1]
			vm.Context().Memory[2] = b[2]
			vm.Context().Memory[3] = b[3]
			require.NoError(t, err)
			cycle, err := vm.Run(app)
			require.NoError(t, err)

			want := cache[i]
			if want {
				assert.Equal(t, int8(1), vm.Context().Memory[4])
			} else {
				assert.Equal(t, int8(0), vm.Context().Memory[4])
			}

			printStats(t, stats, cycle, vm)
		})

		t.Run(fmt.Sprintf("Prime - with extra memory - %d", i), func(t *testing.T) {
			t.Parallel()
			vm := factory(memory)
			instructions := test.ReadFile(t, "../res/prime-number-2.asm")
			app, err := risc.Parse(instructions)
			b := bytes.BytesFromLowBits(int32(i))
			vm.Context().Memory[0] = b[0]
			vm.Context().Memory[1] = b[1]
			vm.Context().Memory[2] = b[2]
			vm.Context().Memory[3] = b[3]
			require.NoError(t, err)
			cycle, err := vm.Run(app)
			require.NoError(t, err)

			want := cache[i]
			if want {
				assert.Equal(t, int8(1), vm.Context().Memory[4])
			} else {
				assert.Equal(t, int8(0), vm.Context().Memory[4])
			}

			printStats(t, stats, cycle, vm)
		})

	}
}

func testSums(t *testing.T, factory func(int) virtualMachine, memory, from, to int, stats bool) {
	for i := from; i < to; i++ {
		t.Run(fmt.Sprintf("Sums - %d", i), func(t *testing.T) {
			t.Parallel()
			vm := factory(memory)
			n := i
			for i := 0; i < n; i++ {
				b := bytes.BytesFromLowBits(int32(i))
				vm.Context().Memory[4*i+0] = b[0]
				vm.Context().Memory[4*i+1] = b[1]
				vm.Context().Memory[4*i+2] = b[2]
				vm.Context().Memory[4*i+3] = b[3]
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

			printStats(t, stats, cycle, vm)
		})
	}
}

func testStringLength(t *testing.T, factory func(int) virtualMachine, memory int, length int, stats bool) {
	t.Run("String length", func(t *testing.T) {
		t.Parallel()
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

		got := bytes.I32FromBytes(vm.Context().Memory[0], vm.Context().Memory[1], vm.Context().Memory[2], vm.Context().Memory[3])
		assert.Equal(t, int32(length), got)

		printStats(t, stats, cycle, vm)
	})
}

func testStringCopy(t *testing.T, factory func(int) virtualMachine, memory int, length int, stats bool) {
	t.Run("String copy", func(t *testing.T) {
		t.Parallel()
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
		for i, v := range vm.Context().Memory {
			assert.Equal(t, int8('1'), v, i)
		}

		printStats(t, stats, cycle, vm)
	})
}

func testBubbleSort(t *testing.T, length int, factory func(int) virtualMachine, stats bool) {
	t.Run("Bubble sort", func(t *testing.T) {
		vm := factory(length * 4)
		for i := 0; i < length; i++ {
			b := bytes.BytesFromLowBits(int32(length - i))
			vm.Context().Memory[4*i+0] = b[0]
			vm.Context().Memory[4*i+1] = b[1]
			vm.Context().Memory[4*i+2] = b[2]
			vm.Context().Memory[4*i+3] = b[3]
		}
		vm.Context().Registers[risc.A0] = 0
		vm.Context().Registers[risc.A1] = int32(length)

		instructions := test.ReadFile(t, "../res/bubble-sort.asm")
		app, err := risc.Parse(instructions)
		require.NoError(t, err)
		cycle, err := vm.Run(app)
		require.NoError(t, err)

		for i := 0; i < length; i++ {
			n := bytes.I32FromBytes(vm.Context().Memory[4*i], vm.Context().Memory[4*i+1], vm.Context().Memory[4*i+2], vm.Context().Memory[4*i+3])
			require.Equal(t, int32(i+1), n)
		}

		printStats(t, stats, cycle, vm)
	})
}

func testConditionalBranch(t *testing.T, factory func(int) virtualMachine, stats bool) {
	t.Run("Conditional branch", func(t *testing.T) {
		t.Parallel()
		vm := factory(40)
		instructions := test.ReadFile(t, "../res/conditional-branch.asm")
		app, err := risc.Parse(instructions)
		require.NoError(t, err)
		cycle, err := vm.Run(app)
		require.NoError(t, err)
		assert.Equal(t, int32(0), vm.Context().Registers[risc.T1])
		assert.Equal(t, int32(2), vm.Context().Registers[risc.T2])
		printStats(t, stats, cycle, vm)
	})
}

func testSpectre(t *testing.T, factory func(int) virtualMachine, stats bool) {
	t.Run("Spectre", func(t *testing.T) {
		t.Parallel()
		vm := factory(40)
		secret := 42
		data := []int{3, 1, 2, 3, 0, 0, 0, 0, 0, secret}
		for idx, i := range data {
			b := bytes.BytesFromLowBits(int32(i))
			vm.Context().Memory[4*idx+0] = b[0]
			vm.Context().Memory[4*idx+1] = b[1]
			vm.Context().Memory[4*idx+2] = b[2]
			vm.Context().Memory[4*idx+3] = b[3]
		}
		instructions := test.ReadFile(t, "../res/spectre.asm")
		app, err := risc.Parse(instructions)
		require.NoError(t, err)
		cycle, err := vm.Run(app)
		require.NoError(t, err)
		got := bytes.I32FromBytes(vm.Context().Memory[0], vm.Context().Memory[1], vm.Context().Memory[2], vm.Context().Memory[3])
		assert.NotEqual(t, int32(secret), got)
		printStats(t, stats, cycle, vm)
	})
}

func printStats(t *testing.T, stats bool, cycle int, vm virtualMachine) {
	if !stats {
		return
	}
	t.Logf("Cycle: %d", cycle)
	s := vm.Stats()
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.Log(k, s[k])
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
		"MVP-7.0",
		"MVP-7.1",
		"MVP-7.2",
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
		versionMVP7_0
		versionMVP7_1
		versionMVP7_2
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
			versionMVP7_0: 301714,
			versionMVP7_1: 301714,
			versionMVP7_2: 301864,
		},
		"Sum": {
			versionMVP1:   10409494,
			versionMVP2:   1634638,
			versionMVP3:   465310,
			versionMVP4:   345743,
			versionMVP5:   341648,
			versionMVP6_0: 333721,
			versionMVP6_1: 321432,
			versionMVP6_2: 321432,
			versionMVP6_3: 321432,
			versionMVP7_0: 137257,
			versionMVP7_1: 137257,
			versionMVP7_2: 126282,
		},
		"String copy": {
			versionMVP1:   32349405,
			versionMVP2:   7280967,
			versionMVP3:   4201911,
			versionMVP4:   3883985,
			versionMVP5:   3853268,
			versionMVP6_0: 3855996,
			versionMVP6_1: 3834899,
			versionMVP6_2: 3834899,
			versionMVP6_3: 1956067,
			versionMVP7_0: 303003,
			versionMVP7_1: 303003,
			versionMVP7_2: 277645,
		},
		"String length": {
			versionMVP1:   19622376,
			versionMVP2:   3953646,
			versionMVP3:   874593,
			versionMVP4:   679223,
			versionMVP5:   668984,
			versionMVP6_0: 671976,
			versionMVP6_1: 641250,
			versionMVP6_2: 641250,
			versionMVP6_3: 641250,
			versionMVP7_0: 163635,
			versionMVP7_1: 163635,
			versionMVP7_2: 160378,
		},
		"Bubble sort": {
			versionMVP1:   158852511,
			versionMVP2:   42909111,
			versionMVP3:   6380745,
			versionMVP4:   4786503,
			versionMVP5:   4746704,
			versionMVP6_0: 2737247,
			versionMVP6_1: 2677345,
			versionMVP6_2: 2677345,
			versionMVP6_3: 2677345,
			versionMVP7_0: 24232735,
			versionMVP7_1: 1229965,
			versionMVP7_2: 949616,
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
		versionMVP7_0: func(m int) virtualMachine {
			return mvp7_0.NewCPU(false, m, 2)
		},
		versionMVP7_1: func(m int) virtualMachine {
			return mvp7_1.NewCPU(false, m, 2)
		},
		versionMVP7_2: func(m int) virtualMachine {
			return mvp7_2.NewCPU(false, m, 3)
		},
	}

	primeOutput := make([]benchResult, totalVersions)
	t.Run("Prime", func(t *testing.T) {
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				t.Parallel()
				v := expected["Prime"][idx]
				if v == 0 {
					t.SkipNow()
				}

				vm := factory(5)
				b := bytes.BytesFromLowBits(int32(benchPrimeNumber))
				vm.Context().Memory[0] = b[0]
				vm.Context().Memory[1] = b[1]
				vm.Context().Memory[2] = b[2]
				vm.Context().Memory[3] = b[3]

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

	sumsOutput := make([]benchResult, totalVersions)
	t.Run("Sum", func(t *testing.T) {
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				t.Parallel()
				v := expected["Sum"][idx]
				if v == 0 {
					t.SkipNow()
				}

				vm := factory(memory)
				n := benchSums
				for i := 0; i < n; i++ {
					b := bytes.BytesFromLowBits(int32(i))
					vm.Context().Memory[4*i+0] = b[0]
					vm.Context().Memory[4*i+1] = b[1]
					vm.Context().Memory[4*i+2] = b[2]
					vm.Context().Memory[4*i+3] = b[3]
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

	cpyOutput := make([]benchResult, totalVersions)
	t.Run("String copy", func(t *testing.T) {
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				t.Parallel()
				v := expected["String copy"][idx]
				if v == 0 {
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

	lengthOutput := make([]benchResult, totalVersions)
	t.Run("String length", func(t *testing.T) {
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				t.Parallel()
				v := expected["String length"][idx]
				if v == 0 {
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

				got := bytes.I32FromBytes(vm.Context().Memory[0], vm.Context().Memory[1], vm.Context().Memory[2], vm.Context().Memory[3])
				assert.Equal(t, int32(length), got)

				assert.Equal(t, v, cycles)
				lengthOutput[idx] = stringLengthStats(cycles)
			})
		}
	})

	bubbleOutput := make([]benchResult, totalVersions)
	t.Run("Bubble sort", func(t *testing.T) {
		for idx, factory := range vms {
			t.Run(versions[idx], func(t *testing.T) {
				t.Parallel()
				v := expected["Bubble sort"][idx]
				if v == 0 {
					t.SkipNow()
				}

				data := benchBubSort
				vm := factory(data * 4)
				for i := 0; i < data; i++ {
					b := bytes.BytesFromLowBits(int32(data - i))
					vm.Context().Memory[4*i+0] = b[0]
					vm.Context().Memory[4*i+1] = b[1]
					vm.Context().Memory[4*i+2] = b[2]
					vm.Context().Memory[4*i+3] = b[3]
				}
				vm.Context().Registers[risc.A0] = 0
				vm.Context().Registers[risc.A1] = int32(data)

				instructions := test.ReadFile(t, "../res/bubble-sort.asm")
				app, err := risc.Parse(instructions)
				require.NoError(t, err)
				cycles, err := vm.Run(app)
				require.NoError(t, err)

				for i := 0; i < data; i++ {
					n := bytes.I32FromBytes(vm.Context().Memory[4*i], vm.Context().Memory[4*i+1], vm.Context().Memory[4*i+2], vm.Context().Memory[4*i+3])
					require.Equal(t, int32(i+1), n)
				}

				assert.Equal(t, v, cycles)
				bubbleOutput[idx] = bubbleSortStats(cycles)
			})
		}
	})

	output := `| Machine | Prime number | Sum of array | String copy | String length | Bubble sort | Avg |
|:------:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|
`
	output += fmt.Sprintf("| Apple M1 | %.1f ns | %.1f ns | %.1f ns | %.1f ns | %.1f ns | 1.0 |\n", m1PrimeExecutionTime, m1SumsExecutionTime, m1StringCopyExecutionTime, m1StringLengthExecutionTime, m1BubbleSortExecutionTime)
	var keys []string
	for _, k := range versions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for idx := range keys {
		sum := primeOutput[idx].slower + sumsOutput[idx].slower + cpyOutput[idx].slower + lengthOutput[idx].slower + bubbleOutput[idx].slower
		avg := sum / 5
		output += fmt.Sprintf("| %s | %.0f ns, %.1fx slower | %.0f ns, %.1fx slower | %.0f ns, %.1fx slower | %.0f ns, %.1fx slower | %.0f ns, %.1fx slower | %.1f |\n", versions[idx], primeOutput[idx].executionNs, primeOutput[idx].slower, sumsOutput[idx].executionNs, sumsOutput[idx].slower, cpyOutput[idx].executionNs, cpyOutput[idx].slower, lengthOutput[idx].executionNs, lengthOutput[idx].slower, bubbleOutput[idx].executionNs, bubbleOutput[idx].slower, avg)
	}
	fmt.Println(output)
}
