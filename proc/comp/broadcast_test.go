package comp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teivah/majorana/proc/comp"
)

func TestBroadcast(t *testing.T) {
	b := comp.NewBroadcast[int](3)

	assert.Equal(t, []int{}, read(b.Read(0)))

	b.Notify(0)
	assert.Equal(t, []int{0}, read(b.Read(0)))
	assert.Equal(t, []int{0}, read(b.Read(1)))
	assert.Equal(t, []int{0}, read(b.Read(2)))

	b.Read(0)[0].Commit()
	assert.Equal(t, []int{}, read(b.Read(0)))
	assert.Equal(t, []int{0}, read(b.Read(1)))
	assert.Equal(t, []int{0}, read(b.Read(2)))
}

func read(events []comp.Event[int]) []int {
	res := make([]int, 0, len(events))
	for _, event := range events {
		res = append(res, event.Data)
	}
	return res
}
