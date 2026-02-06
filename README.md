# SnapAPI Go SDK

Official Go SDK for [SnapAPI](https://snapapi.pics) - Lightning-fast screenshot API for developers.

## Installation

```bash
go get github.com/snapapi-dev/snapapi-go
```

## Quick Start

```go
package main

import (
    "log"
    "os"

    "github.com/snapapi-dev/snapapi-go"
)

func main() {
    client := snapapi.NewClient("sk_live_xxx")

    // Capture a screenshot
    data, err := client.Screenshot(snapapi.ScreenshotOptions{
        URL:    "https://example.com",
        Format: "png",
        Width:  1920,
        Height: 1080,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Save to file
    os.WriteFile("screenshot.png", data, 0644)
}
```

## Usage Examples

### Basic Screenshot

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://example.com",
})
```

### Full Page Screenshot

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:      "https://example.com",
    FullPage: true,
    Format:   "png",
})
```

### Device Presets

Capture screenshots using pre-configured device settings:

```go
// Using device preset
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:    "https://example.com",
    Device: snapapi.DeviceIPhone15Pro,
})

// Or use the convenience method
data, err := client.ScreenshotDevice("https://example.com", snapapi.DeviceIPadPro129, nil)

// Get all available device presets
devices, err := client.GetDevices()
fmt.Printf("Total devices: %d\n", devices.Total)
```

Available device presets:
- **Desktop**: `DeviceDesktop1080p`, `DeviceDesktop1440p`, `DeviceDesktop4K`
- **Mac**: `DeviceMacBookPro13`, `DeviceMacBookPro16`, `DeviceIMac24`
- **iPhone**: `DeviceIPhoneSE`, `DeviceIPhone12`, `DeviceIPhone13`, `DeviceIPhone14`, `DeviceIPhone14Pro`, `DeviceIPhone15`, `DeviceIPhone15Pro`, `DeviceIPhone15ProMax`
- **iPad**: `DeviceIPad`, `DeviceIPadMini`, `DeviceIPadAir`, `DeviceIPadPro11`, `DeviceIPadPro129`
- **Android**: `DevicePixel7`, `DevicePixel8`, `DevicePixel8Pro`, `DeviceSamsungGalaxyS23`, `DeviceSamsungGalaxyS24`, `DeviceSamsungGalaxyTabS9`

### Dark Mode

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:      "https://example.com",
    DarkMode: true,
})
```

### Screenshot from HTML

```go
html := "<html><body><h1>Hello World</h1></body></html>"
data, err := client.ScreenshotFromHTML(html, &snapapi.ScreenshotOptions{
    Width:  800,
    Height: 600,
})
```

### PDF Export

```go
pdfData, err := client.PDF(snapapi.ScreenshotOptions{
    URL: "https://example.com",
    PDFOptions: &snapapi.PDFOptions{
        PageSize:            "a4",
        Landscape:           ptrBool(false),
        MarginTop:           "20mm",
        MarginBottom:        "20mm",
        MarginLeft:          "15mm",
        MarginRight:         "15mm",
        PrintBackground:     ptrBool(true),
        DisplayHeaderFooter: ptrBool(true),
        HeaderTemplate:      `<div style="font-size:10px;text-align:center;width:100%;">Header</div>`,
        FooterTemplate:      `<div style="font-size:10px;text-align:center;width:100%;">Page <span class="pageNumber"></span></div>`,
    },
})

os.WriteFile("document.pdf", pdfData, 0644)

// Helper function
func ptrBool(b bool) *bool { return &b }
```

### Geolocation Emulation

```go
accuracy := 100.0
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://maps.google.com",
    Geolocation: &snapapi.Geolocation{
        Latitude:  48.8566,
        Longitude: 2.3522,
        Accuracy:  &accuracy,
    },
})
```

### Timezone & Locale

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:      "https://example.com",
    Timezone: "America/New_York",
    Locale:   "en-US",
})
```

### Proxy Support

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://example.com",
    Proxy: &snapapi.ProxyConfig{
        Server:   "http://proxy.example.com:8080",
        Username: "user",
        Password: "pass",
    },
})
```

### Hide Elements

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://example.com",
    HideSelectors: []string{
        ".cookie-banner",
        "#popup-modal",
        ".advertisement",
    },
})
```

### Click Before Screenshot

```go
clickDelay := 500
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:           "https://example.com",
    ClickSelector: ".accept-cookies-button",
    ClickDelay:    &clickDelay,
    Delay:         1000,
})
```

### Block Ads, Trackers, Chat Widgets

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:                "https://example.com",
    BlockAds:           true,
    BlockTrackers:      true,
    BlockCookieBanners: true,
    BlockChatWidgets:   true, // Blocks Intercom, Drift, Zendesk, etc.
})
```

### Thumbnail Generation

```go
result, err := client.ScreenshotWithMetadata(snapapi.ScreenshotOptions{
    URL: "https://example.com",
    Thumbnail: &snapapi.ThumbnailOptions{
        Enabled: true,
        Width:   300,
        Height:  200,
        Fit:     "cover", // "cover", "contain", or "fill"
    },
})

// Access both full image and thumbnail
fullImage := result.Data      // base64 encoded
thumbnail := result.Thumbnail // base64 encoded
```

### Fail on HTTP Errors

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:             "https://example.com/404-page",
    FailOnHTTPError: true, // Will return error if page returns 4xx or 5xx
})
if err != nil {
    fmt.Println("Page returned HTTP error")
}
```

### Custom JavaScript Execution

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://example.com",
    JavaScript: `
        document.querySelector('.popup')?.remove();
        document.body.style.background = 'white';
    `,
    Delay: 1000,
})
```

### Custom CSS

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://example.com",
    CSS: `
        body { background: #f0f0f0 !important; }
        .ads, .banner { display: none !important; }
    `,
})
```

### With Cookies (Authenticated Pages)

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://example.com/dashboard",
    Cookies: []snapapi.Cookie{
        {
            Name:   "session",
            Value:  "abc123",
            Domain: "example.com",
        },
    },
})
```

### HTTP Basic Authentication

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "https://example.com/protected",
    HTTPAuth: &snapapi.HTTPAuth{
        Username: "user",
        Password: "pass",
    },
})
```

### Element Screenshot with Clipping

```go
// Capture specific element
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:      "https://example.com",
    Selector: ".hero-section",
})

// Or use manual clipping
clipX := 100
clipY := 100
clipW := 500
clipH := 300
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL:        "https://example.com",
    ClipX:      &clipX,
    ClipY:      &clipY,
    ClipWidth:  &clipW,
    ClipHeight: &clipH,
})
```

### Extract Metadata

```go
fonts := true
colors := true
links := true
httpStatus := true

result, err := client.ScreenshotWithMetadata(snapapi.ScreenshotOptions{
    URL:             "https://example.com",
    IncludeMetadata: true,
    ExtractMetadata: &snapapi.ExtractMetadata{
        Fonts:          &fonts,
        Colors:         &colors,
        Links:          &links,
        HTTPStatusCode: &httpStatus,
    },
})

fmt.Printf("Title: %s\n", result.Metadata.Title)
fmt.Printf("HTTP Status: %d\n", result.Metadata.HTTPStatusCode)
fmt.Printf("Fonts: %v\n", result.Metadata.Fonts)
fmt.Printf("Colors: %v\n", result.Metadata.Colors)
fmt.Printf("Links: %v\n", result.Metadata.Links)
```

### Get Screenshot with Metadata

```go
result, err := client.ScreenshotWithMetadata(snapapi.ScreenshotOptions{
    URL:             "https://example.com",
    IncludeMetadata: true,
})

fmt.Printf("Width: %d\n", result.Width)
fmt.Printf("Height: %d\n", result.Height)
fmt.Printf("File Size: %d\n", result.FileSize)
fmt.Printf("Duration: %dms\n", result.Took)
fmt.Printf("Cached: %v\n", result.Cached)
// result.Data contains base64 encoded image
```

### Batch Screenshots

```go
batch, err := client.Batch(snapapi.BatchOptions{
    URLs: []string{
        "https://example.com",
        "https://example.org",
        "https://example.net",
    },
    Format:     "png",
    WebhookURL: "https://your-server.com/webhook",
})

fmt.Printf("Job ID: %s\n", batch.JobID)

// Check status later
status, err := client.GetBatchStatus(batch.JobID)
if status.Status == "completed" {
    for _, result := range status.Results {
        fmt.Printf("%s: %s\n", result.URL, result.Status)
    }
}
```

### Get API Capabilities

```go
capabilities, err := client.GetCapabilities()
fmt.Printf("API Version: %s\n", capabilities.Version)
fmt.Printf("Capabilities: %v\n", capabilities.Capabilities)
```

### Get API Usage

```go
usage, err := client.GetUsage()
fmt.Printf("Used: %d\n", usage.Used)
fmt.Printf("Limit: %d\n", usage.Limit)
fmt.Printf("Remaining: %d\n", usage.Remaining)
fmt.Printf("Resets at: %s\n", usage.ResetAt)
```

### Screenshot from Markdown

```go
markdown := "# Hello World\n\nThis is **bold** and this is *italic*.\n\n- Item 1\n- Item 2\n- Item 3"
data, err := client.ScreenshotFromMarkdown(markdown, &snapapi.ScreenshotOptions{
    Width:  800,
    Height: 600,
})
if err != nil {
    log.Fatal(err)
}
os.WriteFile("markdown.png", data, 0644)
```

### Extract Content

```go
// Extract as Markdown
result, err := client.ExtractMarkdown("https://example.com")
fmt.Println(result.Content)

// Extract article content
result, err := client.ExtractArticle("https://blog.example.com/post")
fmt.Printf("Title: %s\n", result.Title)
fmt.Println(result.Content)

// Extract with full options
result, err := client.Extract(snapapi.ExtractOptions{
    URL:                "https://example.com",
    Type:               "markdown",
    Selector:           ".main-content",
    BlockAds:           true,
    BlockCookieBanners: true,
    CleanOutput:        true,
})

// Extract plain text
result, err := client.ExtractText("https://example.com")

// Extract structured data
result, err := client.ExtractStructured("https://example.com")

// Extract links
result, err := client.ExtractLinks("https://example.com")
fmt.Printf("Found %d links\n", len(result.Links))

// Extract images
result, err := client.ExtractImages("https://example.com")
fmt.Printf("Found %d images\n", len(result.Images))

// Extract metadata
result, err := client.ExtractMetadata("https://example.com")
fmt.Printf("Title: %s\n", result.Title)
fmt.Printf("Metadata: %v\n", result.Metadata)
```

### Analyze with AI

```go
// Basic analysis
result, err := client.Analyze(snapapi.AnalyzeOptions{
    URL:    "https://example.com",
    Prompt: "Describe the main content and purpose of this webpage",
})
fmt.Println(result.Analysis)

// Analysis with custom provider and model
result, err := client.Analyze(snapapi.AnalyzeOptions{
    URL:      "https://example.com/pricing",
    Prompt:   "Extract all pricing tiers and features as JSON",
    Provider: "openai",
    APIKey:   "sk-xxx",
    Model:    "gpt-4o",
})
fmt.Println(result.Analysis)

// Analysis with screenshot and metadata
result, err := client.Analyze(snapapi.AnalyzeOptions{
    URL:               "https://example.com",
    Prompt:            "Is this website accessible and mobile-friendly?",
    IncludeScreenshot: true,
    IncludeMetadata:   true,
    BlockAds:          true,
    BlockCookieBanners: true,
})
fmt.Println(result.Analysis)
// result.Screenshot contains base64 encoded screenshot
// result.Metadata contains page metadata
```

## Configuration Options

### Client Options

```go
// Custom base URL
client := snapapi.NewClient("sk_live_xxx",
    snapapi.WithBaseURL("https://custom-api.example.com"),
)

// Custom timeout
client := snapapi.NewClient("sk_live_xxx",
    snapapi.WithTimeout(120 * time.Second),
)

// Custom HTTP client
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns: 10,
    },
}
client := snapapi.NewClient("sk_live_xxx",
    snapapi.WithHTTPClient(httpClient),
)
```

### Screenshot Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `URL` | string | - | URL to capture (required if no HTML/Markdown) |
| `HTML` | string | - | HTML content to render (required if no URL/Markdown) |
| `Markdown` | string | - | Markdown content to render as screenshot |
| `Format` | string | `"png"` | `"png"`, `"jpeg"`, `"webp"`, `"avif"`, `"pdf"` |
| `Quality` | *int | `80` | Image quality 1-100 (JPEG/WebP) |
| `Device` | DevicePreset | - | Device preset name |
| `Width` | int | `1280` | Viewport width (100-3840) |
| `Height` | int | `800` | Viewport height (100-2160) |
| `DeviceScaleFactor` | *float64 | `1` | Device pixel ratio (1-3) |
| `IsMobile` | bool | `false` | Emulate mobile device |
| `HasTouch` | bool | `false` | Enable touch events |
| `IsLandscape` | bool | `false` | Landscape orientation |
| `FullPage` | bool | `false` | Capture full scrollable page |
| `FullPageScrollDelay` | *int | `400` | Delay between scroll steps (ms) |
| `FullPageMaxHeight` | *int | - | Max height for full page (px) |
| `Selector` | string | - | CSS selector for element capture |
| `SelectorScrollIntoView` | *bool | - | Scroll element into view |
| `ClipX`, `ClipY` | *int | - | Clip region position |
| `ClipWidth`, `ClipHeight` | *int | - | Clip region size |
| `Delay` | int | `0` | Delay before capture (0-30000ms) |
| `Timeout` | int | `30000` | Max wait time (1000-60000ms) |
| `WaitUntil` | string | `"load"` | `"load"`, `"domcontentloaded"`, `"networkidle"` |
| `WaitForSelector` | string | - | Wait for element before capture |
| `WaitForSelectorTimeout` | *int | - | Timeout for wait selector |
| `DarkMode` | bool | `false` | Emulate dark mode |
| `ReducedMotion` | bool | `false` | Reduce animations |
| `CSS` | string | - | Custom CSS to inject |
| `JavaScript` | string | - | JS to execute before capture |
| `HideSelectors` | []string | - | CSS selectors to hide |
| `ClickSelector` | string | - | Element to click before capture |
| `ClickDelay` | *int | - | Delay after click (ms) |
| `BlockAds` | bool | `false` | Block ads |
| `BlockTrackers` | bool | `false` | Block trackers |
| `BlockCookieBanners` | bool | `false` | Hide cookie banners |
| `BlockChatWidgets` | bool | `false` | Block chat widgets |
| `BlockResources` | []string | - | Resource types to block |
| `UserAgent` | string | - | Custom User-Agent |
| `ExtraHeaders` | map[string]string | - | Custom HTTP headers |
| `Cookies` | []Cookie | - | Cookies to set |
| `HTTPAuth` | *HTTPAuth | - | HTTP basic auth credentials |
| `Proxy` | *ProxyConfig | - | Proxy configuration |
| `Geolocation` | *Geolocation | - | Geolocation coordinates |
| `Timezone` | string | - | Timezone (e.g., "America/New_York") |
| `Locale` | string | - | Locale (e.g., "en-US") |
| `PDFOptions` | *PDFOptions | - | PDF generation options |
| `Thumbnail` | *ThumbnailOptions | - | Thumbnail generation options |
| `FailOnHTTPError` | bool | `false` | Fail on 4xx/5xx responses |
| `Cache` | bool | `false` | Enable caching |
| `CacheTTL` | *int | `86400` | Cache TTL in seconds |
| `ResponseType` | string | `"binary"` | `"binary"`, `"base64"`, `"json"` |
| `IncludeMetadata` | bool | `false` | Include page metadata |
| `ExtractMetadata` | *ExtractMetadata | - | Additional metadata to extract |

### PDF Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `PageSize` | string | `"a4"` | `"a4"`, `"a3"`, `"a5"`, `"letter"`, `"legal"`, `"tabloid"` |
| `Width` | string | - | Custom width (e.g., "210mm") |
| `Height` | string | - | Custom height (e.g., "297mm") |
| `Landscape` | *bool | `false` | Landscape orientation |
| `MarginTop` | string | - | Top margin (e.g., "20mm") |
| `MarginRight` | string | - | Right margin |
| `MarginBottom` | string | - | Bottom margin |
| `MarginLeft` | string | - | Left margin |
| `PrintBackground` | *bool | `true` | Print background graphics |
| `HeaderTemplate` | string | - | HTML template for header |
| `FooterTemplate` | string | - | HTML template for footer |
| `DisplayHeaderFooter` | *bool | `false` | Show header/footer |
| `Scale` | *float64 | `1` | Scale (0.1-2) |
| `PageRanges` | string | - | Page ranges (e.g., "1-5") |
| `PreferCSSPageSize` | *bool | `false` | Use CSS page size |

### Extract Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `URL` | string | - | URL to extract content from (required) |
| `Type` | string | `"markdown"` | `"markdown"`, `"article"`, `"structured"`, `"text"`, `"links"`, `"images"`, `"metadata"` |
| `Selector` | string | - | CSS selector to scope extraction |
| `WaitFor` | string | - | Wait for selector before extraction |
| `Timeout` | int | `30000` | Max wait time (ms) |
| `DarkMode` | bool | `false` | Emulate dark mode |
| `BlockAds` | bool | `false` | Block ads |
| `BlockCookieBanners` | bool | `false` | Hide cookie banners |
| `IncludeImages` | bool | `false` | Include image references |
| `MaxLength` | *int | - | Max content length (characters) |
| `CleanOutput` | bool | `false` | Clean and simplify output |

### Analyze Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `URL` | string | - | URL to analyze (required) |
| `Prompt` | string | - | Analysis prompt/question (required) |
| `Provider` | string | - | AI provider (e.g., `"openai"`, `"anthropic"`) |
| `APIKey` | string | - | API key for the AI provider |
| `Model` | string | - | Model to use (e.g., `"gpt-4o"`) |
| `JSONSchema` | string | - | JSON schema for structured output |
| `Timeout` | int | `60000` | Max wait time (ms) |
| `WaitFor` | string | - | Wait for selector before analysis |
| `BlockAds` | bool | `false` | Block ads |
| `BlockCookieBanners` | bool | `false` | Hide cookie banners |
| `IncludeScreenshot` | bool | `false` | Include page screenshot |
| `IncludeMetadata` | bool | `false` | Include page metadata |
| `MaxContentLength` | *int | - | Max content length for analysis |

## Error Handling

```go
data, err := client.Screenshot(snapapi.ScreenshotOptions{
    URL: "invalid-url",
})
if err != nil {
    if apiErr, ok := err.(*snapapi.APIError); ok {
        fmt.Printf("Code: %s\n", apiErr.Code)           // "INVALID_URL"
        fmt.Printf("Status: %d\n", apiErr.StatusCode)   // 400
        fmt.Printf("Message: %s\n", apiErr.Message)     // "The provided URL is not valid"
        fmt.Printf("Details: %v\n", apiErr.Details)     // map[url:invalid-url]

        // Check if error is retryable
        if apiErr.IsRetryable() {
            // Implement retry logic
        }
    }
}
```

### Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `INVALID_URL` | 400 | URL is malformed or not accessible |
| `INVALID_PARAMS` | 400 | One or more parameters are invalid |
| `UNAUTHORIZED` | 401 | Missing or invalid API key |
| `FORBIDDEN` | 403 | API key doesn't have permission |
| `QUOTA_EXCEEDED` | 429 | Monthly quota exceeded |
| `RATE_LIMITED` | 429 | Too many requests |
| `TIMEOUT` | 504 | Page took too long to load |
| `CAPTURE_FAILED` | 500 | Screenshot capture failed |
| `HTTP_ERROR` | varies | Page returned HTTP error (with FailOnHTTPError) |

## License

MIT
