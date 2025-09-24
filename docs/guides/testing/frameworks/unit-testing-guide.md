# Unit Testing Guide

## Overview

Unit testing is a fundamental testing approach that focuses on testing individual components or units of code in isolation. In the OneMount project, unit tests verify that each component functions correctly on its own, without dependencies on other components.

The OneMount test framework provides specialized utilities for unit testing, making it easier to write comprehensive and effective unit tests.

## Key Concepts

### Unit Testing Principles

1. **Isolation**: Test components in isolation from their dependencies
2. **Determinism**: Tests should produce the same results every time they run
3. **Speed**: Unit tests should run quickly
4. **Coverage**: Aim for high code coverage with unit tests
5. **Simplicity**: Keep unit tests simple and focused on a single behavior

### Test Structure

Unit tests in the OneMount project follow a standard structure:

1. **Setup**: Prepare the test environment and create the component under test
2. **Exercise**: Call the method or function being tested
3. **Verify**: Check that the results match expectations
4. **Teardown**: Clean up any resources created during the test

## Unit Testing Framework

The OneMount unit testing framework is designed to make it easy to write comprehensive unit tests for all components of the OneMount system. It provides a set of utilities and helpers that simplify common testing tasks and promote consistent testing practices.

### UnitTestFramework

The `UnitTestFramework` provides utilities for unit testing:

```
type UnitTestFramework struct {
    // Embedded TestFramework for core functionality
    *TestFramework

    // Additional unit testing utilities
    Assertions *AssertionHelper
    Mocks      *MockHelper
    Fixtures   *FixtureHelper
}
```

### Creating a UnitTestFramework

```
// Create a unit test framework
framework := testutil.NewUnitTestFramework(t)

// Or with custom configuration
config := testutil.TestConfig{
    Environment:    "unit-test",
    Timeout:        10,
    VerboseLogging: true,
}
framework := testutil.NewUnitTestFrameworkWithConfig(t, config)
```

## Key Components

### Test Fixtures

Test fixtures provide a way to set up and tear down test environments. They encapsulate the setup and teardown logic, making tests more readable and maintainable.

```
// Create a fixture
fixture := NewUnitTestFixture("my-fixture")
    .WithSetup(func(t *testing.T) (interface{}, error) {
        // Set up test environment
        return myTestData, nil
    })
    .WithTeardown(func(t *testing.T, fixture interface{}) error {
        // Clean up test environment
        return nil
    })
    .WithData("key", "value")

// Use the fixture in a test
fixture.Use(t, func(t *testing.T, fixture interface{}) {
    // Test logic using the fixture
    data := fixture.(MyTestData)
    // ...
})
```

The `FixtureHelper` provides utilities for managing test fixtures:

```
// Load a test fixture
data := framework.Fixtures.Load("testdata/fixture.json")

// Create a temporary file with content
filePath := framework.Fixtures.CreateTempFile("test-file", []byte("content"))

// Create a temporary directory
dirPath := framework.Fixtures.CreateTempDir("test-dir")

// Clean up fixtures (automatically called at the end of the test)
framework.Fixtures.Cleanup()
```

### Mock Objects

Mock objects provide a way to simulate dependencies for isolated testing. They allow you to control the behavior of dependencies and verify interactions.

```
// Create a mock
mock := NewMock(t, "my-mock")

// Set up expectations
mock.On("method1").WithArgs("arg1", "arg2").Return("result1")
mock.On("method2").WithArgs(1, 2).Return(3).ReturnError(nil)
mock.On("method3").WithArgs().ReturnError(errors.New("test error"))
mock.On("method4").WithArgs("arg").SetTimes(2)

// Call methods on the mock
result1 := mock.Call("method1", "arg1", "arg2")
result2 := mock.Call("method2", 1, 2)
result3 := mock.Call("method3")
mock.Call("method4", "arg")
mock.Call("method4", "arg")

// Verify expectations
mock.Verify()
```

The `MockHelper` provides utilities for creating and managing mocks:

```
// Create a mock
mockGraph := framework.Mocks.NewMockGraphProvider()
mockFS := framework.Mocks.NewMockFileSystemProvider()
mockUI := framework.Mocks.NewMockUIProvider()

// Configure mock behavior
mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
    ID:   "root",
    Name: "root",
})

// Verify mock interactions
framework.Mocks.Verify(func() {
    mockGraph.VerifyCalled("/me/drive/root")
    mockFS.VerifyOperation("/path/to/file.txt", testutil.FSOperationRead)
})
```

### Assertion Utilities

Assertion utilities provide a way to verify expected outcomes in tests. They provide a consistent way to check conditions and report failures.

```
// Create an assert object
assert := NewAssert(t)

// Use assertions
assert.Equal(expected, actual)
assert.NotEqual(expected, actual)
assert.Nil(value)
assert.NotNil(value)
assert.True(value)
assert.False(value)
assert.NoError(err)
assert.Error(err)
assert.ErrorContains(err, "expected error message")
assert.Len(collection, length)
assert.Contains(collection, element)
```

The `AssertionHelper` provides utilities for making assertions in unit tests:

```
// Basic assertions
framework.Assertions.Equal(expected, actual, "values should be equal")
framework.Assertions.NotEqual(notExpected, actual, "values should not be equal")
framework.Assertions.Nil(value, "value should be nil")
framework.Assertions.NotNil(value, "value should not be nil")
framework.Assertions.True(condition, "condition should be true")
framework.Assertions.False(condition, "condition should be false")

// Error assertions
framework.Assertions.Error(err, "should return an error")
framework.Assertions.NoError(err, "should not return an error")
framework.Assertions.ErrorContains(err, "substring", "error should contain substring")
framework.Assertions.ErrorIs(err, expectedErr, "error should match expected error")

// Collection assertions
framework.Assertions.Contains(collection, element, "collection should contain element")
framework.Assertions.NotContains(collection, element, "collection should not contain element")
framework.Assertions.Len(collection, length, "collection should have expected length")
framework.Assertions.Empty(collection, "collection should be empty")
framework.Assertions.NotEmpty(collection, "collection should not be empty")

// Advanced assertions
framework.Assertions.JSONEqual(expected, actual, "JSON should be equal")
framework.Assertions.Subset(superset, subset, "should be a subset")
framework.Assertions.ElementsMatch(expected, actual, "should have same elements")
framework.Assertions.Eventually(condition, timeout, interval, "condition should eventually be true")
```

### Table-Driven Tests

Table-driven tests provide a way to run the same test logic with different inputs and expected outputs. They make it easy to test multiple scenarios with minimal code duplication.

```
// Define test cases
tests := []TableTest{
    {
        Name:     "Addition",
        Input:    []int{1, 2},
        Expected: 3,
    },
    {
        Name:     "Subtraction",
        Input:    []int{5, 3},
        Expected: 2,
    },
    {
        Name:          "Division by zero",
        Input:         []int{5, 0},
        ExpectedError: errors.New("division by zero"),
    },
}

// Run the tests
RunTableTests(t, tests, func(t *testing.T, test TableTest) (interface{}, error) {
    input := test.Input.([]int)
    if len(input) != 2 {
        return nil, errors.New("input must be a slice of 2 integers")
    }

    a, b := input[0], input[1]

    switch test.Name {
    case "Addition":
        return a + b, nil
    case "Subtraction":
        return a - b, nil
    case "Division by zero":
        if b == 0 {
            return nil, errors.New("division by zero")
        }
        return a / b, nil
    default:
        return nil, errors.New("unknown operation")
    }
})
```

### Edge Case Testing

Edge case testing utilities provide a way to test edge cases and error conditions. They make it easy to generate test data for edge cases and simulate error conditions.

```
// Generate edge cases
generator := NewEdgeCaseGenerator()
stringCases := generator.StringEdgeCases()
intCases := generator.IntEdgeCases()
floatCases := generator.FloatEdgeCases()
boolCases := generator.BoolEdgeCases()
timeCases := generator.TimeEdgeCases()
sliceCases := generator.SliceEdgeCases()
mapCases := generator.MapEdgeCases()

// Test with edge cases
for _, str := range stringCases {
    // Test with string edge case
}

// Define error conditions
conditions := []*ErrorCondition{
    NewErrorCondition("No error").
        WithFunc(func() error {
            return nil
        }),
    NewErrorCondition("With error").
        WithFunc(func() error {
            return errors.New("test error")
        }).
        WithExpectedError(errors.New("test error")),
    NewErrorCondition("With recovery").
        WithFunc(func() error {
            return errors.New("test error")
        }).
        WithExpectedError(errors.New("test error")).
        WithRecovery(func() error {
            return nil
        }),
}

// Run error conditions
RunErrorConditions(t, conditions)
```

## Writing Unit Tests

### Basic Unit Test

```
func TestSomeFunction(t *testing.T) {
    // Create a unit test framework
    framework := testutil.NewUnitTestFramework(t)

    // Create the component under test
    component := NewSomeComponent()

    // Call the method being tested
    result, err := component.SomeMethod("input")

    // Verify the results
    framework.Assertions.NoError(err, "method should not return an error")
    framework.Assertions.Equal("expected result", result, "result should match expected value")
}
```

### Testing with Dependencies

```
func TestComponentWithDependencies(t *testing.T) {
    // Create a unit test framework
    framework := testutil.NewUnitTestFramework(t)

    // Create mocks for dependencies
    mockDep1 := framework.Mocks.NewMockDependency1()
    mockDep2 := framework.Mocks.NewMockDependency2()

    // Configure mock behavior
    mockDep1.AddMockResponse("method1", "result1")
    mockDep2.AddMockResponse("method2", "result2")

    // Create the component under test with mock dependencies
    component := NewSomeComponent(mockDep1, mockDep2)

    // Call the method being tested
    result, err := component.MethodThatUsesDependencies()

    // Verify the results
    framework.Assertions.NoError(err, "method should not return an error")
    framework.Assertions.Equal("expected result", result, "result should match expected value")

    // Verify interactions with dependencies
    framework.Mocks.Verify(func() {
        mockDep1.VerifyCalled("method1")
        mockDep2.VerifyCalled("method2")
    })
}
```

### Table-Driven Tests

```
func TestWithMultipleCases(t *testing.T) {
    // Create a unit test framework
    framework := testutil.NewUnitTestFramework(t)

    // Define test cases
    testCases := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "normal case",
            input:    "normal input",
            expected: "expected output",
            wantErr:  false,
        },
        {
            name:     "error case",
            input:    "invalid input",
            expected: "",
            wantErr:  true,
        },
        // More test cases...
    }

    // Run test cases
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Create a new framework for each test case
            caseFramework := testutil.NewUnitTestFramework(t)

            // Create the component under test
            component := NewSomeComponent()

            // Call the method being tested
            result, err := component.SomeMethod(tc.input)

            // Verify the results
            if tc.wantErr {
                caseFramework.Assertions.Error(err, "should return an error")
            } else {
                caseFramework.Assertions.NoError(err, "should not return an error")
                caseFramework.Assertions.Equal(tc.expected, result, "result should match expected value")
            }
        })
    }
}
```

### Testing Error Handling

```
func TestErrorHandling(t *testing.T) {
    // Create a unit test framework
    framework := testutil.NewUnitTestFramework(t)

    // Create mocks for dependencies
    mockDep := framework.Mocks.NewMockDependency()

    // Configure mock to return an error
    expectedErr := errors.New("dependency error")
    mockDep.AddMockError("method", expectedErr)

    // Create the component under test with mock dependency
    component := NewSomeComponent(mockDep)

    // Call the method being tested
    result, err := component.MethodThatHandlesErrors()

    // Verify the error is handled correctly
    framework.Assertions.Error(err, "should return an error")
    framework.Assertions.ErrorIs(err, expectedErr, "should return the dependency error")
    framework.Assertions.Equal("", result, "result should be empty on error")
}
```

### Testing Concurrent Code

```
func TestConcurrentCode(t *testing.T) {
    // Create a unit test framework
    framework := testutil.NewUnitTestFramework(t)

    // Create the component under test
    component := NewConcurrentComponent()

    // Create a wait group to wait for all goroutines
    var wg sync.WaitGroup
    wg.Add(10)

    // Create a mutex to protect shared state
    var mu sync.Mutex
    results := make([]string, 0, 10)

    // Run concurrent operations
    for i := 0; i < 10; i++ {
        go func(id int) {
            defer wg.Done()

            // Call the method being tested
            result, err := component.ConcurrentMethod(id)

            // Safely store the result
            mu.Lock()
            defer mu.Unlock()

            framework.Assertions.NoError(err, "concurrent method should not return an error")
            results = append(results, result)
        }(i)
    }

    // Wait for all goroutines to complete
    wg.Wait()

    // Verify the results
    framework.Assertions.Len(results, 10, "should have 10 results")
    // Additional assertions on the results...
}
```

## Examples

### Testing a Utility Function

```
func TestFormatPath(t *testing.T) {
    // Create a unit test framework
    framework := testutil.NewUnitTestFramework(t)

    // Define test cases
    testCases := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "normal path",
            input:    "/path/to/file.txt",
            expected: "/path/to/file.txt",
        },
        {
            name:     "path with double slashes",
            input:    "/path//to/file.txt",
            expected: "/path/to/file.txt",
        },
        {
            name:     "path with trailing slash",
            input:    "/path/to/directory/",
            expected: "/path/to/directory",
        },
        {
            name:     "empty path",
            input:    "",
            expected: "",
        },
    }

    // Run test cases
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Create a new framework for each test case
            caseFramework := testutil.NewUnitTestFramework(t)

            // Call the function being tested
            result := utils.FormatPath(tc.input)

            // Verify the result
            caseFramework.Assertions.Equal(tc.expected, result, "formatted path should match expected value")
        })
    }
}
```

### Testing a Component with Dependencies

```
func TestFileManager(t *testing.T) {
    // Create a unit test framework
    framework := testutil.NewUnitTestFramework(t)

    // Create mocks for dependencies
    mockFS := framework.Mocks.NewMockFileSystemProvider()
    mockGraph := framework.Mocks.NewMockGraphProvider()

    // Configure mock behavior
    mockFS.AddMockFile("/path/to/file.txt", []byte("file content"), nil)
    mockGraph.AddMockResponse("/me/drive/root:/path/to/file.txt", &graph.DriveItem{
        ID:   "123",
        Name: "file.txt",
        File: &graph.File{
            MimeType: "text/plain",
        },
    })

    // Create the component under test with mock dependencies
    fileManager := fs.NewFileManager(mockFS, mockGraph)

    // Call the method being tested
    info, err := fileManager.GetFileInfo("/path/to/file.txt")

    // Verify the results
    framework.Assertions.NoError(err, "GetFileInfo should not return an error")
    framework.Assertions.NotNil(info, "file info should not be nil")
    framework.Assertions.Equal("file.txt", info.Name, "file name should match")
    framework.Assertions.Equal("text/plain", info.MimeType, "mime type should match")

    // Verify interactions with dependencies
    framework.Mocks.Verify(func() {
        mockFS.VerifyOperation("/path/to/file.txt", testutil.FSOperationRead)
        mockGraph.VerifyCalled("/me/drive/root:/path/to/file.txt")
    })
}
```

### Testing Error Handling in a Component

```
func TestFileManagerErrorHandling(t *testing.T) {
    // Create a unit test framework
    framework := testutil.NewUnitTestFramework(t)

    // Create mocks for dependencies
    mockFS := framework.Mocks.NewMockFileSystemProvider()
    mockGraph := framework.Mocks.NewMockGraphProvider()

    // Configure mock to return an error
    fsErr := errors.New("file not found")
    mockFS.AddMockError("/path/to/nonexistent.txt", testutil.FSOperationRead, fsErr)

    // Create the component under test with mock dependencies
    fileManager := fs.NewFileManager(mockFS, mockGraph)

    // Call the method being tested
    info, err := fileManager.GetFileInfo("/path/to/nonexistent.txt")

    // Verify the error is handled correctly
    framework.Assertions.Error(err, "GetFileInfo should return an error")
    framework.Assertions.ErrorContains(err, "file not found", "error should contain the original error message")
    framework.Assertions.Nil(info, "file info should be nil on error")

    // Verify interactions with dependencies
    framework.Mocks.Verify(func() {
        mockFS.VerifyOperation("/path/to/nonexistent.txt", testutil.FSOperationRead)
        // Graph API should not be called if the file is not found
        mockGraph.VerifyNotCalled("/me/drive/root:/path/to/nonexistent.txt")
    })
}
```

### Complete Unit Test Example

```
package mypackage_test

import (
    "errors"
    "testing"

    "github.com/auriora/onemount/internal/testutil"
)

func TestMyFunction(t *testing.T) {
    // Create a fixture
    fixture := testutil.NewUnitTestFixture("my-fixture")
        .WithSetup(func(t *testing.T) (interface{}, error) {
            // Set up test environment
            return "test-data", nil
        })
        .WithTeardown(func(t *testing.T, fixture interface{}) error {
            // Clean up test environment
            return nil
        })

    // Use the fixture
    fixture.Use(t, func(t *testing.T, fixture interface{}) {
        // Create an assert object
        assert := testutil.NewAssert(t)

        // Create a mock
        mock := testutil.NewMock(t, "my-mock")
        mock.On("method1").WithArgs("arg1").Return("result1")

        // Call the function under test
        result := MyFunction(fixture.(string), mock)

        // Verify the result
        assert.Equal("expected-result", result)

        // Verify mock expectations
        mock.Verify()
    })

    // Test with edge cases
    generator := testutil.NewEdgeCaseGenerator()
    for _, str := range generator.StringEdgeCases() {
        t.Run("Edge case: "+str, func(t *testing.T) {
            assert := testutil.NewAssert(t)
            mock := testutil.NewMock(t, "my-mock")
            mock.On("method1").WithArgs(str).Return("result1")

            result := MyFunction(str, mock)
            assert.Equal("expected-result", result)
            mock.Verify()
        })
    }

    // Test error conditions
    conditions := []*testutil.ErrorCondition{
        testutil.NewErrorCondition("Network error").
            WithFunc(func() error {
                mock := testutil.NewMock(t, "my-mock")
                mock.On("method1").WithArgs("arg1").ReturnError(errors.New("network error"))

                result := MyFunction("arg1", mock)
                if result != "error-result" {
                    return errors.New("expected error-result, got " + result)
                }
                return nil
            }),
    }
    testutil.RunErrorConditions(t, conditions)

    // Table-driven tests
    tests := []testutil.TableTest{
        {
            Name:     "Normal input",
            Input:    "normal",
            Expected: "normal-result",
        },
        {
            Name:     "Special input",
            Input:    "special",
            Expected: "special-result",
        },
    }
    testutil.RunTableTests(t, tests, func(t *testing.T, test testutil.TableTest) (interface{}, error) {
        mock := testutil.NewMock(t, "my-mock")
        mock.On("method1").WithArgs(test.Input.(string)).Return(test.Input.(string) + "-result")

        return MyFunction(test.Input.(string), mock), nil
    })
}
```

## Integration with Existing Test Framework

The unit testing framework is designed to work alongside the existing TestFramework and IntegrationTestEnvironment. It provides additional utilities specifically for unit testing, while the existing framework provides utilities for integration testing and system testing.

To use the unit testing framework with the existing framework:

```
func TestWithBothFrameworks(t *testing.T) {
    // Create a logger
    logger := log.With().Str("component", "test").Logger()

    // Create a test framework
    framework := testutil.NewTestFramework(testutil.TestConfig{
        Environment:    "test",
        Timeout:        30,
        VerboseLogging: true,
    }, &logger)

    // Add cleanup using t.Cleanup to ensure resources are cleaned up
    t.Cleanup(func() {
        framework.CleanupResources()
    })

    // Create a unit test fixture
    fixture := testutil.NewUnitTestFixture("my-fixture")
        .WithSetup(func(t *testing.T) (interface{}, error) {
            // Set up test environment
            return "test-data", nil
        })
        .WithTeardown(func(t *testing.T, fixture interface{}) error {
            // Clean up test environment
            return nil
        })

    // Use the fixture
    fixture.Use(t, func(t *testing.T, fixture interface{}) {
        // Create an assert object
        assert := testutil.NewAssert(t)

        // Run a test with the framework
        result := framework.RunTest("my-test", func(ctx context.Context) error {
            // Test logic using the framework and fixture
            return nil
        })

        // Verify the result
        assert.Equal(testutil.TestStatusPassed, result.Status)
    })
}
```

## Best Practices

### Do's

1. **Test One Thing at a Time**: Each unit test should focus on testing a single behavior or code path.
2. **Use Descriptive Test Names**: Test names should clearly describe what is being tested and the expected outcome.
3. **Use Table-Driven Tests**: For testing multiple similar cases, use table-driven tests to reduce code duplication.
4. **Mock Dependencies**: Use mocks to isolate the component under test from its dependencies.
5. **Test Edge Cases**: Include tests for edge cases, such as empty inputs, boundary values, and error conditions.
6. **Test Error Handling**: Verify that errors are handled correctly and appropriate error messages are returned.
7. **Keep Tests Fast**: Unit tests should run quickly to provide fast feedback during development.
8. **Use Assertions**: Use the assertion utilities provided by the framework to make test code more readable.
9. **Clean Up Resources**: Use the framework's resource management to ensure all resources are cleaned up after tests.
10. **Aim for High Coverage**: Strive for high code coverage with unit tests, but focus on testing behavior rather than implementation details.
11. **Use Test Fixtures**: Use test fixtures to encapsulate setup and teardown logic, making tests more readable and maintainable.
12. **Test Both Success and Failure Cases**: Test both normal operation and error handling.
13. **Keep Tests Independent**: Each test should be independent of other tests.
14. **Keep Tests Simple**: Keep tests simple and focused on a single aspect of behavior.

### Don'ts

1. **Don't Test External Systems**: Unit tests should not depend on external systems or services.
2. **Don't Test Implementation Details**: Focus on testing the public API and behavior, not internal implementation details.
3. **Don't Write Brittle Tests**: Avoid tests that break when implementation details change but behavior remains the same.
4. **Don't Overuse Mocks**: Use mocks for external dependencies, but consider using real implementations for simple internal components.
5. **Don't Test Framework Code**: Focus on testing your own code, not the behavior of the testing framework or standard library.
6. **Don't Write Flaky Tests**: Avoid tests that sometimes pass and sometimes fail without changes to the code.
7. **Don't Ignore Test Failures**: Address test failures promptly to maintain the reliability of the test suite.
8. **Don't Duplicate Assertions**: Use table-driven tests or helper functions to avoid duplicating assertion logic.
9. **Don't Write Tests That Take Too Long**: Keep unit tests fast to maintain a quick feedback loop during development.
10. **Don't Skip Writing Tests for Simple Code**: Even simple code can have bugs, so write tests for all code paths.

## Related Documentation

- Testing Framework (see code comments in `internal/testutil/framework/framework.go`)
- Mock Providers (see code comments in `internal/testutil/mock/` directory)
- [Integration Testing](integration-testing-guide.md)
- [Testing Troubleshooting](../testing-troubleshooting.md)
