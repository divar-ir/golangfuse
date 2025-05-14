package langfuse

import "github.com/divar-ir/golangfuse/internal/models"

type Option func(t *models.Trace)

func WithTags(tags ...string) Option {
	return func(t *models.Trace) {
		t.Tags = tags
	}
}

func WithSessionID(sessionID string) Option {
	return func(t *models.Trace) {
		t.SessionID = sessionID
	}
}
