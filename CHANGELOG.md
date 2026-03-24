# Changelog

All notable changes to the SnapAPI Go SDK are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [3.2.0] - 2026-03-23

### Added
- `DarkMode` and `BlockCookies` fields on `ScreenshotParams` to match API spec
- `ScrollVideo` field on `VideoParams` for scroll-based video recording
- `Selectors` (map) and `WaitFor` fields on `ScrapeParams` for multi-element scraping
- `GeneratePDF(ctx, PDFParams)` alias for `PDF()`
- `GenerateOGImage(ctx, OGImageParams)` alias for `OGImage()`
- Tests for all new fields and method aliases

### Changed
- Codebase split into multi-file structure: `screenshot.go`, `scrape.go`, `extract.go`, `pdf.go`, `video.go`, `options.go`, `retry.go`, `errors.go` for better maintainability
- User-Agent updated to `snapapi-go/3.2.0`

## [3.1.0] - 2026-03-17

### Added
- `ScreenshotToStorage(ctx, ScreenshotToStorageParams)` -- save captures directly to cloud storage
- `ScrapeText(ctx, url)` and `ScrapeHTML(ctx, url)` convenience wrappers
- `ExtractMarkdown(ctx, url)` and `ExtractText(ctx, url)` convenience wrappers
- `client.Storage` namespace -- `List`, `Get`, `Delete` for stored captures
- `client.Scheduled` namespace -- `Create`, `List`, `Get`, `Delete`, `Pause`, `Resume` for recurring captures
- `client.Webhooks` namespace -- `Create`, `List`, `Get`, `Delete` for webhook endpoints
- `client.APIKeys` namespace -- `Create`, `List`, `Revoke` for API key management
- `ErrNotFound` error code constant for HTTP 404 responses
- `APIError.IsQuotaExceeded()` helper method
- `ErrQuotaExceeded` now mapped from HTTP 402 in `mapErrorCode`
- `ErrNotFound` now mapped from HTTP 404 in `mapErrorCode`
- Additional error string mappings: "Bad Request", "Too Many Requests", "Service Unavailable", "Capture Failed"
- GitHub Actions CI workflow with matrix tests (Go 1.21/1.22/1.23), race detector, coverage, golangci-lint, and tag-triggered release
- `TestRetry_RetryAfterOverridesBackoff` -- verifies Retry-After takes precedence over backoff delay
- `TestClient_NamespacesInitialized` -- verifies namespace fields are non-nil after `New()`
- `TestErrorCode_QuotaExceeded` and `TestErrorCode_NotFound` test cases
- Namespace tests for all four new namespaces (Storage, Scheduled, Webhooks, APIKeys)

### Fixed
- **Retry logic bug**: Retry-After sleep was previously added on top of the exponential backoff sleep instead of replacing it. Now Retry-After takes precedence when present.
- **Manual `errors.As` reimplementation removed**: Test helper now uses stdlib `errors.As` directly.
- **Context cancellation test**: No longer leaves a server goroutine blocked after test completion.

### Changed
- `OGImage(ctx, OGImageParams)` method for Open Graph social image generation
- `Ping(ctx)` method for API health check (`GET /v1/ping`)
- `Quota(ctx)` alias for `GetUsage(ctx)`
- `PDFToFile(ctx, filename, PDFParams)` convenience method
- `PDFParams` now supports `HTML`, `PageSize`, `Landscape`, and individual margin fields
- `Authorization: Bearer` header sent alongside `X-Api-Key` for maximum compatibility
- `UsageResult.Limit` field
- User-Agent updated to `snapapi-go/3.1.0`
- PDF generation uses `/v1/screenshot` with `format=pdf` (matching actual API)

## [2.1.0] - 2026-03-16

### Added
- `Analyze(ctx, AnalyzeParams)` method -- `POST /v1/analyze` for LLM-powered page analysis
- `GetUsage(ctx)` method -- `GET /v1/usage` for checking API usage stats
- `ScreenshotToFile(ctx, filename, ScreenshotParams)` convenience method
- Complete `ScreenshotParams` fields: `Scale`, `BlockAds`, `WaitForSelector`, `Clip`, `ScrollY`, `CustomCSS`, `CustomJS`, `Headers`, `UserAgent`, `Proxy`, `AccessKey`, `Selector`
- Complete `ScrapeParams` fields: `Format`, `WaitForSelector`, `Headers`, `Proxy`, `AccessKey`
- Complete `ExtractParams` fields: `IncludeLinks`, `IncludeImages`, `Selector`, `WaitForSelector`, `Headers`, `Proxy`, `AccessKey`
- `ClipRegion` type for screenshot clipping
- `ErrServiceDown` error code for HTTP 503
- `APIError.IsServiceUnavailable()` helper method
- `examples/analyze/` -- LLM analysis example
- `examples/advanced/` -- real-world use cases (monitoring, SEO, PDF reports, thumbnails)

### Changed
- Base URL corrected to `https://api.snapapi.pics` (was incorrectly `https://snapapi.pics`)
- Auth header changed to `X-Api-Key` to match API specification (was `Authorization: Bearer`)
- `ScrapeResult` fields updated to match API: `Data`, `URL`, `Status` (was `Success`, `HTML`, `Text`)
- `ExtractResult` fields updated to match API: `Content`, `URL`, `WordCount` (removed `Success`, `Format`, `ResponseTime`)
- User-Agent updated to `snapapi-go/2.1.0`
- README overhauled with complete API reference, all parameters, and real-world use cases

### Fixed
- API base URL now matches the actual SnapAPI endpoint
- Authentication header now uses the correct `X-Api-Key` format

## [3.0.0] - 2026-03-14

### Added
- `New()` constructor with functional options pattern
- Retry with exponential backoff
- `PDF()` and `Video()` methods
- Table-driven unit tests with `net/http/httptest`
- GitHub Actions CI workflow

### Changed
- Module bumped to v3.0.0
- Default timeout reduced to 30s

## [2.0.0] - 2026-01-15

- Initial public release with basic screenshot, scrape, and extract support.
