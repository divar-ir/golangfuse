package golangfuse

import (
	"net/http"
	"time"

	"resty.dev/v3"
)

type Option func(l *langfuseImpl)

func WithDefaultPromptLabel(label string) Option {
	return func(l *langfuseImpl) {
		l.promptLabel = label
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(l *langfuseImpl) {
		l.restClient = resty.NewWithClient(httpClient)
	}
}

func WithFlushPeriod(period time.Duration) Option {
	return func(l *langfuseImpl) {
		if period <= 0 {
			period = 1 * time.Second // Default to 1 second if invalid period is provided
		}
		l.flushPeriod = period
	}
}
