package snapapi

import (
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.snapapi.pics"
	defaultTimeout = 30 * time.Second
	defaultRetries = 3
	userAgent      = "snapapi-go/3.2.0"
)

// Option configures a Client. Pass options to New().
type Option func(*Client)

// WithBaseURL overrides the API base URL (useful for testing or proxies).
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout sets the HTTP request timeout. Default is 30s.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithRetries sets the number of automatic retries on transient errors.
// Default is 3. Set to 0 to disable retries.
func WithRetries(n int) Option {
	return func(c *Client) {
		c.retries = n
	}
}

// WithRetryDelay sets the base delay for exponential backoff between retries.
// Default is 500ms.
func WithRetryDelay(d time.Duration) Option {
	return func(c *Client) {
		c.retryDelay = d
	}
}

// WithHTTPClient replaces the underlying *http.Client entirely.
// This allows injecting a custom transport (e.g. for mocking in tests).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}
