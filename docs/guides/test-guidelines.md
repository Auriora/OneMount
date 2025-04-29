# Test Guidelines for onedriver

This document outlines the best practices for writing tests in the onedriver project, based on the patterns and practices established during the test code refactoring.

## Table-Driven Tests

Table-driven tests are a powerful pattern for testing multiple scenarios with similar logic. They help reduce code duplication and make it easier to add new test cases.

### When to Use Table-Driven Tests

- When testing multiple scenarios with similar logic
- When testing different inputs with the same expected behavior
- When testing edge cases of the same function

### Example

```go
func TestFileOperations(t *testing.T) {
    // Define test cases
    testCases := []struct {
        name        string
        operation   string
        content     string
        iterations  int
        fileMode    int
        verifyFunc  func(t *testing.T, filePath string, content string, iterations int)
    }{
        {
            name:       "WriteAndRead_ShouldPreserveContent",
            operation:  "write",
            content:    "my hands are typing words\n",
            iterations: 1,
            fileMode:   os.O_CREATE|os.O_RDWR,
            verifyFunc: func(t *testing.T, filePath string, content string, iterations int) {
                read, err := os.ReadFile(filePath)
                require.NoError(t, err, "Failed to read file")
                assert.Equal(t, content, string(read), "File content was not correct")
            },
        },
        // More test cases...
    }

    // Run each test case
    for _, tc := range testCases {
        tc := tc // Capture range variable for parallel execution
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel() // Run subtests in parallel
            
            // Test-specific setup
            
            // Perform the operation
            
            // Verify the results
            tc.verifyFunc(t, filePath, tc.content, tc.iterations)
        })
    }
}
```

## Test Naming Conventions

Clear and descriptive test names help understand what is being tested and what the expected outcome is.

### Naming Pattern

Use the format `Operation_ShouldExpectedResult` for test names, for example:
- `WriteAndRead_ShouldPreserveContent`
- `AppendMultipleTimes_ShouldHaveMultipleLines`
- `SetXAttr_ShouldStoreAttributeValue`

This naming pattern clearly indicates:
1. What operation is being tested
2. What the expected outcome is

## Parallel Test Execution

Running tests in parallel can significantly reduce test execution time. However, tests must be designed to run independently.

### When to Use Parallel Execution

- When tests don't share mutable state
- When tests use unique resources (e.g., different files)
- When tests don't depend on the order of execution

### Example

```go
func TestSomething(t *testing.T) {
    t.Parallel() // Mark the test for parallel execution
    
    // Test logic...
}

// For subtests
t.Run("SubtestName", func(t *testing.T) {
    t.Parallel() // Mark the subtest for parallel execution
    
    // Subtest logic...
})
```

### When Not to Use Parallel Execution

- When tests share mutable state
- When tests depend on the order of execution
- When tests use the same resources (e.g., the same file)

In these cases, document why parallel execution is not used:

```go
// Cannot use t.Parallel() because this test modifies global state
```

## Proper Cleanup

Proper cleanup ensures that tests don't leave behind artifacts that could affect other tests.

### Using t.Cleanup()

The `t.Cleanup()` function registers a function to be called when the test and all its subtests complete. This is the preferred way to clean up resources.

```go
func TestWithCleanup(t *testing.T) {
    // Create a resource
    filePath := filepath.Join(TestDir, "test_file.txt")
    err := os.WriteFile(filePath, []byte("test content"), 0644)
    require.NoError(t, err, "Failed to create test file")
    
    // Register cleanup function
    t.Cleanup(func() {
        if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
            t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
        }
    })
    
    // Test logic...
}
```

### Cleanup Best Practices

1. Always clean up resources created during tests
2. Use `t.Cleanup()` instead of `defer` for better error reporting
3. Check for errors during cleanup and log them
4. Handle the case where the resource might not exist (e.g., `os.IsNotExist(err)`)
5. For table-driven tests, ensure each subtest cleans up its own resources

## Error Handling and Assertions

Proper error handling and assertions make tests more reliable and easier to debug.

### Using require vs. assert

- Use `require` for assertions that should terminate the test on failure
- Use `assert` for assertions that should not terminate the test

```go
// Use require for critical assertions
require.NoError(t, err, "Failed to create test file")
require.NotNil(t, result, "Result should not be nil")

// Use assert for non-critical assertions
assert.Equal(t, expected, actual, "Values should be equal")
assert.True(t, condition, "Condition should be true")
```

### Descriptive Error Messages

Always provide descriptive error messages that include:
1. What was expected
2. What was actually received
3. Context about what's being tested

```go
require.Equal(t, expected, actual, 
    "Value does not match expected. Got %v, expected %v", actual, expected)
```

## Dynamic Waiting

Replace fixed timeouts and sleeps with dynamic waiting to make tests more reliable.

### Using WaitForCondition

```go
testutil.WaitForCondition(t, func() bool {
    // Return true when the condition is met
    _, err := os.Stat(filePath)
    return err == nil
}, 5*time.Second, 100*time.Millisecond, "File was not created within timeout")
```

### When to Use Dynamic Waiting

- When waiting for asynchronous operations to complete
- When testing operations that might take variable time
- When testing operations that depend on external systems

## Test Isolation

Tests should be isolated from each other to prevent interference.

### Isolation Best Practices

1. Use unique resources for each test (e.g., different file names)
2. Clean up resources after tests
3. Don't rely on the state created by other tests
4. Use subtests to group related tests
5. Use parallel execution when possible

## Conclusion

Following these best practices will help ensure that tests in the onedriver project are:
1. Reliable - Tests produce consistent results
2. Maintainable - Tests are easy to understand and modify
3. Efficient - Tests run quickly and don't waste resources
4. Effective - Tests catch bugs and verify functionality