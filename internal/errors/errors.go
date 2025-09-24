// Package errors provides custom error types and error handling utilities for the OneMount project.
// It includes functions for error wrapping and error context propagation.
package errors

import (
	"errors"
	"fmt"
)

// Unwrap unwraps an error to find the underlying cause.
// This is a convenience function that uses errors.Unwrap.
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// Is reports whether any error in err's chain matches target.
// This is a convenience function that uses errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
// This is a convenience function that uses errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Wrap wraps an error with a message.
// This is a convenience function for the common pattern of wrapping an error with context.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted message.
// This is a convenience function for the common pattern of wrapping an error with context.
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// New creates a new error with the given message.
// This is a convenience function that uses errors.New.
func New(message string) error {
	return errors.New(message)
}

// Note: Logging functions have been moved to the logging package.
// Use the equivalent functions from the logging package instead.
