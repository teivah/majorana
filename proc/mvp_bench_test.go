package proc

import (
	"testing"
)

var globalBool bool
var globalInt int
var globalInts32 []int32
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

func BenchmarkStringLength(b *testing.B) {
	b.StopTimer()
	// We recreate the slice to prevent CPU cache hit
	src := make([]byte, 0, benchStringLength)
	for i := 0; i < benchStringLength; i++ {
		src = append(src, byte(i))
	}
	b.StartTimer()

	var local int
	for i := 0; i < b.N; i++ {
		local = strlen(src)
	}
	globalInt = local
}

func BenchmarkBubbleSort(b *testing.B) {
	b.StopTimer()
	// We recreate the slice to prevent CPU cache hit
	src := make([]int32, 0, benchBubSort)
	for i := 0; i < benchBubSort; i++ {
		src = append(src, int32(benchBubSort-i))
	}
	b.StartTimer()

	var local []int32
	for i := 0; i < b.N; i++ {
		bubsort(src, len(src))
		local = src
	}
	globalInts32 = local
}
