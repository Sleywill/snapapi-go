# Changelog

All notable changes to the SnapAPI Go SDK are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

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
