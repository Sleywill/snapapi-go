// Command analyze demonstrates using SnapAPI's LLM analysis endpoint.
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
		snapapi.WithRetries(2),
	)

	ctx := context.Background()

	result, err := client.Analyze(ctx, snapapi.AnalyzeParams{
		URL:      "https://example.com",
		Prompt:   "Summarize the main purpose of this website in 2-3 sentences.",
		Provider: "openai",
		APIKey:   os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		var apiErr *snapapi.APIError
		if errors.As(err, &apiErr) && apiErr.IsServiceUnavailable() {
			fmt.Println("Analyze endpoint is currently unavailable (LLM credits may be exhausted)")
			os.Exit(1)
		}
		log.Fatalf("Analyze failed: %v", err)
	}

	fmt.Printf("Analysis of %s:\n\n%s\n", result.URL, result.Result)
}
