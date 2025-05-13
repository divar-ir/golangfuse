package client

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"resty.dev/v3"
)

const DefaultTimeout = 800 * time.Millisecond

type LangfuseClient interface {
	GetPromptTemplate(ctx context.Context, promptName string) (string, error)
}

type langfuseClient struct {
	restClient  *resty.Client
	endpoint    string
	promptLabel string
}

func New(endpoint, publicKey, secretKey string) LangfuseClient {
	return NewWithHttpClient(http.DefaultClient, endpoint, publicKey, secretKey)
}

func NewWithHttpClient(httpClient *http.Client, endpoint, publicKey, secretKey string) LangfuseClient {
	client := resty.NewWithClient(httpClient).
		SetBasicAuth(publicKey, secretKey)
	return &langfuseClient{
		restClient:  client,
		endpoint:    endpoint,
		promptLabel: "production", // TODO: use option pattern to override this default if needed
	}
}

func (c *langfuseClient) GetPromptTemplate(ctx context.Context, promptName string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()
	promptObject := chatPrompt{}
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

func convertJinjaVariablesToGoTemplate(prompt string) string {
	re := regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)
	return re.ReplaceAllString(prompt, "{{.$1}}")
}
