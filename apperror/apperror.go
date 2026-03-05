// Package apperror provides structured application errors with HTTP status codes,
// machine-readable error codes, and user-friendly messages.
//
// It is designed to unify error handling across HTTP APIs by pairing every error
// with a status code and a stable error code that clients can rely on.
//
// Usage:
//
//	err := apperror.NotFound("user", "123")       // 404
//	err := apperror.BadRequest("invalid email")    // 400
//	err := apperror.Internal("db failed", dbErr)   // 500
//
//	// Check error type:
//	if apperror.IsNotFound(err) { ... }
//
//	// Extract HTTP status from any error:
//	status := apperror.HTTPStatus(err) // 404, 500, etc.
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// ---------------------------------------------------------------------------
// Error struct
// ---------------------------------------------------------------------------

// Error represents a structured application error that carries an HTTP status
// code, a machine-readable error code, a human-readable message, and an
// optional wrapped (original) error for logging purposes.
type Error struct {
	// Code is the HTTP status code associated with this error (e.g. 400, 404).
	Code int

	// ErrorCode is a stable, machine-readable identifier (e.g. "ERR_NOT_FOUND").
	// Clients can switch on this value instead of parsing messages.
	ErrorCode string

	// Message is the user-friendly description of what went wrong.
	Message string

	// Err is the optional underlying error for logging / debugging.
	// It is never exposed to the API consumer directly.
	Err error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.ErrorCode, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.ErrorCode, e.Message)
}

// Unwrap returns the underlying error, enabling errors.Is / errors.As chains.
func (e *Error) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the HTTP status code for this error.
func (e *Error) HTTPStatus() int {
	return e.Code
}

// Wrap returns a new Error with the same code, error-code, and message but
// with a different underlying error attached.
func (e *Error) Wrap(err error) *Error {
	return &Error{
		Code:      e.Code,
		ErrorCode: e.ErrorCode,
		Message:   e.Message,
		Err:       err,
	}
}

// ---------------------------------------------------------------------------
// Error code constants
// ---------------------------------------------------------------------------

const (
	ErrCodeNotFound           = "ERR_NOT_FOUND"
	ErrCodeBadRequest         = "ERR_BAD_REQUEST"
	ErrCodeValidation         = "ERR_VALIDATION"
	ErrCodeUnauthorized       = "ERR_UNAUTHORIZED"
	ErrCodeForbidden          = "ERR_FORBIDDEN"
	ErrCodeConflict           = "ERR_CONFLICT"
	ErrCodeDuplicate          = "ERR_DUPLICATE"
	ErrCodeInternal           = "ERR_INTERNAL"
	ErrCodeServiceUnavailable = "ERR_SERVICE_UNAVAILABLE"
	ErrCodeUnprocessable      = "ERR_UNPROCESSABLE"
	ErrCodeValidationFailed   = "ERR_VALIDATION_FAILED"
	ErrCodeEndpointNotFound   = "ERR_ENDPOINT_NOT_FOUND"
	ErrCodeMethodNotAllowed   = "ERR_METHOD_NOT_ALLOWED"
)

// ---------------------------------------------------------------------------
// Generic constructors – for custom error types
// ---------------------------------------------------------------------------

// New creates an *Error with a custom HTTP status code, machine-readable error
// code, and message. Use this when the pre-built constructors (NotFound,
// BadRequest, etc.) don't cover your use case.
//
//	apperror.New(http.StatusPaymentRequired, "ERR_PAYMENT", "payment required")
func New(httpStatus int, errorCode, message string) *Error {
	return &Error{
		Code:      httpStatus,
		ErrorCode: errorCode,
		Message:   message,
	}
}

// Newf is like New but with a formatted message.
//
//	apperror.Newf(http.StatusPaymentRequired, "ERR_PAYMENT", "plan %s requires payment", plan)
func Newf(httpStatus int, errorCode, format string, args ...any) *Error {
	return New(httpStatus, errorCode, fmt.Sprintf(format, args...))
}

// NewWithErr is like New but wraps an underlying error for logging/debugging.
//
//	apperror.NewWithErr(http.StatusBadGateway, "ERR_UPSTREAM", "upstream failed", err)
func NewWithErr(httpStatus int, errorCode, message string, err error) *Error {
	return &Error{
		Code:      httpStatus,
		ErrorCode: errorCode,
		Message:   message,
		Err:       err,
	}
}

// ---------------------------------------------------------------------------
// Constructors
// ---------------------------------------------------------------------------

// NotFound creates a 404 Not Found error.
//
//	apperror.NotFound("User not found")
func NotFound(message string) *Error {
	return &Error{
		Code:      http.StatusNotFound,
		ErrorCode: ErrCodeNotFound,
		Message:   message,
	}
}

// NotFoundf creates a 404 Not Found error with a formatted message.
//
//	apperror.NotFoundf("user with ID %s not found", id)
func NotFoundf(format string, args ...any) *Error {
	return NotFound(fmt.Sprintf(format, args...))
}

// BadRequest creates a 400 Bad Request error.
func BadRequest(message string) *Error {
	return &Error{
		Code:      http.StatusBadRequest,
		ErrorCode: ErrCodeBadRequest,
		Message:   message,
	}
}

// BadRequestWrap creates a 400 Bad Request error wrapping an underlying error.
func BadRequestWrap(message string, err error) *Error {
	return &Error{
		Code:      http.StatusBadRequest,
		ErrorCode: ErrCodeBadRequest,
		Message:   message,
		Err:       err,
	}
}

// Validation creates a 400 Bad Request error with the validation error code.
func Validation(message string) *Error {
	return &Error{
		Code:      http.StatusBadRequest,
		ErrorCode: ErrCodeValidation,
		Message:   message,
	}
}

// Unauthorized creates a 401 Unauthorized error.
func Unauthorized(message string) *Error {
	return &Error{
		Code:      http.StatusUnauthorized,
		ErrorCode: ErrCodeUnauthorized,
		Message:   message,
	}
}

// Forbidden creates a 403 Forbidden error.
func Forbidden(message string) *Error {
	return &Error{
		Code:      http.StatusForbidden,
		ErrorCode: ErrCodeForbidden,
		Message:   message,
	}
}

// Conflict creates a 409 Conflict error.
func Conflict(message string) *Error {
	return &Error{
		Code:      http.StatusConflict,
		ErrorCode: ErrCodeConflict,
		Message:   message,
	}
}

// Duplicate creates a 409 Conflict error for duplicate resources.
//
//	apperror.Duplicate("User", "email", "foo@bar.com")
func Duplicate(resource, field string, value any) *Error {
	return &Error{
		Code:      http.StatusConflict,
		ErrorCode: ErrCodeDuplicate,
		Message:   fmt.Sprintf("%s with %s '%v' already exists", resource, field, value),
	}
}

// Internal creates a 500 Internal Server Error with both message and cause.
func Internal(message string, err error) *Error {
	return &Error{
		Code:      http.StatusInternalServerError,
		ErrorCode: ErrCodeInternal,
		Message:   message,
		Err:       err,
	}
}

// InternalFromErr creates a 500 Internal Server Error from an error alone.
func InternalFromErr(err error) *Error {
	return &Error{
		Code:      http.StatusInternalServerError,
		ErrorCode: ErrCodeInternal,
		Message:   "internal server error",
		Err:       err,
	}
}

// ServiceUnavailable creates a 503 Service Unavailable error.
func ServiceUnavailable(message string) *Error {
	return &Error{
		Code:      http.StatusServiceUnavailable,
		ErrorCode: ErrCodeServiceUnavailable,
		Message:   message,
	}
}

// Unprocessable creates a 422 Unprocessable Entity error.
func Unprocessable(message string) *Error {
	return &Error{
		Code:      http.StatusUnprocessableEntity,
		ErrorCode: ErrCodeUnprocessable,
		Message:   message,
	}
}

// ---------------------------------------------------------------------------
// Checkers – work with any error via errors.As
// ---------------------------------------------------------------------------

// IsNotFound reports whether err (or any error in its chain) is a 404.
func IsNotFound(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusNotFound
	}
	return false
}

// IsBadRequest reports whether err is a 400.
func IsBadRequest(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusBadRequest
	}
	return false
}

// IsUnauthorized reports whether err is a 401.
func IsUnauthorized(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusUnauthorized
	}
	return false
}

// IsForbidden reports whether err is a 403.
func IsForbidden(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusForbidden
	}
	return false
}

// IsConflict reports whether err is a 409.
func IsConflict(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusConflict
	}
	return false
}

// IsInternal reports whether err is a 500.
func IsInternal(err error) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == http.StatusInternalServerError
	}
	return false
}

// ---------------------------------------------------------------------------
// Utilities
// ---------------------------------------------------------------------------

// HTTPStatus extracts the HTTP status code from an error. If the error is not
// an *Error, it defaults to 500 Internal Server Error.
func HTTPStatus(err error) int {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return http.StatusInternalServerError
}

// Message extracts the user-friendly message from an error. If the error is
// not an *Error, it falls back to err.Error().
func Message(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Message
	}
	return err.Error()
}

// Code extracts the machine-readable error code. Returns empty string for
// non-apperror errors.
func Code(err error) string {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.ErrorCode
	}
	return ""
}
