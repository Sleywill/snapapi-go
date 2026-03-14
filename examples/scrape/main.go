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
		URL: "https://example.com",
	})
	if err != nil {
		log.Fatalf("Scrape failed: %v", err)
	}

	fmt.Printf("Scraped URL: %s\n", result.URL)
	fmt.Printf("Text (%d chars):\n%s\n", len(result.Text), result.Text)
}
