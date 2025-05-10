package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

// setupTestLogger sets up a test logger that writes to a buffer
func setupTestLogger() (*bytes.Buffer, func()) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Save the original logger
	originalLogger := log.Logger

	// Create a new logger that writes to the buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()
	log.Logger = logger

	// Return the buffer and a cleanup function
	return &buf, func() {
		// Restore the original logger
		log.Logger = originalLogger
	}
}

// parseLogEntry parses a JSON log entry from a buffer
func parseLogEntry(buf *bytes.Buffer) (map[string]interface{}, error) {
	var entry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &entry)
	return entry, err
}

// TestUT_SL_01_01_LogContext_WithMethods_AddFields tests the With* methods of LogContext
func TestUT_SL_01_01_LogContext_WithMethods_AddFields(t *testing.T) {
	// Create a new log context
	ctx := NewLogContext("test_operation")

	// Add fields using With* methods
	ctx = ctx.
		WithRequestID("req123").
		WithUserID("user456").
		WithComponent("test_component").
		WithMethod("test_method").
		WithPath("/test/path").
		With("custom_field", "custom_value")

	// Verify that the fields were added
	assert.Equal(t, "req123", ctx.RequestID)
	assert.Equal(t, "user456", ctx.UserID)
	assert.Equal(t, "test_operation", ctx.Operation)
	assert.Equal(t, "test_component", ctx.Component)
	assert.Equal(t, "test_method", ctx.Method)
	assert.Equal(t, "/test/path", ctx.Path)
	assert.Equal(t, "custom_value", ctx.Additional["custom_field"])
}

// TestUT_SL_02_01_LogContext_Logger_IncludesAllFields tests the Logger method of LogContext
func TestUT_SL_02_01_LogContext_Logger_IncludesAllFields(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create a log context with all fields
	ctx := NewLogContext("test_operation").
		WithRequestID("req123").
		WithUserID("user456").
		WithComponent("test_component").
		WithMethod("test_method").
		WithPath("/test/path").
		With("custom_field", "custom_value")

	// Get a logger from the context and log a message
	logger := ctx.Logger()
	logger.Info().Msg("test message")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that all fields are included in the log entry
	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "test message", entry["message"])
	assert.Equal(t, "req123", entry["request_id"])
	assert.Equal(t, "user456", entry["user"])
	assert.Equal(t, "test_operation", entry["operation"])
	assert.Equal(t, "test_component", entry["component"])
	assert.Equal(t, "test_method", entry["method"])
	assert.Equal(t, "/test/path", entry["path"])
	assert.Equal(t, "custom_value", entry["custom_field"])
}

// TestUT_SL_03_01_LogErrorWithContext_IncludesErrorAndContext tests the LogErrorWithContext function
func TestUT_SL_03_01_LogErrorWithContext_IncludesErrorAndContext(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	err := fmt.Errorf("test error")

	// Create a log context
	ctx := NewLogContext("test_operation").
		WithMethod("test_method")

	// Log the error with context
	LogErrorWithContext(err, ctx, "error occurred")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error and context are included in the log entry
	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "error occurred", entry["message"])
	assert.Equal(t, "test error", entry["error"])
	assert.Equal(t, "test_operation", entry["operation"])
	assert.Equal(t, "test_method", entry["method"])
}

// TestUT_SL_03_02_LogErrorWithContext_WithTypedError_IncludesErrorType tests LogErrorWithContext with a typed error
func TestUT_SL_03_02_LogErrorWithContext_WithTypedError_IncludesErrorType(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create a typed error
	err := NewNotFoundError("resource not found", nil)

	// Create a log context
	ctx := NewLogContext("test_operation")

	// Log the error with context
	LogErrorWithContext(err, ctx, "error occurred")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error type is included in the log entry
	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "error occurred", entry["message"])
	assert.Contains(t, entry["error"].(string), "NotFoundError")
	assert.Contains(t, entry["error"].(string), "resource not found")
	assert.Equal(t, "NotFoundError", entry["error_type"])
	assert.Equal(t, float64(404), entry["status_code"])
}

// TestUT_SL_04_01_LogWarnWithContext_IncludesErrorAndContext tests the LogWarnWithContext function
func TestUT_SL_04_01_LogWarnWithContext_IncludesErrorAndContext(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	err := fmt.Errorf("test warning")

	// Create a log context
	ctx := NewLogContext("test_operation").
		WithMethod("test_method")

	// Log the warning with context
	LogWarnWithContext(err, ctx, "warning occurred")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error and context are included in the log entry
	assert.Equal(t, "warn", entry["level"])
	assert.Equal(t, "warning occurred", entry["message"])
	assert.Equal(t, "test warning", entry["error"])
	assert.Equal(t, "test_operation", entry["operation"])
	assert.Equal(t, "test_method", entry["method"])
}

// TestUT_SL_05_01_LogInfoWithContext_IncludesContext tests the LogInfoWithContext function
func TestUT_SL_05_01_LogInfoWithContext_IncludesContext(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create a log context
	ctx := NewLogContext("test_operation").
		WithMethod("test_method")

	// Log an info message with context
	LogInfoWithContext(ctx, "info message")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the context is included in the log entry
	assert.Equal(t, "info", entry["level"])
	assert.Equal(t, "info message", entry["message"])
	assert.Equal(t, "test_operation", entry["operation"])
	assert.Equal(t, "test_method", entry["method"])
}

// TestUT_SL_06_01_WrapAndLogWithContext_WrapsAndLogsError tests the WrapAndLogWithContext function
func TestUT_SL_06_01_WrapAndLogWithContext_WrapsAndLogsError(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	originalErr := fmt.Errorf("original error")

	// Create a log context
	ctx := NewLogContext("test_operation")

	// Wrap and log the error with context
	wrappedErr := WrapAndLogWithContext(originalErr, ctx, "wrapped error")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error was logged with context
	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "wrapped error", entry["message"])
	assert.Contains(t, entry["error"].(string), "wrapped error")
	assert.Contains(t, entry["error"].(string), "original error")
	assert.Equal(t, "test_operation", entry["operation"])

	// Verify that the error was wrapped
	assert.Contains(t, wrappedErr.Error(), "wrapped error")
	assert.Contains(t, wrappedErr.Error(), "original error")
	assert.True(t, Is(wrappedErr, originalErr))
}

// TestUT_SL_07_01_LogAndReturnWithContext_LogsAndReturnsError tests the LogAndReturnWithContext function
func TestUT_SL_07_01_LogAndReturnWithContext_LogsAndReturnsError(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	originalErr := fmt.Errorf("original error")

	// Create a log context
	ctx := NewLogContext("test_operation")

	// Log and return the error with context
	returnedErr := LogAndReturnWithContext(originalErr, ctx, "error message")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error was logged with context
	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "error message", entry["message"])
	assert.Equal(t, "original error", entry["error"])
	assert.Equal(t, "test_operation", entry["operation"])

	// Verify that the original error was returned
	assert.Equal(t, originalErr, returnedErr)
}

// TestUT_SL_08_01_EnrichErrorWithContext_AddsContextToError tests the EnrichErrorWithContext function
func TestUT_SL_08_01_EnrichErrorWithContext_AddsContextToError(t *testing.T) {
	// Create an error
	originalErr := fmt.Errorf("original error")

	// Create a log context with all fields
	ctx := NewLogContext("test_operation").
		WithMethod("test_method").
		WithPath("/test/path")

	// Enrich the error with context
	enrichedErr := EnrichErrorWithContext(originalErr, ctx, "enriched error")

	// Verify that the error message includes the context
	assert.Contains(t, enrichedErr.Error(), "enriched error")
	assert.Contains(t, enrichedErr.Error(), "operation: test_operation")
	assert.Contains(t, enrichedErr.Error(), "method: test_method")
	assert.Contains(t, enrichedErr.Error(), "path: /test/path")
	assert.Contains(t, enrichedErr.Error(), "original error")

	// Verify that the original error is preserved in the error chain
	assert.True(t, Is(enrichedErr, originalErr))
}
