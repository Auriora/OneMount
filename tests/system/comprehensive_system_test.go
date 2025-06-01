// Package system provides comprehensive system testing for OneMount
package system

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/auriora/onemount/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystemST_COMPREHENSIVE_01_AllOperations runs comprehensive system tests using real OneDrive account
//
//	Test Case ID    ST-COMPREHENSIVE-01
//	Title           Comprehensive System Test Suite
//	Description     Run all system tests using real OneDrive account to verify end-to-end functionality
//	Preconditions   1. Real OneDrive account credentials available in ~/.onemount-tests/.auth_tokens.json
//	                2. Network connection available
//	                3. OneDrive account has sufficient storage space
//	Steps           1. Initialize system test suite with real authentication
//	                2. Mount OneMount filesystem
//	                3. Run basic file operations tests
//	                4. Run directory operations tests
//	                5. Run large file operations tests
//	                6. Run special character file tests
//	                7. Run concurrent operations tests
//	                8. Run file permissions tests
//	                9. Run streaming operations tests
//	                10. Run performance tests
//	                11. Run error handling tests
//	                12. Clean up all test data
//	Expected Result All tests pass successfully, demonstrating full system functionality
func TestSystemST_COMPREHENSIVE_01_AllOperations(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping comprehensive system tests in short mode")
	}

	// Check if auth tokens exist
	if _, err := os.Stat(testutil.AuthTokensPath); os.IsNotExist(err) {
		t.Skipf("Skipping system tests: auth tokens not found at %s", testutil.AuthTokensPath)
	}

	// Log test start
	t.Logf("Starting comprehensive system tests")

	// Create system test suite
	suite, err := NewSystemTestSuite(t)
	require.NoError(t, err, "Failed to create system test suite")

	// Setup test environment
	err = suite.Setup()
	require.NoError(t, err, "Failed to setup system test environment")

	// Ensure cleanup runs even if tests fail
	t.Cleanup(func() {
		if err := suite.Cleanup(); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	})

	// Run all test scenarios
	testScenarios := []struct {
		name string
		test func() error
	}{
		{"Basic File Operations", suite.TestBasicFileOperations},
		{"Directory Operations", suite.TestDirectoryOperations},
		{"Large File Operations", suite.TestLargeFileOperations},
		{"Special Character Files", suite.TestSpecialCharacterFiles},
		{"Concurrent Operations", suite.TestConcurrentOperations},
		{"File Permissions", suite.TestFilePermissions},
		{"Streaming Operations", suite.TestStreamingOperations},
	}

	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			t.Logf("Starting test scenario: %s", scenario.name)
			startTime := time.Now()

			err := scenario.test()
			duration := time.Since(startTime)

			if err != nil {
				t.Errorf("Test scenario %s failed after %v: %v", scenario.name, duration, err)
			} else {
				t.Logf("Test scenario %s completed successfully in %v", scenario.name, duration)
			}
		})
	}
}

// TestSystemST_PERFORMANCE_01_UploadDownloadSpeed tests upload and download performance
//
//	Test Case ID    ST-PERFORMANCE-01
//	Title           Upload/Download Performance Test
//	Description     Measure upload and download speeds for various file sizes
//	Preconditions   1. Real OneDrive account credentials available
//	                2. Network connection available
//	Steps           1. Create files of various sizes (1KB, 100KB, 1MB, 10MB)
//	                2. Measure upload time for each file
//	                3. Measure download time for each file
//	                4. Calculate and report speeds
//	Expected Result Performance metrics are within acceptable ranges
func TestSystemST_PERFORMANCE_01_UploadDownloadSpeed(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	if _, err := os.Stat(testutil.AuthTokensPath); os.IsNotExist(err) {
		t.Skipf("Skipping performance tests: auth tokens not found at %s", testutil.AuthTokensPath)
	}

	suite, err := NewSystemTestSuite(t)
	require.NoError(t, err, "Failed to create system test suite")

	err = suite.Setup()
	require.NoError(t, err, "Failed to setup system test environment")

	t.Cleanup(func() {
		if err := suite.Cleanup(); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	})

	// Test different file sizes
	fileSizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"100KB", 100 * 1024},
		{"1MB", 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
	}

	for _, fileSize := range fileSizes {
		t.Run(fmt.Sprintf("Performance_%s", fileSize.name), func(t *testing.T) {
			err := suite.TestPerformance(fileSize.name, fileSize.size)
			assert.NoError(t, err, "Performance test failed for %s", fileSize.name)
		})
	}
}

// TestSystemST_RELIABILITY_01_ErrorRecovery tests error recovery scenarios
//
//	Test Case ID    ST-RELIABILITY-01
//	Title           Error Recovery Test
//	Description     Test system behavior under various error conditions
//	Preconditions   1. Real OneDrive account credentials available
//	                2. Network connection available
//	Steps           1. Test behavior with invalid file names
//	                2. Test behavior with insufficient permissions
//	                3. Test behavior with network interruptions
//	                4. Test behavior with disk space issues
//	Expected Result System handles errors gracefully and recovers appropriately
func TestSystemST_RELIABILITY_01_ErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping reliability tests in short mode")
	}

	if _, err := os.Stat(testutil.AuthTokensPath); os.IsNotExist(err) {
		t.Skipf("Skipping reliability tests: auth tokens not found at %s", testutil.AuthTokensPath)
	}

	suite, err := NewSystemTestSuite(t)
	require.NoError(t, err, "Failed to create system test suite")

	err = suite.Setup()
	require.NoError(t, err, "Failed to setup system test environment")

	t.Cleanup(func() {
		if err := suite.Cleanup(); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	})

	// Test error recovery scenarios
	errorScenarios := []struct {
		name string
		test func() error
	}{
		{"Invalid File Names", suite.TestInvalidFileNames},
		{"Authentication Refresh", suite.TestAuthenticationRefresh},
		{"Disk Space Handling", suite.TestDiskSpaceHandling},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := scenario.test()
			assert.NoError(t, err, "Error recovery test failed for %s", scenario.name)
		})
	}
}

// TestSystemST_INTEGRATION_01_MountUnmount tests mount/unmount operations
//
//	Test Case ID    ST-INTEGRATION-01
//	Title           Mount/Unmount Integration Test
//	Description     Test filesystem mount and unmount operations
//	Preconditions   1. Real OneDrive account credentials available
//	                2. Network connection available
//	Steps           1. Mount filesystem
//	                2. Verify mount point is accessible
//	                3. Create test files
//	                4. Unmount filesystem
//	                5. Verify mount point is no longer accessible
//	                6. Remount filesystem
//	                7. Verify test files are still present
//	Expected Result Mount/unmount operations work correctly and data persists
func TestSystemST_INTEGRATION_01_MountUnmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	if _, err := os.Stat(testutil.AuthTokensPath); os.IsNotExist(err) {
		t.Skipf("Skipping integration tests: auth tokens not found at %s", testutil.AuthTokensPath)
	}

	suite, err := NewSystemTestSuite(t)
	require.NoError(t, err, "Failed to create system test suite")

	err = suite.Setup()
	require.NoError(t, err, "Failed to setup system test environment")

	t.Cleanup(func() {
		if err := suite.Cleanup(); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	})

	// Test mount/unmount cycle
	err = suite.TestMountUnmountCycle()
	assert.NoError(t, err, "Mount/unmount cycle test failed")
}

// TestSystemST_STRESS_01_HighLoad tests system behavior under high load
//
//	Test Case ID    ST-STRESS-01
//	Title           High Load Stress Test
//	Description     Test system behavior under high load conditions
//	Preconditions   1. Real OneDrive account credentials available
//	                2. Network connection available
//	                3. Sufficient system resources
//	Steps           1. Create many files simultaneously
//	                2. Perform many operations concurrently
//	                3. Monitor system resources
//	                4. Verify all operations complete successfully
//	Expected Result System handles high load without failures or resource leaks
func TestSystemST_STRESS_01_HighLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress tests in short mode")
	}

	if _, err := os.Stat(testutil.AuthTokensPath); os.IsNotExist(err) {
		t.Skipf("Skipping stress tests: auth tokens not found at %s", testutil.AuthTokensPath)
	}

	suite, err := NewSystemTestSuite(t)
	require.NoError(t, err, "Failed to create system test suite")

	err = suite.Setup()
	require.NoError(t, err, "Failed to setup system test environment")

	t.Cleanup(func() {
		if err := suite.Cleanup(); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	})

	// Test high load scenarios
	err = suite.TestHighLoadOperations()
	assert.NoError(t, err, "High load test failed")
}
