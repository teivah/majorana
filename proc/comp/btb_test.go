package comp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtb(t *testing.T) {
	btb := NewBranchTargetBuffer(2)
	assert.Equal(t, nilPtr(), btb.Get(1))
	btb.Add(1, 10)
	assert.Equal(t, ptr(10), btb.Get(1))
	btb.Add(2, 20)
	assert.Equal(t, ptr(20), btb.Get(2))
	btb.Add(2, 21)
	assert.Equal(t, ptr(21), btb.Get(2))
	btb.Add(3, 30)
	assert.Equal(t, ptr(30), btb.Get(3))
	assert.Equal(t, nilPtr(), btb.Get(1))
}

func nilPtr() *int32 {
	return nil
}

func ptr(v int32) *int32 {
	return &v
}
