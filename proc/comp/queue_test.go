package comp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/teivah/majorana/proc/comp"
)

func TestQueue(t *testing.T) {
	q := comp.NewQueue[int]()
	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Push(4)
	q.Push(5)
	assert.Equal(t, 5, q.Length())

	var got []int
	for elem := range q.Iterator() {
		v := q.Value(elem)
		got = append(got, v)
		if v == 5 {
			q.Remove(elem)
		}
	}
	assert.Equal(t, []int{1, 2, 3, 4, 5}, got)

	// Check tail deletion
	got = nil
	for elem := range q.Iterator() {
		v := q.Value(elem)
		got = append(got, v)
		if v == 1 {
			q.Remove(elem)
		}
	}
	assert.Equal(t, []int{1, 2, 3, 4}, got)

	// Check head deletion
	got = nil
	for elem := range q.Iterator() {
		v := q.Value(elem)
		got = append(got, v)
		if v == 3 {
			q.Remove(elem)
		}
	}
	assert.Equal(t, []int{2, 3, 4}, got)

	// Check middle deletion
	got = nil
	for elem := range q.Iterator() {
		v := q.Value(elem)
		got = append(got, v)
		q.Remove(elem)
	}
	assert.Equal(t, []int{2, 4}, got)

	// Check remove all
	got = nil
	for elem := range q.Iterator() {
		v := q.Value(elem)
		got = append(got, v)
	}
	assert.Equal(t, 0, len(got))
}
