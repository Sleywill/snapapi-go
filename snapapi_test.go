package snapapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	snapapi "github.com/Sleywill/snapapi-go"
)

// newTestClient returns a client pointing at the given test server.
func newTestClient(t *testing.T, srv *httptest.Server) *snapapi.Client {
	t.Helper()
	return snapapi.New("test-key",
		snapapi.WithBaseURL(srv.URL),
		snapapi.WithRetries(0), // no retries in unit tests
	)
}

// jsonHandler returns a handler that writes the given JSON body with statusCode.
func jsonHandler(statusCode int, body interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(body)
	}
}

// binaryHandler returns a handler that writes arbitrary bytes.
func binaryHandler(statusCode int, data []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write(data)
	}
}

// ─── Screenshot ───────────────────────────────────────────────────────────────

func TestScreenshot_Success(t *testing.T) {
	want := []byte("\x89PNG fake image bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/screenshot" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("unexpected Authorization header: %q", auth)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(want)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
	})
	if err != nil {
		t.Fatalf("Screenshot() error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("Screenshot() body mismatch: got %q, want %q", got, want)
	}
}

func TestScreenshot_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	var apiErr *snapapi.APIError
	if !isAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != snapapi.ErrInvalidParams {
		t.Errorf("unexpected code: %s", apiErr.Code)
	}
}

func TestScreenshot_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(401, map[string]interface{}{
		"statusCode": 401,
		"error":      "Unauthorized",
		"message":    "Invalid API key",
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	var apiErr *snapapi.APIError
	if !isAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != snapapi.ErrUnauthorized {
		t.Errorf("expected UNAUTHORIZED, got %s", apiErr.Code)
	}
	if apiErr.StatusCode != 401 {
		t.Errorf("expected StatusCode 401, got %d", apiErr.StatusCode)
	}
}

func TestScreenshot_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(429)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"statusCode": 429,
			"error":      "Rate Limited",
			"message":    "Too many requests",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	var apiErr *snapapi.APIError
	if !isAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsRateLimited() {
		t.Errorf("expected IsRateLimited() true")
	}
	if apiErr.RetryAfter != 30 {
		t.Errorf("expected RetryAfter=30, got %d", apiErr.RetryAfter)
	}
}

func TestScreenshot_ServerError(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(500, map[string]interface{}{
		"statusCode": 500,
		"error":      "Internal Server Error",
		"message":    "something broke",
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	var apiErr *snapapi.APIError
	if !isAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsServerError() {
		t.Error("expected IsServerError() true")
	}
}

// ─── Scrape ───────────────────────────────────────────────────────────────────

func TestScrape_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"success": true,
		"url":     "https://example.com",
		"text":    "Hello world",
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, err := client.Scrape(context.Background(), snapapi.ScrapeParams{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Scrape() error: %v", err)
	}
	if !result.Success {
		t.Error("expected Success=true")
	}
	if result.Text != "Hello world" {
		t.Errorf("unexpected text: %q", result.Text)
	}
}

func TestScrape_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Scrape(context.Background(), snapapi.ScrapeParams{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

// ─── Extract ─────────────────────────────────────────────────────────────────

func TestExtract_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"success":      true,
		"url":          "https://example.com",
		"format":       "markdown",
		"content":      "# Hello\n\nWorld.",
		"responseTime": 123,
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, err := client.Extract(context.Background(), snapapi.ExtractParams{
		URL:    "https://example.com",
		Format: "markdown",
	})
	if err != nil {
		t.Fatalf("Extract() error: %v", err)
	}
	if result.Content == "" {
		t.Error("expected non-empty Content")
	}
}

// ─── PDF ─────────────────────────────────────────────────────────────────────

func TestPDF_Success(t *testing.T) {
	pdfHeader := []byte("%PDF-1.4 fake")
	srv := httptest.NewServer(binaryHandler(200, pdfHeader))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.PDF(context.Background(), snapapi.PDFParams{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("PDF() error: %v", err)
	}
	if !strings.HasPrefix(string(got), "%PDF") {
		t.Errorf("expected PDF header, got: %q", got)
	}
}

// ─── Quota ───────────────────────────────────────────────────────────────────

func TestQuota_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"used":      42,
		"total":     1000,
		"remaining": 958,
		"resetAt":   "2026-04-01T00:00:00Z",
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	q, err := client.Quota(context.Background())
	if err != nil {
		t.Fatalf("Quota() error: %v", err)
	}
	if q.Used != 42 {
		t.Errorf("expected Used=42, got %d", q.Used)
	}
	if q.Total != 1000 {
		t.Errorf("expected Total=1000, got %d", q.Total)
	}
}

// ─── Retry logic ─────────────────────────────────────────────────────────────

func TestRetry_EventualSuccess(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 500,
				"error":      "Server Error",
				"message":    "temporary failure",
			})
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("fake-png"))
	}))
	defer srv.Close()

	client := snapapi.New("test-key",
		snapapi.WithBaseURL(srv.URL),
		snapapi.WithRetries(3),
		snapapi.WithRetryDelay(1*time.Millisecond),
	)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestRetry_ExhaustedRetries(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(500)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"statusCode": 500,
			"error":      "Server Error",
			"message":    "always failing",
		})
	}))
	defer srv.Close()

	client := snapapi.New("test-key",
		snapapi.WithBaseURL(srv.URL),
		snapapi.WithRetries(2),
		snapapi.WithRetryDelay(1*time.Millisecond),
	)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	// 1 initial + 2 retries = 3 calls
	if calls != 3 {
		t.Errorf("expected 3 total calls, got %d", calls)
	}
}

func TestRetry_NoRetryOn4xx(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(400)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"statusCode": 400,
			"error":      "Bad Request",
			"message":    "invalid params",
		})
	}))
	defer srv.Close()

	client := snapapi.New("test-key",
		snapapi.WithBaseURL(srv.URL),
		snapapi.WithRetries(3),
		snapapi.WithRetryDelay(1*time.Millisecond),
	)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected error")
	}
	// 4xx should NOT be retried
	if calls != 1 {
		t.Errorf("4xx should not be retried; got %d calls", calls)
	}
}

// ─── Error helpers ────────────────────────────────────────────────────────────

func isAPIError(err error, out **snapapi.APIError) bool {
	var ae *snapapi.APIError
	if ok := isAs(err, &ae); ok {
		if out != nil {
			*out = ae
		}
		return true
	}
	return false
}

// isAs wraps the standard errors.As pattern to avoid an import cycle.
func isAs(err error, target interface{}) bool {
	// Inline the check using a type assertion chain.
	type unwrapper interface{ Unwrap() error }
	for err != nil {
		if tryAs(err, target) {
			return true
		}
		u, ok := err.(unwrapper)
		if !ok {
			break
		}
		err = u.Unwrap()
	}
	return false
}

func tryAs(err error, target interface{}) bool {
	ae, ok := target.(**snapapi.APIError)
	if !ok {
		return false
	}
	v, ok := err.(*snapapi.APIError)
	if ok {
		*ae = v
	}
	return ok
}
