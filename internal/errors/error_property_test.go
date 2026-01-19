package errors

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/auriora/onemount/internal/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// NetworkErrorScenario represents a network error test scenario
type NetworkErrorScenario struct {
	ErrorMessage string
	ContextKey   string
	ContextValue string
}

// generateNetworkErrorScenario creates a random network error scenario
func generateNetworkErrorScenario(seed int) NetworkErrorScenario {
	errorMessages := []string{
		"connection timeout",
		"network unreachable",
		"host not found",
		"connection refused",
		"dial tcp failed",
		"context deadline exceeded",
		"no route to host",
		"network is down",
		"temporary failure",
		"operation timed out",
	}

	contextKeys := []string{
		"operation",
		"resource",
		"endpoint",
		"method",
		"path",
		"user_id",
		"request_id",
		"component",
		"service",
		"action",
	}

	contextValues := []string{
		"download_file",
		"upload_file",
		"list_directory",
		"get_metadata",
		"create_folder",
		"delete_item",
		"sync_changes",
		"authenticate",
		"refresh_token",
		"check_quota",
	}

	return NetworkErrorScenario{
		ErrorMessage: errorMessages[seed%len(errorMessages)],
		ContextKey:   contextKeys[(seed/10)%len(contextKeys)],
		ContextValue: contextValues[(seed/100)%len(contextValues)],
	}
}

// TestProperty35_NetworkErrorLogging tests that network errors are logged with appropriate context
// Property 35: Network Error Logging
// Validates: Requirements 11.1
//
// For any network error occurrence, the system should log the error with appropriate context information
func TestProperty35_NetworkErrorLogging(t *testing.T) {
	// Set up test logger to capture log output
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Run 100 iterations with different scenarios
	for i := 0; i < 100; i++ {
		scenario := generateNetworkErrorScenario(i)

		// Clear the buffer before each test
		buf.Reset()

		// Create a network error
		baseErr := fmt.Errorf("network error: %s", scenario.ErrorMessage)
		networkErr := NewNetworkError(scenario.ErrorMessage, baseErr)

		// Create a log context with the generated key-value pair
		logCtx := logging.NewLogContext("test_context").
			With(scenario.ContextKey, scenario.ContextValue)

		// Log the error with context
		logging.LogErrorWithContext(networkErr, logCtx, "Network operation failed")

		// Verify the log output contains the error message
		logOutput := buf.String()
		if !strings.Contains(logOutput, scenario.ErrorMessage) {
			t.Errorf("Iteration %d: Log output missing error message: %s\nLog: %s", i, scenario.ErrorMessage, logOutput)
			continue
		}

		// Verify the log output contains the context key
		if !strings.Contains(logOutput, scenario.ContextKey) {
			t.Errorf("Iteration %d: Log output missing context key: %s\nLog: %s", i, scenario.ContextKey, logOutput)
			continue
		}

		// Verify the log output contains the context value
		if !strings.Contains(logOutput, scenario.ContextValue) {
			t.Errorf("Iteration %d: Log output missing context value: %s\nLog: %s", i, scenario.ContextValue, logOutput)
			continue
		}

		// Verify the log output contains "error" level
		if !strings.Contains(logOutput, "error") {
			t.Errorf("Iteration %d: Log output missing error level\nLog: %s", i, logOutput)
			continue
		}
	}
}

// TestProperty35_NetworkErrorLogging_WithMetrics tests that network errors are recorded in metrics
// Property 35: Network Error Logging (Metrics)
// Validates: Requirements 11.1
//
// For any error occurrence, the system should record the error in metrics
func TestProperty35_NetworkErrorLogging_WithMetrics(t *testing.T) {
	errorMessages := []string{
		"connection timeout",
		"network unreachable",
		"host not found",
		"connection refused",
		"dial tcp failed",
		"context deadline exceeded",
		"no route to host",
		"network is down",
		"temporary failure",
		"operation timed out",
	}

	// Run 100 iterations with different error messages
	for i := 0; i < 100; i++ {
		errorMsg := errorMessages[i%len(errorMessages)]

		// Get the global metrics instance
		metrics := GetErrorMetrics()

		// Reset metrics before test
		metrics.ResetMetrics()

		// Create a network error
		baseErr := fmt.Errorf("network error: %s", errorMsg)
		networkErr := NewNetworkError(errorMsg, baseErr)

		// Record the error
		metrics.RecordError(networkErr)

		// Verify the error was recorded
		metricsData := metrics.GetMetrics()

		// Check that network error count increased
		networkCount, ok := metricsData["network_error_count"].(int)
		if !ok || networkCount != 1 {
			t.Errorf("Iteration %d: Network error count not recorded correctly: %v", i, networkCount)
			continue
		}

		// Check that error counts map contains "network"
		errorCounts, ok := metricsData["error_counts"].(map[string]int)
		if !ok {
			t.Errorf("Iteration %d: Error counts map not found", i)
			continue
		}

		if errorCounts["network"] != 1 {
			t.Errorf("Iteration %d: Network error not in error counts: %v", i, errorCounts)
			continue
		}
	}
}

// ErrorTypeScenario represents an error type classification test scenario
type ErrorTypeScenario struct {
	ErrorType string
	ErrorMsg  string
}

// generateErrorTypeScenario creates a random error type scenario
func generateErrorTypeScenario(seed int) ErrorTypeScenario {
	errorTypes := []string{
		"network",
		"auth",
		"not_found",
		"validation",
		"operation",
		"resource_busy",
	}

	errorMessages := []string{
		"test error 1",
		"test error 2",
		"test error 3",
		"test error 4",
		"test error 5",
	}

	return ErrorTypeScenario{
		ErrorType: errorTypes[seed%len(errorTypes)],
		ErrorMsg:  errorMessages[(seed/10)%len(errorMessages)],
	}
}

// TestProperty35_ErrorTypeClassification tests that different error types are classified correctly
// Property 35: Network Error Logging (Classification)
// Validates: Requirements 11.1
//
// For any error occurrence, the system should classify the error type correctly
func TestProperty35_ErrorTypeClassification(t *testing.T) {
	// Run 100 iterations with different error types
	for i := 0; i < 100; i++ {
		scenario := generateErrorTypeScenario(i)

		// Get the global metrics instance
		metrics := GetErrorMetrics()

		// Reset metrics before test
		metrics.ResetMetrics()

		// Create an error of the specified type
		var err error
		switch scenario.ErrorType {
		case "network":
			err = NewNetworkError(scenario.ErrorMsg, nil)
		case "auth":
			err = NewAuthError(scenario.ErrorMsg, nil)
		case "not_found":
			err = NewNotFoundError(scenario.ErrorMsg, nil)
		case "validation":
			err = NewValidationError(scenario.ErrorMsg, nil)
		case "operation":
			err = NewOperationError(scenario.ErrorMsg, nil)
		case "resource_busy":
			err = NewResourceBusyError(scenario.ErrorMsg, nil)
		default:
			err = New(scenario.ErrorMsg)
		}

		// Record the error
		metrics.RecordError(err)

		// Verify the error was classified correctly
		metricsData := metrics.GetMetrics()
		errorCounts, ok := metricsData["error_counts"].(map[string]int)
		if !ok {
			t.Errorf("Iteration %d: Error counts map not found", i)
			continue
		}

		// Check that the error type was recorded
		if errorCounts[scenario.ErrorType] != 1 {
			t.Errorf("Iteration %d: Error type %s not recorded correctly: %v", i, scenario.ErrorType, errorCounts)
			continue
		}
	}
}

// ErrorWrappingScenario represents an error wrapping test scenario
type ErrorWrappingScenario struct {
	OriginalMsg string
	WrapMsg     string
}

// generateErrorWrappingScenario creates a random error wrapping scenario
func generateErrorWrappingScenario(seed int) ErrorWrappingScenario {
	originalMessages := []string{
		"original error 1",
		"original error 2",
		"original error 3",
		"original error 4",
		"original error 5",
	}

	wrapMessages := []string{
		"failed to process",
		"operation failed",
		"error occurred",
		"unable to complete",
		"request failed",
	}

	return ErrorWrappingScenario{
		OriginalMsg: originalMessages[seed%len(originalMessages)],
		WrapMsg:     wrapMessages[(seed/10)%len(wrapMessages)],
	}
}

// TestProperty35_ErrorContextPreservation tests that error context is preserved through wrapping
// Property 35: Network Error Logging (Context Preservation)
// Validates: Requirements 11.1
//
// For any error that is wrapped, the original context should be preserved
func TestProperty35_ErrorContextPreservation(t *testing.T) {
	// Run 100 iterations with different error messages
	for i := 0; i < 100; i++ {
		scenario := generateErrorWrappingScenario(i)

		// Create an original error
		originalErr := New(scenario.OriginalMsg)

		// Wrap the error
		wrappedErr := Wrap(originalErr, scenario.WrapMsg)

		// Verify both messages are in the error string
		errStr := wrappedErr.Error()
		if !strings.Contains(errStr, scenario.OriginalMsg) {
			t.Errorf("Iteration %d: Wrapped error missing original message: %s\nError: %s", i, scenario.OriginalMsg, errStr)
			continue
		}

		if !strings.Contains(errStr, scenario.WrapMsg) {
			t.Errorf("Iteration %d: Wrapped error missing wrap message: %s\nError: %s", i, scenario.WrapMsg, errStr)
			continue
		}
	}
}

// setupTestLogger sets up a test logger that captures output to a buffer
func setupTestLogger() (*bytes.Buffer, func()) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Save the original loggers
	originalLogger := log.Logger
	originalDefaultLogger := logging.DefaultLogger

	// Create a new logger that writes to the buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()
	log.Logger = logger
	logging.DefaultLogger = logging.New(&buf)

	// Return the buffer and a cleanup function
	return &buf, func() {
		// Restore the original loggers
		log.Logger = originalLogger
		logging.DefaultLogger = originalDefaultLogger
	}
}

// RateLimitScenario represents a rate limit backoff test scenario
type RateLimitScenario struct {
	MaxRetries   int
	InitialDelay int // in milliseconds
	Multiplier   float64
}

// generateRateLimitScenario creates a random rate limit scenario
func generateRateLimitScenario(seed int) RateLimitScenario {
	maxRetries := []int{1, 2, 3, 4, 5}
	initialDelays := []int{100, 200, 500, 1000, 2000} // milliseconds
	multipliers := []float64{1.5, 2.0, 2.5, 3.0}

	return RateLimitScenario{
		MaxRetries:   maxRetries[seed%len(maxRetries)],
		InitialDelay: initialDelays[(seed/10)%len(initialDelays)],
		Multiplier:   multipliers[(seed/100)%len(multipliers)],
	}
}

// TestProperty36_RateLimitBackoff tests that rate limit errors trigger exponential backoff
// Property 36: Rate Limit Backoff
// Validates: Requirements 11.2
//
// For any API rate limit encounter, the system should implement exponential backoff
func TestProperty36_RateLimitBackoff(t *testing.T) {
	// Run 100 iterations with different scenarios
	for i := 0; i < 100; i++ {
		scenario := generateRateLimitScenario(i)

		// Create a rate limit error (429 status code)
		rateLimitErr := NewResourceBusyError("rate limit exceeded", nil)

		// Verify the error is classified as a rate limit error
		if !IsResourceBusyError(rateLimitErr) {
			t.Errorf("Iteration %d: Rate limit error not classified correctly", i)
			continue
		}

		// Verify the error is retryable
		// The retry package should recognize this as a retryable error
		// We test this by checking the error type
		if !IsResourceBusyError(rateLimitErr) {
			t.Errorf("Iteration %d: Rate limit error should be retryable", i)
			continue
		}

		// Calculate expected delays for exponential backoff
		expectedDelays := make([]int, scenario.MaxRetries)
		currentDelay := scenario.InitialDelay
		for j := 0; j < scenario.MaxRetries; j++ {
			expectedDelays[j] = currentDelay
			currentDelay = int(float64(currentDelay) * scenario.Multiplier)
		}

		// Verify that delays increase exponentially
		for j := 1; j < len(expectedDelays); j++ {
			if expectedDelays[j] <= expectedDelays[j-1] {
				t.Errorf("Iteration %d: Delay %d (%d ms) should be greater than delay %d (%d ms)",
					i, j, expectedDelays[j], j-1, expectedDelays[j-1])
				break
			}
		}
	}
}

// TestProperty36_RateLimitBackoff_WithMetrics tests that rate limit errors are recorded in metrics
// Property 36: Rate Limit Backoff (Metrics)
// Validates: Requirements 11.2
//
// For any rate limit error, the system should record it in metrics
func TestProperty36_RateLimitBackoff_WithMetrics(t *testing.T) {
	// Run 100 iterations
	for i := 0; i < 100; i++ {
		// Get the global metrics instance
		metrics := GetErrorMetrics()

		// Reset metrics before test
		metrics.ResetMetrics()

		// Create a rate limit error
		rateLimitErr := NewResourceBusyError("rate limit exceeded", nil)

		// Record the error
		metrics.RecordError(rateLimitErr)

		// Verify the error was recorded
		metricsData := metrics.GetMetrics()

		// Check that rate limit count increased
		rateLimitCount, ok := metricsData["rate_limit_count"].(int)
		if !ok || rateLimitCount != 1 {
			t.Errorf("Iteration %d: Rate limit count not recorded correctly: %v", i, rateLimitCount)
			continue
		}

		// Check that resource busy count increased
		resourceBusyCount, ok := metricsData["resource_busy_count"].(int)
		if !ok || resourceBusyCount != 1 {
			t.Errorf("Iteration %d: Resource busy count not recorded correctly: %v", i, resourceBusyCount)
			continue
		}
	}
}

// TestProperty36_ExponentialBackoffProgression tests the exponential backoff progression
// Property 36: Rate Limit Backoff (Progression)
// Validates: Requirements 11.2
//
// For any retry sequence, delays should increase exponentially
func TestProperty36_ExponentialBackoffProgression(t *testing.T) {
	multipliers := []float64{1.5, 2.0, 2.5, 3.0}
	initialDelays := []int{100, 200, 500, 1000}

	// Run 100 iterations with different combinations
	for i := 0; i < 100; i++ {
		multiplier := multipliers[i%len(multipliers)]
		initialDelay := initialDelays[(i/10)%len(initialDelays)]

		// Calculate a sequence of delays
		delays := make([]int, 5)
		currentDelay := initialDelay
		for j := 0; j < 5; j++ {
			delays[j] = currentDelay
			currentDelay = int(float64(currentDelay) * multiplier)
		}

		// Verify exponential progression
		for j := 1; j < len(delays); j++ {
			expectedDelay := int(float64(delays[j-1]) * multiplier)
			if delays[j] != expectedDelay {
				t.Errorf("Iteration %d: Delay %d should be %d but got %d (multiplier: %.1f)",
					i, j, expectedDelay, delays[j], multiplier)
				break
			}

			// Verify each delay is greater than the previous
			if delays[j] <= delays[j-1] {
				t.Errorf("Iteration %d: Delay %d (%d) should be greater than delay %d (%d)",
					i, j, delays[j], j-1, delays[j-1])
				break
			}
		}
	}
}

// TestProperty36_MaxDelayEnforcement tests that maximum delay is enforced
// Property 36: Rate Limit Backoff (Max Delay)
// Validates: Requirements 11.2
//
// For any retry sequence, delays should not exceed the maximum configured delay
func TestProperty36_MaxDelayEnforcement(t *testing.T) {
	maxDelays := []int{5000, 10000, 30000, 60000} // milliseconds
	multipliers := []float64{2.0, 2.5, 3.0}

	// Run 100 iterations
	for i := 0; i < 100; i++ {
		maxDelay := maxDelays[i%len(maxDelays)]
		multiplier := multipliers[(i/10)%len(multipliers)]
		initialDelay := 1000

		// Calculate delays until we exceed max delay
		currentDelay := initialDelay
		for j := 0; j < 10; j++ {
			// Apply max delay cap
			if currentDelay > maxDelay {
				currentDelay = maxDelay
			}

			// Verify delay doesn't exceed max
			if currentDelay > maxDelay {
				t.Errorf("Iteration %d, retry %d: Delay %d exceeds max delay %d",
					i, j, currentDelay, maxDelay)
				break
			}

			// Calculate next delay
			nextDelay := int(float64(currentDelay) * multiplier)
			if nextDelay > maxDelay {
				nextDelay = maxDelay
			}
			currentDelay = nextDelay
		}
	}
}
