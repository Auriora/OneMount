package logging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

// setupTestLogger sets up a test logger that writes to a buffer
func setupTestLogger() (*bytes.Buffer, func()) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Save the original loggers
	originalLogger := log.Logger
	originalDefaultLogger := DefaultLogger

	// Create a new logger that writes to the buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()
	log.Logger = logger
	DefaultLogger = Logger{zl: logger}

	// Return the buffer and a cleanup function
	return &buf, func() {
		// Restore the original loggers
		log.Logger = originalLogger
		DefaultLogger = originalDefaultLogger
	}
}

// parseLogEntry parses a JSON log entry from a buffer
func parseLogEntry(buf *bytes.Buffer) (map[string]interface{}, error) {
	var entry map[string]interface{}

	// Create a copy of the buffer to avoid consuming it
	bufCopy := bytes.NewBuffer(buf.Bytes())

	// Create a decoder to read JSON objects one at a time
	decoder := json.NewDecoder(bufCopy)

	// Read the first JSON object
	if err := decoder.Decode(&entry); err != nil {
		return nil, err
	}

	return entry, nil
}

// parseLogEntries parses multiple JSON log entries from a buffer
func parseLogEntries(buf *bytes.Buffer) ([]map[string]interface{}, error) {
	var entries []map[string]interface{}

	// Create a copy of the buffer to avoid consuming it
	bufCopy := bytes.NewBuffer(buf.Bytes())

	// Create a decoder to read JSON objects one at a time
	decoder := json.NewDecoder(bufCopy)

	// Read all JSON objects
	for {
		var entry map[string]interface{}
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
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
	assert.Equal(t, "test_method", entry["method_name"])
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
	assert.Equal(t, "test_method", entry["method_name"])
}

// TestUT_SL_03_02_LogErrorWithContext_WithCustomError_IncludesErrorMessage tests LogErrorWithContext with a custom error
func TestUT_SL_03_02_LogErrorWithContext_WithCustomError_IncludesErrorMessage(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create a custom error
	err := fmt.Errorf("resource not found: %w", errors.New("not found error"))

	// Create a log context
	ctx := NewLogContext("test_operation")

	// Log the error with context
	LogErrorWithContext(err, ctx, "error occurred")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error message is included in the log entry
	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "error occurred", entry["message"])
	assert.Contains(t, entry["error"].(string), "resource not found")
	assert.Contains(t, entry["error"].(string), "not found error")
}

// TestUT_SL_04_01_LogErrorAsWarnWithContext_IncludesErrorAndContext tests the LogErrorAsWarnWithContext function
func TestUT_SL_04_01_LogErrorAsWarnWithContext_IncludesErrorAndContext(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	err := fmt.Errorf("test warning")

	// Create a log context
	ctx := NewLogContext("test_operation").
		WithMethod("test_method")

	// Log the warning with context
	LogErrorAsWarnWithContext(err, ctx, "warning occurred")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error and context are included in the log entry
	assert.Equal(t, "warn", entry["level"])
	assert.Equal(t, "warning occurred", entry["message"])
	assert.Equal(t, "test warning", entry["error"])
	assert.Equal(t, "test_operation", entry["operation"])
	assert.Equal(t, "test_method", entry["method_name"])
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
	assert.Equal(t, "test_method", entry["method_name"])
}

// TestUT_SL_06_01_WrapAndLogErrorWithContext_WrapsAndLogsError tests the WrapAndLogErrorWithContext function
func TestUT_SL_06_01_WrapAndLogErrorWithContext_WrapsAndLogsError(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	originalErr := fmt.Errorf("original error")

	// Create a log context
	ctx := NewLogContext("test_operation")

	// Wrap and log the error with context
	wrappedErr := WrapAndLogErrorWithContext(originalErr, ctx, "wrapped error")

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
	assert.True(t, errors.Is(wrappedErr, originalErr))
}

// TestUT_SL_07_01_LogErrorWithContext_LogsError tests the LogErrorWithContext function
// This demonstrates the recommended approach of using LogErrorWithContext and returning the error separately.
func TestUT_SL_07_01_LogErrorWithContext_LogsError(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	originalErr := fmt.Errorf("original error")

	// Create a log context
	ctx := NewLogContext("test_operation")

	// Log the error with context and return it separately
	LogErrorWithContext(originalErr, ctx, "error message")
	returnedErr := originalErr

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
	assert.True(t, errors.Is(enrichedErr, originalErr))
}

// TestUT_SL_09_01_LogErrorAsWarn_IncludesError tests the LogErrorAsWarn function
func TestUT_SL_09_01_LogErrorAsWarn_IncludesError(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	err := fmt.Errorf("test warning")

	// Log the error as a warning
	LogErrorAsWarn(err, "warning occurred")

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error is included in the log entry
	assert.Equal(t, "warn", entry["level"])
	assert.Equal(t, "warning occurred", entry["message"])
	assert.Equal(t, "test warning", entry["error"])
}

// TestUT_SL_09_02_LogErrorAsWarnWithFields_IncludesErrorAndFields tests the LogErrorAsWarnWithFields function
func TestUT_SL_09_02_LogErrorAsWarnWithFields_IncludesErrorAndFields(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	err := fmt.Errorf("test warning")

	// Create fields
	fields := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
	}

	// Log the error as a warning with fields
	LogErrorAsWarnWithFields(err, "warning with fields", fields)

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error and fields are included in the log entry
	assert.Equal(t, "warn", entry["level"])
	assert.Equal(t, "warning with fields", entry["message"])
	assert.Equal(t, "test warning", entry["error"])
	assert.Equal(t, "value1", entry["field1"])
	assert.Equal(t, float64(42), entry["field2"])
}

// TestUT_SL_09_03_WrapAndLogErrorf_WrapsAndLogsError tests the WrapAndLogErrorf function
func TestUT_SL_09_03_WrapAndLogErrorf_WrapsAndLogsError(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create an error
	originalErr := fmt.Errorf("original error")

	// Wrap and log the error with a formatted message
	wrappedErr := WrapAndLogErrorf(originalErr, "wrapped error with value %d", 42)

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify that the error was logged with the formatted message
	assert.Equal(t, "error", entry["level"])
	assert.Equal(t, "wrapped error with value 42", entry["message"])
	assert.Contains(t, entry["error"].(string), "wrapped error with value 42")
	assert.Contains(t, entry["error"].(string), "original error")

	// Verify that the error was wrapped with the formatted message
	assert.Contains(t, wrappedErr.Error(), "wrapped error with value 42")
	assert.Contains(t, wrappedErr.Error(), "original error")
	assert.True(t, errors.Is(wrappedErr, originalErr))
}
