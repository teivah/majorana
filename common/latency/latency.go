// Package latency represents latency in cycles.
// Source https://www.7-cpu.com/cpu/Apple_M1.html
package latency

const (
	RegisterAccess = 1
	L1Access       = 3
	L2Access       = 18
	L3Access       = 18 + 32  // 18 + 10 ns
	MemoryAccess   = 18 + 291 // 19 + 91 ns
	Flush          = 1
)
