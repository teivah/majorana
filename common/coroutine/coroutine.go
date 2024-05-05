package co

import (
	"slices"
)

type Coroutine[A, B any] struct {
	start   func(A) B
	current func(A) B
	list    []func(A) bool
	isStart bool
	// Return true if should stop
	pre func(A) bool
}

func New[A, B any](f func(A) B) Coroutine[A, B] {
	return Coroutine[A, B]{
		start:   f,
		current: f,
		isStart: true,
	}
}

func (c *Coroutine[A, B]) Pre(f func(A) bool) {
	c.pre = f
}

func (c *Coroutine[A, B]) Cycle(a A) B {
	var zero B
	if c.pre != nil {
		if c.pre(a) {
			return zero
		}
	}

	if !c.isStart {
		// If there's already an execution going on, we finish it before to tackle
		// the pre-actions
		return c.current(a)
	}

	length := len(c.list)
	c.list = slices.DeleteFunc(c.list, func(f func(A) bool) bool {
		return f(a)
	})
	if length == 0 {
		return c.current(a)
	}
	return zero
}

func (c *Coroutine[A, B]) Checkpoint(f func(A) B) {
	c.current = f
	c.isStart = false
}

func (c *Coroutine[A, B]) Append(f func(A) bool) {
	c.list = append(c.list, f)
}

func (c *Coroutine[A, B]) ExecuteWithCheckpoint(a A, f func(A) B) B {
	c.current = f
	c.isStart = false
	return f(a)
}

func (c *Coroutine[A, B]) ExecuteWithCheckpointAfter(a A, cycles int, f func(A) B) B {
	c.current = f
	c.isStart = false
	remaining := cycles
	var zero B
	return c.ExecuteWithCheckpoint(a, func(a A) B {
		if remaining > 0 {
			remaining--
			return zero
		}
		return f(a)
	})
}

func (c *Coroutine[A, B]) Reset() {
	c.current = c.start
	c.isStart = true
}

func (c *Coroutine[A, B]) ExecuteWithReset(a A, f func(A) B) B {
	c.Reset()
	return f(a)
}

func (c *Coroutine[A, B]) IsStart() bool {
	return c.isStart && len(c.list) == 0
}
