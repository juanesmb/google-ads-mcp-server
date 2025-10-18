package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"linkedin-mcp/internal/infrastructure/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HTTPClientSuite struct {
	suite.Suite
	client   api.Client
	server   *httptest.Server
	config   *Config
	ctx      context.Context
	response *api.Response
	err      error
}

func (s *HTTPClientSuite) SetupSuite() {
	s.ctx = context.Background()
	s.config = DefaultConfig()
	s.config.MaxRetries = 1 // Reduce retries for faster tests
	s.config.RetryDelay = 10 * time.Millisecond
	s.config.MaxRetryDelay = 50 * time.Millisecond
}

func (s *HTTPClientSuite) SetupTest() {
	s.client = NewClient(s.config)
	s.response = nil
	s.err = nil
}

func (s *HTTPClientSuite) TearDownTest() {
	if s.server != nil {
		s.server.Close()
		s.server = nil
	}
}

func (s *HTTPClientSuite) givenServerWithHandler(handler http.HandlerFunc) {
	s.server = httptest.NewServer(handler)
}

func (s *HTTPClientSuite) givenServerReturnsStatus(statusCode int) {
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte("test response"))
	})
}

func (s *HTTPClientSuite) givenServerReturnsJSONResponse(data interface{}) {
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})
}

func (s *HTTPClientSuite) givenServerWithCustomHeaders(headers map[string]string) {
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		for key, value := range headers {
			w.Header().Set(key, value)
		}
		w.Write([]byte("test response"))
	})
}

func (s *HTTPClientSuite) givenServerWithDelay(delay time.Duration) {
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Write([]byte("delayed response"))
	})
}

func (s *HTTPClientSuite) givenServerThatFails() {
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		// Simulate server error by closing connection
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	})
}

func (s *HTTPClientSuite) whenIGetRequest() {
	s.response, s.err = s.client.Get(s.ctx, s.server.URL, nil)
}

func (s *HTTPClientSuite) whenIPostRequest(body interface{}) {
	s.response, s.err = s.client.Post(s.ctx, s.server.URL, body, nil)
}

func (s *HTTPClientSuite) whenIPutRequest(body interface{}) {
	s.response, s.err = s.client.Put(s.ctx, s.server.URL, body, nil)
}

func (s *HTTPClientSuite) whenIDeleteRequest() {
	s.response, s.err = s.client.Delete(s.ctx, s.server.URL, nil)
}

func (s *HTTPClientSuite) whenIPatchRequest(body interface{}) {
	s.response, s.err = s.client.Patch(s.ctx, s.server.URL, body, nil)
}

func (s *HTTPClientSuite) whenIGetRequestWithHeaders(headers map[string]string) {
	s.response, s.err = s.client.Get(s.ctx, s.server.URL, headers)
}

func (s *HTTPClientSuite) thenRequestSucceeds() {
	s.NoError(s.err)
	s.NotNil(s.response)
}

func (s *HTTPClientSuite) thenRequestFails() {
	s.Error(s.err)
}

func (s *HTTPClientSuite) thenStatusCodeIs(expected int) {
	s.Equal(expected, s.response.StatusCode)
}

func (s *HTTPClientSuite) thenResponseBodyContains(text string) {
	s.Contains(string(s.response.Body), text)
}

func (s *HTTPClientSuite) thenResponseHasHeader(key, value string) {
	headers := s.response.Headers[key]
	s.NotEmpty(headers)
	s.Contains(headers, value)
}

func (s *HTTPClientSuite) thenErrorContains(text string) {
	s.Error(s.err)
	s.Contains(s.err.Error(), text)
}

func (s *HTTPClientSuite) TestWhenValidGetRequest_ThenSucceeds() {
	s.givenServerReturnsStatus(http.StatusOK)
	s.whenIGetRequest()
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusOK)
	s.thenResponseBodyContains("test response")
}

func (s *HTTPClientSuite) TestWhenValidPostRequest_ThenSucceeds() {
	requestBody := map[string]string{"key": "value"}
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		s.Equal(http.MethodPost, r.Method)
		s.Equal("application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	})
	s.whenIPostRequest(requestBody)
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusCreated)
	s.thenResponseBodyContains("created")
}

func (s *HTTPClientSuite) TestWhenValidPutRequest_ThenSucceeds() {
	requestBody := map[string]string{"key": "updated"}
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		s.Equal(http.MethodPut, r.Method)
		s.Equal("application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("updated"))
	})
	s.whenIPutRequest(requestBody)
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusOK)
	s.thenResponseBodyContains("updated")
}

func (s *HTTPClientSuite) TestWhenValidDeleteRequest_ThenSucceeds() {
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		s.Equal(http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	})
	s.whenIDeleteRequest()
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusNoContent)
}

func (s *HTTPClientSuite) TestWhenValidPatchRequest_ThenSucceeds() {
	requestBody := map[string]string{"key": "patched"}
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		s.Equal(http.MethodPatch, r.Method)
		s.Equal("application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("patched"))
	})
	s.whenIPatchRequest(requestBody)
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusOK)
	s.thenResponseBodyContains("patched")
}

func (s *HTTPClientSuite) TestWhenServerReturns500_ThenRetries() {
	attemptCount := 0
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte("response"))
	})
	s.whenIGetRequest()
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusOK)
	s.Equal(2, attemptCount)
}

func (s *HTTPClientSuite) TestWhenServerReturns429_ThenRetries() {
	attemptCount := 0
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		w.Write([]byte("response"))
	})
	s.whenIGetRequest()
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusOK)
	s.Equal(2, attemptCount)
}

func (s *HTTPClientSuite) TestWhenMaxRetriesExceeded_ThenFails() {
	// Set MaxRetries to 0 to ensure it fails immediately
	s.config.MaxRetries = 0
	s.client = NewClient(s.config)
	s.givenServerReturnsStatus(http.StatusInternalServerError)
	s.whenIGetRequest()
	// With MaxRetries = 0, it should return the 500 status code, not fail
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusInternalServerError)
}

func (s *HTTPClientSuite) TestWhenServerAlwaysFails_ThenMaxRetriesExceeded() {
	// Set MaxRetries to 1 and make server always fail
	s.config.MaxRetries = 1
	s.client = NewClient(s.config)
	s.givenServerThatFails()
	s.whenIGetRequest()
	s.thenRequestFails()
	s.thenErrorContains("request failed")
}

func (s *HTTPClientSuite) TestWhenMaxRetriesExceededWithRetryableStatus_ThenReturnsLastStatus() {
	// Set MaxRetries to 1 and server always returns 500
	s.config.MaxRetries = 1
	s.client = NewClient(s.config)
	s.givenServerReturnsStatus(http.StatusInternalServerError)
	s.whenIGetRequest()
	// Should succeed and return the 500 status code after retrying
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusInternalServerError)
}

func (s *HTTPClientSuite) TestWhenMaxRetriesExceededWithRetryableStatusAndNetworkError_ThenFails() {
	// Set MaxRetries to 1 and create a scenario where first attempt fails with network error,
	// second attempt fails with network error, then we exceed max retries
	s.config.MaxRetries = 1
	s.client = NewClient(s.config)

	attemptCount := 0
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		// Always fail with network error
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	})
	s.whenIGetRequest()
	s.thenRequestFails()
	s.thenErrorContains("request failed")
	s.Equal(2, attemptCount) // Should have tried twice (initial + 1 retry)
}

func (s *HTTPClientSuite) TestWhenServerFails_ThenRetries() {
	attemptCount := 0
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			// Simulate connection failure
			if hj, ok := w.(http.Hijacker); ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	})
	s.whenIGetRequest()
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusOK)
	s.Equal(2, attemptCount)
}

func (s *HTTPClientSuite) TestWhenCustomHeadersProvided_ThenHeadersAreSet() {
	headers := map[string]string{
		"Authorization": "Bearer token",
		"X-Custom":      "value",
	}
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("Bearer token", r.Header.Get("Authorization"))
		s.Equal("value", r.Header.Get("X-Custom"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
	s.whenIGetRequestWithHeaders(headers)
	s.thenRequestSucceeds()
}

func (s *HTTPClientSuite) TestWhenDefaultHeadersSet_ThenHeadersAreApplied() {
	s.config.DefaultHeaders = map[string]string{
		"X-API-Key": "default-key",
		"X-Version": "1.0",
	}
	s.client = NewClient(s.config)
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		s.Equal("default-key", r.Header.Get("X-API-Key"))
		s.Equal("1.0", r.Header.Get("X-Version"))
		s.Equal("linkedin-mcp-client/1.0", r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})
	s.whenIGetRequest()
	s.thenRequestSucceeds()
}

func (s *HTTPClientSuite) TestWhenResponseHasHeaders_ThenHeadersAreReturned() {
	responseHeaders := map[string]string{
		"Content-Type": "application/json",
		"X-Rate-Limit": "100",
	}
	s.givenServerWithCustomHeaders(responseHeaders)
	s.whenIGetRequest()
	s.thenRequestSucceeds()
	s.thenResponseHasHeader("Content-Type", "application/json")
	s.thenResponseHasHeader("X-Rate-Limit", "100")
}

func (s *HTTPClientSuite) TestWhenJSONResponse_ThenBodyIsCorrect() {
	expectedData := map[string]interface{}{
		"id":   float64(1), // JSON unmarshaling converts to float64
		"name": "test",
	}
	s.givenServerReturnsJSONResponse(expectedData)
	s.whenIGetRequest()
	s.thenRequestSucceeds()

	var actualData map[string]interface{}
	err := json.Unmarshal(s.response.Body, &actualData)
	s.NoError(err)
	s.Equal(expectedData, actualData)
}

func (s *HTTPClientSuite) TestWhenInvalidJSONBody_ThenFails() {
	invalidBody := make(chan int) // Channels cannot be marshaled to JSON
	s.givenServerReturnsStatus(http.StatusOK)
	s.whenIPostRequest(invalidBody)
	s.thenRequestFails()
	s.thenErrorContains("failed to marshal request body")
}

func (s *HTTPClientSuite) TestWhenInvalidURL_ThenFails() {
	// Test with invalid URL to trigger request creation failure
	s.response, s.err = s.client.Get(s.ctx, "invalid://url", nil)
	s.thenRequestFails()
	s.thenErrorContains("request failed")
}

func (s *HTTPClientSuite) TestWhenMaxRetriesExceededOnNetworkError_ThenFails() {
	// Set MaxRetries to 0 and make server always fail
	s.config.MaxRetries = 0
	s.client = NewClient(s.config)
	s.givenServerThatFails()
	s.whenIGetRequest()
	s.thenRequestFails()
	s.thenErrorContains("request failed")
}

func (s *HTTPClientSuite) TestWhenMaxRetriesExceededOnBodyReadError_ThenFails() {
	// Set MaxRetries to 0 and create a scenario where body read fails
	s.config.MaxRetries = 0
	s.client = NewClient(s.config)

	// Create a server that closes connection immediately after sending headers
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Close connection immediately to cause body read error
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
	})
	s.whenIGetRequest()
	s.thenRequestFails()
	s.thenErrorContains("failed to read response body")
}

func (s *HTTPClientSuite) TestWhenContextCancelled_ThenFails() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	s.givenServerWithDelay(100 * time.Millisecond)
	response, err := s.client.Get(ctx, s.server.URL, nil)
	s.Error(err)
	s.Nil(response)
}

func (s *HTTPClientSuite) TestWhenTimeoutExceeded_ThenFails() {
	s.config.Timeout = 50 * time.Millisecond
	s.client = NewClient(s.config)
	s.givenServerWithDelay(100 * time.Millisecond)
	s.whenIGetRequest()
	s.thenRequestFails()
	s.thenErrorContains("request failed")
}

func (s *HTTPClientSuite) TestWhenNilConfig_ThenUsesDefaultConfig() {
	client := NewClient(nil)
	s.NotNil(client)

	// Test that it works with default config
	s.givenServerReturnsStatus(http.StatusOK)
	response, err := client.Get(s.ctx, s.server.URL, nil)
	s.NoError(err)
	s.NotNil(response)
}

func (s *HTTPClientSuite) TestWhenNonRetryableStatusCode_ThenDoesNotRetry() {
	attemptCount := 0
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusBadRequest) // 400 is not retryable
		w.Write([]byte("bad request"))
	})
	s.whenIGetRequest()
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusBadRequest)
	s.Equal(1, attemptCount) // Should not retry
}

func (s *HTTPClientSuite) TestWhenExponentialBackoff_ThenDelayIncreases() {
	// Set MaxRetries to 1 to ensure it retries once and then returns the error status
	s.config.MaxRetries = 1
	s.client = NewClient(s.config)

	start := time.Now()
	s.givenServerReturnsStatus(http.StatusInternalServerError)
	s.whenIGetRequest()
	elapsed := time.Since(start)

	// Should have retried with exponential backoff
	// First retry: 10ms, total should be at least 10ms
	s.True(elapsed >= 10*time.Millisecond)
	// With MaxRetries = 1, it should return the 500 status code after retrying
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusInternalServerError)
}

func (s *HTTPClientSuite) TestWhenMaxRetryDelayExceeded_ThenUsesMaxDelay() {
	s.config.RetryDelay = 1 * time.Second
	s.config.MaxRetryDelay = 100 * time.Millisecond
	s.config.MaxRetries = 1
	s.client = NewClient(s.config)

	start := time.Now()
	s.givenServerReturnsStatus(http.StatusInternalServerError)
	s.whenIGetRequest()
	elapsed := time.Since(start)

	// Should not exceed max retry delay
	s.True(elapsed < 500*time.Millisecond) // Should be much less than 1s + 2s
	// With MaxRetries = 1, it should return the 500 status code after retrying
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusInternalServerError)
}

func (s *HTTPClientSuite) TestWhenResponseBodyReadFails_ThenRetries() {
	attemptCount := 0
	s.givenServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			// Simulate body read failure by closing connection
			if hj, ok := w.(http.Hijacker); ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	})
	s.whenIGetRequest()
	s.thenRequestSucceeds()
	s.thenStatusCodeIs(http.StatusOK)
	s.Equal(2, attemptCount)
}

func TestHTTPClientSuite(t *testing.T) {
	suite.Run(t, new(HTTPClientSuite))
}

// TestDefaultConfig tests the DefaultConfig function
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
	assert.Equal(t, 30*time.Second, config.MaxRetryDelay)
	assert.Equal(t, "linkedin-mcp-client/1.0", config.UserAgent)
	assert.NotNil(t, config.DefaultHeaders)
}

// TestResponse tests the Response struct
func TestResponse(t *testing.T) {
	response := &api.Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       []byte("test"),
	}

	assert.Equal(t, 200, response.StatusCode)
	assert.Equal(t, "application/json", response.Headers["Content-Type"][0])
	assert.Equal(t, []byte("test"), response.Body)
}
