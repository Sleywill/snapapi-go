# snapapi-go

Official Go SDK for [SnapAPI](https://snapapi.pics) — lightning-fast screenshot, PDF, scraping, and AI web extraction API.

## Installation

```bash
go get github.com/Sleywill/snapapi-go@v2.0.0
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "os"

    snapapi "github.com/Sleywill/snapapi-go"
)

func main() {
    client := snapapi.NewClient("your-api-key")
    ctx := context.Background()

    // Take a screenshot
    img, err := client.Screenshot(ctx, snapapi.ScreenshotOptions{
        URL:      "https://example.com",
        Format:   "png",
        FullPage: true,
    })
    if err != nil {
        panic(err)
    }
    os.WriteFile("screenshot.png", img, 0644)
    fmt.Printf("Saved %d bytes\n", len(img))
}
```

## Authentication

Set your API key when creating the client:

```go
client := snapapi.NewClient("sk_live_...")
```

Or use an environment variable:

```go
client := snapapi.NewClient(os.Getenv("SNAPAPI_KEY"))
```

## Endpoints

### Screenshot — `POST /v1/screenshot`

Capture a webpage as PNG, JPEG, WEBP, AVIF, or PDF.

```go
// Basic screenshot
img, err := client.Screenshot(ctx, snapapi.ScreenshotOptions{
    URL:    "https://example.com",
    Format: "png",
    Width:  1440,
    Height: 900,
})

// Full-page screenshot with dark mode
img, err = client.Screenshot(ctx, snapapi.ScreenshotOptions{
    URL:                "https://example.com",
    FullPage:           true,
    DarkMode:           true,
    BlockAds:           true,
    BlockCookieBanners: true,
})

// Screenshot from HTML
img, err = client.Screenshot(ctx, snapapi.ScreenshotOptions{
    HTML:   "<h1 style='color:red'>Hello!</h1>",
    Format: "png",
})

// PDF generation
pdf, err := client.PDF(ctx, snapapi.ScreenshotOptions{
    URL: "https://example.com",
    PDF: &snapapi.PDFPageOptions{
        PageSize:  "A4",
        Landscape: false,
    },
})

// Screenshot with storage (returns JSON with id+url)
result, err := client.ScreenshotToStorage(ctx, snapapi.ScreenshotOptions{
    URL:    "https://example.com",
    Format: "png",
    Storage: &snapapi.StorageDestination{
        Destination: "s3",
    },
})
fmt.Println(result.URL)
```

### Scrape — `POST /v1/scrape`

Scrape text, HTML, or links from websites.

```go
result, err := client.Scrape(ctx, snapapi.ScrapeOptions{
    URL:   "https://example.com",
    Type:  "text",  // text|html|links
    Pages: 3,
})
for _, item := range result.Results {
    fmt.Printf("Page %d: %s\n", item.Page, item.URL)
}
```

### Extract — `POST /v1/extract`

Extract structured content: articles, metadata, links, images.

```go
// Article extraction
article, err := client.ExtractArticle(ctx, "https://example.com/post")

// Markdown extraction
md, err := client.ExtractMarkdown(ctx, "https://example.com")

// Custom extraction
result, err := client.Extract(ctx, snapapi.ExtractOptions{
    URL:           "https://example.com",
    Type:          "structured",
    BlockAds:      true,
    IncludeImages: true,
    MaxLength:     intPtr(5000),
})
```

### Analyze — `POST /v1/analyze`

AI-powered webpage analysis using OpenAI or Anthropic.

```go
result, err := client.Analyze(ctx, snapapi.AnalyzeOptions{
    URL:               "https://example.com",
    Prompt:            "What is the main purpose of this page?",
    Provider:          "openai",
    APIKey:            "sk-...", // your LLM API key
    IncludeScreenshot: true,
    IncludeMetadata:   true,
})
fmt.Println(result.Analysis)
```

### Storage — `/v1/storage/*`

```go
// List stored files
files, err := client.Storage.ListFiles(ctx)

// Storage usage
usage, err := client.Storage.GetUsage(ctx)
fmt.Printf("Used: %d / %d bytes\n", usage.Used, usage.Limit)

// Configure S3
err = client.Storage.ConfigureS3(ctx, snapapi.S3Config{
    Bucket:          "my-bucket",
    Region:          "us-east-1",
    AccessKeyID:     "AKIA...",
    SecretAccessKey: "...",
})

// Delete a file
err = client.Storage.DeleteFile(ctx, "file-id")
```

### Scheduled — `/v1/scheduled/*`

```go
// Create hourly screenshot job
job, err := client.Scheduled.Create(ctx, snapapi.ScheduledOptions{
    URL:            "https://example.com",
    CronExpression: "0 * * * *",
    Format:         "png",
    FullPage:       true,
    WebhookURL:     "https://myapp.com/webhook",
})

// List all jobs
jobs, err := client.Scheduled.List(ctx)

// Delete a job
err = client.Scheduled.Delete(ctx, job.ID)
```

### Webhooks — `/v1/webhooks/*`

```go
// Register a webhook
hook, err := client.Webhooks.Create(ctx, snapapi.WebhookOptions{
    URL:    "https://myapp.com/snapapi-webhook",
    Events: []string{"screenshot.completed", "scheduled.run"},
    Secret: "my-secret",
})

// List all webhooks
hooks, err := client.Webhooks.List(ctx)

// Delete a webhook
err = client.Webhooks.Delete(ctx, hook.ID)
```

### API Keys — `/v1/keys/*`

```go
// List all keys
keys, err := client.Keys.List(ctx)

// Create a new key
key, err := client.Keys.Create(ctx, "production-key")
fmt.Println(key.Key) // only shown on creation

// Revoke a key
err = client.Keys.Delete(ctx, key.ID)
```

## Client Options

```go
client := snapapi.NewClient(apiKey,
    snapapi.WithTimeout(120*time.Second),
    snapapi.WithBaseURL("https://api.snapapi.pics"),
)
```

## Error Handling

```go
img, err := client.Screenshot(ctx, opts)
if err != nil {
    var apiErr *snapapi.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error %s: %s (HTTP %d)\n", apiErr.Code, apiErr.Message, apiErr.StatusCode)
        if apiErr.IsRetryable() {
            // retry with backoff
        }
    }
    return err
}
```

### Error Codes

| Code | Meaning |
|------|---------|
| `INVALID_PARAMS` | Missing or invalid request parameters |
| `UNAUTHORIZED` | Invalid or missing API key |
| `FORBIDDEN` | Plan does not include this feature |
| `RATE_LIMITED` | Too many requests |
| `TIMEOUT` | Page load timed out |
| `QUOTA_EXCEEDED` | Monthly quota reached |
| `CONNECTION_ERROR` | Network error |
| `SERVER_ERROR` | Upstream server error |

## Running the Example

```bash
SNAPAPI_KEY=your-key go run examples/basic/main.go
```

## License

MIT
