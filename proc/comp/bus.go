package comp

type Bus[T any] interface {
	Flush()
	Add(t T, currentCycle float32)
	Get() T
	Peek() T
	IsBufferFull() bool
	IsEmpty() bool
	IsElementInBuffer() bool
	IsElementInQueue() bool
	Connect(currentCycle float32)
}

type BufferEntry[T any] struct {
	availableFromCycle float32
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

func (bus *BufferedBus[T]) Add(t T, currentCycle float32) {
	bus.buffer = append(bus.buffer, BufferEntry[T]{
		availableFromCycle: currentCycle + 1,
		t:                  t,
	})
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

func (bus *BufferedBus[T]) Connect(currentCycle float32) {
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
