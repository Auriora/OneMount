package fs

import (
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestXattrOperations tests the xattr operations directly on the Inode struct
func TestXattrOperations(t *testing.T) {
	// Create a test inode
	inode := NewInode("test_xattr", 0644|fuse.S_IFREG, nil)
	require.NotNil(t, inode, "Failed to create test inode")

	// Test SetXAttr
	t.Run("SetXAttr", func(t *testing.T) {
		inode.Lock()
		// Initialize the xattrs map if it's nil
		if inode.xattrs == nil {
			inode.xattrs = make(map[string][]byte)
		}

		// Set an xattr
		attrName := "user.test.attr"
		attrValue := []byte("test value")
		inode.xattrs[attrName] = attrValue
		inode.Unlock()

		// Verify the xattr was set
		inode.RLock()
		value, exists := inode.xattrs[attrName]
		inode.RUnlock()

		assert.True(t, exists, "Xattr was not set")
		assert.Equal(t, attrValue, value, "Xattr value does not match")
	})

	// Test GetXAttr
	t.Run("GetXAttr", func(t *testing.T) {
		inode.RLock()
		attrName := "user.test.attr"
		value, exists := inode.xattrs[attrName]
		inode.RUnlock()

		assert.True(t, exists, "Xattr does not exist")
		assert.Equal(t, []byte("test value"), value, "Xattr value does not match")
	})

	// Test ListXAttr
	t.Run("ListXAttr", func(t *testing.T) {
		// Add another xattr
		inode.Lock()
		inode.xattrs["user.test.attr2"] = []byte("another value")
		inode.Unlock()

		// List xattrs
		inode.RLock()
		attrCount := len(inode.xattrs)
		hasAttr1 := false
		hasAttr2 := false
		for name := range inode.xattrs {
			if name == "user.test.attr" {
				hasAttr1 = true
			}
			if name == "user.test.attr2" {
				hasAttr2 = true
			}
		}
		inode.RUnlock()

		assert.Equal(t, 2, attrCount, "Wrong number of xattrs")
		assert.True(t, hasAttr1, "First xattr not found")
		assert.True(t, hasAttr2, "Second xattr not found")
	})

	// Test RemoveXAttr
	t.Run("RemoveXAttr", func(t *testing.T) {
		// Remove an xattr
		inode.Lock()
		delete(inode.xattrs, "user.test.attr")
		inode.Unlock()

		// Verify it was removed
		inode.RLock()
		_, exists := inode.xattrs["user.test.attr"]
		attrCount := len(inode.xattrs)
		inode.RUnlock()

		assert.False(t, exists, "Xattr was not removed")
		assert.Equal(t, 1, attrCount, "Wrong number of xattrs after removal")
	})
}

// TestFileStatusXattr tests the file status functionality that uses xattrs
func TestFileStatusXattr(t *testing.T) {
	// Create a test inode
	inode := NewInode("test_status", 0644|fuse.S_IFREG, nil)
	require.NotNil(t, inode, "Failed to create test inode")

	// Create a test filesystem
	testFS, err := NewFilesystem(auth, "tmp/test_xattr", 30)
	require.NoError(t, err, "Failed to create test filesystem")

	// Insert the inode into the filesystem
	nodeID := testFS.InsertChild("root", inode)
	require.NotZero(t, nodeID, "Failed to insert inode into filesystem")

	// Test setting file status
	t.Run("SetFileStatus", func(t *testing.T) {
		// Set file status
		status := FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Unix(int64(inode.ModTime()), 0),
		}
		testFS.SetFileStatus(inode.ID(), status)

		// Update the file status xattr
		testFS.updateFileStatus(inode)

		// Verify the xattr was set
		inode.RLock()
		statusValue, exists := inode.xattrs["user.onedriver.status"]
		inode.RUnlock()

		assert.True(t, exists, "Status xattr was not set")
		assert.Equal(t, []byte("Local"), statusValue, "Status xattr value does not match")
	})

	// Test setting file status with error
	t.Run("SetFileStatusWithError", func(t *testing.T) {
		// Set file status with error
		status := FileStatusInfo{
			Status:    StatusError,
			ErrorMsg:  "Test error message",
			Timestamp: time.Unix(int64(inode.ModTime()), 0),
		}
		testFS.SetFileStatus(inode.ID(), status)

		// Update the file status xattr
		testFS.updateFileStatus(inode)

		// Verify the status xattr was set
		inode.RLock()
		statusValue, exists := inode.xattrs["user.onedriver.status"]
		assert.True(t, exists, "Status xattr was not set")
		assert.Equal(t, []byte("Error"), statusValue, "Status xattr value does not match")

		// Verify the error xattr was set
		errorValue, errorExists := inode.xattrs["user.onedriver.error"]
		inode.RUnlock()

		assert.True(t, errorExists, "Error xattr was not set")
		assert.Equal(t, []byte("Test error message"), errorValue, "Error xattr value does not match")
	})

	// Test clearing error
	t.Run("ClearError", func(t *testing.T) {
		// Set file status without error
		status := FileStatusInfo{
			Status:    StatusLocal,
			Timestamp: time.Unix(int64(inode.ModTime()), 0),
		}
		testFS.SetFileStatus(inode.ID(), status)

		// Update the file status xattr
		testFS.updateFileStatus(inode)

		// Verify the error xattr was removed
		inode.RLock()
		_, errorExists := inode.xattrs["user.onedriver.error"]
		inode.RUnlock()

		assert.False(t, errorExists, "Error xattr was not removed")
	})
}

// TestFilesystemXattrOperations tests the FUSE xattr operations on the Filesystem struct
func TestFilesystemXattrOperations(t *testing.T) {
	// Create a test filesystem
	testFS, err := NewFilesystem(auth, "tmp/test_fs_xattr", 30)
	require.NoError(t, err, "Failed to create test filesystem")

	// Create a test inode
	inode := NewInode("test_fs_xattr", 0644|fuse.S_IFREG, nil)
	require.NotNil(t, inode, "Failed to create test inode")

	// Insert the inode into the filesystem
	nodeID := testFS.InsertChild("root", inode)
	require.NotZero(t, nodeID, "Failed to insert inode into filesystem")

	// Test SetXAttr
	t.Run("SetXAttr", func(t *testing.T) {
		// Create input parameters
		in := &fuse.SetXAttrIn{InHeader: fuse.InHeader{NodeId: nodeID}}
		attrName := "user.test.fs.attr"
		attrValue := []byte("test filesystem value")

		// Call SetXAttr
		status := testFS.SetXAttr(nil, in, attrName, attrValue)
		assert.Equal(t, fuse.OK, status, "SetXAttr failed")

		// Verify the xattr was set
		inode.RLock()
		value, exists := inode.xattrs[attrName]
		inode.RUnlock()

		assert.True(t, exists, "Xattr was not set")
		assert.Equal(t, attrValue, value, "Xattr value does not match")
	})

	// Test GetXAttr
	t.Run("GetXAttr", func(t *testing.T) {
		// Create input parameters
		header := &fuse.InHeader{NodeId: nodeID}
		attrName := "user.test.fs.attr"

		// First call with zero buffer to get size
		buf := make([]byte, 0)
		size, status := testFS.GetXAttr(nil, header, attrName, buf)
		assert.Equal(t, fuse.OK, status, "GetXAttr size query failed")
		assert.Equal(t, uint32(len([]byte("test filesystem value"))), size, "GetXAttr returned wrong size")

		// Call with properly sized buffer
		buf = make([]byte, size)
		size, status = testFS.GetXAttr(nil, header, attrName, buf)
		assert.Equal(t, fuse.OK, status, "GetXAttr failed")
		assert.Equal(t, uint32(len([]byte("test filesystem value"))), size, "GetXAttr returned wrong size")
		assert.Equal(t, []byte("test filesystem value"), buf, "GetXAttr returned wrong value")
	})

	// Test ListXAttr
	t.Run("ListXAttr", func(t *testing.T) {
		// Set another xattr
		in := &fuse.SetXAttrIn{InHeader: fuse.InHeader{NodeId: nodeID}}
		testFS.SetXAttr(nil, in, "user.test.fs.attr2", []byte("another value"))

		// First call with zero buffer to get size
		buf := make([]byte, 0)
		size, status := testFS.ListXAttr(nil, &in.InHeader, buf)
		assert.Equal(t, fuse.OK, status, "ListXAttr size query failed")
		assert.Greater(t, size, uint32(0), "ListXAttr returned zero size")

		// Call with properly sized buffer
		buf = make([]byte, size)
		size, status = testFS.ListXAttr(nil, &in.InHeader, buf)
		assert.Equal(t, fuse.OK, status, "ListXAttr failed")

		// Parse the null-terminated list of attribute names
		var attrs []string
		start := 0
		for i := 0; i < len(buf); i++ {
			if buf[i] == 0 {
				attrs = append(attrs, string(buf[start:i]))
				start = i + 1
			}
		}

		assert.Contains(t, attrs, "user.test.fs.attr", "ListXAttr did not return first attribute")
		assert.Contains(t, attrs, "user.test.fs.attr2", "ListXAttr did not return second attribute")
	})

	// Test RemoveXAttr
	t.Run("RemoveXAttr", func(t *testing.T) {
		// Create input parameters
		header := &fuse.InHeader{NodeId: nodeID}
		attrName := "user.test.fs.attr"

		// Call RemoveXAttr
		status := testFS.RemoveXAttr(nil, header, attrName)
		assert.Equal(t, fuse.OK, status, "RemoveXAttr failed")

		// Verify the xattr was removed
		inode.RLock()
		_, exists := inode.xattrs[attrName]
		inode.RUnlock()

		assert.False(t, exists, "Xattr was not removed")

		// Try to get the removed xattr
		buf := make([]byte, 100)
		_, status = testFS.GetXAttr(nil, header, attrName, buf)
		assert.NotEqual(t, fuse.OK, status, "GetXAttr succeeded for removed attribute")
	})
}
