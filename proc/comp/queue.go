package comp

import (
	"container/list"
)

type Queue[T any] struct {
	queue  *list.List
	length int
}

func NewQueue[T any](length int) *Queue[T] {
	return &Queue[T]{queue: list.New(), length: length}
}

func (q *Queue[T]) Push(value T) {
	q.queue.PushBack(value)
}

func (q *Queue[T]) Length() int {
	return q.queue.Len()
}

func (q *Queue[T]) IsFull() bool {
	return q.queue.Len() >= q.length
}

func (q *Queue[T]) Iterator() <-chan *list.Element {
	iter := make(chan *list.Element, q.queue.Len())
	go func() {
		defer close(iter)
		for e := q.queue.Front(); e != nil; {
			current := e
			e = e.Next()
			iter <- current
		}
	}()
	return iter
}

func (q *Queue[T]) Value(elem *list.Element) T {
	return elem.Value.(T)
}

func (q *Queue[T]) Remove(elem *list.Element) {
	q.queue.Remove(elem)
}
