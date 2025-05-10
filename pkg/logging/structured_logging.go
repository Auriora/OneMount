// Package logging provides standardized logging utilities for the OneMount project.
// This file defines structured logging functions for errors.
package logging

import (
	"fmt"
)

// Note: Field constants are now defined in constants.go
// Note: LogContext and related methods are now defined in context.go
// Note: LogErrorWithContext is now defined in method_logging_context.go

// LogWarnWithContext logs a warning with the given context
func LogWarnWithContext(err error, ctx LogContext, msg string) {
	if err == nil {
		return
	}

	// Get the logger with context
	logger := ctx.Logger()

	// Create the warning event
	event := logger.Warn().Err(err)

	// Log the message
	event.Msg(msg)
}

// LogInfoWithContext logs an info message with the given context
func LogInfoWithContext(ctx LogContext, msg string) {
	// Get the logger with context
	logger := ctx.Logger()

	// Log the message
	logger.Info().Msg(msg)
}

// LogDebugWithContext logs a debug message with the given context
func LogDebugWithContext(ctx LogContext, msg string) {
	// Get the logger with context
	logger := ctx.Logger()

	// Log the message
	logger.Debug().Msg(msg)
}

// LogTraceWithContext logs a trace message with the given context
func LogTraceWithContext(ctx LogContext, msg string) {
	// Get the logger with context
	logger := ctx.Logger()

	// Log the message
	logger.Trace().Msg(msg)
}

// WrapAndLogWithContext wraps an error, logs it with context, and returns the wrapped error
func WrapAndLogWithContext(err error, ctx LogContext, msg string) error {
	if err == nil {
		return nil
	}

	wrapped := fmt.Errorf("%s: %w", msg, err)
	LogErrorWithContext(wrapped, ctx, msg)
	return wrapped
}

// LogAndReturnWithContext logs an error with context and returns it
func LogAndReturnWithContext(err error, ctx LogContext, msg string) error {
	if err == nil {
		return nil
	}

	LogErrorWithContext(err, ctx, msg)
	return err
}

// EnrichErrorWithContext adds context information to an error
// This is useful when you want to add context to an error without logging it
func EnrichErrorWithContext(err error, ctx LogContext, msg string) error {
	if err == nil {
		return nil
	}

	// Create a message that includes context information
	contextMsg := msg

	// Add operation if available
	if ctx.Operation != "" {
		contextMsg += " (operation: " + ctx.Operation + ")"
	}

	// Add method if available
	if ctx.Method != "" {
		contextMsg += " (method: " + ctx.Method + ")"
	}

	// Add path if available
	if ctx.Path != "" {
		contextMsg += " (path: " + ctx.Path + ")"
	}

	return fmt.Errorf("%s: %w", contextMsg, err)
}
