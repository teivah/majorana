package comp

type SimpleBus[T any] struct {
	pending entry[T]
	current entry[T]
}

type entry[T any] struct {
	exists bool
	t      T
}

func (b *SimpleBus[T]) Flush() {
	b.pending = entry[T]{}
	b.current = entry[T]{}
}

func (b *SimpleBus[T]) Get() (t T, exists bool) {
	if b.current.exists {
		t = b.current.t
		exists = true
	}
	b.current = b.pending
	b.pending = entry[T]{}
	return
}

func (b *SimpleBus[T]) CanAdd() bool {
	return !b.pending.exists
}

func (b *SimpleBus[T]) Add(t T) {
	b.pending = entry[T]{
		exists: true,
		t:      t,
	}
}

func (b *SimpleBus[T]) IsEmpty() bool {
	return !b.pending.exists && !b.current.exists
}

func (b *SimpleBus[T]) Clean() {
	b.pending = entry[T]{}
	b.current = entry[T]{}
}
