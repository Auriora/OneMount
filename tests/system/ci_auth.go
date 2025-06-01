// Package system provides CI-specific authentication for system tests
package system

import (
	"os"
)

// CIAuthConfig holds configuration for CI authentication
type CIAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	TenantID     string `json:"tenant_id"`
	Scope        string `json:"scope"`
}

// TokenResponse represents the response from Azure AD token endpoint
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
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
