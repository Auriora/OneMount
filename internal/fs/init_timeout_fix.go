package fs

import (
	"context"
	"fmt"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/logging"
)

// InitTimeoutConfig defines timeout settings for filesystem initialization operations
type InitTimeoutConfig struct {
	RootItemTimeout     time.Duration // Timeout for fetching root item from Graph API
	AuthValidateTimeout time.Duration // Timeout for validating authentication
	DatabaseTimeout     time.Duration // Timeout for database operations
}

// DefaultInitTimeoutConfig returns sensible default timeout values for initialization
func DefaultInitTimeoutConfig() *InitTimeoutConfig {
	return &InitTimeoutConfig{
		RootItemTimeout:     15 * time.Second, // Timeout for root item fetch
		AuthValidateTimeout: 10 * time.Second, // Timeout for auth validation
		DatabaseTimeout:     30 * time.Second, // Timeout for database operations
	}
}

// SafeGetRootItem fetches the root item with timeout protection to prevent hangs during initialization
func SafeGetRootItem(ctx context.Context, auth *graph.Auth, config *InitTimeoutConfig) (*graph.DriveItem, error) {
	if config == nil {
		config = DefaultInitTimeoutConfig()
	}

	// Create a timeout context for the root item fetch
	timeoutCtx, cancel := context.WithTimeout(ctx, config.RootItemTimeout)
	defer cancel()

	// Channel to receive the result
	resultChan := make(chan struct {
		item *graph.DriveItem
		err  error
	}, 1)

	// Start the API call in a goroutine
	go func() {
		logging.Debug().Msg("Attempting to fetch root item from Graph API")
		item, err := graph.GetItem("root", auth)
		resultChan <- struct {
			item *graph.DriveItem
			err  error
		}{item, err}
	}()

	// Wait for either the result or timeout
	select {
	case result := <-resultChan:
		if result.err != nil {
			logging.Debug().Err(result.err).Msg("Failed to fetch root item from Graph API")
			return nil, result.err
		}
		logging.Debug().Msg("Successfully fetched root item from Graph API")
		return result.item, nil

	case <-timeoutCtx.Done():
		logging.Warn().
			Dur("timeout", config.RootItemTimeout).
			Msg("Root item fetch timed out - this prevents initialization hangs")

		// Return a timeout error that can be handled as offline mode
		return nil, fmt.Errorf("root item fetch timed out after %v: %w", config.RootItemTimeout, timeoutCtx.Err())
	}
}

// ValidateAuthWithTimeout validates authentication with timeout protection
func ValidateAuthWithTimeout(ctx context.Context, auth *graph.Auth, config *InitTimeoutConfig) error {
	if config == nil {
		config = DefaultInitTimeoutConfig()
	}

	// Create a timeout context for auth validation
	timeoutCtx, cancel := context.WithTimeout(ctx, config.AuthValidateTimeout)
	defer cancel()

	// Channel to receive the result
	resultChan := make(chan error, 1)

	// Start the validation in a goroutine
	go func() {
		logging.Debug().Msg("Validating authentication tokens")

		// Check if token is expired first (this is fast)
		if auth.ExpiresAt > 0 {
			expiresAt := time.Unix(auth.ExpiresAt, 0)
			if time.Now().After(expiresAt) {
				logging.Debug().Time("expiresAt", expiresAt).Msg("Token is expired")
				resultChan <- fmt.Errorf("authentication token expired at %v", expiresAt)
				return
			}
		}

		// Try a simple API call to validate the token
		_, err := graph.Get("/me", auth)
		resultChan <- err
	}()

	// Wait for either the result or timeout
	select {
	case err := <-resultChan:
		if err != nil {
			logging.Debug().Err(err).Msg("Authentication validation failed")
			return err
		}
		logging.Debug().Msg("Authentication validation successful")
		return nil

	case <-timeoutCtx.Done():
		logging.Warn().
			Dur("timeout", config.AuthValidateTimeout).
			Msg("Authentication validation timed out")

		return fmt.Errorf("authentication validation timed out after %v: %w", config.AuthValidateTimeout, timeoutCtx.Err())
	}
}

// IsTimeoutError checks if an error is a timeout error that should trigger offline mode
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context timeout errors
	if err == context.DeadlineExceeded {
		return true
	}

	// Check for wrapped timeout errors
	return fmt.Sprintf("%v", err) == "context deadline exceeded" ||
		fmt.Sprintf("%v", err) == "timeout" ||
		fmt.Sprintf("%v", err) == "i/o timeout"
}
