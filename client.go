// Package snapapi provides a Go client for the SnapAPI screenshot and web extraction service.
//
// Example usage:
//
//	client := snapapi.NewClient("your-api-key")
//	data, err := client.Screenshot(ctx, snapapi.ScreenshotOptions{
//	    URL: "https://example.com",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("screenshot.png", data, 0644)
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
	defaultBaseURL = "https://api.snapapi.pics"
	defaultTimeout = 90 * time.Second
	userAgent      = "snapapi-go/2.0.0"
)

// Client is the main SnapAPI client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	// Sub-clients for resource groups
	Storage   *StorageClient
	Scheduled *ScheduledClient
	Webhooks  *WebhooksClient
	Keys      *KeysClient
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout sets a custom timeout for HTTP requests.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient replaces the underlying HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new SnapAPI client with the given API key.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Storage = &StorageClient{c}
	c.Scheduled = &ScheduledClient{c}
	c.Webhooks = &WebhooksClient{c}
	c.Keys = &KeysClient{c}
	return c
}

// ─── Core Endpoints ───────────────────────────────────────────────────────────

// Screenshot captures a screenshot of a URL or HTML/Markdown content.
// Returns raw binary image bytes (PNG/JPEG/WEBP/AVIF) or PDF bytes.
// When Storage options are set, returns a StorageResult JSON instead.
func (c *Client) Screenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error) {
	if opts.URL == "" && opts.HTML == "" && opts.Markdown == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL, HTML, or Markdown is required", StatusCode: 400}
	}
	return c.doRequest(ctx, "POST", "/v1/screenshot", opts)
}

// ScreenshotToStorage captures a screenshot and stores it, returning metadata.
func (c *Client) ScreenshotToStorage(ctx context.Context, opts ScreenshotOptions) (*StorageUploadResult, error) {
	data, err := c.Screenshot(ctx, opts)
	if err != nil {
		return nil, err
	}
	var result StorageUploadResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse storage result: %w", err)
	}
	return &result, nil
}

// Scrape scrapes content from a URL.
func (c *Client) Scrape(ctx context.Context, opts ScrapeOptions) (*ScrapeResult, error) {
	if opts.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	data, err := c.doRequest(ctx, "POST", "/v1/scrape", opts)
	if err != nil {
		return nil, err
	}
	var result ScrapeResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse scrape result: %w", err)
	}
	return &result, nil
}

// Extract extracts content from a webpage.
func (c *Client) Extract(ctx context.Context, opts ExtractOptions) (*ExtractResult, error) {
	if opts.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	data, err := c.doRequest(ctx, "POST", "/v1/extract", opts)
	if err != nil {
		return nil, err
	}
	var result ExtractResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse extract result: %w", err)
	}
	return &result, nil
}

// Analyze performs AI-powered analysis of a webpage.
func (c *Client) Analyze(ctx context.Context, opts AnalyzeOptions) (*AnalyzeResult, error) {
	if opts.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	data, err := c.doRequest(ctx, "POST", "/v1/analyze", opts)
	if err != nil {
		return nil, err
	}
	var result AnalyzeResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse analyze result: %w", err)
	}
	return &result, nil
}

// ─── Convenience helpers ──────────────────────────────────────────────────────

// PDF generates a PDF (sets Format = "pdf").
func (c *Client) PDF(ctx context.Context, opts ScreenshotOptions) ([]byte, error) {
	opts.Format = "pdf"
	return c.Screenshot(ctx, opts)
}

// ExtractMarkdown is a shorthand for Extract with Type "markdown".
func (c *Client) ExtractMarkdown(ctx context.Context, url string) (*ExtractResult, error) {
	return c.Extract(ctx, ExtractOptions{URL: url, Type: "markdown"})
}

// ExtractArticle is a shorthand for Extract with Type "article".
func (c *Client) ExtractArticle(ctx context.Context, url string) (*ExtractResult, error) {
	return c.Extract(ctx, ExtractOptions{URL: url, Type: "article"})
}

// ExtractText is a shorthand for Extract with Type "text".
func (c *Client) ExtractText(ctx context.Context, url string) (*ExtractResult, error) {
	return c.Extract(ctx, ExtractOptions{URL: url, Type: "text"})
}

// ExtractLinks is a shorthand for Extract with Type "links".
func (c *Client) ExtractLinks(ctx context.Context, url string) (*ExtractResult, error) {
	return c.Extract(ctx, ExtractOptions{URL: url, Type: "links"})
}

// ExtractImages is a shorthand for Extract with Type "images".
func (c *Client) ExtractImages(ctx context.Context, url string) (*ExtractResult, error) {
	return c.Extract(ctx, ExtractOptions{URL: url, Type: "images"})
}

// ExtractMetadata is a shorthand for Extract with Type "metadata".
func (c *Client) ExtractMetadata(ctx context.Context, url string) (*ExtractResult, error) {
	return c.Extract(ctx, ExtractOptions{URL: url, Type: "metadata"})
}

// ─── Internal ─────────────────────────────────────────────────────────────────

// doRequest performs an HTTP request and returns raw response bytes.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &APIError{Code: ErrConnectionError, Message: fmt.Sprintf("connection error: %v", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(respBody, resp.StatusCode)
	}
	return respBody, nil
}

// doRequestJSON is like doRequest but unmarshals the response into dst.
func (c *Client) doRequestJSON(ctx context.Context, method, path string, body interface{}, dst interface{}) error {
	data, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	if dst == nil {
		return nil
	}
	return json.Unmarshal(data, dst)
}

// ─── Sub-Clients ──────────────────────────────────────────────────────────────

// StorageClient provides access to /v1/storage/* endpoints.
type StorageClient struct{ c *Client }

// ListFiles returns all stored files.
func (s *StorageClient) ListFiles(ctx context.Context) (*StorageFilesResult, error) {
	var result StorageFilesResult
	return &result, s.c.doRequestJSON(ctx, "GET", "/v1/storage/files", nil, &result)
}

// DeleteFile deletes a stored file by ID.
func (s *StorageClient) DeleteFile(ctx context.Context, id string) error {
	return s.c.doRequestJSON(ctx, "DELETE", "/v1/storage/files/"+id, nil, nil)
}

// GetUsage returns storage usage statistics.
func (s *StorageClient) GetUsage(ctx context.Context) (*StorageUsageResult, error) {
	var result StorageUsageResult
	return &result, s.c.doRequestJSON(ctx, "GET", "/v1/storage/usage", nil, &result)
}

// ConfigureS3 configures an S3-compatible storage backend.
func (s *StorageClient) ConfigureS3(ctx context.Context, opts S3Config) error {
	return s.c.doRequestJSON(ctx, "POST", "/v1/storage/s3", opts, nil)
}

// TestS3 tests the configured S3 storage connection.
func (s *StorageClient) TestS3(ctx context.Context) error {
	return s.c.doRequestJSON(ctx, "POST", "/v1/storage/s3/test", nil, nil)
}

// ScheduledClient provides access to /v1/scheduled/* endpoints.
type ScheduledClient struct{ c *Client }

// Create creates a new scheduled screenshot job.
func (s *ScheduledClient) Create(ctx context.Context, opts ScheduledOptions) (*ScheduledJob, error) {
	var result ScheduledJob
	return &result, s.c.doRequestJSON(ctx, "POST", "/v1/scheduled", opts, &result)
}

// List returns all scheduled jobs.
func (s *ScheduledClient) List(ctx context.Context) (*ScheduledListResult, error) {
	var result ScheduledListResult
	return &result, s.c.doRequestJSON(ctx, "GET", "/v1/scheduled", nil, &result)
}

// Delete removes a scheduled job by ID.
func (s *ScheduledClient) Delete(ctx context.Context, id string) error {
	return s.c.doRequestJSON(ctx, "DELETE", "/v1/scheduled/"+id, nil, nil)
}

// WebhooksClient provides access to /v1/webhooks/* endpoints.
type WebhooksClient struct{ c *Client }

// Create registers a new webhook.
func (w *WebhooksClient) Create(ctx context.Context, opts WebhookOptions) (*Webhook, error) {
	var result Webhook
	return &result, w.c.doRequestJSON(ctx, "POST", "/v1/webhooks", opts, &result)
}

// List returns all registered webhooks.
func (w *WebhooksClient) List(ctx context.Context) (*WebhooksListResult, error) {
	var result WebhooksListResult
	return &result, w.c.doRequestJSON(ctx, "GET", "/v1/webhooks", nil, &result)
}

// Delete removes a webhook by ID.
func (w *WebhooksClient) Delete(ctx context.Context, id string) error {
	return w.c.doRequestJSON(ctx, "DELETE", "/v1/webhooks/"+id, nil, nil)
}

// KeysClient provides access to /v1/keys/* endpoints.
type KeysClient struct{ c *Client }

// List returns all API keys.
func (k *KeysClient) List(ctx context.Context) (*KeysListResult, error) {
	var result KeysListResult
	return &result, k.c.doRequestJSON(ctx, "GET", "/v1/keys", nil, &result)
}

// Create creates a new API key.
func (k *KeysClient) Create(ctx context.Context, name string) (*APIKey, error) {
	var result APIKey
	return &result, k.c.doRequestJSON(ctx, "POST", "/v1/keys", map[string]string{"name": name}, &result)
}

// Delete revokes an API key by ID.
func (k *KeysClient) Delete(ctx context.Context, id string) error {
	return k.c.doRequestJSON(ctx, "DELETE", "/v1/keys/"+id, nil, nil)
}
