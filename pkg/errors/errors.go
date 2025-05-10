// Package errors provides standardized error handling utilities for the OneMount project.
// It includes functions for error wrapping, error logging, and error context propagation.
package errors

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// Standard field names for logging
const (
	FieldError     = "error"     // Error message
	FieldOperation = "operation" // Higher-level operation
	FieldComponent = "component" // Component or module
	FieldPath      = "path"      // File or resource path
	FieldID        = "id"        // Identifier
	FieldUser      = "user"      // User identifier
	FieldStatus    = "status"    // Status code or string
)

// Unwrap unwraps an error to find the underlying cause.
// This is a convenience function that uses errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is reports whether any error in err's chain matches target.
// This is a convenience function that uses errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
// This is a convenience function that uses errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Wrap wraps an error with a message.
// This is a convenience function for the common pattern of wrapping an error with context.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted message.
// This is a convenience function for the common pattern of wrapping an error with context.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// New creates a new error with the given message.
// This is a convenience function that uses errors.New.
func New(message string) error {
	return errors.New(message)
}

// LogError logs an error with additional fields.
// This is a convenience function for logging errors with additional context.
func LogError(err error, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	event := log.Error().Err(err)

	// Add additional fields in pairs (key, value)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			event = event.Interface(key, fields[i+1])
		}
	}

	event.Msg(msg)
}

// LogWarn logs a warning with additional fields.
// This is useful for logging potential issues that don't prevent the application from working.
func LogWarn(err error, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	event := log.Warn().Err(err)

	// Add additional fields in pairs (key, value)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key, ok := fields[i].(string)
			if !ok {
				continue
			}
			event = event.Interface(key, fields[i+1])
		}
	}

	event.Msg(msg)
}

// LogAndReturn logs an error and returns it.
// This is a convenience function for the common pattern of logging an error and then returning it.
func LogAndReturn(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	LogError(err, msg, fields...)
	return err
}

// WrapAndLog wraps an error with a message, logs it, and returns the wrapped error.
// This is a convenience function for the common pattern of wrapping an error, logging it, and then returning it.
func WrapAndLog(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	wrapped := Wrap(err, msg)
	LogError(wrapped, msg, fields...)
	return wrapped
}

// WrapfAndLog wraps an error with a formatted message, logs it, and returns the wrapped error.
// This is a convenience function for the common pattern of wrapping an error, logging it, and then returning it.
func WrapfAndLog(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)
	wrapped := Wrap(err, msg)
	LogError(wrapped, msg)
	return wrapped
}
