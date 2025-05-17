package observer

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type Observer[T any] interface {
	StartObserve(ctx context.Context, period time.Duration)
}

type observerImpl[T any] struct {
	queue        Queue[T]
	eventHandler EventHandler[T]
}

type EventHandler[T any] func(ctx context.Context, events []T) error

func NewObserver[T any](queue Queue[T], eventHandler EventHandler[T]) Observer[T] {
	o := &observerImpl[T]{
		queue:        queue,
		eventHandler: eventHandler,
	}
	return o
}

func (o *observerImpl[T]) StartObserve(ctx context.Context, period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			items := o.queue.Flush()
			if len(items) > 0 {
				err := o.eventHandler(ctx, items)
				if err != nil {
					logrus.WithError(err).
						Error("golangfuse: error in event handler")
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
