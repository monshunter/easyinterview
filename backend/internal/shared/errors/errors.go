package errors

import "fmt"

// APIError mirrors the documented error response object
// (B1 shared-conventions-codified §3.2). HTTP handlers wrap it inside an outer
// `{"error": APIError}` envelope; the envelope itself is owned by the
// transport layer in each domain.
type APIError struct {
	Code      string         `json:"code"`
	Message   string         `json:"message"`
	RequestID string         `json:"requestId"`
	Retryable bool           `json:"retryable"`
	Details   map[string]any `json:"details,omitempty"`
}

// Error implements the standard error interface so APIError can flow through
// existing Go error chains.
func (e *APIError) Error() string {
	if e == nil {
		return "<nil APIError>"
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Wrap returns a new APIError with the given documented code, human-readable
// message, and retryable hint. RequestID and Details are populated by the
// transport layer when available.
func Wrap(code, message string, retryable bool) *APIError {
	return &APIError{
		Code:      code,
		Message:   message,
		Retryable: retryable,
	}
}
