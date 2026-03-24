package snapapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

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
	// DarkMode enables dark mode (prefers-color-scheme: dark). Default: false.
	DarkMode bool `json:"dark_mode,omitempty"`
	// Delay is the time in milliseconds to wait after page load.
	Delay int `json:"delay,omitempty"`
	// Quality sets JPEG/WebP compression quality (1-100). Default: 85.
	Quality int `json:"quality,omitempty"`
	// Scale is the device scale factor. Default: 1.
	Scale float64 `json:"scale,omitempty"`
	// BlockAds enables ad blocking. Default: false.
	BlockAds bool `json:"block_ads,omitempty"`
	// BlockCookies blocks cookie consent banners. Default: false.
	BlockCookies bool `json:"block_cookies,omitempty"`
	// WaitForSelector waits for this CSS selector to appear before capturing.
	WaitForSelector string `json:"wait_for_selector,omitempty"`
	// Selector captures only the element matching this CSS selector.
	Selector string `json:"selector,omitempty"`
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

// GenerateOGImage is an alias for OGImage, provided for convenience.
//
//	ogBytes, err := client.GenerateOGImage(ctx, snapapi.OGImageParams{URL: "https://example.com"})
func (c *Client) GenerateOGImage(ctx context.Context, p OGImageParams) ([]byte, error) {
	return c.OGImage(ctx, p)
}
