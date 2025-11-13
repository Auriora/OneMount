package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_STATUS_09_XAttrSupport_TrackedCorrectly tests that extended attribute support is tracked correctly.
//
//	Test Case ID    IT-FS-STATUS-09
//	Title           Extended Attribute Support Tracking
//	Description     Tests that the filesystem correctly tracks whether extended attributes are supported
//	Preconditions   Filesystem mounted with status tracking enabled
//	Steps           1. Check initial xattr support status
//	                2. Create a test inode and update its status
//	                3. Verify xattr support is tracked
//	                4. Verify statistics include xattr support status
//	Expected Result XAttr support status is tracked and reported in statistics
//	Requirements    8.1, 8.4
//	Notes: This test verifies that xattr support tracking works correctly.
func TestIT_FS_STATUS_09_XAttrSupport_TrackedCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "XAttrSupportTrackingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the filesystem from the fixture
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		assert.True(ok, "Expected UnitTestFixture")

		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		filesystem := fsFixture.FS.(*Filesystem)

		// Initially, xattr support should be false (not yet tested)
		filesystem.xattrSupportedM.RLock()
		initialSupport := filesystem.xattrSupported
		filesystem.xattrSupportedM.RUnlock()

		t.Logf("Initial xattr support status: %v", initialSupport)

		// Create a test inode using NewInodeDriveItem
		testItem := &graph.DriveItem{
			ID:   "test-xattr-support",
			Name: "test_file.txt",
			File: &graph.File{},
			Size: 1024,
		}
		testInode := NewInodeDriveItem(testItem)

		// Insert the inode into the filesystem
		filesystem.InsertNodeID(testInode)
		filesystem.InsertChild(filesystem.root, testInode)

		// Update file status, which should trigger xattr operations
		filesystem.updateFileStatus(testInode)

		// Check that xattr support status is now tracked
		filesystem.xattrSupportedM.RLock()
		finalSupport := filesystem.xattrSupported
		filesystem.xattrSupportedM.RUnlock()

		t.Logf("Final xattr support status: %v", finalSupport)

		// Get statistics and verify xattr support is included
		stats, err := filesystem.GetStats()
		assert.NoError(err, "Should get statistics without error")
		assert.NotNil(stats, "Statistics should not be nil")

		t.Logf("Statistics report xattr support: %v", stats.XAttrSupported)

		// Verify that the xattr support status is consistent
		assert.Equal(finalSupport, stats.XAttrSupported,
			"Statistics xattr support should match filesystem status")

		// Verify that GetQuickStats also includes xattr support
		quickStats, err := filesystem.GetQuickStats()
		assert.NoError(err, "Should get quick statistics without error")
		assert.NotNil(quickStats, "Quick statistics should not be nil")

		assert.Equal(finalSupport, quickStats.XAttrSupported,
			"Quick statistics xattr support should match filesystem status")

		t.Log("✓ XAttr support tracking works correctly")
	})
}
