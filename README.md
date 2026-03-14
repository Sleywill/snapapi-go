# snapapi-go

Official Go SDK for [SnapAPI.pics](https://snapapi.pics) — capture screenshots, generate PDFs, scrape pages, and extract structured content from any URL.

## Requirements

- Go 1.21+
- A SnapAPI API key — get one at [snapapi.pics](https://snapapi.pics)

## Installation

```bash
go get github.com/Sleywill/snapapi-go@v3
```

## Quick start

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
client := snapapi.New("sk_...",
    // HTTP request timeout (default: 30s)
    snapapi.WithTimeout(45*time.Second),
    // Number of automatic retries on 5xx / rate-limit errors (default: 3)
    snapapi.WithRetries(3),
    // Base delay for exponential back-off (default: 500ms)
    snapapi.WithRetryDelay(500*time.Millisecond),
    // Override the API base URL (useful for testing)
    snapapi.WithBaseURL("https://snapapi.pics"),
    // Bring your own *http.Client (e.g. with a custom transport)
    snapapi.WithHTTPClient(myHTTPClient),
)
```

## Endpoints

### Screenshot

```go
img, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
    URL:      "https://example.com",
    Format:   "png",   // "png" or "jpeg"
    FullPage: true,
    Width:    1280,
    Height:   720,
    Wait:     500,     // ms to wait after page load
    Quality:  85,      // JPEG quality (1-100)
})
```

### Scrape

```go
result, err := client.Scrape(ctx, snapapi.ScrapeParams{
    URL:      "https://example.com",
    Selector: "article", // optional CSS selector
    Wait:     1000,
})
fmt.Println(result.Text)
```

### Extract (LLM-ready content)

```go
result, err := client.Extract(ctx, snapapi.ExtractParams{
    URL:    "https://example.com",
    Format: "markdown", // "markdown", "text", or "json"
})
fmt.Println(result.Content)
```

### PDF

```go
pdfBytes, err := client.PDF(ctx, snapapi.PDFParams{
    URL:    "https://example.com",
    Format: "a4",    // "a4" or "letter"
    Margin: "10mm",
})
os.WriteFile("output.pdf", pdfBytes, 0644)
```

### Video

```go
videoBytes, err := client.Video(ctx, snapapi.VideoParams{
    URL:      "https://example.com",
    Duration: 5,
    Format:   "mp4",
    Width:    1280,
    Height:   720,
})
os.WriteFile("capture.mp4", videoBytes, 0644)
```

### Quota

```go
q, err := client.Quota(ctx)
fmt.Printf("Used: %d / %d\n", q.Used, q.Total)
```

## Error handling

Every method returns `(result, error)` and never panics. Errors are always `*APIError`:

```go
img, err := client.Screenshot(ctx, params)
if err != nil {
    var apiErr *snapapi.APIError
    if errors.As(err, &apiErr) {
        switch {
        case apiErr.IsRateLimited():
            fmt.Printf("Rate limited. Retry after %ds\n", apiErr.RetryAfter)
        case apiErr.IsUnauthorized():
            log.Fatal("Invalid API key")
        case apiErr.IsServerError():
            log.Printf("Server error: %s", apiErr.Message)
        default:
            log.Printf("[%s] %s", apiErr.Code, apiErr.Message)
        }
    }
    return
}
```

### Error codes

| Constant | Meaning |
|---|---|
| `ErrInvalidParams` | Bad request parameters |
| `ErrUnauthorized` | Invalid or missing API key |
| `ErrForbidden` | Insufficient permissions |
| `ErrRateLimited` | Rate limit exceeded (check `RetryAfter`) |
| `ErrQuotaExceeded` | Monthly quota exhausted |
| `ErrTimeout` | Request timed out server-side |
| `ErrCaptureFailed` | Browser capture failed |
| `ErrConnectionError` | Network-level failure |
| `ErrServerError` | Unexpected server error (5xx) |

## Context and cancellation

All methods accept `context.Context` as the first parameter:

```go
ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
defer cancel()
img, err := client.Screenshot(ctx, params)
```

## Running the examples

```bash
export SNAPAPI_KEY=sk_your_key

go run ./examples/screenshot/
go run ./examples/scrape/
go run ./examples/extract/
```

## Running tests

```bash
go test -race ./...
```

## License

MIT — see [LICENSE](LICENSE).
