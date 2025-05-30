// Package errors provides error handling utilities for the OneMount project.
package errors

import (
	"sync"
	"time"

	"github.com/auriora/onemount/pkg/logging"
)

// ErrorMetrics tracks error metrics for monitoring purposes
type ErrorMetrics struct {
	// Total number of errors by type
	ErrorCounts map[string]int
	// Number of rate limit errors
	RateLimitCount int
	// Number of network errors
	NetworkErrorCount int
	// Number of authentication errors
	AuthErrorCount int
	// Number of not found errors
	NotFoundErrorCount int
	// Number of validation errors
	ValidationErrorCount int
	// Number of operation errors
	OperationErrorCount int
	// Number of resource busy errors
	ResourceBusyErrorCount int
	// Number of errors by status code
	StatusCodeCounts map[int]int
	// Last error time by type
	LastErrorTime map[string]time.Time
	// Error rate (errors per minute) by type
	ErrorRates map[string]float64

	// Mutex for thread-safe access
	mu sync.RWMutex
}

var (
	// Global error metrics instance
	globalMetrics     *ErrorMetrics
	globalMetricsOnce sync.Once
)

// GetErrorMetrics returns the global error metrics instance
func GetErrorMetrics() *ErrorMetrics {
	globalMetricsOnce.Do(func() {
		globalMetrics = &ErrorMetrics{
			ErrorCounts:      make(map[string]int),
			StatusCodeCounts: make(map[int]int),
			LastErrorTime:    make(map[string]time.Time),
			ErrorRates:       make(map[string]float64),
		}
		// Start a goroutine to periodically log error metrics
		go globalMetrics.monitorErrorRates()
	})
	return globalMetrics
}

// RecordError records an error for monitoring purposes
func (m *ErrorMetrics) RecordError(err error) {
	if err == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Determine error type
	errorType := "unknown"
	switch {
	case IsNetworkError(err):
		errorType = "network"
		m.NetworkErrorCount++
	case IsAuthError(err):
		errorType = "auth"
		m.AuthErrorCount++
	case IsNotFoundError(err):
		errorType = "not_found"
		m.NotFoundErrorCount++
	case IsValidationError(err):
		errorType = "validation"
		m.ValidationErrorCount++
	case IsOperationError(err):
		errorType = "operation"
		m.OperationErrorCount++
	case IsResourceBusyError(err):
		errorType = "resource_busy"
		m.ResourceBusyErrorCount++
		m.RateLimitCount++
	}

	// Update error counts
	m.ErrorCounts[errorType]++

	// Update last error time
	m.LastErrorTime[errorType] = time.Now()

	// Extract status code if available
	if typedErr, ok := err.(*TypedError); ok && typedErr.StatusCode > 0 {
		m.StatusCodeCounts[typedErr.StatusCode]++
	}
}

// monitorErrorRates periodically calculates and logs error rates
func (m *ErrorMetrics) monitorErrorRates() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.calculateErrorRates()
		m.logErrorMetrics()
	}
}

// calculateErrorRates calculates error rates based on error counts and time
func (m *ErrorMetrics) calculateErrorRates() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for errorType, lastTime := range m.LastErrorTime {
		count := m.ErrorCounts[errorType]
		duration := now.Sub(lastTime).Minutes()
		if duration > 0 && count > 0 {
			// Calculate errors per minute
			m.ErrorRates[errorType] = float64(count) / duration
		}
	}
}

// logErrorMetrics logs current error metrics
func (m *ErrorMetrics) logErrorMetrics() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Log overall error counts
	logging.Info().
		Int("total_errors", sumMapValues(m.ErrorCounts)).
		Int("network_errors", m.NetworkErrorCount).
		Int("auth_errors", m.AuthErrorCount).
		Int("not_found_errors", m.NotFoundErrorCount).
		Int("validation_errors", m.ValidationErrorCount).
		Int("operation_errors", m.OperationErrorCount).
		Int("resource_busy_errors", m.ResourceBusyErrorCount).
		Int("rate_limit_errors", m.RateLimitCount).
		Msg("Error metrics summary")

	// Log error rates
	for errorType, rate := range m.ErrorRates {
		logging.Info().
			Str("error_type", errorType).
			Float64("errors_per_minute", rate).
			Msg("Error rate")
	}

	// Log status code distribution
	if len(m.StatusCodeCounts) > 0 {
		logging.Info().
			Interface("status_code_counts", m.StatusCodeCounts).
			Msg("Error status code distribution")
	}
}

// GetMetrics returns a copy of the current error metrics
func (m *ErrorMetrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"error_counts":           m.ErrorCounts,
		"network_error_count":    m.NetworkErrorCount,
		"auth_error_count":       m.AuthErrorCount,
		"not_found_error_count":  m.NotFoundErrorCount,
		"validation_error_count": m.ValidationErrorCount,
		"operation_error_count":  m.OperationErrorCount,
		"resource_busy_count":    m.ResourceBusyErrorCount,
		"rate_limit_count":       m.RateLimitCount,
		"status_code_counts":     m.StatusCodeCounts,
		"error_rates":            m.ErrorRates,
	}
}

// ResetMetrics resets all error metrics
func (m *ErrorMetrics) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ErrorCounts = make(map[string]int)
	m.NetworkErrorCount = 0
	m.AuthErrorCount = 0
	m.NotFoundErrorCount = 0
	m.ValidationErrorCount = 0
	m.OperationErrorCount = 0
	m.ResourceBusyErrorCount = 0
	m.RateLimitCount = 0
	m.StatusCodeCounts = make(map[int]int)
	m.LastErrorTime = make(map[string]time.Time)
	m.ErrorRates = make(map[string]float64)
}

// sumMapValues returns the sum of all values in a map[string]int
func sumMapValues(m map[string]int) int {
	sum := 0
	for _, v := range m {
		sum += v
	}
	return sum
}

// MonitorError records an error for monitoring purposes
func MonitorError(err error) {
	if err == nil {
		return
	}

	// Record the error in the global metrics
	metrics := GetErrorMetrics()
	metrics.RecordError(err)
}

// WrapAndMonitor wraps an error and records it for monitoring
func WrapAndMonitor(err error, message string) error {
	if err == nil {
		return nil
	}

	// Wrap the error
	wrappedErr := Wrap(err, message)

	// Record the error
	MonitorError(wrappedErr)

	return wrappedErr
}
