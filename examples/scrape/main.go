// Command scrape demonstrates scraping a web page with SnapAPI.
package main

import (
	"context"
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
		snapapi.WithTimeout(30*time.Second),
		snapapi.WithRetries(3),
	)

	ctx := context.Background()

	result, err := client.Scrape(ctx, snapapi.ScrapeParams{
		URL:    "https://example.com",
		Format: "html",
	})
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}

	fmt.Printf("Scraped URL: %s (status %d)\n", result.URL, result.Status)
	fmt.Printf("Data (%d chars):\n%s\n", len(result.Data), result.Data)
}
