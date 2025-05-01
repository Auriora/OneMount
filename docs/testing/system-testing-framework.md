# System Testing Framework

This document provides an overview of the system testing framework for the OneMount project, including how to set up a production-like environment, define end-to-end test scenarios, verify system behavior, test with production-like data volumes, and test system configuration options.

## Overview

The system testing framework is designed to test the complete integrated system with minimal mocking, using production-like data volumes and configuration options. It builds on the existing test utilities in the project, including the TestFramework, scenario-based testing, and network simulation.

## Key Components

### SystemTestEnvironment

The `SystemTestEnvironment` provides a controlled environment for system testing. It includes:

- System configuration management
- Test framework for test execution
- System verifier for verifying system behavior
- Data volume generator for generating test data
- Configuration manager for managing system configuration
- Test scenario management

### SystemVerifier

The `SystemVerifier` provides utilities for verifying system behavior, including:

- Verifying file content
- Verifying file existence
- Verifying directory existence and contents
- Verifying system metrics
- Verifying system state

### DataVolumeGenerator

The `DataVolumeGenerator` provides utilities for generating production-like data volumes, including:

- Generating sets of files with specified total size
- Generating nested directory structures
- Generating large files
- Cleaning up generated data

### ConfigurationManager

The `ConfigurationManager` provides utilities for managing system configuration options, including:

- Setting and getting configuration options
- Saving and loading configuration from files
- Resetting configuration to default values

### CommonSystemScenarios

The `CommonSystemScenarios` provides predefined test scenarios for common system testing tasks, including:

- End-to-end file operations
- Configuration option testing
- Large data volume testing
- System behavior verification

## Usage

### Setting Up a System Test Environment

```go
import (
    "context"
    "github.com/rs/zerolog/log"
    "github.com/yourusername/onemount/internal/testutil"
)

// Create a logger
logger := log.With().Str("component", "system-test").Logger()

// Create a context
ctx := context.Background()

// Create a new SystemTestEnvironment
env := testutil.NewSystemTestEnvironment(ctx, &logger)

// Set up the environment
err := env.SetupEnvironment()
if err != nil {
    // Handle error
}

// Add cleanup using t.Cleanup to ensure resources are cleaned up
t.Cleanup(func() {
    env.TeardownEnvironment()
})
```

### Defining End-to-End Test Scenarios

```go
// Create common system scenarios
scenarios := testutil.NewCommonSystemScenarios(env)

// Add a predefined scenario
env.AddScenario(scenarios.EndToEndFileOperationsScenario())

// Define a custom scenario
customScenario := testutil.TestScenario{
    Name:        "Custom System Test",
    Description: "Tests custom system functionality",
    Tags:        []string{"system", "custom"},
    Steps: []testutil.TestStep{
        {
            Name: "Custom step",
            Action: func(ctx context.Context) error {
                // Custom step implementation
                return nil
            },
            Validation: func(ctx context.Context) error {
                // Custom validation
                return nil
            },
        },
        // More steps...
    },
    Assertions: []testutil.TestAssertion{
        {
            Name: "Custom assertion",
            Condition: func(ctx context.Context) bool {
                // Custom assertion condition
                return true
            },
            Message: "Custom assertion failed",
        },
    },
    Cleanup: []testutil.CleanupStep{
        {
            Name: "Custom cleanup",
            Action: func(ctx context.Context) error {
                // Custom cleanup action
                return nil
            },
            AlwaysRun: true,
        },
    },
}

// Add the custom scenario
env.AddScenario(customScenario)
```

### Running Test Scenarios

```go
// Run a specific scenario
err := env.RunScenario("End-to-End File Operations")
if err != nil {
    // Handle error
}

// Run all scenarios
errors := env.RunAllScenarios()
if len(errors) > 0 {
    // Handle errors
}
```

### Verifying System Behavior

```go
// Get the system verifier
verifier := env.GetVerifier()

// Verify file existence
err := verifier.VerifyFileExists("test-file.txt")
if err != nil {
    // Handle error
}

// Verify file content
expectedContent := []byte("Test content")
err = verifier.VerifyFileContent("test-file.txt", expectedContent)
if err != nil {
    // Handle error
}

// Verify directory contents
expectedEntries := []string{"file1.txt", "file2.txt", "subdir"}
err = verifier.VerifyDirectoryContents("test-dir", expectedEntries)
if err != nil {
    // Handle error
}

// Verify system metrics
metrics := map[string]float64{
    "throughput":   100.0, // MB/s
    "latency":      50.0,  // ms
    "cpu_usage":    30.0,  // %
    "memory_usage": 200.0, // MB
}
err = verifier.VerifySystemMetrics(metrics)
if err != nil {
    // Handle error
}

// Verify system state
expectedState := map[string]interface{}{
    "status":        "running",
    "connected":     true,
    "cache_entries": 100,
    "error_count":   0,
}
err = verifier.VerifySystemState(expectedState)
if err != nil {
    // Handle error
}
```

### Testing with Production-Like Data Volumes

```go
// Get the data volume generator
dataGenerator := env.GetDataGenerator()

// Generate a set of files with specified total size
dataDir := filepath.Join(env.GetConfig().BaseDir, "test-data")
err := dataGenerator.GenerateFiles(dataDir, 500) // 500MB of data
if err != nil {
    // Handle error
}

// Generate a nested directory structure
err = dataGenerator.GenerateNestedDirectories(dataDir, 3, 5) // 3 levels deep, 5 files per directory
if err != nil {
    // Handle error
}

// Generate a large file
largePath := filepath.Join(dataDir, "large-file.dat")
err = dataGenerator.GenerateLargeFile(largePath, 100) // 100MB file
if err != nil {
    // Handle error
}

// Clean up generated data
err = dataGenerator.CleanupGeneratedData(dataDir)
if err != nil {
    // Handle error
}
```

### Testing System Configuration Options

```go
// Get the configuration manager
configManager := env.GetConfigManager()

// Set a configuration option
err := configManager.SetConfig("cache_size_mb", 2048)
if err != nil {
    // Handle error
}

// Get a configuration option
value, err := configManager.GetConfig("cache_size_mb")
if err != nil {
    // Handle error
}
cacheSize := value.(int)

// Save configuration to a file
configPath := filepath.Join(env.GetConfig().BaseDir, "config.json")
err = configManager.SaveConfig(configPath)
if err != nil {
    // Handle error
}

// Load configuration from a file
err = configManager.LoadConfig(configPath)
if err != nil {
    // Handle error
}

// Reset configuration to default values
err = configManager.ResetConfig()
if err != nil {
    // Handle error
}
```

## Example: Complete System Test

```go
package system_test

import (
    "context"
    "testing"
    "path/filepath"

    "github.com/rs/zerolog/log"
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestEndToEndFileOperations(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "system-test").Logger()

    // Create a context
    ctx := context.Background()

    // Create a new SystemTestEnvironment
    env := testutil.NewSystemTestEnvironment(ctx, &logger)
    require.NotNil(t, env)

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Create common system scenarios
    scenarios := testutil.NewCommonSystemScenarios(env)

    // Add the end-to-end file operations scenario
    env.AddScenario(scenarios.EndToEndFileOperationsScenario())

    // Run the scenario
    err = env.RunScenario("End-to-End File Operations")
    require.NoError(t, err)
}

func TestConfigurationOptions(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "system-test").Logger()

    // Create a context
    ctx := context.Background()

    // Create a new SystemTestEnvironment
    env := testutil.NewSystemTestEnvironment(ctx, &logger)
    require.NotNil(t, env)

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Create common system scenarios
    scenarios := testutil.NewCommonSystemScenarios(env)

    // Add the configuration test scenario
    env.AddScenario(scenarios.ConfigurationTestScenario())

    // Run the scenario
    err = env.RunScenario("Configuration Options Test")
    require.NoError(t, err)
}

func TestLargeDataVolumes(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "system-test").Logger()

    // Create a context
    ctx := context.Background()

    // Create a new SystemTestEnvironment with production-like data volumes
    env := testutil.NewSystemTestEnvironment(ctx, &logger)
    config := env.GetConfig()
    config.ProductionDataVolumes = true
    config.DataSizeMB = 500 // 500MB of test data
    env.SetConfig(config)
    require.NotNil(t, env)

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Create common system scenarios
    scenarios := testutil.NewCommonSystemScenarios(env)

    // Add the large data volume scenario
    env.AddScenario(scenarios.LargeDataVolumeScenario())

    // Run the scenario
    err = env.RunScenario("Large Data Volume Test")
    require.NoError(t, err)
}

func TestSystemBehavior(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "system-test").Logger()

    // Create a context
    ctx := context.Background()

    // Create a new SystemTestEnvironment
    env := testutil.NewSystemTestEnvironment(ctx, &logger)
    require.NotNil(t, env)

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Create common system scenarios
    scenarios := testutil.NewCommonSystemScenarios(env)

    // Add the system behavior verification scenario
    env.AddScenario(scenarios.SystemBehaviorVerificationScenario())

    // Run the scenario
    err = env.RunScenario("System Behavior Verification")
    require.NoError(t, err)
}
```

## Best Practices

1. Always use `t.Cleanup()` to ensure the environment is torn down, even if tests panic
2. Test with production-like data volumes to catch performance issues
3. Test with different configuration options to ensure system flexibility
4. Use scenario-based testing for complex system tests
5. Include validation steps to verify the results of actions
6. Use cleanup steps to ensure the environment is left in a clean state
7. Test both normal operation and error handling
8. Use descriptive names for scenarios, steps, and assertions
9. Verify system behavior using the provided verifier utilities
10. Test with different network conditions to ensure robustness