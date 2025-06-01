// Package system provides CI-specific authentication for system tests
package system

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/auriora/onemount/pkg/graph"
)

// CIAuthConfig holds configuration for CI authentication
type CIAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TenantID     string `json:"tenant_id"`
	Scope        string `json:"scope"`
}

// CreateCIAuth creates authentication for CI environments using service principal
func CreateCIAuth() (*graph.Auth, error) {
	// Check if we're in CI environment
	if !isCI() {
		return nil, fmt.Errorf("CI authentication should only be used in CI environments")
	}

	// Get credentials from environment variables
	config := CIAuthConfig{
		ClientID:     os.Getenv("AZURE_CLIENT_ID"),
		ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		TenantID:     os.Getenv("AZURE_TENANT_ID"),
		Scope:        "https://graph.microsoft.com/.default",
	}

	// Validate required environment variables
	if config.ClientID == "" {
		return nil, fmt.Errorf("AZURE_CLIENT_ID environment variable is required")
	}
	if config.ClientSecret == "" {
		return nil, fmt.Errorf("AZURE_CLIENT_SECRET environment variable is required")
	}
	if config.TenantID == "" {
		return nil, fmt.Errorf("AZURE_TENANT_ID environment variable is required")
	}

	// Get access token using client credentials flow
	token, err := getServicePrincipalToken(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get service principal token: %w", err)
	}

	// Create auth object
	auth := &graph.Auth{
		AccessToken:  token.AccessToken,
		RefreshToken: "", // Service principal doesn't use refresh tokens
		ExpiresAt:    time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).Unix(),
	}

	return auth, nil
}

// TokenResponse represents the response from Azure AD token endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

// getServicePrincipalToken gets an access token using client credentials flow
func getServicePrincipalToken(config CIAuthConfig) (*TokenResponse, error) {
	// This would typically use an HTTP client to call Azure AD token endpoint
	// For now, we'll return a placeholder that shows the structure

	// In a real implementation, you would:
	// 1. Make POST request to https://login.microsoftonline.com/{tenant}/oauth2/v2.0/token
	// 2. With form data: grant_type=client_credentials, client_id=..., client_secret=..., scope=...
	// 3. Parse the JSON response

	return nil, fmt.Errorf("service principal authentication not yet implemented - requires HTTP client integration")
}

// isCI checks if we're running in a CI environment
func isCI() bool {
	ciEnvVars := []string{
		"CI",             // Generic CI indicator
		"GITHUB_ACTIONS", // GitHub Actions
		"GITLAB_CI",      // GitLab CI
		"JENKINS_URL",    // Jenkins
		"TRAVIS",         // Travis CI
		"CIRCLECI",       // CircleCI
	}

	for _, envVar := range ciEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// SaveCITokensForTesting saves CI tokens in the expected format for system tests
func SaveCITokensForTesting(auth *graph.Auth, path string) error {
	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create tokens file in the format expected by system tests
	tokensData := map[string]interface{}{
		"access_token":  auth.AccessToken,
		"refresh_token": auth.RefreshToken,
		"expires_at":    auth.ExpiresAt,
		"account":       "ci-test-account",
	}

	// Write to file with restricted permissions
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create tokens file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(tokensData); err != nil {
		return fmt.Errorf("failed to write tokens: %w", err)
	}

	return nil
}
