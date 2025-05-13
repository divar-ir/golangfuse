package client

import "time"

type promptItem struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatPrompt struct {
	ID        string       `json:"id"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	ProjectID string       `json:"projectId"`
	CreatedBy string       `json:"createdBy"`
	Prompt    []promptItem `json:"prompt"`
	Name      string       `json:"name"`
	Version   int          `json:"version"`
	Type      string       `json:"type"`
	Config    any          `json:"config"`
	Tags      []string     `json:"tags"`
	Labels    []string     `json:"labels"`
}
