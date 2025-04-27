package fs

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Standard field names for logging
const (
	// Common field names
	FieldMethod    = "method"      // Method or function name
	FieldOperation = "operation"   // Higher-level operation
	FieldComponent = "component"   // Component or module
	FieldDuration  = "duration_ms" // Duration of operation in milliseconds
	FieldError     = "error"       // Error message
	FieldPath      = "path"        // File or resource path
	FieldID        = "id"          // Identifier
	FieldUser      = "user"        // User identifier
	FieldStatus    = "status"      // Status code or string
	FieldSize      = "size"        // Size in bytes
	FieldGoroutine = "goroutine"   // Goroutine ID
	FieldPhase     = "phase"       // Phase of operation (e.g., "entry", "exit")

	// Method logging specific fields
	FieldReturn = "return" // Return value prefix
	FieldParam  = "param"  // Parameter prefix

	// Phase values
	PhaseEntry = "entry" // Method entry
	PhaseExit  = "exit"  // Method exit

	// Message templates
	MsgMethodCalled    = "Method called"
	MsgMethodCompleted = "Method completed"
)

// LogContext represents a logging context that can be passed between functions
type LogContext struct {
	RequestID string
	UserID    string
	Operation string
	// Add other fields as needed
}

// WithLogContext creates a new zerolog.Logger with the given context
func WithLogContext(ctx LogContext) zerolog.Logger {
	logger := log.With()

	if ctx.RequestID != "" {
		logger = logger.Str("request_id", ctx.RequestID)
	}

	if ctx.UserID != "" {
		logger = logger.Str(FieldUser, ctx.UserID)
	}

	if ctx.Operation != "" {
		logger = logger.Str(FieldOperation, ctx.Operation)
	}

	return logger.Logger()
}

// LogMethodCallWithContext logs the entry of a method with context
func LogMethodCallWithContext(methodName string, ctx LogContext) (string, zerolog.Logger, LogContext) {
	// Create a logger with the context
	logger := WithLogContext(ctx)

	// Get the current goroutine ID
	goroutineID := getCurrentGoroutineID()

	// Log method entry
	logger.Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseEntry).
		Str(FieldGoroutine, goroutineID).
		Msg(MsgMethodCalled)

	return methodName, logger, ctx
}

// LogMethodReturnWithContext logs the exit of a method with context
func LogMethodReturnWithContext(methodName string, startTime, logger zerolog.Logger, ctx LogContext, returns ...interface{}) {
	duration := time.Since(startTime)

	// Get the current goroutine ID
	goroutineID := getCurrentGoroutineID()

	// Create log event
	event := logger.Debug().
		Str(FieldMethod, methodName).
		Str(FieldPhase, PhaseExit).
		Str(FieldGoroutine, goroutineID).
		Dur(FieldDuration, duration)

	// Log return values if any
	for i, ret := range returns {
		if ret == nil {
			event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), nil)
		} else {
			event = event.Interface(FieldReturn+fmt.Sprintf("%d", i+1), ret)
		}
	}

	event.Msg(MsgMethodCompleted)
}

// LogError logs an error with context
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
