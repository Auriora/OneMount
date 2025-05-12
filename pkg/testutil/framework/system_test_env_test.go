package framework

import (
	"context"
	"github.com/auriora/onemount/pkg/testutil"
	"os"
	"os/signal"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Use the existing logger implementation in the testutil package

func TestSystemTestEnvironment_Setup(t *testing.T) {
	// Create a logger
	logger := testutil.NewCustomLogger("system-test")

	// Create a context
	ctx := context.Background()

	// Create a new SystemTestEnvironment
	env := NewSystemTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Ensure signal handling is not active
	if env.framework.isHandling {
		t.Log("Signal handling is active at the start of the test, cleaning up")
		// Create a cleanup function that stops signal handling
		cleanup := func() {
			env.framework.signalMu.Lock()
			defer env.framework.signalMu.Unlock()

			if !env.framework.isHandling {
				return
			}

			// Stop receiving signals
			signal.Stop(env.framework.signalChan)

			// Cancel the context to stop the goroutine
			if env.framework.signalCancel != nil {
				env.framework.signalCancel()
			}

			// Only close the channel if it's not nil
			if env.framework.signalChan != nil {
				close(env.framework.signalChan)
				env.framework.signalChan = nil
			}

			// Set the isHandling flag to false
			env.framework.isHandling = false
		}

		// Call the cleanup function
		cleanup()

		// Verify that signal handling is stopped
		assert.False(t, env.framework.isHandling, "Signal handling should be inactive after cleanup")
	}

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)

	// Verify that the base directory was created
	_, err = os.Stat(env.config.BaseDir)
	require.NoError(t, err)

	// Verify that the mount point was created
	_, err = os.Stat(env.config.MountPoint)
	require.NoError(t, err)

	// Clean up
	defer func() {
		err := env.TeardownEnvironment()
		require.NoError(t, err)

		// Verify that the base directory was removed
		_, err = os.Stat(env.config.BaseDir)
		assert.True(t, os.IsNotExist(err))
	}()
}

func TestSystemTestEnvironment_DataGenerator(t *testing.T) {
	// Create a logger
	logger := testutil.NewCustomLogger("system-test")

	// Create a context
	ctx := context.Background()

	// Create a new SystemTestEnvironment
	env := NewSystemTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)

	// Clean up
	defer func() {
		err := env.TeardownEnvironment()
		require.NoError(t, err)
	}()

	// Get the data generator
	dataGenerator := env.GetDataGenerator()
	require.NotNil(t, dataGenerator)

	// Generate a test file
	testDir := filepath.Join(env.config.BaseDir, "test-data")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	testFile := filepath.Join(testDir, "test-file.dat")
	err = dataGenerator.GenerateLargeFile(testFile, 1) // 1MB file
	require.NoError(t, err)

	// Verify that the file was created
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, int64(1024*1024), info.Size())

	// Clean up the test file
	err = dataGenerator.CleanupGeneratedData(testDir)
	require.NoError(t, err)

	// Verify that the directory was removed
	_, err = os.Stat(testDir)
	assert.True(t, os.IsNotExist(err))
}

func TestSystemTestEnvironment_ConfigManager(t *testing.T) {
	// Create a logger
	logger := testutil.NewCustomLogger("system-test")

	// Create a context
	ctx := context.Background()

	// Create a new SystemTestEnvironment
	env := NewSystemTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)

	// Clean up
	defer func() {
		err := env.TeardownEnvironment()
		require.NoError(t, err)
	}()

	// Get the configuration manager
	configManager := env.GetConfigManager()
	require.NotNil(t, configManager)

	// Set a configuration option
	err = configManager.SetConfig("cache_size_mb", 2048)
	require.NoError(t, err)

	// Get the configuration option
	value, err := configManager.GetConfig("cache_size_mb")
	require.NoError(t, err)
	assert.Equal(t, 2048, value)

	// Reset the configuration
	err = configManager.ResetConfig()
	require.NoError(t, err)

	// Verify that the configuration was reset
	value, err = configManager.GetConfig("cache_size_mb")
	require.NoError(t, err)
	assert.Equal(t, 1024, value)
}

func TestSystemTestEnvironment_Verifier(t *testing.T) {
	// Create a logger
	logger := testutil.NewCustomLogger("system-test")

	// Create a context
	ctx := context.Background()

	// Create a new SystemTestEnvironment
	env := NewSystemTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)

	// Clean up
	defer func() {
		err := env.TeardownEnvironment()
		require.NoError(t, err)
	}()

	// Get the verifier
	verifier := env.GetVerifier()
	require.NotNil(t, verifier)

	// Create a test file
	testFile := filepath.Join(env.config.MountPoint, "test-file.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Verify that the file exists
	err = verifier.VerifyFileExists("test-file.txt")
	require.NoError(t, err)

	// Verify the file content
	err = verifier.VerifyFileContent("test-file.txt", []byte("test content"))
	require.NoError(t, err)

	// Create a test directory
	testDir := filepath.Join(env.config.MountPoint, "test-dir")
	err = os.MkdirAll(testDir, 0755)
	require.NoError(t, err)

	// Create some files in the test directory
	err = os.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("file1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(testDir, "file2.txt"), []byte("file2"), 0644)
	require.NoError(t, err)

	// Verify that the directory exists
	err = verifier.VerifyDirectoryExists("test-dir")
	require.NoError(t, err)

	// Verify the directory contents
	err = verifier.VerifyDirectoryContents("test-dir", []string{"file1.txt", "file2.txt"})
	require.NoError(t, err)
}

func TestSystemTestEnvironment_Scenarios(t *testing.T) {
	// Create a logger
	logger := testutil.NewCustomLogger("system-test")

	// Create a context
	ctx := context.Background()

	// Create a new SystemTestEnvironment
	env := NewSystemTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)

	// Clean up
	defer func() {
		err := env.TeardownEnvironment()
		require.NoError(t, err)
	}()

	// Create a simple test scenario
	scenario := TestScenario{
		Name:        "Simple Test Scenario",
		Description: "A simple test scenario for testing the SystemTestEnvironment",
		Tags:        []string{"system", "test"},
		Steps: []TestStep{
			{
				Name: "Create a test file",
				Action: func(ctx context.Context) error {
					testFile := filepath.Join(env.config.MountPoint, "scenario-test-file.txt")
					return os.WriteFile(testFile, []byte("scenario test content"), 0644)
				},
				Validation: func(ctx context.Context) error {
					return env.verifier.VerifyFileExists("scenario-test-file.txt")
				},
			},
			{
				Name: "Verify file content",
				Action: func(ctx context.Context) error {
					return env.verifier.VerifyFileContent("scenario-test-file.txt", []byte("scenario test content"))
				},
			},
		},
		Assertions: []TestAssertion{
			{
				Name: "File operations were successful",
				Condition: func(ctx context.Context) bool {
					err := env.verifier.VerifyFileExists("scenario-test-file.txt")
					return err == nil
				},
				Message: "File operations failed",
			},
		},
		Cleanup: []CleanupStep{
			{
				Name: "Remove test file",
				Action: func(ctx context.Context) error {
					return os.Remove(filepath.Join(env.config.MountPoint, "scenario-test-file.txt"))
				},
				AlwaysRun: true,
			},
		},
	}

	// Add the scenario
	env.AddScenario(scenario)

	// Run the scenario
	err = env.RunScenario("Simple Test Scenario")
	require.NoError(t, err)

	// Verify that the file was removed by the cleanup step
	_, err = os.Stat(filepath.Join(env.config.MountPoint, "scenario-test-file.txt"))
	assert.True(t, os.IsNotExist(err))
}

func TestCommonSystemScenarios(t *testing.T) {
	// Skip this test in automated test runs as it requires more setup
	if os.Getenv("CI") != "" {
		t.Skip("Skipping test in CI environment")
	}

	// Create a logger
	logger := testutil.NewCustomLogger("system-test")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a new SystemTestEnvironment
	env := NewSystemTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)

	// Clean up
	defer func() {
		err := env.TeardownEnvironment()
		require.NoError(t, err)
	}()

	// Create common system scenarios
	scenarios := NewCommonSystemScenarios(env)
	require.NotNil(t, scenarios)

	// Add the end-to-end file operations scenario
	env.AddScenario(scenarios.EndToEndFileOperationsScenario())

	// Add the configuration test scenario
	env.AddScenario(scenarios.ConfigurationTestScenario())

	// Add the large data volume scenario
	env.AddScenario(scenarios.LargeDataVolumeScenario())

	// Add the system behavior verification scenario
	env.AddScenario(scenarios.SystemBehaviorVerificationScenario())

	// Run all scenarios
	errors := env.RunAllScenarios()
	assert.Empty(t, errors)
}

// TestSystemTestEnvironment_SignalHandling tests the interaction between SystemTestEnvironment and signal handling
func TestSystemTestEnvironment_SignalHandling(t *testing.T) {
	// Create a logger
	logger := testutil.NewCustomLogger("system-test")

	// Create a context
	ctx := context.Background()

	// Create a new SystemTestEnvironment
	env := NewSystemTestEnvironment(ctx, logger)
	require.NotNil(t, env)

	// Explicitly set up signal handling on the framework
	cleanup := env.framework.SetupSignalHandling()
	require.NotNil(t, cleanup)

	// Verify that signal handling is set up
	assert.True(t, env.framework.isHandling, "Signal handling should be active")
	assert.NotNil(t, env.framework.signalChan, "Signal channel should not be nil")

	// Set up the environment
	err := env.SetupEnvironment()
	require.NoError(t, err)

	// Clean up signal handling
	cleanup()

	// Verify that signal handling is stopped
	assert.False(t, env.framework.isHandling, "Signal handling should be inactive after cleanup")
	assert.Nil(t, env.framework.signalChan, "Signal channel should be nil after cleanup")

	// Clean up the environment
	defer func() {
		err := env.TeardownEnvironment()
		require.NoError(t, err)
	}()
}
