// Command screenshot demonstrates taking a full-page PNG screenshot with SnapAPI.
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
		snapapi.WithTimeout(45*time.Second),
		snapapi.WithRetries(3),
	)

	ctx := context.Background()

	img, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
		URL:      "https://example.com",
		Format:   "png",
		FullPage: true,
		Width:    1280,
		Height:   720,
	})
	if err != nil {
		var apiErr *snapapi.APIError
		if errors.As(err, &apiErr) {
			fmt.Fprintf(os.Stderr, "API error [%s]: %s\n", apiErr.Code, apiErr.Message)
			if apiErr.IsRateLimited() {
				fmt.Fprintf(os.Stderr, "Retry after %d seconds\n", apiErr.RetryAfter)
			}
		}
		log.Fatal(err)
	}

	if err := os.WriteFile("screenshot.png", img, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Saved screenshot.png (%d bytes)\n", len(img))

	// Check quota remaining
	quota, err := client.Quota(ctx)
	if err != nil {
		log.Printf("Warning: could not fetch quota: %v", err)
	} else {
		fmt.Printf("Quota: %d used / %d total (%d remaining)\n",
			quota.Used, quota.Total, quota.Remaining)
	}
}
