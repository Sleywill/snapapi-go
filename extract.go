package snapapi

import (
	"context"
	"net/http"
)

// ExtractParams holds all parameters for the Extract endpoint.
type ExtractParams struct {
	// URL of the page to extract content from. Required.
	URL string `json:"url"`
	// Format is the output format: "markdown" (default), "text", or "html".
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

// extractAPIResponse is the raw shape the SnapAPI server returns for /v1/extract.
// The API returns {"success":true,"type":"markdown","url":"...","data":"...","responseTime":N}.
type extractAPIResponse struct {
	Success      bool   `json:"success"`
	Type         string `json:"type"`
	URL          string `json:"url"`
	Data         string `json:"data"`
	ResponseTime int    `json:"responseTime"`
}

// ExtractResult is the structured response from the extract endpoint.
type ExtractResult struct {
	// Content is the extracted text (markdown, plain text, or JSON).
	// Populated from the API's "data" field.
	Content string `json:"content"`
	// URL is the final URL after any redirects.
	URL string `json:"url"`
	// WordCount is the approximate number of words in the extracted content.
	WordCount int `json:"word_count"`
	// ResponseTime is the server-side render duration in milliseconds.
	ResponseTime int `json:"responseTime"`
	// Type is the output format that was actually returned (e.g. "markdown").
	Type string `json:"type"`
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
	// The API returns {"success":true,"type":"markdown","url":"...","data":"..."}.
	// We map the "data" field to Content for a consistent SDK interface.
	var raw extractAPIResponse
	if err := c.doJSON(ctx, http.MethodPost, "/v1/extract", p, &raw); err != nil {
		return nil, err
	}
	return &ExtractResult{
		Content:      raw.Data,
		URL:          raw.URL,
		ResponseTime: raw.ResponseTime,
		Type:         raw.Type,
	}, nil
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
