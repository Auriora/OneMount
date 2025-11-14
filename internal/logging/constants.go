// Package logging provides standardized logging utilities for the OneMount project.
// This file defines constants used throughout the logging package.
//
// Consistent field names and message templates are essential for structured logging.
// This file centralizes all constants used in the logging package, including:
//
//   - Standard field names for common concepts (method, operation, component, etc.)
//   - Method logging specific fields (return, param)
//   - Additional field names for structured logging
//   - Phase values (entry, exit)
//   - Message templates
//
// Using these constants ensures consistency across the codebase and makes logs
// easier to parse and analyze.
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
//   - constants.go (this file): Constants used throughout the logging package
package logging

// HumanReadableTimeFormat is the default time layout (YYYY-MM-DD HH:MM:SS) used by console logs.
const HumanReadableTimeFormat = "2006-01-02 15:04:05"

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
	FieldRequestID = "request_id"  // Request identifier for operations spanning multiple functions
	FieldStatus    = "status"      // Status code or string
	FieldSize      = "size"        // Size in bytes
	FieldGoroutine = "goroutine"   // Goroutine ID
	FieldPhase     = "phase"       // Phase of operation (e.g., "entry", "exit")

	// Method logging specific fields
	FieldReturn = "return" // Return value prefix
	FieldParam  = "param"  // Parameter prefix

	// Additional field names for structured logging
	FieldOffset      = "offset"       // Offset in bytes
	FieldCount       = "count"        // Count of items
	FieldRetries     = "retries"      // Number of retries
	FieldStatusCode  = "status_code"  // HTTP status code
	FieldContentType = "content_type" // Content type
	FieldURL         = "url"          // URL
	FieldRequest     = "request"      // Request details
	FieldResponse    = "response"     // Response details
	FieldSource      = "source"       // Source of the error
	FieldTarget      = "target"       // Target of the operation

	// Phase values
	PhaseEntry = "entry" // Method entry
	PhaseExit  = "exit"  // Method exit

	// Message templates
	MsgMethodCalled    = "Method called"
	MsgMethodCompleted = "Method completed"
)
