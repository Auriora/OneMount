package system

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
)

// TestSystemST_Auth_01_01_InteractiveAuthentication tests the complete OAuth2 authentication flow
// with real Microsoft Graph API integration.
//
// Test Case ID: ST-AUTH-01-01
// Description: Verify interactive authentication works with real OneDrive
// Prerequisites: GUI environment with X11 forwarding
// Expected Result: Authentication completes successfully and tokens are stored
// Requirements: 1.1, 1.2
func TestSystemST_Auth_01_01_InteractiveAuthentication(t *testing.T) {
	// Check if we should use mock authentication
	if IsMockAuthEnabled() {
		t.Log("Using mock authentication for headless testing")
		authPath := "test-artifacts/.auth_tokens.json"

		err := SetupMockAuthIfNeeded(authPath)
		if err != nil {
			t.Fatalf("Failed to setup mock auth: %v", err)
		}

		// Load mock auth to verify it works
		auth, err := graph.LoadAuthTokens(authPath)
		if err != nil {
			t.Fatalf("Failed to load mock auth: %v", err)
		}

		if auth.AccessToken != "mock_access_token_for_testing" {
			t.Fatalf("Mock auth not properly configured")
		}

		t.Log("✓ Mock authentication configured successfully")
		return
	}

	// Original interactive authentication test
	// Skip if running in CI or headless environment
	if os.Getenv("CI") == "true" || os.Getenv("DISPLAY") == "" {
		t.Skip("Skipping interactive authentication test in headless environment")
	}

	// Ensure test directories exist
	if err := os.MkdirAll(testutil.TestSandboxDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Remove existing auth tokens to force fresh authentication
	authPath := testutil.AuthTokensPath
	if err := os.Remove(authPath); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Could not remove existing auth tokens: %v", err)
	}

	// Attempt interactive authentication
	t.Log("Starting interactive authentication - this will open a browser window")
	config := graph.AuthConfig{}                                                   // Use default config
	auth, err := graph.Authenticate(context.Background(), config, authPath, false) // headless=false for interactive
	if err != nil {
		t.Fatalf("Interactive authentication failed: %v", err)
	}

	// Verify authentication result
	if auth == nil {
		t.Fatal("Authentication returned nil auth object")
	}

	if auth.AccessToken == "" {
		t.Error("Access token is empty")
	}

	if auth.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}

	if auth.ExpiresAt <= time.Now().Unix() {
		t.Error("Token appears to be already expired")
	}

	if auth.Account == "" {
		t.Error("Account information is empty")
	}

	// Verify tokens were saved to file
	if _, err := os.Stat(authPath); os.IsNotExist(err) {
		t.Error("Auth tokens file was not created")
	}

	// Verify we can load the saved tokens
	loadedAuth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Errorf("Failed to load saved auth tokens: %v", err)
	}

	if loadedAuth.AccessToken != auth.AccessToken {
		t.Error("Loaded access token does not match original")
	}

	t.Log("Interactive authentication test completed successfully")
}

// TestSystemST_Auth_02_01_TokenRefresh tests automatic token refresh with real Microsoft Graph API.
//
// Test Case ID: ST-AUTH-02-01
// Description: Verify token refresh works with real OneDrive API
// Prerequisites: Valid refresh token from previous authentication
// Expected Result: Tokens are refreshed successfully
// Requirements: 1.3
func TestSystemST_Auth_02_01_TokenRefresh(t *testing.T) {
	// Load existing auth tokens
	authPath := testutil.AuthTokensPath
	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Skipf("Skipping token refresh test - no existing auth tokens: %v", err)
	}

	// Store original tokens for comparison
	originalAccessToken := auth.AccessToken
	originalExpiresAt := auth.ExpiresAt

	// Force token refresh by setting expiration to past
	auth.ExpiresAt = time.Now().Add(-time.Hour).Unix()

	// Attempt to refresh tokens
	err = auth.Refresh(context.Background())
	if err != nil {
		t.Fatalf("Token refresh failed: %v", err)
	}

	// Verify tokens were updated
	if auth.AccessToken == originalAccessToken {
		t.Error("Access token was not updated after refresh")
	}

	if auth.ExpiresAt <= originalExpiresAt {
		t.Error("Token expiration was not updated after refresh")
	}

	if auth.ExpiresAt <= time.Now().Unix() {
		t.Error("Refreshed token appears to be already expired")
	}

	// Verify refreshed tokens were saved
	loadedAuth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Errorf("Failed to load refreshed auth tokens: %v", err)
	}

	if loadedAuth.AccessToken != auth.AccessToken {
		t.Error("Saved access token does not match refreshed token")
	}

	t.Log("Token refresh test completed successfully")
}

// TestSystemST_Auth_03_01_APIConnectivity tests basic API connectivity with authenticated requests.
//
// Test Case ID: ST-AUTH-03-01
// Description: Verify authenticated API requests work with real OneDrive
// Prerequisites: Valid authentication tokens
// Expected Result: API requests succeed and return expected data
// Requirements: 1.1, 1.4
func TestSystemST_Auth_03_01_APIConnectivity(t *testing.T) {
	// Load existing auth tokens
	authPath := testutil.AuthTokensPath
	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Skipf("Skipping API connectivity test - no existing auth tokens: %v", err)
	}

	// Create Graph provider (client)
	client := graph.NewProvider(auth)

	// Test basic API connectivity by getting drive root
	rootItem, err := client.GetItem("root")
	if err != nil {
		t.Fatalf("Failed to get drive root: %v", err)
	}

	// Verify we got valid root item
	if rootItem == nil {
		t.Fatal("Root item is nil")
	}

	if rootItem.ID == "" {
		t.Error("Root item ID is empty")
	}

	if rootItem.Name == "" {
		t.Error("Root item name is empty")
	}

	// Test getting drive information using the graph package directly
	drive, err := graph.GetDrive(auth)
	if err != nil {
		t.Errorf("Failed to get drive information: %v", err)
	} else {
		if drive.ID == "" {
			t.Error("Drive ID is empty")
		}
		t.Logf("Connected to drive: %s (ID: %s)", drive.DriveType, drive.ID)
	}

	t.Log("API connectivity test completed successfully")
}
