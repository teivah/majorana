package comp

type BufferEntry[T any] struct {
	availableFromCycle float32
	t                  T
}

type Bus[T any] struct {
	buffer       []BufferEntry[T]
	queue        []T
	queueLength  int
	bufferLength int
}

func NewBus[T any](queueLength, bufferLength int) *Bus[T] {
	return &Bus[T]{
		buffer:       make([]BufferEntry[T], 0),
		queue:        make([]T, 0),
		queueLength:  queueLength,
		bufferLength: bufferLength,
	}
}

func (bus *Bus[T]) Flush() {
	bus.buffer = make([]BufferEntry[T], 0)
	bus.queue = make([]T, 0)
}

func (bus *Bus[T]) Add(t T, currentCycle float32) {
	bus.buffer = append(bus.buffer, BufferEntry[T]{
		availableFromCycle: currentCycle + 1,
		t:                  t,
	})
}

func (bus *Bus[T]) Get() T {
	elem := bus.queue[0]
	bus.queue = bus.queue[1:]
	return elem
}

func (bus *Bus[T]) Peek() T {
	return bus.queue[0]
}

func (bus *Bus[T]) IsBufferFull() bool {
	return len(bus.buffer) == bus.bufferLength
}

func (bus *Bus[T]) IsEmpty() bool {
	return len(bus.queue) == 0 && len(bus.buffer) == 0
}

func (bus *Bus[T]) ContainsElementInBuffer() bool {
	return len(bus.buffer) != 0
}

func (bus *Bus[T]) ContainsElementInQueue() bool {
	return len(bus.queue) != 0
}

func (bus *Bus[T]) Connect(currentCycle float32) {
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
