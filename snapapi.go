// Package snapapi provides an idiomatic Go client for the SnapAPI.pics service.
//
// SnapAPI lets you capture screenshots, generate PDFs, scrape web pages,
// extract structured content, and analyze pages with LLMs -- all via a simple
// HTTP API.
//
// # Quick start
//
//	client := snapapi.New("sk_your_key",
//	    snapapi.WithTimeout(30*time.Second),
//	    snapapi.WithRetries(3),
//	)
//
//	img, err := client.Screenshot(ctx, snapapi.ScreenshotParams{
//	    URL:      "https://example.com",
//	    Format:   "png",
//	    FullPage: true,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.WriteFile("screenshot.png", img, 0644)
//
// # Namespaces
//
// The client exposes sub-namespaces for grouping related endpoints:
//
//	client.Storage    -- manage stored captures
//	client.Scheduled  -- schedule recurring captures
//	client.Webhooks   -- manage webhook endpoints
//	client.APIKeys    -- manage API keys for your account
//
// # Error handling
//
// All methods return a typed *APIError on failure. Use errors.As to inspect it:
//
//	var apiErr *snapapi.APIError
//	if errors.As(err, &apiErr) {
//	    fmt.Println(apiErr.Code, apiErr.StatusCode)
//	}
package snapapi

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Client is the SnapAPI client. Create one with New().
//
// The client is safe for concurrent use by multiple goroutines.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	retries    int
	retryDelay time.Duration

	// Sub-namespace accessors. Populated by New().
	Storage   *StorageNamespace
	Scheduled *ScheduledNamespace
	Webhooks  *WebhooksNamespace
	APIKeys   *APIKeysNamespace
}

// New creates a new SnapAPI client with the given API key.
//
//	client := snapapi.New("sk_...",
//	    snapapi.WithTimeout(45*time.Second),
//	    snapapi.WithRetries(3),
//	)
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		retries:    defaultRetries,
		retryDelay: 500 * time.Millisecond,
	}
	for _, o := range opts {
		o(c)
	}
	// Wire up namespace accessors.
	c.Storage = &StorageNamespace{c: c}
	c.Scheduled = &ScheduledNamespace{c: c}
	c.Webhooks = &WebhooksNamespace{c: c}
	c.APIKeys = &APIKeysNamespace{c: c}
	return c
}

// NewClient is an alias for New, kept for backward compatibility.
func NewClient(apiKey string, opts ...Option) *Client {
	return New(apiKey, opts...)
}

// ---------------------------------------------------------------------------
// Analyze
// ---------------------------------------------------------------------------

// AnalyzeParams holds all parameters for the Analyze endpoint.
type AnalyzeParams struct {
	// URL of the page to analyze. Required.
	URL string `json:"url"`
	// Prompt is the instruction for the LLM (e.g. "Summarize this page").
	Prompt string `json:"prompt,omitempty"`
	// Provider is the LLM provider: "openai", "anthropic", or "google".
	Provider string `json:"provider,omitempty"`
	// APIKey is the LLM provider API key.
	APIKey string `json:"apiKey,omitempty"`
	// JSONSchema constrains the LLM output to match a JSON schema.
	JSONSchema map[string]interface{} `json:"jsonSchema,omitempty"`
}

// AnalyzeResult is the structured response from the analyze endpoint.
type AnalyzeResult struct {
	// Result is the LLM's analysis output.
	Result string `json:"result"`
	// URL is the analyzed URL.
	URL string `json:"url"`
}

// Analyze sends a page to an LLM for analysis.
// Note: This endpoint may return HTTP 503 if LLM credits are exhausted.
//
//	result, err := client.Analyze(ctx, snapapi.AnalyzeParams{
//	    URL:      "https://example.com",
//	    Prompt:   "Summarize this page in 3 sentences.",
//	    Provider: "openai",
//	    APIKey:   "sk-...",
//	})
//	fmt.Println(result.Result)
func (c *Client) Analyze(ctx context.Context, p AnalyzeParams) (*AnalyzeResult, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	var result AnalyzeResult
	if err := c.doJSON(ctx, http.MethodPost, "/v1/analyze", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Usage
// ---------------------------------------------------------------------------

// UsageResult is the response from GET /v1/usage.
type UsageResult struct {
	Used      int    `json:"used"`
	Limit     int    `json:"limit"`
	Total     int    `json:"total"`
	Remaining int    `json:"remaining"`
	ResetAt   string `json:"resetAt,omitempty"`
}

// GetUsage returns the caller's current API usage statistics.
//
//	usage, err := client.GetUsage(ctx)
//	fmt.Printf("Used: %d / %d\n", usage.Used, usage.Limit)
func (c *Client) GetUsage(ctx context.Context) (*UsageResult, error) {
	var result UsageResult
	if err := c.doJSON(ctx, http.MethodGet, "/v1/usage", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Quota is an alias for GetUsage, kept for backward compatibility.
func (c *Client) Quota(ctx context.Context) (*UsageResult, error) {
	return c.GetUsage(ctx)
}

// ---------------------------------------------------------------------------
// Ping
// ---------------------------------------------------------------------------

// PingResult is the response from GET /v1/ping.
type PingResult struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

// Ping checks the API health.
//
//	result, err := client.Ping(ctx)
//	fmt.Println(result.Status)
func (c *Client) Ping(ctx context.Context) (*PingResult, error) {
	var result PingResult
	if err := c.doJSON(ctx, http.MethodGet, "/v1/ping", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Storage namespace
// ---------------------------------------------------------------------------

// StorageNamespace groups all Storage-related API methods.
// Access it via client.Storage.
type StorageNamespace struct{ c *Client }

// StorageItem represents a single item in cloud storage.
type StorageItem struct {
	// Key is the object key / path.
	Key string `json:"key"`
	// URL is the public access URL.
	URL string `json:"url"`
	// Size is the file size in bytes.
	Size int64 `json:"size"`
	// ContentType is the MIME type.
	ContentType string `json:"content_type"`
	// CreatedAt is the ISO 8601 creation timestamp.
	CreatedAt string `json:"created_at"`
}

// StorageListResult is the paginated response from Storage.List.
type StorageListResult struct {
	Items   []StorageItem `json:"items"`
	Total   int           `json:"total"`
	Page    int           `json:"page"`
	PerPage int           `json:"per_page"`
	HasMore bool          `json:"has_more"`
}

// StorageListParams are query parameters for Storage.List.
type StorageListParams struct {
	// Page is the 1-based page number. Default: 1.
	Page int `json:"page,omitempty"`
	// PerPage is the number of items per page. Default: 20, max: 100.
	PerPage int `json:"per_page,omitempty"`
	// Prefix filters results to keys with this prefix.
	Prefix string `json:"prefix,omitempty"`
}

// List returns a paginated list of stored captures.
//
//	items, err := client.Storage.List(ctx, snapapi.StorageListParams{PerPage: 50})
func (s *StorageNamespace) List(ctx context.Context, p StorageListParams) (*StorageListResult, error) {
	path := "/v1/storage"
	sep := "?"
	if p.Page > 0 {
		path += sep + fmt.Sprintf("page=%d", p.Page)
		sep = "&"
	}
	if p.PerPage > 0 {
		path += sep + fmt.Sprintf("per_page=%d", p.PerPage)
		sep = "&"
	}
	if p.Prefix != "" {
		path += sep + "prefix=" + p.Prefix
	}
	var result StorageListResult
	if err := s.c.doJSON(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Get returns metadata for a single stored capture by key.
//
//	item, err := client.Storage.Get(ctx, "reports/home.png")
func (s *StorageNamespace) Get(ctx context.Context, key string) (*StorageItem, error) {
	if key == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "key is required", StatusCode: 400}
	}
	var result StorageItem
	if err := s.c.doJSON(ctx, http.MethodGet, "/v1/storage/"+key, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a stored capture by key.
//
//	err := client.Storage.Delete(ctx, "reports/home.png")
func (s *StorageNamespace) Delete(ctx context.Context, key string) error {
	if key == "" {
		return &APIError{Code: ErrInvalidParams, Message: "key is required", StatusCode: 400}
	}
	_, err := s.c.doRaw(ctx, http.MethodDelete, "/v1/storage/"+key, nil)
	return err
}

// ---------------------------------------------------------------------------
// Scheduled namespace
// ---------------------------------------------------------------------------

// ScheduledNamespace groups all Scheduled-capture API methods.
// Access it via client.Scheduled.
type ScheduledNamespace struct{ c *Client }

// Schedule represents a recurring scheduled capture.
type Schedule struct {
	// ID is the unique schedule identifier.
	ID string `json:"id"`
	// URL is the page to capture on each run.
	URL string `json:"url"`
	// Cron is the cron expression for the schedule (e.g. "0 9 * * 1-5").
	Cron string `json:"cron"`
	// Params holds the screenshot parameters used for each capture.
	Params map[string]interface{} `json:"params,omitempty"`
	// Active indicates whether the schedule is currently running.
	Active bool `json:"active"`
	// CreatedAt is the ISO 8601 creation timestamp.
	CreatedAt string `json:"created_at"`
	// LastRunAt is the ISO 8601 timestamp of the most recent run, if any.
	LastRunAt string `json:"last_run_at,omitempty"`
	// NextRunAt is the ISO 8601 timestamp of the next scheduled run.
	NextRunAt string `json:"next_run_at,omitempty"`
}

// CreateScheduleParams are the parameters for Scheduled.Create.
type CreateScheduleParams struct {
	// URL of the page to capture. Required.
	URL string `json:"url"`
	// Cron is the cron expression for the schedule. Required.
	Cron string `json:"cron"`
	// Params holds additional screenshot options applied on each run.
	Params map[string]interface{} `json:"params,omitempty"`
}

// Create creates a new recurring schedule.
//
//	sched, err := client.Scheduled.Create(ctx, snapapi.CreateScheduleParams{
//	    URL:  "https://example.com",
//	    Cron: "0 9 * * 1-5",
//	})
func (s *ScheduledNamespace) Create(ctx context.Context, p CreateScheduleParams) (*Schedule, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	if p.Cron == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "cron expression is required", StatusCode: 400}
	}
	var result Schedule
	if err := s.c.doJSON(ctx, http.MethodPost, "/v1/scheduled", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns all schedules for the authenticated account.
//
//	schedules, err := client.Scheduled.List(ctx)
func (s *ScheduledNamespace) List(ctx context.Context) ([]Schedule, error) {
	var result []Schedule
	if err := s.c.doJSON(ctx, http.MethodGet, "/v1/scheduled", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Get returns a single schedule by ID.
//
//	sched, err := client.Scheduled.Get(ctx, "sched_abc123")
func (s *ScheduledNamespace) Get(ctx context.Context, id string) (*Schedule, error) {
	if id == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "id is required", StatusCode: 400}
	}
	var result Schedule
	if err := s.c.doJSON(ctx, http.MethodGet, "/v1/scheduled/"+id, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete deletes a schedule by ID.
//
//	err := client.Scheduled.Delete(ctx, "sched_abc123")
func (s *ScheduledNamespace) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &APIError{Code: ErrInvalidParams, Message: "id is required", StatusCode: 400}
	}
	_, err := s.c.doRaw(ctx, http.MethodDelete, "/v1/scheduled/"+id, nil)
	return err
}

// Pause disables a schedule without deleting it.
//
//	err := client.Scheduled.Pause(ctx, "sched_abc123")
func (s *ScheduledNamespace) Pause(ctx context.Context, id string) error {
	if id == "" {
		return &APIError{Code: ErrInvalidParams, Message: "id is required", StatusCode: 400}
	}
	_, err := s.c.doRaw(ctx, http.MethodPost, "/v1/scheduled/"+id+"/pause", nil)
	return err
}

// Resume re-enables a paused schedule.
//
//	err := client.Scheduled.Resume(ctx, "sched_abc123")
func (s *ScheduledNamespace) Resume(ctx context.Context, id string) error {
	if id == "" {
		return &APIError{Code: ErrInvalidParams, Message: "id is required", StatusCode: 400}
	}
	_, err := s.c.doRaw(ctx, http.MethodPost, "/v1/scheduled/"+id+"/resume", nil)
	return err
}

// ---------------------------------------------------------------------------
// Webhooks namespace
// ---------------------------------------------------------------------------

// WebhooksNamespace groups all Webhook management API methods.
// Access it via client.Webhooks.
type WebhooksNamespace struct{ c *Client }

// Webhook represents a webhook endpoint registration.
type Webhook struct {
	// ID is the unique webhook identifier.
	ID string `json:"id"`
	// URL is the endpoint that receives event POST requests.
	URL string `json:"url"`
	// Events is the list of event types this webhook receives.
	Events []string `json:"events"`
	// Active indicates whether the webhook is enabled.
	Active bool `json:"active"`
	// Secret is the signing secret used to verify delivery (write-only on create).
	Secret string `json:"secret,omitempty"`
	// CreatedAt is the ISO 8601 creation timestamp.
	CreatedAt string `json:"created_at"`
}

// CreateWebhookParams are the parameters for Webhooks.Create.
type CreateWebhookParams struct {
	// URL is the HTTPS endpoint to deliver events to. Required.
	URL string `json:"url"`
	// Events is the list of event types to subscribe to.
	// Example: []string{"screenshot.completed", "schedule.run.failed"}
	Events []string `json:"events"`
	// Secret is an optional signing secret for HMAC verification.
	Secret string `json:"secret,omitempty"`
}

// Create registers a new webhook endpoint.
//
//	wh, err := client.Webhooks.Create(ctx, snapapi.CreateWebhookParams{
//	    URL:    "https://myapp.com/hooks/snapapi",
//	    Events: []string{"screenshot.completed"},
//	})
func (w *WebhooksNamespace) Create(ctx context.Context, p CreateWebhookParams) (*Webhook, error) {
	if p.URL == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "URL is required", StatusCode: 400}
	}
	var result Webhook
	if err := w.c.doJSON(ctx, http.MethodPost, "/v1/webhooks", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns all registered webhooks for the account.
//
//	hooks, err := client.Webhooks.List(ctx)
func (w *WebhooksNamespace) List(ctx context.Context) ([]Webhook, error) {
	var result []Webhook
	if err := w.c.doJSON(ctx, http.MethodGet, "/v1/webhooks", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Get returns a single webhook by ID.
//
//	hook, err := client.Webhooks.Get(ctx, "wh_abc123")
func (w *WebhooksNamespace) Get(ctx context.Context, id string) (*Webhook, error) {
	if id == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "id is required", StatusCode: 400}
	}
	var result Webhook
	if err := w.c.doJSON(ctx, http.MethodGet, "/v1/webhooks/"+id, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Delete removes a webhook registration.
//
//	err := client.Webhooks.Delete(ctx, "wh_abc123")
func (w *WebhooksNamespace) Delete(ctx context.Context, id string) error {
	if id == "" {
		return &APIError{Code: ErrInvalidParams, Message: "id is required", StatusCode: 400}
	}
	_, err := w.c.doRaw(ctx, http.MethodDelete, "/v1/webhooks/"+id, nil)
	return err
}

// ---------------------------------------------------------------------------
// APIKeys namespace
// ---------------------------------------------------------------------------

// APIKeysNamespace groups all API-key management methods.
// Access it via client.APIKeys.
type APIKeysNamespace struct{ c *Client }

// APIKey represents one API key belonging to the account.
type APIKey struct {
	// ID is the unique key identifier.
	ID string `json:"id"`
	// Name is a human-readable label for the key.
	Name string `json:"name"`
	// Key is the raw API key string (only returned on creation).
	Key string `json:"key,omitempty"`
	// CreatedAt is the ISO 8601 creation timestamp.
	CreatedAt string `json:"created_at"`
	// LastUsedAt is the ISO 8601 timestamp of the most recent authenticated
	// request made with this key, or empty if never used.
	LastUsedAt string `json:"last_used_at,omitempty"`
	// ExpiresAt is the optional ISO 8601 expiry timestamp.
	ExpiresAt string `json:"expires_at,omitempty"`
}

// CreateAPIKeyParams are the parameters for APIKeys.Create.
type CreateAPIKeyParams struct {
	// Name is a human-readable label for the key. Required.
	Name string `json:"name"`
	// ExpiresAt is an optional ISO 8601 expiry timestamp.
	ExpiresAt string `json:"expires_at,omitempty"`
}

// Create creates a new API key.
// The raw key string is only returned once; store it securely.
//
//	key, err := client.APIKeys.Create(ctx, snapapi.CreateAPIKeyParams{Name: "CI pipeline"})
//	fmt.Println(key.Key) // store this value -- it will not be shown again
func (a *APIKeysNamespace) Create(ctx context.Context, p CreateAPIKeyParams) (*APIKey, error) {
	if p.Name == "" {
		return nil, &APIError{Code: ErrInvalidParams, Message: "name is required", StatusCode: 400}
	}
	var result APIKey
	if err := a.c.doJSON(ctx, http.MethodPost, "/v1/api-keys", p, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// List returns all API keys for the authenticated account (raw key values
// are not included; only metadata).
//
//	keys, err := client.APIKeys.List(ctx)
func (a *APIKeysNamespace) List(ctx context.Context) ([]APIKey, error) {
	var result []APIKey
	if err := a.c.doJSON(ctx, http.MethodGet, "/v1/api-keys", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Revoke permanently deletes an API key by ID.
//
//	err := client.APIKeys.Revoke(ctx, "key_abc123")
func (a *APIKeysNamespace) Revoke(ctx context.Context, id string) error {
	if id == "" {
		return &APIError{Code: ErrInvalidParams, Message: "id is required", StatusCode: 400}
	}
	_, err := a.c.doRaw(ctx, http.MethodDelete, "/v1/api-keys/"+id, nil)
	return err
}
