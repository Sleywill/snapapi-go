package snapapi

import (
	"context"
	"net/http"
)

// VideoParams holds parameters for video recording.
type VideoParams struct {
	// URL of the page to record. Required.
	URL string `json:"url"`
	// Duration in seconds (1–30). Default: 5.
	Duration int `json:"duration,omitempty"`
	// Format: "mp4" (default), "webm", or "gif".
	Format string `json:"format,omitempty"`
	// Width of the viewport in pixels (320–1920). Default: 1280.
	Width int `json:"width,omitempty"`
	// Height of the viewport in pixels (240–1080). Default: 720.
	Height int `json:"height,omitempty"`
	// FPS is the frames per second (10–30). Default: 25.
	FPS int `json:"fps,omitempty"`
	// Scrolling enables scroll-based video recording.
	Scrolling bool `json:"scrolling,omitempty"`
	// ScrollSpeed is the scroll speed in px/s (50–500). Default: 100.
	ScrollSpeed int `json:"scrollSpeed,omitempty"`
	// ScrollDelay is the delay before scrolling starts in ms. Default: 1000.
	ScrollDelay int `json:"scrollDelay,omitempty"`
	// ScrollDuration is the duration of each scroll step in ms. Default: 300.
	ScrollDuration int `json:"scrollDuration,omitempty"`
	// ScrollBy is the pixels scrolled per step. Default: 800.
	ScrollBy int `json:"scrollBy,omitempty"`
	// ScrollEasing is the easing function: "linear" (default), "ease_in", "ease_out", "ease_in_out", "ease_in_out_quint".
	ScrollEasing string `json:"scrollEasing,omitempty"`
	// ScrollBack scrolls back to the top at the end. Default: true (omitted means server default).
	ScrollBack *bool `json:"scrollBack,omitempty"`
	// ScrollComplete stops recording when scrolling finishes.
	ScrollComplete *bool `json:"scrollComplete,omitempty"`
	// DarkMode emulates dark color scheme.
	DarkMode bool `json:"darkMode,omitempty"`
	// BlockAds enables ad blocking.
	BlockAds bool `json:"blockAds,omitempty"`
	// BlockCookieBanners blocks cookie consent banners.
	BlockCookieBanners bool `json:"blockCookieBanners,omitempty"`
	// Delay is the time in ms to wait after page load before recording.
	Delay int `json:"delay,omitempty"`
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
