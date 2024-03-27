package proc

import (
	"fmt"
)

const i57360u = 2_300_000_000
const secondToNanosecond = 1_000_000_000

func stats(test string, cycles float32) {
	s := cycles / i57360u
	ns := s * secondToNanosecond
	fmt.Printf("%s: %.2f cycles, %.2f nanoseconds\n", test, cycles, ns)
}
