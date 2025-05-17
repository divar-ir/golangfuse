package golangfuse

import (
	"time"

	"github.com/divar-ir/golangfuse/internal/constants"
)

type Trace struct {
	ID        string     `json:"id,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
	Name      string     `json:"name,omitempty"`
	UserID    string     `json:"userId,omitempty"`
	Input     any        `json:"input,omitempty"`
	Output    any        `json:"output,omitempty"`
	SessionID string     `json:"sessionId,omitempty"`
	Release   string     `json:"release,omitempty"`
	Version   string     `json:"version,omitempty"`
	Metadata  any        `json:"metadata,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Public    bool       `json:"public,omitempty"`
}

type Ingestion struct {
	Batch []IngestionEvent `json:"batch"`
}

type IngestionEvent struct {
	ID        string                       `json:"id"`
	Timestamp time.Time                    `json:"timestamp"`
	Type      constants.IngestionEventType `json:"type"`
	Body      any                          `json:"body"`
}

type PromptItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatPrompt struct {
	ID        string       `json:"id"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	ProjectID string       `json:"projectId"`
	CreatedBy string       `json:"createdBy"`
	Prompt    []PromptItem `json:"prompt"`
	Name      string       `json:"name"`
	Version   int          `json:"version"`
	Type      string       `json:"type"`
	Config    any          `json:"config"`
	Tags      []string     `json:"tags"`
	Labels    []string     `json:"labels"`
}
