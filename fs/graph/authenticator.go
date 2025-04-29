package graph

import (
	"context"
)

// Authenticator is an interface for authentication operations
type Authenticator interface {
	// Authenticate performs authentication and returns Auth information
	Authenticate() (*Auth, error)

	// AuthenticateWithContext performs authentication with context and returns Auth information
	AuthenticateWithContext(ctx context.Context) (*Auth, error)

	// Refresh refreshes the authentication tokens if expired
	Refresh() error

	// RefreshWithContext refreshes the authentication tokens with context if expired
	RefreshWithContext(ctx context.Context) error

	// GetAuth returns the current Auth information
	GetAuth() *Auth
}

// RealAuthenticator implements Authenticator interface for real authentication
type RealAuthenticator struct {
	config   AuthConfig
	authPath string
	headless bool
	auth     *Auth
}

// NewRealAuthenticator creates a new RealAuthenticator
func NewRealAuthenticator(config AuthConfig, authPath string, headless bool) *RealAuthenticator {
	return &RealAuthenticator{
		config:   config,
		authPath: authPath,
		headless: headless,
	}
}

// Authenticate performs real authentication and returns Auth information
func (ra *RealAuthenticator) Authenticate() (*Auth, error) {
	return ra.AuthenticateWithContext(context.Background())
}

// AuthenticateWithContext performs real authentication with context and returns Auth information
func (ra *RealAuthenticator) AuthenticateWithContext(ctx context.Context) (*Auth, error) {
	var err error
	ra.auth, err = Authenticate(ctx, ra.config, ra.authPath, ra.headless)
	return ra.auth, err
}

// Refresh refreshes the authentication tokens if expired
func (ra *RealAuthenticator) Refresh() error {
	return ra.RefreshWithContext(context.Background())
}

// RefreshWithContext refreshes the authentication tokens with context if expired
func (ra *RealAuthenticator) RefreshWithContext(ctx context.Context) error {
	if ra.auth == nil {
		var err error
		ra.auth, err = ra.AuthenticateWithContext(ctx)
		return err
	}
	return ra.auth.Refresh(ctx)
}

// GetAuth returns the current Auth information
func (ra *RealAuthenticator) GetAuth() *Auth {
	return ra.auth
}

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
func (ma *MockAuthenticator) AuthenticateWithContext(ctx context.Context) (*Auth, error) {
	return ma.auth, nil
}

// Refresh refreshes the mock authentication tokens (no-op for mock)
func (ma *MockAuthenticator) Refresh() error {
	return ma.RefreshWithContext(context.Background())
}

// RefreshWithContext refreshes the mock authentication tokens with context (no-op for mock)
func (ma *MockAuthenticator) RefreshWithContext(ctx context.Context) error {
	return nil
}

// GetAuth returns the current Auth information
func (ma *MockAuthenticator) GetAuth() *Auth {
	return ma.auth
}

// NewAuthenticator creates a new Authenticator based on the configuration
func NewAuthenticator(config AuthConfig, authPath string, headless bool, isMock bool) Authenticator {
	if isMock {
		return NewMockAuthenticator()
	}
	return NewRealAuthenticator(config, authPath, headless)
}
