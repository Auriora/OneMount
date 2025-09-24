package logging

import (
	"testing"
	"time"

	"github.com/auriora/onemount/internal/util"
	"github.com/stretchr/testify/assert"
)

// TestUT_FS_05_01_Logging_MethodEntryAndExit_LogsCorrectly tests the LogMethodEntry and LogMethodExit functions.
//
//	Test Case ID    UT-FS-05-01
//	Title           Method Logging
//	Description     Tests the LogMethodEntry and LogMethodExit functions
//	Preconditions   None
//	Steps           1. Call LogMethodEntry
//	                2. Call LogMethodExit with different types of return values
//	                3. Verify the log output contains the expected information
//	Expected Result Logging functions correctly log method entry and exit
//	Notes: This test verifies that the logging functions correctly log method entry and exit.
func TestUT_FS_05_01_Logging_MethodEntryAndExit_LogsCorrectly(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// 1. Call LogMethodEntry
	methodName := "TestMethod"
	methodName, startTime := LogMethodEntry(methodName, "param1", 42)

	// Parse the log entry
	entry, err := parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify the log entry for method entry
	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, methodName, entry["method"])
	assert.Equal(t, PhaseEntry, entry["phase"])
	assert.Equal(t, MsgMethodCalled, entry["message"])
	assert.NotEmpty(t, entry["goroutine"])

	// Clear the buffer for the next log entry
	buf.Reset()

	// 2. Call LogMethodExit with different types of return values
	returnValues := []interface{}{
		"result1",
		42,
		true,
		nil,
		struct{ Name string }{"test"},
	}
	LogMethodExit(methodName, time.Since(startTime), returnValues...)

	// Parse the log entry
	entry, err = parseLogEntry(buf)
	assert.NoError(t, err)

	// Verify the log entry for method exit
	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, methodName, entry["method"])
	assert.Equal(t, PhaseExit, entry["phase"])
	assert.Equal(t, MsgMethodCompleted, entry["message"])
	assert.NotEmpty(t, entry["goroutine"])
	assert.NotEmpty(t, entry["duration_ms"])

	// Verify return values are logged
	assert.Equal(t, "result1", entry["return1"])
	assert.Equal(t, float64(42), entry["return2"])
	assert.Equal(t, true, entry["return3"])
	assert.Nil(t, entry["return4"])
	assert.Contains(t, entry["return5"], "[struct")
}

// TestUT_FS_06_01_GoroutineID_GetCurrent_ReturnsValidID tests the getCurrentGoroutineID function.
//
//	Test Case ID    UT-FS-06-01
//	Title           Goroutine ID Retrieval
//	Description     Tests the getCurrentGoroutineID function
//	Preconditions   None
//	Steps           1. Call getCurrentGoroutineID
//	                2. Verify the result is not empty and is a number
//	Expected Result getCurrentGoroutineID returns a valid goroutine ID
//	Notes: This test verifies that the getCurrentGoroutineID function returns a valid goroutine ID.
func TestUT_FS_06_01_GoroutineID_GetCurrent_ReturnsValidID(t *testing.T) {
	// 1. Call getCurrentGoroutineID
	goroutineID := util.GetCurrentGoroutineID()

	// 2. Verify the result is not empty and is a number
	assert.NotEmpty(t, goroutineID, "Goroutine ID should not be empty")

	// Verify that the goroutine ID is a number
	// The goroutine ID should be a positive integer
	assert.Regexp(t, "^[0-9]+$", goroutineID, "Goroutine ID should be a number")
}

// TestUT_FS_07_01_Logging_InMethods_LogsCorrectly tests the logging in actual methods.
//
//	Test Case ID    UT-FS-07-01
//	Title           Method Logging in Methods
//	Description     Tests the logging in actual methods
//	Preconditions   None
//	Steps           1. Create a test method with logging
//	                2. Call the method
//	                3. Verify the log output contains the expected information
//	Expected Result Methods correctly log their entry and exit
//	Notes: This test verifies that methods correctly log their entry and exit.
func TestUT_FS_07_01_Logging_InMethods_LogsCorrectly(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// 1. Create a test method with logging
	testMethod := func(a string, b int) string {
		methodName, startTime := LogMethodEntry("TestMethod", a, b)
		result := a + "-result"
		LogMethodExit(methodName, time.Since(startTime), result)
		return result
	}

	// 2. Call the method
	result := testMethod("test", 42)

	// Wait a moment for the deferred function to be executed
	time.Sleep(10 * time.Millisecond)

	// 3. Verify the log output contains the expected information
	// Parse all log entries
	entries, err := parseLogEntries(buf)
	assert.NoError(t, err)
	assert.Len(t, entries, 2, "Expected 2 log entries")

	// Verify the first log entry (method entry)
	entry := entries[0]
	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, "TestMethod", entry["method"])
	assert.Equal(t, PhaseEntry, entry["phase"])
	assert.Equal(t, MsgMethodCalled, entry["message"])
	assert.NotEmpty(t, entry["goroutine"])
	assert.Equal(t, "test", entry["param1"])
	assert.Equal(t, float64(42), entry["param2"])

	// Verify the second log entry (method exit)
	entry = entries[1]
	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, "TestMethod", entry["method"])
	assert.Equal(t, PhaseExit, entry["phase"])
	assert.Equal(t, MsgMethodCompleted, entry["message"])
	assert.NotEmpty(t, entry["goroutine"])
	assert.NotEmpty(t, entry["duration_ms"])
	assert.Equal(t, "test-result", entry["return1"])

	// Verify the result
	assert.Equal(t, "test-result", result)
}

// TestUT_FS_08_01_Logging_MethodEntryAndExitWithContext_LogsCorrectly tests the LogMethodEntryWithContext and LogMethodExitWithContext functions.
//
//	Test Case ID    UT-FS-08-01
//	Title           Method Logging with Context
//	Description     Tests the LogMethodEntryWithContext and LogMethodExitWithContext functions
//	Preconditions   None
//	Steps           1. Create a test method with context-aware logging
//	                2. Call the method
//	                3. Verify the log output contains the expected information
//	Expected Result Methods correctly log their entry and exit with context
//	Notes: This test verifies that the logging functions correctly log method entry and exit with context.
func TestUT_FS_08_01_Logging_MethodEntryAndExitWithContext_LogsCorrectly(t *testing.T) {
	// Set up a test logger
	buf, cleanup := setupTestLogger()
	defer cleanup()

	// Create a log context
	ctx := NewLogContext("test_operation").
		WithRequestID("req123").
		WithUserID("user456").
		WithComponent("test_component").
		WithMethod("test_method").
		WithPath("/test/path")

	// 1. Create a test method with context-aware logging
	testMethodWithContext := func(ctx LogContext, a string, b int) string {
		// Add parameters to context for logging
		ctx = ctx.With("param1", a).With("param2", b)

		// Log method entry with context
		methodName, startTime, logger, ctx := LogMethodEntryWithContext("TestMethodWithContext", ctx)

		// Create the result
		result := a + "-result"

		// Log method exit with context
		LogMethodExitWithContext(methodName, startTime, logger, ctx, result)

		return result
	}

	// 2. Call the method
	result := testMethodWithContext(ctx, "test", 42)

	// Wait a moment for the deferred function to be executed
	time.Sleep(10 * time.Millisecond)

	// 3. Verify the log output contains the expected information
	// Parse all log entries
	entries, err := parseLogEntries(buf)
	assert.NoError(t, err)
	assert.Len(t, entries, 2, "Expected 2 log entries")

	// Verify the first log entry (method entry)
	entry := entries[0]
	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, "TestMethodWithContext", entry["method"])
	assert.Equal(t, PhaseEntry, entry["phase"])
	assert.Equal(t, MsgMethodCalled, entry["message"])
	assert.NotEmpty(t, entry["goroutine"])
	assert.Equal(t, "test", entry["param1"])
	assert.Equal(t, float64(42), entry["param2"])
	assert.Equal(t, "req123", entry["request_id"])
	assert.Equal(t, "user456", entry["user"])
	assert.Equal(t, "test_operation", entry["operation"])
	assert.Equal(t, "test_component", entry["component"])
	assert.Equal(t, "test_method", entry["method_name"])
	assert.Equal(t, "/test/path", entry["path"])

	// Verify the second log entry (method exit)
	entry = entries[1]
	assert.Equal(t, "debug", entry["level"])
	assert.Equal(t, "TestMethodWithContext", entry["method"])
	assert.Equal(t, PhaseExit, entry["phase"])
	assert.Equal(t, MsgMethodCompleted, entry["message"])
	assert.NotEmpty(t, entry["goroutine"])
	assert.NotEmpty(t, entry["duration_ms"])
	assert.Equal(t, "test-result", entry["return1"])
	assert.Equal(t, "req123", entry["request_id"])
	assert.Equal(t, "user456", entry["user"])
	assert.Equal(t, "test_operation", entry["operation"])
	assert.Equal(t, "test_component", entry["component"])
	assert.Equal(t, "test_method", entry["method_name"])
	assert.Equal(t, "/test/path", entry["path"])

	// Verify the result
	assert.Equal(t, "test-result", result)
}
