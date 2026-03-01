package main

import (
	"context"
	"fmt"
	"log"
	"os"

	snapapi "github.com/Sleywill/snapapi-go"
)

func main() {
	apiKey := os.Getenv("SNAPAPI_KEY")
	if apiKey == "" {
		log.Fatal("SNAPAPI_KEY environment variable is required")
	}

	client := snapapi.NewClient(apiKey)
	ctx := context.Background()

	// ── Screenshot ─────────────────────────────────────────────────────────────
	fmt.Println("Taking screenshot of example.com...")
	img, err := client.Screenshot(ctx, snapapi.ScreenshotOptions{
		URL:      "https://example.com",
		Format:   "png",
		FullPage: true,
		Width:    1280,
		Height:   720,
	})
	if err != nil {
		log.Fatalf("Screenshot failed: %v", err)
	}
	if err := os.WriteFile("screenshot.png", img, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Saved screenshot.png (%d bytes)\n", len(img))

	// ── Scrape ─────────────────────────────────────────────────────────────────
	fmt.Println("\nScraping example.com...")
	scrapeResult, err := client.Scrape(ctx, snapapi.ScrapeOptions{
		URL:  "https://example.com",
		Type: "text",
	})
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}
	for _, item := range scrapeResult.Results {
		fmt.Printf("Page %d: %d chars\n", item.Page, len(item.Data))
	}

	// ── Extract ────────────────────────────────────────────────────────────────
	fmt.Println("\nExtracting article from example.com...")
	extractResult, err := client.ExtractArticle(ctx, "https://example.com")
	if err != nil {
		log.Fatalf("Extract failed: %v", err)
	}
	fmt.Printf("Extracted (type=%s) in %dms\n", extractResult.Type, extractResult.ResponseTime)

	// ── List API Keys ──────────────────────────────────────────────────────────
	fmt.Println("\nListing API keys...")
	keys, err := client.Keys.List(ctx)
	if err != nil {
		log.Fatalf("Keys.List failed: %v", err)
	}
	for _, k := range keys.Keys {
		fmt.Printf("  Key: %s (id=%s)\n", k.Name, k.ID)
	}

	// ── Scheduled Screenshot ───────────────────────────────────────────────────
	fmt.Println("\nCreating scheduled job...")
	job, err := client.Scheduled.Create(ctx, snapapi.ScheduledOptions{
		URL:            "https://example.com",
		CronExpression: "0 * * * *", // every hour
		Format:         "png",
		FullPage:       true,
	})
	if err != nil {
		log.Fatalf("Scheduled.Create failed: %v", err)
	}
	fmt.Printf("Created job id=%s, next run: %s\n", job.ID, job.NextRunAt)

	// Clean up
	_ = client.Scheduled.Delete(ctx, job.ID)
	fmt.Println("\nDone!")
}
