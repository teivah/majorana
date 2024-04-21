package comp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teivah/majorana/proc/comp"
	"github.com/teivah/majorana/risc"
)

func TestRat(t *testing.T) {
	rat := comp.NewRAT[risc.RegisterType, int32](2)
	v, exists := rat.Read(risc.T0)
	assert.False(t, exists)

	rat.Write(risc.T0, 1)
	v, exists = rat.Read(risc.T0)
	assert.True(t, exists)
	assert.Equal(t, int32(1), v)
	assert.Equal(t, map[risc.RegisterType]int32{
		risc.T0: 1,
	}, rat.Values())

	rat.Write(risc.T0, 2)
	v, exists = rat.Read(risc.T0)
	assert.True(t, exists)
	assert.Equal(t, int32(2), v)
	assert.Equal(t, map[risc.RegisterType]int32{
		risc.T0: 2,
	}, rat.Values())
}
