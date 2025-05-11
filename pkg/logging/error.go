// Package logging provides standardized logging utilities for the OneMount project.
// This file defines error logging functionality.
//
// Error logging is a critical aspect of application monitoring and debugging. This file
// provides a comprehensive set of functions for logging errors with different levels of
// detail and context. Key features include:
//
//   - Basic error logging with LogError and LogErrorWithFields
//   - Warning-level logging with LogWarn, LogWarnWithFields, and LogWarnWithError
//   - Context-aware error logging with LogErrorWithContext
//   - Combined logging and returning with LogAndReturn, LogErrorAndReturn, and LogErrorWithContextAndReturn
//   - Error wrapping and logging with WrapAndLog and WrapfAndLog
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
func LogError(err error, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	event := Error().Err(err)

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

// LogErrorWithFields logs an error with additional fields
// This is a convenience function for logging errors with additional context
func LogErrorWithFields(err error, msg string, fields map[string]interface{}) {
	if err == nil {
		return
	}

	event := Error().Err(err)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg(msg)
}

// LogWarn logs a warning with additional fields.
// This is useful for logging potential issues that don't prevent the application from working.
func LogWarn(err error, msg string, fields ...interface{}) {
	if err == nil {
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

// LogWarnWithFields logs a warning with additional fields
// This is useful for logging potential issues that don't prevent the application from working
func LogWarnWithFields(msg string, fields map[string]interface{}) {
	event := Warn()

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg(msg)
}

// LogWarnWithError logs a warning with an error
// This is useful for logging non-critical errors that don't prevent the application from working
func LogWarnWithError(err error, msg string, fields ...interface{}) {
	if err == nil {
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

// LogAndReturn logs an error and returns it.
// This is a convenience function for the common pattern of logging an error and then returning it.
func LogAndReturn(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	LogError(err, msg, fields...)
	return err
}

// LogErrorAndReturn logs an error and returns it
// This is a convenience function for the common pattern of logging an error and then returning it
func LogErrorAndReturn(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	LogError(err, msg, fields...)
	return err
}

// LogErrorWithContext logs an error with the given context
func LogErrorWithContext(err error, ctx LogContext, msg string, fields ...interface{}) {
	if err == nil {
		return
	}

	logger := WithLogContext(ctx)
	event := logger.Error().Err(err)

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

// LogErrorWithContextAndReturn logs an error with context and returns it
// This is a convenience function for the common pattern of logging an error with context and then returning it
func LogErrorWithContextAndReturn(err error, ctx LogContext, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	LogErrorWithContext(err, ctx, msg, fields...)
	return err
}

// WrapAndLog wraps an error with a message, logs it, and returns the wrapped error.
// This is a convenience function for the common pattern of wrapping an error, logging it, and then returning it.
func WrapAndLog(err error, msg string, fields ...interface{}) error {
	if err == nil {
		return nil
	}

	// We can't use errors.Wrap here to avoid circular dependency
	wrapped := fmt.Errorf("%s: %w", msg, err)
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
	// We can't use errors.Wrap here to avoid circular dependency
	wrapped := fmt.Errorf("%s: %w", msg, err)
	LogError(wrapped, msg)
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
