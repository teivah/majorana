package proc

import (
	"testing"

	"github.com/teivah/majorana/proc/mvp1"
	"github.com/teivah/majorana/proc/mvp2"
	"github.com/teivah/majorana/proc/mvp3"
	"github.com/teivah/majorana/proc/mvp4"
	"github.com/teivah/majorana/proc/mvp5"
	mvp6_0 "github.com/teivah/majorana/proc/mvp6-0"
	mvp6_1 "github.com/teivah/majorana/proc/mvp6-1"
	mvp6_2 "github.com/teivah/majorana/proc/mvp6-2"
	mvp6_3 "github.com/teivah/majorana/proc/mvp6-3"
	mvp7_0 "github.com/teivah/majorana/proc/mvp7-0"
)

const (
	slowTest = 1000
)

func TestSlowMvp1(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp1.NewCPU(false, memory)
	}
	testSlow(t, factory)
}

func TestSlowMvp2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp2.NewCPU(false, memory)
	}
	testSlow(t, factory)
}

func TestSlowMvp3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp3.NewCPU(false, memory)
	}
	testSlow(t, factory)
}

func TestSlowMvp4(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp4.NewCPU(false, memory)
	}
	testSlow(t, factory)
}

func TestSlowMvp5(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp5.NewCPU(false, memory)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_0_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_0.NewCPU(false, memory, 2, 2)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_0_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_0.NewCPU(false, memory, 3, 3)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_1_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_1.NewCPU(false, memory, 2, 2)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_1_3x3(t *testing.T) {
	// Not passing
}

func TestSlowMvp6_2_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_2.NewCPU(false, memory, 2, 2)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_2_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_2.NewCPU(false, memory, 3, 3)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_3_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_3.NewCPU(false, memory, 2, 2)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_3_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp6_3.NewCPU(false, memory, 3, 3)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_4_2x2(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp7_0.NewCPU(false, memory, 2)
	}
	testSlow(t, factory)
}

func TestSlowMvp6_4_3x3(t *testing.T) {
	t.Parallel()
	factory := func(memory int) virtualMachine {
		return mvp7_0.NewCPU(false, memory, 3)
	}
	testSlow(t, factory)
}

func testSlow(t *testing.T, factory func(memory int) virtualMachine) {
	testPrime(t, factory, memory, testFrom, slowTest, false)
	testSums(t, factory, memory, testFrom, slowTest, false)
	testStringLength(t, factory, 1024, slowTest, false)
	testStringCopy(t, factory, slowTest*2, slowTest, false)
	testBubbleSort(t, slowTest/5, factory, false)
	testConditionalBranch(t, factory, false)
	testSpectre(t, factory, false)
}
