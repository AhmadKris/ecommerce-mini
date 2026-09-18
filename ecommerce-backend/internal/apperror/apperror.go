// Package apperror defines the typed application error used across service
// and handler layers, decoupling client-facing messages from internal causes
// and mapping error categories to HTTP status codes in one place.
package apperror

import (
	"errors"
	"net/http"
)

// Code identifies the category of an AppError, independent of its message.
type Code string

const (
	CodeNotFound        Code = "NOT_FOUND"
	CodeValidation      Code = "VALIDATION_ERROR"
	CodeUnauthorized    Code = "UNAUTHORIZED"
	CodeForbidden       Code = "FORBIDDEN"
	CodeConflict        Code = "CONFLICT"
	CodeDuplicateEntry  Code = "DUPLICATE_ENTRY"
	CodeRateLimited     Code = "RATE_LIMITED"
	CodePayloadTooLarge Code = "PAYLOAD_TOO_LARGE"
	CodeInternal        Code = "INTERNAL_ERROR"
)

// AppError is the single error type that crosses the service → handler
// boundary. Message is safe to show to the client; Err carries the original
// cause and is only ever written to logs.
type AppError struct {
	Code    Code
	Message string
	Errors  []string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

// Unwrap exposes the original cause so callers can still use errors.Is/As
// against sentinel errors from the repository layer.
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus maps the error's Code to the HTTP status the handler middleware
// should write.
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeValidation:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeConflict:
		return http.StatusConflict
	case CodeDuplicateEntry:
		return http.StatusConflict
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodePayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

func NotFound(message string, err error) *AppError {
	return &AppError{Code: CodeNotFound, Message: message, Err: err}
}

func Validation(message string, errs []string) *AppError {
	return &AppError{Code: CodeValidation, Message: message, Errors: errs}
}

func Unauthorized(message string, err error) *AppError {
	return &AppError{Code: CodeUnauthorized, Message: message, Err: err}
}

func Forbidden(message string) *AppError {
	return &AppError{Code: CodeForbidden, Message: message}
}

func Conflict(message string, err error) *AppError {
	return &AppError{Code: CodeConflict, Message: message, Err: err}
}

func DuplicateEntry(message string, err error) *AppError {
	return &AppError{Code: CodeDuplicateEntry, Message: message, Err: err}
}

func RateLimited(message string) *AppError {
	return &AppError{Code: CodeRateLimited, Message: message}
}

func PayloadTooLarge(message string) *AppError {
	return &AppError{Code: CodePayloadTooLarge, Message: message}
}

// Internal wraps an unclassified error with a generic client-facing message.
// The original err is preserved for logging but never reaches the client.
func Internal(err error) *AppError {
	return &AppError{Code: CodeInternal, Message: "Terjadi kesalahan pada server", Err: err}
}

// As is a convenience wrapper around errors.As for the common case of
// extracting an *AppError from an error chain.
func As(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
