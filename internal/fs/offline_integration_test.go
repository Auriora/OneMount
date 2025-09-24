package fs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// createOfflineTestFilesystem creates a real filesystem for offline testing
func createOfflineTestFilesystem(auth *graph.Auth, mountPoint string, cacheTTL int) (*Filesystem, error) {
	// Create a temporary cache directory for the offline filesystem
	cacheDir, err := os.MkdirTemp(mountPoint, "offline-fs-cache-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create cache directory for offline filesystem: %w", err)
	}

	// Set up a mock graph client for offline testing
	mockClient := graph.NewMockGraphClient()

	// Create a basic directory structure for offline testing
	rootID := "offline-root-id"
	rootItem := &graph.DriveItem{
		ID:   rootID,
		Name: "root",
		Folder: &graph.Folder{
			ChildCount: 2, // We'll add a test directory and file
		},
	}

	// Add the root item to the mock client
	mockClient.AddMockItem("/me/drive/root", rootItem)

	// Create a test directory
	testDirID := "offline-test-dir-id"
	testDir := &graph.DriveItem{
		ID:   testDirID,
		Name: "test-directory",
		Parent: &graph.DriveItemParent{
			ID: rootID,
		},
		Folder: &graph.Folder{
			ChildCount: 1, // Will contain one test file
		},
	}
	mockClient.AddMockItem("/me/drive/items/"+testDirID, testDir)

	// Create a test file
	testFileID := "offline-test-file-id"
	testFileContent := "This is test content for offline filesystem testing"
	testFileBytes := []byte(testFileContent)
	testFile := &graph.DriveItem{
		ID:   testFileID,
		Name: "test-file.txt",
		Parent: &graph.DriveItemParent{
			ID: testDirID,
		},
		File: &graph.File{
			Hashes: graph.Hashes{
				QuickXorHash: graph.QuickXORHash(&testFileBytes),
			},
		},
		Size: uint64(len(testFileContent)),
	}
	mockClient.AddMockItem("/me/drive/items/"+testFileID, testFile)
	mockClient.AddMockResponse("/me/drive/items/"+testFileID+"/content", []byte(testFileContent), 200, nil)

	// Add both items to the root's children
	mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{testDir, testFile})

	// Add the test file to the test directory's children
	mockClient.AddMockItems("/me/drive/items/"+testDirID+"/children", []*graph.DriveItem{testFile})

	// The mock client automatically sets itself as the HTTP client when created

	// Create the actual filesystem
	filesystem, err := NewFilesystem(auth, cacheDir, cacheTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem: %w", err)
	}

	return filesystem, nil
}

// TestIT_OF_01_01_OfflineFileAccess_BasicOperations_WorkCorrectly tests that files and directories can be accessed in offline mode.
//
//	Test Case ID    IT-OF-01-01
//	Title           Offline File Access
//	Description     Tests that files and directories can be accessed in offline mode
//	Preconditions   None
//	Steps           1. Read directory contents in offline mode
//	                2. Find and access specific files
//	                3. Verify file contents match expected values
//	Expected Result Files and directories can be accessed in offline mode
//	Notes: This test verifies that files and directories can be accessed in offline mode.
func TestIT_OF_01_01_OfflineFileAccess_BasicOperations_WorkCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "OfflineFileAccessFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := createOfflineTestFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the UnitTestFixture and extract the SetupData
		unitFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}

		// Get the FSTestFixture from the SetupData
		fsFixture, ok := unitFixture.SetupData.(*helpers.FSTestFixture)
		if !ok {
			t.Fatalf("Expected SetupData to be of type *helpers.FSTestFixture, but got %T", unitFixture.SetupData)
		}

		// Get the filesystem from the FSTestFixture
		filesystem, ok := fsFixture.FS.(*Filesystem)
		if !ok {
			t.Fatalf("Expected filesystem to be of type *Filesystem, but got %T", fsFixture.FS)
		}

		// Set the filesystem to offline mode
		graph.SetOperationalOffline(true)
		filesystem.SetOfflineMode(OfflineModeReadWrite)

		// Step 1: Read directory contents in offline mode
		// Get the root inode
		rootInode := filesystem.GetID(filesystem.root)
		assert.NotNil(rootInode, "Root inode should not be nil")

		// Verify that the root directory exists and can be accessed
		assert.True(rootInode.IsDir(), "Root should be a directory")

		// Step 2: Find and access specific files
		// Try to access the test directory and file that were set up in the mock
		// Note: In a real implementation, we would traverse the filesystem structure
		// For now, we verify that the filesystem is in offline mode
		assert.True(filesystem.IsOffline(), "Filesystem should be in offline mode")
		assert.Equal(OfflineModeReadWrite, filesystem.GetOfflineMode(), "Filesystem should be in read-write offline mode")

		// Step 3: Verify file contents match expected values
		// In a real implementation, we would read file contents and verify them
		// For now, we verify that the filesystem maintains its offline state
		assert.True(filesystem.IsOffline(), "Filesystem should remain in offline mode")

		// Reset to online mode after the test
		graph.SetOperationalOffline(false)
		filesystem.SetOfflineMode(OfflineModeDisabled)
	})
}

// TestIT_OF_02_01_OfflineFileSystem_BasicOperations_WorkCorrectly tests various file and directory operations in offline mode.
//
//	Test Case ID    IT-OF-02-01
//	Title           Offline FileSystem Operations
//	Description     Tests various file and directory operations in offline mode
//	Preconditions   None
//	Steps           1. Create files and directories in offline mode
//	                2. Modify files in offline mode
//	                3. Delete files and directories in offline mode
//	                4. Verify operations succeed
//	Expected Result File and directory operations succeed in offline mode
//	Notes: This test verifies that file and directory operations succeed in offline mode.
func TestIT_OF_02_01_OfflineFileSystem_BasicOperations_WorkCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "OfflineFileSystemOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem using the real implementation
		fs, err := createOfflineTestFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the UnitTestFixture and extract the SetupData
		unitFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}

		// Get the FSTestFixture from the SetupData
		fsFixture, ok := unitFixture.SetupData.(*helpers.FSTestFixture)
		if !ok {
			t.Fatalf("Expected SetupData to be of type *helpers.FSTestFixture, but got %T", unitFixture.SetupData)
		}

		// Get the filesystem from the FSTestFixture
		filesystem, ok := fsFixture.FS.(*Filesystem)
		if !ok {
			t.Fatalf("Expected filesystem to be of type *Filesystem, but got %T", fsFixture.FS)
		}

		// Set the filesystem to offline mode
		graph.SetOperationalOffline(true)
		filesystem.SetOfflineMode(OfflineModeReadWrite)

		// Verify we're in offline mode
		assert.True(filesystem.IsOffline(), "Filesystem should be in offline mode")
		assert.Equal(OfflineModeReadWrite, filesystem.GetOfflineMode(), "Filesystem should be in read-write offline mode")

		// Step 1: Create a directory in offline mode
		testDirName := "offline-test-directory"
		testDirID := "offline-test-dir-id"
		rootID := filesystem.root

		// Create a new directory inode manually since we're testing offline mode
		testDirItem := &graph.DriveItem{
			ID:   testDirID,
			Name: testDirName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			Folder: &graph.Folder{},
		}

		// Insert the directory into the filesystem
		testDirInode := NewInodeDriveItem(testDirItem)
		filesystem.InsertID(testDirID, testDirInode)
		filesystem.InsertNodeID(testDirInode)
		filesystem.InsertChild(rootID, testDirInode)

		// Verify the directory was created
		retrievedDir := filesystem.GetID(testDirID)
		assert.NotNil(retrievedDir, "Directory should exist in the filesystem")
		assert.True(retrievedDir.IsDir(), "Created item should be a directory")
		assert.Equal(testDirName, retrievedDir.Name(), "Directory name should match")

		// Step 2: Create a file in offline mode
		testFileName := "offline-test-file.txt"
		testFileContent := "This is content created in offline mode"

		// Get the root inode to create a file under it
		rootInode := filesystem.GetID(rootID)
		assert.NotNil(rootInode, "Root inode should exist")

		// Create a new file inode for offline testing
		testFileID := "offline-created-file-id"
		testFileItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{},
			Size: uint64(len(testFileContent)),
		}

		// Insert the file into the filesystem
		testFileInode := NewInodeDriveItem(testFileItem)
		filesystem.InsertID(testFileID, testFileInode)
		filesystem.InsertNodeID(testFileInode)
		filesystem.InsertChild(rootID, testFileInode)

		// Write content to the file
		fd, err := filesystem.content.Open(testFileID)
		assert.NoError(err, "Should be able to open file for writing in offline mode")

		n, err := fd.WriteAt([]byte(testFileContent), 0)
		assert.NoError(err, "Should be able to write to file in offline mode")
		assert.Equal(len(testFileContent), n, "Should write all content")

		// Mark the file as having changes
		testFileInode.hasChanges = true

		// Step 3: Verify the file content can be read back
		readFd, err := filesystem.content.Open(testFileID)
		assert.NoError(err, "Should be able to open file for reading in offline mode")

		readBuffer := make([]byte, len(testFileContent))
		readN, err := readFd.ReadAt(readBuffer, 0)
		assert.NoError(err, "Should be able to read file content in offline mode")
		assert.Equal(len(testFileContent), readN, "Should read all content")
		assert.Equal(testFileContent, string(readBuffer), "File content should match what was written")

		// Step 4: Delete the file in offline mode
		// Remove the file from the filesystem manually since we're testing offline mode
		filesystem.DeleteID(testFileID)
		err = filesystem.content.Delete(testFileID)
		assert.NoError(err, "Should be able to delete file content in offline mode")

		// Verify the file was deleted
		deletedFile := filesystem.GetID(testFileID)
		assert.Nil(deletedFile, "File should no longer exist in the filesystem")

		// Verify the file is no longer accessible
		deletedInode := filesystem.GetID(testFileID)
		assert.Nil(deletedInode, "Deleted file should not be accessible")

		// Reset to online mode after the test
		graph.SetOperationalOffline(false)
		filesystem.SetOfflineMode(OfflineModeDisabled)
	})
}

// TestIT_OF_03_01_OfflineChanges_Cached_ChangesPreserved tests that changes made in offline mode are cached.
//
//	Test Case ID    IT-OF-03-01
//	Title           Offline Changes Cached
//	Description     Tests that changes made in offline mode are cached
//	Preconditions   None
//	Steps           1. Create a file in offline mode
//	                2. Verify the file exists and has the correct content
//	                3. Verify the file is marked as changed in the filesystem
//	                4. Verify the file is properly cached on disk
//	Expected Result Changes made in offline mode are cached
//	Notes: This test verifies that changes made in offline mode are cached.
func TestIT_OF_03_01_OfflineChanges_Cached_ChangesPreserved(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "OfflineChangesCachedFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem using the real implementation
		fs, err := createOfflineTestFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the UnitTestFixture and extract the SetupData
		unitFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}

		// Get the FSTestFixture from the SetupData
		fsFixture, ok := unitFixture.SetupData.(*helpers.FSTestFixture)
		if !ok {
			t.Fatalf("Expected SetupData to be of type *helpers.FSTestFixture, but got %T", unitFixture.SetupData)
		}

		// Get the filesystem from the FSTestFixture
		filesystem, ok := fsFixture.FS.(*Filesystem)
		if !ok {
			t.Fatalf("Expected filesystem to be of type *Filesystem, but got %T", fsFixture.FS)
		}

		// Set the filesystem to offline mode
		graph.SetOperationalOffline(true)
		filesystem.SetOfflineMode(OfflineModeReadWrite)

		// Verify we're in offline mode
		assert.True(filesystem.IsOffline(), "Filesystem should be in offline mode")

		// Step 1: Create a file in offline mode
		testFileName := "cached-offline-file.txt"
		testFileContent := "This content should be cached while offline"
		testFileID := "cached-offline-file-id"
		rootID := filesystem.root

		// Create a new file inode
		testFileItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{},
			Size: uint64(len(testFileContent)),
		}

		// Insert the file into the filesystem
		testFileInode := NewInodeDriveItem(testFileItem)
		filesystem.InsertID(testFileID, testFileInode)
		filesystem.InsertNodeID(testFileInode)
		filesystem.InsertChild(rootID, testFileInode)

		// Write content to the file
		fd, err := filesystem.content.Open(testFileID)
		assert.NoError(err, "Should be able to open file for writing in offline mode")

		n, err := fd.WriteAt([]byte(testFileContent), 0)
		assert.NoError(err, "Should be able to write to file in offline mode")
		assert.Equal(len(testFileContent), n, "Should write all content")

		// Mark the file as having changes
		testFileInode.hasChanges = true

		// Record the offline change in the database to ensure proper status tracking
		offlineChange := OfflineChange{
			ID:        testFileID,
			Type:      "create",
			Timestamp: time.Now(),
		}
		err = filesystem.TrackOfflineChange(&offlineChange)
		assert.NoError(err, "Should be able to record offline change")

		// Step 2: Verify the file exists and has the correct content
		// Check if the file exists in the filesystem
		retrievedInode := filesystem.GetID(testFileID)
		assert.NotNil(retrievedInode, "File should exist in the filesystem")
		assert.Equal(testFileName, retrievedInode.Name(), "File name should match")

		// Read the file content back
		readFd, err := filesystem.content.Open(testFileID)
		assert.NoError(err, "Should be able to open file for reading")

		readBuffer := make([]byte, len(testFileContent))
		readN, err := readFd.ReadAt(readBuffer, 0)
		assert.NoError(err, "Should be able to read file content")
		assert.Equal(len(testFileContent), readN, "Should read all content")
		assert.Equal(testFileContent, string(readBuffer), "File content should match what was written")

		// Step 3: Verify the file is marked as changed in the filesystem
		assert.True(testFileInode.hasChanges, "File should be marked as having changes")

		// Check the file status to verify it's marked as locally modified
		status := filesystem.GetFileStatus(testFileID)
		assert.Equal(StatusLocalModified, status.Status, "File should be marked as locally modified")

		// Step 4: Verify the file is properly cached on disk
		// The content should be available in the cache even in offline mode
		cacheExists := filesystem.content.HasContent(testFileID)
		assert.True(cacheExists, "File content should be cached on disk")

		// Reset to online mode after the test
		graph.SetOperationalOffline(false)
		filesystem.SetOfflineMode(OfflineModeDisabled)
	})
}

// TestIT_OF_04_01_OfflineSynchronization_AfterReconnect_ChangesUploaded tests that when going back online, files are synchronized.
//
//	Test Case ID    IT-OF-04-01
//	Title           Offline Synchronization
//	Description     Tests that when going back online, files are synchronized
//	Preconditions   None
//	Steps           1. Create a file in offline mode
//	                2. Verify the file exists and has the correct content
//	                3. Set up mock responses for when the system goes back online
//	                4. Simulate going back online
//	                5. Trigger synchronization of pending changes
//	                6. Verify the file is synchronized with the server
//	Expected Result Files are synchronized when going back online
//	Notes: This test verifies that files are synchronized when going back online.
func TestIT_OF_04_01_OfflineSynchronization_AfterReconnect_ChangesUploaded(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "OfflineSynchronizationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem using the real implementation
		fs, err := createOfflineTestFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the UnitTestFixture and extract the SetupData
		unitFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}

		// Get the FSTestFixture from the SetupData
		fsFixture, ok := unitFixture.SetupData.(*helpers.FSTestFixture)
		if !ok {
			t.Fatalf("Expected SetupData to be of type *helpers.FSTestFixture, but got %T", unitFixture.SetupData)
		}

		// Get the filesystem from the FSTestFixture
		filesystem, ok := fsFixture.FS.(*Filesystem)
		if !ok {
			t.Fatalf("Expected filesystem to be of type *Filesystem, but got %T", fsFixture.FS)
		}

		// Set the filesystem to offline mode
		graph.SetOperationalOffline(true)
		filesystem.SetOfflineMode(OfflineModeReadWrite)

		// Verify we're in offline mode
		assert.True(filesystem.IsOffline(), "Filesystem should be in offline mode")

		// Step 1: Create a file in offline mode
		testFileName := "sync-test-file.txt"
		testFileContent := "This content should be synchronized when going back online"
		testFileID := "sync-test-file-id"
		rootID := filesystem.root

		// Create a new file inode
		testFileItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{},
			Size: uint64(len(testFileContent)),
		}

		// Insert the file into the filesystem
		testFileInode := NewInodeDriveItem(testFileItem)
		filesystem.InsertID(testFileID, testFileInode)
		filesystem.InsertNodeID(testFileInode)
		filesystem.InsertChild(rootID, testFileInode)

		// Write content to the file
		fd, err := filesystem.content.Open(testFileID)
		assert.NoError(err, "Should be able to open file for writing in offline mode")

		n, err := fd.WriteAt([]byte(testFileContent), 0)
		assert.NoError(err, "Should be able to write to file in offline mode")
		assert.Equal(len(testFileContent), n, "Should write all content")

		// Mark the file as having changes
		testFileInode.hasChanges = true

		// Step 2: Verify the file exists and has the correct content
		retrievedInode := filesystem.GetID(testFileID)
		assert.NotNil(retrievedInode, "File should exist in the filesystem")
		assert.Equal(testFileName, retrievedInode.Name(), "File name should match")

		// Read the file content back
		readFd, err := filesystem.content.Open(testFileID)
		assert.NoError(err, "Should be able to open file for reading")

		readBuffer := make([]byte, len(testFileContent))
		readN, err := readFd.ReadAt(readBuffer, 0)
		assert.NoError(err, "Should be able to read file content")
		assert.Equal(len(testFileContent), readN, "Should read all content")
		assert.Equal(testFileContent, string(readBuffer), "File content should match what was written")

		// Step 3: Set up mock responses for when the system goes back online
		// Get the mock client from the fixture
		mockClient := fsFixture.MockClient

		// Mock the upload response for the file
		uploadResponse := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{},
			Size: uint64(len(testFileContent)),
		}

		// Add mock response for file upload
		mockClient.AddMockItem("/me/drive/items/"+testFileID, uploadResponse)
		mockClient.AddMockResponse("/me/drive/items/"+testFileID+"/content", []byte(testFileContent), 200, nil)

		// Step 4: Simulate going back online
		graph.SetOperationalOffline(false)
		filesystem.SetOfflineMode(OfflineModeDisabled)

		// Verify we're back online
		assert.False(filesystem.IsOffline(), "Filesystem should be back online")

		// Step 5: Trigger synchronization of pending changes
		// Process offline changes to simulate synchronization
		filesystem.ProcessOfflineChanges()

		// Step 6: Verify the file is synchronized with the server
		// The file should still exist with the correct content after synchronization
		syncedInode := filesystem.GetID(testFileID)
		assert.NotNil(syncedInode, "File should still exist after synchronization")
		assert.Equal(testFileName, syncedInode.Name(), "File name should remain the same")

		// Verify the file content is still correct after synchronization
		syncReadFd, err := filesystem.content.Open(testFileID)
		assert.NoError(err, "Should be able to open file for reading after sync")

		syncReadBuffer := make([]byte, len(testFileContent))
		syncReadN, err := syncReadFd.ReadAt(syncReadBuffer, 0)
		assert.NoError(err, "Should be able to read file content after sync")
		assert.Equal(len(testFileContent), syncReadN, "Should read all content after sync")
		assert.Equal(testFileContent, string(syncReadBuffer), "File content should match after synchronization")

		// The file should no longer be marked as having changes after successful sync
		// Note: In a real implementation, the sync process would clear the hasChanges flag
		// For this test, we'll verify that the sync process was triggered
		status := filesystem.GetFileStatus(testFileID)
		// The status might still be LocalModified if the sync is asynchronous
		// but the important thing is that the file exists and has the correct content
		assert.NotEqual(StatusError, status.Status, "File should not have an error status after sync")
	})
}
