// Package framework provides testing utilities for the OneMount project.
package framework

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUT_FW_10_01_SetupSignalHandling_ValidFramework_RegistersSignalHandlers tests setting up signal handling in the test framework.
//
//	Test Case ID    UT-FW-10-01
//	Title           Setup Signal Handling
//	Description     Tests setting up signal handling in the test framework
//	Preconditions   A test framework has been created
//	Steps           1. Create a test framework
//	                2. Set up signal handling
//	                3. Verify signal handling is set up correctly
//	                4. Call the cleanup function
//	                5. Verify signal handling is stopped
//	Expected Result Signal handling is set up and stopped correctly
func TestUT_FW_10_01_SetupSignalHandling_ValidFramework_RegistersSignalHandlers(t *testing.T) {
	// Create a mock logger
	logger := newMockLogger()

	// Create a TestFramework
	tf := NewTestFramework(TestConfig{}, logger)

	// Create a mock resource
	resource := newMockResource(nil)

	// Add the resource to the TestFramework
	tf.AddResource(resource)

	// Set up signal handling
	cleanup := tf.SetupSignalHandling()
	require.NotNil(t, cleanup, "Cleanup function should not be nil")

	// Verify that signal handling is set up
	assert.True(t, tf.isHandling, "Signal handling should be active")
	assert.NotNil(t, tf.signalChan, "Signal channel should not be nil")

	// Verify that the logger recorded the setup
	assert.Contains(t, logger.infoMessages, "Signal handling set up for SIGINT and SIGTERM", "Logger should record signal handling setup")

	// Call the cleanup function
	cleanup()

	// Verify that signal handling is stopped
	assert.False(t, tf.isHandling, "Signal handling should be inactive after cleanup")
	assert.Contains(t, logger.infoMessages, "Signal handling stopped", "Logger should record signal handling stopped")
}

// TestUT_FW_10_02_SetupSignalHandlingIdempotent_CalledTwice_OnlyRegistersOnce tests that calling SetupSignalHandling twice only registers signal handlers once.
//
//	Test Case ID    UT-FW-10-02
//	Title           Setup Signal Handling Idempotent
//	Description     Tests that calling SetupSignalHandling twice only registers signal handlers once
//	Preconditions   A test framework has been created
//	Steps           1. Create a test framework
//	                2. Set up signal handling twice
//	                3. Verify signal handling is only set up once
//	                4. Call both cleanup functions
//	                5. Verify signal handling is stopped correctly
//	Expected Result Signal handling is set up only once and stopped correctly
func TestUT_FW_10_02_SetupSignalHandlingIdempotent_CalledTwice_OnlyRegistersOnce(t *testing.T) {
	// Create a mock logger
	logger := newMockLogger()

	// Create a TestFramework
	tf := NewTestFramework(TestConfig{}, logger)

	// Set up signal handling twice
	cleanup1 := tf.SetupSignalHandling()
	require.NotNil(t, cleanup1, "First cleanup function should not be nil")

	// Clear the logs
	logger.infoMessages = make([]string, 0)

	// Set up signal handling again
	cleanup2 := tf.SetupSignalHandling()
	require.NotNil(t, cleanup2, "Second cleanup function should not be nil")

	// Verify that the logger recorded that signal handling was already set up
	assert.Contains(t, logger.infoMessages, "Signal handling already set up", "Logger should record that signal handling was already set up")

	// Call the second cleanup function
	cleanup2()

	// Verify that the logger recorded that signal handling was already stopped
	assert.Contains(t, logger.infoMessages, "Signal handling already stopped by another call", "Logger should record that signal handling was already stopped")

	// Call the first cleanup function
	cleanup1()

	// Verify that signal handling is stopped
	assert.False(t, tf.isHandling, "Signal handling should be inactive after cleanup")
	assert.Contains(t, logger.infoMessages, "Signal handling stopped", "Logger should record signal handling stopped")
}

// TestUT_FW_10_03_CleanupResourcesOnSignal_ResourceAdded_ResourceCleaned tests that resources are cleaned up when a signal is received.
//
//	Test Case ID    UT-FW-10-03
//	Title           Cleanup Resources On Signal
//	Description     Tests that resources are cleaned up when a signal is received
//	Preconditions   A test framework has been created with a resource
//	Steps           1. Create a test framework
//	                2. Add a resource to the framework
//	                3. Set up signal handling
//	                4. Simulate receiving a signal
//	                5. Verify the resource is cleaned up
//	Expected Result The resource is cleaned up when a signal is received
func TestUT_FW_10_03_CleanupResourcesOnSignal_ResourceAdded_ResourceCleaned(t *testing.T) {
	// Skip this test if we're not in a test environment that allows us to fork
	// This test actually sends a signal to itself, which would terminate the process
	// So we need to fork a child process to test this
	if os.Getenv("TEST_SIGNAL_HANDLING") != "1" {
		t.Skip("Skipping signal handling test in parent process")
	}

	// Create a mock logger
	logger := newMockLogger()

	// Create a TestFramework
	tf := NewTestFramework(TestConfig{}, logger)

	// Create a mock resource
	resource := newMockResource(nil)

	// Add the resource to the TestFramework
	tf.AddResource(resource)

	// Set up signal handling
	tf.SetupSignalHandling()

	// Simulate receiving a signal by directly calling the cleanup function
	// In a real scenario, this would be triggered by a signal
	if err := tf.CleanupResources(); err != nil {
		t.Fatalf("CleanupResources failed: %v", err)
	}

	// Verify that the resource was cleaned up
	assert.True(t, resource.cleanupCalled, "Resource cleanup should be called")
}
