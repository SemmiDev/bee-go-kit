package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/semmidev/bee-go-kit/apperror"
)

// ---------------------------------------------------------------------------
// HandlerConfig – configurable messages for HTTP error handlers
// ---------------------------------------------------------------------------

// HandlerConfig allows customisation of the messages returned by the
// pre-built error handler functions. Provide your own instance to
// NewErrorHandlers, or use DefaultHandlerConfig for sensible defaults.
type HandlerConfig struct {
	NotFoundMessage      string // 404
	MethodNotAllowedMsg  string // 405
	InternalErrorMessage string // 500
	BadRequestMessage    string // 400
	UnauthorizedMessage  string // 401
	ForbiddenMessage     string // 403
}

// DefaultHandlerConfig returns a HandlerConfig with English defaults.
func DefaultHandlerConfig() HandlerConfig {
	return HandlerConfig{
		NotFoundMessage:      "Endpoint not found",
		MethodNotAllowedMsg:  "Method not allowed",
		InternalErrorMessage: "Internal server error",
		BadRequestMessage:    "Bad request",
		UnauthorizedMessage:  "Unauthorized",
		ForbiddenMessage:     "Forbidden",
	}
}

// ---------------------------------------------------------------------------
// ErrorHandlers – a set of http.HandlerFunc for common error status codes
// ---------------------------------------------------------------------------

// ErrorHandlers holds pre-built http.HandlerFunc for common HTTP error codes.
// Create one via NewErrorHandlers and register them with your router/framework.
type ErrorHandlers struct {
	cfg HandlerConfig
}

// NewErrorHandlers creates an ErrorHandlers set with the given configuration.
func NewErrorHandlers(cfg HandlerConfig) *ErrorHandlers {
	return &ErrorHandlers{cfg: cfg}
}

// NotFound handles 404 responses.
func (h *ErrorHandlers) NotFound(w http.ResponseWriter, _ *http.Request) {
	writeErrorResponse(w, http.StatusNotFound, apperror.ErrCodeEndpointNotFound, h.cfg.NotFoundMessage)
}

// MethodNotAllowed handles 405 responses.
func (h *ErrorHandlers) MethodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	writeErrorResponse(w, http.StatusMethodNotAllowed, apperror.ErrCodeMethodNotAllowed, h.cfg.MethodNotAllowedMsg)
}

// InternalError handles 500 responses.
func (h *ErrorHandlers) InternalError(w http.ResponseWriter, _ *http.Request) {
	writeErrorResponse(w, http.StatusInternalServerError, apperror.ErrCodeInternal, h.cfg.InternalErrorMessage)
}

// BadRequest handles 400 responses.
func (h *ErrorHandlers) BadRequest(w http.ResponseWriter, _ *http.Request) {
	writeErrorResponse(w, http.StatusBadRequest, apperror.ErrCodeBadRequest, h.cfg.BadRequestMessage)
}

// Unauthorized handles 401 responses.
func (h *ErrorHandlers) Unauthorized(w http.ResponseWriter, _ *http.Request) {
	writeErrorResponse(w, http.StatusUnauthorized, apperror.ErrCodeUnauthorized, h.cfg.UnauthorizedMessage)
}

// Forbidden handles 403 responses.
func (h *ErrorHandlers) Forbidden(w http.ResponseWriter, _ *http.Request) {
	writeErrorResponse(w, http.StatusForbidden, apperror.ErrCodeForbidden, h.cfg.ForbiddenMessage)
}

// ---------------------------------------------------------------------------
// Internal helper
// ---------------------------------------------------------------------------

// writeErrorResponse writes a minimal JSON error response.
func writeErrorResponse(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	resp := Response{
		Success:   false,
		ErrorCode: code,
		Message:   message,
	}

	_ = json.NewEncoder(w).Encode(resp)
}
