package comp

type Bus[T any] interface {
	Flush()
	Add(t T, currentCycle int)
	DeleteLast()
	Get() T
	Peek() T
	IsBufferFull() bool
	IsEmpty() bool
	IsElementInBuffer() bool
	IsElementInQueue() bool
	Connect(currentCycle int)
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

func (bus *BufferedBus[T]) Flush() {
	bus.buffer = make([]BufferEntry[T], 0)
	bus.queue = make([]T, 0)
}

func (bus *BufferedBus[T]) Add(t T, currentCycle int) {
	bus.buffer = append(bus.buffer, BufferEntry[T]{
		availableFromCycle: currentCycle + 1,
		t:                  t,
	})
}

func (bus *BufferedBus[T]) DeleteLast() {
	if len(bus.buffer) == 0 {
		return
	}
	bus.buffer = bus.buffer[:len(bus.buffer)-1]
}

func (bus *BufferedBus[T]) Get() T {
	elem := bus.queue[0]
	bus.queue = bus.queue[1:]
	return elem
}

func (bus *BufferedBus[T]) Peek() T {
	return bus.queue[0]
}

func (bus *BufferedBus[T]) IsBufferFull() bool {
	return len(bus.buffer) == bus.bufferLength
}

func (bus *BufferedBus[T]) IsEmpty() bool {
	return len(bus.queue) == 0 && len(bus.buffer) == 0
}

func (bus *BufferedBus[T]) IsElementInBuffer() bool {
	return len(bus.buffer) != 0
}

func (bus *BufferedBus[T]) IsElementInQueue() bool {
	return len(bus.queue) != 0
}

func (bus *BufferedBus[T]) Connect(currentCycle int) {
	if len(bus.queue) == bus.queueLength {
		return
	}

	i := 0
	for ; i < len(bus.buffer); i++ {
		if len(bus.queue) == bus.queueLength {
			break
		}
		entry := bus.buffer[i]
		if entry.availableFromCycle > currentCycle {
			break
		}
		bus.queue = append(bus.queue, entry.t)
	}
	bus.buffer = bus.buffer[i:]
}

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

func (b *SimpleBus[T]) Peek() (t T, res bool) {
	if b.current.exists {
		t = b.current.t
		res = true
	}
	return
}

func (b *SimpleBus[T]) Connect() {
	if b.current.exists {
		return
	}
	b.current = b.pending
	b.pending = entry[T]{}
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
