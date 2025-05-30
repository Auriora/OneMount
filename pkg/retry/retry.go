// Package retry provides utilities for retrying operations that may fail due to transient errors.
package retry

import (
	"context"
	"math/rand"
	"time"

	"github.com/auriora/onemount/pkg/errors"
	"github.com/auriora/onemount/pkg/logging"
)

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryableFuncWithResult is a function that returns a result and can be retried
type RetryableFuncWithResult[T any] func() (T, error)

// Config holds configuration for retry operations
type Config struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// Multiplier is the factor by which the delay increases after each retry
	Multiplier float64

	// Jitter is the maximum random jitter added to the delay
	Jitter float64

	// RetryableErrors is a list of error types that should be retried
	RetryableErrors []RetryableError
}

// RetryableError defines a function that determines if an error should be retried
type RetryableError func(error) bool

// DefaultConfig returns a default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.2,
		RetryableErrors: []RetryableError{
			IsRetryableNetworkError,
			IsRetryableServerError,
			IsRetryableRateLimitError,
		},
	}
}

// IsRetryableNetworkError returns true if the error is a network error that should be retried
func IsRetryableNetworkError(err error) bool {
	return errors.IsNetworkError(err)
}

// IsRetryableServerError returns true if the error is a server error that should be retried
func IsRetryableServerError(err error) bool {
	// Check if it's an operation error (typically 5xx errors)
	return errors.IsOperationError(err)
}

// IsRetryableRateLimitError returns true if the error is a rate limit error that should be retried
func IsRetryableRateLimitError(err error) bool {
	// Check if it's a resource busy error (typically 429 errors)
	return errors.IsResourceBusyError(err)
}

// Do retries the given function with exponential backoff
func Do(ctx context.Context, op RetryableFunc, config Config) error {
	var err error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the operation
		err = op()
		if err == nil {
			return nil
		}

		// Check if we should retry this error
		shouldRetry := false
		for _, retryableError := range config.RetryableErrors {
			if retryableError(err) {
				shouldRetry = true
				break
			}
		}

		// If we shouldn't retry or we've reached the maximum number of retries, return the error
		if !shouldRetry || attempt == config.MaxRetries {
			return err
		}

		// Calculate the next delay with jitter
		jitterRange := float64(delay) * config.Jitter
		jitterAmount := time.Duration(rand.Float64() * jitterRange)
		actualDelay := delay + jitterAmount

		// Log the retry
		logging.Info().
			Err(err).
			Int("attempt", attempt+1).
			Int("maxRetries", config.MaxRetries).
			Dur("delay", actualDelay).
			Msg("Operation failed, retrying after delay")

		// Wait for the delay or until the context is canceled
		select {
		case <-time.After(actualDelay):
			// Continue to the next retry
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "retry canceled by context")
		}

		// Increase the delay for the next retry, but don't exceed the maximum delay
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	// This should never be reached, but just in case
	return err
}

// DoWithResult retries the given function with exponential backoff and returns a result
func DoWithResult[T any](ctx context.Context, op RetryableFuncWithResult[T], config Config) (T, error) {
	var result T
	var err error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the operation
		result, err = op()
		if err == nil {
			return result, nil
		}

		// Check if we should retry this error
		shouldRetry := false
		for _, retryableError := range config.RetryableErrors {
			if retryableError(err) {
				shouldRetry = true
				break
			}
		}

		// If we shouldn't retry or we've reached the maximum number of retries, return the error
		if !shouldRetry || attempt == config.MaxRetries {
			return result, err
		}

		// Calculate the next delay with jitter
		jitterRange := float64(delay) * config.Jitter
		jitterAmount := time.Duration(rand.Float64() * jitterRange)
		actualDelay := delay + jitterAmount

		// Log the retry
		logging.Info().
			Err(err).
			Int("attempt", attempt+1).
			Int("maxRetries", config.MaxRetries).
			Dur("delay", actualDelay).
			Msg("Operation failed, retrying after delay")

		// Wait for the delay or until the context is canceled
		select {
		case <-time.After(actualDelay):
			// Continue to the next retry
		case <-ctx.Done():
			var zero T
			return zero, errors.Wrap(ctx.Err(), "retry canceled by context")
		}

		// Increase the delay for the next retry, but don't exceed the maximum delay
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	// This should never be reached, but just in case
	return result, err
}
