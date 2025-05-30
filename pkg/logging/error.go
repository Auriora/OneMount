// Package logging provides standardized logging utilities for the OneMount project.
// This file defines error logging functionality.
//
// Error logging is a critical aspect of application monitoring and debugging. This file
// provides a comprehensive set of functions for logging errors with different levels of
// detail and context. Key features include:
//
//   - Basic error logging with LogError
//   - Warning-level logging with LogWarn, LogWarnWithFields, and LogWarnWithError
//   - Context-aware error logging with LogErrorWithContext
//   - Error wrapping and logging with WrapAndLogError and WrapAndLogErrorWithContext
//   - Error formatting with FormatErrorWithContext
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go (this file): Error logging functionality
//   - performance.go: Performance optimization utilities
package logging

import (
	"fmt"
)

// LogError logs an error with additional fields.
// This is a convenience function for logging errors with additional context.
// The fields parameter can be either a variadic list of key-value pairs or a map[string]interface{}.
func LogError(err error, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	// Check if error level is enabled before performing operations
	if !IsLevelEnabled(ErrorLevel) {
		return
	}

	event := Error().Err(err)

	// Check if fields is a single map[string]interface{}
	if len(fields) == 1 {
		if fieldsMap, ok := fields[0].(map[string]interface{}); ok {
			for key, value := range fieldsMap {
				event = event.Interface(key, value)
			}
			event.Msg(msg)
			return
		}
	}

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

// LogErrorAsWarn logs an error as a warning with additional fields.
// This is useful for logging potential issues that don't prevent the application from working.
func LogErrorAsWarn(err error, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	// Check if warn level is enabled before performing operations
	if !IsLevelEnabled(WarnLevel) {
		return
	}

	event := Warn().Err(err)

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

// LogErrorAsWarnWithFields logs an error as a warning with additional fields
// This is useful for logging potential issues that don't prevent the application from working
func LogErrorAsWarnWithFields(err error, msg string, fields map[string]interface{}) {
	if err == nil {
		return
	}

	// Check if warn level is enabled before performing operations
	if !IsLevelEnabled(WarnLevel) {
		return
	}

	event := Warn().Err(err)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg(msg)
}

// LogErrorWithContext logs an error with the given context
// The fields parameter can be either a variadic list of key-value pairs or a map[string]interface{}.
func LogErrorWithContext(err error, ctx LogContext, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	// Check if error level is enabled before performing operations
	if !IsLevelEnabled(ErrorLevel) {
		return
	}

	logger := WithLogContext(ctx)
	event := logger.Error().Err(err)

	// Check if fields is a single map[string]interface{}
	if len(fields) == 1 {
		if fieldsMap, ok := fields[0].(map[string]interface{}); ok {
			for key, value := range fieldsMap {
				event = event.Interface(key, value)
			}
			event.Msg(msg)
			return
		}
	}

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

// WrapAndLogError wraps an error with a message, logs it, and returns the wrapped error.
// This is a convenience function for the common pattern of wrapping an error, logging it, and then returning it.
// The fields parameter can be either a variadic list of key-value pairs or a map[string]interface{}.
func WrapAndLogError(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	// We can't use errors.Wrap here to avoid circular dependency
	wrapped := fmt.Errorf("%s: %w", msg, err)

	// Only log if error level is enabled
	if IsLevelEnabled(ErrorLevel) {
		LogError(wrapped, msg, fields...)
	}

	return wrapped
}

// WrapAndLogErrorf wraps an error with a formatted message, logs it, and returns the wrapped error.
func WrapAndLogErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	msg := fmt.Sprintf(format, args...)
	return WrapAndLogError(err, msg)
}

// WrapAndLogErrorWithContext wraps an error with a message, logs it with context, and returns the wrapped error.
// This is a convenience function for the common pattern of wrapping an error, logging it with context, and then returning it.
// The fields parameter can be either a variadic list of key-value pairs or a map[string]interface{}.
func WrapAndLogErrorWithContext(err error, ctx LogContext, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	// We can't use errors.Wrap here to avoid circular dependency
	wrapped := fmt.Errorf("%s: %w", msg, err)

	// Only log if error level is enabled
	if IsLevelEnabled(ErrorLevel) {
		LogErrorWithContext(wrapped, ctx, msg, fields...)
	}

	return wrapped
}

// FormatErrorWithContext formats an error message with additional context
// This is useful for creating descriptive error messages
func FormatErrorWithContext(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	// Format the message with the fields
	formattedMsg := msg
	if len(fields) > 0 {
		formattedMsg += " ("
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				key, ok := fields[i].(string)
				if !ok {
					continue
				}
				formattedMsg += fmt.Sprintf("%s=%v", key, fields[i+1])
				if i+2 < len(fields) {
					formattedMsg += ", "
				}
			}
		}
		formattedMsg += ")"
	}

	// Return a new error with the formatted message and the original error
	return fmt.Errorf("%s: %w", formattedMsg, err)
}
