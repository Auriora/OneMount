# Tutorial: Integration Testing

This tutorial will guide you through the process of writing integration tests using the OneMount test framework. Integration tests verify that different components of the system work together correctly.

> **Note**: All code examples in this tutorial are for illustration purposes only and may need to be adapted to your specific project structure and imports. The examples are not meant to be compiled directly but rather to demonstrate concepts and patterns.

## Table of Contents

1. [Introduction to Integration Testing](#introduction-to-integration-testing)
2. [The IntegrationTestEnvironment Component](#the-integrationtestenvironment-component)
3. [Setting Up an Integration Test Environment](#setting-up-an-integration-test-environment)
4. [Configuring Components](#configuring-components)
5. [Scenario-Based Testing](#scenario-based-testing)
6. [Testing Component Interactions](#testing-component-interactions)
7. [Best Practices](#best-practices)
8. [Complete Example](#complete-example)

## Introduction to Integration Testing

Integration testing is a critical part of the testing process that verifies different components of the system work together correctly. Unlike unit tests, which test individual components in isolation, integration tests focus on the interactions between components.

Integration tests help you:

- Verify that components interact correctly
- Identify issues that only occur when components are integrated
- Ensure that the system works as a whole
- Validate interface contracts between components
- Test end-to-end workflows

The OneMount test framework provides an IntegrationTestEnvironment component that helps you set up and run integration tests.

## The IntegrationTestEnvironment Component

The IntegrationTestEnvironment component provides a controlled environment for integration tests. It allows you to:

- Configure which components are real and which are mocked
- Integrate with the NetworkSimulator for testing under different network conditions
- Manage test data
- Support component isolation via network rules
- Define and execute test scenarios
- Set up and tear down the test environment

You can create an IntegrationTestEnvironment as follows:

```go
// Create a logger
logger := &testutil.TestLogger{}

// Create a context
ctx := context.Background()

// Create a new IntegrationTestEnvironment
env := testutil.NewIntegrationTestEnvironment(ctx, logger)
```

## Setting Up an Integration Test Environment

To set up an integration test environment, you need to:

1. Create an IntegrationTestEnvironment instance
2. Configure which components should be real and which should be mocked
3. Set up the environment
4. Add cleanup to ensure resources are released

Here's an example:

```go
// Create a logger
logger := &testutil.TestLogger{}

// Create a context
ctx := context.Background()

// Create a new IntegrationTestEnvironment
env := testutil.NewIntegrationTestEnvironment(ctx, logger)

// Configure which components should be mocked
env.SetIsolationConfig(testutil.IsolationConfig{
    MockedServices: []string{"graph", "filesystem", "ui"},
    NetworkRules:   []testutil.NetworkRule{},
    DataIsolation:  true,
})

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

## Configuring Components

Once you've set up the environment, you can configure the components for your test. This includes:

1. Getting components from the environment
2. Configuring mock components with expected behavior
3. Setting up test data

Here's an example:

```go
// Get the graph component
graphComponent, err := env.GetComponent("graph")
if err != nil {
    // Handle error
}

// Configure the mock graph component
mockGraph := graphComponent.(*testutil.MockGraphProvider)
mockGraph.AddMockItem("/drive/root", &graph.DriveItem{
    Name: "root",
    // ...
})

// Get the filesystem component
fsComponent, err := env.GetComponent("filesystem")
if err != nil {
    // Handle error
}

// Configure the mock filesystem component
mockFS := fsComponent.(*testutil.MockFileSystemProvider)
mockFS.AddFile("/test.txt", []byte("Hello, World!"), 0644)
```

## Scenario-Based Testing

The IntegrationTestEnvironment supports scenario-based testing, which allows you to define and execute test scenarios with multiple steps. A test scenario consists of:

1. A name and description
2. A series of test steps, each with an action and optional validation
3. Assertions that verify the overall outcome of the scenario
4. Cleanup steps that run after the scenario completes

Here's an example of defining and running a test scenario:

```go
// Create a test scenario
scenario := testutil.TestScenario{
    Name:        "File Operations",
    Description: "Tests file creation, modification, and deletion",
    Steps: []testutil.TestStep{
        {
            Name: "Create file",
            Action: func(ctx context.Context) error {
                // Implementation would create a file
                return nil
            },
            Validation: func(ctx context.Context) error {
                // Implementation would verify file was created
                return nil
            },
        },
        {
            Name: "Modify file",
            Action: func(ctx context.Context) error {
                // Implementation would modify the file
                return nil
            },
            Validation: func(ctx context.Context) error {
                // Implementation would verify file was modified
                return nil
            },
        },
        // More steps...
    },
    Assertions: []testutil.TestAssertion{
        {
            Name: "File operations completed successfully",
            Condition: func(ctx context.Context) bool {
                // Implementation would check if all operations were successful
                return true
            },
            Message: "File operations did not complete successfully",
        },
    },
    Cleanup: []testutil.CleanupStep{
        {
            Name: "Clean up test files",
            Action: func(ctx context.Context) error {
                // Implementation would clean up any remaining test files
                return nil
            },
            AlwaysRun: true,
        },
    },
}

// Add the scenario to the environment
env.AddScenario(scenario)

// Run the scenario
err := env.RunScenario("File Operations")
if err != nil {
    // Handle error
}
```

## Testing Component Interactions

Integration tests should focus on verifying that components interact correctly. This includes:

1. Testing that components communicate correctly
2. Verifying that data is passed correctly between components
3. Testing error handling and edge cases
4. Verifying that components handle different network conditions

Here's an example of testing the interaction between the graph and filesystem components:

```go
// Create a test scenario for component interaction
scenario := testutil.TestScenario{
    Name:        "Graph to Filesystem Interaction",
    Description: "Tests the interaction between the graph and filesystem components",
    Steps: []testutil.TestStep{
        {
            Name: "Download file from Graph API",
            Action: func(ctx context.Context) error {
                // Get the graph component
                graphComponent, err := env.GetComponent("graph")
                if err != nil {
                    return err
                }
                mockGraph := graphComponent.(*testutil.MockGraphProvider)

                // Get the filesystem component
                fsComponent, err := env.GetComponent("filesystem")
                if err != nil {
                    return err
                }
                mockFS := fsComponent.(*testutil.MockFileSystemProvider)

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

                return nil
            },
            Validation: func(ctx context.Context) error {
                // Get the filesystem component
                fsComponent, err := env.GetComponent("filesystem")
                if err != nil {
                    return err
                }
                mockFS := fsComponent.(*testutil.MockFileSystemProvider)

                // Verify the file exists in the filesystem
                content, err := mockFS.ReadFile("/test.txt")
                if err != nil {
                    return err
                }

                // Verify the content
                if string(content) != "Hello, World!" {
                    return fmt.Errorf("file content mismatch: got %q, want %q", string(content), "Hello, World!")
                }

                return nil
            },
        },
    },
}

// Add and run the scenario
env.AddScenario(scenario)
err := env.RunScenario("Graph to Filesystem Interaction")
if err != nil {
    // Handle error
}
```

## Best Practices

When writing integration tests, follow these best practices:

1. **Use clean test environments**: Always start with a clean test environment to ensure that tests are isolated from each other.

2. **Implement proper cleanup**: Always clean up resources after tests to prevent interference with other tests. Use `t.Cleanup()` to ensure cleanup happens even if the test fails.

3. **Test failure scenarios**: Don't just test the happy path. Test how your components handle failures, such as network errors, API errors, and invalid inputs.

4. **Use realistic test data**: Use realistic test data that represents what your system will encounter in production. This helps ensure that your tests catch real-world issues.

5. **Isolate components when necessary**: Use mocks to isolate components when testing specific interactions. This helps identify which component is causing a failure.

6. **Test component boundaries**: Focus on testing the boundaries between components, where data is passed from one component to another. This is where integration issues often occur.

7. **Use dynamic waiting**: Use dynamic waiting instead of fixed timeouts to make tests more reliable. This is especially important for asynchronous operations.

8. **Document test scenarios**: Document what each integration test is verifying to make it clear what's being tested and why.

9. **Test with different network conditions**: Use the NetworkSimulator to test how your components interact under different network conditions.

10. **Verify interactions**: Verify that components interact as expected by checking that the right methods are called with the right arguments.

## Complete Example

Here's a complete example of an integration test:

```go
package mypackage_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestIntegration(t *testing.T) {
    // Create a logger
    logger := &testutil.TestLogger{}

    // Create a context
    ctx := context.Background()

    // Create a test environment
    env := testutil.NewIntegrationTestEnvironment(ctx, logger)
    require.NotNil(t, env)

    // Set up isolation config to mock all components
    env.SetIsolationConfig(testutil.IsolationConfig{
        MockedServices: []string{"graph", "filesystem", "ui"},
        NetworkRules:   []testutil.NetworkRule{},
        DataIsolation:  true,
    })

    // Set up the environment
    err := env.SetupEnvironment()
    require.NoError(t, err)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        env.TeardownEnvironment()
    })

    // Get and configure the graph component
    graphComponent, err := env.GetComponent("graph")
    require.NoError(t, err)
    mockGraph := graphComponent.(*testutil.MockGraphProvider)
    mockGraph.AddMockItem("/drive/root:/test.txt", &graph.DriveItem{
        ID:   "item123",
        Name: "test.txt",
        Size: 1024,
        File: &graph.File{
            MimeType: "text/plain",
        },
    })
    mockGraph.AddMockContent("/drive/root:/test.txt", []byte("Hello, World!"))

    // Get and configure the filesystem component
    fsComponent, err := env.GetComponent("filesystem")
    require.NoError(t, err)
    mockFS := fsComponent.(*testutil.MockFileSystemProvider)

    // Create a test scenario
    scenario := testutil.TestScenario{
        Name:        "File Download",
        Description: "Tests downloading a file from the Graph API to the filesystem",
        Steps: []testutil.TestStep{
            {
                Name: "Download file",
                Action: func(ctx context.Context) error {
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

                    return nil
                },
                Validation: func(ctx context.Context) error {
                    // Verify the file exists in the filesystem
                    content, err := mockFS.ReadFile("/test.txt")
                    if err != nil {
                        return err
                    }

                    // Verify the content
                    if string(content) != "Hello, World!" {
                        return fmt.Errorf("file content mismatch: got %q, want %q", string(content), "Hello, World!")
                    }

                    return nil
                },
            },
        },
        Assertions: []testutil.TestAssertion{
            {
                Name: "File was downloaded successfully",
                Condition: func(ctx context.Context) bool {
                    // Verify the file exists in the filesystem
                    _, err := mockFS.ReadFile("/test.txt")
                    return err == nil
                },
                Message: "File was not downloaded successfully",
            },
        },
        Cleanup: []testutil.CleanupStep{
            {
                Name: "Clean up test files",
                Action: func(ctx context.Context) error {
                    // Remove the test file
                    err := mockFS.Remove("/test.txt")
                    if err != nil && !os.IsNotExist(err) {
                        return err
                    }
                    return nil
                },
                AlwaysRun: true,
            },
        },
    }

    // Add and run the scenario
    env.AddScenario(scenario)
    err = env.RunScenario("File Download")
    require.NoError(t, err)

    // Test with network disconnection
    networkSimulator := env.GetNetworkSimulator()
    err = networkSimulator.Disconnect()
    require.NoError(t, err)

    // Create an offline scenario
    offlineScenario := testutil.TestScenario{
        Name:        "Offline Mode",
        Description: "Tests operation in offline mode",
        Steps: []testutil.TestStep{
            {
                Name: "Verify offline status",
                Action: func(ctx context.Context) error {
                    // Verify that the network is disconnected
                    if networkSimulator.IsConnected() {
                        return fmt.Errorf("network should be disconnected")
                    }
                    return nil
                },
            },
            {
                Name: "Try to download file",
                Action: func(ctx context.Context) error {
                    // Try to get the file from the Graph API
                    _, err := mockGraph.GetItemPath("/test.txt")

                    // Verify that an error is returned
                    if err == nil {
                        return fmt.Errorf("expected error when network is disconnected, got nil")
                    }

                    // Verify that the error is a network error
                    var netErr *graph.NetworkError
                    if !errors.As(err, &netErr) {
                        return fmt.Errorf("expected NetworkError, got %T: %v", err, err)
                    }

                    return nil
                },
            },
        },
    }

    // Add and run the offline scenario
    env.AddScenario(offlineScenario)
    err = env.RunScenario("Offline Mode")
    require.NoError(t, err)

    // Reconnect for cleanup
    err = networkSimulator.Reconnect()
    require.NoError(t, err)
}
```

This example demonstrates:
1. Setting up an integration test environment
2. Configuring mock components
3. Creating and running test scenarios
4. Testing component interactions
5. Testing with network disconnection
6. Proper cleanup of resources

By following these patterns, you can write comprehensive integration tests that verify your components work together correctly under various conditions.
