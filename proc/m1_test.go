package proc

const (
	m1Frequency        = 3_200_000_000
	secondToNanosecond = 1_000_000_000

	m1PrimeExecutionTime        = 31703.
	m1SumsExecutionTime         = 1300.
	m1StringCopyExecutionTime   = 3232.
	m1StringLengthExecutionTime = 3231.
	m1BubbleSortExecutionTime   = 42182.
)

type benchResult struct {
	executionNs float64
	slower      float64
}

func primeStats(cycles int) benchResult {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1PrimeExecutionTime
	return benchResult{ns, slower}
}

func sumStats(cycles int) benchResult {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1SumsExecutionTime
	return benchResult{ns, slower}
}

func stringCopyStats(cycles int) benchResult {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1StringCopyExecutionTime
	return benchResult{ns, slower}
}

func stringLengthStats(cycles int) benchResult {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1StringLengthExecutionTime
	return benchResult{ns, slower}
}

func bubbleSortStats(cycles int) benchResult {
	s := float64(cycles) / m1Frequency
	ns := s * secondToNanosecond
	slower := ns / m1BubbleSortExecutionTime
	return benchResult{ns, slower}
}
