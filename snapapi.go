// Package snapapi provides an idiomatic Go client for the SnapAPI.pics service.
//
// SnapAPI lets you capture screenshots, generate PDFs, scrape web pages,
// extract structured content, and analyze pages with LLMs — all via a simple
// HTTP API.
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
// # Namespaces
//
// The client exposes sub-namespaces for grouping related endpoints:
//
//	client.Storage    -- manage stored captures
//	client.Scheduled  -- schedule recurring captures
//	client.Webhooks   -- manage webhook endpoints
//	client.APIKeys    -- manage API keys for your account
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
	"os"
	"time"
)

const (
	defaultBaseURL = "https://api.snapapi.pics"
	defaultTimeout = 30 * time.Second
	defaultRetries = 3
	userAgent      = "snapapi-go/3.1.0"
)

// Client is the SnapAPI client. Create one with New().
//
// The client is safe for concurrent use by multiple goroutines.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	retries    int
	retryDelay time.Duration

	// Sub-namespace accessors. Populated by New().
	Storage   *StorageNamespace
	Scheduled *ScheduledNamespace
	Webhooks  *WebhooksNamespace
	APIKeys   *APIKeysNamespace
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
	// Wire up namespace accessors.
	c.Storage = &StorageNamespace{c: c}
	c.Scheduled = &ScheduledNamespace{c: c}
	c.Webhooks = &WebhooksNamespace{c: c}
	c.APIKeys = &APIKeysNamespace{c: c}
	return c
}

// NewClient is an alias for New, kept for backward compatibility.
func NewClient(apiKey string, opts ...Option) *Client {
	return New(apiKey, opts...)
}

// ---------------------------------------------------------------------------
// Screenshot
// ---------------------------------------------------------------------------

// ScreenshotParams holds all parameters for the Screenshot endpoint.
type ScreenshotParams struct {
	// URL of the page to capture. Required.
	URL string `json:"url"`
	// Format is the output image format: "png" (default), "jpeg", "webp", or "pdf".
	Format string `json:"format,omitempty"`
	// Width of the viewport in pixels. Default: 1280.
	Width int `json:"width,omitempty"`
	// Height of the viewport in pixels.
	Height int `json:"height,omitempty"`
	// FullPage captures the entire scrollable page. Default: false.
	FullPage bool `json:"full_page,omitempty"`
	// Delay is the time in milliseconds to wait after page load.
	Delay int `json:"delay,omitempty"`
	// Quality sets JPEG/WebP compression quality (1-100). Default: 85.
	Quality int `json:"quality,omitempty"`
	// Scale is the device scale factor. Default: 1.
	Scale float64 `json:"scale,omitempty"`
	// BlockAds enables ad blocking. Default: false.
	BlockAds bool `json:"block_ads,omitempty"`
	// WaitForSelector waits for this CSS selector to appear before capturing.
	WaitForSelector string `json:"wait_for_selector,omitempty"`
	// Clip captures only the specified rectangular region.
	Clip *ClipRegion `json:"clip,omitempty"`
	// ScrollY scrolls the page by this many pixels before capturing.
	ScrollY int `json:"scroll_y,omitempty"`
	// CustomCSS is injected into the page before capturing.
	CustomCSS string `json:"custom_css,omitempty"`
	// CustomJS is executed on the page before capturing.
	CustomJS string `json:"custom_js,omitempty"`
	// Headers are additional HTTP headers sent when loading the page.
	Headers map[string]string `json:"headers,omitempty"`
	// UserAgent overrides the browser's User-Agent string.
	UserAgent string `json:"user_agent,omitempty"`
	// Proxy routes the browser request through this proxy URL.
	Proxy string `json:"proxy,omitempty"`
	// AccessKey is an alternative authentication method via query parameter.
	AccessKey string `json:"access_key,omitempty"`
	// Selector captures only the element matching this CSS selector.
	Selector string `json:"selector,omitempty"`
}

// ClipRegion defines a rectangular region for clipping screenshots.
type ClipRegion struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"w"`
	Height int `json:"h"`
}

// Screenshot captures a screenshot of a URL.
// Returns raw image bytes (PNG, JPEG, WebP, or PDF depending on Params.Format).
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

// ScreenshotToFile captures a screenshot and writes it directly to a file.
// The file is created with mode 0644. Returns the number of bytes written.
//
//	n, err := client.ScreenshotToFile(ctx, "output.png", snapapi.ScreenshotParams{
//	    URL:    "https://example.com",
//	    Format: "png",
//	})
func (c *Client) ScreenshotToFile(ctx context.Context, filename string, p ScreenshotParams) (int, error) {
	data, err := c.Screenshot(ctx, p)
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return 0, fmt.Errorf("snapapi: write file %q: %w", filename, err)
	}
	return len(data), nil
}

// ScreenshotToStorageParams are the parameters for ScreenshotToStorage.
type ScreenshotToStorageParams struct {
	ScreenshotParams
	// StorageKey is the object key / path to store the capture under.
	// If empty the server assigns a key.
	StorageKey string `json:"storage_key,omitempty"`
	// StorageBucket overrides the default storage bucket.
	StorageBucket string `json:"storage_bucket,omitempty"`
}

// StorageCapture is the response returned when a capture is saved to cloud storage.
type StorageCapture struct {
	// URL is the public URL of the stored capture.
	URL string `json:"url"`
	// Key is the storage object key.
	Key string `json:"key"`
	// Bucket is the storage bucket name.
	Bucket string `json:"bucket"`
	// Size is the file size in bytes.
	Size int64 `json:"size"`
	// ContentType is the MIME type of the stored file.
	ContentType string `json:"content_type"`
	// CreatedAt is the ISO 8601 timestamp of when the capture was stored.
	CreatedAt string `json:"created_at"`
}

// ScreenshotToStorage captures a screenshot and stores it directly in cloud
// storage managed by SnapAPI, returning a public URL and metadata.
//
//	capture, err := client.ScreenshotToStorage(ctx, ScreenshotToStorageParams{
//	    ScreenshotParams: snapapi.ScreenshotParams{
//	        URL:    "https://example.com",
//	        Format: "png",
//	    },
//	    StorageKey: "reports/home.png",
//	})
//	fmt.Println(capture.URL)
func (c *Client) ScreenshotToStorage(ctx context.Context, p ScreenshotToStorageParams) (*StorageCapture, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	var result StorageCapture
	if err := c.doJSON(ctx, http.MethodPost, "/v1/screenshot/storage", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Scrape
// ---------------------------------------------------------------------------

// ScrapeParams holds all parameters for the Scrape endpoint.
type ScrapeParams struct {
	// URL of the page to scrape. Required.
	URL string `json:"url"`
	// Selector is a CSS selector to scope scraping to a subtree.
	Selector string `json:"selector,omitempty"`
	// Format is the output format: "html" (default), "text", or "json".
	Format string `json:"format,omitempty"`
	// WaitForSelector waits for this CSS selector to appear before scraping.
	WaitForSelector string `json:"wait_for_selector,omitempty"`
	// Headers are additional HTTP headers sent when loading the page.
	Headers map[string]string `json:"headers,omitempty"`
	// Proxy routes the browser request through this proxy URL.
	Proxy string `json:"proxy,omitempty"`
	// AccessKey is an alternative authentication method via query parameter.
	AccessKey string `json:"access_key,omitempty"`
}

// ScrapeResult is the structured response from the scrape endpoint.
type ScrapeResult struct {
	// Data contains the scraped content (HTML, text, or JSON string).
	Data string `json:"data"`
	// URL is the final URL after any redirects.
	URL string `json:"url"`
	// Status is the HTTP status code of the scraped page.
	Status int `json:"status"`
}

// Scrape fetches text or structured content from a URL.
//
//	data, err := client.Scrape(ctx, snapapi.ScrapeParams{URL: "https://example.com"})
//	fmt.Println(data.Data)
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

// ScrapeText is a convenience wrapper that scrapes the page and returns only
// the plain-text content string (no metadata).
//
//	text, err := client.ScrapeText(ctx, "https://example.com")
func (c *Client) ScrapeText(ctx context.Context, url string) (string, error) {
	result, err := c.Scrape(ctx, ScrapeParams{URL: url, Format: "text"})
	if err != nil {
		return "", err
	}
	return result.Data, nil
}

// ScrapeHTML is a convenience wrapper that scrapes the page and returns only
// the raw HTML string (no metadata).
//
//	html, err := client.ScrapeHTML(ctx, "https://example.com")
func (c *Client) ScrapeHTML(ctx context.Context, url string) (string, error) {
	result, err := c.Scrape(ctx, ScrapeParams{URL: url, Format: "html"})
	if err != nil {
		return "", err
	}
	return result.Data, nil
}

// ---------------------------------------------------------------------------
// Extract
// ---------------------------------------------------------------------------

// ExtractParams holds all parameters for the Extract endpoint.
type ExtractParams struct {
	// URL of the page to extract content from. Required.
	URL string `json:"url"`
	// Format is the output format: "markdown" (default), "text", or "json".
	Format string `json:"format,omitempty"`
	// IncludeLinks includes hyperlinks in the output. Default: true.
	IncludeLinks *bool `json:"include_links,omitempty"`
	// IncludeImages includes image references in the output. Default: false.
	IncludeImages *bool `json:"include_images,omitempty"`
	// Selector scopes extraction to this CSS selector.
	Selector string `json:"selector,omitempty"`
	// WaitForSelector waits for this CSS selector to appear before extracting.
	WaitForSelector string `json:"wait_for_selector,omitempty"`
	// Headers are additional HTTP headers sent when loading the page.
	Headers map[string]string `json:"headers,omitempty"`
	// Proxy routes the browser request through this proxy URL.
	Proxy string `json:"proxy,omitempty"`
	// AccessKey is an alternative authentication method via query parameter.
	AccessKey string `json:"access_key,omitempty"`
}

// ExtractResult is the structured response from the extract endpoint.
type ExtractResult struct {
	// Content is the extracted text (markdown, plain text, or JSON).
	Content string `json:"content"`
	// URL is the final URL after any redirects.
	URL string `json:"url"`
	// WordCount is the approximate number of words in the extracted content.
	WordCount int `json:"word_count"`
}

// Extract extracts readable content from a URL, suitable for LLM consumption.
//
//	content, err := client.Extract(ctx, snapapi.ExtractParams{
//	    URL:    "https://example.com",
//	    Format: "markdown",
//	})
//	fmt.Println(content.Content)
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

// ExtractMarkdown is a convenience wrapper that extracts page content as
// Markdown, returning only the content string.
//
//	md, err := client.ExtractMarkdown(ctx, "https://example.com/blog/post")
func (c *Client) ExtractMarkdown(ctx context.Context, url string) (string, error) {
	result, err := c.Extract(ctx, ExtractParams{URL: url, Format: "markdown"})
	if err != nil {
		return "", err
	}
	return result.Content, nil
}

// ExtractText is a convenience wrapper that extracts page content as plain
// text, returning only the content string.
//
//	text, err := client.ExtractText(ctx, "https://example.com/blog/post")
func (c *Client) ExtractText(ctx context.Context, url string) (string, error) {
	result, err := c.Extract(ctx, ExtractParams{URL: url, Format: "text"})
	if err != nil {
		return "", err
	}
	return result.Content, nil
}

// ---------------------------------------------------------------------------
// Analyze
// ---------------------------------------------------------------------------

// AnalyzeParams holds all parameters for the Analyze endpoint.
type AnalyzeParams struct {
	// URL of the page to analyze. Required.
	URL string `json:"url"`
	// Prompt is the instruction for the LLM (e.g. "Summarize this page").
	Prompt string `json:"prompt,omitempty"`
	// Provider is the LLM provider: "openai", "anthropic", or "google".
	Provider string `json:"provider,omitempty"`
	// APIKey is the LLM provider API key.
	APIKey string `json:"apiKey,omitempty"`
	// JSONSchema constrains the LLM output to match a JSON schema.
	JSONSchema map[string]interface{} `json:"jsonSchema,omitempty"`
}

// AnalyzeResult is the structured response from the analyze endpoint.
type AnalyzeResult struct {
	// Result is the LLM's analysis output.
	Result string `json:"result"`
	// URL is the analyzed URL.
	URL string `json:"url"`
}

// Analyze sends a page to an LLM for analysis.
// Note: This endpoint may return HTTP 503 if LLM credits are exhausted.
//
//	result, err := client.Analyze(ctx, snapapi.AnalyzeParams{
//	    URL:      "https://example.com",
//	    Prompt:   "Summarize this page in 3 sentences.",
//	    Provider: "openai",
//	    APIKey:   "sk-...",
//	})
//	fmt.Println(result.Result)
func (c *Client) Analyze(ctx context.Context, p AnalyzeParams) (*AnalyzeResult, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	var result AnalyzeResult
	if err := c.doJSON(ctx, http.MethodPost, "/v1/analyze", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// PDF
// ---------------------------------------------------------------------------

// PDFParams holds parameters for PDF generation.
type PDFParams struct {
	// URL of the page to convert. Required (unless HTML is set).
	URL string `json:"url,omitempty"`
	// HTML is raw HTML to convert to PDF.
	HTML string `json:"html,omitempty"`
	// PageSize is the paper size: "a4" (default), "letter", etc.
	PageSize string `json:"page_size,omitempty"`
	// Landscape rotates the page to landscape orientation.
	Landscape bool `json:"landscape,omitempty"`
	// MarginTop sets the top margin (e.g. "10mm", "1cm").
	MarginTop string `json:"margin_top,omitempty"`
	// MarginBottom sets the bottom margin.
	MarginBottom string `json:"margin_bottom,omitempty"`
	// MarginLeft sets the left margin.
	MarginLeft string `json:"margin_left,omitempty"`
	// MarginRight sets the right margin.
	MarginRight string `json:"margin_right,omitempty"`
}

// PDF generates a PDF of a URL or HTML content.
// Uses the screenshot endpoint with format=pdf.
//
//	pdfBytes, err := client.PDF(ctx, snapapi.PDFParams{URL: "https://example.com"})
func (c *Client) PDF(ctx context.Context, p PDFParams) ([]byte, error) {
	if p.URL == "" && p.HTML == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL or HTML is required", StatusCode: 400}
	}
	body := map[string]interface{}{
		"format": "pdf",
	}
	if p.URL != "" {
		body["url"] = p.URL
	}
	if p.HTML != "" {
		body["html"] = p.HTML
	}
	if p.PageSize != "" {
		body["page_size"] = p.PageSize
	}
	if p.Landscape {
		body["landscape"] = true
	}
	if p.MarginTop != "" {
		body["margin_top"] = p.MarginTop
	}
	if p.MarginBottom != "" {
		body["margin_bottom"] = p.MarginBottom
	}
	if p.MarginLeft != "" {
		body["margin_left"] = p.MarginLeft
	}
	if p.MarginRight != "" {
		body["margin_right"] = p.MarginRight
	}
	return c.doRaw(ctx, http.MethodPost, "/v1/screenshot", body)
}

// PDFToFile generates a PDF and writes it directly to a file.
// Returns the number of bytes written.
func (c *Client) PDFToFile(ctx context.Context, filename string, p PDFParams) (int, error) {
	data, err := c.PDF(ctx, p)
	if err != nil {
		return 0, err
	}
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return 0, fmt.Errorf("snapapi: write file %q: %w", filename, err)
	}
	return len(data), nil
}

// ---------------------------------------------------------------------------
// OG Image
// ---------------------------------------------------------------------------

// OGImageParams holds parameters for OG image generation.
type OGImageParams struct {
	// URL of the page. Required.
	URL string `json:"url"`
	// Format: "png" (default), "jpeg", "webp".
	Format string `json:"format,omitempty"`
	// Width of the OG image. Default: 1200.
	Width int `json:"width,omitempty"`
	// Height of the OG image. Default: 630.
	Height int `json:"height,omitempty"`
}

// OGImage generates an Open Graph social image for a URL.
// Uses the screenshot endpoint with OG-standard dimensions.
//
//	ogBytes, err := client.OGImage(ctx, snapapi.OGImageParams{URL: "https://example.com"})
func (c *Client) OGImage(ctx context.Context, p OGImageParams) ([]byte, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	width := p.Width
	if width == 0 {
		width = 1200
	}
	height := p.Height
	if height == 0 {
		height = 630
	}
	screenshotParams := ScreenshotParams{
		URL:    p.URL,
		Format: p.Format,
		Width:  width,
		Height: height,
	}
	return c.Screenshot(ctx, screenshotParams)
}

// ---------------------------------------------------------------------------
// Video
// ---------------------------------------------------------------------------

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

// Video records a short video of a URL.
//
//	videoBytes, err := client.Video(ctx, snapapi.VideoParams{URL: "https://example.com"})
func (c *Client) Video(ctx context.Context, p VideoParams) ([]byte, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	return c.doRaw(ctx, http.MethodPost, "/v1/video", p)
}

// ---------------------------------------------------------------------------
// Usage
// ---------------------------------------------------------------------------

// UsageResult is the response from GET /v1/usage.
type UsageResult struct {
	Used      int    `json:"used"`
	Limit     int    `json:"limit"`
	Total     int    `json:"total"`
	Remaining int    `json:"remaining"`
	ResetAt   string `json:"resetAt,omitempty"`
}

// GetUsage returns the caller's current API usage statistics.
//
//	usage, err := client.GetUsage(ctx)
//	fmt.Printf("Used: %d / %d\n", usage.Used, usage.Limit)
func (c *Client) GetUsage(ctx context.Context) (*UsageResult, error) {
	var result UsageResult
	if err := c.doJSON(ctx, http.MethodGet, "/v1/usage", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Quota is an alias for GetUsage, kept for backward compatibility.
func (c *Client) Quota(ctx context.Context) (*UsageResult, error) {
	return c.GetUsage(ctx)
}

// ---------------------------------------------------------------------------
// Ping
// ---------------------------------------------------------------------------

// PingResult is the response from GET /v1/ping.
type PingResult struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

// Ping checks the API health.
//
//	result, err := client.Ping(ctx)
//	fmt.Println(result.Status)
func (c *Client) Ping(ctx context.Context) (*PingResult, error) {
	var result PingResult
	if err := c.doJSON(ctx, http.MethodGet, "/v1/ping", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Internal transport
// ---------------------------------------------------------------------------

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
		// use exponential back-off.
		waitDur := delay
		if apiErr != nil && apiErr.RetryAfter > 0 {
			waitDur = time.Duration(apiErr.RetryAfter) * time.Second
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
