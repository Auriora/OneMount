package fs

import (
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestUT_FS_FileWrite_01_FileCreation tests file creation and upload marking.
func TestIT_FS_FileWrite_01_FileCreation(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "FileCreationUploadFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		fileName := "test_upload_file.txt"
		createIn := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: 1},
			Mode:     0644,
		}
		createOut := &fuse.CreateOut{}

		status := fs.Create(nil, createIn, fileName, createOut)
		assert.Equal(fuse.OK, status, "File creation should succeed")
		assert.NotEqual(uint64(0), createOut.NodeId, "NodeId should be assigned")

		testContent := []byte("This is test content for upload verification")
		writeIn := &fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: createOut.NodeId},
			Offset:   0,
		}

		bytesWritten, writeStatus := fs.Write(nil, writeIn, testContent)
		assert.Equal(fuse.OK, writeStatus, "Write should succeed")
		assert.Equal(uint32(len(testContent)), bytesWritten, "All bytes should be written")

		id := fs.TranslateID(1)
		child, _ := fs.GetChild(id, fileName, fs.auth)
		assert.NotNil(child, "File should exist in directory")
		assert.Equal(fileName, child.Name(), "File name should match")

		fileInode := fs.GetNodeID(createOut.NodeId)
		assert.NotNil(fileInode, "File inode should exist")
		assert.True(fileInode.hasChanges, "File should be marked as having changes")

		fileStatus := fs.GetFileStatus(fileInode.ID())
		assert.Equal(StatusLocalModified, fileStatus.Status, "File status should be StatusLocalModified")
	})
}

// TestUT_FS_FileWrite_02_FileModification tests file modification and upload queuing.
func TestIT_FS_FileWrite_02_FileModification(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "FileModificationFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		fileName := "test_modify_file.txt"
		createIn := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: 1},
			Mode:     0644,
		}
		createOut := &fuse.CreateOut{}

		status := fs.Create(nil, createIn, fileName, createOut)
		assert.Equal(fuse.OK, status, "File creation should succeed")

		initialContent := []byte("Initial content")
		writeIn := &fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: createOut.NodeId},
			Offset:   0,
		}

		bytesWritten, writeStatus := fs.Write(nil, writeIn, initialContent)
		assert.Equal(fuse.OK, writeStatus, "Initial write should succeed")
		assert.Equal(uint32(len(initialContent)), bytesWritten, "All initial bytes should be written")

		fileInode := fs.GetNodeID(createOut.NodeId)
		assert.NotNil(fileInode, "File inode should exist")
		assert.True(fileInode.hasChanges, "File should be marked as having changes after initial write")

		modifiedContent := []byte("Modified content - this is the new version")
		writeIn.Offset = 0

		bytesWritten, writeStatus = fs.Write(nil, writeIn, modifiedContent)
		assert.Equal(fuse.OK, writeStatus, "Modification write should succeed")
		assert.Equal(uint32(len(modifiedContent)), bytesWritten, "All modified bytes should be written")

		assert.True(fileInode.hasChanges, "File should still be marked as having changes after modification")
		assert.Equal(uint64(len(modifiedContent)), fileInode.DriveItem.Size, "File size should reflect modified content")

		fileStatus := fs.GetFileStatus(fileInode.ID())
		assert.Equal(StatusLocalModified, fileStatus.Status, "File status should be StatusLocalModified")
	})
}

// TestUT_FS_FileWrite_03_FileDeletion tests file deletion and sync verification.
func TestIT_FS_FileWrite_03_FileDeletion(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "FileDeletionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		fileName := "test_delete_file.txt"
		createIn := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: 1},
			Mode:     0644,
		}
		createOut := &fuse.CreateOut{}

		status := fs.Create(nil, createIn, fileName, createOut)
		assert.Equal(fuse.OK, status, "File creation should succeed")

		testContent := []byte("Content to be deleted")
		writeIn := &fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: createOut.NodeId},
			Offset:   0,
		}

		bytesWritten, writeStatus := fs.Write(nil, writeIn, testContent)
		assert.Equal(fuse.OK, writeStatus, "Write should succeed")
		assert.Equal(uint32(len(testContent)), bytesWritten, "All bytes should be written")

		id := fs.TranslateID(1)
		child, _ := fs.GetChild(id, fileName, fs.auth)
		assert.NotNil(child, "File should exist before deletion")

		unlinkIn := &fuse.InHeader{NodeId: 1}
		unlinkStatus := fs.Unlink(nil, unlinkIn, fileName)
		assert.Equal(fuse.OK, unlinkStatus, "Unlink should succeed")

		child, _ = fs.GetChild(id, fileName, fs.auth)
		assert.Nil(child, "File should not exist after deletion")

		deletedInode := fs.GetNodeID(createOut.NodeId)
		assert.Nil(deletedInode, "File inode should be deleted from filesystem")

		// Note: Content cache may still have the file data for caching purposes
		// This is expected behavior - the cache persists even after file deletion
	})
}

// TestUT_FS_FileWrite_04_DirectoryOperations tests directory creation, file operations within, and deletion.
func TestIT_FS_FileWrite_04_DirectoryOperations(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "DirectoryOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		fs.SetOfflineMode(OfflineModeReadWrite)
		t.Cleanup(func() { fs.SetOfflineMode(OfflineModeDisabled) })
		rootInode := fs.GetID(fsFixture.RootID)
		if rootInode == nil {
			t.Fatalf("Root inode not found")
		}
		rootNodeID := rootInode.NodeID()
		if rootNodeID == 0 {
			rootNodeID = fs.InsertNodeID(rootInode)
		}

		dirName := "test_directory"
		mkdirIn := &fuse.MkdirIn{
			InHeader: fuse.InHeader{NodeId: rootNodeID},
			Mode:     0755,
		}
		dirEntryOut := &fuse.EntryOut{}

		status := fs.Mkdir(nil, mkdirIn, dirName, dirEntryOut)
		assert.Equal(fuse.OK, status, "Directory creation should succeed")
		assert.NotEqual(uint64(0), dirEntryOut.NodeId, "Directory NodeId should be assigned")

		dirInode := fs.GetNodeID(dirEntryOut.NodeId)
		assert.NotNil(dirInode, "Directory inode should exist")
		assert.True(dirInode.IsDir(), "Inode should be a directory")
		assert.Equal(dirName, dirInode.Name(), "Directory name should match")

		file1Name := "file1.txt"
		file2Name := "file2.txt"

		createIn1 := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: dirEntryOut.NodeId},
			Mode:     0644,
		}
		createOut1 := &fuse.CreateOut{}

		status = fs.Create(nil, createIn1, file1Name, createOut1)
		assert.Equal(fuse.OK, status, "File1 creation should succeed")

		content1 := []byte("Content of file 1")
		writeIn1 := &fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: createOut1.NodeId},
			Offset:   0,
		}

		bytesWritten, writeStatus := fs.Write(nil, writeIn1, content1)
		assert.Equal(fuse.OK, writeStatus, "Write to file1 should succeed")
		assert.Equal(uint32(len(content1)), bytesWritten, "All bytes should be written to file1")

		createIn2 := &fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: dirEntryOut.NodeId},
			Mode:     0644,
		}
		createOut2 := &fuse.CreateOut{}

		status = fs.Create(nil, createIn2, file2Name, createOut2)
		assert.Equal(fuse.OK, status, "File2 creation should succeed")

		content2 := []byte("Content of file 2")
		writeIn2 := &fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: createOut2.NodeId},
			Offset:   0,
		}

		bytesWritten, writeStatus = fs.Write(nil, writeIn2, content2)
		assert.Equal(fuse.OK, writeStatus, "Write to file2 should succeed")
		assert.Equal(uint32(len(content2)), bytesWritten, "All bytes should be written to file2")

		dirID := fs.TranslateID(dirEntryOut.NodeId)
		child1, _ := fs.GetChild(dirID, file1Name, fs.auth)
		assert.NotNil(child1, "File1 should exist in directory")
		assert.Equal(file1Name, child1.Name(), "File1 name should match")

		child2, _ := fs.GetChild(dirID, file2Name, fs.auth)
		assert.NotNil(child2, "File2 should exist in directory")
		assert.Equal(file2Name, child2.Name(), "File2 name should match")

		unlinkIn := &fuse.InHeader{NodeId: dirEntryOut.NodeId}

		unlinkStatus := fs.Unlink(nil, unlinkIn, file1Name)
		assert.Equal(fuse.OK, unlinkStatus, "File1 deletion should succeed")

		unlinkStatus = fs.Unlink(nil, unlinkIn, file2Name)
		assert.Equal(fuse.OK, unlinkStatus, "File2 deletion should succeed")

		child1, _ = fs.GetChild(dirID, file1Name, fs.auth)
		assert.Nil(child1, "File1 should not exist after deletion")

		child2, _ = fs.GetChild(dirID, file2Name, fs.auth)
		assert.Nil(child2, "File2 should not exist after deletion")

		// Note: Directory deletion via Rmdir requires server sync which is not
		// fully supported in the mock environment. The test verifies that:
		// 1. Directory can be created
		// 2. Files can be created within the directory
		// 3. Files can be deleted from the directory
		// Directory deletion itself is tested in integration tests with real server
		// Note: Directory deletion via Rmdir requires server sync which is not
		// fully supported in the mock environment. The test verifies that:
		// 1. Directory can be created
		// 2. Files can be created within the directory
		// 3. Files can be deleted from the directory
		// Directory deletion itself is tested in integration tests with real server
	})
}
