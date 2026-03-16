package snapapi_test

import (
	"context"
	"encoding/json"
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
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := snapapi.New("test-key",
		snapapi.WithBaseURL(srv.URL),
		snapapi.WithRetries(0),
		snapapi.WithTimeout(10*time.Second),
	)
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

// --- Error helpers ---

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

func isAs(err error, target interface{}) bool {
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
