package comp

import "slices"

type Broadcast[T any] struct {
	count     int
	listeners [][]event[T]
}

func NewBroadcast[T any](count int) *Broadcast[T] {
	return &Broadcast[T]{
		count:     count,
		listeners: make([][]event[T], count),
	}
}

type event[T any] struct {
	data T
	read bool
}

func (b *Broadcast[T]) Notify(t T) {
	for i := range b.listeners {
		b.listeners[i] = append(b.listeners[i], event[T]{data: t})
	}
}

type Event[T any] struct {
	Data   T
	Commit func()
}

func (b *Broadcast[T]) Read(id int) []Event[T] {
	// Clean
	b.listeners[id] = slices.DeleteFunc(b.listeners[id], func(e event[T]) bool {
		return e.read
	})
	var events []Event[T]
	for i, event := range b.listeners[id] {
		events = append(events, Event[T]{
			Data: event.data,
			Commit: func() {
				b.listeners[id][i].read = true
			},
		})
	}
	return events
}
