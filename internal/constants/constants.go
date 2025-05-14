package constants

type IngestionEventType string

const (
	IngestionEventTypeTraceCreate      IngestionEventType = "trace-create"
	IngestionEventTypeGenerationCreate IngestionEventType = "generation-create"
	IngestionEventTypeGenerationUpdate IngestionEventType = "generation-update"
	IngestionEventTypeScoreCreate      IngestionEventType = "score-create"
	IngestionEventTypeSpanCreate       IngestionEventType = "span-create"
	IngestionEventTypeSpanUpdate       IngestionEventType = "span-update"
	IngestionEventTypeEventCreate      IngestionEventType = "event-create"
)
