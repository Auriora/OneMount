package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/coreos/go-systemd/v22/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMountpointIsValid tests that we can detect a mountpoint as valid appropriately
func TestMountpointIsValid(t *testing.T) {
	// Create a test directory and file
	err := os.Mkdir("_test", 0755)
	require.NoError(t, err, "Failed to create test directory")

	err = os.WriteFile("_test/.example", []byte("some text\n"), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Setup cleanup to remove the test directory after test completes or fails
	t.Cleanup(func() {
		if err := os.RemoveAll("_test"); err != nil {
			t.Logf("Warning: Failed to clean up test directory: %v", err)
		}
	})

	// Check if the "mount" directory is empty
	dirents, err := os.ReadDir("mount")
	if err == nil {
		t.Logf("Mount directory contents (%d items):", len(dirents))
		for _, dirent := range dirents {
			t.Logf("  %s", dirent.Name())
		}
	} else {
		t.Logf("Error reading mount directory: %v", err)
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
			mountpoint: "mount",
			expected:   true,
			reason:     "mount directory should be a valid mountpoint",
		},
		{
			name:       "TestDirectory",
			mountpoint: "_test",
			expected:   false,
			reason:     "Test directory should not be a valid mountpoint",
		},
		{
			name:       "TestFile",
			mountpoint: "_test/.example",
			expected:   false,
			reason:     "File should not be a valid mountpoint",
		},
	}

	// Run each test case as a subtest
	for _, tc := range tests {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
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
			// Test escaping home directory
			t.Run("Escape", func(t *testing.T) {
				escaped := EscapeHome(tc.unescaped)
				require.Equal(t, tc.escaped, escaped,
					"Failed to correctly escape home in %q (%s)", tc.unescaped, tc.desc)
			})

			// Test unescaping home directory
			t.Run("Unescape", func(t *testing.T) {
				unescaped := UnescapeHome(tc.escaped)
				require.Equal(t, tc.unescaped, unescaped,
					"Failed to correctly unescape home in %q (%s)", tc.escaped, tc.desc)
			})
		})
	}
}

func TestGetAccountName(t *testing.T) {
	t.Parallel()

	wd, _ := os.Getwd()
	escaped := unit.UnitNamePathEscape(filepath.Join(wd, "mount"))

	// we compute the cache directory manually to avoid an import cycle
	cacheDir, _ := os.UserCacheDir()

	// copy auth tokens to cache dir if it doesn't already exist
	// (CI runners will not have this file yet)
	os.MkdirAll(filepath.Join(cacheDir, "onedriver", escaped), 0700)
	dest := filepath.Join(cacheDir, "onedriver", escaped, "auth_tokens.json")
	if _, err := os.Stat(dest); err != nil {
		exec.Command("cp", ".auth_tokens.json", dest).Run()
	}

	_, err := GetAccountName(filepath.Join(cacheDir, "onedriver"), escaped)
	assert.NoError(t, err)
}
