package snapapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// ScreenshotGeolocation holds GPS coordinates for browser geolocation emulation.
type ScreenshotGeolocation struct {
	// Latitude in decimal degrees (-90 to 90).
	Latitude float64 `json:"latitude"`
	// Longitude in decimal degrees (-180 to 180).
	Longitude float64 `json:"longitude"`
	// Accuracy in metres (optional, default: 1).
	Accuracy float64 `json:"accuracy,omitempty"`
}

// ScreenshotProxy holds proxy server configuration.
type ScreenshotProxy struct {
	// Server is the proxy URL, e.g. "http://host:port".
	Server string `json:"server"`
	// Username for proxy authentication (optional).
	Username string `json:"username,omitempty"`
	// Password for proxy authentication (optional).
	Password string `json:"password,omitempty"`
	// Bypass is a list of domains that bypass the proxy (optional).
	Bypass []string `json:"bypass,omitempty"`
}

// ScreenshotCookie defines a browser cookie to inject into the page session.
type ScreenshotCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Expires  int64  `json:"expires,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	SameSite string `json:"sameSite,omitempty"` // "Strict" | "Lax" | "None"
}

// ScreenshotHTTPAuth holds HTTP Basic Authentication credentials.
type ScreenshotHTTPAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ScreenshotParams holds all parameters for the Screenshot endpoint.
type ScreenshotParams struct {
	// URL of the page to capture. Required (unless HTML or Markdown is set).
	URL string `json:"url,omitempty"`
	// HTML is raw HTML to render instead of a URL.
	HTML string `json:"html,omitempty"`
	// Markdown is rendered to HTML before capturing.
	Markdown string `json:"markdown,omitempty"`
	// Format is the output image format: "png" (default), "jpeg", "webp", "avif", or "pdf".
	Format string `json:"format,omitempty"`
	// Width of the viewport in pixels. Default: 1280.
	Width int `json:"width,omitempty"`
	// Height of the viewport in pixels. Default: 800.
	Height int `json:"height,omitempty"`
	// DeviceScaleFactor is the device pixel ratio (1–3). Default: 1.
	DeviceScaleFactor float64 `json:"deviceScaleFactor,omitempty"`
	// Device is a named device viewport preset (overrides Width/Height/DeviceScaleFactor).
	Device string `json:"device,omitempty"`
	// IsMobile emulates a mobile device.
	IsMobile bool `json:"isMobile,omitempty"`
	// HasTouch enables touch events.
	HasTouch bool `json:"hasTouch,omitempty"`
	// IsLandscape emulates landscape orientation.
	IsLandscape bool `json:"isLandscape,omitempty"`
	// FullPage captures the entire scrollable page. Default: false.
	FullPage bool `json:"fullPage,omitempty"`
	// FullPageScrollDelay is the delay between scroll steps in ms. Default: 400.
	FullPageScrollDelay int `json:"fullPageScrollDelay,omitempty"`
	// FullPageMaxHeight is the maximum height for full-page capture in px.
	FullPageMaxHeight int `json:"fullPageMaxHeight,omitempty"`
	// ScrollToBottom scrolls to the bottom of the page before capturing.
	ScrollToBottom bool `json:"scrollToBottom,omitempty"`
	// DarkMode enables dark mode (prefers-color-scheme: dark). Default: false.
	DarkMode bool `json:"darkMode,omitempty"`
	// ReducedMotion reduces CSS animations.
	ReducedMotion bool `json:"reducedMotion,omitempty"`
	// Delay is the time in milliseconds to wait after page load (0–30000).
	Delay int `json:"delay,omitempty"`
	// Quality sets JPEG/WebP compression quality (1–100).
	Quality int `json:"quality,omitempty"`
	// BlockAds enables ad blocking. Default: false.
	BlockAds bool `json:"blockAds,omitempty"`
	// BlockTrackers blocks tracking scripts.
	BlockTrackers bool `json:"blockTrackers,omitempty"`
	// BlockCookieBanners blocks cookie consent banners. Default: false.
	BlockCookieBanners bool `json:"blockCookieBanners,omitempty"`
	// BlockChatWidgets blocks chat/support widgets.
	BlockChatWidgets bool `json:"blockChatWidgets,omitempty"`
	// WaitUntil is the navigation event to wait for: "load", "domcontentloaded", "networkidle".
	WaitUntil string `json:"waitUntil,omitempty"`
	// WaitForSelector waits for this CSS selector to appear before capturing.
	WaitForSelector string `json:"waitForSelector,omitempty"`
	// Selector captures only the element matching this CSS selector.
	Selector string `json:"selector,omitempty"`
	// Clip captures only the specified rectangular region.
	Clip *ClipRegion `json:"clip,omitempty"`
	// ScrollY scrolls the page by this many pixels before capturing.
	ScrollY int `json:"scroll_y,omitempty"`
	// CSS is injected into the page before capturing.
	CSS string `json:"css,omitempty"`
	// JavaScript is executed on the page before capturing.
	JavaScript string `json:"javascript,omitempty"`
	// HideSelectors is a list of CSS selectors to hide before capturing.
	HideSelectors []string `json:"hideSelectors,omitempty"`
	// ClickSelector clicks this CSS selector before capturing.
	ClickSelector string `json:"clickSelector,omitempty"`
	// Headers are additional HTTP headers sent when loading the page.
	Headers map[string]string `json:"extraHeaders,omitempty"`
	// UserAgent overrides the browser's User-Agent string.
	UserAgent string `json:"userAgent,omitempty"`
	// Cookies are injected into the browser session before loading the page.
	Cookies []ScreenshotCookie `json:"cookies,omitempty"`
	// HTTPAuth provides HTTP Basic Authentication credentials.
	HTTPAuth *ScreenshotHTTPAuth `json:"httpAuth,omitempty"`
	// Proxy routes the browser request through a custom proxy server.
	Proxy *ScreenshotProxy `json:"proxy,omitempty"`
	// PremiumProxy uses SnapAPI's managed residential proxy network.
	PremiumProxy bool `json:"premiumProxy,omitempty"`
	// Geolocation emulates GPS coordinates in the browser.
	Geolocation *ScreenshotGeolocation `json:"geolocation,omitempty"`
	// Timezone sets the browser timezone (IANA string, e.g. "America/New_York").
	Timezone string `json:"timezone,omitempty"`
	// Locale sets the browser locale (e.g. "en-US").
	Locale string `json:"locale,omitempty"`
	// Cache enables response caching. Default: false.
	Cache bool `json:"cache,omitempty"`
	// CacheTTL is the cache time-to-live in seconds (60–2592000). Default: 86400.
	CacheTTL int `json:"cacheTtl,omitempty"`
	// WebhookURL delivers the result asynchronously to this URL.
	WebhookURL string `json:"webhookUrl,omitempty"`
	// Async queues the capture and returns a job ID immediately.
	Async bool `json:"async,omitempty"`
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
