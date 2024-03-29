package mvm4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBtb(t *testing.T) {
	btb := newBranchTargetBuffer(2)
	assert.Equal(t, nilPtr(), btb.get(1))
	btb.add(1, 10)
	assert.Equal(t, ptr(10), btb.get(1))
	btb.add(2, 20)
	assert.Equal(t, ptr(20), btb.get(2))
	btb.add(2, 21)
	assert.Equal(t, ptr(21), btb.get(2))
	btb.add(3, 30)
	assert.Equal(t, ptr(30), btb.get(3))
	assert.Equal(t, nilPtr(), btb.get(1))
}

func nilPtr() *int32 {
	return nil
}

func ptr(v int32) *int32 {
	return &v
}
