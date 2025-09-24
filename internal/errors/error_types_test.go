package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUT_ET_01_01_ErrorType_String_ReturnsCorrectString tests the String method of ErrorType.
func TestUT_ET_01_01_ErrorType_String_ReturnsCorrectString(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		expected  string
	}{
		{ErrorTypeUnknown, "UnknownError"},
		{ErrorTypeNetwork, "NetworkError"},
		{ErrorTypeNotFound, "NotFoundError"},
		{ErrorTypeAuth, "AuthError"},
		{ErrorTypeValidation, "ValidationError"},
		{ErrorTypeOperation, "OperationError"},
		{ErrorTypeTimeout, "TimeoutError"},
		{ErrorTypeResourceBusy, "ResourceBusyError"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			assert.Equal(t, test.expected, test.errorType.String())
		})
	}
}

// TestUT_ET_02_01_TypedError_Error_WithUnderlyingError_IncludesAllParts tests the Error method of TypedError with an underlying error.
func TestUT_ET_02_01_TypedError_Error_WithUnderlyingError_IncludesAllParts(t *testing.T) {
	// Create an underlying error
	underlyingErr := fmt.Errorf("underlying error")

	// Create a typed error with the underlying error
	typedErr := &TypedError{
		Type:       ErrorTypeNetwork,
		Message:    "network error message",
		StatusCode: 503,
		Err:        underlyingErr,
	}

	// Verify that the error message includes all parts
	errorMessage := typedErr.Error()
	assert.Contains(t, errorMessage, "NetworkError")
	assert.Contains(t, errorMessage, "network error message")
	assert.Contains(t, errorMessage, "underlying error")
}

// TestUT_ET_02_02_TypedError_Error_WithoutUnderlyingError_IncludesTypeAndMessage tests the Error method of TypedError without an underlying error.
func TestUT_ET_02_02_TypedError_Error_WithoutUnderlyingError_IncludesTypeAndMessage(t *testing.T) {
	// Create a typed error without an underlying error
	typedErr := &TypedError{
		Type:       ErrorTypeNotFound,
		Message:    "resource not found",
		StatusCode: 404,
		Err:        nil,
	}

	// Verify that the error message includes the type and message
	errorMessage := typedErr.Error()
	assert.Contains(t, errorMessage, "NotFoundError")
	assert.Contains(t, errorMessage, "resource not found")
	assert.NotContains(t, errorMessage, "nil")
}

// TestUT_ET_03_01_TypedError_Unwrap_ReturnsUnderlyingError tests the Unwrap method of TypedError.
func TestUT_ET_03_01_TypedError_Unwrap_ReturnsUnderlyingError(t *testing.T) {
	// Create an underlying error
	underlyingErr := fmt.Errorf("underlying error")

	// Create a typed error with the underlying error
	typedErr := &TypedError{
		Type:       ErrorTypeNetwork,
		Message:    "network error message",
		StatusCode: 503,
		Err:        underlyingErr,
	}

	// Verify that Unwrap returns the underlying error
	assert.Equal(t, underlyingErr, typedErr.Unwrap())
}

// TestUT_ET_04_01_NewErrorFunctions_CreateCorrectErrorTypes tests the constructor functions for typed errors.
func TestUT_ET_04_01_NewErrorFunctions_CreateCorrectErrorTypes(t *testing.T) {
	// Create an underlying error
	underlyingErr := fmt.Errorf("underlying error")

	// Test each constructor function
	tests := []struct {
		name           string
		errorFunc      func(string, error) error
		expectedType   ErrorType
		expectedStatus int
	}{
		{"NewNetworkError", NewNetworkError, ErrorTypeNetwork, 503},
		{"NewNotFoundError", NewNotFoundError, ErrorTypeNotFound, 404},
		{"NewAuthError", NewAuthError, ErrorTypeAuth, 401},
		{"NewValidationError", NewValidationError, ErrorTypeValidation, 400},
		{"NewOperationError", NewOperationError, ErrorTypeOperation, 500},
		{"NewTimeoutError", NewTimeoutError, ErrorTypeTimeout, 408},
		{"NewResourceBusyError", NewResourceBusyError, ErrorTypeResourceBusy, 409},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create the error using the constructor function
			err := test.errorFunc("test message", underlyingErr)

			// Verify that the error is a TypedError
			var typedErr *TypedError
			assert.True(t, As(err, &typedErr))

			// Verify that the error has the correct type and status code
			assert.Equal(t, test.expectedType, typedErr.Type)
			assert.Equal(t, test.expectedStatus, typedErr.StatusCode)
			assert.Equal(t, "test message", typedErr.Message)
			assert.Equal(t, underlyingErr, typedErr.Err)
		})
	}
}

// TestUT_ET_05_01_IsErrorTypeFunctions_ReturnCorrectResults tests the Is*Error functions.
func TestUT_ET_05_01_IsErrorTypeFunctions_ReturnCorrectResults(t *testing.T) {
	// Create errors of each type
	networkErr := NewNetworkError("network error", nil)
	notFoundErr := NewNotFoundError("not found error", nil)
	authErr := NewAuthError("auth error", nil)
	validationErr := NewValidationError("validation error", nil)
	operationErr := NewOperationError("operation error", nil)
	timeoutErr := NewTimeoutError("timeout error", nil)
	resourceBusyErr := NewResourceBusyError("resource busy error", nil)

	// Create a regular error
	regularErr := fmt.Errorf("regular error")

	// Test each Is*Error function with each error type
	tests := []struct {
		name     string
		isFunc   func(error) bool
		err      error
		expected bool
	}{
		{"IsNetworkError with NetworkError", IsNetworkError, networkErr, true},
		{"IsNetworkError with NotFoundError", IsNetworkError, notFoundErr, false},
		{"IsNetworkError with regular error", IsNetworkError, regularErr, false},

		{"IsNotFoundError with NotFoundError", IsNotFoundError, notFoundErr, true},
		{"IsNotFoundError with NetworkError", IsNotFoundError, networkErr, false},
		{"IsNotFoundError with regular error", IsNotFoundError, regularErr, false},

		{"IsAuthError with AuthError", IsAuthError, authErr, true},
		{"IsAuthError with NetworkError", IsAuthError, networkErr, false},
		{"IsAuthError with regular error", IsAuthError, regularErr, false},

		{"IsValidationError with ValidationError", IsValidationError, validationErr, true},
		{"IsValidationError with NetworkError", IsValidationError, networkErr, false},
		{"IsValidationError with regular error", IsValidationError, regularErr, false},

		{"IsOperationError with OperationError", IsOperationError, operationErr, true},
		{"IsOperationError with NetworkError", IsOperationError, networkErr, false},
		{"IsOperationError with regular error", IsOperationError, regularErr, false},

		{"IsTimeoutError with TimeoutError", IsTimeoutError, timeoutErr, true},
		{"IsTimeoutError with NetworkError", IsTimeoutError, networkErr, false},
		{"IsTimeoutError with regular error", IsTimeoutError, regularErr, false},

		{"IsResourceBusyError with ResourceBusyError", IsResourceBusyError, resourceBusyErr, true},
		{"IsResourceBusyError with NetworkError", IsResourceBusyError, networkErr, false},
		{"IsResourceBusyError with regular error", IsResourceBusyError, regularErr, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.isFunc(test.err))
		})
	}
}

// TestUT_ET_06_01_ErrorWrapping_PreservesErrorType tests that error wrapping preserves the error type.
func TestUT_ET_06_01_ErrorWrapping_PreservesErrorType(t *testing.T) {
	// Create a typed error
	notFoundErr := NewNotFoundError("resource not found", nil)

	// Wrap the error
	wrappedErr := Wrap(notFoundErr, "wrapped error")

	// Verify that the wrapped error is still recognized as a NotFoundError
	assert.True(t, IsNotFoundError(wrappedErr))

	// Verify that Is works with the original error
	assert.True(t, Is(wrappedErr, notFoundErr))
}

// TestUT_ET_06_02_ErrorChain_PreservesAllTypes tests that an error chain preserves all error types.
func TestUT_ET_06_02_ErrorChain_PreservesAllTypes(t *testing.T) {
	// Create a chain of errors
	baseErr := fmt.Errorf("base error")
	notFoundErr := NewNotFoundError("resource not found", baseErr)
	wrappedErr := Wrap(notFoundErr, "wrapped error")

	// Verify that the wrapped error is still recognized as a NotFoundError
	assert.True(t, IsNotFoundError(wrappedErr))

	// Verify that Is works with all errors in the chain
	assert.True(t, Is(wrappedErr, notFoundErr))
	assert.True(t, Is(wrappedErr, baseErr))
}
