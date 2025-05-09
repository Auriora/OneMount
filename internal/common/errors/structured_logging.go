// Package errors provides standardized error handling utilities for the OneMount project.
// This file defines structured logging functions for errors.
package errors

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Additional field names for structured logging
const (
	FieldMethod      = "method"      // Method or function name
	FieldDuration    = "duration_ms" // Duration of operation in milliseconds
	FieldGoroutine   = "goroutine"   // Goroutine ID
	FieldPhase       = "phase"       // Phase of operation (e.g., "entry", "exit")
	FieldSize        = "size"        // Size in bytes
	FieldOffset      = "offset"      // Offset in bytes
	FieldCount       = "count"       // Count of items
	FieldRetries     = "retries"     // Number of retries
	FieldStatusCode  = "status_code" // HTTP status code
	FieldContentType = "content_type" // Content type
	FieldURL         = "url"         // URL
	FieldRequest     = "request"     // Request details
	FieldResponse    = "response"    // Response details
	FieldSource      = "source"      // Source of the error
	FieldTarget      = "target"      // Target of the operation
)

// LogContext represents a logging context that can be passed between functions
type LogContext struct {
	RequestID  string
	UserID     string
	Operation  string
	Component  string
	Method     string
	Path       string
	Additional map[string]interface{}
}

// NewLogContext creates a new LogContext with the given operation
func NewLogContext(operation string) LogContext {
	return LogContext{
		Operation:  operation,
		Additional: make(map[string]interface{}),
	}
}

// WithRequestID adds a request ID to the log context
func (lc LogContext) WithRequestID(requestID string) LogContext {
	lc.RequestID = requestID
	return lc
}

// WithUserID adds a user ID to the log context
func (lc LogContext) WithUserID(userID string) LogContext {
	lc.UserID = userID
	return lc
}

// WithComponent adds a component to the log context
func (lc LogContext) WithComponent(component string) LogContext {
	lc.Component = component
	return lc
}

// WithMethod adds a method to the log context
func (lc LogContext) WithMethod(method string) LogContext {
	lc.Method = method
	return lc
}

// WithPath adds a path to the log context
func (lc LogContext) WithPath(path string) LogContext {
	lc.Path = path
	return lc
}

// With adds a custom field to the log context
func (lc LogContext) With(key string, value interface{}) LogContext {
	lc.Additional[key] = value
	return lc
}

// Logger returns a zerolog.Logger with the context fields added
func (lc LogContext) Logger() zerolog.Logger {
	logger := log.With()

	if lc.RequestID != "" {
		logger = logger.Str("request_id", lc.RequestID)
	}

	if lc.UserID != "" {
		logger = logger.Str(FieldUser, lc.UserID)
	}

	if lc.Operation != "" {
		logger = logger.Str(FieldOperation, lc.Operation)
	}

	if lc.Component != "" {
		logger = logger.Str(FieldComponent, lc.Component)
	}

	if lc.Method != "" {
		logger = logger.Str(FieldMethod, lc.Method)
	}

	if lc.Path != "" {
		logger = logger.Str(FieldPath, lc.Path)
	}

	// Add any additional fields
	for k, v := range lc.Additional {
		logger = logger.Interface(k, v)
	}

	return logger.Logger()
}

// LogErrorWithContext logs an error with the given context
func LogErrorWithContext(err error, ctx LogContext, msg string) {
	if err == nil {
		return
	}

	// Get the logger with context
	logger := ctx.Logger()

	// Create the error event
	event := logger.Error().Err(err)

	// Add error type information if it's a TypedError
	var typedErr *TypedError
	if As(err, &typedErr) {
		event = event.
			Str("error_type", typedErr.Type.String()).
			Int("status_code", typedErr.StatusCode)
	}

	// Log the message
	event.Msg(msg)
}

// LogWarnWithContext logs a warning with the given context
func LogWarnWithContext(err error, ctx LogContext, msg string) {
	if err == nil {
		return
	}

	// Get the logger with context
	logger := ctx.Logger()

	// Create the warning event
	event := logger.Warn().Err(err)

	// Add error type information if it's a TypedError
	var typedErr *TypedError
	if As(err, &typedErr) {
		event = event.
			Str("error_type", typedErr.Type.String()).
			Int("status_code", typedErr.StatusCode)
	}

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

	wrapped := Wrap(err, msg)
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

	return Wrap(err, contextMsg)
}