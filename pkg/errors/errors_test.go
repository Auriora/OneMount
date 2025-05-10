package errors

import (
	"fmt"
	"testing"

	"github.com/auriora/onemount/pkg/logging"
	"github.com/stretchr/testify/assert"
)

// TestUT_ER_01_01_Wrap_WithMessage_AddsContext tests the Wrap function.
func TestUT_ER_01_01_Wrap_WithMessage_AddsContext(t *testing.T) {
	// Create an original error
	originalErr := New("original error")

	// Wrap the error with context
	wrappedErr := Wrap(originalErr, "context message")

	// Verify that the wrapped error contains both the context and the original error
	assert.Contains(t, wrappedErr.Error(), "context message")
	assert.Contains(t, wrappedErr.Error(), "original error")

	// Verify that errors.Is works with the wrapped error
	assert.True(t, Is(wrappedErr, originalErr))

	// Verify that errors.Unwrap returns the original error
	assert.Equal(t, originalErr, Unwrap(wrappedErr))
}

// TestUT_ER_01_02_Wrap_WithNilError_ReturnsNil tests the Wrap function with a nil error.
func TestUT_ER_01_02_Wrap_WithNilError_ReturnsNil(t *testing.T) {
	// Wrap a nil error
	wrappedErr := Wrap(nil, "context message")

	// Verify that the result is nil
	assert.Nil(t, wrappedErr)
}

// TestUT_ER_02_01_Wrapf_WithFormattedMessage_AddsContext tests the Wrapf function.
func TestUT_ER_02_01_Wrapf_WithFormattedMessage_AddsContext(t *testing.T) {
	// Create an original error
	originalErr := New("original error")

	// Wrap the error with a formatted context message
	wrappedErr := Wrapf(originalErr, "context message with %s", "parameter")

	// Verify that the wrapped error contains the formatted context and the original error
	assert.Contains(t, wrappedErr.Error(), "context message with parameter")
	assert.Contains(t, wrappedErr.Error(), "original error")

	// Verify that errors.Is works with the wrapped error
	assert.True(t, Is(wrappedErr, originalErr))

	// Verify that errors.Unwrap returns the original error
	assert.Equal(t, originalErr, Unwrap(wrappedErr))
}

// TestUT_ER_02_02_Wrapf_WithNilError_ReturnsNil tests the Wrapf function with a nil error.
func TestUT_ER_02_02_Wrapf_WithNilError_ReturnsNil(t *testing.T) {
	// Wrap a nil error with a formatted message
	wrappedErr := Wrapf(nil, "context message with %s", "parameter")

	// Verify that the result is nil
	assert.Nil(t, wrappedErr)
}

// TestUT_ER_03_01_WrapAndLog_WithMessage_WrapsAndLogsError tests the WrapAndLog function.
func TestUT_ER_03_01_WrapAndLog_WithMessage_WrapsAndLogsError(t *testing.T) {
	// Create an original error
	originalErr := New("original error")

	// Wrap and log the error
	wrappedErr := logging.WrapAndLog(originalErr, "context message", "field1", "value1")

	// Verify that the wrapped error contains both the context and the original error
	assert.Contains(t, wrappedErr.Error(), "context message")
	assert.Contains(t, wrappedErr.Error(), "original error")

	// Verify that errors.Is works with the wrapped error
	assert.True(t, Is(wrappedErr, originalErr))
}

// TestUT_ER_03_02_WrapAndLog_WithNilError_ReturnsNil tests the WrapAndLog function with a nil error.
func TestUT_ER_03_02_WrapAndLog_WithNilError_ReturnsNil(t *testing.T) {
	// Wrap and log a nil error
	wrappedErr := WrapAndLog(nil, "context message", "field1", "value1")

	// Verify that the result is nil
	assert.Nil(t, wrappedErr)
}

// TestUT_ER_04_01_WrapfAndLog_WithFormattedMessage_WrapsAndLogsError tests the WrapfAndLog function.
func TestUT_ER_04_01_WrapfAndLog_WithFormattedMessage_WrapsAndLogsError(t *testing.T) {
	// Create an original error
	originalErr := New("original error")

	// Wrap and log the error with a formatted message
	wrappedErr := WrapfAndLog(originalErr, "context message with %s", "parameter")

	// Verify that the wrapped error contains the formatted context and the original error
	assert.Contains(t, wrappedErr.Error(), "context message with parameter")
	assert.Contains(t, wrappedErr.Error(), "original error")

	// Verify that errors.Is works with the wrapped error
	assert.True(t, Is(wrappedErr, originalErr))
}

// TestUT_ER_04_02_WrapfAndLog_WithNilError_ReturnsNil tests the WrapfAndLog function with a nil error.
func TestUT_ER_04_02_WrapfAndLog_WithNilError_ReturnsNil(t *testing.T) {
	// Wrap and log a nil error with a formatted message
	wrappedErr := WrapfAndLog(nil, "context message with %s", "parameter")

	// Verify that the result is nil
	assert.Nil(t, wrappedErr)
}

// TestUT_ER_05_01_LogAndReturn_WithMessage_LogsAndReturnsError tests the LogAndReturn function.
func TestUT_ER_05_01_LogAndReturn_WithMessage_LogsAndReturnsError(t *testing.T) {
	// Create an original error
	originalErr := New("original error")

	// Log and return the error
	returnedErr := LogAndReturn(originalErr, "error message", "field1", "value1")

	// Verify that the returned error is the original error
	assert.Equal(t, originalErr, returnedErr)
}

// TestUT_ER_05_02_LogAndReturn_WithNilError_ReturnsNil tests the LogAndReturn function with a nil error.
func TestUT_ER_05_02_LogAndReturn_WithNilError_ReturnsNil(t *testing.T) {
	// Log and return a nil error
	returnedErr := LogAndReturn(nil, "error message", "field1", "value1")

	// Verify that the result is nil
	assert.Nil(t, returnedErr)
}

// TestUT_ER_06_01_ErrorChain_WithMultipleWraps_PreservesChain tests that error chains are preserved.
func TestUT_ER_06_01_ErrorChain_WithMultipleWraps_PreservesChain(t *testing.T) {
	// Create a chain of errors
	originalErr := New("original error")
	wrappedOnce := Wrap(originalErr, "first wrap")
	wrappedTwice := Wrap(wrappedOnce, "second wrap")
	wrappedThrice := Wrap(wrappedTwice, "third wrap")

	// Verify that the final error contains all the context messages
	assert.Contains(t, wrappedThrice.Error(), "third wrap")
	assert.Contains(t, wrappedThrice.Error(), "second wrap")
	assert.Contains(t, wrappedThrice.Error(), "first wrap")
	assert.Contains(t, wrappedThrice.Error(), "original error")

	// Verify that errors.Is works with the wrapped error
	assert.True(t, Is(wrappedThrice, originalErr))

	// Verify that errors.Unwrap returns the correct error at each level
	assert.Equal(t, wrappedTwice, Unwrap(wrappedThrice))
	assert.Equal(t, wrappedOnce, Unwrap(wrappedTwice))
	assert.Equal(t, originalErr, Unwrap(wrappedOnce))
	assert.Nil(t, Unwrap(originalErr))
}

// TestUT_ER_07_01_As_WithCustomErrorType_FindsMatchingType tests the As function.
func TestUT_ER_07_01_As_WithCustomErrorType_FindsMatchingType(t *testing.T) {
	// Use a simple error type that implements the error interface
	originalErr := fmt.Errorf("original error")

	// Wrap the error
	wrappedErr := Wrap(originalErr, "wrapped")

	// Use As to find the original error in the chain
	var target error
	assert.True(t, As(wrappedErr, &target))
	// When using errors.As, target is set to the first error in the chain that matches the target type
	// In this case, target is set to wrappedErr, not originalErr
	assert.Contains(t, target.Error(), originalErr.Error())
}

// TestUT_ER_08_01_MultipleErrorTypes_InChain_CanBeIdentified tests identifying multiple error types in a chain.
func TestUT_ER_08_01_MultipleErrorTypes_InChain_CanBeIdentified(t *testing.T) {
	// Create a chain of errors using proper error wrapping
	baseErr := New("base error")
	err1 := Wrap(baseErr, "error type 1")
	err2 := Wrap(err1, "error type 2")
	err3 := Wrap(err2, "error type 3")

	// Verify that Is works with the error chain
	assert.True(t, Is(err3, baseErr))
	assert.True(t, Is(err3, err1))
	assert.True(t, Is(err3, err2))

	// Verify that the error messages are preserved in the chain
	assert.Contains(t, err3.Error(), "base error")
	assert.Contains(t, err3.Error(), "error type 1")
	assert.Contains(t, err3.Error(), "error type 2")
	assert.Contains(t, err3.Error(), "error type 3")
}
