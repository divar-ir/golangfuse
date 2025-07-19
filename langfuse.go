package golangfuse

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"resty.dev/v3"
)

type Langfuse interface {
	// Shutdown stops the event buffer and cleans up resources.
	// It also ensures that any pending events are sent before shutdown.
	Shutdown(ctx context.Context) error
	Trace(input, output any, options ...TraceOption)
	GetSystemPromptTemplate(ctx context.Context, promptName string) (string, error)
}

type langfuseImpl struct {
	restClient  *resty.Client
	eventBuffer *eventBuffer
	isShutdown  atomic.Bool
	cancelFunc  context.CancelFunc
	endpoint    string
	promptLabel string
	flushPeriod time.Duration
}

func New(
	ctx context.Context,
	endpoint,
	publicKey,
	secretKey string,
	opts ...Option,
) Langfuse {
	c := &langfuseImpl{
		restClient:  resty.New(),
		endpoint:    endpoint,
		promptLabel: "production",
		flushPeriod: 1 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.restClient = c.restClient.SetBasicAuth(publicKey, secretKey)
	c.eventBuffer = newEventBufferer(c.sendEvents)
	c.startSendingEvents(ctx, c.flushPeriod)
	return c
}

func (c *langfuseImpl) startSendingEvents(ctx context.Context, period time.Duration) {
	ctx, cancel := context.WithCancel(ctx)
	c.cancelFunc = cancel
	go c.eventBuffer.Start(ctx, period)
}

func (c *langfuseImpl) Shutdown(ctx context.Context) error {
	// Check if already shutdown
	if c.isShutdown.CompareAndSwap(false, true) {
		// Cancel the event buffer goroutine
		c.cancelFunc()

		// Flush any remaining events before shutdown
		c.eventBuffer.Flush(ctx)

		return nil
	} else {
		return AlreadyShutdownErr
	}
}

func (c *langfuseImpl) GetSystemPromptTemplate(ctx context.Context, promptName string) (string, error) {
	promptObject := ChatPrompt{}
	resp, err := c.restClient.R().
		SetContext(ctx).
		SetResult(&promptObject).
		Get(fmt.Sprintf("%s/api/public/v2/prompts/%s?label=%s", c.endpoint, promptName, c.promptLabel))
	if err != nil {
		return "", err
	}
	if resp.StatusCode() != 200 {
		return "", fmt.Errorf("unexpected status code (%d), response %s", resp.StatusCode(), resp.String())
	}
	if promptObject.Type != "chat" {
		return "", fmt.Errorf("unexpected prompt type: %s", promptObject.Type)
	}
	if len(promptObject.Prompt) == 0 {
		return "", fmt.Errorf("prompt is empty")
	}
	if promptObject.Prompt[0].Role != "system" {
		return "", fmt.Errorf("prompt role is not system")
	}
	return promptObject.Prompt[0].Content, nil
}

func (c *langfuseImpl) Trace(input, output any, options ...TraceOption) {
	trace := &Trace{
		Input:  input,
		Output: output,
	}
	for _, opt := range options {
		opt(trace)
	}
	c.eventBuffer.Add(IngestionEvent{
		ID:        uuid.NewString(),
		Timestamp: time.Now(),
		Type:      IngestionEventTypeTraceCreate,
		Body:      trace,
	})
}

func (c *langfuseImpl) sendEvents(ctx context.Context, events []IngestionEvent) error {
	i := &Ingestion{
		Batch: events,
	}
	resp, err := c.restClient.R().SetContext(ctx).SetBody(i).
		Post(fmt.Sprintf("%s/api/public/ingestion", c.endpoint))
	if err != nil {
		return fmt.Errorf("failed to send ingestion: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("failed to send ingestion (status = %d): %s", resp.StatusCode(), resp.String())
	}
	return nil
}
