package proc

import (
	"fmt"
)

const (
	m1Frequency        = 3_200_000_000
	secondToNanosecond = 1_000_000_000

	m1PrimeExecutionTime        = 31703.
	m1SumsExecutionTime         = 1300.
	m1StringCopyExecutionTime   = 3232.
	m1StringLengthExecutionTime = 3231.
)

func primeStats(cycles int) string {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1PrimeExecutionTime
	return fmt.Sprintf("%.0f ns, %.1f slower", ns, slower)
}

func sumStats(cycles int) string {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1SumsExecutionTime
	return fmt.Sprintf("%.0f ns, %.1f slower", ns, slower)
}

func stringCopyStats(cycles int) string {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1StringCopyExecutionTime
	return fmt.Sprintf("%.0f ns, %.1f slower", ns, slower)
}

func stringLengthStats(cycles int) string {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1StringLengthExecutionTime
	return fmt.Sprintf("%.0f ns, %.1f slower", ns, slower)
}
