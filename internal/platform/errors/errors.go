package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard errors that can be used directly
var (
	ErrNotFound            = errors.New("resource not found")
	ErrBadRequest          = errors.New("bad request")
	ErrInternalServer      = errors.New("internal server error")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrConflict            = errors.New("conflict")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrValidation          = errors.New("validation error")
	ErrDatabaseQuery       = errors.New("database query error")
	ErrCacheOperation      = errors.New("cache operation error")
	ErrExternalServiceCall = errors.New("external service call error")
)

// ValidationError represents a validation error
type ValidationError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Field     string `json:"field,omitempty"`
	FieldType string `json:"field_type,omitempty"`
}

// Error returns the string representation of the error
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
	}
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(code, message string, field string) *ValidationError {
	return &ValidationError{
		Code:    code,
		Message: message,
		Field:   field,
	}
}

// AppError represents a custom application error
type AppError struct {
	Code    string // Error code, should use constants from code.go
	Message string // Human-readable error message
	Err     error  // Original error (optional)
	Status  int    // HTTP status code
}

// Error returns the string representation of the error
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Err.Error())
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the original error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Status:  http.StatusInternalServerError, // Default status
	}
}

// WithStatus sets the HTTP status code
func (e *AppError) WithStatus(status int) *AppError {
	e.Status = status
	return e
}

// NewNotFound creates a not found error
func NewNotFound(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Status:  http.StatusNotFound,
	}
}

// NewBadRequest creates a bad request error
func NewBadRequest(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Status:  http.StatusBadRequest,
	}
}

// NewUnauthorized creates an unauthorized error
func NewUnauthorized(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Status:  http.StatusUnauthorized,
	}
}

// NewForbidden creates a forbidden error
func NewForbidden(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Status:  http.StatusForbidden,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to an AppError if possible
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// ErrorResponse represents the structure of an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new error response from an error
func NewErrorResponse(err error) ErrorResponse {
	if appErr, ok := AsAppError(err); ok {
		return ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
		}
	}
	return ErrorResponse{
		Code:    fmt.Sprintf("%d", CodeInternalServerError),
		Message: err.Error(),
	}
}
