package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/jstaf/onedriver/fs/graph"
	"github.com/jstaf/onedriver/testutil"
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

// TestInodeProperties tests various properties of inodes, including mode and directory detection
func TestInodeProperties(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name           string
		path           string
		isDirectory    bool
		expectedMode   uint32
		setupFunc      func(t *testing.T, path string) string
		cleanupFunc    func(t *testing.T, path string)
		expectedIsDir  bool
	}{
		{
			name:          "Directory_ShouldHaveCorrectModeAndIsDir",
			path:          "/Documents",
			isDirectory:   true,
			expectedMode:  0755 | fuse.S_IFDIR,
			setupFunc: func(t *testing.T, path string) string {
				// Ensure the Documents directory exists
				docDir := "mount" + path
				if _, err := os.Stat(docDir); os.IsNotExist(err) {
					require.NoError(t, os.Mkdir(docDir, 0755), "Failed to create Documents directory")

					// Wait for the filesystem to process the directory creation
					testutil.WaitForCondition(t, func() bool {
						// Check if the directory exists and is accessible
						_, err := os.Stat(docDir)
						return err == nil
					}, 5*time.Second, 100*time.Millisecond, "Documents directory was not created within timeout")
				}
				return docDir
			},
			cleanupFunc: func(t *testing.T, path string) {
				// No cleanup needed for Documents directory
			},
			expectedIsDir: true,
		},
		{
			name:          "File_ShouldHaveCorrectModeAndIsNotDir",
			path:          "/onedriver_tests/test_inode_properties.txt",
			isDirectory:   false,
			expectedMode:  0644 | fuse.S_IFREG,
			setupFunc: func(t *testing.T, path string) string {
				fullPath := "mount" + path

				// Remove the file if it exists to ensure a clean state
				if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to remove test file: %v", err)
				}

				// Create the test file
				require.NoError(t, os.WriteFile(fullPath, []byte("test"), 0644))

				// Wait for the filesystem to process the file creation
				testutil.WaitForCondition(t, func() bool {
					// Check if the file exists and is accessible
					_, err := os.Stat(fullPath)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "Test file was not created within timeout")

				return fullPath
			},
			cleanupFunc: func(t *testing.T, path string) {
				// Clean up the test file
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up test file: %v", err)
				}
			},
			expectedIsDir: false,
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup test resources
			fullPath := tc.setupFunc(t, tc.path)

			// Register cleanup
			t.Cleanup(func() {
				tc.cleanupFunc(t, fullPath)
			})

			// Get the item from the server
			var item *graph.DriveItem
			var err error

			// Retry getting the item
			assert.Eventually(t, func() bool {
				item, err = graph.GetItemPath(tc.path, auth)
				return err == nil && item != nil
			}, 15*time.Second, time.Second, "Could not get item at path %s", tc.path)

			require.NotNil(t, item, "Item at path %s cannot be nil, err: %v", tc.path, err)

			// Create an inode from the drive item
			inode := NewInodeDriveItem(item)

			// Test mode
			require.Equal(t, tc.expectedMode, inode.Mode(),
				"Mode of %s wrong: %o != %o",
				tc.path, inode.Mode(), tc.expectedMode)

			// Test IsDir
			if tc.expectedIsDir {
				require.True(t, inode.IsDir(), "Item at %s not detected as a directory", tc.path)
			} else {
				require.False(t, inode.IsDir(), "Item at %s incorrectly detected as a directory", tc.path)
			}
		})
	}
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
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to remove test file: %v", err)
	}

	// Create the test file
	require.NoError(t, os.WriteFile(filePath, []byte("argl bargl"), 0644))

	// Wait for the filesystem to process the file creation
	testutil.WaitForCondition(t, func() bool {
		_, err := os.Stat(filePath)
		return err == nil
	}, 5*time.Second, 100*time.Millisecond, "Test file was not created within timeout")

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
