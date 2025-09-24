// Package testutil provides testing utilities for the OneMount project.
package helpers

import (
	"github.com/auriora/onemount/internal/testutil"
	"os"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/logging"
	"time"
)

// GetTestAuth attempts to load authentication tokens from the .auth_tokens.json file
// if it exists, otherwise it returns a mock Auth object.
//
// This function helps tests reuse existing authentication tokens to avoid
// re-authentication when re-running tests.
//
// Set useMockAuth to true to force the use of mock auth instead of loading from the auth tokens file.
func GetTestAuth() *graph.Auth {
	// Ensure test directories exist
	if err := EnsureTestDirectories(); err != nil {
		logging.Warn().Err(err).Msg("Failed to ensure test directories exist")
	}

	// Check if the auth tokens file exists
	if _, err := os.Stat(testutil.AuthTokensPath); err == nil {
		// File exists, try to load it
		auth, err := graph.LoadAuthTokens(testutil.AuthTokensPath)
		if err == nil {
			logging.Debug().Str("path", testutil.AuthTokensPath).Msg("Loaded existing auth tokens for test")
			return auth
		}
		// Log the error but continue to create a mock auth object
		logging.Warn().Err(err).Str("path", testutil.AuthTokensPath).Msg("Failed to load existing auth tokens for test, creating mock auth")
	}

	return createMockAuth()
}

// createMockAuth creates a mock Auth object for testing
func createMockAuth() *graph.Auth {
	// Create a mock auth object
	auth := &graph.Auth{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		Account:      "mock@example.com",
		Path:         testutil.AuthTokensPath,
	}

	return auth
}
