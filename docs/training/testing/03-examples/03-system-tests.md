# System Test Examples

This file contains examples of system tests using the OneMount test framework. System tests verify the behavior of the entire system as a whole, including all its components and their interactions.

> **Note**: All code examples in this document are for illustration purposes only and may need to be adapted to your specific project structure and imports. The examples are not meant to be compiled directly but rather to demonstrate concepts and patterns.

## Table of Contents

1. [Introduction to System Testing](#introduction-to-system-testing)
2. [Basic System Test](#basic-system-test)
3. [End-to-End Workflow Tests](#end-to-end-workflow-tests)
4. [System Configuration Tests](#system-configuration-tests)
5. [System Performance Tests](#system-performance-tests)
6. [System Recovery Tests](#system-recovery-tests)

## Introduction to System Testing

System testing is a critical part of the testing process that verifies the behavior of the entire system as a whole. Unlike unit tests and integration tests, which focus on individual components and their interactions, system tests verify that the entire system works correctly from end to end.

System tests help you:

- Verify that the system meets its requirements
- Identify issues that only occur when all components are working together
- Ensure that the system works correctly in a production-like environment
- Validate end-to-end workflows
- Test system configuration options
- Verify system performance under realistic conditions

The OneMount test framework provides tools for writing and running system tests, including the SystemTestEnvironment component.

## Basic System Test

Here's a basic example of a system test that verifies the end-to-end functionality of the OneMount filesystem:

```go
package system_test

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestBasicSystemFunctionality tests the basic functionality of the OneMount filesystem
func TestBasicSystemFunctionality(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a system test environment
    env := testutil.NewSystemTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the mount point
    mountPoint := env.GetMountPoint()
    require.NotEmpty(t, mountPoint, "Mount point should not be empty")
    
    // Verify that the mount point exists and is a directory
    info, err := os.Stat(mountPoint)
    require.NoError(t, err, "Mount point should exist")
    require.True(t, info.IsDir(), "Mount point should be a directory")
    
    // Create a test file
    testFilePath := filepath.Join(mountPoint, "test.txt")
    testContent := "Hello, OneMount!"
    err = os.WriteFile(testFilePath, []byte(testContent), 0644)
    require.NoError(t, err, "Failed to create test file")
    
    // Wait for the file to be uploaded
    time.Sleep(2 * time.Second)
    
    // Read the file back
    content, err := os.ReadFile(testFilePath)
    require.NoError(t, err, "Failed to read test file")
    require.Equal(t, testContent, string(content), "File content should match")
    
    // List the directory
    files, err := os.ReadDir(mountPoint)
    require.NoError(t, err, "Failed to list directory")
    require.GreaterOrEqual(t, len(files), 1, "Directory should contain at least one file")
    
    // Find the test file in the directory listing
    found := false
    for _, file := range files {
        if file.Name() == "test.txt" {
            found = true
            break
        }
    }
    require.True(t, found, "Test file should be in the directory listing")
    
    // Delete the test file
    err = os.Remove(testFilePath)
    require.NoError(t, err, "Failed to delete test file")
    
    // Wait for the file to be deleted
    time.Sleep(2 * time.Second)
    
    // Verify that the file no longer exists
    _, err = os.Stat(testFilePath)
    require.True(t, os.IsNotExist(err), "File should no longer exist")
}
```

This test:
1. Sets up a system test environment
2. Verifies that the mount point exists and is a directory
3. Creates a test file
4. Reads the file back
5. Lists the directory and verifies that the file is in the listing
6. Deletes the file
7. Verifies that the file no longer exists

## End-to-End Workflow Tests

Here's an example of a system test that verifies an end-to-end workflow:

```go
package system_test

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestDocumentWorkflow tests a complete document workflow
func TestDocumentWorkflow(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a system test environment
    env := testutil.NewSystemTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the mount point
    mountPoint := env.GetMountPoint()
    require.NotEmpty(t, mountPoint, "Mount point should not be empty")
    
    // Create a test scenario
    scenario := testutil.TestScenario{
        Name:        "Document Workflow",
        Description: "Tests a complete document workflow",
        Steps: []testutil.TestStep{
            {
                Name: "Create document directory",
                Action: func(ctx context.Context) error {
                    return os.Mkdir(filepath.Join(mountPoint, "Documents"), 0755)
                },
                Validation: func(ctx context.Context) error {
                    info, err := os.Stat(filepath.Join(mountPoint, "Documents"))
                    if err != nil {
                        return err
                    }
                    if !info.IsDir() {
                        return fmt.Errorf("Documents should be a directory")
                    }
                    return nil
                },
            },
            {
                Name: "Create document",
                Action: func(ctx context.Context) error {
                    return os.WriteFile(
                        filepath.Join(mountPoint, "Documents", "report.txt"),
                        []byte("# Quarterly Report\n\nThis is a quarterly report."),
                        0644,
                    )
                },
                Validation: func(ctx context.Context) error {
                    content, err := os.ReadFile(filepath.Join(mountPoint, "Documents", "report.txt"))
                    if err != nil {
                        return err
                    }
                    if string(content) != "# Quarterly Report\n\nThis is a quarterly report." {
                        return fmt.Errorf("document content does not match")
                    }
                    return nil
                },
            },
            {
                Name: "Create backup directory",
                Action: func(ctx context.Context) error {
                    return os.Mkdir(filepath.Join(mountPoint, "Backups"), 0755)
                },
                Validation: func(ctx context.Context) error {
                    info, err := os.Stat(filepath.Join(mountPoint, "Backups"))
                    if err != nil {
                        return err
                    }
                    if !info.IsDir() {
                        return fmt.Errorf("Backups should be a directory")
                    }
                    return nil
                },
            },
            {
                Name: "Copy document to backup",
                Action: func(ctx context.Context) error {
                    content, err := os.ReadFile(filepath.Join(mountPoint, "Documents", "report.txt"))
                    if err != nil {
                        return err
                    }
                    return os.WriteFile(
                        filepath.Join(mountPoint, "Backups", "report-backup.txt"),
                        content,
                        0644,
                    )
                },
                Validation: func(ctx context.Context) error {
                    content, err := os.ReadFile(filepath.Join(mountPoint, "Backups", "report-backup.txt"))
                    if err != nil {
                        return err
                    }
                    if string(content) != "# Quarterly Report\n\nThis is a quarterly report." {
                        return fmt.Errorf("backup content does not match")
                    }
                    return nil
                },
            },
            {
                Name: "Update document",
                Action: func(ctx context.Context) error {
                    return os.WriteFile(
                        filepath.Join(mountPoint, "Documents", "report.txt"),
                        []byte("# Quarterly Report\n\nThis is an updated quarterly report."),
                        0644,
                    )
                },
                Validation: func(ctx context.Context) error {
                    content, err := os.ReadFile(filepath.Join(mountPoint, "Documents", "report.txt"))
                    if err != nil {
                        return err
                    }
                    if string(content) != "# Quarterly Report\n\nThis is an updated quarterly report." {
                        return fmt.Errorf("updated document content does not match")
                    }
                    return nil
                },
            },
            {
                Name: "Verify backup is unchanged",
                Action: func(ctx context.Context) error {
                    content, err := os.ReadFile(filepath.Join(mountPoint, "Backups", "report-backup.txt"))
                    if err != nil {
                        return err
                    }
                    if string(content) != "# Quarterly Report\n\nThis is a quarterly report." {
                        return fmt.Errorf("backup should be unchanged")
                    }
                    return nil
                },
            },
        },
        Cleanup: []testutil.CleanupStep{
            {
                Name: "Clean up test files",
                Action: func(ctx context.Context) error {
                    // Remove the test files and directories
                    os.Remove(filepath.Join(mountPoint, "Documents", "report.txt"))
                    os.Remove(filepath.Join(mountPoint, "Backups", "report-backup.txt"))
                    os.Remove(filepath.Join(mountPoint, "Documents"))
                    os.Remove(filepath.Join(mountPoint, "Backups"))
                    return nil
                },
                AlwaysRun: true,
            },
        },
    }
    
    // Run the scenario
    err = env.RunScenario(scenario)
    require.NoError(t, err, "Scenario should complete successfully")
}
```

This test:
1. Sets up a system test environment
2. Creates a test scenario with multiple steps
3. Creates a document directory
4. Creates a document
5. Creates a backup directory
6. Copies the document to the backup directory
7. Updates the document
8. Verifies that the backup is unchanged
9. Cleans up the test files and directories

## System Configuration Tests

Here's an example of a system test that verifies different system configurations:

```go
package system_test

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestSystemConfigurations tests different system configurations
func TestSystemConfigurations(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Define different configurations to test
    configurations := []struct {
        name        string
        config      testutil.SystemConfig
        setupFunc   func(t *testing.T, env *testutil.SystemTestEnvironment)
        validateFunc func(t *testing.T, env *testutil.SystemTestEnvironment)
    }{
        {
            name: "DefaultConfig",
            config: testutil.SystemConfig{
                CacheEnabled: true,
                CacheSize:    1024 * 1024 * 100, // 100 MB
                LogLevel:     "info",
            },
            setupFunc: func(t *testing.T, env *testutil.SystemTestEnvironment) {
                // No additional setup needed for default config
            },
            validateFunc: func(t *testing.T, env *testutil.SystemTestEnvironment) {
                // Verify that the cache is enabled
                cacheDir := env.GetCacheDirectory()
                require.NotEmpty(t, cacheDir, "Cache directory should not be empty")
                
                // Verify that the cache directory exists
                info, err := os.Stat(cacheDir)
                require.NoError(t, err, "Cache directory should exist")
                require.True(t, info.IsDir(), "Cache directory should be a directory")
            },
        },
        {
            name: "NoCacheConfig",
            config: testutil.SystemConfig{
                CacheEnabled: false,
                LogLevel:     "info",
            },
            setupFunc: func(t *testing.T, env *testutil.SystemTestEnvironment) {
                // No additional setup needed for no-cache config
            },
            validateFunc: func(t *testing.T, env *testutil.SystemTestEnvironment) {
                // Verify that the cache is disabled
                cacheDir := env.GetCacheDirectory()
                require.Empty(t, cacheDir, "Cache directory should be empty when cache is disabled")
                
                // Create a test file
                mountPoint := env.GetMountPoint()
                testFilePath := filepath.Join(mountPoint, "test.txt")
                testContent := "Hello, OneMount!"
                err := os.WriteFile(testFilePath, []byte(testContent), 0644)
                require.NoError(t, err, "Failed to create test file")
                
                // Wait for the file to be uploaded
                time.Sleep(2 * time.Second)
                
                // Read the file back
                content, err := os.ReadFile(testFilePath)
                require.NoError(t, err, "Failed to read test file")
                require.Equal(t, testContent, string(content), "File content should match")
                
                // Delete the test file
                err = os.Remove(testFilePath)
                require.NoError(t, err, "Failed to delete test file")
            },
        },
        {
            name: "DebugLogConfig",
            config: testutil.SystemConfig{
                CacheEnabled: true,
                CacheSize:    1024 * 1024 * 100, // 100 MB
                LogLevel:     "debug",
            },
            setupFunc: func(t *testing.T, env *testutil.SystemTestEnvironment) {
                // No additional setup needed for debug log config
            },
            validateFunc: func(t *testing.T, env *testutil.SystemTestEnvironment) {
                // Verify that debug logs are being generated
                logFile := env.GetLogFile()
                require.NotEmpty(t, logFile, "Log file should not be empty")
                
                // Verify that the log file exists
                info, err := os.Stat(logFile)
                require.NoError(t, err, "Log file should exist")
                require.False(t, info.IsDir(), "Log file should not be a directory")
                
                // Read the log file
                content, err := os.ReadFile(logFile)
                require.NoError(t, err, "Failed to read log file")
                
                // Verify that the log file contains debug messages
                require.Contains(t, string(content), "level=debug", "Log file should contain debug messages")
            },
        },
    }
    
    // Test each configuration
    for _, cfg := range configurations {
        t.Run(cfg.name, func(t *testing.T) {
            // Create a system test environment with the configuration
            env := testutil.NewSystemTestEnvironment(context.Background(), logger, cfg.config)
            require.NotNil(t, env)
            
            // Set up the environment
            err := env.SetupEnvironment()
            require.NoError(t, err)
            
            // Add cleanup using t.Cleanup to ensure resources are cleaned up
            t.Cleanup(func() {
                env.TeardownEnvironment()
            })
            
            // Run the setup function
            cfg.setupFunc(t, env)
            
            // Run the validation function
            cfg.validateFunc(t, env)
        })
    }
}
```

This test:
1. Defines different system configurations to test
2. Tests each configuration in a separate subtest
3. Sets up a system test environment with the configuration
4. Validates that the system behaves correctly with the configuration

## System Performance Tests

Here's an example of a system test that verifies system performance:

```go
package system_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestSystemPerformance tests the performance of the system
func TestSystemPerformance(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a system test environment
    env := testutil.NewSystemTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the mount point
    mountPoint := env.GetMountPoint()
    require.NotEmpty(t, mountPoint, "Mount point should not be empty")
    
    // Create a performance benchmark
    benchmark := testutil.NewPerformanceBenchmark(testutil.BenchmarkConfig{
        Name:             "SystemPerformance",
        Iterations:       10,
        WarmupIterations: 2,
        Timeout:          5 * time.Minute,
        Thresholds: testutil.PerformanceThresholds{
            LatencyP50: 500 * time.Millisecond,
            LatencyP95: 1 * time.Second,
            LatencyP99: 2 * time.Second,
            Throughput: 10, // operations per second
        },
    }, logger)
    
    // Define operations to benchmark
    operations := map[string]func(ctx context.Context) error{
        "CreateSmallFile": func(ctx context.Context) error {
            // Create a small file (1 KB)
            data := make([]byte, 1024)
            for i := range data {
                data[i] = byte(i % 256)
            }
            fileName := filepath.Join(mountPoint, fmt.Sprintf("small-file-%d.txt", time.Now().UnixNano()))
            return os.WriteFile(fileName, data, 0644)
        },
        "CreateMediumFile": func(ctx context.Context) error {
            // Create a medium file (1 MB)
            data := make([]byte, 1024*1024)
            for i := range data {
                data[i] = byte(i % 256)
            }
            fileName := filepath.Join(mountPoint, fmt.Sprintf("medium-file-%d.txt", time.Now().UnixNano()))
            return os.WriteFile(fileName, data, 0644)
        },
        "ReadSmallFile": func(ctx context.Context) error {
            // Create a small file to read
            data := make([]byte, 1024)
            for i := range data {
                data[i] = byte(i % 256)
            }
            fileName := filepath.Join(mountPoint, "small-file-read.txt")
            err := os.WriteFile(fileName, data, 0644)
            if err != nil {
                return err
            }
            
            // Read the file
            _, err = os.ReadFile(fileName)
            return err
        },
        "ListDirectory": func(ctx context.Context) error {
            // List the directory
            _, err := os.ReadDir(mountPoint)
            return err
        },
    }
    
    // Run benchmarks for each operation
    for name, operation := range operations {
        t.Run(name, func(t *testing.T) {
            // Update the benchmark name
            benchmark.SetName(name)
            
            // Run the benchmark
            results, err := benchmark.Run(operation)
            require.NoError(t, err, "Benchmark should complete successfully")
            
            // Log the results
            t.Logf("Latency (P50): %v", results.LatencyP50)
            t.Logf("Latency (P95): %v", results.LatencyP95)
            t.Logf("Latency (P99): %v", results.LatencyP99)
            t.Logf("Throughput: %v ops/sec", results.Throughput)
            
            // Verify that the results meet the thresholds
            require.LessOrEqual(t, results.LatencyP50, benchmark.GetThresholds().LatencyP50, "P50 latency exceeds threshold")
            require.LessOrEqual(t, results.LatencyP95, benchmark.GetThresholds().LatencyP95, "P95 latency exceeds threshold")
            require.LessOrEqual(t, results.LatencyP99, benchmark.GetThresholds().LatencyP99, "P99 latency exceeds threshold")
            require.GreaterOrEqual(t, results.Throughput, benchmark.GetThresholds().Throughput, "Throughput below threshold")
        })
    }
    
    // Clean up test files
    files, err := os.ReadDir(mountPoint)
    require.NoError(t, err, "Failed to list directory")
    for _, file := range files {
        if file.IsDir() {
            continue
        }
        err := os.Remove(filepath.Join(mountPoint, file.Name()))
        require.NoError(t, err, "Failed to delete test file")
    }
}
```

This test:
1. Sets up a system test environment
2. Creates a performance benchmark
3. Defines operations to benchmark (create small file, create medium file, read small file, list directory)
4. Runs benchmarks for each operation
5. Verifies that the results meet the performance thresholds
6. Cleans up test files

## System Recovery Tests

Here's an example of a system test that verifies system recovery:

```go
package system_test

import (
    "context"
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "time"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestSystemRecovery tests the system's ability to recover from failures
func TestSystemRecovery(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a system test environment
    env := testutil.NewSystemTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the mount point
    mountPoint := env.GetMountPoint()
    require.NotEmpty(t, mountPoint, "Mount point should not be empty")
    
    // Create a test file
    testFilePath := filepath.Join(mountPoint, "recovery-test.txt")
    testContent := "This is a recovery test file."
    err = os.WriteFile(testFilePath, []byte(testContent), 0644)
    require.NoError(t, err, "Failed to create test file")
    
    // Wait for the file to be uploaded
    time.Sleep(2 * time.Second)
    
    // Verify that the file exists
    _, err = os.Stat(testFilePath)
    require.NoError(t, err, "File should exist")
    
    // Simulate a system crash by forcibly stopping the OneMount service
    cmd := exec.Command("sudo", "systemctl", "stop", "onemount@" + os.Getenv("USER") + ".service")
    err = cmd.Run()
    require.NoError(t, err, "Failed to stop OneMount service")
    
    // Wait for the service to stop
    time.Sleep(5 * time.Second)
    
    // Verify that the mount point is no longer accessible
    _, err = os.Stat(mountPoint)
    require.Error(t, err, "Mount point should not be accessible")
    
    // Restart the OneMount service
    cmd = exec.Command("sudo", "systemctl", "start", "onemount@" + os.Getenv("USER") + ".service")
    err = cmd.Run()
    require.NoError(t, err, "Failed to start OneMount service")
    
    // Wait for the service to start
    time.Sleep(10 * time.Second)
    
    // Verify that the mount point is accessible again
    _, err = os.Stat(mountPoint)
    require.NoError(t, err, "Mount point should be accessible")
    
    // Verify that the test file still exists
    content, err := os.ReadFile(testFilePath)
    require.NoError(t, err, "Failed to read test file")
    require.Equal(t, testContent, string(content), "File content should match")
    
    // Delete the test file
    err = os.Remove(testFilePath)
    require.NoError(t, err, "Failed to delete test file")
}
```

This test:
1. Sets up a system test environment
2. Creates a test file
3. Simulates a system crash by stopping the OneMount service
4. Verifies that the mount point is no longer accessible
5. Restarts the OneMount service
6. Verifies that the mount point is accessible again
7. Verifies that the test file still exists and has the correct content
8. Cleans up the test file

These examples demonstrate different aspects of system testing with the OneMount test framework. By following these patterns, you can write comprehensive system tests that verify your system works correctly as a whole.