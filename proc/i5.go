package proc

import (
	"fmt"
)

const i57360u = 2_300_000_000
const secondToNanosecond = 1_000_000_000
const i5ExecutionTime = 253

func stats(cycles float32) {
	s := cycles / i57360u
	ns := s * secondToNanosecond
	faster := ns / i5ExecutionTime
	fmt.Printf("%.2f cycles, %.2f nanoseconds, %.1f faster\n", cycles, ns, faster)
}
