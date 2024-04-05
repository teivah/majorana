package comp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := NewLRUCache(2, 6)
	as := getAssert(t, c)

	as(0, 0, false)
	c.Push(0, []int8{0, 1})
	as(1, 1, true)
	c.Push(2, []int8{2, 3})
	c.Push(4, []int8{4, 5})
	as(0, 0, true)
	as(1, 1, true)
	as(2, 2, true)
	as(5, 5, true)
	as(6, 0, false)

	// Should move it to first
	as(5, 5, true)
	c.Push(6, []int8{6, 7})
	c.Push(8, []int8{8, 9})
	as(5, 5, true)
}

func getAssert(t *testing.T, c *LRUCache) func(int32, int8, bool) {
	t.Helper()
	return func(addr int32, i int8, b bool) {
		val, exists := c.Get(addr)
		assert.Equal(t, i, val)
		assert.Equal(t, b, exists)
	}
}
