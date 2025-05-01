# Unit Test Examples

This file contains examples of unit tests using the OneMount test framework. Unit tests focus on testing individual components in isolation, verifying that they behave as expected.

## Table of Contents

1. [Basic Unit Test](#basic-unit-test)
2. [Table-Driven Tests](#table-driven-tests)
3. [Tests with Mocks](#tests-with-mocks)
4. [Testing Error Handling](#testing-error-handling)
5. [Testing Edge Cases](#testing-edge-cases)

## Basic Unit Test

Here's a basic example of a unit test for a simple function:

```go
package util_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/onemount/internal/util"
)

// TestFormatPath tests the FormatPath function
func TestFormatPath(t *testing.T) {
    // Test case
    path := "/path/with//double/slashes/"
    expected := "/path/with/double/slashes"
    
    // Call the function
    result := util.FormatPath(path)
    
    // Verify the result
    assert.Equal(t, expected, result, "FormatPath should normalize double slashes and remove trailing slash")
}
```

This test:
1. Sets up the test input and expected output
2. Calls the function being tested
3. Verifies that the result matches the expected output

## Table-Driven Tests

Table-driven tests are useful when you want to test multiple scenarios with the same logic:

```go
package util_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/onemount/internal/util"
)

// TestFormatPath_TableDriven tests the FormatPath function with multiple scenarios
func TestFormatPath_TableDriven(t *testing.T) {
    // Define test cases
    testCases := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "NormalPath_ShouldRemainUnchanged",
            input:    "/normal/path",
            expected: "/normal/path",
        },
        {
            name:     "DoubleSlashes_ShouldBeNormalized",
            input:    "/path//with/double//slashes",
            expected: "/path/with/double/slashes",
        },
        {
            name:     "TrailingSlash_ShouldBeRemoved",
            input:    "/path/with/trailing/slash/",
            expected: "/path/with/trailing/slash",
        },
        {
            name:     "EmptyPath_ShouldReturnRoot",
            input:    "",
            expected: "/",
        },
        {
            name:     "RootPath_ShouldRemainRoot",
            input:    "/",
            expected: "/",
        },
    }
    
    // Run each test case
    for _, tc := range testCases {
        tc := tc // Capture range variable for parallel execution
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel
            
            // Call the function
            result := util.FormatPath(tc.input)
            
            // Verify the result
            assert.Equal(t, tc.expected, result, "FormatPath did not produce expected result for %s", tc.input)
        })
    }
}
```

This test:
1. Defines a table of test cases, each with a name, input, and expected output
2. Runs each test case as a subtest
3. Uses `t.Parallel()` to run the subtests in parallel
4. Verifies that the result matches the expected output for each test case

## Tests with Mocks

Here's an example of a unit test that uses mocks to isolate the component being tested:

```go
package fs_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/fs/graph"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestFileManager_GetFile tests the GetFile method of the FileManager
func TestFileManager_GetFile(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a mock graph client
    mockGraph := testutil.NewMockGraphProvider()
    
    // Configure the mock
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
    
    // Create the file manager with the mock
    fileManager := fs.NewFileManager(mockGraph, logger)
    
    // Call the method being tested
    file, err := fileManager.GetFile("/test.txt")
    
    // Verify the result
    require.NoError(t, err, "GetFile should not return an error")
    assert.Equal(t, "test.txt", file.Name, "File name should match")
    assert.Equal(t, int64(1024), file.Size, "File size should match")
    
    // Verify the mock was called correctly
    recorder := mockGraph.GetRecorder()
    assert.True(t, recorder.VerifyCall("GetItemPath", 1), "GetItemPath should be called once")
}
```

This test:
1. Creates a mock graph client
2. Configures the mock to return specific responses
3. Creates the component being tested with the mock
4. Calls the method being tested
5. Verifies that the result is correct
6. Verifies that the mock was called correctly

## Testing Error Handling

Here's an example of a unit test that verifies error handling:

```go
package fs_test

import (
    "errors"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/onemount/internal/fs"
    "github.com/yourusername/onemount/internal/fs/graph"
    "github.com/yourusername/onemount/internal/testutil"
)

// TestFileManager_GetFile_Error tests error handling in the GetFile method
func TestFileManager_GetFile_Error(t *testing.T) {
    // Create a logger
    logger := testutil.NewTestLogger()
    
    // Create a mock graph client
    mockGraph := testutil.NewMockGraphProvider()
    
    // Configure the mock to return an error
    resource := "/me/drive/root:/nonexistent.txt"
    mockGraph.AddErrorResponse(resource, graph.NewGraphError(404, "Not Found"))
    
    // Create the file manager with the mock
    fileManager := fs.NewFileManager(mockGraph, logger)
    
    // Call the method being tested
    file, err := fileManager.GetFile("/nonexistent.txt")
    
    // Verify the error
    assert.Error(t, err, "GetFile should return an error for nonexistent file")
    assert.Nil(t, file, "File should be nil when an error occurs")
    
    // Verify the error type
    var graphErr *graph.GraphError
    assert.True(t, errors.As(err, &graphErr), "Error should be a GraphError")
    assert.Equal(t, 404, graphErr.StatusCode, "Error should have status code 404")
    
    // Verify the mock was called correctly
    recorder := mockGraph.GetRecorder()
    assert.True(t, recorder.VerifyCall("GetItemPath", 1), "GetItemPath should be called once")
}
```

This test:
1. Configures a mock to return an error
2. Calls the method being tested
3. Verifies that an error is returned
4. Verifies the type and properties of the error
5. Verifies that the mock was called correctly

## Testing Edge Cases

Here's an example of a unit test that verifies behavior for edge cases:

```go
package util_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/yourusername/onemount/internal/util"
)

// TestParseSize_EdgeCases tests the ParseSize function with edge cases
func TestParseSize_EdgeCases(t *testing.T) {
    // Test cases
    testCases := []struct {
        name        string
        input       string
        expected    int64
        expectError bool
    }{
        {
            name:        "EmptyString_ShouldReturnError",
            input:       "",
            expected:    0,
            expectError: true,
        },
        {
            name:        "InvalidFormat_ShouldReturnError",
            input:       "not a size",
            expected:    0,
            expectError: true,
        },
        {
            name:        "NegativeNumber_ShouldReturnError",
            input:       "-10MB",
            expected:    0,
            expectError: true,
        },
        {
            name:        "ZeroSize_ShouldReturnZero",
            input:       "0",
            expected:    0,
            expectError: false,
        },
        {
            name:        "VeryLargeSize_ShouldParse",
            input:       "9999TB",
            expected:    9999 * 1024 * 1024 * 1024 * 1024,
            expectError: false,
        },
        {
            name:        "CaseInsensitive_ShouldParse",
            input:       "10mb",
            expected:    10 * 1024 * 1024,
            expectError: false,
        },
    }
    
    // Run each test case
    for _, tc := range testCases {
        tc := tc // Capture range variable for parallel execution
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel
            
            // Call the function
            result, err := util.ParseSize(tc.input)
            
            // Verify the result
            if tc.expectError {
                assert.Error(t, err, "ParseSize should return an error for %s", tc.input)
                assert.Equal(t, tc.expected, result, "ParseSize should return expected value even on error")
            } else {
                assert.NoError(t, err, "ParseSize should not return an error for %s", tc.input)
                assert.Equal(t, tc.expected, result, "ParseSize should return expected value")
            }
        })
    }
}
```

This test:
1. Defines test cases for various edge cases
2. Runs each test case as a subtest
3. Verifies that the function behaves correctly for each edge case
4. Checks both the result and any errors returned

## Best Practices for Unit Tests

1. **Test one thing at a time**: Each test should focus on a single aspect of the component's behavior.
2. **Use descriptive test names**: Test names should describe what is being tested and what the expected outcome is.
3. **Use table-driven tests**: When testing multiple scenarios with the same logic, use table-driven tests to reduce duplication.
4. **Isolate components**: Use mocks to isolate the component being tested from its dependencies.
5. **Test error handling**: Verify that the component handles errors correctly.
6. **Test edge cases**: Verify that the component behaves correctly for edge cases.
7. **Use parallel tests**: When possible, use `t.Parallel()` to run tests in parallel.
8. **Clean up resources**: Use `t.Cleanup()` to ensure resources are cleaned up after the test.
9. **Use assertions appropriately**: Use `assert` for non-critical assertions and `require` for critical assertions.
10. **Keep tests simple**: Tests should be easy to understand and maintain.