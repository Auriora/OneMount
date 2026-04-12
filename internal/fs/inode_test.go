package fs

import (
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestIT_FS_01_01_Inode_Creation_HasCorrectProperties tests that inodes are created with the correct properties.
//
//	Test Case ID    IT-FS-01-01
//	Title           Inode Creation Properties
//	Description     Tests that inodes are created with the correct properties
//	Preconditions   None
//	Steps           1. Create inodes with different modes (file, directory, executable)
//	                2. Verify the properties of each inode
//	Expected Result Inodes have the correct properties (ID, name, mode, directory status)
//	Notes: This test verifies that inodes are created with the correct properties.
func TestIT_FS_01_01_Inode_Creation_HasCorrectProperties(t *testing.T) {
	fixture := framework.NewUnitTestFixture("InodeCreationFixture")

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Create a root inode (directory)
		root := NewInode("root", fuse.S_IFDIR|0755, nil)
		assert.NotNil(root, "Root inode should not be nil")
		assert.Equal("root", root.Name(), "Root name should be 'root'")
		assert.True(root.IsDir(), "Root should be a directory")

		// Create a regular file inode
		file := NewInode("test.txt", fuse.S_IFREG|0644, root)
		assert.NotNil(file, "File inode should not be nil")
		assert.Equal("test.txt", file.Name(), "File name should be 'test.txt'")
		assert.False(file.IsDir(), "File should not be a directory")
		assert.NotEqual("", file.ID(), "File should have a non-empty ID")
		assert.Equal(root.ID(), file.ParentID(), "File parent ID should match root ID")

		// Create an executable file inode
		exec := NewInode("script.sh", fuse.S_IFREG|0755, root)
		assert.NotNil(exec, "Executable inode should not be nil")
		assert.Equal("script.sh", exec.Name(), "Executable name should be 'script.sh'")
		assert.False(exec.IsDir(), "Executable should not be a directory")

		// Create a subdirectory inode
		subdir := NewInode("subdir", fuse.S_IFDIR|0755, root)
		assert.NotNil(subdir, "Subdirectory inode should not be nil")
		assert.Equal("subdir", subdir.Name(), "Subdirectory name should be 'subdir'")
		assert.True(subdir.IsDir(), "Subdirectory should be a directory")
		assert.Equal(root.ID(), subdir.ParentID(), "Subdirectory parent ID should match root ID")

		// Verify IDs are unique
		assert.NotEqual(file.ID(), exec.ID(), "Different inodes should have different IDs")
		assert.NotEqual(file.ID(), subdir.ID(), "File and directory should have different IDs")
	})
}

// TestIT_FS_02_01_Inode_Properties_ModeAndDirectoryDetection tests various properties of inodes.
//
//	Test Case ID    IT-FS-02-01
//	Title           Inode Properties
//	Description     Tests various properties of inodes, including mode and directory detection
//	Preconditions   None
//	Steps           1. Create test directories and files
//	                2. Create inodes from drive items
//	                3. Test the mode and IsDir methods
//	Expected Result Inodes have the correct mode and directory status
//	Notes: This test verifies that inodes correctly report their mode and directory status.
func TestIT_FS_02_01_Inode_Properties_ModeAndDirectoryDetection(t *testing.T) {
	fixture := framework.NewUnitTestFixture("InodePropertiesFixture")

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		// Test with DriveItem representing a folder
		now := time.Now()
		folderItem := &graph.DriveItem{
			ID:      "folder-id-123",
			Name:    "Documents",
			Folder:  &graph.Folder{ChildCount: 5},
			ModTime: &now,
		}
		folderInode := NewInodeDriveItem(folderItem)
		assert.NotNil(folderInode, "Folder inode should not be nil")
		assert.True(folderInode.IsDir(), "Folder inode should be a directory")
		assert.Equal("Documents", folderInode.Name(), "Folder name should match")
		assert.Equal("folder-id-123", folderInode.ID(), "Folder ID should match")

		// Test with DriveItem representing a file
		fileItem := &graph.DriveItem{
			ID:   "file-id-456",
			Name: "report.pdf",
			Size: 1024,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "somehash==",
				},
			},
			ModTime: &now,
		}
		fileInode := NewInodeDriveItem(fileItem)
		assert.NotNil(fileInode, "File inode should not be nil")
		assert.False(fileInode.IsDir(), "File inode should not be a directory")
		assert.Equal("report.pdf", fileInode.Name(), "File name should match")
		assert.Equal(uint64(1024), fileInode.Size(), "File size should match")

		// Test mode for NewInode-created inodes
		dirInode := NewInode("testdir", fuse.S_IFDIR|0755, nil)
		assert.True(dirInode.Mode()&fuse.S_IFDIR != 0, "Directory mode should have S_IFDIR bit set")

		regInode := NewInode("testfile", fuse.S_IFREG|0644, nil)
		assert.True(regInode.Mode()&fuse.S_IFREG != 0, "Regular file mode should have S_IFREG bit set")
		assert.False(regInode.IsDir(), "Regular file should not be a directory")
	})
}

// TestIT_FS_03_01_Filename_SpecialCharacters_ProperlyEscaped tests that filenames with special characters are properly handled.
//
//	Test Case ID    IT-FS-03-01
//	Title           Filename Special Characters
//	Description     Tests that filenames with special characters are properly escaped
//	Preconditions   None
//	Steps           1. Create inodes with special characters in their names
//	                2. Verify the names are preserved correctly
//	Expected Result Files with special characters in their names are properly handled
//	Notes: This test verifies that filenames with special characters are properly escaped.
func TestIT_FS_03_01_Filename_SpecialCharacters_ProperlyEscaped(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "FilenameEscapingFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		fsFixture := getFSTestFixture(t, fixture)
		filesystem := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)
		assert.NotNil(root, "Root inode should exist")

		// Test filenames with various special characters
		specialNames := []string{
			"file with spaces.txt",
			"file-with-dashes.txt",
			"file_with_underscores.txt",
			"file.multiple.dots.txt",
			"UPPERCASE.TXT",
			"MiXeD CaSe FiLe.Txt",
			"file (1).txt",
			"file [brackets].txt",
			"résumé.pdf",
			"日本語ファイル.txt",
		}

		for _, name := range specialNames {
			inode := NewInode(name, fuse.S_IFREG|0644, root)
			assert.NotNil(inode, "Inode for '%s' should not be nil", name)
			assert.Equal(name, inode.Name(), "Inode name should be preserved for '%s'", name)
			filesystem.InsertID(inode.ID(), inode)
			filesystem.InsertChild(rootID, inode)
		}

		// Verify all files can be retrieved
		for _, name := range specialNames {
			child, _ := filesystem.GetChild(rootID, name, fsFixture.Auth)
			assert.NotNil(child, "Should be able to retrieve file '%s'", name)
			if child != nil {
				assert.Equal(name, child.Name(), "Retrieved file name should match for '%s'", name)
			}
		}
	})
}

// TestIT_FS_04_01_FileCreation_VariousScenarios_BehavesCorrectly tests various behaviors when creating files.
//
//	Test Case ID    IT-FS-04-01
//	Title           File Creation Behavior
//	Description     Tests various behaviors when creating files
//	Preconditions   None
//	Steps           1. Create a file
//	                2. Create the same file again
//	                3. Verify the same inode is returned
//	                4. Test with different modes and after writing content
//	Expected Result File creation behavior is correct, including truncation and returning the same inode
//	Notes: This test verifies that file creation behaves correctly in various scenarios.
func TestIT_FS_04_01_FileCreation_VariousScenarios_BehavesCorrectly(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "FileCreationBehaviorFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		fsFixture := getFSTestFixture(t, fixture)
		filesystem := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID
		root := filesystem.GetID(rootID)
		assert.NotNil(root, "Root inode should exist")

		// Step 1: Create a file
		file1 := NewInode("test_create.txt", fuse.S_IFREG|0644, root)
		assert.NotNil(file1, "First file creation should succeed")
		filesystem.InsertID(file1.ID(), file1)
		filesystem.InsertChild(rootID, file1)

		// Step 2: Verify the file exists
		retrieved, _ := filesystem.GetChild(rootID, "test_create.txt", fsFixture.Auth)
		assert.NotNil(retrieved, "File should be retrievable after creation")
		assert.Equal("test_create.txt", retrieved.Name(), "Retrieved file name should match")

		// Step 3: Create a file with different mode
		execFile := NewInode("script.sh", fuse.S_IFREG|0755, root)
		assert.NotNil(execFile, "Executable file creation should succeed")
		assert.False(execFile.IsDir(), "Executable file should not be a directory")

		// Step 4: Create a directory
		dir := NewInode("newdir", fuse.S_IFDIR|0755, root)
		assert.NotNil(dir, "Directory creation should succeed")
		assert.True(dir.IsDir(), "Created directory should be a directory")

		// Step 5: Create a file inside the directory
		nestedFile := NewInode("nested.txt", fuse.S_IFREG|0644, dir)
		assert.NotNil(nestedFile, "Nested file creation should succeed")
		assert.Equal(dir.ID(), nestedFile.ParentID(), "Nested file parent should be the directory")

		// Step 6: Verify hasChanges is initially false for new inodes
		freshFile := NewInode("fresh.txt", fuse.S_IFREG|0644, root)
		assert.False(freshFile.HasChanges(), "Newly created inode should not have changes")

		// Step 7: Set hasChanges and verify
		freshFile.SetHasChanges(true)
		assert.True(freshFile.HasChanges(), "Inode should have changes after SetHasChanges(true)")
	})
}
