// Package snapapi provides a Go client for the SnapAPI screenshot service.
//
// Example usage:
//
//	client := snapapi.NewClient("sk_live_xxx")
//	data, err := client.Screenshot(snapapi.ScreenshotOptions{
//	    URL: "https://example.com",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("screenshot.png", data, 0644)
package snapapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.snapapi.pics"
	defaultTimeout = 60 * time.Second
	userAgent      = "snapapi-go/1.2.0"
)

// Client is a SnapAPI client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ClientOption is a function that configures a Client.
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

// WithHTTPClient sets a custom HTTP client.
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

	return c
}

// Screenshot captures a screenshot of the specified URL or HTML content.
// Returns the raw image bytes for binary/base64 response types,
// or use ScreenshotWithMetadata for JSON response with metadata.
func (c *Client) Screenshot(opts ScreenshotOptions) ([]byte, error) {
	if opts.URL == "" && opts.HTML == "" && opts.Markdown == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "URL, HTML, or Markdown is required",
			StatusCode: 400,
		}
	}

	return c.doRequest("POST", "/v1/screenshot", opts)
}

// ScreenshotWithMetadata captures a screenshot and returns metadata.
func (c *Client) ScreenshotWithMetadata(opts ScreenshotOptions) (*ScreenshotResult, error) {
	opts.ResponseType = "json"

	data, err := c.Screenshot(opts)
	if err != nil {
		return nil, err
	}

	var result ScreenshotResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// Batch captures screenshots of multiple URLs.
func (c *Client) Batch(opts BatchOptions) (*BatchResult, error) {
	if len(opts.URLs) == 0 {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "URLs are required",
			StatusCode: 400,
		}
	}

	data, err := c.doRequest("POST", "/v1/screenshot/batch", opts)
	if err != nil {
		return nil, err
	}

	var result BatchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetBatchStatus retrieves the status of a batch job.
func (c *Client) GetBatchStatus(jobID string) (*BatchStatus, error) {
	data, err := c.doRequest("GET", "/v1/screenshot/batch/"+jobID, nil)
	if err != nil {
		return nil, err
	}

	var status BatchStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &status, nil
}

// GetUsage retrieves your API usage statistics.
func (c *Client) GetUsage() (*Usage, error) {
	data, err := c.doRequest("GET", "/v1/usage", nil)
	if err != nil {
		return nil, err
	}

	var usage Usage
	if err := json.Unmarshal(data, &usage); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &usage, nil
}

// ScreenshotFromHTML captures a screenshot from HTML content.
func (c *Client) ScreenshotFromHTML(html string, opts *ScreenshotOptions) ([]byte, error) {
	if html == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "HTML content is required",
			StatusCode: 400,
		}
	}

	if opts == nil {
		opts = &ScreenshotOptions{}
	}
	opts.HTML = html
	opts.URL = ""

	return c.doRequest("POST", "/v1/screenshot", opts)
}

// ScreenshotDevice captures a screenshot using a device preset.
func (c *Client) ScreenshotDevice(url string, device DevicePreset, opts *ScreenshotOptions) ([]byte, error) {
	if url == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "URL is required",
			StatusCode: 400,
		}
	}

	if opts == nil {
		opts = &ScreenshotOptions{}
	}
	opts.URL = url
	opts.Device = device

	return c.doRequest("POST", "/v1/screenshot", opts)
}

// PDF generates a PDF from a URL or HTML content.
func (c *Client) PDF(opts ScreenshotOptions) ([]byte, error) {
	if opts.URL == "" && opts.HTML == "" && opts.Markdown == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "URL, HTML, or Markdown is required",
			StatusCode: 400,
		}
	}

	opts.Format = "pdf"
	opts.ResponseType = "binary"

	return c.doRequest("POST", "/v1/screenshot", opts)
}

// PDFFromHTML generates a PDF from HTML content.
func (c *Client) PDFFromHTML(html string, pdfOpts *PDFOptions) ([]byte, error) {
	if html == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "HTML content is required",
			StatusCode: 400,
		}
	}

	opts := ScreenshotOptions{
		HTML:         html,
		Format:       "pdf",
		ResponseType: "binary",
		PDFOptions:   pdfOpts,
	}

	return c.doRequest("POST", "/v1/screenshot", opts)
}

// Video captures a video of a webpage with optional scroll animation.
func (c *Client) Video(opts VideoOptions) ([]byte, error) {
	if opts.URL == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "URL is required",
			StatusCode: 400,
		}
	}

	return c.doRequest("POST", "/v1/video", opts)
}

// VideoWithResult captures a video and returns structured result with metadata.
func (c *Client) VideoWithResult(opts VideoOptions) (*VideoResult, error) {
	if opts.URL == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "URL is required",
			StatusCode: 400,
		}
	}

	opts.ResponseType = "json"
	data, err := c.doRequest("POST", "/v1/video", opts)
	if err != nil {
		return nil, err
	}

	var result VideoResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetDevices retrieves available device presets.
func (c *Client) GetDevices() (*DevicesResult, error) {
	data, err := c.doRequest("GET", "/v1/devices", nil)
	if err != nil {
		return nil, err
	}

	var result DevicesResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetCapabilities retrieves API capabilities and features.
func (c *Client) GetCapabilities() (*CapabilitiesResult, error) {
	data, err := c.doRequest("GET", "/v1/capabilities", nil)
	if err != nil {
		return nil, err
	}

	var result CapabilitiesResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ScreenshotFromMarkdown captures a screenshot from Markdown content.
func (c *Client) ScreenshotFromMarkdown(markdown string, opts *ScreenshotOptions) ([]byte, error) {
	if markdown == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "Markdown content is required",
			StatusCode: 400,
		}
	}

	if opts == nil {
		opts = &ScreenshotOptions{}
	}
	opts.Markdown = markdown
	opts.URL = ""
	opts.HTML = ""

	return c.doRequest("POST", "/v1/screenshot", opts)
}

// Extract extracts content from a webpage.
func (c *Client) Extract(opts ExtractOptions) (*ExtractResult, error) {
	if opts.URL == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "URL is required",
			StatusCode: 400,
		}
	}

	data, err := c.doRequest("POST", "/v1/extract", opts)
	if err != nil {
		return nil, err
	}

	var result ExtractResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ExtractMarkdown extracts content from a webpage as Markdown.
func (c *Client) ExtractMarkdown(url string) (*ExtractResult, error) {
	return c.Extract(ExtractOptions{URL: url, Type: "markdown"})
}

// ExtractArticle extracts article content from a webpage.
func (c *Client) ExtractArticle(url string) (*ExtractResult, error) {
	return c.Extract(ExtractOptions{URL: url, Type: "article"})
}

// ExtractStructured extracts structured content from a webpage.
func (c *Client) ExtractStructured(url string) (*ExtractResult, error) {
	return c.Extract(ExtractOptions{URL: url, Type: "structured"})
}

// ExtractText extracts plain text content from a webpage.
func (c *Client) ExtractText(url string) (*ExtractResult, error) {
	return c.Extract(ExtractOptions{URL: url, Type: "text"})
}

// ExtractLinks extracts links from a webpage.
func (c *Client) ExtractLinks(url string) (*ExtractResult, error) {
	return c.Extract(ExtractOptions{URL: url, Type: "links"})
}

// ExtractImages extracts image URLs from a webpage.
func (c *Client) ExtractImages(url string) (*ExtractResult, error) {
	return c.Extract(ExtractOptions{URL: url, Type: "images"})
}

// ExtractMetadata extracts metadata from a webpage.
func (c *Client) ExtractMetadata(url string) (*ExtractResult, error) {
	return c.Extract(ExtractOptions{URL: url, Type: "metadata"})
}

// Analyze performs AI-powered analysis of a webpage.
func (c *Client) Analyze(opts AnalyzeOptions) (*AnalyzeResult, error) {
	if opts.URL == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "URL is required",
			StatusCode: 400,
		}
	}
	if opts.Prompt == "" {
		return nil, &APIError{
			Code:       ErrInvalidParams,
			Message:    "Prompt is required",
			StatusCode: 400,
		}
	}

	data, err := c.doRequest("POST", "/v1/analyze", opts)
	if err != nil {
		return nil, err
	}

	var result AnalyzeResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// doRequest performs an HTTP request to the API.
func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &APIError{
			Code:       ErrConnectionError,
			Message:    fmt.Sprintf("connection error: %v", err),
			StatusCode: 0,
		}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.handleError(respBody, resp.StatusCode)
	}

	return respBody, nil
}

// handleError parses an error response from the API.
func (c *Client) handleError(body []byte, statusCode int) error {
	var errResp errorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return &APIError{
			Code:       "HTTP_ERROR",
			Message:    fmt.Sprintf("HTTP %d", statusCode),
			StatusCode: statusCode,
		}
	}

	return &APIError{
		Code:       errResp.Error.Code,
		Message:    errResp.Error.Message,
		StatusCode: statusCode,
		Details:    errResp.Error.Details,
	}
}
