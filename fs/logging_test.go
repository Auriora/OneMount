package fs

import (
	"bytes"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

	// Test LogMethodCall and LogMethodReturn
	methodName, startTime := LogMethodCall()
	assert.NotEmpty(t, methodName, "Method name should not be empty")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	// Test LogMethodReturn with different types of return values
	LogMethodReturn(methodName, startTime, true)
	assert.Contains(t, buf.String(), "Method called")
	assert.Contains(t, buf.String(), "Method completed")
	assert.Contains(t, buf.String(), "return1")

	// Reset buffer
	buf.Reset()

	// Test with multiple return values
	methodName, startTime = LogMethodCall()
	LogMethodReturn(methodName, startTime, "test", 123, nil)
	assert.Contains(t, buf.String(), "Method called")
	assert.Contains(t, buf.String(), "Method completed")
	assert.Contains(t, buf.String(), "return1")
	assert.Contains(t, buf.String(), "return2")
	assert.Contains(t, buf.String(), "return3")

	// Reset buffer
	buf.Reset()

	// Test with struct return value
	type testStruct struct {
		Name string
		Age  int
	}
	methodName, startTime = LogMethodCall()
	LogMethodReturn(methodName, startTime, &testStruct{Name: "Test", Age: 30})
	assert.Contains(t, buf.String(), "Method called")
	assert.Contains(t, buf.String(), "Method completed")
	assert.Contains(t, buf.String(), "return1")
}

// TestLoggingInMethods tests the logging in actual methods
func TestLoggingInMethods(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()

	// Create a test filesystem
	fs, err := NewFilesystem(nil, t.TempDir(), 30)
	assert.NoError(t, err)

	// Test IsOffline method with logging
	isOffline := fs.IsOffline()
	assert.False(t, isOffline)
	assert.Contains(t, buf.String(), "Method called")
	assert.Contains(t, buf.String(), "method=IsOffline")
	assert.Contains(t, buf.String(), "phase=entry")
	assert.Contains(t, buf.String(), "Method completed")
	assert.Contains(t, buf.String(), "phase=exit")
	assert.Contains(t, buf.String(), "return1=false")

	// Reset buffer
	buf.Reset()

	// Test GetNodeID method with logging
	inode := fs.GetNodeID(999) // This should return nil for a non-existent ID
	assert.Nil(t, inode)
	assert.Contains(t, buf.String(), "Method called")
	assert.Contains(t, buf.String(), "method=GetNodeID")
	assert.Contains(t, buf.String(), "phase=entry")
	assert.Contains(t, buf.String(), "Method completed")
	assert.Contains(t, buf.String(), "phase=exit")
	assert.Contains(t, buf.String(), "return1=null")
}
