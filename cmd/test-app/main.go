package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	snapapi "github.com/Sleywill/snapapi-go"
)

var apiKey = os.Getenv("SNAPAPI_KEY")


var (
	passed  int
	failed  int
	skipped int
)

func test(name string, fn func() error) {
	fmt.Printf("%-60s ", name)
	// Retry once on rate limit with backoff
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = fn()
		if err == nil {
			break
		}
		if apiErr, ok := err.(*snapapi.APIError); ok && apiErr.IsRetryable() && apiErr.Code == "RATE_LIMITED" {
			wait := 65 * time.Second
			fmt.Printf("[rate limited, waiting 65s] ")
			time.Sleep(wait)
			continue
		}
		break
	}
	if err != nil {
		if strings.Contains(err.Error(), "SKIP:") {
			fmt.Printf("⏭️  %s\n", err.Error())
			skipped++
		} else {
			fmt.Printf("❌ %s\n", err.Error())
			failed++
		}
	} else {
		fmt.Println("✅")
		passed++
	}
	// Delay between tests to avoid rate limiting (100 req/min)
	time.Sleep(3 * time.Second)
}

func main() {
	if apiKey == "" {
		fmt.Println("Set SNAPAPI_KEY environment variable")
		os.Exit(1)
	}
	client := snapapi.NewClient(apiKey)

	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println("SnapAPI Go SDK — Comprehensive Test Suite")
	fmt.Println("=" + strings.Repeat("=", 79))
	fmt.Println()

	// ─── 1. Ping ───
	fmt.Println("── GET /v1/ping ──")
	test("Ping", func() error {
		r, err := client.Ping()
		if err != nil {
			return err
		}
		if r.Status != "ok" {
			return fmt.Errorf("expected status ok, got %s", r.Status)
		}
		return nil
	})

	// ─── 2. Capabilities ───
	fmt.Println("\n── GET /v1/capabilities ──")
	test("GetCapabilities", func() error {
		r, err := client.GetCapabilities()
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		if r.Version == "" {
			return fmt.Errorf("empty version")
		}
		return nil
	})

	// ─── 3. Devices ───
	fmt.Println("\n── GET /v1/devices ──")
	test("GetDevices", func() error {
		r, err := client.GetDevices()
		if err != nil {
			return err
		}
		if !r.Success || r.Total == 0 {
			return fmt.Errorf("success=%v total=%d", r.Success, r.Total)
		}
		return nil
	})

	// ─── 4. Usage ───
	fmt.Println("\n── GET /v1/usage ──")
	test("GetUsage", func() error {
		r, err := client.GetUsage()
		if err != nil {
			return err
		}
		if r.Limit == 0 {
			return fmt.Errorf("limit=0")
		}
		fmt.Printf("[%d/%d] ", r.Used, r.Limit)
		return nil
	})

	// ─── 5. Screenshot (URL, various formats) ───
	fmt.Println("\n── POST /v1/screenshot ──")
	test("Screenshot URL (png)", func() error {
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL: "https://example.com", Format: "png", Width: 1280, Height: 720,
		})
		if err != nil {
			return err
		}
		if len(data) < 1000 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot URL (jpeg)", func() error {
		q := 80
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL: "https://example.com", Format: "jpeg", Quality: &q,
		})
		if err != nil {
			return err
		}
		if len(data) < 500 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot URL (webp)", func() error {
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL: "https://example.com", Format: "webp",
		})
		if err != nil {
			return err
		}
		if len(data) < 500 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot URL (avif)", func() error {
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL: "https://example.com", Format: "avif",
		})
		if err != nil {
			return err
		}
		if len(data) < 100 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot fullPage", func() error {
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL: "https://example.com", FullPage: true,
		})
		if err != nil {
			return err
		}
		if len(data) < 1000 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot with device preset", func() error {
		data, err := client.ScreenshotDevice("https://example.com", snapapi.DeviceIPhone15Pro, nil)
		if err != nil {
			return err
		}
		if len(data) < 500 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot with selector", func() error {
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL: "https://example.com", Selector: "h1",
		})
		if err != nil {
			return err
		}
		if len(data) < 100 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot dark mode", func() error {
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL: "https://example.com", DarkMode: true,
		})
		if err != nil {
			return err
		}
		if len(data) < 500 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot custom CSS/JS (Starter+ plan)", func() error {
		_, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL:        "https://example.com",
			CSS:        "body { background: red; }",
			JavaScript: "document.title = 'test';",
		})
		if err != nil {
			if apiErr, ok := err.(*snapapi.APIError); ok && apiErr.StatusCode == 403 {
				return fmt.Errorf("SKIP: requires Starter plan")
			}
			return err
		}
		return nil
	})

	test("Screenshot block ads/cookies/trackers (Starter+)", func() error {
		_, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL:                "https://example.com",
			BlockAds:           true,
			BlockCookieBanners: true,
			BlockTrackers:      true,
			BlockChatWidgets:   true,
		})
		if err != nil {
			if apiErr, ok := err.(*snapapi.APIError); ok && apiErr.StatusCode == 403 {
				return fmt.Errorf("SKIP: requires Starter plan")
			}
			return err
		}
		return nil
	})

	test("Screenshot with delay + waitForSelector", func() error {
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL:             "https://example.com",
			Delay:           500,
			WaitForSelector: "h1",
		})
		if err != nil {
			return err
		}
		if len(data) < 500 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Screenshot with thumbnail", func() error {
		r, err := client.ScreenshotWithMetadata(snapapi.ScreenshotOptions{
			URL:       "https://example.com",
			Thumbnail: &snapapi.ThumbnailOptions{Enabled: true, Width: 200, Height: 150},
		})
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		return nil
	})

	test("Screenshot responseType=json + includeMetadata", func() error {
		r, err := client.ScreenshotWithMetadata(snapapi.ScreenshotOptions{
			URL:             "https://example.com",
			IncludeMetadata: true,
		})
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		if r.Format == "" {
			return fmt.Errorf("empty format")
		}
		fmt.Printf("[%dx%d %s] ", r.Width, r.Height, r.Format)
		return nil
	})

	test("Screenshot responseType=base64", func() error {
		data, err := client.Screenshot(snapapi.ScreenshotOptions{
			URL: "https://example.com", ResponseType: "base64",
		})
		if err != nil {
			return err
		}
		if len(data) < 100 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	// ─── Screenshot from HTML ───
	test("Screenshot from HTML", func() error {
		data, err := client.ScreenshotFromHTML("<html><body><h1>Hello Go SDK!</h1></body></html>", nil)
		if err != nil {
			return err
		}
		if len(data) < 500 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	// ─── Screenshot from Markdown ───
	test("Screenshot from Markdown", func() error {
		data, err := client.ScreenshotFromMarkdown("# Hello\n\nThis is **bold** from Go SDK.", nil)
		if err != nil {
			return err
		}
		if len(data) < 500 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	// ─── 6. PDF (via screenshot endpoint) ───
	fmt.Println("\n── POST /v1/screenshot (format=pdf) ──")
	test("PDF from URL (screenshot endpoint)", func() error {
		data, err := client.PDF(snapapi.ScreenshotOptions{URL: "https://example.com"})
		if err != nil {
			return err
		}
		if len(data) < 500 || string(data[:5]) != "%PDF-" {
			return fmt.Errorf("not a valid PDF (%d bytes)", len(data))
		}
		return nil
	})

	test("PDF from HTML", func() error {
		data, err := client.PDFFromHTML("<h1>Hello PDF</h1><p>Generated by Go SDK</p>", nil)
		if err != nil {
			return err
		}
		if len(data) < 500 || string(data[:5]) != "%PDF-" {
			return fmt.Errorf("not a valid PDF (%d bytes)", len(data))
		}
		return nil
	})

	// ─── 7. Dedicated PDF endpoint ───
	fmt.Println("\n── POST /v1/pdf ──")
	test("PDFDedicated from URL", func() error {
		data, err := client.PDFDedicated(snapapi.PDFDedicatedOptions{URL: "https://example.com"})
		if err != nil {
			return err
		}
		if len(data) < 500 || string(data[:5]) != "%PDF-" {
			return fmt.Errorf("not a valid PDF (%d bytes)", len(data))
		}
		return nil
	})

	test("PDFDedicated with options", func() error {
		landscape := true
		printBg := true
		data, err := client.PDFDedicated(snapapi.PDFDedicatedOptions{
			URL: "https://example.com",
			PDFOptions: &snapapi.PDFOptions{
				PageSize:        "a4",
				Landscape:       &landscape,
				PrintBackground: &printBg,
			},
		})
		if err != nil {
			return err
		}
		if len(data) < 500 || string(data[:5]) != "%PDF-" {
			return fmt.Errorf("not a valid PDF (%d bytes)", len(data))
		}
		return nil
	})

	// ─── 8. Batch ───
	fmt.Println("\n── POST /v1/screenshot/batch ──")
	var batchJobID string
	test("Batch screenshot", func() error {
		r, err := client.Batch(snapapi.BatchOptions{
			URLs:   []string{"https://example.com", "https://httpbin.org/html"},
			Format: "png",
		})
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		batchJobID = r.JobID
		fmt.Printf("[jobId=%s] ", r.JobID)
		return nil
	})

	// ─── 9. Batch status polling ───
	test("Batch status polling", func() error {
		if batchJobID == "" {
			return fmt.Errorf("SKIP: no batch job")
		}
		for i := 0; i < 15; i++ {
			time.Sleep(2 * time.Second)
			s, err := client.GetBatchStatus(batchJobID)
			if err != nil {
				return err
			}
			if s.Status == "completed" {
				fmt.Printf("[%d/%d completed] ", s.Completed, s.Total)
				return nil
			}
			if s.Status == "failed" {
				return fmt.Errorf("batch failed")
			}
		}
		return fmt.Errorf("timeout waiting for batch")
	})

	// ─── 10. Video ───
	fmt.Println("\n── POST /v1/video ──")
	test("Video capture", func() error {
		data, err := client.Video(snapapi.VideoOptions{
			URL:      "https://example.com",
			Duration: 3,
			Width:    1280,
			Height:   720,
		})
		if err != nil {
			return err
		}
		if len(data) < 1000 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		fmt.Printf("[%d bytes] ", len(data))
		return nil
	})

	test("Video with scroll", func() error {
		data, err := client.Video(snapapi.VideoOptions{
			URL:      "https://example.com",
			Duration: 3,
			Scroll:   true,
		})
		if err != nil {
			return err
		}
		if len(data) < 1000 {
			return fmt.Errorf("too small: %d bytes", len(data))
		}
		return nil
	})

	test("Video with JSON result", func() error {
		r, err := client.VideoWithResult(snapapi.VideoOptions{
			URL:      "https://example.com",
			Duration: 3,
		})
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		fmt.Printf("[%dx%d %dms] ", r.Width, r.Height, r.Took)
		return nil
	})

	// ─── 11. Extract ───
	fmt.Println("\n── POST /v1/extract ──")
	extractTypes := []string{"html", "text", "markdown", "article", "links", "images", "metadata", "structured"}
	for _, t := range extractTypes {
		extractType := t
		test(fmt.Sprintf("Extract type=%s", extractType), func() error {
			r, err := client.Extract(snapapi.ExtractOptions{
				URL:  "https://example.com",
				Type: extractType,
			})
			if err != nil {
				return err
			}
			if !r.Success {
				return fmt.Errorf("success=false")
			}
			// Show some info
			switch extractType {
			case "links":
				fmt.Printf("[%d links] ", len(r.Links))
			case "images":
				fmt.Printf("[%d images] ", len(r.Images))
			default:
				if r.Content != "" {
					cl := len(r.Content)
					if cl > 50 {
						cl = 50
					}
					fmt.Printf("[%d chars] ", len(r.Content))
				}
			}
			return nil
		})
	}

	// Convenience methods
	test("ExtractMarkdown convenience", func() error {
		r, err := client.ExtractMarkdown("https://example.com")
		if err != nil {
			return err
		}
		if !r.Success || r.Content == "" {
			return fmt.Errorf("empty content")
		}
		return nil
	})

	test("ExtractText convenience", func() error {
		r, err := client.ExtractText("https://example.com")
		if err != nil {
			return err
		}
		if !r.Success || r.Content == "" {
			return fmt.Errorf("empty content")
		}
		return nil
	})

	test("ExtractArticle convenience", func() error {
		r, err := client.ExtractArticle("https://example.com")
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		return nil
	})

	test("ExtractLinks convenience", func() error {
		r, err := client.ExtractLinks("https://example.com")
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		return nil
	})

	test("ExtractImages convenience", func() error {
		r, err := client.ExtractImages("https://example.com")
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		return nil
	})

	test("ExtractMetadata convenience", func() error {
		r, err := client.ExtractMetadata("https://example.com")
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		return nil
	})

	test("ExtractStructured convenience", func() error {
		r, err := client.ExtractStructured("https://example.com")
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		return nil
	})

	// ─── 12. Analyze ───
	fmt.Println("\n── POST /v1/analyze ──")
	test("Analyze webpage", func() error {
		r, err := client.Analyze(snapapi.AnalyzeOptions{
			URL:    "https://example.com",
			Prompt: "What is this page about? Answer in one sentence.",
		})
		if err != nil {
			return err
		}
		if !r.Success {
			return fmt.Errorf("success=false")
		}
		if r.Analysis == "" {
			return fmt.Errorf("empty analysis")
		}
		cl := len(r.Analysis)
		if cl > 80 {
			cl = 80
		}
		fmt.Printf("[%s...] ", r.Analysis[:cl])
		return nil
	})

	// ─── 13. Async screenshot ───
	fmt.Println("\n── Async Screenshot ──")
	test("Async screenshot + polling", func() error {
		r, err := client.ScreenshotAsync(snapapi.ScreenshotOptions{
			URL: "https://example.com",
		})
		if err != nil {
			// Async may not be supported — check
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
				return fmt.Errorf("SKIP: async not supported")
			}
			return err
		}
		if r.JobID == "" {
			// Might have returned synchronously
			fmt.Printf("[sync response] ")
			return nil
		}
		fmt.Printf("[jobId=%s] ", r.JobID)
		// Poll
		for i := 0; i < 15; i++ {
			time.Sleep(2 * time.Second)
			s, err := client.GetScreenshotStatus(r.JobID)
			if err != nil {
				return err
			}
			if s.Status == "completed" {
				fmt.Printf("[done] ")
				return nil
			}
			if s.Status == "failed" {
				return fmt.Errorf("async failed: %s", s.Error)
			}
		}
		return fmt.Errorf("timeout")
	})

	// ─── 14. Error handling ───
	fmt.Println("\n── Error Handling ──")
	test("Invalid API key", func() error {
		badClient := snapapi.NewClient("invalid-key")
		_, err := badClient.Screenshot(snapapi.ScreenshotOptions{URL: "https://example.com"})
		if err == nil {
			return fmt.Errorf("expected error")
		}
		apiErr, ok := err.(*snapapi.APIError)
		if !ok {
			return fmt.Errorf("expected APIError, got %T", err)
		}
		if apiErr.StatusCode != 401 {
			return fmt.Errorf("expected 401, got %d", apiErr.StatusCode)
		}
		return nil
	})

	test("Missing URL validation", func() error {
		_, err := client.Screenshot(snapapi.ScreenshotOptions{})
		if err == nil {
			return fmt.Errorf("expected error")
		}
		return nil
	})

	test("APIError.IsRetryable()", func() error {
		e := &snapapi.APIError{Code: "RATE_LIMITED", StatusCode: 429}
		if !e.IsRetryable() {
			return fmt.Errorf("rate limited should be retryable")
		}
		e2 := &snapapi.APIError{Code: "UNAUTHORIZED", StatusCode: 401}
		if e2.IsRetryable() {
			return fmt.Errorf("unauthorized should NOT be retryable")
		}
		return nil
	})

	// ─── Summary ───
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("RESULTS: ✅ %d passed  ❌ %d failed  ⏭️  %d skipped\n", passed, failed, skipped)
	fmt.Println(strings.Repeat("=", 80))

	// Dump any failure detail
	if failed > 0 {
		os.Exit(1)
	}

	// Save a sample output
	data, _ := json.MarshalIndent(map[string]int{
		"passed": passed, "failed": failed, "skipped": skipped,
	}, "", "  ")
	os.WriteFile("test-results.json", data, 0644)
}
