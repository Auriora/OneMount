# Getting Started with the OneMount Test Framework

This guide will help you get started with the OneMount test framework. It covers the basics of the framework, how to set up a test environment, and how to write your first test.

## Table of Contents

1. [Introduction to the Test Framework](#introduction-to-the-test-framework)
2. [Key Components](#key-components)
3. [Setting Up a Test Environment](#setting-up-a-test-environment)
4. [Writing Your First Test](#writing-your-first-test)
5. [Running Tests](#running-tests)
6. [Next Steps](#next-steps)

## Introduction to the Test Framework

The OneMount test framework is a comprehensive testing solution designed to support various types of tests for the OneMount filesystem. It provides a centralized test configuration and execution environment, along with utilities for mocking external dependencies, simulating network conditions, and managing test resources.

The framework was implemented in phases, with each phase focusing on specific aspects of the test infrastructure:

- **Phase 1**: Core Test Framework - Basic TestFramework structure, mock providers, and coverage reporting
- **Phase 2**: Mock Infrastructure - Graph API mocks, filesystem mocks, and network condition simulation
- **Phase 3**: Integration and Performance Testing - Integration test environment, scenario-based testing, and performance benchmarking
- **Phase 4**: Advanced Features - Advanced coverage reporting, load testing, and performance metrics collection
- **Phase 5**: Test Types Implementation - Unit, integration, system, and security testing frameworks

## Key Components

The test framework consists of several key components:

1. **TestFramework**: The core component that provides centralized test configuration, setup, and execution. It manages test resources, mock providers, test execution, and context management.

2. **Mocking Infrastructure**: Provides mock implementations of external dependencies and components, such as the Microsoft Graph API, filesystem operations, and UI interactions.

3. **Network Simulation**: Allows testing under different network conditions, such as latency, packet loss, and bandwidth limitations.

4. **Integration Test Environment**: Provides a controlled environment for integration tests, with configurable components and network conditions.

5. **Coverage Reporting**: Tracks and reports test coverage metrics, such as line coverage, function coverage, and branch coverage.

## Setting Up a Test Environment

To set up a test environment, you need to create a `TestFramework` instance with the appropriate configuration. Here's a basic example:

```go
package mypackage_test

import (
    "testing"
    "github.com/rs/zerolog/log"
    "github.com/yourusername/onemount/internal/testutil"
)

func TestMyFeature(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "test").Logger()

    // Create a test configuration
    config := testutil.TestConfig{
        Environment:    "test",
        Timeout:        30,  // 30 seconds
        VerboseLogging: true,
        ArtifactsDir:   "/tmp/test-artifacts",
    }

    // Create a new TestFramework
    framework := testutil.NewTestFramework(config, &logger)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        framework.CleanupResources()
    })

    // Now you can use the framework to run tests
    // ...
}
```

## Writing Your First Test

Let's write a simple test that verifies a file can be created and read:

```go
func TestFileOperations(t *testing.T) {
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

    // Register mock providers
    mockFS := NewMockFileSystemProvider()
    framework.RegisterMockProvider("filesystem", mockFS)

    // Run the test
    result := framework.RunTest("file-creation-test", func(ctx context.Context) error {
        // Create a file
        filePath := "/test.txt"
        content := "Hello, World!"
        err := mockFS.WriteFile(filePath, []byte(content), 0644)
        if err != nil {
            return fmt.Errorf("failed to create file: %w", err)
        }

        // Read the file
        data, err := mockFS.ReadFile(filePath)
        if err != nil {
            return fmt.Errorf("failed to read file: %w", err)
        }

        // Verify the content
        if string(data) != content {
            return fmt.Errorf("file content mismatch: got %q, want %q", string(data), content)
        }

        return nil
    })

    // Check the result
    if result.Status != testutil.TestStatusPassed {
        t.Errorf("Test failed: %v", result.Failures)
    }
}
```

## Running Tests

You can run your tests using the standard Go test command:

```bash
go test ./...
```

To run tests with coverage reporting:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

To run specific tests:

```bash
go test ./path/to/package -run TestName
```

## Next Steps

Now that you've learned the basics of the OneMount test framework, you can:

1. Explore the [Step-by-Step Tutorials](02-tutorials/) to learn how to perform common testing tasks
2. Check out the [Test Examples](03-examples/) to see how different types of tests are implemented
3. Practice your skills with the [Practice Exercises](04-exercises/)
4. Learn about advanced topics in the [Advanced Topics](05-advanced-topics/) section

For more information about best practices for writing tests, refer to the [Test Guidelines](../../guides/test-guidelines.md).