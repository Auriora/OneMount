package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/stretchr/testify/require"
)

// TestIT_FS_01_01_Cache_BasicOperations_WorkCorrectly tests various cache operations.
//
//	Test Case ID    IT-FS-01-01
//	Title           Cache Operations
//	Description     Tests various cache operations
//	Preconditions   None
//	Steps           1. Create a filesystem cache
//	                2. Perform operations on the cache (get path, get children, check pointers)
//	                3. Verify the results of each operation
//	Expected Result Cache operations work correctly
//	Notes: This test verifies that the cache operations work correctly.
func TestIT_FS_01_01_Cache_BasicOperations_WorkCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "CacheOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID

		// Step 1: Test basic cache operations

		// Test GetPath operation
		rootInodeByPath, err := fs.GetPath("/", fs.auth)
		assert.NoError(err, "GetPath should not return error")
		assert.NotNil(rootInodeByPath, "Root inode should exist")
		assert.Equal("/", rootInodeByPath.Path(), "Root path should be /")

		// Test GetID operation
		rootInode := fs.GetID(rootID)
		assert.NotNil(rootInode, "Root inode should exist")
		assert.Equal(rootID, rootInode.ID(), "Root inode ID should match")

		// Step 2: Test cache insertion and retrieval

		// Create a test file item
		testFileID := "test-cache-file-id"
		testFileName := "cache_test_file.txt"
		fileItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Size: 1024,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "test-hash",
				},
			},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Insert the file into the cache
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		nodeID := fs.InsertChild(rootID, fileInode)

		// Verify insertion
		assert.NotEqual(uint64(0), nodeID, "Node ID should be assigned")
		assert.Equal(nodeID, fileInode.NodeID(), "Node ID should match inode")

		// Test retrieval by ID
		retrievedInode := fs.GetID(testFileID)
		assert.NotNil(retrievedInode, "File should be retrievable by ID")
		assert.Equal(testFileID, retrievedInode.ID(), "Retrieved inode ID should match")
		assert.Equal(testFileName, retrievedInode.Name(), "Retrieved inode name should match")

		// Test retrieval by NodeID
		retrievedByNodeID := fs.GetNodeID(nodeID)
		assert.NotNil(retrievedByNodeID, "File should be retrievable by NodeID")
		assert.Equal(testFileID, retrievedByNodeID.ID(), "Retrieved inode ID should match")

		// Step 3: Test GetChild operation
		childInode, err := fs.GetChild(rootID, testFileName, fs.auth)
		assert.NoError(err, "GetChild should not return error")
		assert.NotNil(childInode, "Child should be found")
		assert.Equal(testFileID, childInode.ID(), "Child ID should match")

		// Step 4: Test GetChildrenID operation
		children, err := fs.GetChildrenID(rootID, fs.auth)
		assert.NoError(err, "GetChildrenID should not return error")
		assert.NotNil(children, "Children map should not be nil")
		childKey := strings.ToLower(testFileName)
		childInode, exists := children[childKey]
		assert.True(exists, "Children should contain our test file")
		assert.Equal(testFileID, childInode.ID(), "Child in map should have correct ID")

		// Step 5: Test path operations
		expectedPath := "/" + testFileName
		assert.Equal(expectedPath, fileInode.Path(), "File path should be correct")

		// Test that the file can be found by path
		foundInode, err := fs.GetPath(expectedPath, fs.auth)
		assert.NoError(err, "GetPath should find the file")
		assert.NotNil(foundInode, "Found inode should not be nil")
		assert.Equal(testFileID, foundInode.ID(), "Found inode ID should match")

		// Step 6: Test cache cleanup and management

		// Test that cache pointers are working correctly
		assert.Equal(fileInode, fs.GetNodeID(nodeID), "Node ID pointer should be consistent")
		assert.Equal(fileInode, fs.GetID(testFileID), "ID pointer should be consistent")
	})
}

// TestUT_FS_01_02_Cache_SkipsXDGVolumeInfoFromServer verifies that remote
// .xdg-volume-info entries returned by the Graph API are ignored so the local
// virtual file can be used instead.
func TestUT_FS_01_02_Cache_SkipsXDGVolumeInfoFromServer(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "SkipXDGVolumeInfoFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		xdgItem := &graph.DriveItem{
			ID:   "remote-xdg-id",
			Name: ".xdg-volume-info",
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{QuickXorHash: "fakehash=="},
			},
		}
		regularItem := &graph.DriveItem{
			ID:   "regular-file-id",
			Name: "regular.txt",
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{QuickXorHash: "anotherhash=="},
			},
			Size: 42,
		}

		mockClient.AddMockItem("/me/drive/items/"+rootID, &graph.DriveItem{ID: rootID, Name: "root", Folder: &graph.Folder{}})
		mockClient.AddMockItem("/me/drive/items/"+regularItem.ID, regularItem)
		mockClient.AddMockItem("/me/drive/items/"+xdgItem.ID, xdgItem)
		mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{xdgItem, regularItem})

		children, err := fs.GetChildrenID(rootID, fsFixture.Auth)
		assert.NoError(err, "GetChildrenID should not error")
		assert.NotNil(children, "Children map should not be nil")
		if _, exists := children[strings.ToLower(xdgItem.Name)]; exists {
			t.Fatalf("Remote .xdg-volume-info should be ignored but was present in children map")
		}
		assert.Nil(fs.GetID(xdgItem.ID), "Remote .xdg-volume-info should not be cached")
	})
}

// TestGetChildrenIDUsesMetadataStoreWhenOffline ensures cached metadata can satisfy directory listings without Graph access.
func TestGetChildrenIDUsesMetadataStoreWhenOffline(t *testing.T) {
	tempSandbox := filepath.Join(os.TempDir(), "onemount-tests")
	originalSandbox := testutil.TestSandboxDir
	originalTmp := testutil.TestSandboxTmpDir
	originalAuth := testutil.AuthTokensPath
	originalLog := testutil.TestLogPath
	originalGraph := testutil.GraphTestDir
	originalMount := testutil.TestMountPoint
	originalDir := testutil.TestDir
	originalSystemMount := testutil.SystemTestMountPoint
	originalSystemData := testutil.SystemTestDataDir
	originalSystemLog := testutil.SystemTestLogPath

	testutil.TestSandboxDir = tempSandbox
	testutil.TestSandboxTmpDir = filepath.Join(tempSandbox, "tmp")
	testutil.AuthTokensPath = filepath.Join(tempSandbox, ".auth_tokens.json")
	testutil.TestLogPath = filepath.Join(tempSandbox, "logs", "fusefs_tests.log")
	testutil.GraphTestDir = filepath.Join(tempSandbox, "graph_test_dir")
	testutil.TestMountPoint = filepath.Join(testutil.TestSandboxTmpDir, "mount")
	testutil.TestDir = filepath.Join(testutil.TestMountPoint, "onemount_tests")
	testutil.SystemTestMountPoint = filepath.Join(testutil.TestSandboxTmpDir, "system-test-mount")
	testutil.SystemTestDataDir = filepath.Join(tempSandbox, "system-test-data")
	testutil.SystemTestLogPath = filepath.Join(tempSandbox, "logs", "system_tests.log")

	t.Cleanup(func() {
		testutil.TestSandboxDir = originalSandbox
		testutil.TestSandboxTmpDir = originalTmp
		testutil.AuthTokensPath = originalAuth
		testutil.TestLogPath = originalLog
		testutil.GraphTestDir = originalGraph
		testutil.TestMountPoint = originalMount
		testutil.TestDir = originalDir
		testutil.SystemTestMountPoint = originalSystemMount
		testutil.SystemTestDataDir = originalSystemData
		testutil.SystemTestLogPath = originalSystemLog
	})

	if err := helpers.EnsureTestDirectories(); err != nil {
		t.Fatalf("Failed to prepare test directories: %v", err)
	}
	fixture := helpers.SetupFSTestFixture(t, "MetadataChildrenRecoveryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, f interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := f.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", f)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID

		file := helpers.CreateMockFile(fsFixture.MockClient, rootID, "metadata-recovery.txt", "metadata-recovery-id", "hello metadata")
		children, err := fs.GetChildrenID(rootID, fs.auth)
		assert.NoError(err, "Initial metadata fetch should succeed")
		childInode, exists := children[strings.ToLower(file.Name)]
		if !assert.True(exists, "Fetched children should include the mock file") {
			return
		}

		fs.metadata.Delete(childInode.ID())
		parent := fs.GetID(rootID)
		assert.NotNil(parent, "Root inode should still exist")
		parent.mu.Lock()
		parent.children = nil
		parent.subdir = 0
		parent.mu.Unlock()

		graph.SetOperationalOffline(true)
		defer graph.SetOperationalOffline(false)

		restoredChildren, err := fs.GetChildrenID(rootID, fs.auth)
		assert.NoError(err, "GetChildrenID should read from structured metadata while offline")
		restored, ok := restoredChildren[strings.ToLower(file.Name)]
		if !assert.True(ok, "Children map should include entry restored from metadata") {
			return
		}
		assert.Equal(childInode.ID(), restored.ID(), "Restored inode should match original child ID")
	})
}

func TestGetPathUsesMetadataStoreWhenOffline(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "MetadataPathRecoveryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, data interface{}) {
			unitTestFixture, ok := data.(*framework.UnitTestFixture)
			if !ok {
				t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", data)
			}
			fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
			fs := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID

			fileItem := helpers.CreateMockFile(fsFixture.MockClient, rootID, "metadata-path.txt", "metadata-path-id", "path-metadata-content")

			_, err := fs.GetPath("/"+fileItem.Name, fs.auth)
			require.NoError(t, err, "Initial GetPath should succeed online")

			fs.metadata.Delete(fileItem.ID)
			parent := fs.GetID(rootID)
			require.NotNil(t, parent, "Root inode should exist in memory")
			parent.mu.Lock()
			parent.children = nil
			parent.subdir = 0
			parent.mu.Unlock()

			graph.SetOperationalOffline(true)
			defer graph.SetOperationalOffline(false)

			inode, err := fs.GetPath("/"+fileItem.Name, fs.auth)
			require.NoError(t, err, "GetPath should resolve from metadata store while offline")
			require.NotNil(t, inode)
			require.Equal(t, fileItem.ID, inode.ID(), "Offline GetPath should return the same inode")
		})
	})
}
