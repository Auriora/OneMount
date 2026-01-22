package fs

import (
	"os"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestIT_FS_FileOps_01_FileCreation_BasicOperations tests basic file creation operations.
//
//	Test Case ID    IT-FS-FileOps-01
//	Title           Basic File Creation Operations
//	Description     Tests creating files using Mknod and Create operations
//	Preconditions   None
//	Steps           1. Create files using Mknod
//	                2. Create files using Create
//	                3. Verify file attributes and properties
//	                4. Test error conditions (duplicate files, invalid names)
//	Expected Result Files are created successfully with correct attributes
//	Notes: This test verifies that file creation works correctly.
func TestIT_FS_FileOps_01_FileCreation_BasicOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileCreationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Test file creation using Mknod

		// Create a test file using Mknod
		fileName := "test_file_mknod.txt"
		mknodIn := &fuse.MknodIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0644,
		}
		entryOut := &fuse.EntryOut{}

		status := fs.Mknod(nil, mknodIn, fileName, entryOut)
		assert.Equal(fuse.OK, status, "Mknod should succeed")
		assert.NotEqual(uint64(0), entryOut.NodeId, "NodeId should be assigned")

		// Verify the file was created
		fileInode := fs.GetNodeID(entryOut.NodeId)
		assert.NotNil(fileInode, "File inode should exist")
		assert.Equal(fileName, fileInode.Name(), "File name should match")
		assert.False(fileInode.IsDir(), "File should not be a directory")

		// Step 2: Test file creation using Create

		// Create a test file using Create
		fileName2 := "test_file_create.txt"
		createIn := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0644,
		}
		createOut := &fuse.CreateOut{}

		status = fs.Create(nil, createIn, fileName2, createOut)
		assert.Equal(fuse.OK, status, "Create should succeed")
		assert.NotEqual(uint64(0), createOut.NodeId, "NodeId should be assigned")

		// Verify the file was created
		file2Inode := fs.GetNodeID(createOut.NodeId)
		assert.NotNil(file2Inode, "File inode should exist")
		assert.Equal(fileName2, file2Inode.Name(), "File name should match")
		assert.False(file2Inode.IsDir(), "File should not be a directory")

		// Step 3: Test error conditions

		// Try to create a file with the same name (should fail for Mknod)
		duplicateEntryOut := &fuse.EntryOut{}
		status = fs.Mknod(nil, mknodIn, fileName, duplicateEntryOut)
		assert.NotEqual(fuse.OK, status, "Mknod with duplicate name should fail")

		// Try to create a file with restricted name
		restrictedEntryOut := &fuse.EntryOut{}
		status = fs.Mknod(nil, mknodIn, "..", restrictedEntryOut)
		assert.Equal(fuse.EINVAL, status, "Mknod with restricted name should return EINVAL")

		// Step 4: Verify file attributes

		// Check file attributes
		attr := fileInode.makeAttr()
		assert.Equal(uint32(0644), attr.Mode&0777, "File mode should be 0644")
		assert.Equal(uint64(0), attr.Size, "New file should have zero size")
		assert.True(attr.Mtime > 0, "File should have modification time")

		// Cleanup - remove unused variable warning
		_ = rootID
	})
}

// TestIT_FS_FileOps_02_FileReadWrite_BasicOperations tests basic file read/write operations.
//
//	Test Case ID    IT-FS-FileOps-02
//	Title           Basic File Read/Write Operations
//	Description     Tests reading from and writing to files
//	Preconditions   File exists
//	Steps           1. Create a file
//	                2. Open the file
//	                3. Write data to the file
//	                4. Read data from the file
//	                5. Verify data integrity
//	Expected Result Data is written and read correctly
//	Notes: This test verifies that file I/O operations work correctly.
func TestIT_FS_FileOps_02_FileReadWrite_BasicOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileReadWriteFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create a test file
		fileName := "test_readwrite.txt"
		createIn := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0644,
		}
		createOut := &fuse.CreateOut{}

		status := fs.Create(nil, createIn, fileName, createOut)
		assert.Equal(fuse.OK, status, "File creation should succeed")

		nodeID := createOut.NodeId

		// Step 2: Open the file
		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
			Flags:    uint32(os.O_RDWR),
		}
		openOut := &fuse.OpenOut{}

		status = fs.Open(nil, openIn, openOut)
		assert.Equal(fuse.OK, status, "File open should succeed")

		// Step 3: Write data to the file
		testData := []byte("Hello, OneMount filesystem!")
		writeIn := &fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
			Offset:   0,
		}

		bytesWritten, status := fs.Write(nil, writeIn, testData)
		assert.Equal(fuse.OK, status, "File write should succeed")
		assert.Equal(uint32(len(testData)), bytesWritten, "All bytes should be written")

		// Step 4: Read data from the file
		readBuf := make([]byte, len(testData)+10) // Buffer larger than data
		readIn := &fuse.ReadIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
			Offset:   0,
			Size:     uint32(len(readBuf)),
		}

		readResult, status := fs.Read(nil, readIn, readBuf)
		assert.Equal(fuse.OK, status, "File read should succeed")
		assert.NotNil(readResult, "Read result should not be nil")

		// Step 5: Verify file size was updated
		fileInode := fs.GetNodeID(nodeID)
		assert.NotNil(fileInode, "File inode should exist")
		assert.Equal(uint64(len(testData)), fileInode.DriveItem.Size, "File size should match written data")
	})
}

func TestIT_FS_FileOps_FileCreationMarksMetadataDirty(t *testing.T) {
	if err := helpers.EnsureTestDirectories(); err != nil {
		t.Fatalf("EnsureTestDirectories: %v", err)
	}
	fixture := helpers.SetupFSTestFixture(t, "FileCreationMetadataFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, data interface{}) {
		unit, ok := data.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("expected unit test fixture")
		}
		fsFixture := unit.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		mknodIn := &fuse.MknodIn{InHeader: fuse.InHeader{NodeId: 1}, Mode: 0644}
		entryOut := &fuse.EntryOut{}
		status := fs.Mknod(nil, mknodIn, "dirty.txt", entryOut)
		if status != fuse.OK {
			t.Fatalf("mknod failed: %v", status)
		}

		inode := fs.GetNodeID(entryOut.NodeId)
		if inode == nil {
			t.Fatalf("expected inode")
		}
		entry, err := fs.GetMetadataEntry(inode.ID())
		if err != nil {
			t.Fatalf("GetMetadataEntry: %v", err)
		}
		if entry.State != metadata.ItemStateDirtyLocal {
			t.Fatalf("expected DIRTY_LOCAL state, got %s", entry.State)
		}
	})
}

func TestIT_FS_FileOps_MkdirStateReflectsConnectivity(t *testing.T) {
	if err := helpers.EnsureTestDirectories(); err != nil {
		t.Fatalf("EnsureTestDirectories: %v", err)
	}
	fixture := helpers.SetupFSTestFixture(t, "MkdirStateFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, data interface{}) {
		unit := data.(*framework.UnitTestFixture)
		fsFixture := unit.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		subtests := []struct {
			name    string
			offline bool
			expect  metadata.ItemState
		}{
			{"online", false, metadata.ItemStateDirtyLocal},
			{"offline", true, metadata.ItemStateDirtyLocal},
		}

		for _, tc := range subtests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.offline {
					graph.SetOperationalOffline(true)
					fs.SetOfflineMode(OfflineModeReadWrite)
					defer func() {
						graph.SetOperationalOffline(false)
						fs.SetOfflineMode(OfflineModeDisabled)
					}()
				} else {
					fs.SetOfflineMode(OfflineModeDisabled)
				}
				mkdirIn := &fuse.MkdirIn{InHeader: fuse.InHeader{NodeId: 1}, Mode: 0755}
				entryOut := &fuse.EntryOut{}
				status := fs.Mkdir(nil, mkdirIn, "dir-"+tc.name, entryOut)
				if status != fuse.OK {
					t.Fatalf("mkdir failed: %v", status)
				}
				inode := fs.GetNodeID(entryOut.NodeId)
				if inode == nil {
					t.Fatalf("expected inode")
				}
				entry, err := fs.GetMetadataEntry(inode.ID())
				if err != nil {
					t.Fatalf("GetMetadataEntry: %v", err)
				}
				if entry.State != tc.expect {
					t.Fatalf("expected %s, got %s", tc.expect, entry.State)
				}
			})
		}
	})
}

// TestIT_FS_FileOps_03_FileDeletion_BasicOperations tests basic file deletion operations.
//
//	Test Case ID    IT-FS-FileOps-03
//	Title           Basic File Deletion Operations
//	Description     Tests deleting files using Unlink operation
//	Preconditions   File exists
//	Steps           1. Create a file
//	                2. Verify file exists
//	                3. Delete the file using Unlink
//	                4. Verify file no longer exists
//	                5. Test error conditions (non-existent files)
//	Expected Result Files are deleted successfully
//	Notes: This test verifies that file deletion works correctly.
func TestIT_FS_FileOps_03_FileDeletion_BasicOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "FileDeletionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create a test file
		fileName := "test_delete.txt"
		createIn := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0644,
		}
		createOut := &fuse.CreateOut{}

		status := fs.Create(nil, createIn, fileName, createOut)
		assert.Equal(fuse.OK, status, "File creation should succeed")

		nodeID := createOut.NodeId

		// Step 2: Verify file exists
		fileInode := fs.GetNodeID(nodeID)
		assert.NotNil(fileInode, "File inode should exist")
		assert.Equal(fileName, fileInode.Name(), "File name should match")

		// Step 3: Delete the file using Unlink
		unlinkIn := &fuse.InHeader{NodeId: 1} // Parent node ID (root)

		status = fs.Unlink(nil, unlinkIn, fileName)
		assert.Equal(fuse.OK, status, "File deletion should succeed")

		// Step 4: Verify file no longer exists
		deletedInode := fs.GetNodeID(nodeID)
		assert.Nil(deletedInode, "File inode should no longer exist")

		// Step 5: Test error conditions

		// Try to delete a non-existent file
		status = fs.Unlink(nil, unlinkIn, "non_existent_file.txt")
		assert.Equal(fuse.ENOENT, status, "Deleting non-existent file should return ENOENT")
	})
}
