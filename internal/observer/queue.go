package observer

import "sync"

type Queue[T any] interface {
	Enqueue(event T)
	Flush() []T
}

type queueImpl[T any] struct {
	items []T
	mu    sync.Mutex
}

func NewQueue[T any]() Queue[T] {
	return &queueImpl[T]{}
}

func (q *queueImpl[T]) Enqueue(event T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, event)
}

func (q *queueImpl[T]) Flush() []T {
	q.mu.Lock()
	defer q.mu.Unlock()
	items := q.items
	q.items = []T{}
	return items
}
