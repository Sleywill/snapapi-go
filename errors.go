package snapapi

import (
	"encoding/json"
	"fmt"
)

// Error code constants.
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

// APIError represents a structured error returned by the SnapAPI.
type APIError struct {
	Code       string                   `json:"code"`
	Message    string                   `json:"message"`
	StatusCode int                      `json:"statusCode"`
	Details    []map[string]interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if len(e.Details) > 0 {
		return fmt.Sprintf("[%s] %s (HTTP %d): %v", e.Code, e.Message, e.StatusCode, e.Details)
	}
	return fmt.Sprintf("[%s] %s (HTTP %d)", e.Code, e.Message, e.StatusCode)
}

// IsRetryable reports whether the error may succeed on retry.
func (e *APIError) IsRetryable() bool {
	return e.Code == ErrRateLimited || e.Code == ErrTimeout || e.StatusCode >= 500
}

// rawAPIError is the JSON shape returned by the API on errors.
type rawAPIError struct {
	StatusCode int                      `json:"statusCode"`
	Error      string                   `json:"error"`
	Message    string                   `json:"message"`
	Details    []map[string]interface{} `json:"details,omitempty"`
}

// parseAPIError builds an *APIError from a raw response body and HTTP status code.
func parseAPIError(body []byte, statusCode int) *APIError {
	var raw rawAPIError
	if err := json.Unmarshal(body, &raw); err != nil || raw.Message == "" {
		return &APIError{
			Code:       "HTTP_ERROR",
			Message:    fmt.Sprintf("HTTP %d: %s", statusCode, string(body)),
			StatusCode: statusCode,
		}
	}

	code := mapCode(raw.Error, statusCode)
	return &APIError{
		Code:       code,
		Message:    raw.Message,
		StatusCode: statusCode,
		Details:    raw.Details,
	}
}

func mapCode(errType string, statusCode int) string {
	switch statusCode {
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 429:
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
