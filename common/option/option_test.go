package option

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptional_Get(t *testing.T) {
	v, exists := Of[int](1).Get()
	assert.Equal(t, 1, v)
	assert.True(t, exists)

	v, exists = None[int]().Get()
	assert.False(t, exists)
}
