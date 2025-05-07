package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// Standard errors to use across the application
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInternalServer     = errors.New("internal server error")
	ErrBadRequest         = errors.New("bad request")
	ErrTimeout            = errors.New("request timeout")
	ErrConflict           = errors.New("conflict")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// AppError represents an application error with HTTP status code and optional metadata
type AppError struct {
	Err        error
	StatusCode int
	Code       string
	Metadata   map[string]interface{}
}

// Error returns the error message
func (e *AppError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithCode adds an error code to the error
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// Is implements error Is interface
func (e *AppError) Is(target error) bool {
	return errors.Is(e.Err, target)
}

// New creates a new AppError
func New(err error, statusCode int) *AppError {
	if err == nil {
		err = ErrInternalServer
	}
	return &AppError{
		Err:        err,
		StatusCode: statusCode,
	}
}

// NewWithMessage creates a new AppError with a formatted message
func NewWithMessage(format string, args ...interface{}) *AppError {
	return &AppError{
		Err:        fmt.Errorf(format, args...),
		StatusCode: http.StatusInternalServerError,
	}
}

// NotFound creates a not found error
func NotFound(message string) *AppError {
	if message == "" {
		return New(ErrNotFound, http.StatusNotFound)
	}
	return New(fmt.Errorf("%s: %w", message, ErrNotFound), http.StatusNotFound)
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	if message == "" {
		return New(ErrBadRequest, http.StatusBadRequest)
	}
	return New(fmt.Errorf("%s: %w", message, ErrBadRequest), http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	if message == "" {
		return New(ErrUnauthorized, http.StatusUnauthorized)
	}
	return New(fmt.Errorf("%s: %w", message, ErrUnauthorized), http.StatusUnauthorized)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	if message == "" {
		return New(ErrForbidden, http.StatusForbidden)
	}
	return New(fmt.Errorf("%s: %w", message, ErrForbidden), http.StatusForbidden)
}

// Internal creates an internal server error
func Internal(err error) *AppError {
	if err == nil {
		return New(ErrInternalServer, http.StatusInternalServerError)
	}
	return New(fmt.Errorf("%s: %w", err.Error(), ErrInternalServer), http.StatusInternalServerError)
}

// Conflict creates a conflict error
func Conflict(message string) *AppError {
	if message == "" {
		return New(ErrConflict, http.StatusConflict)
	}
	return New(fmt.Errorf("%s: %w", message, ErrConflict), http.StatusConflict)
}

// FromError converts a standard error to an AppError
func FromError(err error) *AppError {
	if err == nil {
		return nil
	}

	// Check if it's already an AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// Create an appropriate AppError based on the error
	switch {
	case errors.Is(err, ErrNotFound):
		return NotFound("")
	case errors.Is(err, ErrBadRequest):
		return BadRequest("")
	case errors.Is(err, ErrUnauthorized):
		return Unauthorized("")
	case errors.Is(err, ErrForbidden):
		return Forbidden("")
	case errors.Is(err, ErrConflict):
		return Conflict("")
	default:
		return Internal(err)
	}
}
