package common

import (
	"github.com/bcherrington/onedriver/internal/testutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const configTestDir = "configs/resources/test"

// We should load config correctly.
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	conf := LoadConfig(filepath.Join(configTestDir, "config-test.yml"))

	home, err := os.UserHomeDir()
	require.NoError(t, err, "Failed to get user home directory")
	assert.Equal(t, filepath.Join(home, "somewhere/else"), conf.CacheDir)
	assert.Equal(t, "warn", conf.LogLevel)
}

func TestConfigMerge(t *testing.T) {
	t.Parallel()

	conf := LoadConfig(filepath.Join(configTestDir, "config-test-merge.yml"))

	assert.Equal(t, "debug", conf.LogLevel)
	assert.Equal(t, "/some/directory", conf.CacheDir)
}

// We should come up with the defaults if there is no config file.
func TestLoadNonexistentConfig(t *testing.T) {
	t.Parallel()

	conf := LoadConfig(filepath.Join(configTestDir, "does-not-exist.yml"))

	home, err := os.UserHomeDir()
	require.NoError(t, err, "Failed to get user home directory")
	assert.Equal(t, filepath.Join(home, ".cache/onedriver"), conf.CacheDir)
	assert.Equal(t, "debug", conf.LogLevel)
}

func TestWriteConfig(t *testing.T) {
	t.Parallel()

	configPath := testutil.TestSandboxTmpDir + "/nested/config.yml"

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.RemoveAll(testutil.TestSandboxTmpDir + "/nested"); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test directory: %v", err)
		}
	})

	conf := LoadConfig(filepath.Join(configTestDir, "config-test.yml"))
	require.NoError(t, conf.WriteConfig(configPath), "Failed to write config file")
}
