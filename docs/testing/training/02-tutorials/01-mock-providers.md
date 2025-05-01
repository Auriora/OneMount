# Tutorial: Setting Up and Using Mock Providers

This tutorial will guide you through the process of setting up and using mock providers in your tests. Mock providers are essential for isolating components during testing and simulating external dependencies.

## Table of Contents

1. [Introduction to Mock Providers](#introduction-to-mock-providers)
2. [Available Mock Providers](#available-mock-providers)
3. [Setting Up Mock Providers](#setting-up-mock-providers)
4. [Configuring Mock Responses](#configuring-mock-responses)
5. [Verifying Interactions](#verifying-interactions)
6. [Best Practices](#best-practices)
7. [Complete Example](#complete-example)

## Introduction to Mock Providers

Mock providers are implementations of interfaces that simulate the behavior of real components or external dependencies. They allow you to:

- Test components in isolation
- Control the behavior of dependencies
- Simulate error conditions and edge cases
- Verify interactions with dependencies

The OneMount test framework provides a comprehensive mocking infrastructure that includes mock providers for various components of the system.

## Available Mock Providers

The OneMount test framework includes the following mock providers:

1. **MockGraphProvider**: Simulates the Microsoft Graph API with configurable responses, network condition simulation, and call recording.
2. **MockFileSystemProvider**: Simulates filesystem operations with a virtual filesystem.
3. **MockUIProvider**: Simulates UI interactions and events.

Each mock provider implements the corresponding interface from the real system, allowing you to use them as drop-in replacements for the real components.

## Setting Up Mock Providers

To use mock providers in your tests, you need to:

1. Create instances of the mock providers
2. Register them with the TestFramework
3. Configure their behavior

Here's an example of setting up mock providers:

```go
// Create a logger
logger := log.With().Str("component", "test").Logger()

// Create a test configuration
config := testutil.TestConfig{
    Environment:    "test",
    Timeout:        30,
    VerboseLogging: true,
}

// Create a new TestFramework
framework := testutil.NewTestFramework(config, &logger)

// Add cleanup using t.Cleanup to ensure resources are cleaned up
t.Cleanup(func() {
    framework.CleanupResources()
})

// Create and register mock providers
mockGraph := testutil.NewMockGraphProvider()
framework.RegisterMockProvider("graph", mockGraph)

mockFS := testutil.NewMockFileSystemProvider()
framework.RegisterMockProvider("filesystem", mockFS)

mockUI := testutil.NewMockUIProvider()
framework.RegisterMockProvider("ui", mockUI)
```

## Configuring Mock Responses

Once you've set up your mock providers, you need to configure their behavior. Each mock provider has methods for configuring responses for different operations.

### Configuring MockGraphProvider

```go
// Configure responses for specific resources
resource := "/me/drive/root:/test.txt"
mockItem := &graph.DriveItem{
    ID:   "item123",
    Name: "test.txt",
    Size: 1024,
    File: &graph.File{
        MimeType: "text/plain",
    },
}
mockGraph.AddMockItem(resource, mockItem)

// Configure error responses
mockGraph.AddErrorResponse("/me/drive/root:/nonexistent.txt", graph.NewGraphError(404, "Not Found"))

// Configure custom behavior
mockGraph.SetConfig(graph.MockConfig{
    ErrorRate:     0.05, // 5% random error rate
    ResponseDelay: 50 * time.Millisecond,
    CustomBehavior: map[string]interface{}{
        "retryCount": 3,
    },
})
```

### Configuring MockFileSystemProvider

```go
// Add files to the virtual filesystem
mockFS.AddFile("/test.txt", []byte("Hello, World!"), 0644)
mockFS.AddDirectory("/testdir", 0755)

// Configure error conditions
mockFS.AddErrorResponse("/error.txt", fs.ErrNotExist)

// Configure custom behavior
mockFS.SetConfig(fs.MockConfig{
    ErrorRate:     0.05, // 5% random error rate
    ResponseDelay: 10 * time.Millisecond,
})
```

### Configuring MockUIProvider

```go
// Configure UI events
mockUI.AddEvent("click", "button1", map[string]interface{}{
    "x": 100,
    "y": 200,
})

// Configure UI state
mockUI.SetState("window1", map[string]interface{}{
    "visible": true,
    "width":   800,
    "height":  600,
})
```

## Verifying Interactions

After running your tests, you can verify that the mock providers were used correctly by checking the recorded interactions.

```go
// Get the recorder from the mock provider
recorder := mockGraph.GetRecorder()

// Verify that a specific method was called
if !recorder.VerifyCall("GetItemPath", 1) {
    t.Errorf("Expected GetItemPath to be called once")
}

// Get all recorded calls
calls := recorder.GetCalls()
for _, call := range calls {
    fmt.Printf("Method: %s, Args: %v, Timestamp: %v\n", call.Method, call.Args, call.Timestamp)
}

// Verify call arguments
for _, call := range calls {
    if call.Method == "GetItemPath" {
        path := call.Args[0].(string)
        if path != "/test.txt" {
            t.Errorf("Expected path to be /test.txt, got %s", path)
        }
    }
}
```

## Best Practices

When using mock providers, follow these best practices:

1. **Keep mocks simple**: Configure only the behavior you need for your test.
2. **Verify important interactions**: Check that the component under test interacts with the mock as expected.
3. **Test error handling**: Configure your mocks to return errors to test error handling in your code.
4. **Use realistic data**: Configure your mocks with data that resembles what the real component would return.
5. **Clean up resources**: Always clean up resources after your tests.
6. **Isolate tests**: Each test should have its own mock providers with their own configuration.
7. **Document mock behavior**: Comment your mock configuration to explain what behavior you're simulating.

## Complete Example

Here's a complete example of using mock providers in a test:

```go
package mypackage_test

import (
    "context"
    "testing"
    "time"

    "github.com/rs/zerolog/log"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/fs/graph"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestFileDownload(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "test").Logger()

    // Create a test configuration
    config := testutil.TestConfig{
        Environment:    "test",
        Timeout:        30,
        VerboseLogging: true,
    }

    // Create a new TestFramework
    framework := testutil.NewTestFramework(config, &logger)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        framework.CleanupResources()
    })

    // Create and register mock providers
    mockGraph := testutil.NewMockGraphProvider()
    framework.RegisterMockProvider("graph", mockGraph)

    mockFS := testutil.NewMockFileSystemProvider()
    framework.RegisterMockProvider("filesystem", mockFS)

    // Configure mock responses
    resource := "/me/drive/root:/test.txt"
    mockItem := &graph.DriveItem{
        ID:   "item123",
        Name: "test.txt",
        Size: 1024,
        File: &graph.File{
            MimeType: "text/plain",
        },
    }
    mockGraph.AddMockItem(resource, mockItem)
    mockGraph.AddMockContent(resource, []byte("Hello, World!"))

    // Run the test
    result := framework.RunTest("file-download-test", func(ctx context.Context) error {
        // Get the file from the Graph API
        item, err := mockGraph.GetItemPath("/test.txt")
        if err != nil {
            return err
        }

        // Download the file content
        content, err := mockGraph.GetItemContent(item.ID)
        if err != nil {
            return err
        }

        // Write the file to the filesystem
        err = mockFS.WriteFile("/test.txt", content, 0644)
        if err != nil {
            return err
        }

        // Read the file from the filesystem
        readContent, err := mockFS.ReadFile("/test.txt")
        if err != nil {
            return err
        }

        // Verify the content
        if string(readContent) != "Hello, World!" {
            return fmt.Errorf("file content mismatch: got %q, want %q", string(readContent), "Hello, World!")
        }

        return nil
    })

    // Check the result
    require.Equal(t, testutil.TestStatusPassed, result.Status, "Test failed: %v", result.Failures)

    // Verify interactions with the mock providers
    graphRecorder := mockGraph.GetRecorder()
    require.True(t, graphRecorder.VerifyCall("GetItemPath", 1), "Expected GetItemPath to be called once")
    require.True(t, graphRecorder.VerifyCall("GetItemContent", 1), "Expected GetItemContent to be called once")

    fsRecorder := mockFS.GetRecorder()
    require.True(t, fsRecorder.VerifyCall("WriteFile", 1), "Expected WriteFile to be called once")
    require.True(t, fsRecorder.VerifyCall("ReadFile", 1), "Expected ReadFile to be called once")
}
```

This example demonstrates:
1. Setting up mock providers
2. Configuring mock responses
3. Running a test that uses the mock providers
4. Verifying interactions with the mock providers

By following this pattern, you can write tests that are isolated, deterministic, and focused on the behavior you want to test.