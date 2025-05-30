// Package logging provides standardized logging utilities for the OneMount project.
// This file defines structured logging functions for errors.
//
// Structured logging with context is a powerful way to add consistent contextual
// information to log entries. This file provides functions for logging at different
// levels (error, warn, info, debug, trace) with context, as well as utilities for
// enriching errors with context information.
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
//   - constants.go: Constants used throughout the logging package
//   - console_writer.go: Console writer functionality
//   - structured_logging.go (this file): Structured logging functions
package logging

import (
	"fmt"
)

// LogErrorAsWarnWithContext logs an error as a warning with the given context
func LogErrorAsWarnWithContext(err error, ctx LogContext, msg string) {
	if err == nil {
		return
	}

	// Check if warn level is enabled before performing operations
	if !IsLevelEnabled(WarnLevel) {
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
	// Check if info level is enabled before performing operations
	if !IsLevelEnabled(InfoLevel) {
		return
	}

	// Get the logger with context
	logger := ctx.Logger()

	// Log the message
	logger.Info().Msg(msg)
}

// LogDebugWithContext logs a debug message with the given context
func LogDebugWithContext(ctx LogContext, msg string) {
	// Check if debug level is enabled before performing operations
	if !IsLevelEnabled(DebugLevel) {
		return
	}

	// Get the logger with context
	logger := ctx.Logger()

	// Log the message
	logger.Debug().Msg(msg)
}

// LogTraceWithContext logs a trace message with the given context
func LogTraceWithContext(ctx LogContext, msg string) {
	// Check if trace level is enabled before performing operations
	if !IsLevelEnabled(TraceLevel) {
		return
	}

	// Get the logger with context
	logger := ctx.Logger()

	// Log the message
	logger.Trace().Msg(msg)
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
