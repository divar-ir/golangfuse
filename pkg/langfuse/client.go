package langfuse

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/divar-ir/golangfuse/internal/constants"
	"github.com/divar-ir/golangfuse/internal/models"
	"github.com/divar-ir/golangfuse/internal/observer"
	"github.com/divar-ir/golangfuse/pkg/errs"
	"github.com/google/uuid"
	"resty.dev/v3"
)

const DefaultTimeout = 800 * time.Millisecond

type Client interface {
	StartSendingEvents(ctx context.Context, period time.Duration) error
	Trace(input, output any, options ...Option)
	GetPromptTemplate(ctx context.Context, promptName string) (string, error)
}

type clientImpl struct {
	restClient             *resty.Client
	eventObserver          observer.Observer[models.IngestionEvent]
	eventQueue             observer.Queue[models.IngestionEvent]
	isSendingEventsStarted atomic.Bool
	endpoint               string
	promptLabel            string
}

func New(endpoint, publicKey, secretKey string) Client {
	return NewWithHttpClient(http.DefaultClient, endpoint, publicKey, secretKey)
}

func NewWithHttpClient(httpClient *http.Client, endpoint, publicKey, secretKey string) Client {
	client := resty.NewWithClient(httpClient).SetBasicAuth(publicKey, secretKey)
	c := &clientImpl{
		restClient:  client,
		eventQueue:  observer.NewQueue[models.IngestionEvent](),
		endpoint:    endpoint,
		promptLabel: "production", // TODO: use option pattern to override this default if needed
	}
	c.eventObserver = observer.NewObserver[models.IngestionEvent](c.eventQueue, c.sendEvents)
	return c
}

func (c *clientImpl) StartSendingEvents(ctx context.Context, period time.Duration) error {
	if c.isSendingEventsStarted.CompareAndSwap(false, true) {
		go c.eventObserver.StartObserve(ctx, period)
		return nil
	} else {
		return errs.AlreadyStartedErr
	}
}

func (c *clientImpl) GetPromptTemplate(ctx context.Context, promptName string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()
	promptObject := models.ChatPrompt{}
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

func (c *clientImpl) Trace(input, output any, options ...Option) {
	trace := &models.Trace{
		Input:  input,
		Output: output,
	}
	for _, opt := range options {
		opt(trace)
	}
	c.eventQueue.Enqueue(models.IngestionEvent{
		ID:        uuid.NewString(),
		Timestamp: time.Now(),
		Type:      constants.IngestionEventTypeTraceCreate,
		Body:      trace,
	})
}

func (c *clientImpl) sendEvents(ctx context.Context, events []models.IngestionEvent) error {
	i := &models.Ingestion{
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
