package snapapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

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

// GeneratePDF is an alias for PDF, provided for convenience.
//
//	pdfBytes, err := client.GeneratePDF(ctx, snapapi.PDFParams{URL: "https://example.com"})
func (c *Client) GeneratePDF(ctx context.Context, p PDFParams) ([]byte, error) {
	return c.PDF(ctx, p)
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
