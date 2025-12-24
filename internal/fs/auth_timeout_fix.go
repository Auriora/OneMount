package fs

import (
	"context"
	"fmt"
	"time"

	"github.com/auriora/onemount/internal/graph"
)

// AuthTimeoutConfig defines timeout settings for authentication operations
type AuthTimeoutConfig struct {
	APICallTimeout      time.Duration // Timeout for individual API calls
	TokenRefreshTimeout time.Duration // Timeout for token refresh operations
	MaxRetries          int           // Maximum number of retries for failed operations
}

// DefaultAuthTimeoutConfig returns sensible default timeout values
func DefaultAuthTimeoutConfig() *AuthTimeoutConfig {
	return &AuthTimeoutConfig{
		APICallTimeout:      10 * time.Second,
		TokenRefreshTimeout: 30 * time.Second,
		MaxRetries:          3,
	}
}

// SafeAuthWrapper wraps authentication operations with proper timeout handling
type SafeAuthWrapper struct {
	auth   *graph.Auth
	config *AuthTimeoutConfig
}

// NewSafeAuthWrapper creates a new authentication wrapper with timeout protection
func NewSafeAuthWrapper(auth *graph.Auth, config *AuthTimeoutConfig) *SafeAuthWrapper {
	if config == nil {
		config = DefaultAuthTimeoutConfig()
	}
	return &SafeAuthWrapper{
		auth:   auth,
		config: config,
	}
}

// IsTokenExpired checks if the authentication token is expired
func (w *SafeAuthWrapper) IsTokenExpired() bool {
	if w.auth == nil {
		return true
	}
	expiresAt := time.Unix(w.auth.ExpiresAt, 0)
	return time.Now().After(expiresAt)
}

// RefreshTokenIfNeeded refreshes the token if it's expired or about to expire
func (w *SafeAuthWrapper) RefreshTokenIfNeeded() error {
	if !w.IsTokenExpired() {
		// Check if token expires within the next 5 minutes
		expiresAt := time.Unix(w.auth.ExpiresAt, 0)
		if time.Until(expiresAt) > 5*time.Minute {
			return nil // Token is still valid for a while
		}
	}

	// Token is expired or about to expire, refresh it
	ctx, cancel := context.WithTimeout(context.Background(), w.config.TokenRefreshTimeout)
	defer cancel()

	// Attempt to refresh the token
	err := w.auth.Refresh(ctx)
	if err != nil {
		return fmt.Errorf("failed to refresh expired token: %w", err)
	}

	return nil
}

// SafeAPICall makes an API call with timeout protection and automatic token refresh
func (w *SafeAuthWrapper) SafeAPICall(endpoint string) ([]byte, error) {
	// First, ensure token is valid
	if err := w.RefreshTokenIfNeeded(); err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Make the API call with timeout
	ctx, cancel := context.WithTimeout(context.Background(), w.config.APICallTimeout)
	defer cancel()

	// Retry logic for transient failures
	var lastErr error
	for attempt := 0; attempt < w.config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("API call timed out after %v", w.config.APICallTimeout)
		default:
		}

		data, err := graph.Get(endpoint, w.auth)
		if err == nil {
			return data, nil
		}

		lastErr = err
		if attempt < w.config.MaxRetries-1 {
			// Wait before retry, but respect context timeout
			select {
			case <-time.After(time.Duration(attempt+1) * time.Second):
			case <-ctx.Done():
				return nil, fmt.Errorf("API call timed out during retry: %w", ctx.Err())
			}
		}
	}

	return nil, fmt.Errorf("API call failed after %d attempts: %w", w.config.MaxRetries, lastErr)
}

// GetAuth returns the underlying auth object (use with caution)
func (w *SafeAuthWrapper) GetAuth() *graph.Auth {
	return w.auth
}

// ValidateConnection tests the connection with a simple API call
func (w *SafeAuthWrapper) ValidateConnection() error {
	_, err := w.SafeAPICall("/me")
	return err
}
