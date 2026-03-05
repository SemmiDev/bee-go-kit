// Package httputil provides a standard API response envelope, JSON writing
// helpers, and pre-built HTTP error handlers for common status codes.
//
// The response format is designed to be consistent across all endpoints:
//
//	{
//	  "success": true|false,
//	  "error_code": "ERR_...",      // only on failure
//	  "message": "...",
//	  "data": { ... },              // only on success
//	  "validation": { ... },        // only on validation failure
//	  "paging": { ... }             // only for paginated lists
//	}
package httputil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/semmidev/bee-go-kit/apperror"
	"github.com/semmidev/bee-go-kit/pagination"
)

// ---------------------------------------------------------------------------
// Response struct
// ---------------------------------------------------------------------------

// Response is the standard API response envelope. Every endpoint should return
// data wrapped in this structure to ensure a consistent contract for clients.
type Response struct {
	// Success indicates whether the request was successful.
	Success bool `json:"success"`

	// ErrorCode is a stable machine-readable code (populated only on failure).
	ErrorCode string `json:"error_code,omitempty"`

	// Message is a short human-readable description of the result.
	Message string `json:"message"`

	// Data holds the response payload (populated only on success).
	Data any `json:"data,omitempty"`

	// Validation holds per-field validation error messages (populated only on
	// validation failure).
	Validation map[string]string `json:"validation,omitempty"`

	// Paging holds pagination metadata (populated only for list endpoints).
	Paging *pagination.Paging `json:"paging,omitempty"`
}

// ---------------------------------------------------------------------------
// Error transformer – extensibility hook
// ---------------------------------------------------------------------------

// ErrorTransformer is a function that transforms a custom error into an HTTP
// status code and Response. Return (0, Response{}) to signal that the default
// logic should handle this error.
//
// Example:
//
//	httputil.SetErrorTransformer(func(err error) (int, httputil.Response) {
//	    var domainErr *domain.Error
//	    if errors.As(err, &domainErr) {
//	        return domainErr.Status, httputil.Response{
//	            Success:   false,
//	            ErrorCode: domainErr.Code,
//	            Message:   domainErr.Message,
//	        }
//	    }
//	    return 0, httputil.Response{} // fall through to default
//	})
type ErrorTransformer func(err error) (int, Response)

var errorTransformer ErrorTransformer

// SetErrorTransformer registers a custom error transformer that is called
// before the default apperror logic in FromError. Pass nil to remove it.
func SetErrorTransformer(fn ErrorTransformer) {
	errorTransformer = fn
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// FromError creates a Response from any error. If an ErrorTransformer is set
// and returns a non-zero status, that result is used. Otherwise, if the error
// is an *apperror.Error, the status code and error code are extracted from it;
// if not, a 500 Internal Server Error is returned.
//
// Returns (httpStatusCode, response).
func FromError(err error) (int, Response) {
	// Try custom transformer first.
	if errorTransformer != nil {
		if status, resp := errorTransformer(err); status != 0 {
			return status, resp
		}
	}

	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		return appErr.Code, Response{
			Success:   false,
			ErrorCode: appErr.ErrorCode,
			Message:   appErr.Message,
		}
	}

	return http.StatusInternalServerError, Response{
		Success:   false,
		ErrorCode: apperror.ErrCodeInternal,
		Message:   fmt.Sprintf("internal server error: %s", err.Error()),
	}
}

// Success creates a success Response (without pagination).
func Success(message string, data any) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// SuccessWithPaging creates a success Response with pagination metadata.
func SuccessWithPaging(message string, data any, paging *pagination.Paging) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
		Paging:  paging,
	}
}

// ValidationError creates a 400 validation error Response with per-field
// error messages.
func ValidationError(message string, validationErrors map[string]string) Response {
	return Response{
		Success:    false,
		ErrorCode:  apperror.ErrCodeValidationFailed,
		Message:    message,
		Validation: validationErrors,
	}
}

// ---------------------------------------------------------------------------
// JSON writers – framework-agnostic (work with net/http directly)
// ---------------------------------------------------------------------------

// WriteJSON writes an HTTP response with the given status code and encodes
// body as JSON. HTML escaping is disabled so SVG / raw HTML in data survives.
func WriteJSON(w http.ResponseWriter, statusCode int, body any) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false) // preserve <, >, & in data (important for SVG)

	if err := enc.Encode(body); err != nil {
		// Fallback: write a minimal error body (headers not yet sent).
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"success":false,"message":"JSON encoding error"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write(buf.Bytes())
}

// WriteSuccess is a convenience wrapper that writes a success Response.
func WriteSuccess(w http.ResponseWriter, statusCode int, message string, data any) {
	WriteJSON(w, statusCode, Success(message, data))
}

// WriteSuccessWithPaging is a convenience wrapper that writes a paginated
// success Response.
func WriteSuccessWithPaging(w http.ResponseWriter, statusCode int, message string, data any, paging *pagination.Paging) {
	WriteJSON(w, statusCode, SuccessWithPaging(message, data, paging))
}

// WriteError writes an error Response. The status code is extracted from the
// error if it is an *apperror.Error; otherwise 500 is used.
func WriteError(w http.ResponseWriter, err error) {
	status, resp := FromError(err)
	WriteJSON(w, status, resp)
}

// WriteValidationError writes a 400 validation error Response.
func WriteValidationError(w http.ResponseWriter, message string, validationErrors map[string]string) {
	WriteJSON(w, http.StatusBadRequest, ValidationError(message, validationErrors))
}
