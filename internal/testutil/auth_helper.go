// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"os"
	"path/filepath"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/rs/zerolog/log"
	"time"
)

// GetTestAuth attempts to load authentication tokens from the .auth_tokens.json file
// if it exists, otherwise it returns a mock Auth object.
//
// This function helps tests reuse existing authentication tokens to avoid
// re-authentication when re-running tests.
func GetTestAuth() *graph.Auth {
	// Ensure test directories exist
	if err := EnsureTestDirectories(); err != nil {
		log.Warn().Err(err).Msg("Failed to ensure test directories exist")
	}

	// Check if the auth tokens file exists
	if _, err := os.Stat(AuthTokensPath); err == nil {
		// File exists, try to load it
		auth, err := graph.LoadAuthTokens(AuthTokensPath)
		if err == nil {
			log.Debug().Str("path", AuthTokensPath).Msg("Loaded existing auth tokens for test")
			return auth
		}
		// Log the error but continue to create a mock auth object
		log.Warn().Err(err).Str("path", AuthTokensPath).Msg("Failed to load existing auth tokens for test, creating mock auth")
	}

	// Create a mock auth object
	auth := &graph.Auth{
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		Account:      "mock@example.com",
		Path:         AuthTokensPath,
	}

	return auth
}

// EnsureTestDirectories ensures that all required test directories exist
func EnsureTestDirectories() error {
	// Create the test sandbox directory if it doesn't exist
	if err := os.MkdirAll(TestSandboxDir, 0755); err != nil {
		return err
	}

	// Create the temporary directory if it doesn't exist
	if err := os.MkdirAll(TestSandboxTmpDir, 0755); err != nil {
		return err
	}

	// Create the logs directory if it doesn't exist
	logsDir := filepath.Dir(TestLogPath)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	return nil
}
