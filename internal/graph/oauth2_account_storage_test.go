package graph

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHashAccount tests the account email hashing function
func TestHashAccount(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{
			name:     "lowercase email",
			email:    "user@example.com",
			expected: "b4c9a289323b21a0", // First 16 chars of SHA256
		},
		{
			name:     "uppercase email",
			email:    "USER@EXAMPLE.COM",
			expected: "b4c9a289323b21a0", // Should be same as lowercase
		},
		{
			name:     "mixed case email",
			email:    "User@Example.Com",
			expected: "b4c9a289323b21a0", // Should be same as lowercase
		},
		{
			name:     "email with whitespace",
			email:    "  user@example.com  ",
			expected: "b4c9a289323b21a0", // Should be same after trimming
		},
		{
			name:     "different email",
			email:    "other@example.com",
			expected: "5b71ed5f946240dc", // Different hash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hashAccount(tt.email)

			// Check hash length
			if len(result) != 16 {
				t.Errorf("hashAccount() returned hash of length %d, expected 16", len(result))
			}

			// Check hash value
			if result != tt.expected {
				t.Errorf("hashAccount(%q) = %q, expected %q", tt.email, result, tt.expected)
			}

			// Check hash is hex
			for _, c := range result {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("hashAccount() returned non-hex character: %c", c)
				}
			}
		})
	}
}

// TestHashAccountStability tests that hashing is stable across multiple calls
func TestHashAccountStability(t *testing.T) {
	email := "test@example.com"

	// Hash the same email multiple times
	hash1 := hashAccount(email)
	hash2 := hashAccount(email)
	hash3 := hashAccount(email)

	// All hashes should be identical
	if hash1 != hash2 || hash2 != hash3 {
		t.Errorf("hashAccount() is not stable: got %q, %q, %q", hash1, hash2, hash3)
	}
}

// TestHashAccountCollisionResistance tests that different emails produce different hashes
func TestHashAccountCollisionResistance(t *testing.T) {
	emails := []string{
		"user1@example.com",
		"user2@example.com",
		"user@example1.com",
		"user@example2.com",
		"admin@example.com",
		"test@example.com",
	}

	hashes := make(map[string]string)

	for _, email := range emails {
		hash := hashAccount(email)

		// Check for collisions
		if existingEmail, exists := hashes[hash]; exists {
			t.Errorf("Hash collision detected: %q and %q both hash to %q", email, existingEmail, hash)
		}

		hashes[hash] = email
	}
}

// TestGetAuthTokensPathByAccount tests account-based token path generation
func TestGetAuthTokensPathByAccount(t *testing.T) {
	tests := []struct {
		name         string
		cacheDir     string
		accountEmail string
		wantContains []string
		wantEmpty    bool
	}{
		{
			name:         "valid account",
			cacheDir:     "/home/user/.cache/onemount",
			accountEmail: "user@example.com",
			wantContains: []string{
				"/home/user/.cache/onemount",
				"accounts",
				"b4c9a289323b21a0", // Hash of user@example.com
				"auth_tokens.json",
			},
		},
		{
			name:         "different account",
			cacheDir:     "/home/user/.cache/onemount",
			accountEmail: "other@example.com",
			wantContains: []string{
				"/home/user/.cache/onemount",
				"accounts",
				"5b71ed5f946240dc", // Hash of other@example.com
				"auth_tokens.json",
			},
		},
		{
			name:         "empty account email",
			cacheDir:     "/home/user/.cache/onemount",
			accountEmail: "",
			wantEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetAuthTokensPathByAccount(tt.cacheDir, tt.accountEmail)

			if tt.wantEmpty {
				if result != "" {
					t.Errorf("GetAuthTokensPathByAccount() with empty email = %q, expected empty string", result)
				}
				return
			}

			// Check that path contains expected components
			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("GetAuthTokensPathByAccount() = %q, should contain %q", result, want)
				}
			}

			// Check path structure
			if !strings.HasPrefix(result, tt.cacheDir) {
				t.Errorf("GetAuthTokensPathByAccount() = %q, should start with %q", result, tt.cacheDir)
			}

			if !strings.HasSuffix(result, "auth_tokens.json") {
				t.Errorf("GetAuthTokensPathByAccount() = %q, should end with auth_tokens.json", result)
			}
		})
	}
}

// TestFindAuthTokens tests token search with fallback logic
func TestFindAuthTokens(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		setup        func() (cacheDir, instance, accountEmail string)
		wantLocation string // "account", "instance", "legacy", or "new"
	}{
		{
			name: "account-based location exists",
			setup: func() (string, string, string) {
				cacheDir := filepath.Join(tmpDir, "test1")
				accountEmail := "user@example.com"
				accountPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)

				// Create account-based token file
				os.MkdirAll(filepath.Dir(accountPath), 0700)
				os.WriteFile(accountPath, []byte(`{"access_token":"test"}`), 0600)

				return cacheDir, "test-instance", accountEmail
			},
			wantLocation: "account",
		},
		{
			name: "instance-based location exists (should migrate)",
			setup: func() (string, string, string) {
				cacheDir := filepath.Join(tmpDir, "test2")
				instance := "test-instance"
				accountEmail := "user@example.com"
				instancePath := GetAuthTokensPath(cacheDir, instance)

				// Create instance-based token file
				os.MkdirAll(filepath.Dir(instancePath), 0700)
				os.WriteFile(instancePath, []byte(`{"access_token":"test"}`), 0600)

				return cacheDir, instance, accountEmail
			},
			wantLocation: "account", // Should migrate to account-based
		},
		{
			name: "legacy location exists (should migrate)",
			setup: func() (string, string, string) {
				cacheDir := filepath.Join(tmpDir, "test3")
				accountEmail := "user@example.com"
				legacyPath := GetAuthTokensPathFromCacheDir(cacheDir)

				// Create legacy token file
				os.MkdirAll(filepath.Dir(legacyPath), 0700)
				os.WriteFile(legacyPath, []byte(`{"access_token":"test"}`), 0600)

				return cacheDir, "test-instance", accountEmail
			},
			wantLocation: "account", // Should migrate to account-based
		},
		{
			name: "no existing tokens (should return new account-based path)",
			setup: func() (string, string, string) {
				cacheDir := filepath.Join(tmpDir, "test4")
				return cacheDir, "test-instance", "user@example.com"
			},
			wantLocation: "new",
		},
		{
			name: "no account email (should return legacy path)",
			setup: func() (string, string, string) {
				cacheDir := filepath.Join(tmpDir, "test5")
				return cacheDir, "test-instance", ""
			},
			wantLocation: "legacy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheDir, instance, accountEmail := tt.setup()

			result, err := FindAuthTokens(cacheDir, instance, accountEmail)
			if err != nil {
				t.Errorf("FindAuthTokens() error = %v", err)
				return
			}

			// Verify result based on expected location
			switch tt.wantLocation {
			case "account":
				if accountEmail != "" {
					expectedPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
					if result != expectedPath {
						t.Errorf("FindAuthTokens() = %q, expected account-based path %q", result, expectedPath)
					}
				}
			case "instance":
				expectedPath := GetAuthTokensPath(cacheDir, instance)
				if result != expectedPath {
					t.Errorf("FindAuthTokens() = %q, expected instance-based path %q", result, expectedPath)
				}
			case "legacy":
				expectedPath := GetAuthTokensPathFromCacheDir(cacheDir)
				if result != expectedPath {
					t.Errorf("FindAuthTokens() = %q, expected legacy path %q", result, expectedPath)
				}
			case "new":
				if accountEmail != "" {
					expectedPath := GetAuthTokensPathByAccount(cacheDir, accountEmail)
					if result != expectedPath {
						t.Errorf("FindAuthTokens() = %q, expected new account-based path %q", result, expectedPath)
					}
				}
			}
		})
	}
}

// TestMigrateTokens tests token migration functionality
func TestMigrateTokens(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		setup   func() (oldPath, newPath string)
		wantErr bool
	}{
		{
			name: "successful migration",
			setup: func() (string, string) {
				oldPath := filepath.Join(tmpDir, "old", "auth_tokens.json")
				newPath := filepath.Join(tmpDir, "new", "auth_tokens.json")

				// Create old token file
				os.MkdirAll(filepath.Dir(oldPath), 0700)
				os.WriteFile(oldPath, []byte(`{"access_token":"test123"}`), 0600)

				return oldPath, newPath
			},
			wantErr: false,
		},
		{
			name: "migration with existing new location",
			setup: func() (string, string) {
				oldPath := filepath.Join(tmpDir, "old2", "auth_tokens.json")
				newPath := filepath.Join(tmpDir, "new2", "auth_tokens.json")

				// Create both old and new token files
				os.MkdirAll(filepath.Dir(oldPath), 0700)
				os.WriteFile(oldPath, []byte(`{"access_token":"old"}`), 0600)
				os.MkdirAll(filepath.Dir(newPath), 0700)
				os.WriteFile(newPath, []byte(`{"access_token":"new"}`), 0600)

				return oldPath, newPath
			},
			wantErr: false, // Should not error, just skip migration
		},
		{
			name: "migration with same paths",
			setup: func() (string, string) {
				path := filepath.Join(tmpDir, "same", "auth_tokens.json")

				// Create token file
				os.MkdirAll(filepath.Dir(path), 0700)
				os.WriteFile(path, []byte(`{"access_token":"test"}`), 0600)

				return path, path
			},
			wantErr: false, // Should not error, just skip migration
		},
		{
			name: "migration with missing old file",
			setup: func() (string, string) {
				oldPath := filepath.Join(tmpDir, "missing", "auth_tokens.json")
				newPath := filepath.Join(tmpDir, "new3", "auth_tokens.json")

				// Don't create old file
				return oldPath, newPath
			},
			wantErr: true,
		},
		{
			name: "migration with empty paths",
			setup: func() (string, string) {
				return "", ""
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldPath, newPath := tt.setup()

			err := migrateTokens(oldPath, newPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("migrateTokens() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && oldPath != "" && newPath != "" && oldPath != newPath {
				// Check if new file already existed before migration
				newFileExistedBefore := false
				if _, err := os.Stat(newPath); err == nil {
					// New file existed, verify it wasn't overwritten
					newData, _ := os.ReadFile(newPath)
					if string(newData) == `{"access_token":"new"}` {
						newFileExistedBefore = true
					}
				}

				if !newFileExistedBefore {
					// Verify new file exists
					if _, err := os.Stat(newPath); err != nil {
						t.Errorf("migrateTokens() did not create new file at %q", newPath)
					}

					// Verify old file still exists (not deleted)
					if _, err := os.Stat(oldPath); err != nil {
						t.Errorf("migrateTokens() deleted old file at %q (should preserve it)", oldPath)
					}

					// Verify content was copied correctly
					oldData, _ := os.ReadFile(oldPath)
					newData, _ := os.ReadFile(newPath)
					if string(oldData) != string(newData) {
						t.Errorf("migrateTokens() content mismatch: old=%q, new=%q", oldData, newData)
					}
				}
			}
		})
	}
}

// TestMigrateTokensPermissions tests that migrated tokens have correct permissions
func TestMigrateTokensPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	oldPath := filepath.Join(tmpDir, "old", "auth_tokens.json")
	newPath := filepath.Join(tmpDir, "new", "auth_tokens.json")

	// Create old token file
	os.MkdirAll(filepath.Dir(oldPath), 0700)
	os.WriteFile(oldPath, []byte(`{"access_token":"test"}`), 0600)

	// Migrate
	err := migrateTokens(oldPath, newPath)
	if err != nil {
		t.Fatalf("migrateTokens() error = %v", err)
	}

	// Check new file permissions
	info, err := os.Stat(newPath)
	if err != nil {
		t.Fatalf("Failed to stat new file: %v", err)
	}

	mode := info.Mode()
	if mode.Perm() != 0600 {
		t.Errorf("New file has permissions %o, expected 0600", mode.Perm())
	}

	// Check new directory permissions
	dirInfo, err := os.Stat(filepath.Dir(newPath))
	if err != nil {
		t.Fatalf("Failed to stat new directory: %v", err)
	}

	dirMode := dirInfo.Mode()
	if dirMode.Perm() != 0700 {
		t.Errorf("New directory has permissions %o, expected 0700", dirMode.Perm())
	}
}

// TestFileExists tests the fileExists helper function
func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name  string
		setup func() string
		want  bool
	}{
		{
			name: "file exists",
			setup: func() string {
				path := filepath.Join(tmpDir, "exists.txt")
				os.WriteFile(path, []byte("test"), 0600)
				return path
			},
			want: true,
		},
		{
			name: "file does not exist",
			setup: func() string {
				return filepath.Join(tmpDir, "notexists.txt")
			},
			want: false,
		},
		{
			name: "directory exists (not a file)",
			setup: func() string {
				path := filepath.Join(tmpDir, "dir")
				os.MkdirAll(path, 0700)
				return path
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result := fileExists(path)

			if result != tt.want {
				t.Errorf("fileExists(%q) = %v, want %v", path, result, tt.want)
			}
		})
	}
}

// TestAuthenticateWithAccountStorage_Migration tests the migration behavior
func TestAuthenticateWithAccountStorage_Migration(t *testing.T) {
	// This is a unit test that verifies the migration logic without actual authentication
	// It tests that existing tokens in old locations are found and migrated

	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "cache")
	instance := "test-instance"

	// Create tokens in instance-based location
	instancePath := GetAuthTokensPath(cacheDir, instance)
	os.MkdirAll(filepath.Dir(instancePath), 0700)

	// Create a mock auth token file
	mockAuth := Auth{
		Account:      "test@example.com",
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    9999999999, // Far future
	}
	mockAuth.ToFile(instancePath)

	// Verify instance-based file exists
	if _, err := os.Stat(instancePath); err != nil {
		t.Fatalf("Failed to create instance-based token file: %v", err)
	}

	// Calculate expected account-based path
	expectedAccountPath := GetAuthTokensPathByAccount(cacheDir, mockAuth.Account)

	// Verify account-based file doesn't exist yet
	if _, err := os.Stat(expectedAccountPath); err == nil {
		t.Fatal("Account-based token file should not exist before migration")
	}

	// Note: We can't actually call AuthenticateWithAccountStorage here because it requires
	// real OAuth2 authentication. Instead, we test the migration logic directly.

	// Test the migration logic
	auth := &Auth{}
	if err := auth.FromFile(instancePath); err != nil {
		t.Fatalf("Failed to load tokens from instance-based location: %v", err)
	}

	// Migrate to account-based location
	if auth.Account != "" {
		accountPath := GetAuthTokensPathByAccount(cacheDir, auth.Account)
		if err := migrateTokens(instancePath, accountPath); err != nil {
			t.Fatalf("Failed to migrate tokens: %v", err)
		}

		// Verify account-based file now exists
		if _, err := os.Stat(accountPath); err != nil {
			t.Errorf("Account-based token file should exist after migration: %v", err)
		}

		// Verify content is correct
		migratedAuth := &Auth{}
		if err := migratedAuth.FromFile(accountPath); err != nil {
			t.Fatalf("Failed to load migrated tokens: %v", err)
		}

		if migratedAuth.Account != mockAuth.Account {
			t.Errorf("Migrated account = %q, expected %q", migratedAuth.Account, mockAuth.Account)
		}

		if migratedAuth.AccessToken != mockAuth.AccessToken {
			t.Errorf("Migrated access token = %q, expected %q", migratedAuth.AccessToken, mockAuth.AccessToken)
		}
	}
}
