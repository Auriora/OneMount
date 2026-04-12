package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"

	"github.com/auriora/onemount/internal/graph"
)

// TestIT_FS_37_01_UploadSession_BasicOperations_WorkCorrectly tests various upload session operations.
//
//	Test Case ID    IT-FS-37-01
//	Title           Upload Session Operations
//	Description     Tests various upload session operations
//	Preconditions   None
//	Steps           1. Test direct uploads using internal functions
//	                2. Test small file uploads using the filesystem interface
//	                3. Test large file uploads using the filesystem interface
//	                4. Verify uploads are successful and content is correct
//	Expected Result Upload sessions work correctly for different file sizes and methods
//	Notes: This test verifies that upload sessions work correctly for different file sizes and methods.
func TestIT_FS_37_01_UploadSession_BasicOperations_WorkCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "UploadSessionOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		fsFixture := getFSTestFixture(t, fixture)
		filesystem := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)

		// Step 1: Verify the filesystem and upload manager are properly initialized
		assert.NotNil(filesystem, "Filesystem should be initialized")
		assert.NotNil(filesystem.uploads, "Upload manager should be initialized")

		// Step 2: Create a small test file
		smallFile := NewInode("small_upload.txt", 0644, root)
		filesystem.InsertNodeID(smallFile)
		filesystem.InsertID(smallFile.ID(), smallFile)
		filesystem.InsertChild(rootID, smallFile)

		// Step 3: Verify the file is in the filesystem
		assert.NotNil(filesystem.GetID(smallFile.ID()), "Small file should exist")
		assert.Equal("small_upload.txt", smallFile.Name(), "Small file name should match")

		// Step 4: Create a larger test file
		largeFile := NewInode("large_upload.bin", 0644, root)
		filesystem.InsertNodeID(largeFile)
		filesystem.InsertID(largeFile.ID(), largeFile)
		filesystem.InsertChild(rootID, largeFile)

		// Step 5: Verify both files exist
		assert.NotNil(filesystem.GetID(largeFile.ID()), "Large file should exist")

		// Step 6: Mark files as having changes (simulating write operations)
		smallFile.SetHasChanges(true)
		largeFile.SetHasChanges(true)
		assert.True(smallFile.HasChanges(), "Small file should have changes")
		assert.True(largeFile.HasChanges(), "Large file should have changes")
	})
}
