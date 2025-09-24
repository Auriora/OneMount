package graph

import (
	"context"
)

// MockAuthenticator implements Authenticator interface for mock authentication
type MockAuthenticator struct {
	auth *Auth
}

// NewMockAuthenticator creates a new MockAuthenticator
func NewMockAuthenticator() *MockAuthenticator {
	mockClient := NewMockGraphClient()
	return &MockAuthenticator{
		auth: &mockClient.Auth,
	}
}

// Authenticate performs mock authentication and returns Auth information
func (ma *MockAuthenticator) Authenticate() (*Auth, error) {
	return ma.AuthenticateWithContext(context.Background())
}

// AuthenticateWithContext performs mock authentication with context and returns Auth information
func (ma *MockAuthenticator) AuthenticateWithContext(_ctx context.Context) (*Auth, error) {
	return ma.auth, nil
}

// Refresh refreshes the mock authentication tokens (no-op for mock)
func (ma *MockAuthenticator) Refresh() error {
	return ma.RefreshWithContext(context.Background())
}

// RefreshWithContext refreshes the mock authentication tokens with context (no-op for mock)
func (ma *MockAuthenticator) RefreshWithContext(_ctx context.Context) error {
	return nil
}

// GetAuth returns the current Auth information
func (ma *MockAuthenticator) GetAuth() *Auth {
	return ma.auth
}
