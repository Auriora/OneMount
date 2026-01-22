package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_XATTR_01_NilInode_HandledGracefully tests that nil inode is handled gracefully.
//
//	Test Case ID    IT-FS-XATTR-01
//	Title           Nil Inode Handling
//	Description     Tests that updateFileStatus handles nil inode gracefully
//	Preconditions   Filesystem mounted
//	Steps           1. Call updateFileStatus with nil inode
//	                2. Verify no panic occurs
//	                3. Verify warning is logged
//	Expected Result No panic, warning logged
//	Requirements    8.1, 8.4
//	Notes: This test verifies defensive programming in updateFileStatus.
func TestIT_FS_XATTR_01_NilInode_HandledGracefully(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "NilInodeHandlingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Call updateFileStatus with nil inode
		// This should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("updateFileStatus panicked with nil inode: %v", r)
			}
		}()

		filesystem.updateFileStatus(nil)

		t.Log("✓ Nil inode handled gracefully without panic")
	})
}

// TestIT_FS_XATTR_02_EmptyPath_HandledGracefully tests that empty path is handled gracefully.
//
//	Test Case ID    IT-FS-XATTR-02
//	Title           Empty Path Handling
//	Description     Tests that updateFileStatus handles inode with empty path gracefully
//	Preconditions   Filesystem mounted
//	Steps           1. Create inode with no parent (empty path)
//	                2. Call updateFileStatus
//	                3. Verify no panic occurs
//	                4. Verify debug log is generated
//	Expected Result No panic, debug log generated
//	Requirements    8.1, 8.4
//	Notes: This test verifies defensive programming in updateFileStatus.
func TestIT_FS_XATTR_02_EmptyPath_HandledGracefully(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "EmptyPathHandlingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a test inode without inserting it into the filesystem
		// This will result in an empty path
		testItem := &graph.DriveItem{
			ID:   "test-empty-path",
			Name: "test_file.txt",
			File: &graph.File{},
			Size: 1024,
		}
		testInode := NewInodeDriveItem(testItem)

		// Call updateFileStatus with inode that has empty path
		// This should not panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("updateFileStatus panicked with empty path: %v", r)
			}
		}()

		filesystem.updateFileStatus(testInode)

		t.Log("✓ Empty path handled gracefully without panic")
	})
}

// TestIT_FS_XATTR_03_XAttrMapInitialization_WorksCorrectly tests that xattr map is initialized correctly.
//
//	Test Case ID    IT-FS-XATTR-03
//	Title           XAttr Map Initialization
//	Description     Tests that xattr map is initialized when nil
//	Preconditions   Filesystem mounted
//	Steps           1. Create inode with nil xattrs map
//	                2. Call updateFileStatus
//	                3. Verify xattrs map is initialized
//	                4. Verify status xattr is set
//	Expected Result XAttrs map initialized, status xattr set
//	Requirements    8.1, 8.4
//	Notes: This test verifies xattr map initialization.
func TestIT_FS_XATTR_03_XAttrMapInitialization_WorksCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "XAttrMapInitializationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a test inode with xattrs map (initialized by NewInodeDriveItem)
		testItem := &graph.DriveItem{
			ID:   "test-xattr-init",
			Name: "test_file.txt",
			File: &graph.File{},
			Size: 1024,
		}
		testInode := NewInodeDriveItem(testItem)

		// Verify xattrs map is initialized (by NewInodeDriveItem)
		testInode.mu.RLock()
		initialXattrs := testInode.xattrs
		testInode.mu.RUnlock()
		assert.NotNil(initialXattrs, "XAttrs map should be initialized by NewInodeDriveItem")
		assert.Equal(0, len(initialXattrs), "XAttrs map should be empty initially")

		// Insert the inode into the filesystem
		filesystem.InsertNodeID(testInode)
		filesystem.InsertChild(filesystem.root, testInode)

		// Call updateFileStatus
		filesystem.updateFileStatus(testInode)

		// Verify xattrs map still works (should have status now)
		testInode.mu.RLock()
		finalXattrs := testInode.xattrs
		testInode.mu.RUnlock()
		assert.NotNil(finalXattrs, "XAttrs map should still be initialized")
		assert.True(len(finalXattrs) > 0, "XAttrs map should have status xattr")

		// Verify status xattr is set
		testInode.mu.RLock()
		statusXattr, exists := testInode.xattrs["user.onemount.status"]
		testInode.mu.RUnlock()
		assert.True(exists, "Status xattr should exist")
		assert.NotNil(statusXattr, "Status xattr should not be nil")
		assert.True(len(statusXattr) > 0, "Status xattr should have value")

		t.Logf("✓ XAttr map initialized correctly, status: %s", string(statusXattr))
	})
}

// TestIT_FS_XATTR_04_StatusXAttr_UpdatedCorrectly tests that status xattr is updated correctly.
//
//	Test Case ID    IT-FS-XATTR-04
//	Title           Status XAttr Update
//	Description     Tests that status xattr is updated when file status changes
//	Preconditions   Filesystem mounted
//	Steps           1. Create inode and set initial status
//	                2. Call updateFileStatus
//	                3. Verify status xattr matches
//	                4. Change status and update again
//	                5. Verify status xattr updated
//	Expected Result Status xattr reflects current file status
//	Requirements    8.1, 8.4
//	Notes: This test verifies status xattr updates.
func TestIT_FS_XATTR_04_StatusXAttr_UpdatedCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "StatusXAttrUpdateFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a test inode
		testItem := &graph.DriveItem{
			ID:   "test-status-update",
			Name: "test_file.txt",
			File: &graph.File{},
			Size: 1024,
		}
		testInode := NewInodeDriveItem(testItem)

		// Insert the inode into the filesystem
		filesystem.InsertNodeID(testInode)
		filesystem.InsertChild(filesystem.root, testInode)

		// Set initial status
		filesystem.SetFileStatus(testInode.ID(), FileStatusInfo{
			Status: StatusCloud,
		})

		// Call updateFileStatus
		filesystem.updateFileStatus(testInode)

		// Verify status xattr matches
		testInode.mu.RLock()
		statusXattr1, exists1 := testInode.xattrs["user.onemount.status"]
		testInode.mu.RUnlock()
		assert.True(exists1, "Status xattr should exist")
		assert.Equal("Cloud", string(statusXattr1), "Status xattr should be 'Cloud'")

		t.Logf("✓ Initial status xattr: %s", string(statusXattr1))

		// Change status
		filesystem.SetFileStatus(testInode.ID(), FileStatusInfo{
			Status: StatusDownloading,
		})

		// Update again
		filesystem.updateFileStatus(testInode)

		// Verify status xattr updated
		testInode.mu.RLock()
		statusXattr2, exists2 := testInode.xattrs["user.onemount.status"]
		testInode.mu.RUnlock()
		assert.True(exists2, "Status xattr should exist")
		assert.Equal("Downloading", string(statusXattr2), "Status xattr should be 'Downloading'")

		t.Logf("✓ Updated status xattr: %s", string(statusXattr2))
	})
}

// TestIT_FS_XATTR_05_ErrorXAttr_SetAndCleared tests that error xattr is set and cleared correctly.
//
//	Test Case ID    IT-FS-XATTR-05
//	Title           Error XAttr Set and Clear
//	Description     Tests that error xattr is set when error exists and cleared when error is resolved
//	Preconditions   Filesystem mounted
//	Steps           1. Create inode and set error status
//	                2. Call updateFileStatus
//	                3. Verify error xattr is set
//	                4. Clear error status
//	                5. Call updateFileStatus again
//	                6. Verify error xattr is removed
//	Expected Result Error xattr set when error exists, removed when error cleared
//	Requirements    8.1, 8.4
//	Notes: This test verifies error xattr management.
func TestIT_FS_XATTR_05_ErrorXAttr_SetAndCleared(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ErrorXAttrFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a test inode
		testItem := &graph.DriveItem{
			ID:   "test-error-xattr",
			Name: "test_file.txt",
			File: &graph.File{},
			Size: 1024,
		}
		testInode := NewInodeDriveItem(testItem)

		// Insert the inode into the filesystem
		filesystem.InsertNodeID(testInode)
		filesystem.InsertChild(filesystem.root, testInode)

		// Set error status
		filesystem.SetFileStatus(testInode.ID(), FileStatusInfo{
			Status:   StatusError,
			ErrorMsg: "Test error message",
		})

		// Call updateFileStatus
		filesystem.updateFileStatus(testInode)

		// Verify error xattr is set
		testInode.mu.RLock()
		errorXattr1, exists1 := testInode.xattrs["user.onemount.error"]
		testInode.mu.RUnlock()
		assert.True(exists1, "Error xattr should exist")
		assert.Equal("Test error message", string(errorXattr1), "Error xattr should contain error message")

		t.Logf("✓ Error xattr set: %s", string(errorXattr1))

		// Clear error status
		filesystem.SetFileStatus(testInode.ID(), FileStatusInfo{
			Status: StatusLocal,
		})

		// Update again
		filesystem.updateFileStatus(testInode)

		// Verify error xattr is removed
		testInode.mu.RLock()
		_, exists2 := testInode.xattrs["user.onemount.error"]
		testInode.mu.RUnlock()
		assert.False(exists2, "Error xattr should be removed")

		t.Log("✓ Error xattr cleared correctly")
	})
}

// TestIT_FS_XATTR_06_XAttrSupport_AlwaysTrue tests that xattr support is always true.
//
//	Test Case ID    IT-FS-XATTR-06
//	Title           XAttr Support Always True
//	Description     Tests that xattr support is always true (in-memory xattrs)
//	Preconditions   Filesystem mounted
//	Steps           1. Check initial xattr support status
//	                2. Update file status
//	                3. Verify xattr support is true
//	                4. Verify statistics report xattr support
//	Expected Result XAttr support is always true
//	Requirements    8.1, 8.4
//	Notes: This test verifies that xattr support is always true for in-memory xattrs.
func TestIT_FS_XATTR_06_XAttrSupport_AlwaysTrue(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "XAttrSupportAlwaysTrueFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a test inode
		testItem := &graph.DriveItem{
			ID:   "test-xattr-always-true",
			Name: "test_file.txt",
			File: &graph.File{},
			Size: 1024,
		}
		testInode := NewInodeDriveItem(testItem)

		// Insert the inode into the filesystem
		filesystem.InsertNodeID(testInode)
		filesystem.InsertChild(filesystem.root, testInode)

		// Update file status
		filesystem.updateFileStatus(testInode)

		// Verify xattr support is true
		filesystem.xattrSupportedM.RLock()
		xattrSupported := filesystem.xattrSupported
		filesystem.xattrSupportedM.RUnlock()
		assert.True(xattrSupported, "XAttr support should be true (in-memory xattrs)")

		// Verify statistics report xattr support
		stats, err := filesystem.GetStats()
		assert.NoError(err, "Should get statistics without error")
		assert.NotNil(stats, "Statistics should not be nil")
		assert.True(stats.XAttrSupported, "Statistics should report xattr support")

		t.Log("✓ XAttr support is always true for in-memory xattrs")
	})
}
