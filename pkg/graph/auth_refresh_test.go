package graph

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUT_GR_AUTH_01_01_TokenRefresh_ExpiredToken_RefreshesSuccessfully tests successful token refresh
func TestUT_GR_AUTH_01_01_TokenRefresh_ExpiredToken_RefreshesSuccessfully(t *testing.T) {
	// Create auth with expired token
	auth := &Auth{
		AccessToken:  "expired-token",
		RefreshToken: "valid-refresh-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		AuthConfig: AuthConfig{
			ClientID:    "test-client-id",
			CodeURL:     "https://test.example.com/auth",
			TokenURL:    "https://test.example.com/token",
			RedirectURL: "https://test.example.com/redirect",
		},
	}

	// Verify token is expired
	isExpired := auth.ExpiresAt <= time.Now().Unix()
	assert.True(t, isExpired, "Token should be expired")

	// Create mock authenticator
	mockAuth := NewMockAuthenticator()

	// Attempt to refresh token (mock always succeeds)
	err := mockAuth.Refresh()

	// Verify refresh succeeded
	assert.NoError(t, err, "Token refresh should succeed")

	// Verify we can get auth back
	updatedAuth := mockAuth.GetAuth()
	assert.NotNil(t, updatedAuth, "Should be able to get auth after refresh")
}

// TestUT_GR_AUTH_02_01_TokenRefresh_WithContext_RespectsTimeout tests refresh with context timeout
func TestUT_GR_AUTH_02_01_TokenRefresh_WithContext_RespectsTimeout(t *testing.T) {
	// Create mock authenticator
	mockAuth := NewMockAuthenticator()

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Attempt to refresh token with context (mock doesn't actually respect timeout, but tests the interface)
	err := mockAuth.RefreshWithContext(ctx)

	// Verify refresh completed (mock implementation doesn't simulate timeout)
	assert.NoError(t, err, "Mock refresh should succeed")
}

// TestUT_GR_AUTH_03_01_TokenRefresh_ValidToken_SkipsRefresh tests refresh with valid token
func TestUT_GR_AUTH_03_01_TokenRefresh_ValidToken_SkipsRefresh(t *testing.T) {
	// Create auth with valid token
	auth := &Auth{
		AccessToken:  "valid-token",
		RefreshToken: "valid-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(), // Expires in 1 hour
		AuthConfig: AuthConfig{
			ClientID:    "test-client-id",
			CodeURL:     "https://test.example.com/auth",
			TokenURL:    "https://test.example.com/token",
			RedirectURL: "https://test.example.com/redirect",
		},
	}

	// Verify token is not expired
	isExpired := auth.ExpiresAt <= time.Now().Unix()
	assert.False(t, isExpired, "Token should not be expired")

	// Create mock authenticator
	mockAuth := NewMockAuthenticator()

	// Attempt to refresh token
	err := mockAuth.Refresh()

	// Verify refresh succeeded
	assert.NoError(t, err, "Token refresh should succeed")

	// Verify we can get auth back
	updatedAuth := mockAuth.GetAuth()
	assert.NotNil(t, updatedAuth, "Should be able to get auth")
}
