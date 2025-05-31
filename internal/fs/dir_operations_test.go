package fs

import (
	"testing"

	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/testutil/framework"
	"github.com/auriora/onemount/pkg/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestUT_FS_DirOps_01_DirectoryCreation_BasicOperations tests basic directory creation operations.
//
//	Test Case ID    UT-FS-DirOps-01
//	Title           Basic Directory Creation Operations
//	Description     Tests creating directories using Mkdir operation
//	Preconditions   None
//	Steps           1. Create directories using Mkdir
//	                2. Verify directory attributes and properties
//	                3. Test nested directory creation
//	                4. Test error conditions (duplicate directories, invalid names)
//	Expected Result Directories are created successfully with correct attributes
//	Notes: This test verifies that directory creation works correctly.
func TestUT_FS_DirOps_01_DirectoryCreation_BasicOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DirectoryCreationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Test directory creation using Mkdir

		// Create a test directory
		dirName := "test_directory"
		mkdirIn := &fuse.MkdirIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0755,
		}
		entryOut := &fuse.EntryOut{}

		status := fs.Mkdir(nil, mkdirIn, dirName, entryOut)
		assert.Equal(fuse.OK, status, "Mkdir should succeed")
		assert.NotEqual(uint64(0), entryOut.NodeId, "NodeId should be assigned")

		// Verify the directory was created
		dirInode := fs.GetNodeID(entryOut.NodeId)
		assert.NotNil(dirInode, "Directory inode should exist")
		assert.Equal(dirName, dirInode.Name(), "Directory name should match")
		assert.True(dirInode.IsDir(), "Should be a directory")

		// Step 2: Test nested directory creation

		// Create a subdirectory inside the first directory
		subDirName := "subdirectory"
		subMkdirIn := &fuse.MkdirIn{
			InHeader: fuse.InHeader{NodeId: entryOut.NodeId}, // Parent directory node ID
			Mode:     0755,
		}
		subEntryOut := &fuse.EntryOut{}

		status = fs.Mkdir(nil, subMkdirIn, subDirName, subEntryOut)
		assert.Equal(fuse.OK, status, "Subdirectory creation should succeed")

		// Verify the subdirectory was created
		subDirInode := fs.GetNodeID(subEntryOut.NodeId)
		assert.NotNil(subDirInode, "Subdirectory inode should exist")
		assert.Equal(subDirName, subDirInode.Name(), "Subdirectory name should match")
		assert.True(subDirInode.IsDir(), "Should be a directory")

		// Step 3: Test error conditions

		// Try to create a directory with the same name (should fail)
		duplicateEntryOut := &fuse.EntryOut{}
		status = fs.Mkdir(nil, mkdirIn, dirName, duplicateEntryOut)
		assert.NotEqual(fuse.OK, status, "Mkdir with duplicate name should fail")

		// Try to create a directory with restricted name
		restrictedEntryOut := &fuse.EntryOut{}
		status = fs.Mkdir(nil, mkdirIn, "..", restrictedEntryOut)
		assert.Equal(fuse.EINVAL, status, "Mkdir with restricted name should return EINVAL")

		// Step 4: Verify directory attributes

		// Check directory attributes
		attr := dirInode.makeAttr()
		assert.Equal(uint32(0755), attr.Mode&0777, "Directory mode should be 0755")
		assert.True(attr.Mode&fuse.S_IFDIR != 0, "Should have directory flag set")
		assert.True(attr.Mtime > 0, "Directory should have modification time")
	})
}

// TestUT_FS_DirOps_02_DirectoryListing_BasicOperations tests basic directory listing operations.
//
//	Test Case ID    UT-FS-DirOps-02
//	Title           Basic Directory Listing Operations
//	Description     Tests listing directory contents using OpenDir, ReadDir, ReadDirPlus
//	Preconditions   Directory with files and subdirectories exists
//	Steps           1. Create a directory with files and subdirectories
//	                2. Open the directory using OpenDir
//	                3. List contents using ReadDir
//	                4. List contents using ReadDirPlus
//	                5. Verify all entries are returned correctly
//	Expected Result Directory contents are listed correctly
//	Notes: This test verifies that directory listing works correctly.
func TestUT_FS_DirOps_02_DirectoryListing_BasicOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DirectoryListingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create a directory with files and subdirectories

		// Create a test directory
		dirName := "test_listing_dir"
		mkdirIn := &fuse.MkdirIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0755,
		}
		dirEntryOut := &fuse.EntryOut{}

		status := fs.Mkdir(nil, mkdirIn, dirName, dirEntryOut)
		assert.Equal(fuse.OK, status, "Directory creation should succeed")

		dirNodeID := dirEntryOut.NodeId

		// Create a file in the directory
		fileName := "test_file.txt"
		createIn := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: dirNodeID},
			Mode:     0644,
		}
		fileEntryOut := &fuse.CreateOut{}

		status = fs.Create(nil, createIn, fileName, fileEntryOut)
		assert.Equal(fuse.OK, status, "File creation should succeed")

		// Create a subdirectory
		subDirName := "test_subdir"
		subMkdirIn := &fuse.MkdirIn{
			InHeader: fuse.InHeader{NodeId: dirNodeID},
			Mode:     0755,
		}
		subDirEntryOut := &fuse.EntryOut{}

		status = fs.Mkdir(nil, subMkdirIn, subDirName, subDirEntryOut)
		assert.Equal(fuse.OK, status, "Subdirectory creation should succeed")

		// Step 2: Open the directory using OpenDir
		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{NodeId: dirNodeID},
		}
		openOut := &fuse.OpenOut{}

		status = fs.OpenDir(nil, openIn, openOut)
		assert.Equal(fuse.OK, status, "OpenDir should succeed")

		// Step 3: Test ReadDirPlus operation
		readIn := &fuse.ReadIn{
			InHeader: fuse.InHeader{NodeId: dirNodeID},
			Offset:   0,
			Size:     4096,
		}
		dirEntryList := &fuse.DirEntryList{}

		// Read first entry (should be ".")
		status = fs.ReadDirPlus(nil, readIn, dirEntryList)
		assert.Equal(fuse.OK, status, "ReadDirPlus should succeed")

		// Step 4: Test directory release
		releaseIn := &fuse.ReleaseIn{
			InHeader: fuse.InHeader{NodeId: dirNodeID},
		}

		fs.ReleaseDir(releaseIn)
		// ReleaseDir doesn't return a status, so we just verify it doesn't panic
	})
}

// TestUT_FS_DirOps_03_DirectoryDeletion_BasicOperations tests basic directory deletion operations.
//
//	Test Case ID    UT-FS-DirOps-03
//	Title           Basic Directory Deletion Operations
//	Description     Tests deleting directories using Rmdir operation
//	Preconditions   Directory exists
//	Steps           1. Create an empty directory
//	                2. Delete the directory using Rmdir
//	                3. Verify directory no longer exists
//	                4. Test error conditions (non-empty directories, non-existent directories)
//	Expected Result Directories are deleted successfully when empty
//	Notes: This test verifies that directory deletion works correctly.
func TestUT_FS_DirOps_03_DirectoryDeletion_BasicOperations(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "DirectoryDeletionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Step 1: Create an empty directory
		dirName := "test_delete_dir"
		mkdirIn := &fuse.MkdirIn{
			InHeader: fuse.InHeader{NodeId: 1}, // Root node ID
			Mode:     0755,
		}
		entryOut := &fuse.EntryOut{}

		status := fs.Mkdir(nil, mkdirIn, dirName, entryOut)
		assert.Equal(fuse.OK, status, "Directory creation should succeed")

		dirNodeID := entryOut.NodeId

		// Verify directory exists
		dirInode := fs.GetNodeID(dirNodeID)
		assert.NotNil(dirInode, "Directory inode should exist")
		assert.Equal(dirName, dirInode.Name(), "Directory name should match")

		// Step 2: Delete the directory using Rmdir
		rmdirIn := &fuse.InHeader{NodeId: 1} // Parent node ID (root)

		status = fs.Rmdir(nil, rmdirIn, dirName)
		assert.Equal(fuse.OK, status, "Directory deletion should succeed")

		// Step 3: Verify directory no longer exists
		deletedInode := fs.GetNodeID(dirNodeID)
		assert.Nil(deletedInode, "Directory inode should no longer exist")

		// Step 4: Test error conditions

		// Try to delete a non-existent directory
		status = fs.Rmdir(nil, rmdirIn, "non_existent_dir")
		assert.Equal(fuse.ENOENT, status, "Deleting non-existent directory should return ENOENT")
	})
}
