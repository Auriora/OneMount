package fs

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Initialize the xattrs map
	inode.Lock()
	if inode.xattrs == nil {
		inode.xattrs = make(map[string][]byte)
	}
	inode.Unlock()

	// Define constants for test data
	const (
		attrName1 = "user.test.attr"
		attrName2 = "user.test.attr2"
	)
	attrValue1 := []byte("test value")
	attrValue2 := []byte("another value")

	// Test SetXAttr
	t.Run("SetXAttr_ShouldStoreAttributeValue", func(t *testing.T) {
		// Set an xattr
		inode.Lock()
		inode.xattrs[attrName1] = attrValue1
		inode.Unlock()

		// Verify the xattr was set
		inode.RLock()
		value, exists := inode.xattrs[attrName1]
		inode.RUnlock()

		require.True(t, exists, "Xattr was not set: %s", attrName1)
		assert.Equal(t, attrValue1, value, "Xattr value does not match. Got %v, expected %v", value, attrValue1)
	})

	// Test GetXAttr
	t.Run("GetXAttr_ShouldRetrieveAttributeValue", func(t *testing.T) {
		// Verify the xattr exists and has the correct value
		inode.RLock()
		value, exists := inode.xattrs[attrName1]
		inode.RUnlock()

		require.True(t, exists, "Xattr does not exist: %s", attrName1)
		assert.Equal(t, attrValue1, value, "Xattr value does not match. Got %v, expected %v", value, attrValue1)
	})

	// Test ListXAttr
	t.Run("ListXAttr_ShouldReturnAllAttributes", func(t *testing.T) {
		// Add another xattr
		inode.Lock()
		inode.xattrs[attrName2] = attrValue2
		inode.Unlock()

		// List xattrs
		inode.RLock()
		attrCount := len(inode.xattrs)
		attributes := make(map[string][]byte)
		for name, value := range inode.xattrs {
			attributes[name] = value
		}
		inode.RUnlock()

		// Verify both xattrs are present
		require.Equal(t, 2, attrCount, "Wrong number of xattrs. Got %d, expected 2", attrCount)

		value1, exists1 := attributes[attrName1]
		require.True(t, exists1, "First xattr not found: %s", attrName1)
		assert.Equal(t, attrValue1, value1, "First xattr value does not match. Got %v, expected %v", value1, attrValue1)

		value2, exists2 := attributes[attrName2]
		require.True(t, exists2, "Second xattr not found: %s", attrName2)
		assert.Equal(t, attrValue2, value2, "Second xattr value does not match. Got %v, expected %v", value2, attrValue2)
	})

	// Test RemoveXAttr
	t.Run("RemoveXAttr_ShouldDeleteAttribute", func(t *testing.T) {
		// Remove an xattr
		inode.Lock()
		delete(inode.xattrs, attrName1)
		inode.Unlock()

		// Verify it was removed
		inode.RLock()
		_, exists := inode.xattrs[attrName1]
		attrCount := len(inode.xattrs)
		remainingAttrs := make([]string, 0, attrCount)
		for name := range inode.xattrs {
			remainingAttrs = append(remainingAttrs, name)
		}
		inode.RUnlock()

		require.False(t, exists, "Xattr was not removed: %s", attrName1)
		require.Equal(t, 1, attrCount, "Wrong number of xattrs after removal. Got %d, expected 1", attrCount)
		assert.Equal(t, attrName2, remainingAttrs[0], "Remaining xattr is not the expected one. Got %s, expected %s",
			remainingAttrs[0], attrName2)
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
		statusValue, exists := inode.xattrs["user.onemount.status"]
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
		statusValue, exists := inode.xattrs["user.onemount.status"]
		assert.True(t, exists, "Status xattr was not set")
		assert.Equal(t, []byte("Error"), statusValue, "Status xattr value does not match")

		// Verify the error xattr was set
		errorValue, errorExists := inode.xattrs["user.onemount.error"]
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
		_, errorExists := inode.xattrs["user.onemount.error"]
		inode.RUnlock()

		assert.False(t, errorExists, "Error xattr was not removed")
	})
}

// TestFilesystemXattrOperations tests the FUSE xattr operations on the Filesystem struct
func TestFilesystemXattrOperations(t *testing.T) {
	// Create a unique test directory path
	testDir := filepath.Join("tmp", fmt.Sprintf("test_fs_xattr_%d", time.Now().UnixNano()))

	// Create a test filesystem
	testFS, err := NewFilesystem(auth, testDir, 30)
	require.NoError(t, err, "Failed to create test filesystem")

	// Setup cleanup to remove the test directory after test completes or fails
	t.Cleanup(func() {
		if err := os.RemoveAll(testDir); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test directory %s: %v", testDir, err)
		}
	})

	// Create a test inode
	inode := NewInode("test_fs_xattr", 0644|fuse.S_IFREG, nil)
	require.NotNil(t, inode, "Failed to create test inode")

	// Insert the inode into the filesystem
	nodeID := testFS.InsertChild("root", inode)
	require.NotZero(t, nodeID, "Failed to insert inode into filesystem")

	// Define constants for test data
	const (
		attrName1 = "user.test.fs.attr"
		attrName2 = "user.test.fs.attr2"
	)
	attrValue1 := []byte("test filesystem value")
	attrValue2 := []byte("another value")

	// Test SetXAttr
	t.Run("SetXAttr_ShouldStoreAttributeInFilesystem", func(t *testing.T) {
		// Create input parameters
		in := &fuse.SetXAttrIn{InHeader: fuse.InHeader{NodeId: nodeID}}

		// Call SetXAttr
		status := testFS.SetXAttr(nil, in, attrName1, attrValue1)
		require.Equal(t, fuse.OK, status, "SetXAttr failed with status: %v", status)

		// Verify the xattr was set
		inode.RLock()
		value, exists := inode.xattrs[attrName1]
		inode.RUnlock()

		require.True(t, exists, "Xattr was not set: %s", attrName1)
		assert.Equal(t, attrValue1, value, "Xattr value does not match. Got %v, expected %v", value, attrValue1)
	})

	// Test GetXAttr
	t.Run("GetXAttr_ShouldRetrieveAttributeFromFilesystem", func(t *testing.T) {
		// Create input parameters
		header := &fuse.InHeader{NodeId: nodeID}

		// First call with zero buffer to get size
		buf := make([]byte, 0)
		size, status := testFS.GetXAttr(nil, header, attrName1, buf)
		require.Equal(t, fuse.OK, status, "GetXAttr size query failed with status: %v", status)
		require.Equal(t, uint32(len(attrValue1)), size,
			"GetXAttr returned wrong size. Got %d, expected %d", size, len(attrValue1))

		// Call with properly sized buffer
		buf = make([]byte, size)
		size, status = testFS.GetXAttr(nil, header, attrName1, buf)
		require.Equal(t, fuse.OK, status, "GetXAttr failed with status: %v", status)
		require.Equal(t, uint32(len(attrValue1)), size,
			"GetXAttr returned wrong size. Got %d, expected %d", size, len(attrValue1))
		assert.Equal(t, attrValue1, buf,
			"GetXAttr returned wrong value. Got %v, expected %v", buf, attrValue1)
	})

	// Test ListXAttr
	t.Run("ListXAttr_ShouldReturnAllAttributesFromFilesystem", func(t *testing.T) {
		// Set another xattr
		in := &fuse.SetXAttrIn{InHeader: fuse.InHeader{NodeId: nodeID}}
		status := testFS.SetXAttr(nil, in, attrName2, attrValue2)
		require.Equal(t, fuse.OK, status, "Failed to set second xattr with status: %v", status)

		// First call with zero buffer to get size
		buf := make([]byte, 0)
		size, status := testFS.ListXAttr(nil, &in.InHeader, buf)
		require.Equal(t, fuse.OK, status, "ListXAttr size query failed with status: %v", status)
		require.Greater(t, size, uint32(0), "ListXAttr returned zero size")

		// Call with properly sized buffer
		buf = make([]byte, size)
		size, status = testFS.ListXAttr(nil, &in.InHeader, buf)
		require.Equal(t, fuse.OK, status, "ListXAttr failed with status: %v", status)

		// Parse the null-terminated list of attribute names
		var attrs []string
		start := 0
		for i := 0; i < len(buf); i++ {
			if buf[i] == 0 {
				attrs = append(attrs, string(buf[start:i]))
				start = i + 1
			}
		}

		// Verify both attributes are present
		require.GreaterOrEqual(t, len(attrs), 2,
			"ListXAttr returned too few attributes. Got %d, expected at least 2", len(attrs))
		assert.Contains(t, attrs, attrName1,
			"ListXAttr did not return first attribute: %s. Got attributes: %v", attrName1, attrs)
		assert.Contains(t, attrs, attrName2,
			"ListXAttr did not return second attribute: %s. Got attributes: %v", attrName2, attrs)
	})

	// Test RemoveXAttr
	t.Run("RemoveXAttr_ShouldDeleteAttributeFromFilesystem", func(t *testing.T) {
		// Create input parameters
		header := &fuse.InHeader{NodeId: nodeID}

		// Call RemoveXAttr
		status := testFS.RemoveXAttr(nil, header, attrName1)
		require.Equal(t, fuse.OK, status, "RemoveXAttr failed with status: %v", status)

		// Verify the xattr was removed
		inode.RLock()
		_, exists := inode.xattrs[attrName1]
		inode.RUnlock()

		require.False(t, exists, "Xattr was not removed: %s", attrName1)

		// Try to get the removed xattr
		buf := make([]byte, 100)
		_, status = testFS.GetXAttr(nil, header, attrName1, buf)
		assert.NotEqual(t, fuse.OK, status,
			"GetXAttr succeeded for removed attribute: %s. Expected failure but got status: %v",
			attrName1, status)
	})
}
