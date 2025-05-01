# Unit Testing

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

The OneMount unit testing framework extends Go's standard testing package with additional utilities specifically designed for unit testing OneMount components.

### UnitTestFramework

The `UnitTestFramework` provides utilities for unit testing:

```go
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

```go
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

### AssertionHelper

The `AssertionHelper` provides utilities for making assertions in unit tests:

```go
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

### MockHelper

The `MockHelper` provides utilities for creating and managing mocks:

```go
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

### FixtureHelper

The `FixtureHelper` provides utilities for managing test fixtures:

```go
// Load a test fixture
data := framework.Fixtures.Load("testdata/fixture.json")

// Create a temporary file with content
filePath := framework.Fixtures.CreateTempFile("test-file", []byte("content"))

// Create a temporary directory
dirPath := framework.Fixtures.CreateTempDir("test-dir")

// Clean up fixtures (automatically called at the end of the test)
framework.Fixtures.Cleanup()
```

## Writing Unit Tests

### Basic Unit Test

```go
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

```go
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

```go
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

```go
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

```go
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

## Examples

### Testing a Utility Function

```go
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

```go
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

```go
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

## Related Documentation

- [Test Framework Core](test-framework-core.md)
- [Mock Providers](mock-providers.md)
- [Integration Testing](integration-testing.md)
- [Unit Testing Best Practices](unit-testing-best-practices.md)
- [Troubleshooting](troubleshooting.md)