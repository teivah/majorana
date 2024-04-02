package comp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teivah/majorana/proc/comp"
)

func TestSimpleBus(t *testing.T) {
	b := &comp.SimpleBus[int]{}
	var val int
	var exists bool

	val, exists = b.Get()
	busAssert(t, 0, false, val, exists)

	assert.Equal(t, true, b.CanAdd())
	b.Add(1)
	val, exists = b.Get()
	busAssert(t, 0, false, val, exists)
	val, exists = b.Get()
	busAssert(t, 1, true, val, exists)
	val, exists = b.Get()
	busAssert(t, 0, false, val, exists)

	assert.Equal(t, true, b.CanAdd())
	b.Add(1)
	assert.Equal(t, false, b.CanAdd())
	val, exists = b.Get()
	busAssert(t, 0, false, val, exists)
	assert.Equal(t, true, b.CanAdd())
	b.Add(2)
	val, exists = b.Get()
	busAssert(t, 1, true, val, exists)
	val, exists = b.Get()
	busAssert(t, 2, true, val, exists)
	val, exists = b.Get()
	busAssert(t, 0, false, val, exists)
}

func busAssert(t *testing.T, expectedVal int, expectedExists bool, val int, exists bool) {
	assert.Equal(t, expectedVal, val)
	assert.Equal(t, expectedExists, exists)
}
