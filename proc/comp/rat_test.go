package comp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

func TestRat(t *testing.T) {
	rat := comp.NewRAT(2)
	v, exists := rat.Read(risc.T0)
	assert.False(t, exists)

	rat.Write(risc.T0, 1)
	v, exists = rat.Read(risc.T0)
	assert.True(t, exists)
	assert.Equal(t, int32(1), v)

	rat.Write(risc.T0, 2)
	v, exists = rat.Read(risc.T0)
	assert.True(t, exists)
	assert.Equal(t, int32(2), v)

	rat.Write(risc.T1, 3)
	v, exists = rat.Read(risc.T1)
	assert.True(t, exists)
	assert.Equal(t, int32(3), v)

	rat.Write(risc.T2, 3)
	v, exists = rat.Read(risc.T0)
	assert.False(t, exists)
}
