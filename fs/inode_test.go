package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// verify that items automatically get created with an ID of "local-"
func TestConstructor(t *testing.T) {
	inode := NewInode("Test Create", 0644|fuse.S_IFREG, nil)
	require.True(t, inode.ID() != "" && isLocalID(inode.ID()),
		"Expected an ID beginning with \"local-\", got \"%s\" instead",
		inode.ID())
}

// verify that the mode of items fetched are correctly set when fetched from
// server
func TestMode(t *testing.T) {
	// Ensure the Documents directory exists
	docDir := "mount/Documents"
	if _, err := os.Stat(docDir); os.IsNotExist(err) {
		require.NoError(t, os.Mkdir(docDir, 0755), "Failed to create Documents directory")
		// Give the filesystem time to process the directory creation
		time.Sleep(2 * time.Second)
	}

	// Test directory mode
	var item *graph.DriveItem
	var err error

	// Retry getting the Documents directory
	assert.Eventually(t, func() bool {
		item, err = graph.GetItemPath("/Documents", auth)
		return err == nil && item != nil
	}, 10*time.Second, time.Second, "Could not get Documents directory")

	require.NotNil(t, item, "Documents directory item cannot be nil, err: %v", err)
	inode := NewInodeDriveItem(item)
	require.Equal(t, uint32(0755|fuse.S_IFDIR), inode.Mode(),
		"mode of /Documents wrong: %o != %o",
		inode.Mode(), 0755|fuse.S_IFDIR)

	// Test file mode
	fname := "/onedriver_tests/test_mode.txt"
	fullPath := "mount" + fname

	// Remove the file if it exists to ensure a clean state
	os.Remove(fullPath)

	// Create the test file
	require.NoError(t, os.WriteFile(fullPath, []byte("test"), 0644))

	// Give the filesystem time to process the file creation
	time.Sleep(2 * time.Second)

	// Retry getting the test file
	assert.Eventually(t, func() bool {
		item, err = graph.GetItemPath(fname, auth)
		return err == nil && item != nil
	}, 15*time.Second, time.Second, "Could not get test file")

	require.NotNil(t, item, "Test file item cannot be nil, err: %v", err)
	inode = NewInodeDriveItem(item)
	require.Equal(t, uint32(0644|fuse.S_IFREG), inode.Mode(),
		"mode of file wrong: %o != %o",
		inode.Mode(), 0644|fuse.S_IFREG)
}

// Do we properly detect whether something is a directory or not?
func TestIsDir(t *testing.T) {
	// Ensure the Documents directory exists
	docDir := "mount/Documents"
	if _, err := os.Stat(docDir); os.IsNotExist(err) {
		require.NoError(t, os.Mkdir(docDir, 0755), "Failed to create Documents directory")
		// Give the filesystem time to process the directory creation
		time.Sleep(2 * time.Second)
	}

	// Test directory detection
	var item *graph.DriveItem
	var err error

	// Retry getting the Documents directory
	assert.Eventually(t, func() bool {
		item, err = graph.GetItemPath("/Documents", auth)
		return err == nil && item != nil
	}, 10*time.Second, time.Second, "Could not get Documents directory")

	require.NotNil(t, item, "Documents directory item cannot be nil, err: %v", err)
	inode := NewInodeDriveItem(item)
	require.True(t, inode.IsDir(), "/Documents not detected as a directory")

	// Test file detection
	fname := "/onedriver_tests/test_is_dir.txt"
	fullPath := "mount" + fname

	// Remove the file if it exists to ensure a clean state
	os.Remove(fullPath)

	// Create the test file
	require.NoError(t, os.WriteFile(fullPath, []byte("test"), 0644))

	// Give the filesystem time to process the file creation
	time.Sleep(2 * time.Second)

	// Retry getting the test file
	assert.Eventually(t, func() bool {
		item, err = graph.GetItemPath(fname, auth)
		if err == nil && item != nil {
			inode = NewInodeDriveItem(item)
			require.False(t, inode.IsDir(), "File created with mode 644 not detected as file")
			return true
		}
		return false
	}, 15*time.Second, time.Second, "Could not create item.")
}

// A filename like .~lock.libreoffice-test.docx# will fail to upload unless the
// filename is escaped.
func TestFilenameEscape(t *testing.T) {
	// Ensure the test directory exists
	err := os.MkdirAll(TestDir, 0755)
	require.NoError(t, err, "Failed to create test directory")

	// Use a special filename that needs escaping
	fname := `.~lock.libreoffice-test.docx#`
	filePath := filepath.Join(TestDir, fname)

	// Remove the file if it exists to ensure a clean state
	os.Remove(filePath)

	// Create the test file
	require.NoError(t, os.WriteFile(filePath, []byte("argl bargl"), 0644))

	// Give the filesystem time to process the file creation
	time.Sleep(2 * time.Second)

	// Make sure it made it to the server
	// Increase timeout and add more detailed logging
	assert.Eventually(t, func() bool {
		children, err := graph.GetItemChildrenPath("/onedriver_tests", auth)
		if err != nil {
			t.Logf("Error getting children: %v", err)
			return false
		}

		// Log all children to help debug
		t.Logf("Found %d children in /onedriver_tests", len(children))
		for i, child := range children {
			t.Logf("Child %d: %s", i, child.Name)
			if child.Name == fname {
				t.Logf("Found matching file: %s", child.Name)
				return true
			}
		}

		// If we didn't find the file, check if it exists locally
		if _, err := os.Stat(filePath); err != nil {
			t.Logf("File doesn't exist locally either: %v", err)
		} else {
			t.Logf("File exists locally but not on server yet")
		}

		return false
	}, 30*time.Second, 5*time.Second, "Could not find file: %s", fname)
}

// When running creat() on an existing file, we should truncate the existing file and
// return the original inode.
// Related to: https://github.com/jstaf/onedriver/issues/99
func TestDoubleCreate(t *testing.T) {
	fname := "double_create.txt"

	parent, err := fs.GetPath("/onedriver_tests", auth)
	require.NoError(t, err)

	fs.Create(
		context.Background().Done(),
		&fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: parent.NodeID()},
			Mode:     0644,
		},
		fname,
		&fuse.CreateOut{},
	)
	child, err := fs.GetChild(parent.ID(), fname, auth)

	// we clean up after ourselves to prevent failing some of the offline tests
	defer fs.Unlink(context.Background().Done(), &fuse.InHeader{NodeId: parent.nodeID}, fname)

	require.NoError(t, err)
	require.NotNil(t, child, "Could not find child post-create")
	childID := child.ID()

	fs.Create(
		context.Background().Done(),
		&fuse.CreateIn{
			InHeader: fuse.InHeader{NodeId: parent.NodeID()},
			Mode:     0644,
		},
		fname,
		&fuse.CreateOut{},
	)
	child, err = fs.GetChild(parent.ID(), fname, auth)
	require.NoError(t, err)
	require.NotNil(t, child, "Could not find child post-create")
	assert.Equal(t, childID, child.ID(),
		"IDs did not match when create run twice on same file.",
	)
}
