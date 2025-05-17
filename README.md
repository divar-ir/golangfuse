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

client := langfuse.New(
"https://your-langfuse-instance.com",
"your-public-key",
"your-secret-key",
)
ctx := context.Background()
err := client.StartSendingEvents(ctx, 5*time.Second)
if err != nil {
    // handle error
}

// Get a prompt template
promptTemplate, err := client.GetPromptTemplate(ctx, "your-prompt-name")

// Call LLM API and get the response
// ...
client.Trace("user input", "model output")
```

## Development

### Running Tests

```
go test ./...
```

## License

[License information](/LICENSE)
