// Package logging provides standardized logging utilities for the OneMount project.
// This file provides utilities for generating and managing request IDs and user IDs.
//
// Request IDs are used to track operations that span multiple functions or goroutines.
// They provide a way to correlate log entries from different parts of the codebase
// that are part of the same logical operation.
//
// User IDs are used to identify the user who initiated an operation.
//
// This file is part of the consolidated logging package structure, which includes:
//   - logger.go: Core logger implementation and level management
//   - context.go: Context-aware logging functionality
//   - method.go: Method entry/exit logging (both with and without context)
//   - error.go: Error logging functionality
//   - performance.go: Performance optimization utilities
//   - constants.go: Constants used throughout the logging package
//   - request_id.go (this file): Utilities for generating and managing request IDs and user IDs
package logging

import (
	"fmt"
	"math/rand"
	"os/user"
	"sync/atomic"
	"time"
)

// Counter for generating unique request IDs
var requestIDCounter uint64

// GenerateRequestID generates a unique request ID.
// The ID is a combination of a timestamp and a counter to ensure uniqueness.
// Format: <timestamp>-<counter>-<random>
func GenerateRequestID() string {
	// Get the current timestamp in milliseconds
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Increment the counter atomically
	counter := atomic.AddUint64(&requestIDCounter, 1)

	// Add a random component to further ensure uniqueness
	random := rand.Intn(10000)

	// Format: <timestamp>-<counter>-<random>
	return fmt.Sprintf("%d-%d-%d", timestamp, counter, random)
}

// GetCurrentUserID returns the username of the current user.
// This can be used as a user ID for logging user-initiated operations.
// If the username cannot be determined, it returns "unknown".
func GetCurrentUserID() string {
	currentUser, err := user.Current()
	if err != nil {
		// If we can't get the current user, return a default value
		return "unknown"
	}
	return currentUser.Username
}

// NewLogContextWithRequestID creates a new LogContext with a unique request ID and the given operation.
// This is a convenience function for creating a LogContext with a request ID in one step.
func NewLogContextWithRequestID(operation string) LogContext {
	return NewLogContext(operation).WithRequestID(GenerateRequestID())
}

// NewLogContextWithRequestAndUserID creates a new LogContext with a unique request ID,
// the current user's ID, and the given operation.
// This is a convenience function for creating a LogContext with both request ID and user ID in one step.
func NewLogContextWithRequestAndUserID(operation string) LogContext {
	return NewLogContext(operation).
		WithRequestID(GenerateRequestID()).
		WithUserID(GetCurrentUserID())
}
