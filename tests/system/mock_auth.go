package system

import (
	"encoding/json"
	"os"
	"time"

	"github.com/auriora/onemount/internal/graph"
)

// MockAuthConfig represents configuration for mock authentication
type MockAuthConfig struct {
	EnableMockAuth bool   `json:"enable_mock_auth"`
	MockAccount    string `json:"mock_account"`
}

// CreateMockAuthTokens creates mock authentication tokens for headless testing
func CreateMockAuthTokens(authPath string) error {
	mockAuth := &graph.Auth{
		AccessToken:  "mock_access_token_for_testing",
		RefreshToken: "mock_refresh_token_for_testing",
		ExpiresAt:    time.Now().Add(24 * time.Hour).Unix(), // Valid for 24 hours (Unix timestamp)
		Account:      "mock_user_id_12345",
	}

	data, err := json.MarshalIndent(mockAuth, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(authPath, data, 0600)
}

// IsMockAuthEnabled checks if mock authentication should be used
func IsMockAuthEnabled() bool {
	// Check environment variable
	if os.Getenv("ONEMOUNT_USE_MOCK_AUTH") == "true" {
		return true
	}

	// Check if we're in CI environment
	if os.Getenv("CI") == "true" || os.Getenv("GITHUB_ACTIONS") == "true" {
		return true
	}

	return false
}

// SetupMockAuthIfNeeded sets up mock authentication if needed for headless testing
func SetupMockAuthIfNeeded(authPath string) error {
	if !IsMockAuthEnabled() {
		return nil
	}

	// Check if auth file already exists and is valid
	if _, err := os.Stat(authPath); err == nil {
		// File exists, check if it's mock auth
		auth, err := graph.LoadAuthTokens(authPath)
		if err == nil && auth.AccessToken == "mock_access_token_for_testing" {
			// Already has mock auth
			return nil
		}
	}

	// Create mock auth tokens
	return CreateMockAuthTokens(authPath)
}
