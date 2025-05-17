// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"fmt"
	"time"
)

// MockAuth represents a mock authentication object for testing
// This is a simplified version of the graph.Auth struct that doesn't require importing the graph package
type MockAuth struct {
	// AuthConfig fields
	ClientID    string
	CodeURL     string
	TokenURL    string
	RedirectURL string

	// Auth fields
	Account      string
	ExpiresIn    int64
	ExpiresAt    int64
	AccessToken  string
	RefreshToken string
	Path         string
}

// GetMockAuth returns a mock authentication object for testing
// This function can be used by tests that need an Auth object but want to avoid import cycles
func GetMockAuth() *MockAuth {
	// Create a mock auth object with default values
	return &MockAuth{
		// AuthConfig fields
		ClientID:    "3470c3fa-bc10-45ab-a0a9-2d30836485d1", // Default client ID
		CodeURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
		RedirectURL: "https://login.live.com/oauth20_desktop.srf",

		// Auth fields
		Account:      "mock@example.com",
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		AccessToken:  "mock-access-token",
		RefreshToken: "mock-refresh-token",
		Path:         AuthTokensPath,
	}
}

// String returns a string representation of the MockAuth object
func (m *MockAuth) String() string {
	return fmt.Sprintf("MockAuth{Account: %s, AccessToken: %s, RefreshToken: %s, ExpiresAt: %d}",
		m.Account, m.AccessToken, m.RefreshToken, m.ExpiresAt)
}
