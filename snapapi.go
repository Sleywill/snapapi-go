// Package snapapi provides an idiomatic Go client for the SnapAPI.pics service.
//
// SnapAPI lets you capture screenshots, generate PDFs, scrape web pages, and
// extract structured content from URLs — all via a simple HTTP API.
//
// # Quick start
//
//	client := snapapi.New("sk_your_key",
//	    snapapi.WithTimeout(30*time.Second),
//	    snapapi.WithRetries(3),
//	)
//
//	img, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
//	    URL:      "https://example.com",
//	    Format:   "png",
//	    FullPage: true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("screenshot.png", img, 0644)
//
// # Error handling
//
// All methods return a typed *APIError on failure. Use errors.As to inspect it:
//
//	var apiErr *snapapi.APIError
//	if errors.As(err, &apiErr) {
//	    fmt.Println(apiErr.Code, apiErr.StatusCode)
//	}
package snapapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://snapapi.pics"
	defaultTimeout = 30 * time.Second
	defaultRetries = 3
	userAgent      = "snapapi-go/3.0.0"
)

// Client is the SnapAPI client. Create one with New().
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	retries    int
	retryDelay time.Duration
}

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

// New creates a new SnapAPI client with the given API key.
//
//	client := snapapi.New("sk_...",
//	    snapapi.WithTimeout(45*time.Second),
//	    snapapi.WithRetries(3),
//	)
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		retries:    defaultRetries,
		retryDelay: 500 * time.Millisecond,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// NewClient is an alias for New, kept for compatibility.
func NewClient(apiKey string, opts ...Option) *Client {
	return New(apiKey, opts...)
}

// Screenshot captures a screenshot of a URL.
// Returns raw image bytes (PNG or JPEG depending on Params.Format).
//
//	img, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
//	    URL:      "https://example.com",
//	    Format:   "png",
//	    FullPage: true,
//	})
func (c *Client) Screenshot(ctx context.Context, p ScreenshotParams) ([]byte, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	return c.doRaw(ctx, http.MethodPost, "/v1/screenshot", p)
}

// ScreenshotParams holds parameters for Screenshot.
type ScreenshotParams struct {
	// URL of the page to capture. Required.
	URL string `json:"url"`
	// Format is the output image format: "png" (default) or "jpeg".
	Format string `json:"format,omitempty"`
	// Width of the viewport in pixels. Default: 1280.
	Width int `json:"width,omitempty"`
	// Height of the viewport in pixels. Default: 720.
	Height int `json:"height,omitempty"`
	// FullPage captures the entire scrollable page. Default: false.
	FullPage bool `json:"full_page,omitempty"`
	// Wait is additional time in milliseconds to wait after page load.
	Wait int `json:"wait,omitempty"`
	// Delay is an alias for Wait (some API versions use this field name).
	Delay int `json:"delay,omitempty"`
	// Quality sets JPEG compression quality (1–100). Only for format="jpeg".
	Quality int `json:"quality,omitempty"`
	// Selector captures only the element matching this CSS selector.
	Selector string `json:"selector,omitempty"`
}

// Scrape fetches text or structured content from a URL.
//
//	data, err := client.Scrape(ctx, snapapi.ScrapeParams{URL: "https://example.com"})
func (c *Client) Scrape(ctx context.Context, p ScrapeParams) (*ScrapeResult, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	var result ScrapeResult
	if err := c.doJSON(ctx, http.MethodPost, "/v1/scrape", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ScrapeParams holds parameters for Scrape.
type ScrapeParams struct {
	// URL of the page to scrape. Required.
	URL string `json:"url"`
	// Selector is a CSS selector to scope scraping to a subtree.
	Selector string `json:"selector,omitempty"`
	// Wait is milliseconds to wait after page load before scraping.
	Wait int `json:"wait,omitempty"`
}

// ScrapeResult is the structured response from the scrape endpoint.
type ScrapeResult struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	// HTML is the raw HTML of the scraped page or element.
	HTML string `json:"html,omitempty"`
	// Text is the plain-text content.
	Text string `json:"text,omitempty"`
}

// Extract extracts readable content from a URL, suitable for LLM consumption.
//
//	content, err := client.Extract(ctx, snapapi.ExtractParams{
//	    URL:    "https://example.com",
//	    Format: "markdown",
//	})
func (c *Client) Extract(ctx context.Context, p ExtractParams) (*ExtractResult, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	var result ExtractResult
	if err := c.doJSON(ctx, http.MethodPost, "/v1/extract", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ExtractParams holds parameters for Extract.
type ExtractParams struct {
	// URL of the page to extract content from. Required.
	URL string `json:"url"`
	// Format is the output format: "markdown", "text", or "json".
	Format string `json:"format,omitempty"`
	// Wait is milliseconds to wait after page load.
	Wait int `json:"wait,omitempty"`
}

// ExtractResult is the structured response from the extract endpoint.
type ExtractResult struct {
	Success      bool   `json:"success"`
	URL          string `json:"url"`
	Format       string `json:"format"`
	Content      string `json:"content"`
	ResponseTime int    `json:"responseTime"`
}

// PDF generates a PDF of a URL.
//
//	pdfBytes, err := client.PDF(ctx, snapapi.PDFParams{URL: "https://example.com"})
func (c *Client) PDF(ctx context.Context, p PDFParams) ([]byte, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	return c.doRaw(ctx, http.MethodPost, "/v1/pdf", p)
}

// PDFParams holds parameters for PDF generation.
type PDFParams struct {
	// URL of the page to convert. Required.
	URL string `json:"url"`
	// Format is the paper size: "a4" (default) or "letter".
	Format string `json:"format,omitempty"`
	// Margin sets page margins (e.g. "10mm", "1cm").
	Margin string `json:"margin,omitempty"`
}

// Video records a short video of a URL.
//
//	videoBytes, err := client.Video(ctx, snapapi.VideoParams{URL: "https://example.com"})
func (c *Client) Video(ctx context.Context, p VideoParams) ([]byte, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	return c.doRaw(ctx, http.MethodPost, "/v1/video", p)
}

// VideoParams holds parameters for video recording.
type VideoParams struct {
	// URL of the page to record. Required.
	URL string `json:"url"`
	// Duration in seconds. Default: 5.
	Duration int `json:"duration,omitempty"`
	// Format: "webm" (default), "mp4", or "gif".
	Format string `json:"format,omitempty"`
	// Width of the viewport in pixels.
	Width int `json:"width,omitempty"`
	// Height of the viewport in pixels.
	Height int `json:"height,omitempty"`
}

// Quota returns the caller's current API quota usage.
//
//	q, err := client.Quota(ctx)
//	fmt.Printf("Used: %d / %d\n", q.Used, q.Total)
func (c *Client) Quota(ctx context.Context) (*QuotaResult, error) {
	var result QuotaResult
	if err := c.doJSON(ctx, http.MethodGet, "/v1/quota", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// QuotaResult is the response from GET /v1/quota.
type QuotaResult struct {
	Used      int    `json:"used"`
	Total     int    `json:"total"`
	Remaining int    `json:"remaining"`
	ResetAt   string `json:"resetAt,omitempty"`
}

// ─── Internal transport ───────────────────────────────────────────────────────

// doRaw executes a request and returns the raw response body.
// Retries on transient errors with exponential back-off.
func (c *Client) doRaw(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var (
		lastErr error
		delay   = c.retryDelay
	)
	attempts := c.retries + 1
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2
		}

		data, err := c.roundTrip(ctx, method, path, body)
		if err == nil {
			return data, nil
		}

		var apiErr *APIError
		if isRetryable(err, &apiErr) && i < attempts-1 {
			// Respect Retry-After if present (stored in APIError.RetryAfter)
			if apiErr != nil && apiErr.RetryAfter > 0 {
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(time.Duration(apiErr.RetryAfter) * time.Second):
				}
			}
			lastErr = err
			continue
		}
		return nil, err
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
