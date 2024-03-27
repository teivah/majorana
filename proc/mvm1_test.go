package proc

import (
	"testing"

	"github.com/teivah/ettore/risc"
	"github.com/teivah/ettore/test"
)

func TestPrimeNumber(t *testing.T) {
	bits := risc.BytesFromLowBits(1109)
	test.RunAssert(t, map[risc.RegisterType]int32{}, 5,
		map[int]int8{0: bits[0], 1: bits[1], 2: bits[2], 3: bits[3]},
		test.ReadFile(t, "../res/prime-number.asm"),
		map[risc.RegisterType]int32{risc.A0: 4},
		map[int]int8{4: 1})
}
