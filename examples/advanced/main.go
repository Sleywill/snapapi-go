// Command advanced demonstrates real-world use cases with the SnapAPI Go SDK.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	snapapi "github.com/Sleywill/snapapi-go"
)

func main() {
	apiKey := os.Getenv("SNAPAPI_KEY")
	if apiKey == "" {
		log.Fatal("SNAPAPI_KEY environment variable is required")
	}

	client := snapapi.New(apiKey,
		snapapi.WithTimeout(60*time.Second),
		snapapi.WithRetries(3),
	)

	ctx := context.Background()

	// --- Use case 1: Website monitoring with scheduled screenshots ---
	fmt.Println("=== Website Monitoring ===")
	monitorURLs := []string{
		"https://example.com",
		"https://example.org",
	}
	for _, u := range monitorURLs {
		_, err := client.ScreenshotToFile(ctx, fmt.Sprintf("monitor_%s.png", sanitize(u)), snapapi.ScreenshotParams{
			URL:      u,
			Format:   "png",
			FullPage: true,
			Width:    1280,
		})
		if err != nil {
			log.Printf("Failed to capture %s: %v", u, err)
			continue
		}
		fmt.Printf("Captured: %s\n", u)
	}

	// --- Use case 2: SEO audit - extract content for analysis ---
	fmt.Println("\n=== SEO Content Extraction ===")
	content, err := client.Extract(ctx, snapapi.ExtractParams{
		URL:    "https://example.com",
		Format: "text",
	})
	if err != nil {
		log.Printf("Extract failed: %v", err)
	} else {
		fmt.Printf("Word count: %d\n", content.WordCount)
		if len(content.Content) > 200 {
			fmt.Printf("Preview: %s...\n", content.Content[:200])
		} else {
			fmt.Printf("Content: %s\n", content.Content)
		}
	}

	// --- Use case 3: Generate PDF report ---
	fmt.Println("\n=== PDF Report Generation ===")
	pdfBytes, err := client.PDF(ctx, snapapi.PDFParams{
		URL:    "https://example.com",
		Format: "a4",
		Margin: "10mm",
	})
	if err != nil {
		log.Printf("PDF generation failed: %v", err)
	} else {
		if err := os.WriteFile("report.pdf", pdfBytes, 0644); err != nil {
			log.Printf("Failed to write PDF: %v", err)
		} else {
			fmt.Printf("Saved report.pdf (%d bytes)\n", len(pdfBytes))
		}
	}

	// --- Use case 4: Social media thumbnail with custom dimensions ---
	fmt.Println("\n=== Social Media Thumbnail ===")
	_, err = client.ScreenshotToFile(ctx, "og_image.png", snapapi.ScreenshotParams{
		URL:    "https://example.com",
		Format: "png",
		Width:  1200,
		Height: 630,
		Clip: &snapapi.ClipRegion{
			X:      0,
			Y:      0,
			Width:  1200,
			Height: 630,
		},
	})
	if err != nil {
		log.Printf("Thumbnail capture failed: %v", err)
	} else {
		fmt.Println("Saved og_image.png (1200x630)")
	}

	// --- Use case 5: Competitor price tracking via scraping ---
	fmt.Println("\n=== Competitor Scraping ===")
	scrapeResult, err := client.Scrape(ctx, snapapi.ScrapeParams{
		URL:      "https://example.com",
		Selector: "body",
		Format:   "text",
	})
	if err != nil {
		var apiErr *snapapi.APIError
		if errors.As(err, &apiErr) {
			fmt.Printf("Scrape error [%s]: %s\n", apiErr.Code, apiErr.Message)
		} else {
			log.Printf("Scrape failed: %v", err)
		}
	} else {
		fmt.Printf("Scraped %s (status %d, %d chars)\n",
			scrapeResult.URL, scrapeResult.Status, len(scrapeResult.Data))
	}

	// --- Check remaining usage ---
	fmt.Println("\n=== Usage ===")
	usage, err := client.GetUsage(ctx)
	if err != nil {
		log.Printf("Could not fetch usage: %v", err)
	} else {
		fmt.Printf("API Usage: %d / %d (%d remaining)\n",
			usage.Used, usage.Total, usage.Remaining)
	}
}

func sanitize(url string) string {
	result := make([]byte, 0, len(url))
	for _, c := range []byte(url) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result = append(result, c)
		} else if c >= 'A' && c <= 'Z' {
			result = append(result, c+32)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}
