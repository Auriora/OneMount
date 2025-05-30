package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUT_RT_01_01_Do_WithSuccessfulOperation_ReturnsNoError tests that Do returns no error when the operation succeeds
func TestUT_RT_01_01_Do_WithSuccessfulOperation_ReturnsNoError(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a config with no retries
	config := Config{
		MaxRetries:      0,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0.1,
		RetryableErrors: []RetryableError{},
	}

	// Create a successful operation
	op := func() error {
		return nil
	}

	// Execute the operation with retry
	err := Do(ctx, op, config)

	// Verify that no error is returned
	assert.NoError(t, err)
}

// TestUT_RT_01_02_Do_WithNonRetryableError_ReturnsError tests that Do returns an error when the operation fails with a non-retryable error
func TestUT_RT_01_02_Do_WithNonRetryableError_ReturnsError(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a config with retries
	config := Config{
		MaxRetries:      3,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0.1,
		RetryableErrors: []RetryableError{},
	}

	// Create an operation that always fails with a non-retryable error
	expectedErr := errors.New("non-retryable error")
	op := func() error {
		return expectedErr
	}

	// Execute the operation with retry
	err := Do(ctx, op, config)

	// Verify that the error is returned
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

// TestUT_RT_01_03_Do_WithRetryableError_EventuallySucceeds tests that Do retries and eventually succeeds
func TestUT_RT_01_03_Do_WithRetryableError_EventuallySucceeds(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a config with retries
	config := Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryableErrors: []RetryableError{
			func(err error) bool { return err.Error() == "retryable error" },
		},
	}

	// Create an operation that fails a few times and then succeeds
	attempts := 0
	op := func() error {
		attempts++
		if attempts <= 2 {
			return errors.New("retryable error")
		}
		return nil
	}

	// Execute the operation with retry
	err := Do(ctx, op, config)

	// Verify that no error is returned and the operation was retried
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

// TestUT_RT_01_04_Do_WithRetryableError_ExceedsMaxRetries tests that Do returns an error when max retries is exceeded
func TestUT_RT_01_04_Do_WithRetryableError_ExceedsMaxRetries(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a config with retries
	config := Config{
		MaxRetries:   2,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryableErrors: []RetryableError{
			func(err error) bool { return err.Error() == "retryable error" },
		},
	}

	// Create an operation that always fails with a retryable error
	expectedErr := errors.New("retryable error")
	attempts := 0
	op := func() error {
		attempts++
		return expectedErr
	}

	// Execute the operation with retry
	err := Do(ctx, op, config)

	// Verify that the error is returned and the operation was retried the maximum number of times
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 3, attempts) // Initial attempt + 2 retries
}

// TestUT_RT_01_05_Do_WithCanceledContext_ReturnsError tests that Do returns an error when the context is canceled
func TestUT_RT_01_05_Do_WithCanceledContext_ReturnsError(t *testing.T) {
	// Create a context that's already canceled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create a config with retries
	config := Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second, // Long delay to ensure context cancellation takes effect
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryableErrors: []RetryableError{
			func(err error) bool { return err.Error() == "retryable error" },
		},
	}

	// Create an operation that fails with a retryable error
	op := func() error {
		return errors.New("retryable error")
	}

	// Execute the operation with retry
	err := Do(ctx, op, config)

	// Verify that a context canceled error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "retry canceled by context")
}

// TestUT_RT_02_01_DoWithResult_WithSuccessfulOperation_ReturnsResult tests that DoWithResult returns a result when the operation succeeds
func TestUT_RT_02_01_DoWithResult_WithSuccessfulOperation_ReturnsResult(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a config with no retries
	config := Config{
		MaxRetries:      0,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0.1,
		RetryableErrors: []RetryableError{},
	}

	// Create a successful operation
	expectedResult := "success"
	op := func() (string, error) {
		return expectedResult, nil
	}

	// Execute the operation with retry
	result, err := DoWithResult(ctx, op, config)

	// Verify that no error is returned and the result is correct
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

// TestUT_RT_02_02_DoWithResult_WithNonRetryableError_ReturnsError tests that DoWithResult returns an error when the operation fails with a non-retryable error
func TestUT_RT_02_02_DoWithResult_WithNonRetryableError_ReturnsError(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a config with retries
	config := Config{
		MaxRetries:      3,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0.1,
		RetryableErrors: []RetryableError{},
	}

	// Create an operation that always fails with a non-retryable error
	expectedErr := errors.New("non-retryable error")
	op := func() (string, error) {
		return "", expectedErr
	}

	// Execute the operation with retry
	result, err := DoWithResult(ctx, op, config)

	// Verify that the error is returned and the result is empty
	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, "", result)
}

// TestUT_RT_02_03_DoWithResult_WithRetryableError_EventuallySucceeds tests that DoWithResult retries and eventually succeeds
func TestUT_RT_02_03_DoWithResult_WithRetryableError_EventuallySucceeds(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// Create a config with retries
	config := Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryableErrors: []RetryableError{
			func(err error) bool { return err.Error() == "retryable error" },
		},
	}

	// Create an operation that fails a few times and then succeeds
	attempts := 0
	expectedResult := "success"
	op := func() (string, error) {
		attempts++
		if attempts <= 2 {
			return "", errors.New("retryable error")
		}
		return expectedResult, nil
	}

	// Execute the operation with retry
	result, err := DoWithResult(ctx, op, config)

	// Verify that no error is returned, the result is correct, and the operation was retried
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
	assert.Equal(t, 3, attempts)
}

// TestUT_RT_03_01_DefaultConfig_ReturnsExpectedValues tests that DefaultConfig returns the expected values
func TestUT_RT_03_01_DefaultConfig_ReturnsExpectedValues(t *testing.T) {
	// Get the default config
	config := DefaultConfig()

	// Verify the values
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, 1*time.Second, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.Equal(t, 2.0, config.Multiplier)
	assert.Equal(t, 0.2, config.Jitter)
	assert.Len(t, config.RetryableErrors, 3)
}
