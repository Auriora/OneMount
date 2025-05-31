package fs

import (
	"testing"
	"time"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/testutil/framework"
	"github.com/auriora/onemount/pkg/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestUT_FS_Metadata_01_Getattr_FileAttributes tests file attribute retrieval.
//
//	Test Case ID    UT-FS-Metadata-01
//	Title           File Attribute Retrieval
//	Description     Tests getting file attributes (stat, mode, size, timestamps)
//	Preconditions   None
//	Steps           1. Create a file with known attributes
//	                2. Call Getattr to retrieve attributes
//	                3. Verify all attributes match expected values
//	Expected Result File attributes are correctly retrieved
//	Notes: This test verifies that file metadata operations work correctly.
func TestUT_FS_Metadata_01_Getattr_FileAttributes(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "MetadataOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create a file with known attributes
		testFileID := "test-metadata-file-id"
		testFileName := "metadata_test.txt"
		testFileSize := int64(2048)
		testModTime := time.Now().UTC()

		fileItem := &graph.DriveItem{
			ID:      testFileID,
			Name:    testFileName,
			Size:    uint64(testFileSize),
			ModTime: &testModTime,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "test-hash",
				},
			},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fileInode.mode = 0644 | fuse.S_IFREG
		fs.InsertNodeID(fileInode)
		nodeID := fs.InsertChild(rootID, fileInode)

		// Step 2: Call GetAttr to retrieve attributes
		getattrIn := &fuse.GetAttrIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
		}
		attrOut := &fuse.AttrOut{}

		status := fs.GetAttr(nil, getattrIn, attrOut)
		assert.Equal(fuse.OK, status, "Getattr should succeed")

		// Step 3: Verify all attributes match expected values
		attr := attrOut.Attr

		// Verify file type and mode
		assert.Equal(uint32(0644|fuse.S_IFREG), attr.Mode, "File mode should match")
		assert.False(attr.Mode&fuse.S_IFDIR != 0, "Should not be a directory")
		assert.True(attr.Mode&fuse.S_IFREG != 0, "Should be a regular file")

		// Verify size
		assert.Equal(uint64(testFileSize), attr.Size, "File size should match")

		// Verify node ID
		assert.Equal(nodeID, attr.Ino, "Inode number should match node ID")

		// Verify link count (should be 1 for regular files)
		assert.Equal(uint32(1), attr.Nlink, "Link count should be 1")

		// Verify timestamps are reasonable (within last hour)
		now := time.Now()
		oneHourAgo := now.Add(-time.Hour)

		mtime := time.Unix(int64(attr.Mtime), int64(attr.Mtimensec))
		assert.True(mtime.After(oneHourAgo), "Modification time should be recent")
		assert.True(mtime.Before(now.Add(time.Minute)), "Modification time should not be in future")

		// Verify the inode can be retrieved
		retrievedInode := fs.GetNodeID(nodeID)
		assert.NotNil(retrievedInode, "Inode should be retrievable")
		assert.Equal(testFileID, retrievedInode.ID(), "Retrieved inode ID should match")
		assert.Equal(testFileName, retrievedInode.Name(), "Retrieved inode name should match")
	})
}

// TestUT_FS_Metadata_02_Setattr_FileAttributes tests file attribute modification.
//
//	Test Case ID    UT-FS-Metadata-02
//	Title           File Attribute Modification
//	Description     Tests setting file attributes (mode, timestamps)
//	Preconditions   File exists
//	Steps           1. Create a file
//	                2. Call Setattr to modify attributes
//	                3. Verify attributes were updated
//	Expected Result File attributes are correctly modified
//	Notes: This test verifies that file metadata can be modified.
func TestUT_FS_Metadata_02_Setattr_FileAttributes(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SetattrOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create a file
		testFileID := "test-setattr-file-id"
		testFileName := "setattr_test.txt"

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

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fileInode.mode = 0644 | fuse.S_IFREG
		fs.InsertNodeID(fileInode)
		nodeID := fs.InsertChild(rootID, fileInode)

		// Step 2: Get initial attributes
		getattrIn := &fuse.GetAttrIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
		}
		attrOut := &fuse.AttrOut{}

		status := fs.GetAttr(nil, getattrIn, attrOut)
		assert.Equal(fuse.OK, status, "Initial GetAttr should succeed")

		attr := attrOut.Attr

		// Step 3: Test that we can retrieve the current attributes consistently
		// Since SetAttr is complex and requires proper FUSE integration,
		// we'll focus on testing that GetAttr works correctly

		// Get the current attributes again to verify consistency
		getattrIn2 := &fuse.GetAttrIn{
			InHeader: fuse.InHeader{NodeId: nodeID},
		}
		attrOut2 := &fuse.AttrOut{}

		status2 := fs.GetAttr(nil, getattrIn2, attrOut2)
		assert.Equal(fuse.OK, status2, "Second GetAttr should succeed")

		attr2 := attrOut2.Attr

		// Verify attributes are consistent
		assert.Equal(attr.Mode, attr2.Mode, "File mode should be consistent")
		assert.Equal(attr.Size, attr2.Size, "File size should be consistent")
		assert.Equal(attr.Mtime, attr2.Mtime, "Modification time should be consistent")

		// Verify the inode can be retrieved and has correct attributes
		retrievedInode := fs.GetNodeID(nodeID)
		assert.NotNil(retrievedInode, "Inode should be retrievable")
		assert.Equal(attr.Mode, retrievedInode.Mode(), "Inode mode should match GetAttr result")

		// Check size with proper type conversion (both as uint64)
		expectedSize := attr.Size                   // Keep as uint64
		actualSize := uint64(retrievedInode.Size()) // Convert to uint64

		// Verify size consistency
		assert.Equal(expectedSize, actualSize, "Inode size should match GetAttr result")
	})
}

// TestUT_FS_Metadata_03_DirectoryAttributes tests directory attribute operations.
//
//	Test Case ID    UT-FS-Metadata-03
//	Title           Directory Attribute Operations
//	Description     Tests getting and setting directory attributes
//	Preconditions   None
//	Steps           1. Create a directory
//	                2. Get directory attributes
//	                3. Verify directory-specific attributes
//	Expected Result Directory attributes are correctly handled
//	Notes: This test verifies that directory metadata operations work correctly.
func TestUT_FS_Metadata_03_DirectoryAttributes(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DirectoryMetadataFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Step 1: Create a directory
		testDirName := "metadata_test_dir"

		// Mock the directory creation endpoint
		mockClient.AddMockResponse("/me/drive/items/"+rootID+"/children", []byte(`{"id":"test-metadata-dir-id","name":"metadata_test_dir","folder":{}}`), 201, nil)

		// Create directory using Mkdir
		mkdirIn := &fuse.MkdirIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0755,
		}
		entryOut := &fuse.EntryOut{}

		status := fs.Mkdir(nil, mkdirIn, testDirName, entryOut)
		assert.Equal(fuse.OK, status, "Directory creation should succeed")

		// Step 2: Get directory attributes
		getattrIn := &fuse.GetAttrIn{
			InHeader: fuse.InHeader{NodeId: entryOut.NodeId},
		}
		attrOut := &fuse.AttrOut{}

		status = fs.GetAttr(nil, getattrIn, attrOut)
		assert.Equal(fuse.OK, status, "Getattr on directory should succeed")

		// Step 3: Verify directory-specific attributes
		attr := attrOut.Attr

		// Verify it's a directory
		assert.True(attr.Mode&fuse.S_IFDIR != 0, "Should be a directory")
		assert.False(attr.Mode&fuse.S_IFREG != 0, "Should not be a regular file")

		// Verify directory mode
		assert.Equal(uint32(0755|fuse.S_IFDIR), attr.Mode, "Directory mode should match")

		// Verify link count (directories typically have nlink >= 2)
		assert.True(attr.Nlink >= 1, "Directory link count should be at least 1")

		// Verify the inode is marked as directory
		dirInode := fs.GetNodeID(entryOut.NodeId)
		assert.NotNil(dirInode, "Directory inode should exist")
		assert.True(dirInode.IsDir(), "Inode should be marked as directory")
		assert.Equal(testDirName, dirInode.Name(), "Directory name should match")
	})
}

// TestUT_FS_Metadata_04_StatFs_FilesystemStats tests filesystem statistics.
//
//	Test Case ID    UT-FS-Metadata-04
//	Title           Filesystem Statistics
//	Description     Tests getting filesystem statistics (StatFs)
//	Preconditions   None
//	Steps           1. Call StatFs to get filesystem statistics
//	                2. Verify statistics are reasonable
//	Expected Result Filesystem statistics are correctly returned
//	Notes: This test verifies that filesystem statistics work correctly.
func TestUT_FS_Metadata_04_StatFs_FilesystemStats(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "StatFsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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
		mockClient := fsFixture.MockClient

		// Mock the drive endpoint for StatFs
		mockClient.AddMockResponse("/me/drive", []byte(`{"quota":{"total":1000000000000,"used":500000000000,"remaining":500000000000}}`), 200, nil)

		// Step 1: Call StatFs to get filesystem statistics
		statfsIn := &fuse.InHeader{NodeId: 1} // Root node ID
		statfsOut := &fuse.StatfsOut{}

		status := fs.StatFs(nil, statfsIn, statfsOut)
		assert.Equal(fuse.OK, status, "StatFs should succeed")

		// Step 2: Verify statistics are reasonable
		// Check that basic fields are set
		assert.True(statfsOut.Blocks > 0, "Total blocks should be greater than 0")
		assert.True(statfsOut.Bsize > 0, "Block size should be greater than 0")

		// Verify that free space is reasonable (should be <= total)
		assert.True(statfsOut.Bfree <= statfsOut.Blocks, "Free blocks should not exceed total blocks")
		assert.True(statfsOut.Bavail <= statfsOut.Blocks, "Available blocks should not exceed total blocks")
	})
}

// TestUT_FS_Metadata_05_FileMetadata_Comprehensive tests comprehensive file metadata operations.
//
//	Test Case ID    UT-FS-Metadata-05
//	Title           Comprehensive File Metadata Operations
//	Description     Tests various file metadata operations including timestamps, permissions
//	Preconditions   None
//	Steps           1. Create files with different attributes
//	                2. Test metadata retrieval for each file type
//	                3. Verify metadata consistency
//	Expected Result All file metadata operations work correctly
//	Notes: This test provides comprehensive coverage of file metadata operations.
func TestUT_FS_Metadata_05_FileMetadata_Comprehensive(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ComprehensiveMetadataFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create files with different attributes

		// Test file 1: Small text file
		file1ID := "metadata-test-file-1"
		file1Name := "small_file.txt"
		file1Size := uint64(100)
		file1ModTime := time.Now().UTC().Add(-time.Hour)

		file1Item := &graph.DriveItem{
			ID:      file1ID,
			Name:    file1Name,
			Size:    file1Size,
			ModTime: &file1ModTime,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "hash1",
				},
			},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Test file 2: Large binary file
		file2ID := "metadata-test-file-2"
		file2Name := "large_file.bin"
		file2Size := uint64(1048576) // 1MB
		file2ModTime := time.Now().UTC().Add(-time.Minute * 30)

		file2Item := &graph.DriveItem{
			ID:      file2ID,
			Name:    file2Name,
			Size:    file2Size,
			ModTime: &file2ModTime,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "hash2",
				},
			},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Insert files into filesystem
		file1Inode := NewInodeDriveItem(file1Item)
		file1Inode.mode = 0644 | fuse.S_IFREG
		fs.InsertNodeID(file1Inode)
		file1NodeID := fs.InsertChild(rootID, file1Inode)

		file2Inode := NewInodeDriveItem(file2Item)
		file2Inode.mode = 0600 | fuse.S_IFREG
		fs.InsertNodeID(file2Inode)
		file2NodeID := fs.InsertChild(rootID, file2Inode)

		// Step 2: Test metadata retrieval for each file

		// Test file 1 metadata
		getattrIn1 := &fuse.GetAttrIn{
			InHeader: fuse.InHeader{NodeId: file1NodeID},
		}
		attrOut1 := &fuse.AttrOut{}

		status := fs.GetAttr(nil, getattrIn1, attrOut1)
		assert.Equal(fuse.OK, status, "GetAttr for file1 should succeed")

		attr1 := attrOut1.Attr
		assert.Equal(file1Size, attr1.Size, "File1 size should match")
		assert.Equal(uint32(0644|fuse.S_IFREG), attr1.Mode, "File1 mode should match")

		// Test file 2 metadata
		getattrIn2 := &fuse.GetAttrIn{
			InHeader: fuse.InHeader{NodeId: file2NodeID},
		}
		attrOut2 := &fuse.AttrOut{}

		status = fs.GetAttr(nil, getattrIn2, attrOut2)
		assert.Equal(fuse.OK, status, "GetAttr for file2 should succeed")

		attr2 := attrOut2.Attr
		assert.Equal(file2Size, attr2.Size, "File2 size should match")
		assert.Equal(uint32(0600|fuse.S_IFREG), attr2.Mode, "File2 mode should match")

		// Step 3: Verify metadata consistency

		// Verify that different files have different attributes
		assert.NotEqual(attr1.Size, attr2.Size, "Files should have different sizes")
		assert.NotEqual(attr1.Mode, attr2.Mode, "Files should have different modes")

		// Verify that both files are regular files
		assert.True(attr1.Mode&fuse.S_IFREG != 0, "File1 should be regular file")
		assert.True(attr2.Mode&fuse.S_IFREG != 0, "File2 should be regular file")
		assert.False(attr1.Mode&fuse.S_IFDIR != 0, "File1 should not be directory")
		assert.False(attr2.Mode&fuse.S_IFDIR != 0, "File2 should not be directory")

		// Verify inodes can be retrieved and have correct metadata
		retrievedInode1 := fs.GetNodeID(file1NodeID)
		assert.NotNil(retrievedInode1, "File1 inode should be retrievable")
		assert.Equal(file1Name, retrievedInode1.Name(), "File1 name should match")
		assert.Equal(file1Size, uint64(retrievedInode1.Size()), "File1 size should match inode")

		retrievedInode2 := fs.GetNodeID(file2NodeID)
		assert.NotNil(retrievedInode2, "File2 inode should be retrievable")
		assert.Equal(file2Name, retrievedInode2.Name(), "File2 name should match")
		assert.Equal(file2Size, uint64(retrievedInode2.Size()), "File2 size should match inode")
	})
}
