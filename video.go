package snapapi

import (
	"context"
	"net/http"
)

// VideoParams holds parameters for video recording.
type VideoParams struct {
	// URL of the page to record. Required.
	URL string `json:"url"`
	// Duration in seconds. Default: 5.
	Duration int `json:"duration,omitempty"`
	// Format: "webm" (default), "mp4", or "gif".
	Format string `json:"format,omitempty"`
	// Width of the viewport in pixels.
	Width int `json:"width,omitempty"`
	// Height of the viewport in pixels.
	Height int `json:"height,omitempty"`
	// ScrollVideo enables scroll-based video recording (scrolls through the page).
	ScrollVideo bool `json:"scrollVideo,omitempty"`
}

// Video records a short video of a URL.
//
//	videoBytes, err := client.Video(ctx, snapapi.VideoParams{URL: "https://example.com"})
func (c *Client) Video(ctx context.Context, p VideoParams) ([]byte, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	return c.doRaw(ctx, http.MethodPost, "/v1/video", p)
}
