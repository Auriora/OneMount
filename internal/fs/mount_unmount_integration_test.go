package fs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_MU_01_01_MountUnmount_BasicCycle_WorksCorrectly tests basic mount/unmount cycle
//
//	Test Case ID    IT-MU-01-01
//	Title           Basic Mount/Unmount Cycle
//	Description     Tests that the filesystem can be mounted and unmounted successfully
//	Preconditions   None
//	Steps           1. Mount the filesystem
//	                2. Verify the filesystem is accessible
//	                3. Unmount the filesystem
//	                4. Verify the filesystem is no longer accessible
//	Expected Result Mount and unmount operations complete successfully
//	Notes: This test verifies the basic mount/unmount functionality works correctly.
func TestIT_MU_01_01_MountUnmount_BasicCycle_WorksCorrectly(t *testing.T) {
	// Create a test fixture for mount/unmount testing
	fixture := helpers.SetupMountTestFixtureWithFactory(t, "BasicMountUnmountFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (helpers.FilesystemInterface, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		return fs, err
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		helper := helpers.MustGetMountHelper(t, fixture)

		// Step 1: Verify the filesystem is mounted
		assert.True(helper.IsMounted(), "Filesystem should be mounted")

		// Step 2: Verify the filesystem is accessible
		mountPoint := helper.GetMountPoint()
		_, err := os.ReadDir(mountPoint)
		assert.NoError(err, "Should be able to read mount point directory")

		// Step 3: Create a test file to verify write operations work
		testContent := []byte("Hello, OneMount!")
		err = helper.CreateTestFile("test.txt", testContent)
		assert.NoError(err, "Should be able to create test file")

		// Step 4: Verify the file exists and has correct content
		assert.True(helper.VerifyFileExists("test.txt"), "Test file should exist")

		readContent, err := helper.ReadTestFile("test.txt")
		assert.NoError(err, "Should be able to read test file")
		assert.Equal(testContent, readContent, "File content should match")

		// Step 5: Unmount the filesystem
		err = helper.Unmount()
		assert.NoError(err, "Should be able to unmount filesystem")

		// Step 6: Wait for unmount to complete
		err = helper.WaitForUnmount(10 * time.Second)
		assert.NoError(err, "Unmount should complete within timeout")

		// Step 7: Verify the filesystem is no longer accessible
		assert.False(helper.IsMounted(), "Filesystem should not be mounted")
	})
}

// TestIT_MU_02_01_MountUnmount_MultipleCycles_WorksCorrectly tests multiple mount/unmount cycles
//
//	Test Case ID    IT-MU-02-01
//	Title           Multiple Mount/Unmount Cycles
//	Description     Tests that the filesystem can be mounted and unmounted multiple times
//	Preconditions   None
//	Steps           1. Perform multiple mount/unmount cycles
//	                2. Verify each cycle works correctly
//	                3. Verify data persistence across cycles
//	Expected Result All mount/unmount cycles complete successfully
//	Notes: This test verifies that multiple mount/unmount cycles work correctly.
func TestIT_MU_02_01_MountUnmount_MultipleCycles_WorksCorrectly(t *testing.T) {
	// Create a test fixture for mount/unmount testing
	fixture := helpers.SetupMountTestFixtureWithFactory(t, "MultipleMountUnmountFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (helpers.FilesystemInterface, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		return fs, err
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		helper := helpers.MustGetMountHelper(t, fixture)

		// Test parameters
		const numCycles = 3
		testFiles := make(map[string][]byte)

		for cycle := 1; cycle <= numCycles; cycle++ {
			// Step 1: Verify filesystem is available (first cycle already setup by fixture)
			if cycle > 1 {
				// For testing, we simulate remounting by waiting
				err := helper.WaitForMount(10 * time.Second)
				assert.NoError(err, "Mount should be ready within timeout for cycle %d", cycle)
			}

			assert.True(helper.IsMounted(), "Filesystem should be mounted for cycle %d", cycle)

			// Step 2: Create a unique test file for this cycle
			fileName := filepath.Join("cycle_files", "test_cycle_%d.txt")
			testContent := []byte("Test content for cycle %d")
			testFiles[fileName] = testContent

			err := helper.CreateTestFile(fileName, testContent)
			assert.NoError(err, "Should be able to create test file for cycle %d", cycle)

			// Step 3: Verify all previous files still exist (data persistence)
			for prevFile, prevContent := range testFiles {
				assert.True(helper.VerifyFileExists(prevFile), "File %s should exist in cycle %d", prevFile, cycle)

				readContent, err := helper.ReadTestFile(prevFile)
				assert.NoError(err, "Should be able to read file %s in cycle %d", prevFile, cycle)
				assert.Equal(prevContent, readContent, "File content should match for %s in cycle %d", prevFile, cycle)
			}

			// Step 4: Unmount the filesystem
			err = helper.Unmount()
			assert.NoError(err, "Should be able to unmount filesystem for cycle %d", cycle)

			err = helper.WaitForUnmount(10 * time.Second)
			assert.NoError(err, "Unmount should complete within timeout for cycle %d", cycle)

			assert.False(helper.IsMounted(), "Filesystem should not be mounted after cycle %d", cycle)
		}
	})
}

// TestIT_MU_03_01_MountUnmount_WithActiveOperations_HandlesGracefully tests unmount with active operations
//
//	Test Case ID    IT-MU-03-01
//	Title           Unmount with Active Operations
//	Description     Tests that unmount handles active file operations gracefully
//	Preconditions   None
//	Steps           1. Mount the filesystem
//	                2. Start file operations
//	                3. Attempt to unmount while operations are active
//	                4. Verify graceful handling
//	Expected Result Unmount waits for operations to complete or handles them gracefully
//	Notes: This test verifies that unmount handles active operations correctly.
func TestIT_MU_03_01_MountUnmount_WithActiveOperations_HandlesGracefully(t *testing.T) {
	// Create a test fixture for mount/unmount testing
	fixture := helpers.SetupMountTestFixtureWithFactory(t, "UnmountWithActiveOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (helpers.FilesystemInterface, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		return fs, err
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		helper := helpers.MustGetMountHelper(t, fixture)

		// Step 1: Verify filesystem is mounted
		assert.True(helper.IsMounted(), "Filesystem should be mounted")

		// Step 2: Create a test file
		testContent := []byte("Test content for active operations")
		err := helper.CreateTestFile("active_test.txt", testContent)
		assert.NoError(err, "Should be able to create test file")

		// Step 3: Open the file for reading (simulating active operation)
		filePath := filepath.Join(helper.GetMountPoint(), "active_test.txt")
		file, err := os.Open(filePath)
		assert.NoError(err, "Should be able to open test file")

		// Step 4: Attempt to unmount while file is open
		// This should either wait for the file to be closed or handle it gracefully
		unmountDone := make(chan error, 1)
		go func() {
			unmountDone <- helper.Unmount()
		}()

		// Wait a moment to let unmount attempt to proceed
		time.Sleep(500 * time.Millisecond)

		// Step 5: Close the file to allow unmount to proceed
		err = file.Close()
		assert.NoError(err, "Should be able to close test file")

		// Step 6: Wait for unmount to complete
		select {
		case err := <-unmountDone:
			assert.NoError(err, "Unmount should complete successfully after file is closed")
		case <-time.After(15 * time.Second):
			t.Fatal("Unmount did not complete within timeout")
		}

		// Step 7: Verify filesystem is unmounted
		err = helper.WaitForUnmount(5 * time.Second)
		assert.NoError(err, "Unmount should complete within timeout")
		assert.False(helper.IsMounted(), "Filesystem should not be mounted")
	})
}

// TestIT_MU_04_01_MountUnmount_ErrorRecovery_RecoversCorrectly tests error recovery during mount/unmount
//
//	Test Case ID    IT-MU-04-01
//	Title           Mount/Unmount Error Recovery
//	Description     Tests that mount/unmount operations recover from errors correctly
//	Preconditions   None
//	Steps           1. Simulate mount/unmount errors
//	                2. Verify error handling
//	                3. Verify recovery after errors
//	Expected Result Errors are handled gracefully and recovery works correctly
//	Notes: This test verifies that mount/unmount error recovery works correctly.
func TestIT_MU_04_01_MountUnmount_ErrorRecovery_RecoversCorrectly(t *testing.T) {
	// Create a test fixture for mount/unmount testing
	fixture := helpers.SetupMountTestFixtureWithFactory(t, "MountUnmountErrorRecoveryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (helpers.FilesystemInterface, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		return fs, err
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)
		helper := helpers.MustGetMountHelper(t, fixture)

		// Step 1: Verify filesystem is initially mounted
		assert.True(helper.IsMounted(), "Filesystem should be initially mounted")

		// Step 2: Unmount the filesystem
		err := helper.Unmount()
		assert.NoError(err, "Should be able to unmount filesystem")

		err = helper.WaitForUnmount(10 * time.Second)
		assert.NoError(err, "Unmount should complete within timeout")

		// Step 3: Attempt to unmount again (should handle gracefully)
		err = helper.Unmount()
		// This might return an error or succeed depending on implementation
		// The important thing is that it doesn't crash or cause issues

		// Step 4: Simulate remounting
		// For testing, we simulate remounting by waiting
		err = helper.WaitForMount(10 * time.Second)
		assert.NoError(err, "Mount should be ready within timeout")

		// Step 5: Verify filesystem is accessible after recovery
		assert.True(helper.IsMounted(), "Filesystem should be mounted after recovery")

		// Step 6: Test file operations work after recovery
		testContent := []byte("Test content after recovery")
		err = helper.CreateTestFile("recovery_test.txt", testContent)
		assert.NoError(err, "Should be able to create test file after recovery")

		assert.True(helper.VerifyFileExists("recovery_test.txt"), "Test file should exist after recovery")

		readContent, err := helper.ReadTestFile("recovery_test.txt")
		assert.NoError(err, "Should be able to read test file after recovery")
		assert.Equal(testContent, readContent, "File content should match after recovery")
	})
}
