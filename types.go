package snapapi

// DevicePreset represents a device preset name.
type DevicePreset string

// Available device presets
const (
	DeviceDesktop1080p       DevicePreset = "desktop-1080p"
	DeviceDesktop1440p       DevicePreset = "desktop-1440p"
	DeviceDesktop4K          DevicePreset = "desktop-4k"
	DeviceMacBookPro13       DevicePreset = "macbook-pro-13"
	DeviceMacBookPro16       DevicePreset = "macbook-pro-16"
	DeviceIMac24             DevicePreset = "imac-24"
	DeviceIPhoneSE           DevicePreset = "iphone-se"
	DeviceIPhone12           DevicePreset = "iphone-12"
	DeviceIPhone13           DevicePreset = "iphone-13"
	DeviceIPhone14           DevicePreset = "iphone-14"
	DeviceIPhone14Pro        DevicePreset = "iphone-14-pro"
	DeviceIPhone15           DevicePreset = "iphone-15"
	DeviceIPhone15Pro        DevicePreset = "iphone-15-pro"
	DeviceIPhone15ProMax     DevicePreset = "iphone-15-pro-max"
	DeviceIPad               DevicePreset = "ipad"
	DeviceIPadMini           DevicePreset = "ipad-mini"
	DeviceIPadAir            DevicePreset = "ipad-air"
	DeviceIPadPro11          DevicePreset = "ipad-pro-11"
	DeviceIPadPro129         DevicePreset = "ipad-pro-12.9"
	DevicePixel7             DevicePreset = "pixel-7"
	DevicePixel8             DevicePreset = "pixel-8"
	DevicePixel8Pro          DevicePreset = "pixel-8-pro"
	DeviceSamsungGalaxyS23   DevicePreset = "samsung-galaxy-s23"
	DeviceSamsungGalaxyS24   DevicePreset = "samsung-galaxy-s24"
	DeviceSamsungGalaxyTabS9 DevicePreset = "samsung-galaxy-tab-s9"
)

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
	Server   string   `json:"server"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	Bypass   []string `json:"bypass,omitempty"`
}

// Geolocation represents geolocation coordinates.
type Geolocation struct {
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Accuracy  *float64 `json:"accuracy,omitempty"`
}

// PDFOptions represents PDF generation options.
type PDFOptions struct {
	PageSize            string   `json:"pageSize,omitempty"`
	Width               string   `json:"width,omitempty"`
	Height              string   `json:"height,omitempty"`
	Landscape           *bool    `json:"landscape,omitempty"`
	MarginTop           string   `json:"marginTop,omitempty"`
	MarginRight         string   `json:"marginRight,omitempty"`
	MarginBottom        string   `json:"marginBottom,omitempty"`
	MarginLeft          string   `json:"marginLeft,omitempty"`
	PrintBackground     *bool    `json:"printBackground,omitempty"`
	HeaderTemplate      string   `json:"headerTemplate,omitempty"`
	FooterTemplate      string   `json:"footerTemplate,omitempty"`
	DisplayHeaderFooter *bool    `json:"displayHeaderFooter,omitempty"`
	Scale               *float64 `json:"scale,omitempty"`
	PageRanges          string   `json:"pageRanges,omitempty"`
	PreferCSSPageSize   *bool    `json:"preferCSSPageSize,omitempty"`
}

// ThumbnailOptions represents thumbnail generation options.
type ThumbnailOptions struct {
	Enabled bool   `json:"enabled"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	Fit     string `json:"fit,omitempty"`
}

// ExtractMetadata represents options for additional metadata extraction.
type ExtractMetadata struct {
	Fonts          *bool `json:"fonts,omitempty"`
	Colors         *bool `json:"colors,omitempty"`
	Links          *bool `json:"links,omitempty"`
	HTTPStatusCode *bool `json:"httpStatusCode,omitempty"`
}

// ScreenshotOptions represents options for taking a screenshot.
type ScreenshotOptions struct {
	URL                    string            `json:"url,omitempty"`
	HTML                   string            `json:"html,omitempty"`
	Format                 string            `json:"format,omitempty"`
	Quality                *int              `json:"quality,omitempty"`
	Device                 DevicePreset      `json:"device,omitempty"`
	Width                  int               `json:"width,omitempty"`
	Height                 int               `json:"height,omitempty"`
	DeviceScaleFactor      *float64          `json:"deviceScaleFactor,omitempty"`
	IsMobile               bool              `json:"isMobile,omitempty"`
	HasTouch               bool              `json:"hasTouch,omitempty"`
	IsLandscape            bool              `json:"isLandscape,omitempty"`
	FullPage               bool              `json:"fullPage,omitempty"`
	FullPageScrollDelay    *int              `json:"fullPageScrollDelay,omitempty"`
	FullPageMaxHeight      *int              `json:"fullPageMaxHeight,omitempty"`
	Selector               string            `json:"selector,omitempty"`
	SelectorScrollIntoView *bool             `json:"selectorScrollIntoView,omitempty"`
	ClipX                  *int              `json:"clipX,omitempty"`
	ClipY                  *int              `json:"clipY,omitempty"`
	ClipWidth              *int              `json:"clipWidth,omitempty"`
	ClipHeight             *int              `json:"clipHeight,omitempty"`
	Delay                  int               `json:"delay,omitempty"`
	Timeout                int               `json:"timeout,omitempty"`
	WaitUntil              string            `json:"waitUntil,omitempty"`
	WaitForSelector        string            `json:"waitForSelector,omitempty"`
	WaitForSelectorTimeout *int              `json:"waitForSelectorTimeout,omitempty"`
	DarkMode               bool              `json:"darkMode,omitempty"`
	ReducedMotion          bool              `json:"reducedMotion,omitempty"`
	CSS                    string            `json:"css,omitempty"`
	JavaScript             string            `json:"javascript,omitempty"`
	HideSelectors          []string          `json:"hideSelectors,omitempty"`
	ClickSelector          string            `json:"clickSelector,omitempty"`
	ClickDelay             *int              `json:"clickDelay,omitempty"`
	BlockAds               bool              `json:"blockAds,omitempty"`
	BlockTrackers          bool              `json:"blockTrackers,omitempty"`
	BlockCookieBanners     bool              `json:"blockCookieBanners,omitempty"`
	BlockChatWidgets       bool              `json:"blockChatWidgets,omitempty"`
	BlockResources         []string          `json:"blockResources,omitempty"`
	UserAgent              string            `json:"userAgent,omitempty"`
	ExtraHeaders           map[string]string `json:"extraHeaders,omitempty"`
	Cookies                []Cookie          `json:"cookies,omitempty"`
	HTTPAuth               *HTTPAuth         `json:"httpAuth,omitempty"`
	Proxy                  *ProxyConfig      `json:"proxy,omitempty"`
	Geolocation            *Geolocation      `json:"geolocation,omitempty"`
	Timezone               string            `json:"timezone,omitempty"`
	Locale                 string            `json:"locale,omitempty"`
	PDFOptions             *PDFOptions       `json:"pdfOptions,omitempty"`
	Thumbnail              *ThumbnailOptions `json:"thumbnail,omitempty"`
	FailOnHTTPError        bool              `json:"failOnHttpError,omitempty"`
	Cache                  bool              `json:"cache,omitempty"`
	CacheTTL               *int              `json:"cacheTtl,omitempty"`
	ResponseType           string            `json:"responseType,omitempty"`
	IncludeMetadata        bool              `json:"includeMetadata,omitempty"`
	ExtractMetadata        *ExtractMetadata  `json:"extractMetadata,omitempty"`
	FailIfContentMissing   []string          `json:"failIfContentMissing,omitempty"`
	FailIfContentContains  []string          `json:"failIfContentContains,omitempty"`
}

// ScrollEasing represents the easing function for scroll animation.
type ScrollEasing string

const (
	ScrollEasingLinear        ScrollEasing = "linear"
	ScrollEasingEaseIn        ScrollEasing = "ease_in"
	ScrollEasingEaseOut       ScrollEasing = "ease_out"
	ScrollEasingEaseInOut     ScrollEasing = "ease_in_out"
	ScrollEasingEaseInOutQuint ScrollEasing = "ease_in_out_quint"
)

// VideoOptions represents options for capturing a video.
type VideoOptions struct {
	URL               string            `json:"url"`
	Format            string            `json:"format,omitempty"`
	Quality           *int              `json:"quality,omitempty"`
	Width             int               `json:"width,omitempty"`
	Height            int               `json:"height,omitempty"`
	Device            DevicePreset      `json:"device,omitempty"`
	Duration          int               `json:"duration,omitempty"`
	FPS               int               `json:"fps,omitempty"`
	Delay             int               `json:"delay,omitempty"`
	Timeout           int               `json:"timeout,omitempty"`
	WaitUntil         string            `json:"waitUntil,omitempty"`
	WaitForSelector   string            `json:"waitForSelector,omitempty"`
	DarkMode          bool              `json:"darkMode,omitempty"`
	BlockAds          bool              `json:"blockAds,omitempty"`
	BlockCookieBanners bool             `json:"blockCookieBanners,omitempty"`
	CSS               string            `json:"css,omitempty"`
	JavaScript        string            `json:"javascript,omitempty"`
	HideSelectors     []string          `json:"hideSelectors,omitempty"`
	UserAgent         string            `json:"userAgent,omitempty"`
	Cookies           []Cookie          `json:"cookies,omitempty"`
	ResponseType      string            `json:"responseType,omitempty"`
	Scroll            bool              `json:"scroll,omitempty"`
	ScrollDelay       *int              `json:"scrollDelay,omitempty"`
	ScrollDuration    *int              `json:"scrollDuration,omitempty"`
	ScrollBy          *int              `json:"scrollBy,omitempty"`
	ScrollEasing      ScrollEasing      `json:"scrollEasing,omitempty"`
	ScrollBack        bool              `json:"scrollBack,omitempty"`
	ScrollComplete    bool              `json:"scrollComplete,omitempty"`
}

// VideoResult represents the result of a video capture.
type VideoResult struct {
	Success  bool   `json:"success"`
	Data     string `json:"data,omitempty"`
	Format   string `json:"format"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	FileSize int    `json:"fileSize"`
	Duration int    `json:"duration"`
	Took     int    `json:"took"`
}

// ScreenshotMetadata represents page metadata from screenshot.
type ScreenshotMetadata struct {
	Title          string   `json:"title,omitempty"`
	Description    string   `json:"description,omitempty"`
	Favicon        string   `json:"favicon,omitempty"`
	OGTitle        string   `json:"ogTitle,omitempty"`
	OGDescription  string   `json:"ogDescription,omitempty"`
	OGImage        string   `json:"ogImage,omitempty"`
	HTTPStatusCode int      `json:"httpStatusCode,omitempty"`
	Fonts          []string `json:"fonts,omitempty"`
	Colors         []string `json:"colors,omitempty"`
	Links          []string `json:"links,omitempty"`
}

// ScreenshotResult represents the result of a screenshot with metadata.
type ScreenshotResult struct {
	Success   bool                `json:"success"`
	Data      string              `json:"data"`
	Width     int                 `json:"width"`
	Height    int                 `json:"height"`
	FileSize  int                 `json:"fileSize"`
	Took      int                 `json:"took"`
	Format    string              `json:"format"`
	Cached    bool                `json:"cached"`
	Metadata  *ScreenshotMetadata `json:"metadata,omitempty"`
	Thumbnail string              `json:"thumbnail,omitempty"`
}

// BatchOptions represents options for batch screenshot operations.
type BatchOptions struct {
	URLs               []string `json:"urls"`
	Format             string   `json:"format,omitempty"`
	Quality            *int     `json:"quality,omitempty"`
	Width              int      `json:"width,omitempty"`
	Height             int      `json:"height,omitempty"`
	FullPage           bool     `json:"fullPage,omitempty"`
	DarkMode           bool     `json:"darkMode,omitempty"`
	BlockAds           bool     `json:"blockAds,omitempty"`
	BlockCookieBanners bool     `json:"blockCookieBanners,omitempty"`
	WebhookURL         string   `json:"webhookUrl,omitempty"`
}

// BatchResult represents the result of a batch operation.
type BatchResult struct {
	Success   bool   `json:"success"`
	JobID     string `json:"jobId"`
	Status    string `json:"status"`
	Total     int    `json:"total"`
	Completed int    `json:"completed,omitempty"`
	Failed    int    `json:"failed,omitempty"`
}

// BatchItemResult represents the result of a single item in a batch.
type BatchItemResult struct {
	URL      string `json:"url"`
	Status   string `json:"status"`
	Data     string `json:"data,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration int    `json:"duration,omitempty"`
}

// BatchStatus represents the status of a batch job.
type BatchStatus struct {
	Success     bool              `json:"success"`
	JobID       string            `json:"jobId"`
	Status      string            `json:"status"`
	Total       int               `json:"total"`
	Completed   int               `json:"completed"`
	Failed      int               `json:"failed"`
	Results     []BatchItemResult `json:"results,omitempty"`
	CreatedAt   string            `json:"createdAt,omitempty"`
	CompletedAt string            `json:"completedAt,omitempty"`
}

// DeviceInfo represents device preset information.
type DeviceInfo struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Width             int     `json:"width"`
	Height            int     `json:"height"`
	DeviceScaleFactor float64 `json:"deviceScaleFactor"`
	IsMobile          bool    `json:"isMobile"`
}

// DevicesResult represents the result of GetDevices.
type DevicesResult struct {
	Success bool                   `json:"success"`
	Devices map[string][]DeviceInfo `json:"devices"`
	Total   int                    `json:"total"`
}

// CapabilitiesResult represents the result of GetCapabilities.
type CapabilitiesResult struct {
	Success      bool                   `json:"success"`
	Version      string                 `json:"version"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// Usage represents API usage statistics.
type Usage struct {
	Used      int    `json:"used"`
	Limit     int    `json:"limit"`
	Remaining int    `json:"remaining"`
	ResetAt   string `json:"resetAt"`
}
