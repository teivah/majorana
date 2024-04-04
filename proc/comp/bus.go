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

type BufferEntry[T any] struct {
	availableFromCycle int
	t                  T
}

type BufferedBus[T any] struct {
	buffer       []BufferEntry[T]
	queue        []T
	queueLength  int
	bufferLength int
}

func NewBufferedBus[T any](queueLength, bufferLength int) *BufferedBus[T] {
	return &BufferedBus[T]{
		buffer:       make([]BufferEntry[T], 0),
		queue:        make([]T, 0),
		queueLength:  queueLength,
		bufferLength: bufferLength,
	}
}

func (b *BufferedBus[T]) InLength() int {
	return b.queueLength
}

func (b *BufferedBus[T]) OutLength() int {
	return b.bufferLength
}

func (b *BufferedBus[T]) Clean() {
	b.buffer = make([]BufferEntry[T], 0)
	b.queue = make([]T, 0)
}

func (b *BufferedBus[T]) Add(t T, currentCycle int) {
	b.buffer = append(b.buffer, BufferEntry[T]{
		availableFromCycle: currentCycle + 1,
		t:                  t,
	})
}

func (b *BufferedBus[T]) Revert(t T, currentCycle int) {
	b.buffer = append([]BufferEntry[T]{
		{
			availableFromCycle: currentCycle,
			t:                  t,
		},
	}, b.buffer...)
}

func (b *BufferedBus[T]) DeleteLast() {
	if len(b.buffer) == 0 {
		return
	}
	b.buffer = b.buffer[:len(b.buffer)-1]
}

func (b *BufferedBus[T]) Get() (T, bool) {
	var zero T
	if len(b.queue) == 0 {
		return zero, false
	}
	elem := b.queue[0]
	b.queue = b.queue[1:]
	return elem, true
}

func (b *BufferedBus[T]) CanGet() bool {
	return len(b.queue) != 0
}

func (b *BufferedBus[T]) CanAdd() bool {
	return len(b.buffer) != b.bufferLength
}

func (b *BufferedBus[T]) RemainingToAdd() int {
	return b.bufferLength - len(b.buffer)
}

func (b *BufferedBus[T]) PendingRead() int {
	return len(b.queue)
}

func (b *BufferedBus[T]) IsEmpty() bool {
	return len(b.queue) == 0 && len(b.buffer) == 0
}

func (b *BufferedBus[T]) Connect(currentCycle int) {
	if len(b.queue) == b.queueLength {
		return
	}

	i := 0
	for ; i < len(b.buffer); i++ {
		if len(b.queue) == b.queueLength {
			break
		}
		entry := b.buffer[i]
		if entry.availableFromCycle > currentCycle {
			break
		}
		b.queue = append(b.queue, entry.t)
	}
	b.buffer = b.buffer[i:]
}
