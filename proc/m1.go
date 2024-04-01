package proc

import (
	"testing"
)

const (
	m1Frequency        = 3_200_000_000
	secondToNanosecond = 1_000_000_000

	m1PrimeExecutionTime = 70.29
	m1SumsExecutionTime  = 1300.
)

func primeStats(t *testing.T, cycles int) {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1PrimeExecutionTime
	t.Logf("%d cycles, %.2f nanoseconds, %.1f slower\n", cycles, ns, slower)
}

func sumStats(t *testing.T, cycles int) {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1SumsExecutionTime
	t.Logf("%d cycles, %.2f nanoseconds, %.1f slower\n", cycles, ns, slower)
}
