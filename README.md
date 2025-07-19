# Golangfuse

A Go SDK for [Langfuse](https://langfuse.com/), enabling prompt management and observability for LLM applications.

## Features

- **Prompt Template Management**: Fetch and use prompt templates from Langfuse
- **Observability**: Track and send events for LLM interactions
- **Batch Processing**: Efficiently send events in batches
- **Authentication**: Built-in support for API key authentication

## Installation

```bash
go get github.com/divar-ir/golangfuse
```

## Usage

```go
import "github.com/divar-ir/golangfuse/pkg/langfuse"

ctx := context.Background()
client := langfuse.New(
	ctx,
    "https://your-langfuse-instance.com",
    "your-public-key",
    "your-secret-key",
)

// Get a prompt template
promptTemplate, err := client.GetPromptTemplate(ctx, "your-prompt-name")

// Call LLM API and get the response
// ...
// Log the trace
client.Trace("user input", "model output")
```

## Development

### Running Tests

```
go test ./...
```

## License

[License information](/LICENSE)
