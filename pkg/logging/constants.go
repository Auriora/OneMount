// Package logging provides standardized logging utilities for the OneMount project.
// This file defines constants used throughout the logging package.
package logging

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
