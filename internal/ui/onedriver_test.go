package ui

import (
	"github.com/bcherrington/onemount/internal/testutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMountpointIsValid tests that we can detect a mountpoint as valid appropriately
func TestMountpointIsValid(t *testing.T) {
	t.Parallel()
	// Create a unique test directory name to avoid conflicts in parallel tests
	testDirName := "_test_" + t.Name()
	testDirName = strings.ReplaceAll(testDirName, "/", "_") // Replace slashes for subtests

	// Create a test directory and file
	err := os.Mkdir(testDirName, 0755)
	require.NoError(t, err, "Failed to create test directory")

	err = os.WriteFile(filepath.Join(testDirName, ".example"), []byte("some text\n"), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Setup cleanup to remove the test directory after test completes or fails
	t.Cleanup(func() {
		if err := os.RemoveAll(testDirName); err != nil {
			t.Logf("Warning: Failed to clean up test directory %s: %v", testDirName, err)
		}
	})

	// For the test mount directory, ensure it exists and is empty
	// This is test-specific setup code that was previously in the production code
	mountDir := testutil.TestMountPoint
	if _, err := os.Stat(mountDir); err == nil {
		// Directory exists, make sure it's empty
		dirents, err := os.ReadDir(mountDir)
		if err != nil {
			// If we can't read the directory, it might be a stale mount point
			t.Logf("Mount directory exists but can't be read, attempting to remove: %v", err)
			err := os.Remove(mountDir)
			if err != nil {
				t.Logf("Failed to remove stale mount directory: %v", err)
				return
			}
			err = os.Mkdir(mountDir, 0700)
			require.NoError(t, err, "Failed to recreate mount directory")
		} else if len(dirents) > 0 {
			// Directory has contents, log them
			t.Logf("Mount directory contents (%d items):", len(dirents))
			for _, dirent := range dirents {
				t.Logf("  %s", dirent.Name())
			}
			// Remove all files in the directory
			for _, dirent := range dirents {
				err := os.RemoveAll(filepath.Join(mountDir, dirent.Name()))
				if err != nil {
					t.Logf("Warning: Failed to remove %s: %v", dirent.Name(), err)
				}
			}
		}
	} else if os.IsNotExist(err) {
		// Directory doesn't exist, create it
		err = os.MkdirAll(mountDir, 0700)
		require.NoError(t, err, "Failed to create mount directory")
	}

	// Define test cases
	tests := []struct {
		name       string
		mountpoint string
		expected   bool
		reason     string
	}{
		{
			name:       "EmptyPath",
			mountpoint: "",
			expected:   false,
			reason:     "Empty path should not be valid",
		},
		{
			name:       "FsDirectory",
			mountpoint: "fs",
			expected:   false,
			reason:     "fs directory should not be a valid mountpoint",
		},
		{
			name:       "NonexistentPath",
			mountpoint: "does_not_exist",
			expected:   false,
			reason:     "Nonexistent path should not be valid",
		},
		{
			name:       "MountDirectory",
			mountpoint: testutil.TestMountPoint,
			expected:   true,
			reason:     "mount directory should be a valid mountpoint",
		},
		{
			name:       "TestDirectory",
			mountpoint: testDirName,
			expected:   false,
			reason:     "Test directory should not be a valid mountpoint",
		},
		{
			name:       "TestFile",
			mountpoint: filepath.Join(testDirName, ".example"),
			expected:   false,
			reason:     "File should not be a valid mountpoint",
		},
	}

	// Run each test case as a subtest
	for _, tc := range tests {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Check if the mountpoint is valid
			isValid := MountpointIsValid(tc.mountpoint)

			// Assert the result
			if tc.expected {
				require.True(t, isValid,
					"Expected mountpoint %q to be valid: %s", tc.mountpoint, tc.reason)
			} else {
				require.False(t, isValid,
					"Expected mountpoint %q to be invalid: %s", tc.mountpoint, tc.reason)
			}
		})
	}
}

// TestHomeEscapeUnescape tests that we can convert paths from ~/some_path to /home/username/some_path and back
func TestHomeEscapeUnescape(t *testing.T) {
	t.Parallel()
	homedir, err := os.UserHomeDir()
	require.NoError(t, err, "Failed to get user home directory")

	// Define test cases
	tests := []struct {
		name      string
		unescaped string
		escaped   string
		desc      string
	}{
		{
			name:      "HomeDirectory",
			unescaped: homedir + "/test",
			escaped:   "~/test",
			desc:      "Path in home directory",
		},
		{
			name:      "NonHomeDirectory",
			unescaped: "/opt/test",
			escaped:   "/opt/test",
			desc:      "Path outside home directory",
		},
		{
			name:      "PathWithTilde",
			unescaped: "/opt/test/~test.lock#",
			escaped:   "/opt/test/~test.lock#",
			desc:      "Path with tilde character",
		},
	}

	// Run each test case as a subtest
	for _, tc := range tests {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Test escaping home directory
			t.Run("Escape", func(t *testing.T) {
				t.Parallel()
				escaped := EscapeHome(tc.unescaped)
				require.Equal(t, tc.escaped, escaped,
					"Failed to correctly escape home in %q (%s)", tc.unescaped, tc.desc)
			})

			// Test unescaping home directory
			t.Run("Unescape", func(t *testing.T) {
				t.Parallel()
				unescaped := UnescapeHome(tc.escaped)
				require.Equal(t, tc.unescaped, unescaped,
					"Failed to correctly unescape home in %q (%s)", tc.escaped, tc.desc)
			})
		})
	}
}

// TestGetAccountName tests the GetAccountName function with various scenarios
func TestGetAccountName(t *testing.T) {
	t.Parallel()

	// Get current working directory and create escaped path for test instance
	wd, err := os.Getwd()
	require.NoError(t, err, "Failed to get working directory")
	escaped := unit.UnitNamePathEscape(filepath.Join(wd, testutil.TestMountPoint))

	// Get user cache directory
	cacheDir, err := os.UserCacheDir()
	require.NoError(t, err, "Failed to get user cache directory")

	// Create test directory structure
	testCacheDir := filepath.Join(cacheDir, "onemount")
	testInstanceDir := filepath.Join(testCacheDir, escaped)
	err = os.MkdirAll(testInstanceDir, 0700)
	require.NoError(t, err, "Failed to create test instance directory")

	// Define paths for test files
	validTokensPath := filepath.Join(testInstanceDir, "auth_tokens.json")
	invalidTokensPath := filepath.Join(testInstanceDir, "invalid_tokens.json")
	emptyAccountPath := filepath.Join(testInstanceDir, "empty_account.json")

	// Setup cleanup to remove test files after test completes or fails
	t.Cleanup(func() {
		// Only remove the files we create in this test, not the valid tokens file
		// that might be used by other tests
		if err := os.Remove(invalidTokensPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up invalid tokens file: %v", err)
		}
		if err := os.Remove(emptyAccountPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up empty account file: %v", err)
		}
	})

	// Copy auth tokens to cache dir if it doesn't already exist
	// (CI runners will not have this file yet)
	if _, err := os.Stat(validTokensPath); err != nil {
		cmd := exec.Command("cp", testutil.AuthTokensPath, validTokensPath)
		err = cmd.Run()
		require.NoError(t, err, "Failed to copy auth tokens file")
	}

	// Create invalid JSON file
	err = os.WriteFile(invalidTokensPath, []byte("this is not valid json"), 0644)
	require.NoError(t, err, "Failed to create invalid tokens file")

	// Create JSON file with empty account
	err = os.WriteFile(emptyAccountPath, []byte(`{"Account":"","AccessToken":"test","RefreshToken":"test","ExpiresAt":0}`), 0644)
	require.NoError(t, err, "Failed to create empty account file")

	// Define test cases
	tests := []struct {
		name          string
		instance      string
		tokenFile     string
		expectedError bool
		errorContains string
	}{
		{
			name:          "ValidTokens",
			instance:      escaped,
			tokenFile:     "auth_tokens.json",
			expectedError: false,
		},
		{
			name:          "InvalidJSON",
			instance:      escaped,
			tokenFile:     "invalid_tokens.json",
			expectedError: true,
			errorContains: "invalid",
		},
		{
			name:          "EmptyAccount",
			instance:      escaped,
			tokenFile:     "empty_account.json",
			expectedError: false, // Not an error, just returns empty string
		},
		{
			name:          "NonexistentFile",
			instance:      escaped,
			tokenFile:     "nonexistent.json",
			expectedError: true,
			errorContains: "no such file",
		},
	}

	// Run each test case as a subtest
	for _, tc := range tests {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a custom instance path for this test
			customInstance := escaped
			if tc.tokenFile != "auth_tokens.json" {
				// For non-default token files, we need to modify the instance path
				// to point to a directory with the specific token file
				customInstance = escaped + "_" + tc.tokenFile
				customInstanceDir := filepath.Join(testCacheDir, customInstance)
				err := os.MkdirAll(customInstanceDir, 0700)
				require.NoError(t, err, "Failed to create custom instance directory")

				// Copy the test token file to the standard auth_tokens.json name in the custom instance directory
				srcPath := filepath.Join(testInstanceDir, tc.tokenFile)
				destPath := filepath.Join(customInstanceDir, "auth_tokens.json")

				if tc.tokenFile != "nonexistent.json" {
					// Only try to copy if the source file exists
					err = os.WriteFile(destPath, []byte{}, 0644) // Create empty file first
					require.NoError(t, err, "Failed to create empty destination file")

					srcData, err := os.ReadFile(srcPath)
					require.NoError(t, err, "Failed to read source token file")

					err = os.WriteFile(destPath, srcData, 0644)
					require.NoError(t, err, "Failed to write to destination token file")
				}

				// Setup cleanup for this custom instance
				t.Cleanup(func() {
					if err := os.RemoveAll(customInstanceDir); err != nil {
						t.Logf("Warning: Failed to clean up custom instance directory: %v", err)
					}
				})
			}

			// Call the function being tested
			account, err := GetAccountName(testCacheDir, customInstance)

			// Check if the result matches expectations
			if tc.expectedError {
				require.Error(t, err, "Expected an error but got none")
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains,
						"Error message did not contain expected text")
				}
			} else {
				require.NoError(t, err, "Got unexpected error")
				if tc.name == "EmptyAccount" {
					assert.Empty(t, account, "Expected empty account but got: %s", account)
				} else {
					assert.NotEmpty(t, account, "Expected non-empty account but got empty string")
				}
			}
		})
	}
}
