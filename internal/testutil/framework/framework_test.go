package framework

import (
	"context"
	"errors"
	"github.com/auriora/onemount/internal/testutil"
	"testing"
	"time"
)

// mockLogger is a simple implementation of the Logger interface for testing
type mockLogger struct {
	debugMessages []string
	infoMessages  []string
	warnMessages  []string
	errorMessages []string
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugMessages: make([]string, 0),
		infoMessages:  make([]string, 0),
		warnMessages:  make([]string, 0),
		errorMessages: make([]string, 0),
	}
}

func (l *mockLogger) Debug(msg string, args ...interface{}) {
	l.debugMessages = append(l.debugMessages, msg)
}

func (l *mockLogger) Info(msg string, args ...interface{}) {
	l.infoMessages = append(l.infoMessages, msg)
}

func (l *mockLogger) Warn(msg string, args ...interface{}) {
	l.warnMessages = append(l.warnMessages, msg)
}

func (l *mockLogger) Error(msg string, args ...interface{}) {
	l.errorMessages = append(l.errorMessages, msg)
}

// mockResource is a simple implementation of the TestResource interface for testing
type mockResource struct {
	cleanupCalled bool
	cleanupError  error
}

func newMockResource(cleanupError error) *mockResource {
	return &mockResource{
		cleanupCalled: false,
		cleanupError:  cleanupError,
	}
}

func (r *mockResource) Cleanup() error {
	r.cleanupCalled = true
	return r.cleanupError
}

// mockMockProvider is a simple implementation of the MockProvider interface for testing
type mockMockProvider struct {
	setupCalled    bool
	teardownCalled bool
	resetCalled    bool
	setupError     error
	teardownError  error
	resetError     error
}

func newMockMockProvider() *mockMockProvider {
	return &mockMockProvider{
		setupCalled:    false,
		teardownCalled: false,
		resetCalled:    false,
	}
}

func (p *mockMockProvider) Setup() error {
	p.setupCalled = true
	return p.setupError
}

func (p *mockMockProvider) Teardown() error {
	p.teardownCalled = true
	return p.teardownError
}

func (p *mockMockProvider) Reset() error {
	p.resetCalled = true
	return p.resetError
}

// TestUT_FW_01_01_NewTestFramework_ValidConfig_CreatesFramework tests the creation of a new test framework with a valid configuration.
//
//	Test Case ID    UT-FW-01-01
//	Title           Test Framework Creation
//	Description     Tests the creation of a new test framework with a valid configuration
//	Preconditions   None
//	Steps           1. Create a test configuration with valid values
//	                2. Create a mock logger
//	                3. Call NewTestFramework with the configuration and logger
//	                4. Verify the framework is created with the correct properties
//	Expected Result A new test framework is created with all properties set correctly
func TestUT_FW_01_01_NewTestFramework_ValidConfig_CreatesFramework(t *testing.T) {
	expectedArtifactsDir := testutil.GetDefaultArtifactsDir()
	config := TestConfig{
		Environment:    "test",
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   expectedArtifactsDir,
	}
	logger := newMockLogger()

	framework := NewTestFramework(config, logger)

	if framework == nil {
		t.Fatal("Expected non-nil TestFramework")
	}

	if framework.Config.Environment != "test" {
		t.Errorf("Expected Environment to be 'test', got '%s'", framework.Config.Environment)
	}

	if framework.Config.Timeout != 30 {
		t.Errorf("Expected Timeout to be 30, got %d", framework.Config.Timeout)
	}

	if !framework.Config.VerboseLogging {
		t.Error("Expected VerboseLogging to be true")
	}

	if framework.Config.ArtifactsDir != expectedArtifactsDir {
		t.Errorf("Expected ArtifactsDir to be '%s', got '%s'", expectedArtifactsDir, framework.Config.ArtifactsDir)
	}

	if framework.resources == nil {
		t.Error("Expected resources to be initialized")
	}

	if framework.mockProviders == nil {
		t.Error("Expected mockProviders to be initialized")
	}

	if framework.ctx == nil {
		t.Error("Expected ctx to be initialized")
	}

	if framework.logger != logger {
		t.Error("Expected logger to be set correctly")
	}
}

// TestUT_FW_02_01_AddResource_ValidResource_AddsToResourcesList tests adding a resource to the test framework.
//
//	Test Case ID    UT-FW-02-01
//	Title           Add Resource to Framework
//	Description     Tests adding a resource to the test framework
//	Preconditions   A test framework has been created
//	Steps           1. Create a mock resource
//	                2. Call AddResource with the mock resource
//	                3. Verify the resource is added to the resources list
//	Expected Result The resource is successfully added to the framework's resources list
func TestUT_FW_02_01_AddResource_ValidResource_AddsToResourcesList(t *testing.T) {
	framework := NewTestFramework(TestConfig{}, newMockLogger())
	resource := newMockResource(nil)

	framework.AddResource(resource)

	if len(framework.resources) != 1 {
		t.Errorf("Expected resources length to be 1, got %d", len(framework.resources))
	}

	if framework.resources[0] != resource {
		t.Error("Expected resource to be added to resources")
	}
}

// TestUT_FW_03_01_CleanupResources_ResourceWithError_ReturnsError tests cleaning up resources when one resource returns an error.
//
//	Test Case ID    UT-FW-03-01
//	Title           Cleanup Resources with Error
//	Description     Tests cleaning up resources when one resource returns an error
//	Preconditions   A test framework has been created with multiple resources
//	Steps           1. Create a test framework
//	                2. Add multiple resources, one with a cleanup error
//	                3. Call CleanupResources
//	                4. Verify all resources had Cleanup called
//	                5. Verify the error is returned
//	                6. Verify the resources list is cleared
//	Expected Result All resources are cleaned up, the error is returned, and the resources list is cleared
func TestUT_FW_03_01_CleanupResources_ResourceWithError_ReturnsError(t *testing.T) {
	framework := NewTestFramework(TestConfig{}, newMockLogger())
	resource1 := newMockResource(nil)
	resource2 := newMockResource(errors.New("cleanup error"))
	resource3 := newMockResource(nil)

	framework.AddResource(resource1)
	framework.AddResource(resource2)
	framework.AddResource(resource3)

	err := framework.CleanupResources()

	if err == nil {
		t.Error("Expected error from CleanupResources")
	}

	if !resource1.cleanupCalled {
		t.Error("Expected resource1.Cleanup to be called")
	}

	if !resource2.cleanupCalled {
		t.Error("Expected resource2.Cleanup to be called")
	}

	if !resource3.cleanupCalled {
		t.Error("Expected resource3.Cleanup to be called")
	}

	if len(framework.resources) != 0 {
		t.Errorf("Expected resources to be cleared, got length %d", len(framework.resources))
	}
}

// TestUT_FW_04_01_RegisterAndGetMockProvider_ValidProvider_RegistersAndRetrieves tests registering and retrieving a mock provider.
//
//	Test Case ID    UT-FW-04-01
//	Title           Register and Get Mock Provider
//	Description     Tests registering and retrieving a mock provider
//	Preconditions   A test framework has been created
//	Steps           1. Create a test framework
//	                2. Create a mock provider
//	                3. Register the mock provider with a name
//	                4. Retrieve the provider using the same name
//	                5. Verify the retrieved provider is the same as the registered one
//	                6. Try to retrieve a non-existent provider
//	                7. Verify the non-existent provider is not found
//	Expected Result The provider is successfully registered and retrieved, and non-existent providers are not found
func TestUT_FW_04_01_RegisterAndGetMockProvider_ValidProvider_RegistersAndRetrieves(t *testing.T) {
	framework := NewTestFramework(TestConfig{}, newMockLogger())
	provider := newMockMockProvider()

	framework.RegisterMockProvider("test-provider", provider)

	retrievedProvider, exists := framework.GetMockProvider("test-provider")
	if !exists {
		t.Error("Expected provider to exist")
	}

	if retrievedProvider != provider {
		t.Error("Expected retrieved provider to be the same as registered provider")
	}

	_, exists = framework.GetMockProvider("non-existent-provider")
	if exists {
		t.Error("Expected non-existent provider to not exist")
	}
}

// TestUT_FW_05_01_RunTest_VariousTestCases_ReturnsCorrectResults tests running tests with different outcomes.
//
//	Test Case ID    UT-FW-05-01
//	Title           Run Test with Various Outcomes
//	Description     Tests running tests that succeed, fail, and timeout
//	Preconditions   None
//	Steps           1. Create a test framework
//	                2. Run a successful test and verify the result
//	                3. Run a failing test and verify the result
//	                4. Run a test that times out and verify the result
//	Expected Result Each test returns the correct result status and failure information
func TestUT_FW_05_01_RunTest_VariousTestCases_ReturnsCorrectResults(t *testing.T) {
	logger := newMockLogger()
	framework := NewTestFramework(TestConfig{}, logger)

	// Test successful test
	result := framework.RunTest("successful-test", func(ctx context.Context) error {
		return nil
	})

	if result.Status != TestStatusPassed {
		t.Errorf("Expected status to be TestStatusPassed, got %s", result.Status)
	}

	if len(result.Failures) != 0 {
		t.Errorf("Expected no failures, got %d", len(result.Failures))
	}

	if result.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	// Test failed test
	result = framework.RunTest("failed-test", func(ctx context.Context) error {
		return errors.New("test error")
	})

	if result.Status != TestStatusFailed {
		t.Errorf("Expected status to be TestStatusFailed, got %s", result.Status)
	}

	if len(result.Failures) != 1 {
		t.Errorf("Expected 1 failure, got %d", len(result.Failures))
	}

	if result.Failures[0].Message != "test error" {
		t.Errorf("Expected failure message to be 'test error', got '%s'", result.Failures[0].Message)
	}

	// Test timeout
	framework = NewTestFramework(TestConfig{Timeout: 1}, logger)
	result = framework.RunTest("timeout-test", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			return nil
		}
	})

	if result.Status != TestStatusFailed {
		t.Errorf("Expected status to be TestStatusFailed, got %s", result.Status)
	}

	if len(result.Failures) != 1 {
		t.Errorf("Expected 1 failure, got %d", len(result.Failures))
	}

	if result.Failures[0].Message != "context deadline exceeded" {
		t.Errorf("Expected failure message to be 'context deadline exceeded', got '%s'", result.Failures[0].Message)
	}
}

// TestUT_FW_06_01_RunTestSuite_MixedResults_ReturnsCorrectCounts tests running a test suite with mixed results.
//
//	Test Case ID    UT-FW-06-01
//	Title           Run Test Suite with Mixed Results
//	Description     Tests running a test suite with a mix of passing and failing tests
//	Preconditions   None
//	Steps           1. Create a test framework
//	                2. Define a map of test functions with mixed results
//	                3. Run the test suite
//	                4. Verify the correct number of results
//	                5. Count the number of passed and failed tests
//	                6. Verify the counts match expectations
//	Expected Result The test suite returns the correct number of results with the expected pass/fail counts
func TestUT_FW_06_01_RunTestSuite_MixedResults_ReturnsCorrectCounts(t *testing.T) {
	logger := newMockLogger()
	framework := NewTestFramework(TestConfig{}, logger)

	tests := map[string]func(ctx context.Context) error{
		"test1": func(ctx context.Context) error {
			return nil
		},
		"test2": func(ctx context.Context) error {
			return errors.New("test2 error")
		},
		"test3": func(ctx context.Context) error {
			return nil
		},
	}

	results := framework.RunTestSuite("test-suite", tests)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	passCount := 0
	failCount := 0
	for _, result := range results {
		if result.Status == TestStatusPassed {
			passCount++
		} else if result.Status == TestStatusFailed {
			failCount++
		}
	}

	if passCount != 2 {
		t.Errorf("Expected 2 passed tests, got %d", passCount)
	}

	if failCount != 1 {
		t.Errorf("Expected 1 failed test, got %d", failCount)
	}
}

// TestUT_FW_07_01_WithTimeout_ShortTimeout_ContextExpires tests creating a context with a timeout.
//
//	Test Case ID    UT-FW-07-01
//	Title           Context with Timeout
//	Description     Tests creating a context with a timeout that expires
//	Preconditions   None
//	Steps           1. Create a test framework
//	                2. Create a context with a short timeout
//	                3. Wait for the context to expire
//	Expected Result The context expires within the expected time
func TestUT_FW_07_01_WithTimeout_ShortTimeout_ContextExpires(t *testing.T) {
	framework := NewTestFramework(TestConfig{}, newMockLogger())
	ctx, cancel := framework.WithTimeout(100 * time.Millisecond)
	defer cancel() // Ensure the cancel function is called to avoid context leaks

	select {
	case <-ctx.Done():
		// Expected behavior
	case <-time.After(200 * time.Millisecond):
		t.Error("Context did not timeout as expected")
	}
}

// TestUT_FW_08_01_WithCancel_ImmediateCancel_ContextCancels tests creating a context with a cancel function.
//
//	Test Case ID    UT-FW-08-01
//	Title           Context with Cancel Function
//	Description     Tests creating a context with a cancel function and canceling it
//	Preconditions   None
//	Steps           1. Create a test framework
//	                2. Create a context with a cancel function
//	                3. Call the cancel function
//	                4. Verify the context is canceled
//	Expected Result The context is canceled immediately after calling the cancel function
func TestUT_FW_08_01_WithCancel_ImmediateCancel_ContextCancels(t *testing.T) {
	framework := NewTestFramework(TestConfig{}, newMockLogger())
	ctx, cancel := framework.WithCancel()

	cancel()

	select {
	case <-ctx.Done():
		// Expected behavior
	case <-time.After(100 * time.Millisecond):
		t.Error("Context was not canceled as expected")
	}
}

// TestUT_FW_09_01_SetContext_CustomContext_UsedInTests tests setting a custom context in the framework.
//
//	Test Case ID    UT-FW-09-01
//	Title           Set Custom Context
//	Description     Tests setting a custom context in the framework and using it in tests
//	Preconditions   None
//	Steps           1. Create a test framework
//	                2. Create a custom context with a value
//	                3. Set the custom context in the framework
//	                4. Verify the context is set correctly
//	                5. Run a test that uses the context value
//	                6. Verify the test passes
//	Expected Result The custom context is set correctly and used in tests
func TestUT_FW_09_01_SetContext_CustomContext_UsedInTests(t *testing.T) {
	framework := NewTestFramework(TestConfig{}, newMockLogger())
	customCtx := context.WithValue(context.Background(), "key", "value")

	framework.SetContext(customCtx)

	if framework.ctx != customCtx {
		t.Error("Expected context to be set correctly")
	}

	// Test that the context is used in RunTest
	result := framework.RunTest("context-test", func(ctx context.Context) error {
		val := ctx.Value("key")
		if val != "value" {
			return errors.New("context value not set correctly")
		}
		return nil
	})

	if result.Status != TestStatusPassed {
		t.Errorf("Expected status to be TestStatusPassed, got %s", result.Status)
	}
}
