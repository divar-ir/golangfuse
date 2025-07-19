package golangfuse

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type EventsFlushFunc func(ctx context.Context, events []IngestionEvent) error

type eventBuffer struct {
	bufferedEvents []IngestionEvent
	mu             sync.Mutex
	flushHandler   EventsFlushFunc
}

func newEventBufferer(flushHandler EventsFlushFunc) *eventBuffer {
	return &eventBuffer{flushHandler: flushHandler}
}

func (i *eventBuffer) Start(ctx context.Context, period time.Duration) {
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			i.Flush(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (i *eventBuffer) Add(event IngestionEvent) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.bufferedEvents = append(i.bufferedEvents, event)
}

func (i *eventBuffer) Flush(ctx context.Context) {
	i.mu.Lock()
	items := i.bufferedEvents
	i.bufferedEvents = nil
	i.mu.Unlock()
	if len(items) > 0 {
		err := i.flushHandler(ctx, items)
		if err != nil {
			logrus.WithError(err).Error("golangfuse: error in event flush handler")
		} else {
			logrus.Tracef("golangfuse: flushed %d events", len(items))
		}
	}
}
