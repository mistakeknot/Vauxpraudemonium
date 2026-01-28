package httpapi

import (
	"encoding/json"
	"net/http"
)

// Meta carries optional pagination or response metadata.
type Meta struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

// Error represents a standardized API error payload.
type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   any    `json:"details,omitempty"`
	Retryable bool   `json:"retryable,omitempty"`
}

// Envelope is the standard response wrapper for API endpoints.
type Envelope struct {
	OK    bool   `json:"ok"`
	Data  any    `json:"data,omitempty"`
	Meta  *Meta  `json:"meta,omitempty"`
	Error *Error `json:"error,omitempty"`
}

// OKEnvelope builds a success response.
func OKEnvelope(data any, meta *Meta) Envelope {
	return Envelope{OK: true, Data: data, Meta: meta}
}

// ErrorEnvelope builds an error response.
func ErrorEnvelope(code, message string, details any, retryable bool) Envelope {
	return Envelope{OK: false, Error: &Error{Code: code, Message: message, Details: details, Retryable: retryable}}
}

// WriteJSON writes a JSON response with proper headers.
func WriteJSON(w http.ResponseWriter, status int, env Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(env)
}

// WriteOK writes a success response.
func WriteOK(w http.ResponseWriter, status int, data any, meta *Meta) {
	WriteJSON(w, status, OKEnvelope(data, meta))
}

// WriteError writes an error response.
func WriteError(w http.ResponseWriter, status int, code, message string, details any, retryable bool) {
	WriteJSON(w, status, ErrorEnvelope(code, message, details, retryable))
}

const (
	ErrInvalidRequest = "invalid_request"
	ErrUnauthorized   = "unauthorized"
	ErrForbidden      = "forbidden"
	ErrNotFound       = "not_found"
	ErrConflict       = "conflict"
	ErrRateLimited    = "rate_limited"
	ErrJobPending     = "job_pending"
	ErrJobFailed      = "job_failed"
	ErrInternal       = "internal_error"
)
