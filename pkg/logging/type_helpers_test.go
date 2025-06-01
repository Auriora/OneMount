package logging

import (
	"reflect"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTypeLogger(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		expectedType string
	}{
		{"bool", true, "logging.boolLogger"},
		{"int", 42, "logging.intLogger"},
		{"int8", int8(42), "logging.int8Logger"},
		{"int16", int16(42), "logging.int16Logger"},
		{"int32", int32(42), "logging.int32Logger"},
		{"int64", int64(42), "logging.int64Logger"},
		{"uint", uint(42), "logging.uintLogger"},
		{"uint8", uint8(42), "logging.uint8Logger"},
		{"uint16", uint16(42), "logging.uint16Logger"},
		{"uint32", uint32(42), "logging.uint32Logger"},
		{"uint64", uint64(42), "logging.uint64Logger"},
		{"float32", float32(42.5), "logging.float32Logger"},
		{"float64", float64(42.5), "logging.float64Logger"},
		{"string", "test", "logging.stringLogger"},
		{"time.Time", time.Now(), "logging.timeLogger"},
		{"[]byte", []byte("test"), "logging.byteSliceLogger"},
		{"[]string", []string{"test"}, "logging.stringSliceLogger"},
		{"fuse.Status", fuse.OK, "logging.fuseStatusLogger"},
		{"interface{}", struct{}{}, "logging.interfaceLogger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := getTypeLogger(reflect.TypeOf(tt.value))
			loggerType := reflect.TypeOf(logger).String()
			assert.Equal(t, tt.expectedType, loggerType)
		})
	}
}

func TestFuseStatusLogger(t *testing.T) {
	tests := []struct {
		name   string
		status fuse.Status
	}{
		{"OK", fuse.OK},
		{"ENOENT", fuse.ENOENT},
		{"EIO", fuse.EIO},
		{"EBADF", fuse.EBADF},
		{"EREMOTEIO", fuse.EREMOTEIO},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := fuseStatusLogger{}

			// This should not panic - the main test
			require.NotPanics(t, func() {
				// Use a real Event from the logging system
				event := Debug()
				logger.LogValue(event, "status", tt.status)
			})
		})
	}
}

func TestLogValueWithTypeLogger_FuseStatus(t *testing.T) {
	tests := []struct {
		name   string
		status fuse.Status
	}{
		{"OK", fuse.OK},
		{"ENOENT", fuse.ENOENT},
		{"EIO", fuse.EIO},
		{"EBADF", fuse.EBADF},
		{"EREMOTEIO", fuse.EREMOTEIO},
		{"EINVAL", fuse.EINVAL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			require.NotPanics(t, func() {
				event := Debug()
				logValueWithTypeLogger(event, "return1", tt.status)
			})
		})
	}
}

func TestLogReturn_FuseStatus(t *testing.T) {
	// This should not panic - this was the original issue
	require.NotPanics(t, func() {
		event := Debug()
		LogReturn(event, 0, fuse.OK)
	})
}

func TestLogParam_FuseStatus(t *testing.T) {
	// This should not panic
	require.NotPanics(t, func() {
		event := Debug()
		LogParam(event, 0, fuse.ENOENT)
	})
}

// TestTypeLoggerPanicRegression tests that we don't get panics for type conversion issues
func TestTypeLoggerPanicRegression(t *testing.T) {
	// This test specifically addresses the original panic:
	// "panic: interface conversion: interface {} is fuse.Status, not int32"

	// Test various fuse.Status values that could cause the original panic
	statusValues := []fuse.Status{
		fuse.OK,
		fuse.ENOENT,
		fuse.EIO,
		fuse.EBADF,
		fuse.EREMOTEIO,
		fuse.EINVAL,
	}

	for i, status := range statusValues {
		t.Run(status.String(), func(t *testing.T) {
			// None of these should panic
			require.NotPanics(t, func() {
				event := Debug()
				LogReturn(event, i, status)
			})

			require.NotPanics(t, func() {
				event := Debug()
				LogParam(event, i, status)
			})

			require.NotPanics(t, func() {
				event := Debug()
				logValueWithTypeLogger(event, "test", status)
			})
		})
	}
}
