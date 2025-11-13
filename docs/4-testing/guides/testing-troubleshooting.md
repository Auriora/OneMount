# Testing Troubleshooting Guide

This guide provides solutions for common issues encountered when using the OneMount test framework.

## Common Issues

### Tests Fail to Start

#### Symptoms
- Tests fail immediately without running any test logic
- Error messages about initialization failures
- Context cancellation errors

#### Possible Causes
1. **Missing or incorrect configuration**
2. **Resource initialization failures**
3. **Context cancellation or timeout**

#### Solutions

1. **Check test configuration**
   ```go
   // Ensure configuration is properly set
   config := testutil.TestConfig{
       Environment:    "test",
       Timeout:        30,  // Ensure timeout is sufficient
       VerboseLogging: true,
       ArtifactsDir:   "/tmp/test-artifacts",
   }
   ```

2. **Verify resource initialization**
   ```go
   // Add error handling for resource initialization
   resource, err := NewSomeResource()
   if err != nil {
       t.Fatalf("Failed to initialize resource: %v", err)
   }
   framework.AddResource(resource)
   ```

3. **Check for context cancellation**
   ```go
   // Use a longer timeout if needed
   ctx := framework.WithTimeout(5 * time.Minute)
   
   // Or use context without timeout for setup
   ctx := context.Background()
   ```

### Mock Providers Not Working Correctly

#### Symptoms
- Tests fail with unexpected behavior from mock providers
- Error messages about missing responses
- Verification failures for mock provider interactions

#### Possible Causes
1. **Mock provider not registered or initialized**
2. **Missing mock responses**
3. **Incorrect mock configuration**

#### Solutions

1. **Verify mock provider registration**
   ```go
   // Register the mock provider
   mockGraph := testutil.NewMockGraphProvider()
   framework.RegisterMockProvider("graph", mockGraph)
   
   // Verify the provider is registered
   provider, exists := framework.GetMockProvider("graph")
   if !exists {
       t.Fatal("Graph provider not registered")
   }
   ```

2. **Add necessary mock responses**
   ```go
   // Add mock responses for all endpoints used in the test
   mockGraph.AddMockResponse("/me/drive/root", &graph.DriveItem{
       ID:   "root",
       Name: "root",
   })
   
   mockGraph.AddMockResponse("/me/drive/root/children", []*graph.DriveItem{
       {ID: "item1", Name: "item1"},
       {ID: "item2", Name: "item2"},
   })
   ```

3. **Check mock provider configuration**
   ```go
   // Configure the mock provider correctly
   mockGraph.Config.SimulateAuth = true
   mockGraph.Config.RecordCalls = true
   mockGraph.SetAuthState(testutil.AuthStateValid)
   ```

### Network Simulation Issues

#### Symptoms
- Tests behave inconsistently with network simulation
- Network conditions not applied correctly
- Disconnection/reconnection not working as expected

#### Possible Causes
1. **Network simulator not initialized**
2. **Incorrect network condition settings**
3. **Missing network reconnection after tests**

#### Solutions

1. **Verify network simulator initialization**
   ```go
   // Get the network simulator
   simulator := framework.GetNetworkSimulator()
   if simulator == nil {
       t.Fatal("Network simulator not initialized")
   }
   ```

2. **Check network condition settings**
   ```go
   // Set explicit network conditions
   err := framework.SetNetworkConditions(100*time.Millisecond, 0.1, 1000)
   if err != nil {
       t.Fatalf("Failed to set network conditions: %v", err)
   }
   
   // Or use a preset
   err = framework.ApplyNetworkPreset(testutil.SlowNetwork)
   if err != nil {
       t.Fatalf("Failed to apply network preset: %v", err)
   }
   ```

3. **Ensure network reconnection after tests**
   ```go
   // Always reconnect the network in cleanup
   t.Cleanup(func() {
       if !framework.IsNetworkConnected() {
           framework.ReconnectNetwork()
       }
       framework.CleanupResources()
   })
   ```

### Resource Cleanup Failures

#### Symptoms
- Tests leave behind resources
- Subsequent tests fail due to resource conflicts
- Error messages about resource cleanup failures

#### Possible Causes
1. **Missing cleanup calls**
2. **Cleanup not executed due to panics**
3. **Resources not properly registered with the framework**

#### Solutions

1. **Use t.Cleanup for guaranteed cleanup**
   ```go
   // Use t.Cleanup to ensure resources are cleaned up
   t.Cleanup(func() {
       framework.CleanupResources()
   })
   ```

2. **Handle panics in cleanup**
   ```go
   // Add panic recovery in cleanup
   t.Cleanup(func() {
       defer func() {
           if r := recover(); r != nil {
               t.Logf("Panic during cleanup: %v", r)
           }
       }()
       framework.CleanupResources()
   })
   ```

3. **Register all resources with the framework**
   ```go
   // Register all resources that need cleanup
   tempDir := createTempDir()
   framework.AddResource(tempDir)
   
   mockFile := createTempFile()
   framework.AddResource(mockFile)
   ```

### Integration Test Environment Issues

#### Symptoms
- Integration tests fail to set up the environment
- Component isolation not working as expected
- Test data management issues

#### Possible Causes
1. **Incorrect isolation configuration**
2. **Component setup failures**
3. **Test data not loaded or cleaned up properly**

#### Solutions

1. **Check isolation configuration**
   ```go
   // Set up proper isolation configuration
   env.SetIsolationConfig(testutil.IsolationConfig{
       MockedServices: []string{"graph", "filesystem", "ui"},
       NetworkRules: []testutil.NetworkRule{
           {
               Source:      "filesystem",
               Destination: "graph",
               Allow:       true,
           },
       },
       DataIsolation: true,
   })
   ```

2. **Verify component setup**
   ```go
   // Set up the environment with error handling
   err := env.SetupEnvironment()
   if err != nil {
       t.Fatalf("Failed to set up environment: %v", err)
   }
   
   // Get and verify components
   graphComponent, err := env.GetComponent("graph")
   if err != nil {
       t.Fatalf("Failed to get graph component: %v", err)
   }
   ```

3. **Manage test data properly**
   ```go
   // Load test data with error handling
   testDataManager := env.GetTestDataManager()
   err := testDataManager.LoadTestData("test-data-set")
   if err != nil {
       t.Fatalf("Failed to load test data: %v", err)
   }
   
   // Clean up test data in cleanup
   t.Cleanup(func() {
       testDataManager.CleanupTestData()
   })
   ```

### Test Scenario Execution Issues

#### Symptoms
- Test scenarios fail to execute
- Steps or assertions fail unexpectedly
- Cleanup steps not executed

#### Possible Causes
1. **Scenario not properly defined**
2. **Step actions or validations failing**
3. **Assertion conditions not met**

#### Solutions

1. **Check scenario definition**
   ```go
   // Define a complete scenario with all required components
   scenario := testutil.TestScenario{
       Name:        "Test Scenario",
       Description: "Tests some functionality",
       Steps: []testutil.TestStep{
           {
               Name: "Step 1",
               Action: func(ctx context.Context) error {
                   // Action implementation
                   return nil
               },
               Validation: func(ctx context.Context) error {
                   // Validation implementation
                   return nil
               },
           },
           // More steps...
       },
       Assertions: []testutil.TestAssertion{
           {
               Name: "Assertion 1",
               Condition: func(ctx context.Context) bool {
                   // Condition implementation
                   return true
               },
               Message: "Assertion 1 failed",
           },
           // More assertions...
       },
       Cleanup: []testutil.CleanupStep{
           {
               Name: "Cleanup 1",
               Action: func(ctx context.Context) error {
                   // Cleanup implementation
                   return nil
               },
               AlwaysRun: true,
           },
           // More cleanup steps...
       },
   }
   ```

2. **Debug step actions and validations**
   ```go
   // Add logging to step actions and validations
   Action: func(ctx context.Context) error {
       logger.Info("Executing step action", "step", "Step 1")
       // Action implementation
       result := someOperation()
       logger.Info("Step action result", "result", result)
       return nil
   },
   Validation: func(ctx context.Context) error {
       logger.Info("Validating step", "step", "Step 1")
       // Validation implementation
       if !someCondition {
           return errors.New("validation failed: condition not met")
       }
       return nil
   },
   ```

3. **Check assertion conditions**
   ```go
   // Add logging to assertion conditions
   Condition: func(ctx context.Context) bool {
       logger.Info("Checking assertion", "assertion", "Assertion 1")
       // Condition implementation
       result := someCondition()
       logger.Info("Assertion result", "result", result)
       return result
   },
   ```

### Performance Testing Issues

#### Symptoms
- Performance tests produce inconsistent results
- Metrics collection fails
- Thresholds not properly checked

#### Possible Causes
1. **Inconsistent test environment**
2. **Insufficient test duration or iterations**
3. **Incorrect metric collection or threshold configuration**

#### Solutions

1. **Standardize the test environment**
   ```go
   // Use a consistent environment for performance tests
   framework := testutil.NewTestFramework(testutil.TestConfig{
       Environment:    "performance",
       Timeout:        300,  // Longer timeout for performance tests
       VerboseLogging: false,  // Reduce logging overhead
   }, &logger)
   
   // Apply consistent network conditions
   framework.ApplyNetworkPreset(testutil.FastNetwork)
   ```

2. **Increase test duration or iterations**
   ```go
   // Use longer duration for load tests
   scenario := testutil.LoadTestScenario{
       Name:        "Sustained Load",
       Description: "Tests performance under sustained load",
       Duration:    5 * time.Minute,  // Longer duration
       Concurrency: 10,
       // ...
   }
   
   // Or increase iterations for benchmark tests
   benchmark := testutil.PerformanceBenchmark{
       Name:       "File Upload Benchmark",
       Iterations: 100,  // More iterations
       // ...
   }
   ```

3. **Check metric collection and thresholds**
   ```go
   // Configure proper metrics collection
   benchmark.EnableMetrics([]string{"latency", "throughput", "cpu", "memory"})
   
   // Set appropriate thresholds
   benchmark.SetThresholds(testutil.PerformanceThresholds{
       MaxLatency:    100 * time.Millisecond,
       MinThroughput: 10,  // operations per second
       MaxCPU:        50,  // percent
       MaxMemory:     100 * 1024 * 1024,  // 100 MB
   })
   ```

### Security Testing Issues

#### Symptoms
- Security tests fail unexpectedly
- Vulnerability scanning not working
- Security control verification issues

#### Possible Causes
1. **Missing security test dependencies**
2. **Incorrect security test configuration**
3. **Environment restrictions preventing security tests**

#### Solutions

1. **Check security test dependencies**
   ```go
   // Verify security test dependencies
   if !testutil.SecurityTestDependenciesAvailable() {
       t.Skip("Security test dependencies not available")
   }
   ```

2. **Configure security tests properly**
   ```go
   // Configure security tests with proper settings
   securityFramework := testutil.NewSecurityTestFramework(testutil.SecurityTestConfig{
       ScanDepth:           "deep",
       IncludeNetworkScan:  true,
       IncludeVulnScan:     true,
       SkipPrivilegedTests: !isPrivilegedEnvironment(),
   })
   ```

3. **Handle environment restrictions**
   ```go
   // Check for environment restrictions
   if !testutil.CanRunPrivilegedTests() {
       t.Skip("Cannot run privileged security tests in this environment")
   }
   
   // Or modify tests to work in restricted environments
   securityScenarios := testutil.NewSecurityTestScenarios(securityFramework)
   if testutil.IsRestrictedEnvironment() {
       // Use non-privileged scenarios
       env.AddScenario(securityScenarios.NonPrivilegedAuthenticationTestScenario())
   } else {
       // Use full scenarios
       env.AddScenario(securityScenarios.AuthenticationTestScenario())
   }
   ```

## Debugging Techniques

### Enabling Verbose Logging

To get more detailed information about what's happening during tests, enable verbose logging:

```go
// Enable verbose logging in the test configuration
config := testutil.TestConfig{
    Environment:    "test",
    Timeout:        30,
    VerboseLogging: true,
}

// Or set the environment variable
os.Setenv("ONEMOUNT_TEST_VERBOSE", "1")
```

### Using the Debug Logger

The test framework provides a debug logger that can be used to output detailed information:

```go
// Get the logger from the framework
logger := framework.GetLogger()

// Log detailed information
logger.Debug("Detailed information", 
    "component", "test",
    "operation", "some-operation",
    "value", someValue)
```

### Inspecting Test Results

Examine the test results to understand what went wrong:

```go
// Run a test and inspect the result
result := framework.RunTest("test-name", func(ctx context.Context) error {
    // Test logic
    return nil
})

// Check the status
if result.Status != testutil.TestStatusPassed {
    // Inspect failures
    for _, failure := range result.Failures {
        fmt.Printf("Failure: %s at %s\n", failure.Message, failure.Location)
        fmt.Printf("Expected: %v\n", failure.Expected)
        fmt.Printf("Actual: %v\n", failure.Actual)
    }
}

// Check artifacts
for _, artifact := range result.Artifacts {
    fmt.Printf("Artifact: %s (%s) at %s\n", artifact.Name, artifact.Type, artifact.Location)
}
```

### Using Test Artifacts

Test artifacts can provide valuable information for debugging:

```go
// Create a test artifact
framework.CreateArtifact("debug-info", "text/plain", []byte("Debug information"))

// Or save a file as an artifact
framework.SaveFileAsArtifact("/path/to/log/file.log", "log", "text/plain")
```

### Isolating Components

Isolate components to identify the source of issues:

```go
// Test with only specific components mocked
env.SetIsolationConfig(testutil.IsolationConfig{
    MockedServices: []string{"graph"},  // Only mock the graph service
    NetworkRules:   []testutil.NetworkRule{},
    DataIsolation:  true,
})
```

### Step-by-Step Execution

Run test steps individually to identify where issues occur:

```go
// Define steps
steps := []struct {
    name string
    fn   func() error
}{
    {"Step 1", func() error { return step1() }},
    {"Step 2", func() error { return step2() }},
    {"Step 3", func() error { return step3() }},
}

// Run steps individually
for _, step := range steps {
    t.Logf("Running step: %s", step.name)
    if err := step.fn(); err != nil {
        t.Fatalf("Step %s failed: %v", step.name, err)
    }
}
```

## Getting Help

If you're still experiencing issues after trying the solutions in this guide, there are several ways to get help:

1. **Check the documentation**: Review the documentation for the specific component you're having issues with.

2. **Search for similar issues**: Check if others have encountered similar issues by searching the issue tracker.

3. **Ask for help**: Reach out to the development team for assistance.

4. **Contribute improvements**: If you find and fix an issue, consider contributing your solution back to the project.