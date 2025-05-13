package client_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/divar-ir/go-http-mock/pkg/httpmock"
	"github.com/divar-ir/golangfuse/pkg/client"
	"github.com/stretchr/testify/suite"
)

type LangfuseClientTest struct {
	suite.Suite
	ctx context.Context
}

func (s *LangfuseClientTest) SetupTest() {
	s.ctx = context.Background()
}

func (s *LangfuseClientTest) TestGetPromptTemplateShouldParseApiResponse() {
	// Given
	const promptName = "test-prompt"
	c := s.getLangfuseClientForTest(promptName, "This is system prompt")

	// When
	promptTemplate, err := c.GetPromptTemplate(s.ctx, promptName)
	s.Require().NoError(err)
	s.Require().Equal("This is system prompt", promptTemplate)
}

func (s *LangfuseClientTest) TestGetPromptTemplateShouldReturnVariablesWithGolangFormat() {
	// Given
	const promptName = "test-prompt"
	jinjaVarialbeVariants := []string{
		"{{myVar}}",
		"{{myVar }}",
		"{{ myVar}}",
		"{{ myVar }}",
	}

	for _, v := range jinjaVarialbeVariants {
		c := s.getLangfuseClientForTest(promptName,
			fmt.Sprintf("This is system prompt %s variable.", v))

		// When
		promptTemplate, err := c.GetPromptTemplate(s.ctx, promptName)

		// Then
		s.Require().NoError(err)
		s.Require().Equal("This is system prompt {{.myVar}} variable.", promptTemplate)
	}
}

func (s *LangfuseClientTest) getLangfuseClientForTest(promptName, promptContent string) client.LangfuseClient {
	apiResponse := fmt.Sprintf(`{
  "id" : "id",
  "createdAt" : "2025-05-06T10:48:43.828Z",
  "updatedAt" : "2025-05-06T12:37:40.637Z",
  "projectId" : "project-id",
  "createdBy" : "creator-id",
  "prompt" : [ {
    "role" : "system",
    "content" : "%s"
	}, {
	"role" : "user",
	"content" : "user message"
	} ],
	"name" : "%s",
	"version" : 1,
	"type" : "chat",
	"isActive" : null,
	"config" : { },
	"tags" : [ "test-tag" ],
	"labels" : [ "abc-label" ],
	"commitMessage" : null,
	"resolutionGraph" : null
}`, promptContent, promptName)
	return client.NewWithHttpClient(
		httpmock.NewMockClient(http.StatusOK, apiResponse),
		"https://langfuse3.data.divar.cloud", "pk", "sk")
}

func (s *LangfuseClientTest) TestShouldSetBasicAuth() {
	// Given
	var sentAuthHeader string
	c := client.NewWithHttpClient(
		&http.Client{
			Transport: httpmock.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				sentAuthHeader = req.Header.Get("Authorization")
				return nil, errors.New("http mock called")
			}),
		},
		"https://test.com",
		"test-pk",
		"test-sk",
	)

	// When
	_, err := c.GetPromptTemplate(s.ctx, "test-prompt")
	s.Require().ErrorContains(err, "http mock called")

	// Then
	s.Require().Equal(
		"Basic "+base64.StdEncoding.EncodeToString([]byte("test-pk:test-sk")),
		sentAuthHeader,
	)
}

func TestLangfuseClient(t *testing.T) {
	suite.Run(t, new(LangfuseClientTest))
}
