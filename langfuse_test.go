package golangfuse_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/divar-ir/go-http-mock/pkg/httpmock"
	"github.com/divar-ir/golangfuse"
	"github.com/stretchr/testify/suite"
)

type ClientTest struct {
	suite.Suite
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (s *ClientTest) SetupTest() {
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
}

func (s *ClientTest) TearDownTest() {
	s.ctxCancel()
}

func (s *ClientTest) TestGetPromptTemplateShouldParseApiResponse() {
	// Given
	const promptName = "test-prompt"
	c := s.getLangfuseClientForTest(promptName, "This is system prompt")

	// When
	promptTemplate, err := c.GetPromptTemplate(s.ctx, promptName)
	s.Require().NoError(err)
	s.Require().Equal("This is system prompt", promptTemplate)
}

func (s *ClientTest) TestGetPromptTemplateShouldReturnVariablesWithGolangFormat() {
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

func (s *ClientTest) getLangfuseClientForTest(promptName, promptContent string) golangfuse.Langfuse {
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
	return golangfuse.NewWithHttpClient(
		httpmock.NewMockClient(http.StatusOK, apiResponse),
		"https://langfuse3.data.divar.cloud", "pk", "sk")
}

func (s *ClientTest) TestShouldSetBasicAuth() {
	// Given
	var sentAuthHeader string
	c := s.getClientWithMockedHttpTransport(func(req *http.Request) (*http.Response, error) {
		sentAuthHeader = req.Header.Get("Authorization")
		return nil, errors.New("http mock called")
	})

	// When
	_, err := c.GetPromptTemplate(s.ctx, "test-prompt")
	s.Require().ErrorContains(err, "http mock called")

	// Then
	s.Require().Equal(
		"Basic "+base64.StdEncoding.EncodeToString([]byte("test-pk:test-sk")),
		sentAuthHeader,
	)
}

func (s *ClientTest) TestStartSendingShouldReturnErrorIfAlreadyStarted() {
	// Given
	c := s.getClient()
	err := c.StartSendingEvents(s.ctx, 1*time.Microsecond)
	s.Require().NoError(err)

	// When
	err = c.StartSendingEvents(s.ctx, 1*time.Microsecond)

	// Then
	s.Require().ErrorContains(err, "already started")
}

func (s *ClientTest) TestShouldSendEvent() {
	// Given
	var sentRequestBody []byte
	wg := &sync.WaitGroup{}
	wg.Add(1)
	c := s.getClientWithMockedHttpTransport(func(req *http.Request) (*http.Response, error) {
		defer wg.Done()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		sentRequestBody = body
		return &http.Response{}, nil
	})
	err := c.StartSendingEvents(s.ctx, 1*time.Microsecond)
	s.Require().NoError(err)

	// When
	c.Trace("input", "output")
	wg.Wait()

	// Then
	type requestBody struct {
		Batch []struct {
			Body struct {
				Input  string `json:"input"`
				Output string `json:"output"`
			}
		} `json:"batch"`
	}
	bodyObj := requestBody{}
	err = json.Unmarshal(sentRequestBody, &bodyObj)
	s.Require().NoError(err)
	s.Require().Len(bodyObj.Batch, 1)
	s.Require().Equal("input", bodyObj.Batch[0].Body.Input)
	s.Require().Equal("output", bodyObj.Batch[0].Body.Output)
}

func (s *ClientTest) TestShouldNotSendAnythingWhenNoEventIsReported() {
	// Given
	httpCallHappened := false
	c := s.getClientWithMockedHttpTransport(func(req *http.Request) (*http.Response, error) {
		httpCallHappened = true
		return &http.Response{}, nil
	})

	// When
	err := c.StartSendingEvents(s.ctx, 1*time.Microsecond)
	s.Require().NoError(err)
	time.Sleep(1 * time.Millisecond)

	// Then
	s.Require().False(httpCallHappened, "http call happened, unexpectedly")
}

func (s *ClientTest) TestShouldSendEventsInBatch() {
	// Given
	var sentRequestBody []byte
	wg := &sync.WaitGroup{}
	wg.Add(1)
	c := s.getClientWithMockedHttpTransport(func(req *http.Request) (*http.Response, error) {
		defer wg.Done()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		sentRequestBody = body
		return &http.Response{}, nil
	})
	err := c.StartSendingEvents(s.ctx, 1*time.Microsecond)
	s.Require().NoError(err)

	// When
	c.Trace("input", "output")
	c.Trace("input", "output")
	wg.Wait()

	// Then
	type requestBody struct {
		Batch []struct{} `json:"batch"`
	}
	bodyObj := requestBody{}
	err = json.Unmarshal(sentRequestBody, &bodyObj)
	s.Require().NoError(err)
	s.Require().Len(bodyObj.Batch, 2)
}

func (s *ClientTest) getClient() golangfuse.Langfuse {
	return s.getClientWithMockedHttpTransport(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusMultiStatus,
			Body:       io.NopCloser(strings.NewReader("{}")),
		}, nil
	})
}

func (s *ClientTest) getClientWithMockedHttpTransport(transport httpmock.RoundTripFunc) golangfuse.Langfuse {
	return golangfuse.NewWithHttpClient(
		&http.Client{Transport: transport},
		"https://test.com",
		"test-pk",
		"test-sk",
	)
}

func TestLangfuseClient(t *testing.T) {
	suite.Run(t, new(ClientTest))
}
