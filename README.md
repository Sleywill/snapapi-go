# SnapAPI Go SDK

Official Go SDK for [SnapAPI](https://snapapi.pics) — Lightning-fast screenshot, PDF, video, extraction, and AI analysis API.

[![Go Reference](https://pkg.go.dev/badge/github.com/Sleywill/snapapi-go.svg)](https://pkg.go.dev/github.com/Sleywill/snapapi-go)

## Installation

```bash
go get github.com/Sleywill/snapapi-go
```

## Quick Start

```go
package main

import (
    "log"
    "os"

    snapapi "github.com/Sleywill/snapapi-go"
)

func main() {
    client := snapapi.NewClient("sk_live_xxx")

    data, err := client.Screenshot(snapapi.ScreenshotOptions{
        URL:    "https://example.com",
        Format: "png",
        Width:  1920,
        Height: 1080,
    })
    if err != nil {
        log.Fatal(err)
    }

    os.WriteFile("screenshot.png", data, 0644)
}
```

## Features

- **Screenshots** — PNG, JPEG, WebP, AVIF formats with device presets, full page, selectors, dark mode
- **PDF Generation** — Dedicated endpoint or via screenshot endpoint with full PDF options
- **Video Capture** — Record webpage videos with scroll animations
- **Batch Screenshots** — Up to 10 URLs in parallel with status polling
- **Content Extraction** — HTML, text, markdown, article, links, images, metadata, structured data
- **AI Analysis** — AI-powered webpage analysis with custom prompts
- **Async Mode** — Asynchronous screenshots with job polling

## Client Configuration

```go
// Default client
client := snapapi.NewClient("sk_live_xxx")

// With options
client := snapapi.NewClient("sk_live_xxx",
    snapapi.WithBaseURL("https://custom-api.example.com"),
    snapapi.WithTimeout(120 * time.Second),
    snapapi.WithHTTPClient(&http.Client{...}),
)
```

## API Reference

### Health Check

```go
result, err := client.Ping()
// result.Status == "ok"
```

### Screenshots

```go
// Basic screenshot
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://example.com",
})

// All format options
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:    "https://example.com",
    Format: "webp", // png, jpeg, webp, avif, pdf
})

// Full page
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:      "https://example.com",
    FullPage: true,
})

// Device preset
data, err := client.ScreenshotDevice("https://example.com", snapapi.DeviceIPhone15Pro, nil)

// Element selector
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:      "https://example.com",
    Selector: "h1",
})

// Dark mode
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:      "https://example.com",
    DarkMode: true,
})

// Custom CSS and JavaScript (Starter+ plan)
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:        "https://example.com",
    CSS:        "body { background: red; }",
    JavaScript: "document.title = 'Hello';",
})

// Block ads, cookies, trackers (Starter+ plan)
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:                "https://example.com",
    BlockAds:           true,
    BlockCookieBanners: true,
    BlockTrackers:      true,
    BlockChatWidgets:   true,
})

// Wait strategies
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:             "https://example.com",
    Delay:           1000,
    WaitForSelector: "h1",
    WaitUntil:       "networkidle",
})

// Thumbnail generation
q := 80
result, err := client.ScreenshotWithMetadata(snapapi.ScreenshotOptions{
    URL:       "https://example.com",
    Quality:   &q,
    Thumbnail: &snapapi.ThumbnailOptions{Enabled: true, Width: 200, Height: 150},
})

// JSON response with metadata
result, err := client.ScreenshotWithMetadata(snapapi.ScreenshotOptions{
    URL:             "https://example.com",
    IncludeMetadata: true,
})
// result.Width, result.Height, result.Format, result.Metadata

// Base64 response
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:          "https://example.com",
    ResponseType: "base64",
})

// From HTML content
data, err := client.ScreenshotFromHTML("<h1>Hello World</h1>", nil)

// From Markdown content
data, err := client.ScreenshotFromMarkdown("# Hello\n\nThis is **bold**.", nil)
```

### Available Device Presets

```go
snapapi.DeviceDesktop1080p       // 1920x1080
snapapi.DeviceDesktop1440p       // 2560x1440
snapapi.DeviceDesktop4K          // 3840x2160
snapapi.DeviceMacBookPro13       // 1440x900 @2x
snapapi.DeviceMacBookPro16       // 1728x1117 @2x
snapapi.DeviceIMac24             // 2240x1260 @2x
snapapi.DeviceIPhoneSE           // 375x667 @2x
snapapi.DeviceIPhone14Pro        // 393x852 @3x
snapapi.DeviceIPhone15Pro        // 393x852 @3x
snapapi.DeviceIPhone15ProMax     // 430x932 @3x
snapapi.DeviceIPadPro11          // 834x1194 @2x
snapapi.DevicePixel8Pro          // 448x998 @2.625x
snapapi.DeviceSamsungGalaxyS24   // 360x780 @3x
// ... and more
```

### PDF Generation

```go
// Via screenshot endpoint
data, err := client.PDF(snapapi.ScreenshotOptions{
    URL: "https://example.com",
})
os.WriteFile("page.pdf", data, 0644)

// From HTML
data, err := client.PDFFromHTML("<h1>Invoice</h1><p>Total: $100</p>", &snapapi.PDFOptions{
    PageSize:  "a4",
    Landscape: boolPtr(true),
})

// Dedicated PDF endpoint with full options
data, err := client.PDFDedicated(snapapi.PDFDedicatedOptions{
    URL: "https://example.com",
    PDFOptions: &snapapi.PDFOptions{
        PageSize:        "a4",
        Landscape:       boolPtr(true),
        PrintBackground: boolPtr(true),
        MarginTop:       "1cm",
        MarginBottom:    "1cm",
    },
})
```

### Video Capture

```go
// Basic video
data, err := client.Video(snapapi.VideoOptions{
    URL:      "https://example.com",
    Duration: 5,
    Width:    1280,
    Height:   720,
})
os.WriteFile("video.mp4", data, 0644)

// Video with scroll animation
data, err := client.Video(snapapi.VideoOptions{
    URL:      "https://example.com",
    Duration: 5,
    Scroll:   true,
    ScrollEasing: snapapi.ScrollEasingEaseInOut,
})

// Video with metadata
result, err := client.VideoWithResult(snapapi.VideoOptions{
    URL:      "https://example.com",
    Duration: 3,
})
// result.Width, result.Height, result.Duration, result.FileSize
```

### Batch Screenshots

```go
// Submit batch job
result, err := client.Batch(snapapi.BatchOptions{
    URLs:   []string{"https://example.com", "https://google.com"},
    Format: "png",
})
fmt.Println("Job ID:", result.JobID)

// Poll for completion
for {
    status, err := client.GetBatchStatus(result.JobID)
    if err != nil {
        log.Fatal(err)
    }
    if status.Status == "completed" {
        for _, item := range status.Results {
            fmt.Printf("%s: %s\n", item.URL, item.Status)
        }
        break
    }
    time.Sleep(2 * time.Second)
}
```

### Content Extraction

```go
// Extract with options
result, err := client.Extract(snapapi.ExtractOptions{
    URL:  "https://example.com",
    Type: "markdown",
})
fmt.Println(result.Content)

// Convenience methods
markdown, err := client.ExtractMarkdown("https://example.com")
text, err := client.ExtractText("https://example.com")
article, err := client.ExtractArticle("https://example.com")
links, err := client.ExtractLinks("https://example.com")     // result.Links
images, err := client.ExtractImages("https://example.com")   // result.Images
metadata, err := client.ExtractMetadata("https://example.com")
structured, err := client.ExtractStructured("https://example.com")
```

### AI Analysis

```go
result, err := client.Analyze(snapapi.AnalyzeOptions{
    URL:    "https://example.com",
    Prompt: "What is this page about? Summarize in 3 bullet points.",
})
fmt.Println(result.Analysis)
```

### Async Screenshots

```go
// Submit async job
result, err := client.ScreenshotAsync(snapapi.ScreenshotOptions{
    URL: "https://example.com",
})

// Poll for result
for {
    status, err := client.GetScreenshotStatus(result.JobID)
    if err != nil {
        log.Fatal(err)
    }
    if status.Status == "completed" {
        fmt.Printf("Screenshot ready: %dx%d\n", status.Result.Width, status.Result.Height)
        break
    }
    time.Sleep(2 * time.Second)
}
```

### Usage & Info

```go
// API usage
usage, err := client.GetUsage()
fmt.Printf("Used: %d/%d (resets %s)\n", usage.Used, usage.Limit, usage.ResetAt)

// Available devices
devices, err := client.GetDevices()
fmt.Printf("%d device presets\n", devices.Total)

// API capabilities
caps, err := client.GetCapabilities()
fmt.Println("Version:", caps.Version)
```

## Error Handling

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{URL: "https://example.com"})
if err != nil {
    if apiErr, ok := err.(*snapapi.APIError); ok {
        fmt.Printf("Code: %s\n", apiErr.Code)           // e.g. "RATE_LIMITED"
        fmt.Printf("Message: %s\n", apiErr.Message)
        fmt.Printf("Status: %d\n", apiErr.StatusCode)    // e.g. 429
        fmt.Printf("Retryable: %v\n", apiErr.IsRetryable())
    }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `INVALID_URL` | Invalid URL provided |
| `INVALID_PARAMS` | Invalid parameters |
| `UNAUTHORIZED` | Invalid API key |
| `FORBIDDEN` | Feature requires higher plan |
| `QUOTA_EXCEEDED` | Monthly quota exceeded |
| `RATE_LIMITED` | Rate limit hit (retryable) |
| `TIMEOUT` | Request timeout (retryable) |
| `CAPTURE_FAILED` | Screenshot capture failed |

## Complete Test App

See [`cmd/test-app/main.go`](cmd/test-app/main.go) for a comprehensive test suite that exercises every SDK method against the live API.

## License

MIT
