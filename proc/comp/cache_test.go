package comp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := NewLRUCache(2, 6)
	ass := getAssert(t, c)

	ass(0, 0, false)
	c.PushLine(0, []int8{0, 1})
	ass(1, 1, true)
	c.PushLine(2, []int8{2, 3})
	c.PushLine(4, []int8{4, 5})
	ass(0, 0, true)
	ass(1, 1, true)
	ass(2, 2, true)
	ass(5, 5, true)
	ass(6, 0, false)

	// Should move it to first
	ass(5, 5, true)
	c.PushLine(6, []int8{6, 7})
	c.PushLine(8, []int8{8, 9})
	ass(5, 5, true)

	v, exists := c.GetCacheLine(4)
	assert.True(t, exists)
	assert.Equal(t, []int8{4, 5}, v)

	c.EvictCacheLine(4)
	ass(4, 0, false)
	ass(5, 0, false)
	ass(6, 6, true)

	_, exists = c.GetCacheLine(4)
	assert.False(t, exists)
}

func getAssert(t *testing.T, c *LRUCache) func(int32, int8, bool) {
	t.Helper()
	return func(addr int32, i int8, b bool) {
		val, exists := c.Get(addr)
		assert.Equal(t, i, val)
		assert.Equal(t, b, exists)
	}
}
