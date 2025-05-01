# Unit Testing Framework

This document provides an overview of the unit testing framework for the OneMount project, including utilities for creating test fixtures, mocking dependencies, assertion utilities, table-driven tests, and testing edge cases and error conditions.

## Overview

The unit testing framework is designed to make it easy to write comprehensive unit tests for all components of the OneMount system. It provides a set of utilities and helpers that simplify common testing tasks and promote consistent testing practices.

## Key Components

### Test Fixtures

Test fixtures provide a way to set up and tear down test environments. They encapsulate the setup and teardown logic, making tests more readable and maintainable.

```go
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

### Mock Objects

Mock objects provide a way to simulate dependencies for isolated testing. They allow you to control the behavior of dependencies and verify interactions.

```go
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

### Assertion Utilities

Assertion utilities provide a way to verify expected outcomes in tests. They provide a consistent way to check conditions and report failures.

```go
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

### Table-Driven Tests

Table-driven tests provide a way to run the same test logic with different inputs and expected outputs. They make it easy to test multiple scenarios with minimal code duplication.

```go
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

```go
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

## Best Practices

1. **Use Test Fixtures**: Use test fixtures to encapsulate setup and teardown logic, making tests more readable and maintainable.
2. **Mock Dependencies**: Use mock objects to simulate dependencies for isolated testing.
3. **Use Assertions**: Use assertion utilities for consistent verification of expected outcomes.
4. **Use Table-Driven Tests**: Use table-driven tests for testing multiple scenarios with minimal code duplication.
5. **Test Edge Cases**: Use edge case testing utilities to ensure your code handles edge cases and error conditions correctly.
6. **Clean Up Resources**: Always clean up resources after tests, even if tests fail.
7. **Test Both Success and Failure Cases**: Test both normal operation and error handling.
8. **Keep Tests Independent**: Each test should be independent of other tests.
9. **Use Descriptive Test Names**: Use descriptive names for tests that explain what is being tested.
10. **Keep Tests Simple**: Keep tests simple and focused on a single aspect of behavior.

## Example: Complete Unit Test

```go
package mypackage_test

import (
    "errors"
    "testing"

    "github.com/bcherrington/onemount/internal/testutil"
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

```go
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

## Conclusion

The unit testing framework provides a comprehensive set of utilities for writing unit tests for the OneMount project. By using these utilities, you can write tests that are more readable, maintainable, and comprehensive, ensuring the reliability and quality of the codebase.