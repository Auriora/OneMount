// Package logging provides standardized logging utilities for the OneMount project.
// This file defines the LogContext struct and related methods for context-based logging.
//
// The LogContext struct allows for consistent logging of contextual information across
// multiple function calls. It provides a fluent interface for building context with
// common fields like request ID, user ID, operation, component, method, and path.
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go (this file): Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
package logging

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

// Logger returns a Logger with the context fields added
func (lc LogContext) Logger() Logger {
	logger := DefaultLogger.With()

	if lc.RequestID != "" {
		logger = logger.Str(FieldRequestID, lc.RequestID)
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
		logger = logger.Str("method_name", lc.Method)
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

// WithLogContext creates a new Logger with the given context
func WithLogContext(ctx LogContext) Logger {
	logger := DefaultLogger.With()

	if ctx.RequestID != "" {
		logger = logger.Str(FieldRequestID, ctx.RequestID)
	}

	if ctx.UserID != "" {
		logger = logger.Str(FieldUser, ctx.UserID)
	}

	if ctx.Operation != "" {
		logger = logger.Str(FieldOperation, ctx.Operation)
	}

	if ctx.Component != "" {
		logger = logger.Str(FieldComponent, ctx.Component)
	}

	if ctx.Method != "" {
		logger = logger.Str("method_name", ctx.Method)
	}

	if ctx.Path != "" {
		logger = logger.Str(FieldPath, ctx.Path)
	}

	// Add any additional fields
	for k, v := range ctx.Additional {
		logger = logger.Interface(k, v)
	}

	return logger.Logger()
}
