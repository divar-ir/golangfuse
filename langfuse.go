package golangfuse

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/divar-ir/golangfuse/internal/constants"
	"github.com/divar-ir/golangfuse/internal/observer"
	"github.com/google/uuid"
	"resty.dev/v3"
)

type Langfuse interface {
	StartSendingEvents(ctx context.Context, period time.Duration) error
	Trace(input, output any, options ...TraceOption)
	GetPromptTemplate(ctx context.Context, promptName string) (string, error)
}

type langfuseImpl struct {
	restClient             *resty.Client
	eventObserver          observer.Observer[IngestionEvent]
	eventQueue             observer.Queue[IngestionEvent]
	isSendingEventsStarted atomic.Bool
	endpoint               string
	promptLabel            string
}

func New(endpoint, publicKey, secretKey string) Langfuse {
	return NewWithHttpClient(http.DefaultClient, endpoint, publicKey, secretKey)
}

func NewWithHttpClient(httpClient *http.Client, endpoint, publicKey, secretKey string) Langfuse {
	client := resty.NewWithClient(httpClient).SetBasicAuth(publicKey, secretKey)
	c := &langfuseImpl{
		restClient:  client,
		eventQueue:  observer.NewQueue[IngestionEvent](),
		endpoint:    endpoint,
		promptLabel: "production", // TODO: use option pattern to override this default if needed
	}
	c.eventObserver = observer.NewObserver[IngestionEvent](c.eventQueue, c.sendEvents)
	return c
}

func (c *langfuseImpl) StartSendingEvents(ctx context.Context, period time.Duration) error {
	if c.isSendingEventsStarted.CompareAndSwap(false, true) {
		go c.eventObserver.StartObserve(ctx, period)
		return nil
	} else {
		return AlreadyStartedErr
	}
}

func (c *langfuseImpl) GetPromptTemplate(ctx context.Context, promptName string) (string, error) {
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
	return convertJinjaVariablesToGoTemplate(promptObject.Prompt[0].Content), nil
}

func (c *langfuseImpl) Trace(input, output any, options ...TraceOption) {
	trace := &Trace{
		Input:  input,
		Output: output,
	}
	for _, opt := range options {
		opt(trace)
	}
	c.eventQueue.Enqueue(IngestionEvent{
		ID:        uuid.NewString(),
		Timestamp: time.Now(),
		Type:      constants.IngestionEventTypeTraceCreate,
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

func convertJinjaVariablesToGoTemplate(prompt string) string {
	re := regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)
	return re.ReplaceAllString(prompt, "{{.$1}}")
}
