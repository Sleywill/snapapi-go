# SnapAPI Go SDK

Official Go SDK for [SnapAPI.pics](https://snapapi.pics) — capture screenshots, generate PDFs, scrape pages, extract structured content, and analyze web pages with LLMs.

[![Go Reference](https://pkg.go.dev/badge/github.com/Sleywill/snapapi-go.svg)](https://pkg.go.dev/github.com/Sleywill/snapapi-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Sleywill/snapapi-go)](https://goreportcard.com/report/github.com/Sleywill/snapapi-go)
[![CI](https://github.com/Sleywill/snapapi-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Sleywill/snapapi-go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go 1.21+](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go&logoColor=white)](https://go.dev/)

## Features

- **Screenshot** -- Capture full-page or viewport screenshots in PNG, JPEG, WebP
- **PDF** -- Generate PDFs from any URL with configurable paper size and margins
- **Scrape** -- Extract HTML, text, or JSON from any web page
- **Extract** -- Get clean, LLM-ready content in Markdown, text, or JSON
- **Analyze** -- Send page content to OpenAI, Anthropic, or Google LLMs for analysis
- **Video** -- Record short browser session videos
- **Usage** -- Monitor your API quota in real time
- **Retry with backoff** -- Automatic exponential backoff on 5xx and rate-limit errors
- **Context support** -- Full `context.Context` integration for cancellation and timeouts

## Requirements

- Go 1.21+
- A SnapAPI API key -- get one free at [snapapi.pics](https://snapapi.pics)

## Installation

```bash
go get github.com/Sleywill/snapapi-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "os"
    "time"

    snapapi "github.com/Sleywill/snapapi-go"
)

func main() {
    client := snapapi.New(os.Getenv("SNAPAPI_KEY"),
        snapapi.WithTimeout(30*time.Second),
        snapapi.WithRetries(3),
    )
    ctx := context.Background()

    // Capture a screenshot and save to file
    img, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
        URL:      "https://example.com",
        Format:   "png",
        FullPage: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    os.WriteFile("screenshot.png", img, 0644)
}
```

## Configuration

```go
client := snapapi.New("sk_live_...",
    // HTTP request timeout (default: 30s)
    snapapi.WithTimeout(45*time.Second),
    // Number of automatic retries on 5xx / rate-limit errors (default: 3)
    snapapi.WithRetries(3),
    // Base delay for exponential back-off (default: 500ms)
    snapapi.WithRetryDelay(500*time.Millisecond),
    // Override the API base URL (useful for testing)
    snapapi.WithBaseURL("https://api.snapapi.pics"),
    // Bring your own *http.Client (e.g. with a custom transport)
    snapapi.WithHTTPClient(myHTTPClient),
)
```

## Complete API Reference

### Screenshot -- `POST /v1/screenshot`

Capture a screenshot of any URL. Returns raw image bytes.

```go
img, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
    URL:             "https://example.com",
    Format:          "png",       // "png", "jpeg", "webp", or "pdf"
    Width:           1280,        // viewport width in pixels
    Height:          720,         // viewport height in pixels
    FullPage:        true,        // capture entire scrollable page
    Delay:           500,         // ms to wait after page load
    Quality:         85,          // JPEG/WebP quality (1-100)
    Scale:           2.0,         // device scale factor (retina)
    BlockAds:        true,        // enable ad blocking
    BlockCookies:    true,        // block cookie consent banners
    DarkMode:        true,        // enable prefers-color-scheme: dark
    WaitForSelector: ".loaded",   // wait for CSS selector to appear
    Selector:        "#hero",     // capture only this element
    ScrollY:         500,         // scroll down before capturing
    CustomCSS:       "body { background: white; }",
    CustomJS:        "document.querySelector('.popup')?.remove();",
    UserAgent:       "MyBot/1.0",
    Proxy:           "http://proxy:8080",
    Headers:         map[string]string{"Cookie": "session=abc"},
    Clip:            &snapapi.ClipRegion{X: 0, Y: 0, Width: 800, Height: 600},
})
os.WriteFile("screenshot.png", img, 0644)
```

### ScreenshotToFile

Convenience method that captures and writes directly to disk:

```go
n, err := client.ScreenshotToFile(ctx, "output.png", snapapi.ScreenshotParams{
    URL:      "https://example.com",
    Format:   "png",
    FullPage: true,
})
fmt.Printf("Wrote %d bytes\n", n)
```

### Scrape -- `POST /v1/scrape`

Fetch HTML, text, or structured data from a URL:

```go
result, err := client.Scrape(ctx, snapapi.ScrapeParams{
    URL:             "https://example.com",
    Selector:        "article",        // scope to CSS selector
    Selectors:       map[string]string{"title": "h1", "body": "article"}, // named multi-element
    Format:          "html",           // "html", "text", or "json"
    WaitFor:         ".content-ready", // wait for selector/timeout before scraping
    WaitForSelector: ".content-ready", // wait for dynamic content
    Headers:         map[string]string{"Accept-Language": "en-US"},
    Proxy:           "http://proxy:8080",
})
fmt.Println(result.Data)     // scraped content
fmt.Println(result.URL)      // final URL after redirects
fmt.Println(result.Status)   // HTTP status of the scraped page
```

### Extract -- `POST /v1/extract`

Extract clean, readable content optimized for LLM consumption:

```go
result, err := client.Extract(ctx, snapapi.ExtractParams{
    URL:             "https://example.com/blog/post",
    Format:          "markdown",    // "markdown", "text", or "json"
    IncludeLinks:    boolPtr(true), // include hyperlinks (default: true)
    IncludeImages:   boolPtr(false),// include image refs (default: false)
    Selector:        "main",        // scope extraction
    WaitForSelector: ".loaded",
    Headers:         map[string]string{"Authorization": "Bearer token"},
})
fmt.Println(result.Content)    // clean markdown/text
fmt.Println(result.WordCount)  // approximate word count
fmt.Println(result.URL)        // final URL

// Helper for bool pointers
func boolPtr(b bool) *bool { return &b }
```

### Analyze -- `POST /v1/analyze`

Send a page to an LLM for analysis:

```go
result, err := client.Analyze(ctx, snapapi.AnalyzeParams{
    URL:      "https://example.com",
    Prompt:   "Summarize this page in 3 bullet points.",
    Provider: "openai",       // "openai", "anthropic", or "google"
    APIKey:   "sk-...",       // your LLM provider API key
    JSONSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "summary": map[string]string{"type": "string"},
        },
    },
})
fmt.Println(result.Result)
```

> **Note:** The analyze endpoint may return HTTP 503 if LLM credits are exhausted.
> Check `apiErr.IsServiceUnavailable()` to handle this gracefully.

### PDF -- `POST /v1/screenshot` (format=pdf)

Generate a PDF from any URL or raw HTML:

```go
pdfBytes, err := client.PDF(ctx, snapapi.PDFParams{
    URL:          "https://example.com",
    PageSize:     "a4",   // "a4" or "letter"
    MarginTop:    "10mm",
    MarginBottom: "10mm",
    MarginLeft:   "10mm",
    MarginRight:  "10mm",
    Landscape:    false,
})
os.WriteFile("output.pdf", pdfBytes, 0644)

// Or generate from raw HTML:
pdfBytes, err = client.PDF(ctx, snapapi.PDFParams{
    HTML: "<h1>Invoice #1234</h1><p>Total: $99</p>",
})

// Convenience method to write directly to disk:
n, err := client.PDFToFile(ctx, "invoice.pdf", snapapi.PDFParams{URL: "https://example.com/invoice/1"})
fmt.Printf("Wrote %d bytes\n", n)
```

### Video -- `POST /v1/video`

Record a short browser session video:

```go
videoBytes, err := client.Video(ctx, snapapi.VideoParams{
    URL:         "https://example.com",
    Duration:    5,       // seconds
    Format:      "mp4",   // "webm", "mp4", or "gif"
    Width:       1280,
    Height:      720,
    ScrollVideo: true,    // scroll-based recording
})
os.WriteFile("capture.mp4", videoBytes, 0644)
```

### GetUsage -- `GET /v1/usage`

Check your current API usage:

```go
usage, err := client.GetUsage(ctx)
fmt.Printf("Used: %d / %d (%d remaining, resets %s)\n",
    usage.Used, usage.Total, usage.Remaining, usage.ResetAt)
```

### OGImage

Convenience wrapper that captures a screenshot at OG-standard dimensions:

```go
og, err := client.OGImage(ctx, snapapi.OGImageParams{
    URL:    "https://example.com/blog/my-post",
    Format: "png",
    Width:  1200, // default
    Height: 630,  // default
})
os.WriteFile("og.png", og, 0644)
```

### Ping -- `GET /v1/ping`

Health check:

```go
result, err := client.Ping(ctx)
fmt.Println(result.Status) // "ok"
```

### Convenience methods

```go
// Scrape shortcuts
text, err := client.ScrapeText(ctx, "https://example.com")
html, err := client.ScrapeHTML(ctx, "https://example.com")

// Extract shortcuts
md,   err := client.ExtractMarkdown(ctx, "https://example.com/blog/post")
text, err  = client.ExtractText(ctx, "https://example.com/blog/post")
```

### ScreenshotToStorage

Capture and store directly in SnapAPI-managed cloud storage:

```go
capture, err := client.ScreenshotToStorage(ctx, snapapi.ScreenshotToStorageParams{
    ScreenshotParams: snapapi.ScreenshotParams{
        URL:    "https://example.com",
        Format: "png",
    },
    StorageKey: "reports/2026-03-17/home.png",
})
fmt.Println(capture.URL)  // public CDN URL
fmt.Println(capture.Size) // bytes
```

## Namespaces

The client exposes four sub-namespaces for managing account resources:

### Storage -- `client.Storage`

```go
// List stored captures
items, err := client.Storage.List(ctx, snapapi.StorageListParams{PerPage: 50})

// Get metadata for one capture
item, err := client.Storage.Get(ctx, "reports/home.png")
fmt.Println(item.URL)

// Delete a stored capture
err = client.Storage.Delete(ctx, "reports/home.png")
```

### Scheduled -- `client.Scheduled`

```go
// Create a recurring schedule (cron syntax)
sched, err := client.Scheduled.Create(ctx, snapapi.CreateScheduleParams{
    URL:  "https://example.com",
    Cron: "0 9 * * 1-5", // weekdays at 09:00 UTC
})
fmt.Println(sched.ID, sched.NextRunAt)

// List / Get / Delete / Pause / Resume
schedules, err := client.Scheduled.List(ctx)
err = client.Scheduled.Pause(ctx, sched.ID)
err = client.Scheduled.Resume(ctx, sched.ID)
err = client.Scheduled.Delete(ctx, sched.ID)
```

### Webhooks -- `client.Webhooks`

```go
// Register a webhook
wh, err := client.Webhooks.Create(ctx, snapapi.CreateWebhookParams{
    URL:    "https://myapp.com/hooks/snapapi",
    Events: []string{"screenshot.completed", "schedule.run.failed"},
    Secret: "my-signing-secret", // used to verify HMAC signature
})
fmt.Println(wh.ID)

// List / Get / Delete
hooks, err := client.Webhooks.List(ctx)
err = client.Webhooks.Delete(ctx, wh.ID)
```

### APIKeys -- `client.APIKeys`

```go
// Create a new key (raw key only returned once)
key, err := client.APIKeys.Create(ctx, snapapi.CreateAPIKeyParams{
    Name: "CI pipeline",
})
fmt.Println(key.Key) // store securely -- not shown again

// List keys (metadata only, no raw key values)
keys, err := client.APIKeys.List(ctx)

// Revoke (permanently delete) a key
err = client.APIKeys.Revoke(ctx, key.ID)
```

## Error Handling

Every method returns `(result, error)` and never panics. API errors are typed as `*APIError`.

### Using errors.Is() with sentinel errors

The SDK provides sentinel errors for concise error matching:

```go
import "errors"

img, err := client.Screenshot(ctx, params)
if errors.Is(err, snapapi.ErrRateLimit) {
    log.Println("Rate limited, backing off...")
} else if errors.Is(err, snapapi.ErrAuth) {
    log.Fatal("Invalid API key")
} else if errors.Is(err, snapapi.ErrQuota) {
    log.Fatal("Monthly quota exhausted")
} else if errors.Is(err, snapapi.ErrServer) {
    log.Println("Server error, will retry")
} else if errors.Is(err, snapapi.ErrNetwork) {
    log.Println("Network failure")
} else if errors.Is(err, snapapi.ErrValidation) {
    log.Println("Bad request parameters")
}
```

| Sentinel | Matches |
|---|---|
| `ErrRateLimit` | HTTP 429 |
| `ErrAuth` | HTTP 401, 403 |
| `ErrQuota` | HTTP 402 |
| `ErrValidation` | HTTP 400 |
| `ErrServer` | HTTP 5xx |
| `ErrNetwork` | Connection failures |

### Using errors.As() for full detail

For access to the full error payload, use `errors.As`:

```go
import "errors"

img, err := client.Screenshot(ctx, params)
if err != nil {
    var apiErr *snapapi.APIError
    if errors.As(err, &apiErr) {
        switch {
        case apiErr.IsRateLimited():
            fmt.Printf("Rate limited. Retry after %ds\n", apiErr.RetryAfter)
        case apiErr.IsUnauthorized():
            log.Fatal("Invalid API key")
        case apiErr.IsQuotaExceeded():
            log.Fatal("Monthly quota exhausted. Upgrade your plan.")
        case apiErr.IsServerError():
            log.Printf("Server error: %s", apiErr.Message)
        case apiErr.IsServiceUnavailable():
            log.Printf("Service down: %s", apiErr.Message)
        default:
            log.Printf("[%s] %s (HTTP %d)", apiErr.Code, apiErr.Message, apiErr.StatusCode)
        }
    }
    return
}
```

### Error Codes

| Constant | HTTP Status | Description |
|---|---|---|
| `ErrInvalidParams` | 400 | Bad request parameters |
| `ErrUnauthorized` | 401 | Invalid or missing API key |
| `ErrQuotaExceeded` | 402 | Monthly quota exhausted |
| `ErrForbidden` | 403 | Insufficient permissions |
| `ErrNotFound` | 404 | Resource not found |
| `ErrRateLimited` | 429 | Rate limit exceeded (check `RetryAfter`) |
| `ErrTimeout` | -- | Request timed out server-side |
| `ErrCaptureFailed` | -- | Browser capture failed |
| `ErrConnectionError` | -- | Network-level failure |
| `ErrServerError` | 5xx | Unexpected server error |
| `ErrServiceDown` | 503 | Service temporarily unavailable |

### APIError Methods

| Method | Description |
|---|---|
| `IsRateLimited()` | Returns true if the error is HTTP 429 |
| `IsUnauthorized()` | Returns true if the error is HTTP 401/403 |
| `IsQuotaExceeded()` | Returns true if the error is HTTP 402 |
| `IsServerError()` | Returns true for any 5xx status |
| `IsServiceUnavailable()` | Returns true for HTTP 503 |

## Retry Behavior

The SDK automatically retries on transient failures:

- **Retried:** 5xx errors, 429 rate limits, network timeouts
- **Not retried:** 4xx client errors (400, 401, 403)
- **Backoff:** Exponential with configurable base delay (default 500ms)
- **Retry-After:** Honored when the server provides this header

Disable retries:

```go
client := snapapi.New("sk_...", snapapi.WithRetries(0))
```

## Context and Cancellation

All methods accept `context.Context` as the first parameter for timeouts and cancellation:

```go
// Timeout after 10 seconds
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
img, err := client.Screenshot(ctx, params)
```

## Real-World Use Cases

### Website Monitoring

```go
// Capture screenshots of your sites on a schedule
urls := []string{"https://mysite.com", "https://mysite.com/pricing"}
for _, u := range urls {
    filename := fmt.Sprintf("monitor_%d.png", time.Now().Unix())
    _, err := client.ScreenshotToFile(ctx, filename, snapapi.ScreenshotParams{
        URL:      u,
        Format:   "png",
        FullPage: true,
    })
    if err != nil {
        log.Printf("Failed to capture %s: %v", u, err)
    }
}
```

### SEO Audit Tool

```go
// Extract content and analyze for SEO quality
content, _ := client.Extract(ctx, snapapi.ExtractParams{
    URL:    "https://example.com",
    Format: "text",
})
fmt.Printf("Page has %d words\n", content.WordCount)

// Use the analyze endpoint for deeper insights
analysis, _ := client.Analyze(ctx, snapapi.AnalyzeParams{
    URL:      "https://example.com",
    Prompt:   "Analyze this page for SEO. List missing meta tags, keyword density issues, and content structure problems.",
    Provider: "openai",
    APIKey:   os.Getenv("OPENAI_API_KEY"),
})
fmt.Println(analysis.Result)
```

### LLM Content Pipeline

```go
// Extract content from multiple pages and feed to your LLM pipeline
urls := []string{
    "https://blog.example.com/post-1",
    "https://blog.example.com/post-2",
}
for _, u := range urls {
    result, err := client.Extract(ctx, snapapi.ExtractParams{
        URL:    u,
        Format: "markdown",
    })
    if err != nil {
        log.Printf("Failed: %v", err)
        continue
    }
    // Feed result.Content to your LLM, RAG pipeline, or embedding model
    fmt.Printf("Extracted %d words from %s\n", result.WordCount, result.URL)
}
```

### Competitor Price Tracking

```go
// Scrape competitor pricing pages
result, err := client.Scrape(ctx, snapapi.ScrapeParams{
    URL:             "https://competitor.com/pricing",
    Selector:        ".pricing-table",
    Format:          "html",
    WaitForSelector: ".price",
})
if err != nil {
    log.Fatal(err)
}
// Parse result.Data for price information
fmt.Printf("Scraped %d chars from pricing page\n", len(result.Data))
```

### Social Media Thumbnail Generation

```go
// Generate OG images sized for social platforms
_, err := client.ScreenshotToFile(ctx, "og_twitter.png", snapapi.ScreenshotParams{
    URL:    "https://mysite.com/blog/my-post",
    Format: "png",
    Width:  1200,
    Height: 628,
    Clip:   &snapapi.ClipRegion{X: 0, Y: 0, Width: 1200, Height: 628},
})
```

### PDF Report Generation

```go
// Generate PDF invoices or reports
pdfBytes, err := client.PDF(ctx, snapapi.PDFParams{
    URL:    "https://myapp.com/invoice/12345",
    Format: "a4",
    Margin: "15mm",
})
if err != nil {
    log.Fatal(err)
}
os.WriteFile("invoice_12345.pdf", pdfBytes, 0644)
```

## Running the Examples

```bash
export SNAPAPI_KEY=sk_live_your_key_here

go run ./examples/screenshot/
go run ./examples/scrape/
go run ./examples/extract/
go run ./examples/analyze/
go run ./examples/advanced/
```

## Running Tests

```bash
go test -race -v ./...
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) before submitting a PR.

```bash
git clone https://github.com/Sleywill/snapapi-go.git
cd snapapi-go
go mod download
go test -race ./...
```

Found a bug? [Open an issue](https://github.com/Sleywill/snapapi-go/issues/new?template=bug_report.md).
Have an idea? [Request a feature](https://github.com/Sleywill/snapapi-go/issues/new?template=feature_request.md).

## Links

- [SnapAPI Website](https://snapapi.pics)
- [API Documentation](https://snapapi.pics/docs)
- [Go Package Reference](https://pkg.go.dev/github.com/Sleywill/snapapi-go)
- [GitHub Repository](https://github.com/Sleywill/snapapi-go)
- [Changelog](./CHANGELOG.md)
- [Report Issues](https://github.com/Sleywill/snapapi-go/issues)

## License

MIT — see [LICENSE](LICENSE).
