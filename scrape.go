package snapapi

import (
	"context"
	"net/http"
)

// ScrapeParams holds all parameters for the Scrape endpoint.
type ScrapeParams struct {
	// URL of the page to scrape. Required.
	URL string `json:"url"`
	// Selector is a CSS selector to scope scraping to a subtree.
	Selector string `json:"selector,omitempty"`
	// Selectors is a map of named CSS selectors to extract multiple elements.
	// Each key is a name and each value is a CSS selector string.
	Selectors map[string]string `json:"selectors,omitempty"`
	// Format is the output format: "html" (default), "text", or "json".
	Format string `json:"format,omitempty"`
	// WaitFor is a CSS selector or timeout to wait for before scraping.
	WaitFor string `json:"waitFor,omitempty"`
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
