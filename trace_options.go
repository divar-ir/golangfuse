package golangfuse

type TraceOption func(t *Trace)

func WithTags(tags ...string) TraceOption {
	return func(t *Trace) {
		t.Tags = tags
	}
}

func WithSessionID(sessionID string) TraceOption {
	return func(t *Trace) {
		t.SessionID = sessionID
	}
}
