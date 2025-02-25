package errors

import (
	"fmt"
	"net/http"
	"time"
)

// AppError represents an application error
type AppError struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	Internal   error       `json:"-"`
	RequestID  string      `json:"request_id,omitempty"`
	ErrorCode  string      `json:"error_code,omitempty"`
	Timestamp  int64       `json:"timestamp"`
	Path       string      `json:"path,omitempty"`
	Validation []string    `json:"validation,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// NewError creates a new AppError
func NewError(code int, message string, opts ...ErrorOption) *AppError {
	err := &AppError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
	}

	for _, opt := range opts {
		opt(err)
	}

	return err
}

// ErrorOption is a function that configures an AppError
type ErrorOption func(*AppError)

// WithInternal adds an internal error
func WithInternal(err error) ErrorOption {
	return func(e *AppError) {
		e.Internal = err
	}
}

// WithDetails adds details to the error
func WithDetails(details interface{}) ErrorOption {
	return func(e *AppError) {
		e.Details = details
	}
}

// WithRequestID adds a request ID
func WithRequestID(requestID string) ErrorOption {
	return func(e *AppError) {
		e.RequestID = requestID
	}
}

// WithPath adds the request path
func WithPath(path string) ErrorOption {
	return func(e *AppError) {
		e.Path = path
	}
}

// WithValidation adds validation errors
func WithValidation(validationErrors []string) ErrorOption {
	return func(e *AppError) {
		e.Validation = validationErrors
	}
}

// Common error constructors
func BadRequest(message string, opts ...ErrorOption) *AppError {
	return NewError(http.StatusBadRequest, message, opts...)
}

func Unauthorized(message string, opts ...ErrorOption) *AppError {
	return NewError(http.StatusUnauthorized, message, opts...)
}

func Forbidden(message string, opts ...ErrorOption) *AppError {
	return NewError(http.StatusForbidden, message, opts...)
}

func NotFound(message string, opts ...ErrorOption) *AppError {
	return NewError(http.StatusNotFound, message, opts...)
}

func InternalServer(message string, opts ...ErrorOption) *AppError {
	return NewError(http.StatusInternalServerError, message, opts...)
}

func Conflict(message string, opts ...ErrorOption) *AppError {
	return NewError(http.StatusConflict, message, opts...)
}