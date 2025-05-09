// Package errors provides standardized error handling utilities for the OneMount project.
// This file defines specialized error types for common error scenarios.
package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents the type of error that occurred.
type ErrorType int

// Error types
const (
	// ErrorTypeUnknown represents an unknown error.
	ErrorTypeUnknown ErrorType = iota
	
	// ErrorTypeNetwork represents a network-related error.
	ErrorTypeNetwork
	
	// ErrorTypeNotFound represents a resource not found error.
	ErrorTypeNotFound
	
	// ErrorTypeAuth represents an authentication or authorization error.
	ErrorTypeAuth
	
	// ErrorTypeValidation represents a validation error.
	ErrorTypeValidation
	
	// ErrorTypeOperation represents an operation error.
	ErrorTypeOperation
	
	// ErrorTypeTimeout represents a timeout error.
	ErrorTypeTimeout
	
	// ErrorTypeResourceBusy represents a resource busy error.
	ErrorTypeResourceBusy
)

// String returns the string representation of the error type.
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeNetwork:
		return "NetworkError"
	case ErrorTypeNotFound:
		return "NotFoundError"
	case ErrorTypeAuth:
		return "AuthError"
	case ErrorTypeValidation:
		return "ValidationError"
	case ErrorTypeOperation:
		return "OperationError"
	case ErrorTypeTimeout:
		return "TimeoutError"
	case ErrorTypeResourceBusy:
		return "ResourceBusyError"
	default:
		return "UnknownError"
	}
}

// TypedError is an error with a specific type and optional HTTP status code.
type TypedError struct {
	Type       ErrorType
	Message    string
	StatusCode int
	Err        error
}

// Error returns the error message.
func (e *TypedError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error.
func (e *TypedError) Unwrap() error {
	return e.Err
}

// NewNetworkError creates a new network error.
func NewNetworkError(message string, err error) error {
	return &TypedError{
		Type:       ErrorTypeNetwork,
		Message:    message,
		StatusCode: http.StatusServiceUnavailable,
		Err:        err,
	}
}

// NewNotFoundError creates a new not found error.
func NewNotFoundError(message string, err error) error {
	return &TypedError{
		Type:       ErrorTypeNotFound,
		Message:    message,
		StatusCode: http.StatusNotFound,
		Err:        err,
	}
}

// NewAuthError creates a new authentication error.
func NewAuthError(message string, err error) error {
	return &TypedError{
		Type:       ErrorTypeAuth,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
		Err:        err,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string, err error) error {
	return &TypedError{
		Type:       ErrorTypeValidation,
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

// NewOperationError creates a new operation error.
func NewOperationError(message string, err error) error {
	return &TypedError{
		Type:       ErrorTypeOperation,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(message string, err error) error {
	return &TypedError{
		Type:       ErrorTypeTimeout,
		Message:    message,
		StatusCode: http.StatusRequestTimeout,
		Err:        err,
	}
}

// NewResourceBusyError creates a new resource busy error.
func NewResourceBusyError(message string, err error) error {
	return &TypedError{
		Type:       ErrorTypeResourceBusy,
		Message:    message,
		StatusCode: http.StatusConflict,
		Err:        err,
	}
}

// IsNetworkError checks if the error is a network error.
func IsNetworkError(err error) bool {
	var typedErr *TypedError
	if As(err, &typedErr) {
		return typedErr.Type == ErrorTypeNetwork
	}
	return false
}

// IsNotFoundError checks if the error is a not found error.
func IsNotFoundError(err error) bool {
	var typedErr *TypedError
	if As(err, &typedErr) {
		return typedErr.Type == ErrorTypeNotFound
	}
	return false
}

// IsAuthError checks if the error is an authentication error.
func IsAuthError(err error) bool {
	var typedErr *TypedError
	if As(err, &typedErr) {
		return typedErr.Type == ErrorTypeAuth
	}
	return false
}

// IsValidationError checks if the error is a validation error.
func IsValidationError(err error) bool {
	var typedErr *TypedError
	if As(err, &typedErr) {
		return typedErr.Type == ErrorTypeValidation
	}
	return false
}

// IsOperationError checks if the error is an operation error.
func IsOperationError(err error) bool {
	var typedErr *TypedError
	if As(err, &typedErr) {
		return typedErr.Type == ErrorTypeOperation
	}
	return false
}

// IsTimeoutError checks if the error is a timeout error.
func IsTimeoutError(err error) bool {
	var typedErr *TypedError
	if As(err, &typedErr) {
		return typedErr.Type == ErrorTypeTimeout
	}
	return false
}

// IsResourceBusyError checks if the error is a resource busy error.
func IsResourceBusyError(err error) bool {
	var typedErr *TypedError
	if As(err, &typedErr) {
		return typedErr.Type == ErrorTypeResourceBusy
	}
	return false
}