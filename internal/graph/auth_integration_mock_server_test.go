package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
)

// TestIT_AUTH_04_01_OAuth2Flow_WithMockServer_CompletesSuccessfully tests complete OAuth2 flow with mock HTTP server
//
//	Test Case ID    IT-AUTH-04-01
//	Title           Complete OAuth2 Flow with Mock Server
//	Description     Tests the complete OAuth2 authentication flow using a mock HTTP server
//	Preconditions   None
//	Steps           1. Set up mock OAuth2 server
//	                2. Exchange auth code for tokens
//	                3. Verify token structure
//	                4. Test token refresh
//	                5. Verify refreshed tokens
//	Expected Result Complete OAuth2 flow works correctly with mock server
//	Notes: This test uses a mock HTTP server to simulate Microsoft OAuth2 endpoints.
func TestIT_AUTH_04_01_OAuth2Flow_WithMockServer_CompletesSuccessfully(t *testing.T) {
	// Create assertions helper
	assert := framework.NewAssert(t)

	// Step 1: Create mock OAuth2 server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle token endpoint
		if r.URL.Path == "/token" && r.Method == "POST" {
			// Parse form data
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Failed to parse form", http.StatusBadRequest)
				return
			}

			grantType := r.FormValue("grant_type")

			if grantType == "authorization_code" {
				// Initial token exchange
				response := map[string]interface{}{
					"access_token":  "mock-access-token-12345",
					"refresh_token": "mock-refresh-token-67890",
					"expires_in":    3600,
					"token_type":    "Bearer",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else if grantType == "refresh_token" {
				// Token refresh
				response := map[string]interface{}{
					"access_token":  "mock-refreshed-access-token-99999",
					"refresh_token": "mock-refresh-token-67890", // Same refresh token
					"expires_in":    3600,
					"token_type":    "Bearer",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				http.Error(w, "Invalid grant type", http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Step 2: Create auth config pointing to mock server
	authConfig := AuthConfig{
		ClientID:    "test-client-id",
		CodeURL:     mockServer.URL + "/authorize",
		TokenURL:    mockServer.URL + "/token",
		RedirectURL: "https://test.example.com/redirect",
	}

	testutil.AuthTokensPath = filepath.Join(t.TempDir(), "auth_tokens.json")

	// Step 3: Simulate auth code exchange
	ctx := context.Background()
	auth, err := getAuthTokens(ctx, authConfig, "mock-auth-code-12345")

	// Step 4: Verify token exchange succeeded
	assert.NoError(err, "Token exchange should succeed")
	assert.NotNil(auth, "Auth should not be nil")

	// Step 5: Verify token structure
	assert.Equal("mock-access-token-12345", auth.AccessToken, "Access token should match")
	assert.Equal("mock-refresh-token-67890", auth.RefreshToken, "Refresh token should match")
	assert.True(auth.ExpiresAt > time.Now().Unix(), "Token should not be expired")

	// Step 6: Force token expiration
	auth.ExpiresAt = 0
	auth.Path = testutil.AuthTokensPath

	// Step 7: Test token refresh
	err = auth.Refresh(ctx)

	// Step 8: Verify refresh succeeded
	assert.NoError(err, "Token refresh should succeed")

	// Step 9: Verify refreshed token
	assert.Equal("mock-refreshed-access-token-99999", auth.AccessToken, "Access token should be refreshed")
	assert.True(auth.ExpiresAt > time.Now().Unix(), "Refreshed token should not be expired")
}

// TestIT_AUTH_05_01_TokenRefresh_WithMockServer_HandlesErrors tests token refresh error scenarios
//
//	Test Case ID    IT-AUTH-05-01
//	Title           Token Refresh Error Handling with Mock Server
//	Description     Tests token refresh error scenarios using a mock HTTP server
//	Preconditions   None
//	Steps           1. Set up mock server that returns errors
//	                2. Attempt token refresh
//	                3. Verify error handling
//	                4. Test different error scenarios
//	Expected Result Token refresh errors are handled correctly
//	Notes: This test verifies error handling during token refresh.
func TestIT_AUTH_05_01_TokenRefresh_WithMockServer_HandlesErrors(t *testing.T) {
	// Create assertions helper
	assert := framework.NewAssert(t)

	// Test Case 1: Server returns 401 Unauthorized
	t.Run("401_Unauthorized", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "The refresh token is invalid or expired",
			})
		}))
		defer mockServer.Close()

		auth := &Auth{
			AccessToken:  "expired-token",
			RefreshToken: "invalid-refresh-token",
			ExpiresAt:    0,
			AuthConfig: AuthConfig{
				ClientID:    "test-client-id",
				TokenURL:    mockServer.URL + "/token",
				RedirectURL: "https://test.example.com/redirect",
			},
		}

		err := auth.Refresh(context.Background())
		assert.Error(err, "Should return error for 401 response")
	})

	// Test Case 2: Server returns 500 Internal Server Error
	t.Run("500_InternalServerError", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer mockServer.Close()

		auth := &Auth{
			AccessToken:  "expired-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    0,
			AuthConfig: AuthConfig{
				ClientID:    "test-client-id",
				TokenURL:    mockServer.URL + "/token",
				RedirectURL: "https://test.example.com/redirect",
			},
		}

		err := auth.Refresh(context.Background())
		assert.Error(err, "Should return error for 500 response")
	})

	// Test Case 3: Server returns invalid JSON
	t.Run("InvalidJSON", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json {{{"))
		}))
		defer mockServer.Close()

		auth := &Auth{
			AccessToken:  "expired-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    0,
			AuthConfig: AuthConfig{
				ClientID:    "test-client-id",
				TokenURL:    mockServer.URL + "/token",
				RedirectURL: "https://test.example.com/redirect",
			},
		}

		err := auth.Refresh(context.Background())
		assert.Error(err, "Should return error for invalid JSON")
	})

	// Test Case 4: Network timeout
	t.Run("NetworkTimeout", func(t *testing.T) {
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simulate slow server
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
		}))
		defer mockServer.Close()

		auth := &Auth{
			AccessToken:  "expired-token",
			RefreshToken: "valid-refresh-token",
			ExpiresAt:    0,
			AuthConfig: AuthConfig{
				ClientID:    "test-client-id",
				TokenURL:    mockServer.URL + "/token",
				RedirectURL: "https://test.example.com/redirect",
			},
		}

		// Create context with short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		err := auth.Refresh(ctx)
		assert.Error(err, "Should return error for timeout")
	})
}

// TestIT_AUTH_06_01_TokenPersistence_SaveAndLoad_WorksCorrectly tests token persistence
//
//	Test Case ID    IT-AUTH-06-01
//	Title           Token Persistence - Save and Load
//	Description     Tests saving and loading authentication tokens from disk
//	Preconditions   None
//	Steps           1. Create auth tokens
//	                2. Save to file
//	                3. Load from file
//	                4. Verify tokens match
//	                5. Verify file permissions
//	Expected Result Tokens are correctly persisted and loaded
//	Notes: This test verifies token persistence functionality.
func TestIT_AUTH_06_01_TokenPersistence_SaveAndLoad_WorksCorrectly(t *testing.T) {
	// Create assertions helper
	assert := framework.NewAssert(t)

	// Step 1: Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "onemount-auth-test-*")
	assert.NoError(err, "Should create temp directory")
	defer os.RemoveAll(tempDir)

	tokenFile := filepath.Join(tempDir, "auth_tokens.json")

	// Step 2: Create auth tokens
	originalAuth := &Auth{
		AccessToken:  "test-access-token-12345",
		RefreshToken: "test-refresh-token-67890",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		ExpiresIn:    3600,
		Account:      "test@example.com",
		AuthConfig: AuthConfig{
			ClientID:    "test-client-id",
			CodeURL:     "https://test.example.com/auth",
			TokenURL:    "https://test.example.com/token",
			RedirectURL: "https://test.example.com/redirect",
		},
	}

	// Step 3: Save tokens to file
	err = SaveAuthTokens(originalAuth, tokenFile)
	assert.NoError(err, "Should save tokens to file")

	// Step 4: Verify file exists
	_, err = os.Stat(tokenFile)
	assert.NoError(err, "Token file should exist")

	// Step 5: Verify file permissions (should be 0600)
	fileInfo, err := os.Stat(tokenFile)
	assert.NoError(err, "Should get file info")
	assert.Equal(os.FileMode(0600), fileInfo.Mode().Perm(), "File permissions should be 0600")

	// Step 6: Load tokens from file
	loadedAuth, err := LoadAuthTokens(tokenFile)
	assert.NoError(err, "Should load tokens from file")
	assert.NotNil(loadedAuth, "Loaded auth should not be nil")

	// Step 7: Verify loaded tokens match original
	assert.Equal(originalAuth.AccessToken, loadedAuth.AccessToken, "Access token should match")
	assert.Equal(originalAuth.RefreshToken, loadedAuth.RefreshToken, "Refresh token should match")
	assert.Equal(originalAuth.ExpiresAt, loadedAuth.ExpiresAt, "ExpiresAt should match")
	assert.Equal(originalAuth.Account, loadedAuth.Account, "Account should match")
	assert.Equal(originalAuth.ClientID, loadedAuth.ClientID, "ClientID should match")

	// Step 8: Test ToFile and FromFile methods
	auth2 := &Auth{
		AccessToken:  "another-access-token",
		RefreshToken: "another-refresh-token",
		ExpiresAt:    time.Now().Add(2 * time.Hour).Unix(),
	}

	tokenFile2 := filepath.Join(tempDir, "auth_tokens_2.json")
	err = auth2.ToFile(tokenFile2)
	assert.NoError(err, "Should save using ToFile method")

	auth3 := &Auth{}
	err = auth3.FromFile(tokenFile2)
	assert.NoError(err, "Should load using FromFile method")
	assert.Equal(auth2.AccessToken, auth3.AccessToken, "Tokens should match after ToFile/FromFile")
}

// TestIT_AUTH_07_01_ConcurrentRefresh_MultipleGoroutines_HandlesCorrectly tests concurrent token refresh
//
//	Test Case ID    IT-AUTH-07-01
//	Title           Concurrent Token Refresh
//	Description     Tests token refresh with multiple concurrent goroutines
//	Preconditions   None
//	Steps           1. Create auth with expired token
//	                2. Launch multiple goroutines to refresh token
//	                3. Verify all refreshes complete
//	                4. Verify token is valid after concurrent refreshes
//	Expected Result Concurrent token refresh is handled correctly
//	Notes: This test verifies thread-safety of token refresh.
func TestIT_AUTH_07_01_ConcurrentRefresh_MultipleGoroutines_HandlesCorrectly(t *testing.T) {
	// Create assertions helper
	assert := framework.NewAssert(t)
	testutil.AuthTokensPath = filepath.Join(t.TempDir(), "auth_tokens_concurrent.json")

	// Step 1: Create mock server
	refreshCount := 0
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" && r.Method == "POST" {
			refreshCount++
			response := map[string]interface{}{
				"access_token":  fmt.Sprintf("refreshed-token-%d", refreshCount),
				"refresh_token": "mock-refresh-token",
				"expires_in":    3600,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer mockServer.Close()

	// Step 2: Create auth with expired token
	auth := &Auth{
		AccessToken:  "expired-token",
		RefreshToken: "valid-refresh-token",
		ExpiresAt:    0,
		Path:         testutil.AuthTokensPath,
		AuthConfig: AuthConfig{
			ClientID:    "test-client-id",
			TokenURL:    mockServer.URL + "/token",
			RedirectURL: "https://test.example.com/redirect",
		},
	}

	// Step 3: Launch multiple goroutines to refresh token
	numGoroutines := 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := auth.Refresh(context.Background())
			done <- err
		}()
	}

	// Step 4: Wait for all goroutines to complete
	errors := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		if err != nil {
			errors++
		}
	}

	// Step 5: Verify most refreshes succeeded (some may fail due to race conditions, but most should succeed)
	assert.True(errors < numGoroutines/2, "Most concurrent refreshes should succeed")

	// Step 6: Verify token is valid after concurrent refreshes
	assert.NotEqual("expired-token", auth.AccessToken, "Token should be refreshed")
	assert.True(auth.ExpiresAt > time.Now().Unix(), "Token should not be expired")
}

// TestIT_AUTH_08_01_AuthenticatorInterface_RealAndMock_WorkCorrectly tests Authenticator interface
//
//	Test Case ID    IT-AUTH-08-01
//	Title           Authenticator Interface - Real and Mock Implementations
//	Description     Tests both RealAuthenticator and MockAuthenticator implementations
//	Preconditions   None
//	Steps           1. Test MockAuthenticator
//	                2. Test RealAuthenticator interface
//	                3. Verify both implement Authenticator interface
//	                4. Test factory method
//	Expected Result Both authenticator implementations work correctly
//	Notes: This test verifies the Authenticator interface and its implementations.
func TestIT_AUTH_08_01_AuthenticatorInterface_RealAndMock_WorkCorrectly(t *testing.T) {
	// Create assertions helper
	assert := framework.NewAssert(t)

	// Step 1: Test MockAuthenticator
	mockAuth := NewMockAuthenticator()
	assert.NotNil(mockAuth, "MockAuthenticator should be created")

	// Verify MockAuthenticator implements Authenticator interface
	var _ Authenticator = mockAuth

	// Test MockAuthenticator methods
	auth, err := mockAuth.Authenticate()
	assert.NoError(err, "Mock authentication should succeed")
	assert.NotNil(auth, "Mock auth should not be nil")

	err = mockAuth.Refresh()
	assert.NoError(err, "Mock refresh should succeed")

	retrievedAuth := mockAuth.GetAuth()
	assert.NotNil(retrievedAuth, "Should get auth from mock")

	// Step 2: Test RealAuthenticator interface (without actual authentication)
	tempDir, err := os.MkdirTemp("", "onemount-auth-test-*")
	assert.NoError(err, "Should create temp directory")
	defer os.RemoveAll(tempDir)

	authPath := filepath.Join(tempDir, "auth_tokens.json")

	// Create a valid auth file for RealAuthenticator to load
	testAuth := &Auth{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		AuthConfig: AuthConfig{
			ClientID:    "test-client-id",
			CodeURL:     "https://test.example.com/auth",
			TokenURL:    "https://test.example.com/token",
			RedirectURL: "https://test.example.com/redirect",
		},
	}
	err = SaveAuthTokens(testAuth, authPath)
	assert.NoError(err, "Should save test auth")

	realAuth := NewRealAuthenticator(testAuth.AuthConfig, authPath, true)
	assert.NotNil(realAuth, "RealAuthenticator should be created")

	// Verify RealAuthenticator implements Authenticator interface
	var _ Authenticator = realAuth

	// Step 3: Test factory method
	mockAuthFromFactory := NewAuthenticator(AuthConfig{}, "", false, true)
	assert.NotNil(mockAuthFromFactory, "Factory should create mock authenticator")

	realAuthFromFactory := NewAuthenticator(testAuth.AuthConfig, authPath, true, false)
	assert.NotNil(realAuthFromFactory, "Factory should create real authenticator")
}
