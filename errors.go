package snapapi

import "fmt"

// Error codes returned by the API.
const (
	ErrInvalidURL      = "INVALID_URL"
	ErrInvalidParams   = "INVALID_PARAMS"
	ErrUnauthorized    = "UNAUTHORIZED"
	ErrForbidden       = "FORBIDDEN"
	ErrQuotaExceeded   = "QUOTA_EXCEEDED"
	ErrRateLimited     = "RATE_LIMITED"
	ErrTimeout         = "TIMEOUT"
	ErrCaptureFailed   = "CAPTURE_FAILED"
	ErrConnectionError = "CONNECTION_ERROR"
)

// APIError represents an error returned by the SnapAPI.
type APIError struct {
	Code       string                   `json:"code"`
	Message    string                   `json:"message"`
	StatusCode int                      `json:"statusCode"`
	Details    []map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("[%s] %s (HTTP %d): %v", e.Code, e.Message, e.StatusCode, e.Details)
	}
	return fmt.Sprintf("[%s] %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
}

// IsRetryable returns true if the error is retryable.
func (e *APIError) IsRetryable() bool {
	return e.Code == ErrRateLimited || e.Code == ErrTimeout || e.StatusCode >= 500
}

// errorResponse represents the error response structure from the API.
// The API returns: {"statusCode": N, "error": "ErrorType", "message": "...", "details": [...]}
type errorResponse struct {
	StatusCode int                      `json:"statusCode"`
	Error      string                   `json:"error"`
	Message    string                   `json:"message"`
	Details    []map[string]interface{} `json:"details,omitempty"`
}
