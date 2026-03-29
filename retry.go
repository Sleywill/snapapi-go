package snapapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// doRaw executes a request and returns the raw response body.
// Retries on transient errors (5xx, 429, network failures) with exponential
// back-off. When the server sends a Retry-After header that value is used as
// the wait duration instead of the computed back-off.
func (c *Client) doRaw(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var (
		lastErr error
		delay   = c.retryDelay
	)
	attempts := c.retries + 1
	for i := 0; i < attempts; i++ {
		data, err := c.roundTrip(ctx, method, path, body)
		if err == nil {
			return data, nil
		}

		// Determine whether this attempt is retryable.
		var apiErr *APIError
		if !isRetryable(err, &apiErr) || i >= attempts-1 {
			return nil, err
		}

		// Compute the sleep duration: honour Retry-After if present, otherwise
		// use exponential back-off with jitter to avoid thundering herd.
		waitDur := delay
		if apiErr != nil && apiErr.RetryAfter > 0 {
			waitDur = time.Duration(apiErr.RetryAfter) * time.Second
		} else {
			// Add up to 25% jitter to avoid synchronized retries.
			jitter := time.Duration(rand.Int63n(int64(delay / 4)))
			waitDur = delay + jitter
		}
		delay *= 2

		lastErr = err
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(waitDur):
		}
	}
	return nil, lastErr
}

// doJSON is like doRaw but JSON-unmarshals the response into dst.
func (c *Client) doJSON(ctx context.Context, method, path string, body interface{}, dst interface{}) error {
	data, err := c.doRaw(ctx, method, path, body)
	if err != nil {
		return err
	}
	if dst == nil {
		return nil
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("snapapi: decode response: %w", err)
	}
	return nil
}

// roundTrip performs a single HTTP request/response cycle.
func (c *Client) roundTrip(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("snapapi: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("snapapi: build request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &APIError{
			Code:    ErrConnectionError,
			Message: err.Error(),
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("snapapi: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(respBody, resp.StatusCode, resp.Header)
	}
	return respBody, nil
}
