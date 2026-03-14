# Changelog

All notable changes to the SnapAPI Go SDK are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [3.0.0] - 2026-03-14

### Added
- `New()` constructor with functional options pattern (`WithTimeout`, `WithRetries`, `WithRetryDelay`, `WithBaseURL`, `WithHTTPClient`)
- `WithRetries(n int)` option — exponential back-off with configurable base delay
- `WithRetryDelay(d time.Duration)` option — controls the initial retry delay
- `Quota(ctx)` method — `GET /v1/quota` returns `QuotaResult{Used, Total, Remaining, ResetAt}`
- `Video(ctx, VideoParams)` method — `POST /v1/video`
- `PDF(ctx, PDFParams)` method — `POST /v1/pdf`
- Strongly-typed parameter structs: `ScreenshotParams`, `ScrapeParams`, `ExtractParams`, `PDFParams`, `VideoParams`
- `APIError.RetryAfter` field parsed from `Retry-After` response header
- `APIError.IsRateLimited()`, `IsUnauthorized()`, `IsServerError()` helpers
- `isRetryable()` internal helper — 5xx and rate-limit errors are retried; 4xx are not
- `Authorization: Bearer <key>` header (replaces `x-api-key`)
- Table-driven unit tests with `net/http/httptest` (no network required)
- GitHub Actions CI workflow (Go 1.21 / 1.22 / 1.23, golangci-lint)
- Examples: `examples/screenshot/`, `examples/scrape/`, `examples/extract/`

### Changed
- Module bumped to **v3.0.0**
- Default timeout reduced from 90s to **30s** (overridable)
- Default base URL changed from `https://api.snapapi.pics` to `https://snapapi.pics`
- `NewClient()` kept as alias for `New()` for backward compatibility

### Removed
- Sub-clients (`Storage`, `Scheduled`, `Webhooks`, `Keys`) — removed to match minimal API scope; will be re-added in a future release
- `Analyze()` — server-side endpoint is currently broken

## [2.0.0] - 2026-01-15

- Initial public release with basic screenshot, scrape, and extract support.
