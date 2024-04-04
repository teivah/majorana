package proc

import (
	"testing"
)

var globalBool bool
var globalInt int
var globalBytes []byte

func BenchmarkPrime(b *testing.B) {
	var local bool
	for i := 0; i < b.N; i++ {
		local = isPrime(benchPrimeNumber)
	}
	globalBool = local
}

func BenchmarkSums(b *testing.B) {
	b.StopTimer()
	// We recreate the slice to prevent CPU cache hit
	n := benchSums
	s := make([]int, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	b.StartTimer()

	var local int
	for i := 0; i < b.N; i++ {
		local = sumArray(s)
	}
	globalInt = local
}

func BenchmarkStringCopy(b *testing.B) {
	b.StopTimer()
	// We recreate the slice to prevent CPU cache hit
	src := make([]byte, 0, benchStringCopy)
	for i := 0; i < benchStringCopy; i++ {
		src = append(src, byte(i))
	}
	dst := make([]byte, benchStringCopy)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		strncpy(dst, src, benchStringCopy)
	}
	globalBytes = dst
}
