package co

type Coroutine[A, B any] struct {
	start   func(A) B
	current func(A) B
	isStart bool
	pre     func(A)
}

func New[A, B any](f func(A) B) Coroutine[A, B] {
	return Coroutine[A, B]{
		start:   f,
		current: f,
		isStart: true,
	}
}

func (c *Coroutine[A, B]) Pre(f func(A)) {
	c.pre = f
}

func (c *Coroutine[A, B]) Cycle(a A) B {
	if c.pre != nil {
		c.pre(a)
	}
	return c.current(a)
}

func (c *Coroutine[A, B]) Checkpoint(f func(A) B) {
	c.current = f
	c.isStart = false
}

func (c *Coroutine[A, B]) ExecuteWithCheckpoint(a A, f func(A) B) B {
	c.current = f
	c.isStart = false
	return f(a)
}

func (c *Coroutine[A, B]) Reset() {
	c.current = c.start
	c.isStart = true
}

func (c *Coroutine[A, B]) IsStart() bool {
	return c.isStart
}
