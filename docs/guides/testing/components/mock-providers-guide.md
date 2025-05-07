# Mock Providers Documentation

## Overview

Mock providers are a crucial component of the OneMount test framework that allow simulating external dependencies for isolated testing. They provide controlled implementations of system components, enabling tests to run without relying on actual external services or components.

The OneMount test framework includes several mock providers for different aspects of the system:

1. **MockGraphProvider**: Simulates Microsoft Graph API responses
2. **MockFileSystemProvider**: Simulates filesystem operations
3. **MockUIProvider**: Simulates UI interactions

## Key Concepts

### Mock Providers

Mock providers are controlled implementations of system components that allow tests to run without relying on actual external services or components. They are essential for:

1. **Isolation**: Testing components in isolation from external dependencies
2. **Controlled Testing**: Creating predictable test environments with predefined responses
3. **Simulation**: Simulating various scenarios, including error conditions
4. **Verification**: Verifying that components interact correctly with external services

### Common Interface

All mock providers implement the `MockProvider` interface:

```go
type MockProvider interface {
    // Setup initializes the mock provider.
    Setup() error
    // Teardown cleans up the mock provider.
    Teardown() error
    // Reset resets the mock provider to its initial state.
    Reset() error
}
```

This common interface ensures that all mock providers can be managed consistently by the test framework.

### Mock Provider Types

The OneMount test framework includes several specialized mock providers:

1. **MockGraphProvider**: Simulates Microsoft Graph API responses
2. **MockFileSystemProvider**: Simulates filesystem operations
3. **MockUIProvider**: Simulates UI interactions

## Components

### MockGraphProvider

The `MockGraphProvider` simulates the Microsoft Graph API, allowing tests to run without making actual network requests to Microsoft's services.

### Features

- Simulate Graph API responses for various endpoints
- Record API calls for verification
- Configure response delays and errors
- Provide predefined responses for common scenarios
- Support for simulating different authentication states

### API Reference

#### Types

```go
type MockGraphProvider struct {
    // Configuration for the mock provider
    Config MockGraphConfig
    // Recorded API calls
    Calls []GraphAPICall
    // Predefined responses
    Responses map[string]interface{}
    // Current authentication state
    AuthState AuthenticationState
    // Network simulator for simulating network conditions
    NetworkSim NetworkSimulator
    // Mutex for thread safety
    mu sync.Mutex
}

type GraphAPICall struct {
    // Endpoint that was called
    Endpoint string
    // HTTP method used
    Method string
    // Request body
    Body []byte
    // Query parameters
    Params map[string]string
    // Headers
    Headers map[string]string
    // Timestamp of the call
    Timestamp time.Time
}

type MockGraphConfig struct {
    // Whether to simulate authentication
    SimulateAuth bool
    // Default response delay
    DefaultDelay time.Duration
    // Whether to record API calls
    RecordCalls bool
    // Default error rate (0.0 to 1.0)
    ErrorRate float64
}

type AuthenticationState string

const (
    // AuthStateNone indicates no authentication
    AuthStateNone AuthenticationState = "NONE"
    // AuthStateValid indicates valid authentication
    AuthStateValid AuthenticationState = "VALID"
    // AuthStateExpired indicates expired authentication
    AuthStateExpired AuthenticationState = "EXPIRED"
    // AuthStateInvalid indicates invalid authentication
    AuthStateInvalid AuthenticationState = "INVALID"
)
```

#### Functions

```go
func NewMockGraphProvider() *MockGraphProvider
```

Creates a new `MockGraphProvider` with default settings.

#### Methods

```go
func (m *MockGraphProvider) Setup() error
```

Initializes the mock provider.

```go
func (m *MockGraphProvider) Teardown() error
```

Cleans up the mock provider.

```go
func (m *MockGraphProvider) Reset() error
```

Resets the mock provider to its initial state.

```go
func (m *MockGraphProvider) SetAuthState(state AuthenticationState)
```

Sets the authentication state.

```go
func (m *MockGraphProvider) AddMockResponse(endpoint string, response interface{})
```

Adds a mock response for a specific endpoint.

```go
func (m *MockGraphProvider) AddMockError(endpoint string, err error)
```

Adds a mock error for a specific endpoint.

```go
func (m *MockGraphProvider) SetResponseDelay(endpoint string, delay time.Duration)
```

Sets the response delay for a specific endpoint.

```go
func (m *MockGraphProvider) GetCalls(endpoint string) []GraphAPICall
```

Gets all recorded calls to a specific endpoint.

```go
func (m *MockGraphProvider) VerifyCalled(endpoint string) bool
```

Verifies that a specific endpoint was called.

```go
func (m *MockGraphProvider) VerifyCalledWithParams(endpoint string, params map[string]string) bool
```

Verifies that a specific endpoint was called with specific parameters.

### Usage Examples

```go
// Create a mock Graph provider
mockGraph := testutil.NewMockGraphProvider()

// Set up the provider
err := mockGraph.Setup()
if err != nil {
    // Handle error
}

// Add a mock response for a specific endpoint
mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
    ID:   "root",
    Name: "root",
    // ...
})

// Add a mock response for a specific file
mockGraph.AddMockResponse("/me/drive/items/123", &graph.DriveItem{
    ID:   "123",
    Name: "example.txt",
    File: &graph.File{
        MimeType: "text/plain",
    },
    // ...
})

// Add a mock error for a specific endpoint
mockGraph.AddMockError("/me/drive/items/456", errors.New("item not found"))

// Set the authentication state
mockGraph.SetAuthState(testutil.AuthStateValid)

// Set a response delay for a specific endpoint
mockGraph.SetResponseDelay("/me/drive/root/children", 500*time.Millisecond)

// Use the mock provider in a test
result := framework.RunTest("graph-test", func(ctx context.Context) error {
    // Get the mock provider from the framework
    provider, exists := framework.GetMockProvider("graph")
    if !exists {
        return errors.New("graph provider not found")
    }

    mockGraph := provider.(*testutil.MockGraphProvider)

    // Use the mock provider
    // ...

    // Verify that a specific endpoint was called
    if !mockGraph.VerifyCalled("/me/drive/root") {
        return errors.New("expected /me/drive/root to be called")
    }

    // Verify that a specific endpoint was called with specific parameters
    params := map[string]string{
        "$select": "id,name,size",
        "$expand": "children",
    }
    if !mockGraph.VerifyCalledWithParams("/me/drive/root", params) {
        return errors.New("expected /me/drive/root to be called with specific parameters")
    }

    return nil
})
```

### MockFileSystemProvider

The `MockFileSystemProvider` simulates filesystem operations, allowing tests to run without making actual changes to the filesystem.

### Features

- Simulate filesystem operations (create, read, update, delete)
- Record filesystem operations for verification
- Configure operation delays and errors
- Provide a virtual filesystem for testing
- Support for simulating different filesystem states

### API Reference

#### Types

```go
type MockFileSystemProvider struct {
    // Configuration for the mock provider
    Config MockFSConfig
    // Recorded filesystem operations
    Operations []FSOperation
    // Virtual filesystem state
    Files map[string]*MockFile
    // Network simulator for simulating network conditions
    NetworkSim NetworkSimulator
    // Mutex for thread safety
    mu sync.Mutex
}

type FSOperation struct {
    // Type of operation
    Type FSOperationType
    // Path of the file or directory
    Path string
    // Data involved in the operation
    Data []byte
    // Timestamp of the operation
    Timestamp time.Time
}

type FSOperationType string

const (
    // FSOperationCreate indicates a file creation operation
    FSOperationCreate FSOperationType = "CREATE"
    // FSOperationRead indicates a file read operation
    FSOperationRead FSOperationType = "READ"
    // FSOperationUpdate indicates a file update operation
    FSOperationUpdate FSOperationType = "UPDATE"
    // FSOperationDelete indicates a file deletion operation
    FSOperationDelete FSOperationType = "DELETE"
    // FSOperationList indicates a directory listing operation
    FSOperationList FSOperationType = "LIST"
)

type MockFSConfig struct {
    // Default operation delay
    DefaultDelay time.Duration
    // Whether to record operations
    RecordOperations bool
    // Default error rate (0.0 to 1.0)
    ErrorRate float64
}

type MockFile struct {
    // Name of the file
    Name string
    // Content of the file
    Content []byte
    // Whether it's a directory
    IsDir bool
    // Children if it's a directory
    Children map[string]*MockFile
    // Metadata
    Metadata map[string]interface{}
}
```

#### Functions

```go
func NewMockFileSystemProvider() *MockFileSystemProvider
```

Creates a new `MockFileSystemProvider` with default settings.

#### Methods

```go
func (m *MockFileSystemProvider) Setup() error
```

Initializes the mock provider.

```go
func (m *MockFileSystemProvider) Teardown() error
```

Cleans up the mock provider.

```go
func (m *MockFileSystemProvider) Reset() error
```

Resets the mock provider to its initial state.

```go
func (m *MockFileSystemProvider) AddMockFile(path string, content []byte, metadata map[string]interface{})
```

Adds a mock file to the virtual filesystem.

```go
func (m *MockFileSystemProvider) AddMockDirectory(path string)
```

Adds a mock directory to the virtual filesystem.

```go
func (m *MockFileSystemProvider) AddMockError(path string, operation FSOperationType, err error)
```

Adds a mock error for a specific path and operation.

```go
func (m *MockFileSystemProvider) SetOperationDelay(path string, operation FSOperationType, delay time.Duration)
```

Sets the operation delay for a specific path and operation.

```go
func (m *MockFileSystemProvider) GetOperations(path string, operation FSOperationType) []FSOperation
```

Gets all recorded operations for a specific path and operation type.

```go
func (m *MockFileSystemProvider) VerifyOperation(path string, operation FSOperationType) bool
```

Verifies that a specific operation was performed on a specific path.

### Usage Examples

```go
// Create a mock filesystem provider
mockFS := testutil.NewMockFileSystemProvider()

// Set up the provider
err := mockFS.Setup()
if err != nil {
    // Handle error
}

// Add a mock file
content := []byte("Hello, World!")
metadata := map[string]interface{}{
    "size": len(content),
    "modified": time.Now(),
}
mockFS.AddMockFile("/path/to/file.txt", content, metadata)

// Add a mock directory
mockFS.AddMockDirectory("/path/to/dir")

// Add a mock error for a specific path and operation
mockFS.AddMockError("/path/to/nonexistent", testutil.FSOperationRead, errors.New("file not found"))

// Set an operation delay for a specific path and operation
mockFS.SetOperationDelay("/path/to/large/file.bin", testutil.FSOperationRead, 500*time.Millisecond)

// Use the mock provider in a test
result := framework.RunTest("filesystem-test", func(ctx context.Context) error {
    // Get the mock provider from the framework
    provider, exists := framework.GetMockProvider("filesystem")
    if !exists {
        return errors.New("filesystem provider not found")
    }

    mockFS := provider.(*testutil.MockFileSystemProvider)

    // Use the mock provider
    // ...

    // Verify that a specific operation was performed
    if !mockFS.VerifyOperation("/path/to/file.txt", testutil.FSOperationRead) {
        return errors.New("expected read operation on /path/to/file.txt")
    }

    return nil
})
```

### MockUIProvider

The `MockUIProvider` simulates UI interactions, allowing tests to run without requiring actual user interface components.

### Features

- Simulate UI events and interactions
- Record UI operations for verification
- Configure operation delays and errors
- Provide predefined responses for common UI interactions
- Support for simulating different UI states

### API Reference

#### Types

```go
type MockUIProvider struct {
    // Configuration for the mock provider
    Config MockUIConfig
    // Recorded UI operations
    Operations []UIOperation
    // Predefined responses
    Responses map[string]interface{}
    // Current UI state
    State map[string]interface{}
    // Mutex for thread safety
    mu sync.Mutex
}

type UIOperation struct {
    // Type of operation
    Type UIOperationType
    // Component ID
    ComponentID string
    // Event data
    Data map[string]interface{}
    // Timestamp of the operation
    Timestamp time.Time
}

type UIOperationType string

const (
    // UIOperationClick indicates a click operation
    UIOperationClick UIOperationType = "CLICK"
    // UIOperationInput indicates an input operation
    UIOperationInput UIOperationType = "INPUT"
    // UIOperationSelect indicates a selection operation
    UIOperationSelect UIOperationType = "SELECT"
    // UIOperationDrag indicates a drag operation
    UIOperationDrag UIOperationType = "DRAG"
    // UIOperationDrop indicates a drop operation
    UIOperationDrop UIOperationType = "DROP"
)

type MockUIConfig struct {
    // Default operation delay
    DefaultDelay time.Duration
    // Whether to record operations
    RecordOperations bool
    // Default error rate (0.0 to 1.0)
    ErrorRate float64
}
```

#### Functions

```go
func NewMockUIProvider() *MockUIProvider
```

Creates a new `MockUIProvider` with default settings.

#### Methods

```go
func (m *MockUIProvider) Setup() error
```

Initializes the mock provider.

```go
func (m *MockUIProvider) Teardown() error
```

Cleans up the mock provider.

```go
func (m *MockUIProvider) Reset() error
```

Resets the mock provider to its initial state.

```go
func (m *MockUIProvider) AddMockResponse(componentID string, operationType UIOperationType, response interface{})
```

Adds a mock response for a specific component and operation.

```go
func (m *MockUIProvider) AddMockError(componentID string, operationType UIOperationType, err error)
```

Adds a mock error for a specific component and operation.

```go
func (m *MockUIProvider) SetOperationDelay(componentID string, operationType UIOperationType, delay time.Duration)
```

Sets the operation delay for a specific component and operation.

```go
func (m *MockUIProvider) GetOperations(componentID string, operationType UIOperationType) []UIOperation
```

Gets all recorded operations for a specific component and operation type.

```go
func (m *MockUIProvider) VerifyOperation(componentID string, operationType UIOperationType) bool
```

Verifies that a specific operation was performed on a specific component.

```go
func (m *MockUIProvider) SimulateUIEvent(componentID string, operationType UIOperationType, data map[string]interface{})
```

Simulates a UI event.

### Usage Examples

```go
// Create a mock UI provider
mockUI := testutil.NewMockUIProvider()

// Set up the provider
err := mockUI.Setup()
if err != nil {
    // Handle error
}

// Add a mock response for a specific component and operation
mockUI.AddMockResponse("login-button", testutil.UIOperationClick, map[string]interface{}{
    "success": true,
})

// Add a mock error for a specific component and operation
mockUI.AddMockError("invalid-button", testutil.UIOperationClick, errors.New("button not found"))

// Set an operation delay for a specific component and operation
mockUI.SetOperationDelay("slow-button", testutil.UIOperationClick, 500*time.Millisecond)

// Use the mock provider in a test
result := framework.RunTest("ui-test", func(ctx context.Context) error {
    // Get the mock provider from the framework
    provider, exists := framework.GetMockProvider("ui")
    if !exists {
        return errors.New("ui provider not found")
    }

    mockUI := provider.(*testutil.MockUIProvider)

    // Simulate a UI event
    mockUI.SimulateUIEvent("login-button", testutil.UIOperationClick, map[string]interface{}{
        "x": 100,
        "y": 200,
    })

    // Verify that a specific operation was performed
    if !mockUI.VerifyOperation("login-button", testutil.UIOperationClick) {
        return errors.New("expected click operation on login-button")
    }

    return nil
})
```

## Getting Started

To use mock providers in your tests, follow these steps:

1. **Create and configure mock providers**
2. **Register them with the test framework**
3. **Use them in your tests**
4. **Verify interactions after test execution**

### Integration with TestFramework

Mock providers are integrated with the `TestFramework` to provide simulated components for tests:

```go
// Create a test framework
framework := testutil.NewTestFramework(config, &logger)

// Register mock providers
mockGraph := testutil.NewMockGraphProvider()
framework.RegisterMockProvider("graph", mockGraph)

mockFS := testutil.NewMockFileSystemProvider()
framework.RegisterMockProvider("filesystem", mockFS)

mockUI := testutil.NewMockUIProvider()
framework.RegisterMockProvider("ui", mockUI)

// Get a registered mock provider
provider, exists := framework.GetMockProvider("graph")
if exists {
    mockGraph := provider.(*testutil.MockGraphProvider)
    // Use the mock provider
}
```

### Basic Usage Example

Here's a complete example of using mock providers in a test:

```go
func TestWithMockProviders(t *testing.T) {
    // Create a test framework
    framework := testutil.NewTestFramework(testutil.TestConfig{}, nil)

    // Create and register mock providers
    mockGraph := testutil.NewMockGraphProvider()
    framework.RegisterMockProvider("graph", mockGraph)

    mockFS := testutil.NewMockFileSystemProvider()
    framework.RegisterMockProvider("filesystem", mockFS)

    // Configure mock providers
    mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
        ID:   "root",
        Name: "root",
    })

    mockFS.AddMockFile("/path/to/file.txt", []byte("test content"), nil)

    // Run a test that uses the mock providers
    result := framework.RunTest("mock-provider-test", func(ctx context.Context) error {
        // Get the mock providers from the framework
        graphProvider, _ := framework.GetMockProvider("graph")
        fsProvider, _ := framework.GetMockProvider("filesystem")

        mockGraph := graphProvider.(*testutil.MockGraphProvider)
        mockFS := fsProvider.(*testutil.MockFileSystemProvider)

        // Use the mock providers in your test
        // ...

        // Verify interactions with mock providers
        if !mockGraph.VerifyCalled("/me/drive/root") {
            return errors.New("expected call to /me/drive/root")
        }

        if !mockFS.VerifyOperation("/path/to/file.txt", testutil.FSOperationRead) {
            return errors.New("expected read operation on /path/to/file.txt")
        }

        return nil
    })

    // Check the test result
    if !result.Success {
        t.Errorf("Test failed: %v", result.Error)
    }
}

## Best Practices

1. **Use Descriptive Names**: Register mock providers with descriptive names that clearly indicate their purpose.

2. **Configure Before Use**: Configure mock providers with appropriate responses and behaviors before using them in tests.

3. **Verify Interactions**: Use the verification methods to ensure that the expected interactions occurred during the test.

4. **Clean Up After Tests**: Always clean up mock providers after tests to ensure a clean state for subsequent tests.

5. **Simulate Edge Cases**: Use mock providers to simulate edge cases and error conditions that might be difficult to reproduce with real components.

6. **Combine with Network Simulation**: Use mock providers in conjunction with network simulation to test how components behave under different network conditions.

7. **Record and Verify**: Record operations and verify them to ensure that the system interacts with the mocked components as expected.

8. **Simulate Realistic Behavior**: Configure mock providers to simulate realistic behavior, including delays and occasional errors.

9. **Test Error Handling**: Use mock providers to test how the system handles errors from external components.

10. **Isolate Components**: Use mock providers to isolate the component under test from its dependencies.

## Troubleshooting

When working with mock providers, you might encounter these common issues:

### Configuration Issues

- **Mock provider not found**: Ensure that you've registered the mock provider with the correct name and that you're using the same name when retrieving it.
- **Incorrect mock response**: Verify that you've added the mock response for the correct endpoint and with the correct parameters.
- **Mock provider not initialized**: Make sure you've called `Setup()` on the mock provider before using it.

### Verification Issues

- **Expected call not verified**: If `VerifyCalled()` returns false, check that you're verifying the correct endpoint and that the call was actually made during the test.
- **Expected operation not verified**: If `VerifyOperation()` returns false, check that you're verifying the correct path and operation type.
- **Unexpected call parameters**: When using `VerifyCalledWithParams()`, ensure that the parameters exactly match what was used in the call.

### Type Assertion Issues

- **Type assertion panic**: When retrieving a mock provider from the framework, ensure that you're casting it to the correct type.

```go
// Correct
provider, exists := framework.GetMockProvider("graph")
if exists {
    mockGraph := provider.(*testutil.MockGraphProvider)
    // Use mockGraph
}

// Incorrect - will panic if provider is not a MockGraphProvider
provider, _ := framework.GetMockProvider("graph")
mockGraph := provider.(*testutil.MockGraphProvider)
```

For more detailed troubleshooting information, see the [Testing Troubleshooting Guide](../testing-troubleshooting.md).

## Related Resources

- [Testing Framework Guide](../frameworks/testing-framework-guide.md): Core test configuration and execution
- [Network Simulator](network-simulator-guide.md): Network condition simulation
- [Integration Testing Guide](../frameworks/integration-testing-guide.md): Integration testing environment
- [Performance Testing Framework](../frameworks/performance-testing-guide.md): Performance testing utilities
- [System Testing Framework](../frameworks/system-testing-guide.md): System testing utilities
- [Test Guidelines](../test-guidelines.md): General testing guidelines
