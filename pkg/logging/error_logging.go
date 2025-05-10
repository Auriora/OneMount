// Package logging provides standardized logging utilities for the OneMount project.
// This file defines error logging functions.
package logging

import (
	"fmt"
	"github.com/rs/zerolog/log"
)

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
