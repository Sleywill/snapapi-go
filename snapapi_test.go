package snapapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// assertAuthHeader checks that both X-Api-Key and Authorization headers are set.
func assertAuthHeader(t *testing.T, r *http.Request) {
	t.Helper()
	if key := r.Header.Get("X-Api-Key"); key != "test-key" {
		t.Errorf("expected X-Api-Key=test-key, got %q", key)
	}
	if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
		t.Errorf("expected Authorization=Bearer test-key, got %q", auth)
	}
}

// --- Screenshot ---

func TestScreenshot_Success(t *testing.T) {
	want := []byte("\x89PNG fake image bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/screenshot" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		assertAuthHeader(t, r)
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

func TestScreenshot_FullParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["url"] != "https://example.com" {
			t.Errorf("unexpected url: %v", body["url"])
		}
		if body["full_page"] != true {
			t.Errorf("expected full_page=true")
		}
		if body["block_ads"] != true {
			t.Errorf("expected block_ads=true")
		}
		if body["custom_css"] != "body { background: red; }" {
			t.Errorf("unexpected custom_css: %v", body["custom_css"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{
		URL:       "https://example.com",
		Format:    "png",
		FullPage:  true,
		Width:     1920,
		Height:    1080,
		BlockAds:  true,
		CustomCSS: "body { background: red; }",
		Scale:     2.0,
		Headers:   map[string]string{"X-Custom": "value"},
	})
	if err != nil {
		t.Fatalf("Screenshot() error: %v", err)
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

// --- ScreenshotToFile ---

func TestScreenshotToFile_Success(t *testing.T) {
	want := []byte("\x89PNG fake image bytes")
	srv := httptest.NewServer(binaryHandler(200, want))
	defer srv.Close()

	client := newTestClient(t, srv)
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")

	n, err := client.ScreenshotToFile(context.Background(), path, snapapi.ScreenshotParams{
		URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("ScreenshotToFile() error: %v", err)
	}
	if n != len(want) {
		t.Errorf("expected %d bytes written, got %d", len(want), n)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(data) != string(want) {
		t.Errorf("file content mismatch")
	}
}

// --- Scrape ---

func TestScrape_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"data":   "<html>Hello</html>",
		"url":    "https://example.com",
		"status": 200,
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, err := client.Scrape(context.Background(), snapapi.ScrapeParams{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Scrape() error: %v", err)
	}
	if result.Data != "<html>Hello</html>" {
		t.Errorf("unexpected data: %q", result.Data)
	}
	if result.Status != 200 {
		t.Errorf("expected status=200, got %d", result.Status)
	}
}

func TestScrape_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Scrape(context.Background(), snapapi.ScrapeParams{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

// --- Extract ---

func TestExtract_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"content":    "# Hello\n\nWorld.",
		"url":        "https://example.com",
		"word_count": 2,
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
	if result.WordCount != 2 {
		t.Errorf("expected word_count=2, got %d", result.WordCount)
	}
}

// --- Analyze ---

func TestAnalyze_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"result": "This is a test page.",
		"url":    "https://example.com",
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, err := client.Analyze(context.Background(), snapapi.AnalyzeParams{
		URL:      "https://example.com",
		Prompt:   "Summarize this page.",
		Provider: "openai",
		APIKey:   "sk-test",
	})
	if err != nil {
		t.Fatalf("Analyze() error: %v", err)
	}
	if result.Result != "This is a test page." {
		t.Errorf("unexpected result: %q", result.Result)
	}
}

func TestAnalyze_ServiceUnavailable(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(503, map[string]interface{}{
		"statusCode": 503,
		"error":      "Service Unavailable",
		"message":    "LLM credits exhausted",
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Analyze(context.Background(), snapapi.AnalyzeParams{URL: "https://example.com"})
	var apiErr *snapapi.APIError
	if !isAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if !apiErr.IsServiceUnavailable() {
		t.Error("expected IsServiceUnavailable() true")
	}
}

func TestAnalyze_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Analyze(context.Background(), snapapi.AnalyzeParams{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

// --- PDF ---

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

// --- Usage ---

func TestGetUsage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/usage" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		assertAuthHeader(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"used":      42,
			"total":     1000,
			"remaining": 958,
			"resetAt":   "2026-04-01T00:00:00Z",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	u, err := client.GetUsage(context.Background())
	if err != nil {
		t.Fatalf("GetUsage() error: %v", err)
	}
	if u.Used != 42 {
		t.Errorf("expected Used=42, got %d", u.Used)
	}
	if u.Total != 1000 {
		t.Errorf("expected Total=1000, got %d", u.Total)
	}
}

// --- Retry logic ---

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

// --- Context cancellation ---

func TestContextCancellation(t *testing.T) {
	// The server uses a short sleep so the test completes quickly while
	// still outlasting the client-side deadline.  We avoid blocking on any
	// channel so that httptest.Server.Close() can drain the connection.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := snapapi.New("test-key",
		snapapi.WithBaseURL(srv.URL),
		snapapi.WithRetries(0),
		snapapi.WithTimeout(10*time.Second),
	)
	// Context times out after 50ms; server sleeps 200ms — the client sees a
	// deadline exceeded error before the server ever responds.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Screenshot(ctx, snapapi.ScreenshotParams{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected context deadline error")
	}
}

// --- OG Image ---

func TestOGImage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/screenshot" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["width"] != float64(1200) {
			t.Errorf("expected width=1200, got %v", body["width"])
		}
		if body["height"] != float64(630) {
			t.Errorf("expected height=630, got %v", body["height"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("fake-og-image"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.OGImage(context.Background(), snapapi.OGImageParams{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("OGImage() error: %v", err)
	}
	if string(got) != "fake-og-image" {
		t.Errorf("unexpected body: %q", got)
	}
}

func TestOGImage_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.OGImage(context.Background(), snapapi.OGImageParams{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

// --- Ping ---

func TestPing_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/ping" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		assertAuthHeader(t, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "ok",
			"timestamp": 1710540000000,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, err := client.Ping(context.Background())
	if err != nil {
		t.Fatalf("Ping() error: %v", err)
	}
	if result.Status != "ok" {
		t.Errorf("expected status=ok, got %q", result.Status)
	}
}

// --- PDF with HTML ---

func TestPDF_WithHTML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["html"] != "<h1>Invoice</h1>" {
			t.Errorf("unexpected html: %v", body["html"])
		}
		if body["format"] != "pdf" {
			t.Errorf("expected format=pdf, got %v", body["format"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("%PDF-1.4 html"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.PDF(context.Background(), snapapi.PDFParams{HTML: "<h1>Invoice</h1>"})
	if err != nil {
		t.Fatalf("PDF() error: %v", err)
	}
	if !strings.HasPrefix(string(got), "%PDF") {
		t.Errorf("expected PDF header, got: %q", got)
	}
}

func TestPDF_MissingURLAndHTML(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.PDF(context.Background(), snapapi.PDFParams{})
	if err == nil {
		t.Fatal("expected error for missing URL and HTML")
	}
}

// --- Quota alias ---

func TestQuota_IsAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"used": 5, "total": 100, "remaining": 95,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	u, err := client.Quota(context.Background())
	if err != nil {
		t.Fatalf("Quota() error: %v", err)
	}
	if u.Used != 5 {
		t.Errorf("expected Used=5, got %d", u.Used)
	}
}

// --- Video ---

func TestVideo_Success(t *testing.T) {
	srv := httptest.NewServer(binaryHandler(200, []byte("fake-video-data")))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.Video(context.Background(), snapapi.VideoParams{URL: "https://example.com", Duration: 5})
	if err != nil {
		t.Fatalf("Video() error: %v", err)
	}
	if string(got) != "fake-video-data" {
		t.Errorf("unexpected body: %q", got)
	}
}

func TestVideo_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Video(context.Background(), snapapi.VideoParams{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

// --- ScrapeText / ScrapeHTML convenience methods ---

func TestScrapeText_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"data":   "Hello world",
		"url":    "https://example.com",
		"status": 200,
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	text, err := client.ScrapeText(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("ScrapeText() error: %v", err)
	}
	if text != "Hello world" {
		t.Errorf("unexpected text: %q", text)
	}
}

func TestScrapeHTML_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"data":   "<html><body>Hello</body></html>",
		"url":    "https://example.com",
		"status": 200,
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	html, err := client.ScrapeHTML(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("ScrapeHTML() error: %v", err)
	}
	if !strings.Contains(html, "<html>") {
		t.Errorf("unexpected html: %q", html)
	}
}

// --- ExtractMarkdown / ExtractText convenience methods ---

func TestExtractMarkdown_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"content":    "# Title\n\nBody.",
		"url":        "https://example.com",
		"word_count": 2,
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	md, err := client.ExtractMarkdown(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("ExtractMarkdown() error: %v", err)
	}
	if !strings.HasPrefix(md, "# Title") {
		t.Errorf("unexpected markdown: %q", md)
	}
}

func TestExtractText_Success(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(200, map[string]interface{}{
		"content":    "Title Body.",
		"url":        "https://example.com",
		"word_count": 2,
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	text, err := client.ExtractText(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("ExtractText() error: %v", err)
	}
	if text != "Title Body." {
		t.Errorf("unexpected text: %q", text)
	}
}

// --- ScreenshotToStorage ---

func TestScreenshotToStorage_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/screenshot/storage" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"url":          "https://storage.snapapi.pics/reports/home.png",
			"key":          "reports/home.png",
			"bucket":       "snapapi-captures",
			"size":         45678,
			"content_type": "image/png",
			"created_at":   "2026-03-17T10:00:00Z",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	capture, err := client.ScreenshotToStorage(context.Background(), snapapi.ScreenshotToStorageParams{
		ScreenshotParams: snapapi.ScreenshotParams{
			URL:    "https://example.com",
			Format: "png",
		},
		StorageKey: "reports/home.png",
	})
	if err != nil {
		t.Fatalf("ScreenshotToStorage() error: %v", err)
	}
	if capture.Key != "reports/home.png" {
		t.Errorf("unexpected key: %q", capture.Key)
	}
	if capture.Size != 45678 {
		t.Errorf("unexpected size: %d", capture.Size)
	}
}

func TestScreenshotToStorage_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.ScreenshotToStorage(context.Background(), snapapi.ScreenshotToStorageParams{})
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

// --- Storage namespace ---

func TestStorage_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/storage") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{"key": "home.png", "url": "https://cdn.example.com/home.png", "size": 1234, "content_type": "image/png", "created_at": "2026-03-17T10:00:00Z"},
			},
			"total":    1,
			"page":     1,
			"per_page": 20,
			"has_more": false,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	result, err := client.Storage.List(context.Background(), snapapi.StorageListParams{})
	if err != nil {
		t.Fatalf("Storage.List() error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Key != "home.png" {
		t.Errorf("unexpected key: %q", result.Items[0].Key)
	}
}

func TestStorage_Get(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/storage/home.png" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"key": "home.png", "url": "https://cdn.example.com/home.png",
			"size": 1234, "content_type": "image/png", "created_at": "2026-03-17T10:00:00Z",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	item, err := client.Storage.Get(context.Background(), "home.png")
	if err != nil {
		t.Fatalf("Storage.Get() error: %v", err)
	}
	if item.Key != "home.png" {
		t.Errorf("unexpected key: %q", item.Key)
	}
}

func TestStorage_Get_MissingKey(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Storage.Get(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestStorage_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/storage/home.png" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	if err := client.Storage.Delete(context.Background(), "home.png"); err != nil {
		t.Fatalf("Storage.Delete() error: %v", err)
	}
}

// --- Scheduled namespace ---

func TestScheduled_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/scheduled" {
			t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["url"] != "https://example.com" {
			t.Errorf("unexpected url: %v", body["url"])
		}
		if body["cron"] != "0 9 * * 1-5" {
			t.Errorf("unexpected cron: %v", body["cron"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "sched_abc123", "url": "https://example.com",
			"cron": "0 9 * * 1-5", "active": true, "created_at": "2026-03-17T10:00:00Z",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	sched, err := client.Scheduled.Create(context.Background(), snapapi.CreateScheduleParams{
		URL:  "https://example.com",
		Cron: "0 9 * * 1-5",
	})
	if err != nil {
		t.Fatalf("Scheduled.Create() error: %v", err)
	}
	if sched.ID != "sched_abc123" {
		t.Errorf("unexpected ID: %q", sched.ID)
	}
}

func TestScheduled_Create_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Scheduled.Create(context.Background(), snapapi.CreateScheduleParams{Cron: "* * * * *"})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestScheduled_Create_MissingCron(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Scheduled.Create(context.Background(), snapapi.CreateScheduleParams{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected error for missing cron")
	}
}

func TestScheduled_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	if err := client.Scheduled.Delete(context.Background(), "sched_abc123"); err != nil {
		t.Fatalf("Scheduled.Delete() error: %v", err)
	}
}

func TestScheduled_Pause_Resume(t *testing.T) {
	var receivedPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPaths = append(receivedPaths, r.URL.Path)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_ = client.Scheduled.Pause(context.Background(), "sched_abc123")
	_ = client.Scheduled.Resume(context.Background(), "sched_abc123")

	if len(receivedPaths) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(receivedPaths))
	}
	if receivedPaths[0] != "/v1/scheduled/sched_abc123/pause" {
		t.Errorf("unexpected pause path: %s", receivedPaths[0])
	}
	if receivedPaths[1] != "/v1/scheduled/sched_abc123/resume" {
		t.Errorf("unexpected resume path: %s", receivedPaths[1])
	}
}

// --- Webhooks namespace ---

func TestWebhooks_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/webhooks" {
			t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["url"] != "https://myapp.com/hooks/snapapi" {
			t.Errorf("unexpected url: %v", body["url"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "wh_abc123", "url": "https://myapp.com/hooks/snapapi",
			"events": []string{"screenshot.completed"}, "active": true,
			"created_at": "2026-03-17T10:00:00Z",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	wh, err := client.Webhooks.Create(context.Background(), snapapi.CreateWebhookParams{
		URL:    "https://myapp.com/hooks/snapapi",
		Events: []string{"screenshot.completed"},
	})
	if err != nil {
		t.Fatalf("Webhooks.Create() error: %v", err)
	}
	if wh.ID != "wh_abc123" {
		t.Errorf("unexpected ID: %q", wh.ID)
	}
}

func TestWebhooks_Create_MissingURL(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.Webhooks.Create(context.Background(), snapapi.CreateWebhookParams{})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestWebhooks_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/webhooks" {
			t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "wh_1", "url": "https://myapp.com/hooks/1", "events": []string{"screenshot.completed"}, "active": true, "created_at": "2026-03-17T10:00:00Z"},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	hooks, err := client.Webhooks.List(context.Background())
	if err != nil {
		t.Fatalf("Webhooks.List() error: %v", err)
	}
	if len(hooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(hooks))
	}
	if hooks[0].ID != "wh_1" {
		t.Errorf("unexpected ID: %q", hooks[0].ID)
	}
}

func TestWebhooks_Delete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/webhooks/wh_abc123" {
			t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	if err := client.Webhooks.Delete(context.Background(), "wh_abc123"); err != nil {
		t.Fatalf("Webhooks.Delete() error: %v", err)
	}
}

// --- APIKeys namespace ---

func TestAPIKeys_Create(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/api-keys" {
			t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "CI pipeline" {
			t.Errorf("unexpected name: %v", body["name"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "key_abc123", "name": "CI pipeline",
			"key": "sk_live_newkey", "created_at": "2026-03-17T10:00:00Z",
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	key, err := client.APIKeys.Create(context.Background(), snapapi.CreateAPIKeyParams{Name: "CI pipeline"})
	if err != nil {
		t.Fatalf("APIKeys.Create() error: %v", err)
	}
	if key.Key != "sk_live_newkey" {
		t.Errorf("unexpected key: %q", key.Key)
	}
}

func TestAPIKeys_Create_MissingName(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	_, err := client.APIKeys.Create(context.Background(), snapapi.CreateAPIKeyParams{})
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}

func TestAPIKeys_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/api-keys" {
			t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": "key_1", "name": "prod", "created_at": "2026-03-17T10:00:00Z"},
			{"id": "key_2", "name": "staging", "created_at": "2026-03-17T10:00:00Z"},
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	keys, err := client.APIKeys.List(context.Background())
	if err != nil {
		t.Fatalf("APIKeys.List() error: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestAPIKeys_Revoke(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/api-keys/key_abc123" {
			t.Errorf("unexpected method/path: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	if err := client.APIKeys.Revoke(context.Background(), "key_abc123"); err != nil {
		t.Fatalf("APIKeys.Revoke() error: %v", err)
	}
}

func TestAPIKeys_Revoke_MissingID(t *testing.T) {
	client := snapapi.New("test-key", snapapi.WithRetries(0))
	err := client.APIKeys.Revoke(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}

// --- Error code mapping ---

func TestErrorCode_QuotaExceeded(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(402, map[string]interface{}{
		"statusCode": 402,
		"error":      "Payment Required",
		"message":    "Monthly quota exceeded",
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	var apiErr *snapapi.APIError
	if !isAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != snapapi.ErrQuotaExceeded {
		t.Errorf("expected QUOTA_EXCEEDED, got %q", apiErr.Code)
	}
	if !apiErr.IsQuotaExceeded() {
		t.Error("IsQuotaExceeded() should be true")
	}
}

func TestErrorCode_NotFound(t *testing.T) {
	srv := httptest.NewServer(jsonHandler(404, map[string]interface{}{
		"statusCode": 404,
		"error":      "Not Found",
		"message":    "Resource not found",
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	var apiErr *snapapi.APIError
	if !isAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Code != snapapi.ErrNotFound {
		t.Errorf("expected NOT_FOUND, got %q", apiErr.Code)
	}
}

// --- Retry: Retry-After overrides backoff ---

func TestRetry_RetryAfterOverridesBackoff(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			// Send Retry-After of 0 seconds so the test stays fast.
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(429)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"statusCode": 429, "error": "Rate Limited", "message": "slow down",
			})
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	client := snapapi.New("test-key",
		snapapi.WithBaseURL(srv.URL),
		snapapi.WithRetries(1),
		snapapi.WithRetryDelay(10*time.Second), // would make the test very slow if used
	)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

// --- Namespace: nil guard / thread safety ---

func TestClient_NamespacesInitialized(t *testing.T) {
	client := snapapi.New("test-key")
	if client.Storage == nil {
		t.Error("Storage namespace is nil")
	}
	if client.Scheduled == nil {
		t.Error("Scheduled namespace is nil")
	}
	if client.Webhooks == nil {
		t.Error("Webhooks namespace is nil")
	}
	if client.APIKeys == nil {
		t.Error("APIKeys namespace is nil")
	}
}

// --- New fields: DarkMode, BlockCookies, ScrollVideo ---

func TestScreenshot_DarkModeAndBlockCookies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["dark_mode"] != true {
			t.Errorf("expected dark_mode=true, got %v", body["dark_mode"])
		}
		if body["block_cookies"] != true {
			t.Errorf("expected block_cookies=true, got %v", body["block_cookies"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Screenshot(context.Background(), snapapi.ScreenshotParams{
		URL:          "https://example.com",
		DarkMode:     true,
		BlockCookies: true,
	})
	if err != nil {
		t.Fatalf("Screenshot() error: %v", err)
	}
}

func TestVideo_ScrollVideo(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["scrollVideo"] != true {
			t.Errorf("expected scrollVideo=true, got %v", body["scrollVideo"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("fake-scroll-video"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.Video(context.Background(), snapapi.VideoParams{
		URL:         "https://example.com",
		ScrollVideo: true,
	})
	if err != nil {
		t.Fatalf("Video() error: %v", err)
	}
	if string(got) != "fake-scroll-video" {
		t.Errorf("unexpected body: %q", got)
	}
}

func TestScrape_SelectorsAndWaitFor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		selectors, ok := body["selectors"].(map[string]interface{})
		if !ok {
			t.Errorf("expected selectors map, got %T", body["selectors"])
		}
		if selectors["title"] != "h1" {
			t.Errorf("expected selectors.title=h1, got %v", selectors["title"])
		}
		if body["waitFor"] != ".loaded" {
			t.Errorf("expected waitFor=.loaded, got %v", body["waitFor"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": "test", "url": "https://example.com", "status": 200,
		})
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	_, err := client.Scrape(context.Background(), snapapi.ScrapeParams{
		URL:       "https://example.com",
		Selectors: map[string]string{"title": "h1"},
		WaitFor:   ".loaded",
	})
	if err != nil {
		t.Fatalf("Scrape() error: %v", err)
	}
}

// --- Method aliases ---

func TestGeneratePDF_IsAlias(t *testing.T) {
	pdfHeader := []byte("%PDF-1.4 alias")
	srv := httptest.NewServer(binaryHandler(200, pdfHeader))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.GeneratePDF(context.Background(), snapapi.PDFParams{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("GeneratePDF() error: %v", err)
	}
	if !strings.HasPrefix(string(got), "%PDF") {
		t.Errorf("expected PDF header, got: %q", got)
	}
}

func TestGenerateOGImage_IsAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["width"] != float64(1200) {
			t.Errorf("expected width=1200, got %v", body["width"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("og-image"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv)
	got, err := client.GenerateOGImage(context.Background(), snapapi.OGImageParams{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("GenerateOGImage() error: %v", err)
	}
	if string(got) != "og-image" {
		t.Errorf("unexpected body: %q", got)
	}
}

// --- Error helpers ---

// isAPIError uses errors.As to check whether err is (or wraps) *snapapi.APIError.
// When it returns true, *out is set to the unwrapped *APIError.
func isAPIError(err error, out **snapapi.APIError) bool {
	var ae *snapapi.APIError
	if errors.As(err, &ae) {
		if out != nil {
			*out = ae
		}
		return true
	}
	return false
}
