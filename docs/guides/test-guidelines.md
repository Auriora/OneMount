# Test Guidelines for OneMount

This document outlines the best practices for writing tests in the OneMount project, based on the patterns and practices established during the test code refactoring and the test architecture design.

## Test Architecture Overview

OneMount follows a comprehensive test architecture designed to ensure code quality and reliability. For detailed information about the test architecture, refer to the [Test Architecture Design](/docs/design/test-architecture-design.md) document.

This guide covers the following key areas:
1. **Test Types and Classification** - The different types of tests used in the project
2. **Test ID Structure and Naming Conventions** - How to structure and name your tests
3. **Best Practices for Using Mocks** - How to effectively use mock components
4. **Guidelines for Achieving Good Test Coverage** - How to ensure comprehensive test coverage
5. **Best Practices for Performance Testing** - How to write effective performance tests
6. **Guidelines for Integration Testing** - How to test component interactions

### Key Components

The test architecture consists of the following key components:

1. **Test Framework**: Provides centralized test configuration, setup, and execution
2. **Mocking Infrastructure**: Simulates external dependencies and components
3. **Test Coverage Reporting**: Tracks and reports test coverage metrics
4. **Integration Test Framework**: Verifies interaction between components
5. **Performance Benchmarks**: Measures key performance indicators

### Test Directory Structure

```
internal/testutil/           # Test utilities
├── common/                  # Common test utilities
├── fs/                      # Filesystem test utilities
└── graph/                   # Graph API test utilities
```

## Test Types and Classification

The OneMount project uses several types of tests to ensure comprehensive coverage of the codebase. Each test type has a specific purpose and scope:

### Unit Tests (UT)

Unit tests focus on testing individual functions or small components in isolation. They verify that each unit of code works as expected independently of other parts of the system.

Characteristics:
- Fast execution
- No external dependencies
- Test a single function or small component
- Use mocks for dependencies

Example: `TestUT_FS_01_01_FileOperations_BasicReadWrite_SuccessfullyPreservesContent`

### Integration Tests (IT)

Integration tests verify the interaction between multiple components. They ensure that different parts of the system work together correctly.

Characteristics:
- Test multiple components together
- May involve external dependencies
- Verify component interactions
- More complex setup than unit tests

Example: `TestIT_GR_02_03_GraphAPI_NetworkDisconnection_RetriesAndSucceeds`

### System Tests (ST)

System tests evaluate the system as a whole, including the filesystem, API integration, and UI. They verify that the entire system meets the requirements.

Characteristics:
- Test the complete system
- Involve all components
- Verify end-to-end functionality
- Simulate real user scenarios

Example: `TestST_FS_01_01_FileSystem_MountAndAccess_SuccessfullyAccessesFiles`

### Performance Tests (PT)

Performance tests measure the performance characteristics of the system, such as response time, throughput, and resource usage.

Characteristics:
- Measure performance metrics
- Test under various load conditions
- Identify performance bottlenecks
- Establish performance baselines

Example: `TestPT_FS_01_01_FileDownload_LargeFiles_CompletesWithinTimeLimit`

### Load Tests (LT)

Load tests evaluate the system's behavior under expected and peak load conditions.

Characteristics:
- Test with multiple concurrent users/operations
- Verify system stability under load
- Identify resource limitations
- Ensure consistent performance

Example: `TestLT_FS_01_01_ConcurrentAccess_MultipleUsers_MaintainsPerformance`

### Scenario Tests (SC)

Scenario tests verify specific user scenarios or workflows.

Characteristics:
- Test complete user workflows
- Verify business requirements
- Simulate real-world usage patterns
- Focus on user experience

Example: `TestSC_UI_01_01_UserLogin_ValidCredentials_SuccessfullyAccessesDrive`

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

## Test ID Structure and Naming Conventions

Clear and descriptive test names help understand what is being tested and what the expected outcome is. OneMount follows a structured approach to test IDs and function naming. For a complete reference of test case definitions, see [Test Case Definitions](/docs/testing/old%20tests/test_case_definitions.md).

### Test ID Structure

The test ID structure follows this pattern:

```
<TYPE>_<COMPONENT>_<TESTNUMBER>_<SUBTESTNUMER>
```

Where:
- `<TYPE>` is the test type (2 letters):
  - UT - Unit Test
  - IT - Integration Test
  - ST - System Test
  - PT - Performance Test
  - LT - Load Test
  - SC - Scenario Test
  - UA - User Acceptance Test
  - etc.
- `<COMPONENT>` is the component being tested (2/3 letters):
  - FS - File System
  - GR - Graph
  - UI - User Interface
  - CMD - Command
  - etc.
- `<TESTNUMBER>` is a 2-digit number uniquely identifying the test
- `<SUBTESTNUMER>` is a 2-digit number uniquely identifying the sub-test or test variant

To prevent duplicate test IDs, use the Test ID Registry to check if a test ID is already in use and to register new test IDs. See the [Test ID Registry](#test-id-registry) section for more information.

### Test Function Naming Convention

Test function names follow this pattern:

```
Test<TYPE>_<COMPONENT>_<TESTNUMBER>_<SUBTESTNUMER>_<UNIT-OF-WORK>_<STATE-UNDER-TEST>_<EXPECTED-BEHAVIOR>
```

Where:
- `<TYPE>`, `<COMPONENT>`, `<TESTNUMBER>`, and `<SUBTESTNUMER>` are the same as in the test ID structure
- `<UNIT-OF-WORK>` represents a single method, a class, or multiple classes
- `<STATE-UNDER-TEST>` represents the inputs or conditions being tested
- `<EXPECTED-BEHAVIOR>` represents the output or result

Examples:
- `TestUT_FS_01_01_FileOperations_BasicReadWrite_SuccessfullyPreservesContent`
- `TestIT_GR_02_03_GraphAPI_NetworkDisconnection_RetriesAndSucceeds`
- `TestUT_UI_05_02_AccountSelector_MultipleAccounts_DisplaysAllAccounts`

This naming pattern clearly indicates:
1. The type of test (unit, integration, etc.)
2. The component being tested
3. The specific test and sub-test numbers
4. What operation is being tested
5. Under what conditions it's being tested
6. What the expected outcome is

### Legacy Naming Pattern

For older tests that haven't been migrated to the new naming convention, the format `Operation_ShouldExpectedResult` is used, for example:
- `WriteAndRead_ShouldPreserveContent`
- `AppendMultipleTimes_ShouldHaveMultipleLines`
- `SetXAttr_ShouldStoreAttributeValue`

All new tests should follow the structured naming convention described above.

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

## Mocking Infrastructure

Mocking is essential for isolating components during testing and simulating external dependencies.

### Available Mock Components

OneMount provides the following mock components:

1. **MockGraphClient**: Simulates the Microsoft Graph API with configurable responses, network condition simulation, and call recording
2. **MockFileSystem**: Simulates filesystem operations
3. **MockUIComponent**: Simulates UI interactions

### Best Practices for Using Mocks

Following these best practices will help you use mocks effectively:

1. **Only mock external dependencies**: Mock only the external dependencies that your code interacts with, not the internal implementation details. This ensures that your tests verify the behavior of your code, not the implementation of the mocks.

2. **Keep mock configurations separate from test logic**: Define mock configurations at the beginning of your test to make it clear what behavior you're testing.

3. **Record and verify mock interactions**: Use the mock recorder to verify that your code interacts with the mocked dependencies as expected. This helps ensure that your code is using the dependencies correctly.

4. **Configure mocks to simulate realistic scenarios**: Configure your mocks to simulate realistic scenarios, including error conditions, network latency, and other real-world conditions.

5. **Use consistent mock configurations**: Use consistent mock configurations across related tests to ensure that your tests are testing the same behavior.

### Using Mock Components

When testing components that interact with external systems, use the appropriate mock component to simulate the external system's behavior. This allows you to test your code in isolation and control the behavior of the external system.

For example, when testing code that interacts with the Microsoft Graph API, use the MockGraphClient to simulate API responses:

```go
// Example of using MockGraphClient in a test
mockClient := graph.NewMockGraphClient()

// Configure responses for specific resources
resource := "/me/drive/root:/test.txt"
mockItem := &graph.DriveItem{
    ID:   "item123",
    Name: "test.txt",
}
mockClient.AddMockItem(resource, mockItem)

// Simulate network conditions (latency, packet loss, bandwidth)
mockClient.SetNetworkConditions(100*time.Millisecond, 0.1, 1024) // 100ms latency, 10% packet loss, 1MB/s bandwidth

// Configure custom behavior
mockClient.SetConfig(graph.MockConfig{
    ErrorRate: 0.05, // 5% random error rate
    ResponseDelay: 50*time.Millisecond,
    CustomBehavior: map[string]interface{}{
        "retryCount": 3,
    },
})

// Use the mock client in your test
item, err := mockClient.GetItemPath("/test.txt")
require.NoError(t, err)
assert.Equal(t, "test.txt", item.Name)

// Verify calls were recorded
recorder := mockClient.GetRecorder()
calls := recorder.GetCalls()
assert.True(t, recorder.VerifyCall("GetItemPath", 1))

// Examine call details
for _, call := range calls {
    fmt.Printf("Method: %s, Args: %v, Timestamp: %v\n", call.Method, call.Args, call.Timestamp)
}
```

## Test Coverage Reporting

Test coverage reporting helps identify areas of the codebase that need additional testing.

### Coverage Metrics

OneMount tracks the following coverage metrics:

1. **Line Coverage**: Percentage of code lines executed during tests
2. **Function Coverage**: Percentage of functions called during tests
3. **Branch Coverage**: Percentage of code branches executed during tests
4. **Package Coverage**: Coverage metrics aggregated by package

### Guidelines for Achieving Good Test Coverage

Follow these guidelines to achieve good test coverage:

1. **Set realistic coverage goals**: Set realistic coverage goals for different parts of the codebase. Not all code needs the same level of coverage. Critical components should have higher coverage goals than less critical ones.

2. **Focus on critical path coverage**: Prioritize testing the critical paths through your code. These are the paths that are most frequently used or have the highest impact on the system's functionality.

3. **Track coverage trends over time**: Monitor coverage trends over time to identify areas where coverage is decreasing. This can help you catch regressions early.

4. **Test edge cases**: Ensure your tests cover edge cases, such as empty inputs, boundary conditions, and error scenarios. These are often the source of bugs.

5. **Balance coverage with test quality**: High coverage doesn't necessarily mean good tests. Focus on writing meaningful tests that verify the correct behavior of your code, not just tests that increase coverage.

6. **Use coverage reports to identify gaps**: Use coverage reports to identify areas of the code that aren't being tested. Focus your testing efforts on these areas.

7. **Include both positive and negative tests**: Test both the expected behavior (positive tests) and error handling (negative tests) to ensure your code handles all scenarios correctly.

### Running Tests with Coverage

To run tests with coverage reporting:

```bash
# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Setting Coverage Thresholds

OneMount uses coverage thresholds to ensure that test coverage remains at an acceptable level. These thresholds are defined in the CoverageThresholds struct:

```go
// CoverageThresholds defines minimum acceptable coverage levels
type CoverageThresholds struct {
    LineCoverage   float64 // Minimum acceptable line coverage percentage
    FuncCoverage   float64 // Minimum acceptable function coverage percentage
    BranchCoverage float64 // Minimum acceptable branch coverage percentage
}
```

The current thresholds are:
- Line Coverage: 80%
- Function Coverage: 90%
- Branch Coverage: 70%

These thresholds are enforced by the CI/CD pipeline, which will fail if the coverage falls below these levels.

## Integration Testing

Integration tests verify the interaction between different components of the system.

### Integration Test Types

OneMount uses the following types of integration tests:

1. **Component Integration**: Tests interaction between internal components
2. **External Integration**: Tests interaction with external systems (OneDrive API)
3. **End-to-End**: Tests complete user workflows

### Guidelines for Integration Testing

Follow these guidelines when writing integration tests:

1. **Use clean test environments**: Always start with a clean test environment to ensure that tests are isolated from each other. Use the IntegrationTestEnvironment to set up a controlled environment for your tests.

2. **Implement proper cleanup**: Always clean up resources after tests to prevent interference with other tests. Use `defer env.Cleanup()` to ensure cleanup happens even if the test fails.

3. **Test failure scenarios**: Don't just test the happy path. Test how your components handle failures, such as network errors, API errors, and invalid inputs.

4. **Use realistic test data**: Use realistic test data that represents what your system will encounter in production. This helps ensure that your tests catch real-world issues.

5. **Isolate components when necessary**: Use mocks to isolate components when testing specific interactions. This helps identify which component is causing a failure.

6. **Test component boundaries**: Focus on testing the boundaries between components, where data is passed from one component to another. This is where integration issues often occur.

7. **Use dynamic waiting**: Use dynamic waiting instead of fixed timeouts to make tests more reliable. This is especially important for asynchronous operations.

8. **Document test scenarios**: Document what each integration test is verifying to make it clear what's being tested and why.

### Writing Integration Tests

When writing integration tests, focus on testing the interaction between components rather than the internal implementation details of each component. Use the IntegrationTestEnvironment to set up a controlled environment for your tests:

```go
// Example of an integration test
func TestFileUploadIntegration(t *testing.T) {
    // Setup test environment
    env := testutil.NewIntegrationTestEnvironment()
    defer env.Cleanup()

    // Create test file
    filePath := filepath.Join(env.MountPoint, "test.txt")
    err := os.WriteFile(filePath, []byte("test content"), 0644)
    require.NoError(t, err)

    // Wait for file to be uploaded
    testutil.WaitForCondition(t, func() bool {
        // Check if file exists on remote
        return env.FileExistsOnRemote("test.txt")
    }, 10*time.Second, 100*time.Millisecond, "File was not uploaded within timeout")

    // Verify file content on remote
    content, err := env.GetRemoteFileContent("test.txt")
    require.NoError(t, err)
    assert.Equal(t, "test content", string(content))
}
```

### Testing Failure Scenarios

It's important to test how your components handle failures. Here's an example of testing a failure scenario:

```go
// Example of testing a failure scenario
func TestFileUploadFailure(t *testing.T) {
    // Setup test environment with network simulation
    env := testutil.NewIntegrationTestEnvironment()
    defer env.Cleanup()

    // Configure network simulator to simulate disconnection
    env.NetworkSimulator.SetConditions(500, 0.5, 1024) // 500ms latency, 50% packet loss, 1MB/s bandwidth

    // Create test file
    filePath := filepath.Join(env.MountPoint, "test.txt")
    err := os.WriteFile(filePath, []byte("test content"), 0644)
    require.NoError(t, err)

    // Simulate network disconnection
    env.NetworkSimulator.Disconnect()

    // Verify file is marked as pending upload
    status, err := env.GetFileStatus(filePath)
    require.NoError(t, err)
    assert.Equal(t, "pending_upload", status)

    // Restore network connection
    env.NetworkSimulator.Reconnect()

    // Wait for file to be uploaded
    testutil.WaitForCondition(t, func() bool {
        return env.FileExistsOnRemote("test.txt")
    }, 10*time.Second, 100*time.Millisecond, "File was not uploaded after network reconnection")

    // Verify file content on remote
    content, err := env.GetRemoteFileContent("test.txt")
    require.NoError(t, err)
    assert.Equal(t, "test content", string(content))
}
```

## Performance Benchmarking

Performance benchmarks measure the performance of critical operations.

### Key Performance Metrics

OneMount measures the following performance metrics:

1. **Latency**: Response time for operations
2. **Throughput**: Operations per second
3. **Resource Usage**: CPU, memory, and network utilization
4. **Scalability**: Performance under increasing load

### Best Practices for Performance Testing

Follow these best practices when writing performance tests:

1. **Use realistic data sets**: Use data sets that are representative of real-world usage. This includes file sizes, file types, and directory structures that users are likely to encounter.

2. **Include baseline measurements**: Always include baseline measurements to provide context for your performance results. This helps identify performance regressions over time.

3. **Test under various load conditions**: Test performance under different load conditions, including normal load, peak load, and stress conditions. This helps identify how the system behaves under different scenarios.

4. **Measure resource usage**: Monitor resource usage (CPU, memory, disk I/O, network I/O) during performance tests to identify resource bottlenecks.

5. **Test on representative hardware**: Run performance tests on hardware that is representative of the target environment. This helps ensure that performance results are relevant to real-world usage.

6. **Automate performance tests**: Automate performance tests to run regularly as part of the CI/CD pipeline. This helps catch performance regressions early.

7. **Set performance thresholds**: Define performance thresholds for critical operations. If performance falls below these thresholds, the tests should fail.

8. **Test with realistic network conditions**: For operations that involve network communication, test with realistic network conditions, including latency, packet loss, and bandwidth limitations.

### Writing Performance Benchmarks

Use Go's built-in benchmarking framework to write performance benchmarks:

```go
// Example of a performance benchmark
func BenchmarkFileDownload(b *testing.B) {
    // Setup benchmark environment
    env := testutil.NewBenchmarkEnvironment()
    defer env.Cleanup()

    // Create test file on remote
    env.CreateRemoteFile("benchmark.txt", generateTestData(1024*1024)) // 1MB file

    // Run benchmark
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        filePath := filepath.Join(env.MountPoint, "benchmark.txt")
        _, err := os.ReadFile(filePath)
        require.NoError(b, err)
    }
}
```

### Testing Under Different Load Conditions

It's important to test performance under different load conditions. Here's an example of a load test:

```go
// Example of a load test
func TestFileDownloadUnderLoad(t *testing.T) {
    // Setup benchmark environment
    env := testutil.NewBenchmarkEnvironment()
    defer env.Cleanup()

    // Create test files on remote
    for i := 0; i < 100; i++ {
        fileName := fmt.Sprintf("benchmark_%d.txt", i)
        env.CreateRemoteFile(fileName, generateTestData(1024*1024)) // 1MB files
    }

    // Configure load test
    loadTest := &testutil.LoadTest{
        Concurrency: 10,                // 10 concurrent operations
        Duration:    30 * time.Second,  // Run for 30 seconds
        RampUp:      5 * time.Second,   // Ramp up over 5 seconds
        Scenario: func(ctx context.Context) error {
            // Select a random file
            fileIndex := rand.Intn(100)
            fileName := fmt.Sprintf("benchmark_%d.txt", fileIndex)
            filePath := filepath.Join(env.MountPoint, fileName)

            // Read the file
            _, err := os.ReadFile(filePath)
            return err
        },
    }

    // Run load test
    metrics, err := env.RunLoadTest(loadTest)
    require.NoError(t, err)

    // Verify performance metrics
    assert.Less(t, metrics.P95Latency, 500*time.Millisecond, "95th percentile latency should be less than 500ms")
    assert.Greater(t, metrics.Throughput, 20.0, "Throughput should be at least 20 operations per second")
    assert.Less(t, metrics.ErrorRate, 0.01, "Error rate should be less than 1%")
    assert.Less(t, metrics.ResourceUsage.CPUUsage, 80.0, "CPU usage should be less than 80%")
    assert.Less(t, metrics.ResourceUsage.MemoryUsage, 500*1024*1024, "Memory usage should be less than 500MB")
}
```

## Test ID Registry

The OneMount project uses a centralized registry for test IDs to prevent duplicate test IDs. The registry is implemented as a Python script (`test_id_registry.py`) that provides the following functionality:

1. **Check if a test ID is already in use**: Before creating a new test, check if the test ID is already in use.
2. **Get the next available test number**: Get the next available test number for a given test type, module, and feature number.
3. **Register new test IDs**: Register a new test ID before implementing the test to ensure it's reserved for your use.
4. **List all test IDs**: List all test IDs in the registry, optionally filtered by test type, module, and feature number.

### Using the Test ID Registry

#### Updating the Registry

Before using the registry, you should update it to ensure it contains all the latest test IDs:

```bash
./test_id_registry.py update
```

#### Checking if a Test ID is Already in Use

To check if a test ID is already in use:

```bash
./test_id_registry.py check UT_FS_01_01
```

#### Getting the Next Available Test Number

To get the next available test number for a given test type, module, and feature number:

```bash
./test_id_registry.py next UT FS 01
```

#### Registering a New Test ID

To register a new test ID:

```bash
./test_id_registry.py register UT FS 01 FileOperations_BasicReadWrite SuccessfullyPreservesContent
```

#### Listing Test IDs

To list all test IDs in the registry:

```bash
./test_id_registry.py list
```

For more information, see the [Test ID Registry README](../../test_id_registry_README.md).

## Development Scripts and Output

### Script Locations

OneMount uses various scripts to automate development and testing tasks. These scripts should be organized as follows:

1. **Developer Scripts**: All developer scripts should be placed in the `scripts/developer` directory. This includes scripts for:
   - Test automation
   - Development environment setup
   - Code generation
   - Performance analysis
   - Debugging utilities

2. **Script Output**: All script output should be directed to the `tmp/` directory. This includes:
   - Log files
   - Generated reports
   - Temporary data files
   - Debug output
   - Performance metrics

### Best Practices for Scripts

1. **Documentation**: Each script should include a header comment that explains its purpose, usage, and any required parameters.
2. **Error Handling**: Scripts should include proper error handling and provide meaningful error messages.
3. **Cleanup**: Scripts should clean up temporary files when they complete, especially if they fail.
4. **Logging**: Scripts should log their actions to make debugging easier.

## Documentation Guidelines

When working on the OneMount project, follow these documentation guidelines:

1. **Update Existing Documentation**: Whenever possible, update existing documentation rather than creating new documents. This helps maintain a cohesive documentation set and prevents fragmentation.

2. **New Documentation Location**: If new documentation is needed, place it in the appropriate subdirectory of the `docs/` folder:
   - `docs/guides/` - For developer guides and best practices
   - `docs/design/` - For design documentation
   - `docs/implementation/` - For implementation details
   - `docs/requirements/` - For project requirements
   - `docs/testing/` - For testing documentation

3. **Link to New Documentation**: When creating new documentation, add links to it from relevant existing documentation to ensure discoverability.

4. **Check Existing Documentation**: Before starting work on a task, always check the existing documentation in the `docs/` directory for relevant information.

## Using Junie AI Prompts

Junie AI prompts are used throughout the OneMount project to provide guidance for implementing features, fixing bugs, and improving code quality. These prompts are structured instructions that can be used with AI assistants to generate code, documentation, or guidance.

### Junie Prompt Structure

A typical Junie prompt follows this structure:

```
Create/Implement/Enhance [component] based on [reference document]. The implementation should:
1. [Requirement 1]
2. [Requirement 2]
3. [Requirement 3]
...

[Additional context or examples]
```

### When to Use Junie Prompts

Include Junie prompts in your documentation when:

1. **Implementing New Features**: Provide prompts that describe how to implement the feature.
2. **Fixing Bugs**: Provide prompts that describe how to fix the bug.
3. **Refactoring Code**: Provide prompts that describe how to refactor the code.
4. **Writing Tests**: Provide prompts that describe how to write tests for a component.
5. **Documenting Code**: Provide prompts that describe how to document a component.

### Example Junie Prompt for Test Implementation

```
Implement a unit test for the FileUploadManager.UploadFile method. The test should:
1. Create a mock GraphClient that simulates successful and failed uploads
2. Test the retry mechanism by simulating temporary network failures
3. Verify that the file status is updated correctly during the upload process
4. Test edge cases such as empty files, large files, and files with special characters
5. Ensure proper cleanup of resources after the test

Reference the test-guidelines.md document for best practices on writing unit tests.
```

### Finding Existing Junie Prompts

Existing Junie prompts can be found in the implementation plan documents in the `docs/implementation/` directory. These prompts can be adapted for similar tasks or used as templates for new prompts.

## Conclusion

Following these best practices will help ensure that tests in the OneMount project are:
1. Reliable - Tests produce consistent results
2. Maintainable - Tests are easy to understand and modify
3. Efficient - Tests run quickly and don't waste resources
4. Effective - Tests catch bugs and verify functionality

Additionally, following the guidelines for scripts, documentation, and Junie prompts will help maintain a consistent and high-quality codebase and documentation set.
