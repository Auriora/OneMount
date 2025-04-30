package testutil

import (
	"context"
	"errors"
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

func TestNewTestFramework(t *testing.T) {
	config := TestConfig{
		Environment:    "test",
		Timeout:        30,
		VerboseLogging: true,
		ArtifactsDir:   "/tmp/test-artifacts",
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

	if framework.Config.ArtifactsDir != "/tmp/test-artifacts" {
		t.Errorf("Expected ArtifactsDir to be '/tmp/test-artifacts', got '%s'", framework.Config.ArtifactsDir)
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

func TestAddResource(t *testing.T) {
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

func TestCleanupResources(t *testing.T) {
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

func TestRegisterAndGetMockProvider(t *testing.T) {
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

func TestRunTest(t *testing.T) {
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

func TestRunTestSuite(t *testing.T) {
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

func TestWithTimeout(t *testing.T) {
	framework := NewTestFramework(TestConfig{}, newMockLogger())
	ctx := framework.WithTimeout(100 * time.Millisecond)

	select {
	case <-ctx.Done():
		// Expected behavior
	case <-time.After(200 * time.Millisecond):
		t.Error("Context did not timeout as expected")
	}
}

func TestWithCancel(t *testing.T) {
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

func TestSetContext(t *testing.T) {
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
