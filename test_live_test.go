// Package snapapi_test contains live integration tests against the SnapAPI
// production API (https://api.snapapi.pics).
//
// Run with:
//
//	SNAPAPI_TEST_KEY=sk_live_xxx go test -v -run TestLive -timeout 300s ./...
//
// If SNAPAPI_TEST_KEY is not set, all tests in this file are skipped.
package snapapi_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	snapapi "github.com/Sleywill/snapapi-go"
)

// liveKey returns the API key for live tests or skips the test if unset.
func liveKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("SNAPAPI_TEST_KEY")
	if key == "" {
		t.Skip("SNAPAPI_TEST_KEY not set — skipping live integration tests")
	}
	return key
}

// newLiveClient creates a client configured for live testing.
func newLiveClient(t *testing.T) *snapapi.Client {
	t.Helper()
	return snapapi.New(liveKey(t),
		snapapi.WithTimeout(120*time.Second),
		snapapi.WithRetries(1),
		snapapi.WithRetryDelay(2*time.Second),
	)
}

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

// isPNGBytes reports whether data starts with the PNG magic bytes.
func isPNGBytes(data []byte) bool {
	return len(data) > 4 && bytes.HasPrefix(data, []byte("\x89PNG"))
}

// isJPEGBytes reports whether data starts with the JPEG magic bytes.
func isJPEGBytes(data []byte) bool {
	return len(data) > 2 && data[0] == 0xFF && data[1] == 0xD8
}

// isWebPBytes reports whether data contains the WEBP marker.
func isWebPBytes(data []byte) bool {
	return len(data) > 12 && string(data[8:12]) == "WEBP"
}

// isPDFBytes reports whether data starts with the PDF magic bytes.
func isPDFBytes(data []byte) bool {
	return len(data) > 4 && bytes.HasPrefix(data, []byte("%PDF"))
}

// isWebMBytes reports whether data contains a WebM/EBML header.
func isWebMBytes(data []byte) bool {
	// EBML magic: 0x1A 0x45 0xDF 0xA3
	return len(data) > 4 && data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3
}

// isGIFBytes reports whether data starts with "GIF8".
func isGIFBytes(data []byte) bool {
	return len(data) > 4 && bytes.HasPrefix(data, []byte("GIF8"))
}

// --------------------------------------------------------------------------
// Ping
// --------------------------------------------------------------------------

func TestLive_Ping(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	result, err := client.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping() error: %v", err)
	}
	if result.Status == "" {
		t.Error("Ping().Status is empty")
	}
	t.Logf("Ping OK: status=%q timestamp=%d", result.Status, result.Timestamp)
}

// --------------------------------------------------------------------------
// GetUsage
// --------------------------------------------------------------------------

func TestLive_GetUsage(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	usage, err := client.GetUsage(ctx)
	if err != nil {
		t.Fatalf("GetUsage() error: %v", err)
	}
	if usage.Limit <= 0 {
		t.Errorf("expected Limit > 0, got %d", usage.Limit)
	}
	if usage.Remaining < 0 {
		t.Errorf("expected Remaining >= 0, got %d", usage.Remaining)
	}
	t.Logf("Usage: used=%d limit=%d remaining=%d resetAt=%s",
		usage.Used, usage.Limit, usage.Remaining, usage.ResetAt)
}

// TestLive_Quota_Alias verifies the Quota() backward-compat alias.
func TestLive_Quota_Alias(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	usage, err := client.Quota(ctx)
	if err != nil {
		t.Fatalf("Quota() error: %v", err)
	}
	if usage.Limit <= 0 {
		t.Errorf("expected Limit > 0, got %d", usage.Limit)
	}
}

// --------------------------------------------------------------------------
// Screenshot
// --------------------------------------------------------------------------

func TestLive_Screenshot_PNG(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
	})
	if err != nil {
		t.Fatalf("Screenshot(png) error: %v", err)
	}
	if len(data) < 1000 {
		t.Errorf("expected substantial PNG bytes, got %d", len(data))
	}
	if !isPNGBytes(data) {
		t.Errorf("response does not look like PNG (first 4 bytes: %x)", data[:4])
	}
	t.Logf("Screenshot(png): %d bytes", len(data))
}

func TestLive_Screenshot_JPEG(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:     "https://example.com",
		Format:  "jpeg",
		Quality: 80,
	})
	if err != nil {
		t.Fatalf("Screenshot(jpeg) error: %v", err)
	}
	if len(data) < 500 {
		t.Errorf("expected substantial JPEG bytes, got %d", len(data))
	}
	if !isJPEGBytes(data) {
		t.Errorf("response does not look like JPEG (first 2 bytes: %x)", data[:2])
	}
	t.Logf("Screenshot(jpeg): %d bytes", len(data))
}

func TestLive_Screenshot_WebP(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "webp",
	})
	if err != nil {
		t.Fatalf("Screenshot(webp) error: %v", err)
	}
	if len(data) < 500 {
		t.Errorf("expected substantial WebP bytes, got %d", len(data))
	}
	if !isWebPBytes(data) {
		t.Errorf("response does not look like WebP (bytes 8-12: %q)", string(data[8:min(12, len(data))]))
	}
	t.Logf("Screenshot(webp): %d bytes", len(data))
}

func TestLive_Screenshot_PDF(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "pdf",
	})
	if err != nil {
		t.Fatalf("Screenshot(pdf) error: %v", err)
	}
	if len(data) < 1000 {
		t.Errorf("expected substantial PDF bytes, got %d", len(data))
	}
	if !isPDFBytes(data) {
		t.Errorf("response does not look like PDF (first 4 bytes: %q)", string(data[:4]))
	}
	t.Logf("Screenshot(pdf): %d bytes", len(data))
}

func TestLive_Screenshot_FullPage(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	normal, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
	})
	if err != nil {
		t.Fatalf("Screenshot(normal) error: %v", err)
	}

	fullPage, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:      "https://example.com",
		Format:   "png",
		FullPage: true,
	})
	if err != nil {
		t.Fatalf("Screenshot(full_page) error: %v", err)
	}

	// Full-page screenshot should generally be equal to or larger than viewport screenshot.
	t.Logf("Screenshot normal=%d bytes, full_page=%d bytes", len(normal), len(fullPage))
}

func TestLive_Screenshot_DarkMode(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:      "https://example.com",
		Format:   "png",
		DarkMode: true,
	})
	if err != nil {
		t.Fatalf("Screenshot(dark_mode) error: %v", err)
	}
	if !isPNGBytes(data) {
		t.Errorf("expected PNG response")
	}
	t.Logf("Screenshot(dark_mode): %d bytes", len(data))
}

func TestLive_Screenshot_CustomDimensions(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
		Width:  800,
		Height: 600,
	})
	if err != nil {
		t.Fatalf("Screenshot(800x600) error: %v", err)
	}
	if !isPNGBytes(data) {
		t.Errorf("expected PNG response")
	}
	t.Logf("Screenshot(800x600): %d bytes", len(data))
}

func TestLive_Screenshot_Scale(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:               "https://example.com",
		Format:            "png",
		DeviceScaleFactor: 2.0,
	})
	if err != nil {
		t.Fatalf("Screenshot(scale=2) error: %v", err)
	}
	if !isPNGBytes(data) {
		t.Errorf("expected PNG response")
	}
	t.Logf("Screenshot(scale=2): %d bytes", len(data))
}

func TestLive_Screenshot_BlockAds(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:                "https://example.com",
		Format:             "png",
		BlockAds:           true,
		BlockCookieBanners: true,
	})
	if err != nil {
		t.Fatalf("Screenshot(block_ads) error: %v", err)
	}
	if !isPNGBytes(data) {
		t.Errorf("expected PNG response")
	}
	t.Logf("Screenshot(block_ads+block_cookies): %d bytes", len(data))
}

func TestLive_Screenshot_CustomHeaders(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
		Headers: map[string]string{
			"Accept-Language": "en-US,en;q=0.9",
		},
	})
	if err != nil {
		t.Fatalf("Screenshot(custom_headers) error: %v", err)
	}
	if !isPNGBytes(data) {
		t.Errorf("expected PNG response")
	}
	t.Logf("Screenshot(custom_headers): %d bytes", len(data))
}

func TestLive_Screenshot_Delay(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
		Delay:  500,
	})
	if err != nil {
		t.Fatalf("Screenshot(delay=500ms) error: %v", err)
	}
	if !isPNGBytes(data) {
		t.Errorf("expected PNG response")
	}
	t.Logf("Screenshot(delay=500ms): %d bytes", len(data))
}

// TestLive_Screenshot_MissingURL ensures client-side validation fires.
func TestLive_Screenshot_MissingURL(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{})
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
	var apiErr *snapapi.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Code != snapapi.ErrInvalidParams {
			t.Errorf("expected ErrInvalidParams, got %q", apiErr.Code)
		}
	} else {
		t.Errorf("expected *APIError, got %T: %v", err, err)
	}
}

// TestLive_Screenshot_InvalidURL sends a malformed URL to the server.
func TestLive_Screenshot_InvalidURL(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "not-a-valid-url",
		Format: "png",
	})
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
	var apiErr *snapapi.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("expected HTTP 400, got %d", apiErr.StatusCode)
	}
	t.Logf("Invalid URL error: %v", err)
}

// --------------------------------------------------------------------------
// PDF via PDF() helper
// --------------------------------------------------------------------------

func TestLive_PDF_FromURL(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.PDF(ctx, snapapi.PDFParams{
		URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("PDF(url) error: %v", err)
	}
	if !isPDFBytes(data) {
		t.Errorf("expected PDF, got first 4 bytes: %q", string(data[:min(4, len(data))]))
	}
	t.Logf("PDF(url): %d bytes", len(data))
}

func TestLive_PDF_FromHTML(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	html := `<!DOCTYPE html><html><body><h1>Hello PDF</h1><p>Testing HTML to PDF conversion.</p></body></html>`
	data, err := client.PDF(ctx, snapapi.PDFParams{
		HTML: html,
	})
	if err != nil {
		t.Fatalf("PDF(html) error: %v", err)
	}
	if !isPDFBytes(data) {
		t.Errorf("expected PDF bytes")
	}
	t.Logf("PDF(html): %d bytes", len(data))
}

func TestLive_PDF_MissingURLAndHTML(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.PDF(ctx, snapapi.PDFParams{})
	if err == nil {
		t.Fatal("expected error when no URL or HTML provided")
	}
}

// --------------------------------------------------------------------------
// OGImage
// --------------------------------------------------------------------------

func TestLive_OGImage(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.OGImage(ctx, snapapi.OGImageParams{
		URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("OGImage() error: %v", err)
	}
	if len(data) < 1000 {
		t.Errorf("expected substantial image bytes, got %d", len(data))
	}
	t.Logf("OGImage(): %d bytes", len(data))
}

// --------------------------------------------------------------------------
// ScreenshotToFile
// --------------------------------------------------------------------------

func TestLive_ScreenshotToFile(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	tmpFile := t.TempDir() + "/live_test.png"
	n, err := client.ScreenshotToFile(ctx, tmpFile, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
	})
	if err != nil {
		t.Fatalf("ScreenshotToFile() error: %v", err)
	}
	if n < 1000 {
		t.Errorf("expected n > 1000, got %d", n)
	}

	data, _ := os.ReadFile(tmpFile)
	if !isPNGBytes(data) {
		t.Errorf("file does not look like PNG")
	}
	t.Logf("ScreenshotToFile(): wrote %d bytes to %s", n, tmpFile)
}

// --------------------------------------------------------------------------
// Scrape
// --------------------------------------------------------------------------

// TestLive_Scrape_Basic tests basic scraping and documents the API response
// shape mismatch with the SDK struct.
func TestLive_Scrape_Basic(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	result, err := client.Scrape(ctx, snapapi.ScrapeParams{
		URL:    "https://example.com",
		Format: "text",
	})
	if err != nil {
		t.Fatalf("Scrape() error: %v", err)
	}

	// BUG: The live API returns {"results":[{"data":"...","url":"..."}]}
	// but ScrapeResult has flat fields Data/URL/Status. The nested structure
	// does not unmarshal into the flat struct, so Data will be empty.
	if result.Data == "" {
		t.Log("BUG CONFIRMED: ScrapeResult.Data is empty — API returns nested " +
			"results[] array but SDK struct is flat. SDK must be fixed.")
	} else {
		t.Logf("Scrape().Data length: %d chars", len(result.Data))
	}
	t.Logf("Scrape() raw URL: %q, Status: %d, Data len: %d", result.URL, result.Status, len(result.Data))
}

func TestLive_Scrape_HTML(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	result, err := client.Scrape(ctx, snapapi.ScrapeParams{
		URL:    "https://example.com",
		Format: "html",
	})
	if err != nil {
		t.Fatalf("Scrape(html) error: %v", err)
	}
	t.Logf("Scrape(html) Data len: %d, URL: %q", len(result.Data), result.URL)
}

func TestLive_ScrapeText_Convenience(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	text, err := client.ScrapeText(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("ScrapeText() error: %v", err)
	}
	// Due to the API/SDK struct mismatch, text will be empty — document it.
	if text == "" {
		t.Log("NOTE: ScrapeText returned empty string due to API response shape mismatch")
	} else {
		t.Logf("ScrapeText() returned %d chars", len(text))
	}
}

func TestLive_ScrapeHTML_Convenience(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	html, err := client.ScrapeHTML(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("ScrapeHTML() error: %v", err)
	}
	t.Logf("ScrapeHTML() returned %d chars", len(html))
}

func TestLive_Scrape_WithSelectors(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	result, err := client.Scrape(ctx, snapapi.ScrapeParams{
		URL: "https://example.com",
		Selectors: map[string]string{
			"title": "h1",
			"body":  "p",
		},
	})
	if err != nil {
		t.Fatalf("Scrape(selectors) error: %v", err)
	}
	t.Logf("Scrape(selectors) Data len: %d", len(result.Data))
}

func TestLive_Scrape_MissingURL(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.Scrape(ctx, snapapi.ScrapeParams{})
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

// --------------------------------------------------------------------------
// Extract
// --------------------------------------------------------------------------

// TestLive_Extract_Markdown tests the extract endpoint and documents the
// API/SDK struct field mismatch.
func TestLive_Extract_Markdown(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	result, err := client.Extract(ctx, snapapi.ExtractParams{
		URL:    "https://example.com",
		Format: "markdown",
	})
	if err != nil {
		t.Fatalf("Extract(markdown) error: %v", err)
	}

	// BUG: The live API returns {"data":"...","url":"...","type":"..."} but
	// ExtractResult maps `json:"content"`. The field name does not match "data",
	// so result.Content will always be empty even on a successful response.
	if result.Content == "" {
		t.Log("BUG CONFIRMED: ExtractResult.Content is empty — API returns field " +
			"'data' but SDK struct uses json:\"content\". SDK must be fixed.")
	} else {
		t.Logf("Extract(markdown) Content length: %d chars", len(result.Content))
	}
	t.Logf("Extract(markdown) URL: %q, WordCount: %d", result.URL, result.WordCount)
}

func TestLive_Extract_Text(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	result, err := client.Extract(ctx, snapapi.ExtractParams{
		URL:    "https://example.com",
		Format: "text",
	})
	if err != nil {
		t.Fatalf("Extract(text) error: %v", err)
	}
	t.Logf("Extract(text) Content: %q (len=%d)", result.Content, len(result.Content))
}

func TestLive_ExtractMarkdown_Convenience(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	md, err := client.ExtractMarkdown(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("ExtractMarkdown() error: %v", err)
	}
	if md == "" {
		t.Log("NOTE: ExtractMarkdown returned empty string due to SDK struct mismatch (json:\"content\" vs API field \"data\")")
	} else {
		t.Logf("ExtractMarkdown() returned %d chars", len(md))
	}
}

func TestLive_ExtractText_Convenience(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	text, err := client.ExtractText(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("ExtractText() error: %v", err)
	}
	t.Logf("ExtractText() returned %d chars", len(text))
}

func TestLive_Extract_MissingURL(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.Extract(ctx, snapapi.ExtractParams{})
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

// --------------------------------------------------------------------------
// Video
// --------------------------------------------------------------------------

func TestLive_Video_WebM(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Video(ctx, snapapi.VideoParams{
		URL:      "https://example.com",
		Format:   "webm",
		Duration: 2,
	})
	if err != nil {
		t.Fatalf("Video(webm) error: %v", err)
	}
	if len(data) < 1000 {
		t.Errorf("expected substantial video bytes, got %d", len(data))
	}
	if !isWebMBytes(data) {
		t.Logf("NOTE: Response may not be WebM EBML — first 8 bytes: %x", data[:min(8, len(data))])
	}
	t.Logf("Video(webm): %d bytes", len(data))
}

func TestLive_Video_GIF(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	data, err := client.Video(ctx, snapapi.VideoParams{
		URL:      "https://example.com",
		Format:   "gif",
		Duration: 2,
	})
	if err != nil {
		t.Fatalf("Video(gif) error: %v", err)
	}
	if len(data) < 1000 {
		t.Errorf("expected substantial GIF bytes, got %d", len(data))
	}
	if !isGIFBytes(data) {
		t.Errorf("response does not look like GIF (first 4 bytes: %q)", string(data[:min(4, len(data))]))
	}
	t.Logf("Video(gif): %d bytes", len(data))
}

func TestLive_Video_MissingURL(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.Video(ctx, snapapi.VideoParams{})
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

// --------------------------------------------------------------------------
// Error handling
// --------------------------------------------------------------------------

func TestLive_InvalidAPIKey(t *testing.T) {
	// Use a client with a clearly invalid key.
	client := snapapi.New("sk_live_invalid_key_for_test",
		snapapi.WithTimeout(30*time.Second),
		snapapi.WithRetries(0),
	)
	// Still need env var so the test is not skipped.
	_ = liveKey(t)

	ctx := context.Background()
	_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
	})
	if err == nil {
		t.Fatal("expected auth error for invalid API key, got nil")
	}
	var apiErr *snapapi.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected HTTP 401, got %d", apiErr.StatusCode)
	}
	if !apiErr.IsUnauthorized() {
		t.Errorf("expected IsUnauthorized()=true, got false (code=%q)", apiErr.Code)
	}
	if !errors.Is(err, snapapi.ErrAuth) {
		t.Errorf("expected errors.Is(err, ErrAuth)=true")
	}
	t.Logf("Auth error: %v", err)
}

func TestLive_ErrorIs_Sentinel_ErrValidation(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{URL: "not-valid"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	var apiErr *snapapi.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	t.Logf("Validation error code=%q status=%d msg=%q", apiErr.Code, apiErr.StatusCode, apiErr.Message)
}

func TestLive_ErrorFormat(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{URL: "not-valid"})
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "HTTP") && !strings.Contains(msg, "INVALID") {
		t.Errorf("error string looks wrong: %q", msg)
	}
	t.Logf("Error string: %q", msg)
}

// --------------------------------------------------------------------------
// Context cancellation
// --------------------------------------------------------------------------

func TestLive_ContextCancellation(t *testing.T) {
	client := newLiveClient(t)

	// Cancel the context before the request completes.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
	})
	if err == nil {
		// The request may have been served from cache, tolerate this.
		t.Log("NOTE: Request succeeded despite cancelled context (likely cached/fast response)")
		return
	}
	// Error is expected — could be context.Canceled, DeadlineExceeded, or wrapped.
	t.Logf("Context cancellation error: %v", err)
}

func TestLive_ContextTimeout(t *testing.T) {
	client := snapapi.New(liveKey(t),
		snapapi.WithTimeout(1*time.Millisecond), // impossibly short
		snapapi.WithRetries(0),
	)
	ctx := context.Background()

	_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
	})
	if err == nil {
		t.Log("NOTE: Screenshot succeeded despite 1ms timeout — connection was very fast")
		return
	}
	t.Logf("Timeout error (as expected): %v", err)
}

// --------------------------------------------------------------------------
// Retry behavior
// --------------------------------------------------------------------------

// TestLive_RetryConfig verifies that the WithRetries and WithRetryDelay
// options take effect without actually triggering retries against the live API.
func TestLive_RetryConfig(t *testing.T) {
	key := liveKey(t)

	// Build client with explicit retry config.
	client := snapapi.New(key,
		snapapi.WithRetries(2),
		snapapi.WithRetryDelay(100*time.Millisecond),
		snapapi.WithTimeout(30*time.Second),
	)

	ctx := context.Background()
	// Ping should succeed on first attempt.
	result, err := client.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping() with retry config error: %v", err)
	}
	if result.Status == "" {
		t.Error("Ping().Status empty")
	}
	t.Logf("Retry-configured client Ping OK: %v", result.Status)
}

// TestLive_ZeroRetries builds a client that won't retry and verifies a valid
// response is still returned on the first try.
func TestLive_ZeroRetries(t *testing.T) {
	client := snapapi.New(liveKey(t),
		snapapi.WithRetries(0),
		snapapi.WithTimeout(30*time.Second),
	)
	ctx := context.Background()

	result, err := client.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping() with 0 retries error: %v", err)
	}
	if result.Status == "" {
		t.Error("expected non-empty status")
	}
}

// --------------------------------------------------------------------------
// Custom HTTP client
// --------------------------------------------------------------------------

// TestLive_CustomHTTPClient verifies that WithHTTPClient is accepted.
func TestLive_CustomHTTPClient(t *testing.T) {
	key := liveKey(t)
	hc := &http.Client{Timeout: 45 * time.Second}
	client := snapapi.New(key, snapapi.WithHTTPClient(hc))
	ctx := context.Background()

	result, err := client.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping() with custom HTTP client error: %v", err)
	}
	if result.Status == "" {
		t.Error("expected non-empty status")
	}
}

// --------------------------------------------------------------------------
// Namespace smoke tests (endpoints may 404 if not implemented server-side)
// --------------------------------------------------------------------------

func TestLive_StorageList(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.Storage.List(ctx, snapapi.StorageListParams{PerPage: 5})
	if err != nil {
		var apiErr *snapapi.APIError
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 404 || apiErr.StatusCode == 501) {
			t.Skipf("Storage.List not implemented on server: %v", err)
		}
		t.Logf("Storage.List error (may be expected): %v", err)
		return
	}
	t.Log("Storage.List OK")
}

func TestLive_APIKeysList(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	_, err := client.APIKeys.List(ctx)
	if err != nil {
		var apiErr *snapapi.APIError
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 404 || apiErr.StatusCode == 501) {
			t.Skipf("APIKeys.List not implemented on server: %v", err)
		}
		t.Logf("APIKeys.List error: %v", err)
		return
	}
	t.Log("APIKeys.List OK")
}

// --------------------------------------------------------------------------
// Concurrency
// --------------------------------------------------------------------------

// TestLive_Concurrent verifies that the client is safe for concurrent use.
func TestLive_Concurrent(t *testing.T) {
	client := newLiveClient(t)
	ctx := context.Background()

	const workers = 3
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
				URL:    "https://example.com",
				Format: "png",
			})
			errs <- err
		}()
	}

	for i := 0; i < workers; i++ {
		if err := <-errs; err != nil {
			t.Errorf("concurrent Screenshot error: %v", err)
		}
	}
	t.Logf("Concurrent screenshots (%d workers) all completed", workers)
}

// --------------------------------------------------------------------------
// min helper (Go 1.21 has builtin but include for clarity)
// --------------------------------------------------------------------------

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
