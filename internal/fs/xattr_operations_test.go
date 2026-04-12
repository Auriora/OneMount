package fs

import (
	"syscall"
	"testing"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestIT_FS_32_01_XAttr_BasicOperations_WorkCorrectly tests basic extended attribute operations.
func TestIT_FS_32_01_XAttr_BasicOperations_WorkCorrectly(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "XAttrBasicOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a test file
		file := NewInode("xattr_test.txt", fuse.S_IFREG|0644, root)
		filesystem.InsertNodeID(file)
		filesystem.InsertID(file.ID(), file)
		filesystem.InsertChild(rootID, file)

		nodeID := file.NodeID()

		// Step 1: SetXAttr — set an extended attribute
		setIn := &fuse.SetXAttrIn{InHeader: fuse.InHeader{NodeId: nodeID}}
		status := filesystem.SetXAttr(nil, setIn, "user.test", []byte("test_value"))
		assert.Equal(fuse.OK, status, "SetXAttr should succeed")

		// Step 2: GetXAttr — retrieve the attribute (size query first)
		getHeader := &fuse.InHeader{NodeId: nodeID}
		size, status := filesystem.GetXAttr(nil, getHeader, "user.test", nil)
		assert.Equal(fuse.OK, status, "GetXAttr size query should succeed")
		assert.Equal(uint32(10), size, "GetXAttr should return correct size for 'test_value'")

		// Step 3: GetXAttr — retrieve the actual value
		buf := make([]byte, size)
		n, status := filesystem.GetXAttr(nil, getHeader, "user.test", buf)
		assert.Equal(fuse.OK, status, "GetXAttr should succeed")
		assert.Equal(uint32(10), n, "GetXAttr should return correct byte count")
		assert.Equal("test_value", string(buf[:n]), "GetXAttr should return correct value")

		// Step 4: GetXAttr — non-existent attribute
		_, status = filesystem.GetXAttr(nil, getHeader, "user.nonexistent", nil)
		assert.Equal(fuse.Status(syscall.ENODATA), status, "GetXAttr for non-existent attr should return ENODATA")

		// Step 5: ListXAttr — list attributes
		listHeader := &fuse.InHeader{NodeId: nodeID}
		listSize, status := filesystem.ListXAttr(nil, listHeader, nil)
		assert.Equal(fuse.OK, status, "ListXAttr size query should succeed")
		assert.True(listSize > 0, "ListXAttr should return non-zero size")

		listBuf := make([]byte, listSize)
		_, status = filesystem.ListXAttr(nil, listHeader, listBuf)
		assert.Equal(fuse.OK, status, "ListXAttr should succeed")

		// Step 6: RemoveXAttr — remove the attribute
		removeHeader := &fuse.InHeader{NodeId: nodeID}
		status = filesystem.RemoveXAttr(nil, removeHeader, "user.test")
		assert.Equal(fuse.OK, status, "RemoveXAttr should succeed")

		// Step 7: Verify removal
		_, status = filesystem.GetXAttr(nil, getHeader, "user.test", nil)
		assert.Equal(fuse.Status(syscall.ENODATA), status, "GetXAttr after removal should return ENODATA")

		// Step 8: RemoveXAttr on already-removed attr
		status = filesystem.RemoveXAttr(nil, removeHeader, "user.test")
		assert.Equal(fuse.Status(syscall.ENODATA), status, "RemoveXAttr on non-existent attr should return ENODATA")
	})
}

// TestIT_FS_33_01_FileStatus_XAttr_StatusCorrectlyReported tests file status extended attributes.
func TestIT_FS_33_01_FileStatus_XAttr_StatusCorrectlyReported(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "FileStatusXAttrFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create test files with different statuses
		file1 := NewInode("local_file.txt", fuse.S_IFREG|0644, root)
		filesystem.InsertNodeID(file1)
		filesystem.InsertID(file1.ID(), file1)
		filesystem.InsertChild(rootID, file1)
		filesystem.SetFileStatus(file1.ID(), FileStatusInfo{Status: StatusLocal})

		file2 := NewInode("cloud_file.txt", fuse.S_IFREG|0644, root)
		filesystem.InsertNodeID(file2)
		filesystem.InsertID(file2.ID(), file2)
		filesystem.InsertChild(rootID, file2)
		filesystem.SetFileStatus(file2.ID(), FileStatusInfo{Status: StatusCloud})

		// Verify statuses can be retrieved
		status1 := filesystem.GetFileStatus(file1.ID())
		assert.Equal(StatusLocal, status1.Status, "File 1 should have Local status")

		status2 := filesystem.GetFileStatus(file2.ID())
		assert.Equal(StatusCloud, status2.Status, "File 2 should have Cloud status")

		// Verify unknown file returns zero-value status (StatusCloud is iota/0)
		unknownStatus := filesystem.GetFileStatus("nonexistent-id")
		assert.Equal(StatusCloud, unknownStatus.Status, "Non-existent file should have zero-value status (StatusCloud)")
	})
}

// TestIT_FS_34_01_Filesystem_XAttrOperations_WorkCorrectly tests filesystem-level extended attribute operations.
func TestIT_FS_34_01_Filesystem_XAttrOperations_WorkCorrectly(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "FilesystemXAttrOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Create a test file
		file := NewInode("multi_xattr.txt", fuse.S_IFREG|0644, root)
		filesystem.InsertNodeID(file)
		filesystem.InsertID(file.ID(), file)
		filesystem.InsertChild(rootID, file)

		nodeID := file.NodeID()

		// Set multiple xattrs
		attrs := map[string]string{
			"user.author":  "test_user",
			"user.version": "1.0",
			"user.tag":     "important",
		}

		for name, value := range attrs {
			setIn := &fuse.SetXAttrIn{InHeader: fuse.InHeader{NodeId: nodeID}}
			status := filesystem.SetXAttr(nil, setIn, name, []byte(value))
			assert.Equal(fuse.OK, status, "SetXAttr should succeed for %s", name)
		}

		// Verify all xattrs can be retrieved
		getHeader := &fuse.InHeader{NodeId: nodeID}
		for name, expectedValue := range attrs {
			size, status := filesystem.GetXAttr(nil, getHeader, name, nil)
			assert.Equal(fuse.OK, status, "GetXAttr size query should succeed for %s", name)

			buf := make([]byte, size)
			_, status = filesystem.GetXAttr(nil, getHeader, name, buf)
			assert.Equal(fuse.OK, status, "GetXAttr should succeed for %s", name)
			assert.Equal(expectedValue, string(buf), "Value should match for %s", name)
		}

		// Overwrite an existing xattr
		setIn := &fuse.SetXAttrIn{InHeader: fuse.InHeader{NodeId: nodeID}}
		status := filesystem.SetXAttr(nil, setIn, "user.version", []byte("2.0"))
		assert.Equal(fuse.OK, status, "Overwriting xattr should succeed")

		buf := make([]byte, 3)
		_, status = filesystem.GetXAttr(nil, getHeader, "user.version", buf)
		assert.Equal(fuse.OK, status, "GetXAttr after overwrite should succeed")
		assert.Equal("2.0", string(buf), "Overwritten value should be updated")

		// Test GetXAttr with buffer too small
		smallBuf := make([]byte, 1)
		_, status = filesystem.GetXAttr(nil, getHeader, "user.author", smallBuf)
		assert.Equal(fuse.Status(syscall.ERANGE), status, "GetXAttr with small buffer should return ERANGE")

		// Test operations on non-existent node
		badHeader := &fuse.InHeader{NodeId: 99999}
		_, status = filesystem.GetXAttr(nil, badHeader, "user.test", nil)
		assert.Equal(fuse.ENOENT, status, "GetXAttr on non-existent node should return ENOENT")
	})
}
