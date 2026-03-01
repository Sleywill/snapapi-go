package snapapi

// ─── Shared ───────────────────────────────────────────────────────────────────

// Cookie represents a browser cookie.
type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain,omitempty"`
	Path     string `json:"path,omitempty"`
	Expires  int64  `json:"expires,omitempty"`
	HTTPOnly bool   `json:"httpOnly,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	SameSite string `json:"sameSite,omitempty"`
}

// HTTPAuth represents HTTP basic authentication credentials.
type HTTPAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ProxyConfig represents proxy configuration.
type ProxyConfig struct {
	Server   string `json:"server"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Geolocation represents geolocation coordinates.
type Geolocation struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Accuracy  *float64 `json:"accuracy,omitempty"`
}

// PDFPageOptions represents PDF page layout options.
type PDFPageOptions struct {
	PageSize  string `json:"pageSize,omitempty"`
	Landscape bool   `json:"landscape,omitempty"`
	// Margins
	MarginTop    string `json:"marginTop,omitempty"`
	MarginRight  string `json:"marginRight,omitempty"`
	MarginBottom string `json:"marginBottom,omitempty"`
	MarginLeft   string `json:"marginLeft,omitempty"`
}

// StorageDestination holds storage destination options embedded in Screenshot requests.
type StorageDestination struct {
	Destination string `json:"destination,omitempty"` // e.g. "s3"
	Format      string `json:"format,omitempty"`
}

// ─── Screenshot ───────────────────────────────────────────────────────────────

// ScreenshotOptions represents options for the POST /v1/screenshot endpoint.
type ScreenshotOptions struct {
	// Source – one of URL, HTML, or Markdown must be set
	URL      string `json:"url,omitempty"`
	HTML     string `json:"html,omitempty"`
	Markdown string `json:"markdown,omitempty"`

	// Output format
	Format  string `json:"format,omitempty"`  // png|jpeg|webp|avif|pdf
	Quality *int   `json:"quality,omitempty"` // 0-100

	// Viewport
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Device string `json:"device,omitempty"`

	// Page behaviour
	FullPage        bool   `json:"fullPage,omitempty"`
	Selector        string `json:"selector,omitempty"`
	Delay           int    `json:"delay,omitempty"`
	Timeout         int    `json:"timeout,omitempty"`
	WaitUntil       string `json:"waitUntil,omitempty"`
	WaitForSelector string `json:"waitForSelector,omitempty"`

	// Visual
	DarkMode bool `json:"darkMode,omitempty"`

	// Scripting
	CSS           string   `json:"css,omitempty"`
	JavaScript    string   `json:"javascript,omitempty"`
	HideSelectors []string `json:"hideSelectors,omitempty"`
	ClickSelector string   `json:"clickSelector,omitempty"`

	// Blocking
	BlockAds           bool `json:"blockAds,omitempty"`
	BlockTrackers      bool `json:"blockTrackers,omitempty"`
	BlockCookieBanners bool `json:"blockCookieBanners,omitempty"`

	// Identity / auth
	UserAgent    string            `json:"userAgent,omitempty"`
	ExtraHeaders map[string]string `json:"extraHeaders,omitempty"`
	Cookies      []Cookie          `json:"cookies,omitempty"`
	HTTPAuth     *HTTPAuth         `json:"httpAuth,omitempty"`

	// Proxy
	Proxy        string `json:"proxy,omitempty"`
	PremiumProxy bool   `json:"premiumProxy,omitempty"`

	// Environment
	Geolocation *Geolocation `json:"geolocation,omitempty"`
	Timezone    string       `json:"timezone,omitempty"`

	// PDF
	PDF *PDFPageOptions `json:"pdf,omitempty"`

	// Storage (returns JSON with id+url instead of binary)
	Storage *StorageDestination `json:"storage,omitempty"`

	// Webhook
	WebhookURL string `json:"webhookUrl,omitempty"`
}

// StorageUploadResult is returned when Storage options are set.
type StorageUploadResult struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// ─── Scrape ──────────────────────────────────────────────────────────────────

// ScrapeOptions represents options for the POST /v1/scrape endpoint.
type ScrapeOptions struct {
	URL             string `json:"url"`
	Type            string `json:"type,omitempty"` // text|html|links
	Pages           int    `json:"pages,omitempty"`
	WaitMs          int    `json:"waitMs,omitempty"`
	Proxy           string `json:"proxy,omitempty"`
	PremiumProxy    bool   `json:"premiumProxy,omitempty"`
	BlockResources  bool   `json:"blockResources,omitempty"`
	Locale          string `json:"locale,omitempty"`
}

// ScrapeResult is the response from /v1/scrape.
type ScrapeResult struct {
	Success  bool          `json:"success"`
	Results  []ScrapeItem  `json:"results"`
}

// ScrapeItem is a single scraped page.
type ScrapeItem struct {
	Page int    `json:"page"`
	URL  string `json:"url"`
	Data string `json:"data"`
}

// ─── Extract ─────────────────────────────────────────────────────────────────

// ExtractOptions represents options for the POST /v1/extract endpoint.
type ExtractOptions struct {
	URL                string `json:"url"`
	Type               string `json:"type,omitempty"` // html|text|markdown|article|links|images|metadata|structured
	Selector           string `json:"selector,omitempty"`
	WaitFor            string `json:"waitFor,omitempty"`
	Timeout            int    `json:"timeout,omitempty"`
	DarkMode           bool   `json:"darkMode,omitempty"`
	BlockAds           bool   `json:"blockAds,omitempty"`
	BlockCookieBanners bool   `json:"blockCookieBanners,omitempty"`
	IncludeImages      bool   `json:"includeImages,omitempty"`
	MaxLength          *int   `json:"maxLength,omitempty"`
}

// ExtractResult is the response from /v1/extract.
type ExtractResult struct {
	Success      bool        `json:"success"`
	Type         string      `json:"type"`
	URL          string      `json:"url"`
	Data         interface{} `json:"data"` // string or structured object depending on type
	ResponseTime int         `json:"responseTime"`
}

// ─── Analyze ─────────────────────────────────────────────────────────────────

// AnalyzeOptions represents options for the POST /v1/analyze endpoint.
type AnalyzeOptions struct {
	URL               string `json:"url"`
	Prompt            string `json:"prompt,omitempty"`
	Provider          string `json:"provider,omitempty"` // openai|anthropic
	APIKey            string `json:"apiKey,omitempty"`
	Model             string `json:"model,omitempty"`
	JSONSchema        string `json:"jsonSchema,omitempty"`
	IncludeScreenshot bool   `json:"includeScreenshot,omitempty"`
	IncludeMetadata   bool   `json:"includeMetadata,omitempty"`
	MaxContentLength  *int   `json:"maxContentLength,omitempty"`
}

// AnalyzeResult is the response from /v1/analyze.
type AnalyzeResult struct {
	Success      bool                   `json:"success"`
	URL          string                 `json:"url"`
	Metadata     map[string]interface{} `json:"metadata"`
	Analysis     interface{}            `json:"analysis"` // string or structured object
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
	ResponseTime int                    `json:"responseTime"`
}

// ─── Storage ──────────────────────────────────────────────────────────────────

// StorageFile represents a file in storage.
type StorageFile struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	Format    string `json:"format"`
	CreatedAt string `json:"createdAt"`
}

// StorageFilesResult is the response from GET /v1/storage/files.
type StorageFilesResult struct {
	Success bool          `json:"success"`
	Files   []StorageFile `json:"files"`
	Total   int           `json:"total"`
}

// StorageUsageResult is the response from GET /v1/storage/usage.
type StorageUsageResult struct {
	Success bool  `json:"success"`
	Used    int64 `json:"used"`
	Limit   int64 `json:"limit"`
}

// S3Config is the body for POST /v1/storage/s3.
type S3Config struct {
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Endpoint        string `json:"endpoint,omitempty"`
	PublicURL       string `json:"publicUrl,omitempty"`
}

// ─── Scheduled ───────────────────────────────────────────────────────────────

// ScheduledOptions is the body for POST /v1/scheduled.
type ScheduledOptions struct {
	URL            string `json:"url"`
	CronExpression string `json:"cronExpression"`
	Format         string `json:"format,omitempty"`
	Width          int    `json:"width,omitempty"`
	Height         int    `json:"height,omitempty"`
	FullPage       bool   `json:"fullPage,omitempty"`
	WebhookURL     string `json:"webhookUrl,omitempty"`
}

// ScheduledJob is a single scheduled job.
type ScheduledJob struct {
	ID             string `json:"id"`
	URL            string `json:"url"`
	CronExpression string `json:"cronExpression"`
	Format         string `json:"format"`
	Active         bool   `json:"active"`
	CreatedAt      string `json:"createdAt"`
	LastRunAt      string `json:"lastRunAt,omitempty"`
	NextRunAt      string `json:"nextRunAt,omitempty"`
}

// ScheduledListResult is the response from GET /v1/scheduled.
type ScheduledListResult struct {
	Success bool           `json:"success"`
	Jobs    []ScheduledJob `json:"jobs"`
}

// ─── Webhooks ─────────────────────────────────────────────────────────────────

// WebhookOptions is the body for POST /v1/webhooks.
type WebhookOptions struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Secret string   `json:"secret,omitempty"`
}

// Webhook represents a registered webhook.
type Webhook struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
	CreatedAt string   `json:"createdAt"`
}

// WebhooksListResult is the response from GET /v1/webhooks.
type WebhooksListResult struct {
	Success  bool      `json:"success"`
	Webhooks []Webhook `json:"webhooks"`
}

// ─── Keys ─────────────────────────────────────────────────────────────────────

// APIKey represents an API key.
type APIKey struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Key       string `json:"key,omitempty"` // only present on creation
	CreatedAt string `json:"createdAt"`
}

// KeysListResult is the response from GET /v1/keys.
type KeysListResult struct {
	Success bool     `json:"success"`
	Keys    []APIKey `json:"keys"`
}
