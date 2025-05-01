# Integration Test Examples

This file contains examples of integration tests using the OneMount test framework. Integration tests focus on verifying that different components of the system work together correctly.

> **Note**: All code examples in this document are for illustration purposes only and may need to be adapted to your specific project structure and imports. The examples are not meant to be compiled directly but rather to demonstrate concepts and patterns.

## Table of Contents

1. [Basic Integration Test](#basic-integration-test)
2. [Testing Component Interactions](#testing-component-interactions)
3. [Testing with Network Conditions](#testing-with-network-conditions)
4. [Scenario-Based Integration Tests](#scenario-based-integration-tests)
5. [Testing Error Handling](#testing-error-handling)

## Basic Integration Test

Here's a basic example of an integration test that verifies the interaction between the Graph API client and the filesystem:

```go
package fs_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/fs/graph"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestFileDownloadIntegration tests the integration between the Graph API client and the filesystem
func TestFileDownloadIntegration(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up isolation config to mock the Graph API
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph"},
        DataIsolation:  true,
    })
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the graph component
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)
    
    // Configure the mock graph component
    mockGraph.AddMockItem("/drive/root:/test.txt", &graph.DriveItem{
        ID:   "item123",
        Name: "test.txt",
        Size: 1024,
        File: &graph.File{
            MimeType: "text/plain",
        },
    })
    mockGraph.AddMockContent("/drive/root:/test.txt", []byte("Hello, World!"))
    
    // Create a file manager with the mock graph client
    fileManager := fs.NewFileManager(mockGraph, logger)
    
    // Download the file
    file, err := fileManager.GetFile("/test.txt")
    require.NoError(t, err, "Failed to get file")
    require.NotNil(t, file, "File should not be nil")
    
    // Verify the file properties
    require.Equal(t, "test.txt", file.Name, "File name should match")
    require.Equal(t, int64(1024), file.Size, "File size should match")
    
    // Download the file content
    content, err := fileManager.GetFileContent(file)
    require.NoError(t, err, "Failed to get file content")
    require.Equal(t, "Hello, World!", string(content), "File content should match")
    
    // Verify that the graph client was called correctly
    recorder := mockGraph.GetRecorder()
    require.True(t, recorder.VerifyCall("GetItemPath", 1), "GetItemPath should be called once")
    require.True(t, recorder.VerifyCall("GetItemContent", 1), "GetItemContent should be called once")
}
```

This test:
1. Sets up an integration test environment
2. Configures a mock Graph API client
3. Creates a file manager with the mock client
4. Downloads a file and its content
5. Verifies that the file properties and content are correct
6. Verifies that the Graph API client was called correctly

## Testing Component Interactions

Here's an example of testing the interaction between multiple components:

```go
package fs_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/fs/graph"
    "github.com/yourusername/onemount/internal/fs/cache"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestFileSystemCacheIntegration tests the integration between the filesystem and cache
func TestFileSystemCacheIntegration(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up isolation config to mock the Graph API and use a real cache
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph"},
        DataIsolation:  true,
    })
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the graph component
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)
    
    // Configure the mock graph component
    mockGraph.AddMockItem("/drive/root:/test.txt", &graph.DriveItem{
        ID:   "item123",
        Name: "test.txt",
        Size: 1024,
        File: &graph.File{
            MimeType: "text/plain",
        },
    })
    mockGraph.AddMockContent("/drive/root:/test.txt", []byte("Hello, World!"))
    
    // Create a temporary directory for the cache
    tempDir := t.TempDir()
    
    // Create a cache
    cacheManager := cache.NewCacheManager(tempDir, logger)
    
    // Create a file manager with the mock graph client and real cache
    fileManager := fs.NewFileManager(mockGraph, logger)
    fileManager.SetCacheManager(cacheManager)
    
    // Download the file (should be cached)
    file, err := fileManager.GetFile("/test.txt")
    require.NoError(t, err, "Failed to get file")
    
    // Download the file content (should be cached)
    content, err := fileManager.GetFileContent(file)
    require.NoError(t, err, "Failed to get file content")
    require.Equal(t, "Hello, World!", string(content), "File content should match")
    
    // Verify that the graph client was called
    recorder := mockGraph.GetRecorder()
    require.True(t, recorder.VerifyCall("GetItemPath", 1), "GetItemPath should be called once")
    require.True(t, recorder.VerifyCall("GetItemContent", 1), "GetItemContent should be called once")
    
    // Reset the recorder
    recorder.Reset()
    
    // Get the file again (should be served from cache)
    file, err = fileManager.GetFile("/test.txt")
    require.NoError(t, err, "Failed to get file from cache")
    
    // Get the file content again (should be served from cache)
    content, err = fileManager.GetFileContent(file)
    require.NoError(t, err, "Failed to get file content from cache")
    require.Equal(t, "Hello, World!", string(content), "Cached content should match")
    
    // Verify that the graph client was not called again
    require.False(t, recorder.VerifyCall("GetItemPath", 1), "GetItemPath should not be called again")
    require.False(t, recorder.VerifyCall("GetItemContent", 1), "GetItemContent should not be called again")
    
    // Verify that the cache contains the file
    require.True(t, cacheManager.HasFile("item123"), "Cache should contain the file")
    
    // Get the file from the cache directly
    cachedContent, err := cacheManager.GetFile("item123")
    require.NoError(t, err, "Failed to get file from cache directly")
    require.Equal(t, "Hello, World!", string(cachedContent), "Directly cached content should match")
}
```

This test:
1. Sets up an integration test environment
2. Configures a mock Graph API client and a real cache
3. Creates a file manager with the mock client and real cache
4. Downloads a file and its content
5. Verifies that the file is cached correctly
6. Verifies that subsequent requests are served from the cache

## Testing with Network Conditions

Here's an example of testing how components interact under different network conditions:

```go
package fs_test

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/fs/graph"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestNetworkConditionsIntegration tests how components interact under different network conditions
func TestNetworkConditionsIntegration(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up isolation config to mock the Graph API
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph"},
        DataIsolation:  true,
    })
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the graph component
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)
    
    // Configure the mock graph component
    mockGraph.AddMockItem("/drive/root:/test.txt", &graph.DriveItem{
        ID:   "item123",
        Name: "test.txt",
        Size: 1024,
        File: &graph.File{
            MimeType: "text/plain",
        },
    })
    mockGraph.AddMockContent("/drive/root:/test.txt", []byte("Hello, World!"))
    
    // Create a file manager with the mock graph client
    fileManager := fs.NewFileManager(mockGraph, logger)
    
    // Get the network simulator
    networkSimulator := env.GetNetworkSimulator()
    
    // Test with fast network
    networkSimulator.ApplyNetworkPreset(testutil.FastNetwork)
    
    // Measure download time with fast network
    startTime := time.Now()
    file, err := fileManager.GetFile("/test.txt")
    require.NoError(t, err, "Failed to get file with fast network")
    content, err := fileManager.GetFileContent(file)
    require.NoError(t, err, "Failed to get file content with fast network")
    fastNetworkDuration := time.Since(startTime)
    
    // Verify the content
    require.Equal(t, "Hello, World!", string(content), "File content should match with fast network")
    
    // Test with slow network
    networkSimulator.ApplyNetworkPreset(testutil.SlowNetwork)
    
    // Measure download time with slow network
    startTime = time.Now()
    file, err = fileManager.GetFile("/test.txt")
    require.NoError(t, err, "Failed to get file with slow network")
    content, err = fileManager.GetFileContent(file)
    require.NoError(t, err, "Failed to get file content with slow network")
    slowNetworkDuration := time.Since(startTime)
    
    // Verify the content
    require.Equal(t, "Hello, World!", string(content), "File content should match with slow network")
    
    // Verify that the slow network test took longer
    require.Greater(t, slowNetworkDuration, fastNetworkDuration, "Slow network test should take longer than fast network test")
    
    // Test with network disconnection
    networkSimulator.Disconnect()
    
    // Try to download the file with disconnected network
    _, err = fileManager.GetFile("/test.txt")
    require.Error(t, err, "Should fail to get file with disconnected network")
    
    // Verify that the error is a network error
    var netErr *graph.NetworkError
    require.ErrorAs(t, err, &netErr, "Error should be a NetworkError")
    
    // Reconnect the network
    networkSimulator.Reconnect()
    
    // Verify that the file can be downloaded again
    file, err = fileManager.GetFile("/test.txt")
    require.NoError(t, err, "Failed to get file after reconnecting")
    content, err = fileManager.GetFileContent(file)
    require.NoError(t, err, "Failed to get file content after reconnecting")
    require.Equal(t, "Hello, World!", string(content), "File content should match after reconnecting")
}
```

This test:
1. Sets up an integration test environment
2. Configures a mock Graph API client
3. Tests file downloads under different network conditions (fast, slow, disconnected)
4. Verifies that the file can be downloaded with fast and slow networks
5. Verifies that the slow network test takes longer
6. Verifies that the file cannot be downloaded with a disconnected network
7. Verifies that the file can be downloaded again after reconnecting

## Scenario-Based Integration Tests

Here's an example of a scenario-based integration test:

```go
package fs_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/fs/graph"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestFileOperationsScenario tests a complete file operations scenario
func TestFileOperationsScenario(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up isolation config to mock the Graph API
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph"},
        DataIsolation:  true,
    })
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the graph component
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)
    
    // Configure the mock graph component
    mockGraph.AddMockItem("/drive/root:/", &graph.DriveItem{
        ID:   "root",
        Name: "",
        Folder: &graph.Folder{
            ChildCount: 0,
        },
    })
    
    // Create a file manager with the mock graph client
    fileManager := fs.NewFileManager(mockGraph, logger)
    
    // Create a test scenario
    scenario := testutil.TestScenario{
        Name:        "File Operations",
        Description: "Tests file creation, modification, and deletion",
        Steps: []testutil.TestStep{
            {
                Name: "Create file",
                Action: func(ctx context.Context) error {
                    // Create a new file
                    return fileManager.CreateFile("/test.txt", []byte("Initial content"))
                },
                Validation: func(ctx context.Context) error {
                    // Verify that the file exists
                    file, err := fileManager.GetFile("/test.txt")
                    if err != nil {
                        return err
                    }
                    if file.Name != "test.txt" {
                        return fmt.Errorf("expected file name 'test.txt', got '%s'", file.Name)
                    }
                    
                    // Verify the content
                    content, err := fileManager.GetFileContent(file)
                    if err != nil {
                        return err
                    }
                    if string(content) != "Initial content" {
                        return fmt.Errorf("expected content 'Initial content', got '%s'", string(content))
                    }
                    
                    return nil
                },
            },
            {
                Name: "Modify file",
                Action: func(ctx context.Context) error {
                    // Modify the file
                    return fileManager.UpdateFile("/test.txt", []byte("Updated content"))
                },
                Validation: func(ctx context.Context) error {
                    // Verify that the file was updated
                    file, err := fileManager.GetFile("/test.txt")
                    if err != nil {
                        return err
                    }
                    
                    // Verify the content
                    content, err := fileManager.GetFileContent(file)
                    if err != nil {
                        return err
                    }
                    if string(content) != "Updated content" {
                        return fmt.Errorf("expected content 'Updated content', got '%s'", string(content))
                    }
                    
                    return nil
                },
            },
            {
                Name: "Delete file",
                Action: func(ctx context.Context) error {
                    // Delete the file
                    return fileManager.DeleteFile("/test.txt")
                },
                Validation: func(ctx context.Context) error {
                    // Verify that the file no longer exists
                    _, err := fileManager.GetFile("/test.txt")
                    if err == nil {
                        return fmt.Errorf("file should not exist after deletion")
                    }
                    
                    // Verify that the error is a "not found" error
                    var notFoundErr *fs.NotFoundError
                    if !errors.As(err, &notFoundErr) {
                        return fmt.Errorf("expected NotFoundError, got %T: %v", err, err)
                    }
                    
                    return nil
                },
            },
        },
        Assertions: []testutil.TestAssertion{
            {
                Name: "File operations completed successfully",
                Condition: func(ctx context.Context) bool {
                    // Verify that the file no longer exists
                    _, err := fileManager.GetFile("/test.txt")
                    return err != nil
                },
                Message: "File operations did not complete successfully",
            },
        },
    }
    
    // Add the scenario to the environment
    env.AddScenario(scenario)
    
    // Run the scenario
    err = env.RunScenario("File Operations")
    require.NoError(t, err, "Scenario should complete successfully")
}
```

This test:
1. Sets up an integration test environment
2. Configures a mock Graph API client
3. Creates a test scenario with multiple steps (create, modify, delete)
4. Runs the scenario and verifies that each step completes successfully
5. Verifies that the overall scenario completes successfully

## Testing Error Handling

Here's an example of testing error handling in an integration test:

```go
package fs_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/fs/graph"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestErrorHandlingIntegration tests how components handle errors
func TestErrorHandlingIntegration(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(context.Background(), logger)
    require.NotNil(t, env)
    
    // Set up isolation config to mock the Graph API
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph"},
        DataIsolation:  true,
    })
    
    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)
    
    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })
    
    // Get the graph component
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)
    
    // Configure the mock graph component to return errors
    mockGraph.AddErrorResponse("/drive/root:/not-found.txt", graph.NewGraphError(404, "Not Found"))
    mockGraph.AddErrorResponse("/drive/root:/permission-denied.txt", graph.NewGraphError(403, "Permission Denied"))
    mockGraph.AddErrorResponse("/drive/root:/server-error.txt", graph.NewGraphError(500, "Internal Server Error"))
    
    // Create a file manager with the mock graph client
    fileManager := fs.NewFileManager(mockGraph, logger)
    
    // Test handling of "not found" error
    t.Run("NotFoundError", func(t *testing.T) {
        // Try to get a non-existent file
        file, err := fileManager.GetFile("/not-found.txt")
        
        // Verify that an error is returned
        require.Error(t, err, "Should fail to get non-existent file")
        require.Nil(t, file, "File should be nil")
        
        // Verify that the error is a "not found" error
        var notFoundErr *fs.NotFoundError
        require.ErrorAs(t, err, &notFoundErr, "Error should be a NotFoundError")
        require.Equal(t, "/not-found.txt", notFoundErr.Path, "Error should contain the correct path")
    })
    
    // Test handling of "permission denied" error
    t.Run("PermissionDeniedError", func(t *testing.T) {
        // Try to get a file with permission denied
        file, err := fileManager.GetFile("/permission-denied.txt")
        
        // Verify that an error is returned
        require.Error(t, err, "Should fail to get file with permission denied")
        require.Nil(t, file, "File should be nil")
        
        // Verify that the error is a "permission denied" error
        var permissionErr *fs.PermissionError
        require.ErrorAs(t, err, &permissionErr, "Error should be a PermissionError")
        require.Equal(t, "/permission-denied.txt", permissionErr.Path, "Error should contain the correct path")
    })
    
    // Test handling of server error
    t.Run("ServerError", func(t *testing.T) {
        // Try to get a file with server error
        file, err := fileManager.GetFile("/server-error.txt")
        
        // Verify that an error is returned
        require.Error(t, err, "Should fail to get file with server error")
        require.Nil(t, file, "File should be nil")
        
        // Verify that the error is a server error
        var serverErr *fs.ServerError
        require.ErrorAs(t, err, &serverErr, "Error should be a ServerError")
        require.Equal(t, 500, serverErr.StatusCode, "Error should contain the correct status code")
    })
    
    // Test retry behavior
    t.Run("RetryBehavior", func(t *testing.T) {
        // Configure the mock to return a server error for the first 2 calls, then succeed
        mockGraph.AddRetryResponse("/drive/root:/retry.txt", 
            []testutil.MockResponse{
                {Error: graph.NewGraphError(500, "Internal Server Error")},
                {Error: graph.NewGraphError(500, "Internal Server Error")},
                {Item: &graph.DriveItem{
                    ID:   "retry123",
                    Name: "retry.txt",
                    Size: 1024,
                    File: &graph.File{
                        MimeType: "text/plain",
                    },
                }},
            })
        mockGraph.AddMockContent("/drive/root:/retry.txt", []byte("Retry succeeded"))
        
        // Try to get the file (should succeed after retries)
        file, err := fileManager.GetFile("/retry.txt")
        require.NoError(t, err, "Should succeed after retries")
        require.NotNil(t, file, "File should not be nil")
        
        // Verify the file properties
        require.Equal(t, "retry.txt", file.Name, "File name should match")
        
        // Verify that the graph client was called multiple times
        recorder := mockGraph.GetRecorder()
        require.Equal(t, 3, recorder.GetCallCount("GetItemPath"), "GetItemPath should be called 3 times")
    })
}
```

This test:
1. Sets up an integration test environment
2. Configures a mock Graph API client to return different types of errors
3. Tests how the file manager handles different error conditions
4. Verifies that the correct error types are returned
5. Tests retry behavior when temporary errors occur

These examples demonstrate different aspects of integration testing with the OneMount test framework. By following these patterns, you can write comprehensive integration tests that verify your components work together correctly under various conditions.