# Test Architecture Design

## 1. Introduction

This document outlines the test architecture for the OneMount project, including the design of the test framework, mocking infrastructure, test coverage reporting, integration testing, and performance benchmarking.

### 1.1 Purpose

The purpose of this test architecture is to ensure comprehensive test coverage across all components of the OneMount system, enabling reliable verification of functionality, performance, and compatibility. This architecture supports the project's quality assurance goals by providing a structured approach to testing at all levels.

### 1.2 Scope

This test architecture applies to all components of the OneMount system, including:
- Filesystem operations
- Microsoft Graph API integration
- User interface components
- System integration with desktop environments
- Performance and reliability aspects

## 2. Test Framework Architecture

### 2.1 Overview

The OneMount test framework is designed as a layered architecture that supports different types of testing while promoting code reuse and maintainability.

```
┌─────────────────────────────────────────────────────────────┐
│                     Test Suites                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Unit Tests │  │Integration  │  │  Performance        │  │
│  │             │  │   Tests     │  │    Benchmarks       │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                     Test Utilities                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Common     │  │  FS         │  │  Graph              │  │
│  │  Utilities  │  │  Utilities  │  │  Utilities          │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
├─────────────────────────────────────────────────────────────┤
│                     Mock Infrastructure                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  │
│  │  Mock FS    │  │  Mock Graph │  │  Mock UI            │  │
│  │  Components │  │  API        │  │  Components         │  │
│  └─────────────┘  └─────────────┘  └─────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Key Components

#### 2.2.1 TestFramework Class

The `TestFramework` class serves as the central component for test configuration, setup, and execution. It provides:

- Test environment initialization
- Configuration management for different test scenarios
- Resource allocation and cleanup
- Test execution orchestration
- Reporting interfaces
- Context management for timeout/cancellation
- Structured logging

```go
// Example code for the TestFramework class
// This is a conceptual example and not meant to be compiled directly
package example

import (
    "context"
    "time"
)

// TestConfig defines configuration options for the test environment
type TestConfig struct {
    // Configuration fields would go here
}

// TestResource represents a resource that needs cleanup after tests
type TestResource interface {
    Cleanup() error
}

// MockProvider is an interface for mock components
type MockProvider interface {
    // Mock provider methods would go here
}

// CoverageReporter collects and reports test coverage
type CoverageReporter interface {
    // Coverage reporting methods would go here
}

// Logger provides structured logging for tests
type Logger interface {
    // Logging methods would go here
}

// TestStatus represents the status of a test
type TestStatus string

const (
    TestStatusPassed  TestStatus = "PASSED"
    TestStatusFailed  TestStatus = "FAILED"
    TestStatusSkipped TestStatus = "SKIPPED"
)

// TestFailure represents a test failure
type TestFailure struct {
    Message   string
    Location  string
    Expected  interface{}
    Actual    interface{}
}

// TestArtifact represents a test artifact
type TestArtifact struct {
    Name     string
    Type     string
    Location string
}

// TestResult represents the result of a test
type TestResult struct {
    Name       string
    Duration   time.Duration
    Status     TestStatus
    Failures   []TestFailure
    Artifacts  []TestArtifact
}

// TestLifecycle defines hooks for test lifecycle events
type TestLifecycle interface {
    BeforeTest(ctx context.Context) error
    AfterTest(ctx context.Context) error
    OnFailure(ctx context.Context, failure TestFailure) error
}

// TestFramework provides centralized test configuration and execution
type TestFramework struct {
    // Configuration for the test environment
    Config TestConfig

    // Test resources that need cleanup
    resources []TestResource

    // Mock providers
    mockProviders map[string]MockProvider

    // Coverage reporting
    coverageReporter CoverageReporter

    // Context for timeout/cancellation
    ctx context.Context

    // Structured logging
    logger Logger
}
```

#### 2.2.2 Mock Providers

Mock providers implement interfaces that mimic the behavior of real components:

- `MockGraphProvider`: Simulates Microsoft Graph API responses
- `MockFileSystemProvider`: Simulates filesystem operations
- `MockUIProvider`: Simulates UI interactions

#### 2.2.3 Test Utilities

Test utilities provide common functionality needed across different tests:

- `common`: General test utilities for assertions, setup, and teardown
- `fs`: Filesystem-specific test utilities
- `graph`: Microsoft Graph API test utilities
- `ui`: UI-specific test utilities

## 3. Mocking Infrastructure

### 3.1 Mock Design Principles

The mocking infrastructure follows these principles:

1. **Interface-based**: All mocks implement the same interfaces as the real components
2. **Configurable**: Mocks can be configured to return specific responses or errors
3. **Stateful**: Mocks can maintain state to simulate real-world behavior
4. **Observable**: Mocks record interactions for verification

### 3.2 Mock Components

#### 3.2.1 Graph API Mocks

```go
// Example code for the MockGraphClient
// This is a conceptual example and not meant to be compiled directly
package example

import "time"

// MockCall represents a record of a method call on a mock
type MockCall struct {
    Method string
    Args   []interface{}
    Result interface{}
}

// NetworkConditions simulates different network scenarios
type NetworkConditions struct {
    Latency    int
    PacketLoss float64
    Bandwidth  int
}

// MockRecorder records and verifies mock interactions
type MockRecorder interface {
    RecordCall(method string, args ...interface{})
    GetCalls() []MockCall
    VerifyCall(method string, times int) bool
}

// MockConfig defines configuration for mock behavior
type MockConfig struct {
    Latency        time.Duration
    ErrorRate      float64
    ResponseDelay  time.Duration
    CustomBehavior map[string]interface{}
}

// MockGraphClient implements the GraphClient interface
type MockGraphClient struct {
    // Configured responses for different API calls
    responses map[string]interface{}

    // Record of calls made to the mock
    calls []MockCall

    // Simulated network conditions
    networkConditions NetworkConditions

    // Mock recorder for verification
    recorder MockRecorder

    // Configuration for mock behavior
    config MockConfig
}
```

#### 3.2.2 Filesystem Mocks

```go
// Example code for the MockFileSystem
// This is a conceptual example and not meant to be compiled directly
package example

// MockCall represents a record of a method call on a mock
type MockCall struct {
    Method string
    Args   []interface{}
    Result interface{}
}

// MockRecorder records and verifies mock interactions
type MockRecorder interface {
    RecordCall(method string, args ...interface{})
    GetCalls() []MockCall
    VerifyCall(method string, times int) bool
}

// MockConfig defines configuration for mock behavior
type MockConfig struct {
    Latency        int64
    ErrorRate      float64
    ResponseDelay  int64
    CustomBehavior map[string]interface{}
}

// MockFile represents a file in the mock filesystem
type MockFile struct {
    Name     string
    Content  []byte
    Metadata map[string]interface{}
}

// FSOperation represents a filesystem operation
type FSOperation struct {
    Type      string
    Path      string
    Timestamp int64
}

// ErrorConditions simulates different error scenarios
type ErrorConditions struct {
    ReadErrors  map[string]error
    WriteErrors map[string]error
    ListErrors  map[string]error
}

// MockFileSystem implements the FileSystem interface
type MockFileSystem struct {
    // Virtual filesystem state
    files map[string]*MockFile

    // Record of operations
    operations []FSOperation

    // Simulated error conditions
    errorConditions ErrorConditions

    // Mock recorder for verification
    recorder MockRecorder

    // Configuration for mock behavior
    config MockConfig
}
```

#### 3.2.3 UI Component Mocks

```go
// Example code for the MockUIComponent
// This is a conceptual example and not meant to be compiled directly
package example

// MockCall represents a record of a method call on a mock
type MockCall struct {
    Method string
    Args   []interface{}
    Result interface{}
}

// MockRecorder records and verifies mock interactions
type MockRecorder interface {
    RecordCall(method string, args ...interface{})
    GetCalls() []MockCall
    VerifyCall(method string, times int) bool
}

// MockConfig defines configuration for mock behavior
type MockConfig struct {
    Latency        int64
    ErrorRate      float64
    ResponseDelay  int64
    CustomBehavior map[string]interface{}
}

// UIState represents the current state of a UI component
type UIState struct {
    Visible   bool
    Enabled   bool
    Text      string
    Properties map[string]interface{}
}

// UIEvent represents a user interaction with a UI component
type UIEvent struct {
    Type      string
    Target    string
    Timestamp int64
    Data      map[string]interface{}
}

// UIResponse represents a response to a UI interaction
type UIResponse struct {
    Success   bool
    Data      interface{}
    Error     error
}

// MockUIComponent implements UI component interfaces
type MockUIComponent struct {
    // Simulated UI state
    state UIState

    // Record of UI events
    events []UIEvent

    // Configured responses to UI interactions
    responses map[string]UIResponse

    // Mock recorder for verification
    recorder MockRecorder

    // Configuration for mock behavior
    config MockConfig
}
```

## 4. Test Coverage Reporting

### 4.1 Coverage Metrics

The test coverage framework tracks the following metrics:

1. **Line Coverage**: Percentage of code lines executed during tests
2. **Function Coverage**: Percentage of functions called during tests
3. **Branch Coverage**: Percentage of code branches executed during tests
4. **Package Coverage**: Coverage metrics aggregated by package

### 4.2 Coverage Collection

Coverage data is collected using Go's built-in coverage tools:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 4.3 Coverage Reporting

A custom coverage reporter aggregates and visualizes coverage data:

```go
// Example code for the CoverageReporter
// This is a conceptual example and not meant to be compiled directly
package example

import "time"

// PackageCoverage represents coverage data for a single package
type PackageCoverage struct {
    PackageName    string
    LineCoverage   float64
    FuncCoverage   float64
    BranchCoverage float64
    Files          map[string]FileCoverage
}

// FileCoverage represents coverage data for a single file
type FileCoverage struct {
    FileName       string
    LineCoverage   float64
    FuncCoverage   float64
    BranchCoverage float64
}

// HistoricalCoverage represents coverage data at a point in time
type HistoricalCoverage struct {
    Timestamp      int64
    TotalCoverage  float64
    PackageCoverage map[string]float64
}

// CoverageThresholds defines minimum acceptable coverage levels
type CoverageThresholds struct {
    LineCoverage   float64
    FuncCoverage   float64
    BranchCoverage float64
}

// CoverageGoal defines coverage goals for packages
type CoverageGoal struct {
    Package    string
    MinLine    float64
    MinBranch  float64
    MinFunc    float64
    Deadline   time.Time
}

// CoverageTrend represents coverage trend analysis
type CoverageTrend struct {
    Timestamp    time.Time
    TotalChange  float64
    PackageDeltas map[string]float64
    Regressions  []CoverageRegression
}

// CoverageRegression represents a coverage regression
type CoverageRegression struct {
    Package    string
    OldCoverage float64
    NewCoverage float64
    Delta      float64
}

// CoverageReporter collects and reports test coverage
type CoverageReporter struct {
    // Coverage data by package
    packageCoverage map[string]PackageCoverage

    // Historical coverage data
    historicalData []HistoricalCoverage

    // Coverage thresholds
    thresholds CoverageThresholds

    // Coverage goals
    goals []CoverageGoal

    // Coverage trends
    trends []CoverageTrend
}
```

## 5. Integration Test Framework

### 5.1 Integration Test Design

Integration tests verify the interaction between different components:

1. **Component Integration**: Tests interaction between internal components
2. **External Integration**: Tests interaction with external systems (OneDrive API)
3. **End-to-End**: Tests complete user workflows

### 5.2 Integration Test Environment

```go
// Example code for the IntegrationTestEnvironment
// This is a conceptual example and not meant to be compiled directly
package example

import "context"

// NetworkSimulator simulates different network conditions for testing
type NetworkSimulator interface {
    // Simulate network conditions
    SetConditions(latency int, packetLoss float64, bandwidth int) error

    // Simulate network disconnection
    Disconnect() error

    // Restore network connection
    Reconnect() error
}

// TestDataManager manages test data for integration tests
type TestDataManager interface {
    // Load test data
    LoadTestData(dataSet string) error

    // Clean up test data
    CleanupTestData() error

    // Get test data item
    GetTestData(key string) interface{}
}

// TestStep represents a step in a test scenario
type TestStep struct {
    Name        string
    Action      func(ctx context.Context) error
    Validation  func(ctx context.Context) error
}

// TestAssertion represents an assertion in a test scenario
type TestAssertion struct {
    Name        string
    Condition   func(ctx context.Context) bool
    Message     string
}

// CleanupStep represents a cleanup step in a test scenario
type CleanupStep struct {
    Name        string
    Action      func(ctx context.Context) error
}

// TestScenario represents a scenario-based test
type TestScenario struct {
    Name        string
    Steps       []TestStep
    Assertions  []TestAssertion
    Cleanup     []CleanupStep
}

// NetworkRule represents a network isolation rule
type NetworkRule struct {
    Source      string
    Destination string
    Allow       bool
}

// IsolationConfig defines component isolation for tests
type IsolationConfig struct {
    MockedServices []string
    NetworkRules   []NetworkRule
    DataIsolation  bool
}

// IntegrationTestEnvironment provides a controlled environment for integration tests
type IntegrationTestEnvironment struct {
    // Real or mock components based on configuration
    components map[string]interface{}

    // Network simulation
    networkSimulator NetworkSimulator

    // Test data management
    testData TestDataManager

    // Test scenarios
    scenarios []TestScenario

    // Component isolation
    isolation IsolationConfig
}
```

### 5.3 Integration Test Scenarios

Key integration test scenarios include:

1. **Authentication Flow**: Verify the complete authentication process
2. **File Operations**: Test file creation, modification, and deletion
3. **Offline Mode**: Test transition to and from offline mode
4. **Error Handling**: Verify proper handling of API errors and network issues

## 6. Performance Benchmarks

### 6.1 Benchmark Framework

The benchmark framework measures key performance indicators:

```go
// Example code for the PerformanceBenchmark
// This is a conceptual example and not meant to be compiled directly
package example

import (
    "testing"
    "time"
)

// TestScenario represents a scenario-based test
type TestScenario struct {
    Name        string
    Steps       []interface{}
    Assertions  []interface{}
    Cleanup     []interface{}
}

// PerformanceThresholds defines minimum acceptable performance levels
type PerformanceThresholds struct {
    MaxLatency      int64  // Maximum acceptable latency in milliseconds
    MinThroughput   int64  // Minimum acceptable operations per second
    MaxMemoryUsage  int64  // Maximum acceptable memory usage in MB
    MaxCPUUsage     float64 // Maximum acceptable CPU usage percentage
}

// ResourceMetrics represents resource usage metrics
type ResourceMetrics struct {
    CPUUsage    float64
    MemoryUsage int64
    DiskIO      int64
    NetworkIO   int64
}

// LoadTest defines load testing parameters
type LoadTest struct {
    Concurrency int
    Duration    time.Duration
    RampUp      time.Duration
    Scenario    TestScenario
}

// PerformanceMetrics represents performance test metrics
type PerformanceMetrics struct {
    Latencies    []time.Duration
    Throughput   float64
    ErrorRate    float64
    ResourceUsage ResourceMetrics
    Custom       map[string]float64
}

// PerformanceBenchmark defines a performance test
type PerformanceBenchmark struct {
    // Name and description
    Name string
    Description string

    // Setup and teardown functions
    Setup func() error
    Teardown func() error

    // The benchmark function
    BenchmarkFunc func(b *testing.B)

    // Performance thresholds
    thresholds PerformanceThresholds

    // Performance metrics
    metrics PerformanceMetrics

    // Load test configuration
    loadTest *LoadTest
}
```

### 6.2 Key Performance Metrics

The framework measures:

1. **Latency**: Response time for operations
2. **Throughput**: Operations per second
3. **Resource Usage**: CPU, memory, and network utilization
4. **Scalability**: Performance under increasing load

### 6.3 Benchmark Scenarios

Key benchmark scenarios include:

1. **File Download Performance**: Measure download speed and latency
2. **File Upload Performance**: Measure upload speed and reliability
3. **Metadata Operations**: Measure performance of directory listing and metadata retrieval
4. **Concurrent Operations**: Measure performance under concurrent access

## 7. Types of Testing

### 7.1 Unit Testing

Unit testing focuses on testing individual classes, methods, and functions in isolation to verify implementation according to the design.

#### Architecture for Unit Testing

- **Scope**: Individual classes, methods, and functions
- **Dependencies**: Mocked using the mock infrastructure
- **Test Data**: Small, focused test data sets
- **Execution Environment**: Local development environment

#### Key Components for Unit Testing

1. **Test Fixtures**: Predefined test data and environment setup
2. **Mock Objects**: Simulated dependencies for isolated testing
3. **Assertion Framework**: Verification of expected outcomes

#### Unit Testing Process

1. **Setup**: Prepare test environment and mock dependencies
2. **Execution**: Execute the unit under test
3. **Verification**: Verify the expected outcome
4. **Teardown**: Clean up test environment

### 7.2 Integration Testing

Integration testing focuses on testing the interaction between different components when integrated to verify implementation according to the design and architecture.

#### Architecture for Integration Testing

- **Scope**: Component interactions and interfaces
- **Dependencies**: Mix of real and mocked components
- **Test Data**: Realistic test data sets
- **Execution Environment**: Controlled test environment

#### Key Components for Integration Testing

1. **Integration Test Environment**: Controlled environment for component integration
2. **Component Wiring**: Configuration of component interactions
3. **Interface Contracts**: Verification of interface compliance

#### Integration Testing Process

1. **Setup**: Prepare integrated components and test data
2. **Execution**: Execute test scenarios involving multiple components
3. **Verification**: Verify component interactions and data flow
4. **Teardown**: Clean up test environment and data

### 7.3 System Testing

System testing focuses on testing the complete integrated system to verify the implementation of all components working as a whole according to the design and architecture.

#### Architecture for System Testing

- **Scope**: End-to-end system functionality
- **Dependencies**: Real components with minimal mocking
- **Test Data**: Production-like test data
- **Execution Environment**: Production-like environment

#### Key Components for System Testing

1. **System Test Environment**: Production-like environment for system testing
2. **End-to-End Scenarios**: Complete user workflows
3. **System Configuration**: Production-like system configuration

#### System Testing Process

1. **Setup**: Prepare system environment and test data
2. **Execution**: Execute end-to-end test scenarios
3. **Verification**: Verify system behavior and data integrity
4. **Teardown**: Clean up system environment and data

### 7.4 Performance Testing

Performance testing focuses on testing system performance under normal conditions and under load to verify meeting of performance requirements.

#### Architecture for Performance Testing

- **Scope**: System performance characteristics
- **Dependencies**: Real components with production configuration
- **Test Data**: Large, realistic test data sets
- **Execution Environment**: Production-like environment with monitoring

#### Key Components for Performance Testing

1. **Load Generation**: Tools for generating realistic load
2. **Performance Monitoring**: Collection of performance metrics
3. **Performance Analysis**: Analysis of performance data

#### Performance Testing Process

1. **Setup**: Prepare performance test environment and monitoring
2. **Baseline**: Establish performance baseline
3. **Load Testing**: Apply various load patterns
4. **Analysis**: Analyze performance metrics against requirements
5. **Reporting**: Report performance results and recommendations

### 7.5 Security Testing

Security testing focuses on identifying vulnerabilities and security issues to verify meeting of security requirements.

#### Architecture for Security Testing

- **Scope**: System security posture
- **Dependencies**: Real components with security configuration
- **Test Data**: Security test data and attack vectors
- **Execution Environment**: Isolated security testing environment

#### Key Components for Security Testing

1. **Security Scanning**: Tools for identifying vulnerabilities
2. **Penetration Testing**: Simulated attacks on the system
3. **Security Analysis**: Analysis of security findings

#### Security Testing Process

1. **Setup**: Prepare security test environment
2. **Scanning**: Scan for known vulnerabilities
3. **Penetration Testing**: Attempt to exploit vulnerabilities
4. **Analysis**: Analyze security findings
5. **Remediation**: Address identified security issues

### 7.6 Acceptance Testing

Acceptance testing focuses on validating that the system meets user requirements to verify meeting of functional requirements and use cases.

#### Architecture for Acceptance Testing

- **Scope**: User-facing functionality
- **Dependencies**: Complete system with production configuration
- **Test Data**: User-centric test data
- **Execution Environment**: Production-like environment

#### Key Components for Acceptance Testing

1. **User Scenarios**: Test cases based on user stories
2. **Acceptance Criteria**: Verification of requirement fulfillment
3. **User Feedback**: Collection of user feedback

#### Acceptance Testing Process

1. **Setup**: Prepare acceptance test environment
2. **Execution**: Execute user scenarios
3. **Verification**: Verify acceptance criteria
4. **Feedback**: Collect and analyze user feedback

## 8. Implementation Plan

### 8.1 Phase 1: Core Framework

1. Implement basic TestFramework structure
2. Set up basic mock providers
3. Implement basic coverage reporting

### 8.2 Phase 2: Mock Infrastructure

1. Implement Graph API mocks with recording
2. Implement filesystem mocks with configurable behavior
3. Add network condition simulation

### 8.3 Phase 3: Integration and Performance

1. Set up integration test environment
2. Implement scenario-based testing
3. Add basic performance benchmarking

### 8.4 Phase 4: Enhancement

1. Add advanced coverage reporting
2. Implement load testing
3. Add performance metrics collection

## 9. Best Practices

### 9.1 Test Organization

- Use table-driven tests for similar test cases
- Group related tests in test suites
- Use meaningful test names that describe the scenario

### 9.2 Mock Usage

- Only mock external dependencies
- Keep mock configurations separate from test logic
- Record and verify mock interactions

### 9.3 Coverage

- Set realistic coverage goals
- Focus on critical path coverage
- Track coverage trends over time

### 9.4 Performance Testing

- Use realistic data sets
- Include baseline measurements
- Test under various load conditions

### 9.5 Integration Testing

- Use clean test environments
- Implement proper cleanup
- Test failure scenarios

## 10. Test Sandbox Guidelines

### 10.1 Overview

The test-sandbox directory is used as the main test working folder for the OneMount project. It contains various test artifacts, including log files, mount points, test files, authentication tokens, and temporary files. This section provides guidelines for using the test-sandbox directory effectively.

### 10.2 Current Structure and Usage

The test-sandbox directory contains the following test artifacts:

1. **Log Files**: `fusefs_tests.log` - Contains logs from test runs
2. **Mount Point**: `tmp/mount` - Where the filesystem is mounted during tests
3. **Test Files**: `dmel.fa` - A large test file used for upload session tests
4. **Authentication Tokens**: `.auth_tokens.json` - Contains authentication tokens for tests
5. **Test Database**: `tmp` - Contains the test database and other temporary files
6. **Content and Thumbnails**: `tmp/test/content` and `tmp/test/thumbnails` - Directories for test content and thumbnails

The test-sandbox directory is defined in `internal/testutil/test_constants.go` with the following structure:

```
test-sandbox/                  (TestSandboxDir)
├── .auth_tokens.json          (AuthTokensPath)
├── dmel.fa                    (DmelfaDir)
├── fusefs_tests.log           (TestLogPath)
└── tmp/                       (TestSandboxTmpDir)
    ├── test/
    │   ├── content/
    │   └── thumbnails/
    └── mount/                 (TestMountPoint)
        └── onemount_tests/    (TestDir)
            └── delta/         (DeltaDir)
```

### 10.3 Recommended Structure Outside the Project

To improve test isolation and avoid cluttering the project directory, it is recommended to move the test-sandbox directory outside of the project. This can be achieved by:

1. Creating a dedicated directory for test artifacts outside the project directory
2. Updating the constants in `internal/testutil/test_constants.go` to use this external directory
3. Ensuring all tests use the constants from `testutil` rather than hardcoded paths

The recommended structure is:

```
$HOME/.onemount-tests/                  (TestSandboxDir)
├── .auth_tokens.json                   (AuthTokensPath)
├── dmel.fa                             (DmelfaDir)
├── logs/
│   └── fusefs_tests.log                (TestLogPath)
├── tmp/                                (TestSandboxTmpDir)
│   ├── test/
│   │   ├── content/
│   │   └── thumbnails/
│   └── mount/                          (TestMountPoint)
│       └── onemount_tests/             (TestDir)
│           └── delta/                  (DeltaDir)
└── graph_test_dir/                     (New directory for graph tests)
```

### 10.4 Best Practices for Test Working Folders

#### 10.4.1 Proper Usage in Tests

1. **Use Constants**: Always use the constants defined in `internal/testutil/test_constants.go` rather than hardcoded paths.
2. **Avoid Direct Manipulation**: Do not directly manipulate the test-sandbox directory in tests. Use the provided utility functions in `internal/testutil/setup.go`.
3. **Respect Directory Structure**: Maintain the directory structure defined in the constants. Do not create additional directories or files in the test-sandbox directory unless necessary.
4. **Test Isolation**: Each test should operate in its own subdirectory to avoid conflicts with other tests.
5. **Resource Limits**: Be mindful of resource usage, especially when creating large files or many small files.

#### 10.4.2 Cleanup Procedures

1. **Clean Up After Tests**: Always clean up any files or directories created during tests.
2. **Use t.Cleanup()**: Use the `t.Cleanup()` function to register cleanup functions that will be called even if tests fail.
3. **Temporary Files**: Store temporary files in the `tmp` directory, which is cleaned up between test runs.
4. **Persistent Files**: Store files that need to persist between test runs (e.g., authentication tokens) in the root of the test-sandbox directory.
5. **Unmount Before Cleanup**: Always unmount the filesystem before attempting to clean up the mount point.

#### 10.4.3 Isolation Between Tests

1. **Unique Test Directories**: Each test should use a unique directory to avoid conflicts with other tests.
2. **Parallel Tests**: When running tests in parallel, ensure they do not share resources.
3. **Clean State**: Start each test with a clean state by removing and recreating test directories.
4. **Independent Tests**: Tests should not depend on the state created by other tests.
5. **Mock Dependencies**: Use mock implementations of external dependencies to improve isolation.

#### 10.4.4 Naming Conventions

1. **Descriptive Names**: Use descriptive names for test files and directories.
2. **Test-Specific Prefixes**: Prefix test files and directories with the test name to avoid conflicts.
3. **Temporary File Suffix**: Use a `.tmp` suffix for temporary files.
4. **Test Data Files**: Store test data files in a `testdata` directory.
5. **Log Files**: Store log files in a `logs` directory with descriptive names.

#### 10.4.5 Resource Management

1. **Limit File Sizes**: Keep test files as small as possible while still being useful for testing.
2. **Clean Up Resources**: Always clean up resources after tests, especially large files.
3. **Reuse Test Files**: Reuse test files when possible instead of creating new ones.
4. **Monitor Resource Usage**: Use the profiler to monitor resource usage during tests.
5. **Limit Concurrent Operations**: Use semaphores to limit concurrent operations and prevent resource exhaustion.

### 10.5 Specific Recommendations for Test Artifacts

#### 10.5.1 fusefs_tests.log

- Move to `$HOME/.onemount-tests/logs/fusefs_tests.log`
- Implement log rotation to prevent the log file from growing too large
- Add timestamps to log entries for better debugging

#### 10.5.2 mount-point

- Move to `$HOME/.onemount-tests/tmp/mount`
- Ensure it's unmounted and cleaned up after tests
- Use a unique mount point for each test run to avoid conflicts

#### 10.5.3 dmel.fa

- Move to `$HOME/.onemount-tests/dmel.fa`
- Consider generating this file on demand instead of storing it
- Implement a mechanism to verify the file's integrity before using it

#### 10.5.4 graph_test_dir

- Create a new directory at `$HOME/.onemount-tests/graph_test_dir`
- Use this directory for graph API tests
- Implement proper cleanup procedures for this directory

#### 10.5.5 test/

- Move to `$HOME/.onemount-tests/tmp/test`
- Ensure it's cleaned up between test runs
- Use subdirectories for different types of tests (e.g., content, thumbnails)

### 10.6 Implementation Considerations

When implementing these recommendations, consider the following:

1. **Backward Compatibility**: Ensure that existing tests continue to work with the new structure.
2. **Environment Variables**: Use environment variables to allow overriding the test-sandbox location.
3. **Documentation**: Update documentation to reflect the new structure and best practices.
4. **CI/CD Integration**: Ensure that CI/CD pipelines are updated to use the new structure.
5. **Test Helpers**: Create helper functions to simplify working with the new structure.

## 11. Conclusion

This test architecture provides a comprehensive framework for ensuring the quality and reliability of the OneMount system. By implementing this architecture, the project will benefit from:

1. **Improved Code Quality**: Through comprehensive testing at all levels
2. **Faster Development**: Through reliable test automation
3. **Better Reliability**: Through systematic testing of edge cases
4. **Performance Assurance**: Through consistent performance benchmarking

The test architecture will evolve alongside the OneMount system to address new testing requirements and improve test coverage and efficiency.
