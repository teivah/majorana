package cache_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teivah/majorana/common/cache"
)

func TestLRUCache_Get(t *testing.T) {
	c := cache.NewLRUCache[int, int](3)
	c.Put(1, 1)
	c.Put(2, 2)
	c.Put(3, 3)

	v, exists := c.Find([]int{1, 2, 3})
	assert.True(t, exists)
	assert.Equal(t, 1, v)

	v, exists = c.Find([]int{1, 2, 3})
	assert.True(t, exists)
	assert.Equal(t, 2, v)

	v, exists = c.Find([]int{1, 2, 3})
	assert.True(t, exists)
	assert.Equal(t, 3, v)

	v, exists = c.Find([]int{1, 2, 3})
	assert.True(t, exists)
	assert.Equal(t, 1, v)

	_, exists = c.Find([]int{4})
	assert.False(t, exists)
}
