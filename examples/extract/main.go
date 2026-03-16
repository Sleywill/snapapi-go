// Command extract demonstrates extracting Markdown content from a URL with SnapAPI.
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

	result, err := client.Extract(ctx, snapapi.ExtractParams{
		URL:    "https://example.com",
		Format: "markdown",
	})
	if err != nil {
		log.Fatalf("Extract failed: %v", err)
	}

	fmt.Printf("Extracted (%d words):\n\n%s\n", result.WordCount, result.Content)
}
