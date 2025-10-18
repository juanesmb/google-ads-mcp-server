package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"linkedin-mcp/internal/infrastructure/api"
)

type Config struct {
	Timeout        time.Duration
	MaxRetries     int
	RetryDelay     time.Duration
	MaxRetryDelay  time.Duration
	UserAgent      string
	DefaultHeaders map[string]string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
		MaxRetryDelay:  30 * time.Second,
		UserAgent:      "linkedin-mcp-client/1.0",
		DefaultHeaders: make(map[string]string),
	}
}

type httpClient struct {
	client *http.Client
	config *Config
}

func NewClient(config *Config) api.Client {
	if config == nil {
		config = DefaultConfig()
	}

	client := &http.Client{
		Timeout: config.Timeout,
	}

	return &httpClient{
		client: client,
		config: config,
	}
}

func (c *httpClient) Get(ctx context.Context, url string, headers map[string]string) (*api.Response, error) {
	return c.doRequest(ctx, http.MethodGet, url, nil, headers)
}

func (c *httpClient) Post(ctx context.Context, url string, body interface{}, headers map[string]string) (*api.Response, error) {
	return c.doRequest(ctx, http.MethodPost, url, body, headers)
}

func (c *httpClient) Put(ctx context.Context, url string, body interface{}, headers map[string]string) (*api.Response, error) {
	return c.doRequest(ctx, http.MethodPut, url, body, headers)
}

func (c *httpClient) Delete(ctx context.Context, url string, headers map[string]string) (*api.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, url, nil, headers)
}

func (c *httpClient) Patch(ctx context.Context, url string, body interface{}, headers map[string]string) (*api.Response, error) {
	return c.doRequest(ctx, http.MethodPatch, url, body, headers)
}

func (c *httpClient) doRequest(ctx context.Context, method, url string, body interface{}, headers map[string]string) (*api.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			bodyBytes, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		c.setHeaders(req, headers)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			if attempt < c.config.MaxRetries {
				c.waitBeforeRetry(attempt)

				continue
			}

			return nil, lastErr
		}

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			if attempt < c.config.MaxRetries {
				c.waitBeforeRetry(attempt)

				continue
			}
			return nil, lastErr
		}
		resp.Body.Close()

		if c.shouldRetry(resp.StatusCode) && attempt < c.config.MaxRetries {
			lastErr = fmt.Errorf("received status %d, retrying", resp.StatusCode)
			c.waitBeforeRetry(attempt)

			continue
		}

		return &api.Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       responseBody,
		}, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *httpClient) setHeaders(req *http.Request, headers map[string]string) {
	for key, value := range c.config.DefaultHeaders {
		req.Header.Set(key, value)
	}

	req.Header.Set("User-Agent", c.config.UserAgent)

	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}
}

func (c *httpClient) shouldRetry(statusCode int) bool {
	return statusCode >= 500 || statusCode == 429 || statusCode == 408
}

func (c *httpClient) waitBeforeRetry(attempt int) {
	delay := time.Duration(float64(c.config.RetryDelay) * math.Pow(2, float64(attempt)))
	if delay > c.config.MaxRetryDelay {
		delay = c.config.MaxRetryDelay
	}
	time.Sleep(delay)
}
