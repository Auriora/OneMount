package testutil

import (
	"context"
	"os"
	"testing"
	"time"
)

// WaitForCondition waits for a condition to be true with a configurable timeout and polling interval
func WaitForCondition(t *testing.T, condition func() bool, timeout, pollInterval time.Duration, message string) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(pollInterval)
	}

	t.Fatalf("Timed out waiting for condition: %s", message)
}

// WaitForConditionWithContext waits for a condition to be true with a context for cancellation
func WaitForConditionWithContext(ctx context.Context, t *testing.T, condition func() bool, pollInterval time.Duration, message string) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("Context cancelled while waiting for condition: %s - %v", message, ctx.Err())
			return
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// RetryWithBackoff retries an operation with exponential backoff until it succeeds or times out
func RetryWithBackoff(t *testing.T, operation func() error, maxRetries int, initialBackoff, maxBackoff time.Duration, message string) error {
	var err error
	backoff := initialBackoff

	for i := 0; i < maxRetries; i++ {
		err = operation()
		if err == nil {
			return nil
		}

		t.Logf("Retry %d/%d failed: %v - %s", i+1, maxRetries, err, message)

		if i < maxRetries-1 {
			time.Sleep(backoff)
			// Exponential backoff with jitter
			backoff = time.Duration(float64(backoff) * 1.5)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	return err
}

// RunWithTimeout runs an operation with a timeout
func RunWithTimeout(t *testing.T, operation func() error, timeout time.Duration, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		t.Fatalf("Operation timed out after %v: %s", timeout, message)
		return ctx.Err()
	}
}

// RunConcurrently runs multiple operations concurrently and waits for all to complete
func RunConcurrently(t *testing.T, operations []func() error, timeout time.Duration, message string) []error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	results := make(chan error, len(operations))
	errors := make([]error, len(operations))

	for i, op := range operations {
		go func(index int, operation func() error) {
			results <- operation()
		}(i, op)
	}

	for i := 0; i < len(operations); i++ {
		select {
		case err := <-results:
			errors[i] = err
		case <-ctx.Done():
			t.Fatalf("Concurrent operations timed out after %v: %s", timeout, message)
			return errors
		}
	}

	return errors
}

// WaitForFileChange waits for a file to change (by checking its modification time)
func WaitForFileChange(t *testing.T, path string, timeout, pollInterval time.Duration) {
	initialStat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat file %s: %v", path, err)
	}

	initialModTime := initialStat.ModTime()

	WaitForCondition(t, func() bool {
		currentStat, err := os.Stat(path)
		if err != nil {
			return false
		}
		return currentStat.ModTime().After(initialModTime)
	}, timeout, pollInterval, "File did not change within timeout")
}

// WaitForFileExistence waits for a file to exist
func WaitForFileExistence(t *testing.T, path string, shouldExist bool, timeout, pollInterval time.Duration) {
	message := "File did not exist within timeout"
	if !shouldExist {
		message = "File did not get deleted within timeout"
	}

	WaitForCondition(t, func() bool {
		return FileExists(path) == shouldExist
	}, timeout, pollInterval, message)
}
