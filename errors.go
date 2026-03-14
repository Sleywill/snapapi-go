package snapapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// Error code constants returned in APIError.Code.
const (
	ErrInvalidParams   = "INVALID_PARAMS"
	ErrUnauthorized    = "UNAUTHORIZED"
	ErrForbidden       = "FORBIDDEN"
	ErrQuotaExceeded   = "QUOTA_EXCEEDED"
	ErrRateLimited     = "RATE_LIMITED"
	ErrTimeout         = "TIMEOUT"
	ErrCaptureFailed   = "CAPTURE_FAILED"
	ErrConnectionError = "CONNECTION_ERROR"
	ErrServerError     = "SERVER_ERROR"
)

// APIError is the structured error type returned by every Client method.
//
// Use errors.As to access the fields:
//
//	var apiErr *snapapi.APIError
//	if errors.As(err, &apiErr) {
//	    switch apiErr.Code {
//	    case snapapi.ErrRateLimited:
//	        time.Sleep(time.Duration(apiErr.RetryAfter) * time.Second)
//	    case snapapi.ErrUnauthorized:
//	        log.Fatal("check your API key")
//	    }
//	}
type APIError struct {
	// Code is a machine-readable error identifier (one of the Err* constants).
	Code string `json:"code"`
	// Message is a human-readable description of the error.
	Message string `json:"message"`
	// StatusCode is the HTTP status code (e.g. 401, 429, 500).
	StatusCode int `json:"statusCode"`
	// RetryAfter is the number of seconds to wait before retrying (from the
	// Retry-After header). Zero means no hint was provided.
	RetryAfter int `json:"retryAfter,omitempty"`
	// Details holds any additional error detail objects returned by the API.
	Details []map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("[%s] %s (HTTP %d): %v", e.Code, e.Message, e.StatusCode, e.Details)
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("[%s] %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// IsRateLimited reports whether the error is a rate-limit (429) response.
func (e *APIError) IsRateLimited() bool { return e.Code == ErrRateLimited }

// IsUnauthorized reports whether the error is an auth failure (401/403).
func (e *APIError) IsUnauthorized() bool {
	return e.Code == ErrUnauthorized || e.Code == ErrForbidden
}

// IsServerError reports whether the error is a 5xx server-side failure.
func (e *APIError) IsServerError() bool { return e.StatusCode >= 500 }

// isRetryable reports whether err may succeed on a subsequent attempt and
// populates apiErr if the error is an *APIError.
func isRetryable(err error, apiErr **APIError) bool {
	var ae *APIError
	if errors.As(err, &ae) {
		if apiErr != nil {
			*apiErr = ae
		}
		return ae.Code == ErrRateLimited || ae.Code == ErrTimeout || ae.StatusCode >= 500
	}
	// Network-level errors (no HTTP status) are always retryable.
	return true
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// rawAPIError matches the JSON shape the SnapAPI server returns on errors.
type rawAPIError struct {
	StatusCode int                      `json:"statusCode"`
	Error      string                   `json:"error"`
	Message    string                   `json:"message"`
	Details    []map[string]interface{} `json:"details,omitempty"`
}

// parseAPIError builds an *APIError from a raw response body, HTTP status code,
// and response headers (used to extract Retry-After).
func parseAPIError(body []byte, statusCode int, headers http.Header) *APIError {
	var raw rawAPIError
	if err := json.Unmarshal(body, &raw); err != nil || raw.Message == "" {
		return &APIError{
			Code:       "HTTP_ERROR",
			Message:    fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
			StatusCode: statusCode,
		}
	}

	ae := &APIError{
		Code:       mapErrorCode(raw.Error, statusCode),
		Message:    raw.Message,
		StatusCode: statusCode,
		Details:    raw.Details,
	}

	// Parse Retry-After header (seconds or HTTP-date; we handle seconds only).
	if ra := headers.Get("Retry-After"); ra != "" {
		if n, err := strconv.Atoi(ra); err == nil {
			ae.RetryAfter = n
		}
	}
	return ae
}

func mapErrorCode(errType string, statusCode int) string {
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusTooManyRequests:
		return ErrRateLimited
	}
	switch errType {
	case "Validation Error":
		return ErrInvalidParams
	case "Unauthorized":
		return ErrUnauthorized
	case "Forbidden":
		return ErrForbidden
	case "Rate Limited":
		return ErrRateLimited
	case "Timeout":
		return ErrTimeout
	default:
		if statusCode >= 500 {
			return ErrServerError
		}
		return errType
	}
}
