package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bcherrington/onedriver/internal/fs/graph"
	"github.com/bcherrington/onedriver/internal/testutil"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInodeCreation verifies that inodes are created with the correct properties
func TestInodeCreation(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name           string
		itemName       string
		mode           uint32
		parent         *Inode
		expectedIDType string
		verifyFunc     func(t *testing.T, inode *Inode)
	}{
		{
			name:           "RegularFile_ShouldHaveLocalID",
			itemName:       "Test Regular File",
			mode:           0644 | fuse.S_IFREG,
			parent:         nil,
			expectedIDType: "local",
			verifyFunc: func(t *testing.T, inode *Inode) {
				require.True(t, isLocalID(inode.ID()),
					"Expected an ID beginning with \"local-\", got \"%s\" instead", inode.ID())
				require.Equal(t, "Test Regular File", inode.Name(),
					"Inode name does not match expected value")
				require.Equal(t, uint32(0644|fuse.S_IFREG), inode.Mode(),
					"Inode mode does not match expected value")
				require.False(t, inode.IsDir(),
					"Regular file incorrectly detected as a directory")
			},
		},
		{
			name:           "Directory_ShouldHaveLocalID",
			itemName:       "Test Directory",
			mode:           0755 | fuse.S_IFDIR,
			parent:         nil,
			expectedIDType: "local",
			verifyFunc: func(t *testing.T, inode *Inode) {
				require.True(t, isLocalID(inode.ID()),
					"Expected an ID beginning with \"local-\", got \"%s\" instead", inode.ID())
				require.Equal(t, "Test Directory", inode.Name(),
					"Inode name does not match expected value")
				require.Equal(t, uint32(0755|fuse.S_IFDIR), inode.Mode(),
					"Inode mode does not match expected value")
				require.True(t, inode.IsDir(),
					"Directory not detected as a directory")
			},
		},
		{
			name:           "ExecutableFile_ShouldHaveLocalID",
			itemName:       "Test Executable",
			mode:           0755 | fuse.S_IFREG,
			parent:         nil,
			expectedIDType: "local",
			verifyFunc: func(t *testing.T, inode *Inode) {
				require.True(t, isLocalID(inode.ID()),
					"Expected an ID beginning with \"local-\", got \"%s\" instead", inode.ID())
				require.Equal(t, "Test Executable", inode.Name(),
					"Inode name does not match expected value")
				require.Equal(t, uint32(0755|fuse.S_IFREG), inode.Mode(),
					"Inode mode does not match expected value")
				require.False(t, inode.IsDir(),
					"Executable file incorrectly detected as a directory")
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Create the inode
			inode := NewInode(tc.itemName, tc.mode, tc.parent)

			// Verify the inode properties
			tc.verifyFunc(t, inode)
		})
	}
}

// TestInodeProperties tests various properties of inodes, including mode and directory detection
func TestInodeProperties(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name          string
		path          string
		isDirectory   bool
		expectedMode  uint32
		setupFunc     func(t *testing.T, path string) string
		cleanupFunc   func(t *testing.T, path string)
		expectedIsDir bool
	}{
		{
			name:         "Directory_ShouldHaveCorrectModeAndIsDir",
			path:         "/Onedriver-Documents",
			isDirectory:  true,
			expectedMode: 0755 | fuse.S_IFDIR,
			setupFunc: func(t *testing.T, path string) string {
				// Ensure the Documents directory exists
				docDir := testutil.TestMountPoint + path
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
			name:         "File_ShouldHaveCorrectModeAndIsNotDir",
			path:         "/onedriver_tests/test_inode_properties.txt",
			isDirectory:  false,
			expectedMode: 0644 | fuse.S_IFREG,
			setupFunc: func(t *testing.T, path string) string {
				fullPath := testutil.TestMountPoint + path

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

// TestFilenameEscaping verifies that filenames with special characters are properly escaped
// and can be successfully uploaded to the server.
func TestFilenameEscaping(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name        string
		filename    string
		content     string
		description string
	}{
		{
			name:        "LibreOfficeTemp_ShouldBeEscaped",
			filename:    `.~lock.libreoffice-test.docx#`,
			content:     "libreoffice temp file content",
			description: "LibreOffice temporary lock file",
		},
		{
			name:        "FilenameWithHash_ShouldBeEscaped",
			filename:    `test#file.txt`,
			content:     "file with hash content",
			description: "Filename containing hash character",
		},
		{
			name:        "FilenameWithQuestionMark_ShouldBeEscaped",
			filename:    `test?file.txt`,
			content:     "file with question mark content",
			description: "Filename containing question mark character",
		},
		{
			name:        "FilenameWithAsterisk_ShouldBeEscaped",
			filename:    `test*file.txt`,
			content:     "file with asterisk content",
			description: "Filename containing asterisk character",
		},
	}

	// Ensure the test directory exists
	err := os.MkdirAll(TestDir, 0755)
	require.NoError(t, err, "Failed to create test directory")

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Create a unique file path for this test
			filePath := filepath.Join(TestDir, tc.filename)

			// Clean up after the test
			t.Cleanup(func() {
				if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
				}
			})

			// Remove the file if it exists to ensure a clean state
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				t.Logf("Warning: Failed to remove existing test file: %v", err)
			}

			// Create the test file
			require.NoError(t, os.WriteFile(filePath, []byte(tc.content), 0644),
				"Failed to create test file with special characters: %s", tc.filename)

			// Wait for the filesystem to process the file creation
			message := "Test file " + tc.filename + " was not created within timeout"
			testutil.WaitForCondition(t, func() bool {
				_, err := os.Stat(filePath)
				return err == nil
			}, 5*time.Second, 100*time.Millisecond, message)

			// Make sure it made it to the server
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
					if child.Name == tc.filename {
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
			}, 30*time.Second, 5*time.Second, "Could not find file "+tc.filename+" ("+tc.description+") on server")

			// Verify the file content
			content, err := os.ReadFile(filePath)
			require.NoError(t, err, "Failed to read test file: %s", tc.filename)
			assert.Equal(t, tc.content, string(content),
				"File content does not match expected value for %s", tc.filename)
		})
	}
}

// TestFileCreationBehavior verifies various behaviors when creating files, including
// creating a file that already exists (which should truncate the existing file and
// return the original inode).
// Related to: https://github.com/bcherrington/onedriver/issues/99
func TestFileCreationBehavior(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name        string
		filename    string
		initialMode uint32
		secondMode  uint32
		content     string
		description string
	}{
		{
			name:        "CreateTwice_ShouldReturnSameInode",
			filename:    "double_create.txt",
			initialMode: 0644,
			secondMode:  0644,
			content:     "test content",
			description: "Creating the same file twice should return the same inode",
		},
		{
			name:        "CreateWithDifferentMode_ShouldReturnSameInode",
			filename:    "double_create_different_mode.txt",
			initialMode: 0644,
			secondMode:  0755,
			content:     "test content with different mode",
			description: "Creating the same file with different mode should return the same inode",
		},
		{
			name:        "CreateAfterWrite_ShouldTruncateAndReturnSameInode",
			filename:    "create_after_write.txt",
			initialMode: 0644,
			secondMode:  0644,
			content:     "this content should be truncated",
			description: "Creating a file after writing to it should truncate the file and return the same inode",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Get the parent directory
			parent, err := fs.GetPath("/onedriver_tests", auth)
			require.NoError(t, err, "Failed to get parent directory")

			// Create the file for the first time
			fs.Create(
				context.Background().Done(),
				&fuse.CreateIn{
					InHeader: fuse.InHeader{NodeId: parent.NodeID()},
					Mode:     tc.initialMode,
				},
				tc.filename,
				&fuse.CreateOut{},
			)

			// Get the child after first creation
			child, err := fs.GetChild(parent.ID(), tc.filename, auth)

			// Clean up after ourselves to prevent failing some of the offline tests
			t.Cleanup(func() {
				fs.Unlink(context.Background().Done(), &fuse.InHeader{NodeId: parent.nodeID}, tc.filename)
			})

			// Verify the child was created successfully
			require.NoError(t, err, "Failed to get child after first creation")
			require.NotNil(t, child, "Child not found after first creation")

			// Store the original ID
			childID := child.ID()
			t.Logf("Original file ID: %s", childID)

			// For the "CreateAfterWrite" test case, write content to the file
			if tc.name == "CreateAfterWrite_ShouldTruncateAndReturnSameInode" {
				// Write content to the file
				filePath := filepath.Join("tmp", "mount/onedriver_tests", tc.filename)
				err := os.WriteFile(filePath, []byte(tc.content), 0644)
				require.NoError(t, err, "Failed to write content to file")

				// Verify the content was written
				content, err := os.ReadFile(filePath)
				require.NoError(t, err, "Failed to read file content")
				assert.Equal(t, tc.content, string(content), "File content does not match expected value")
			}

			// Create the file for the second time
			fs.Create(
				context.Background().Done(),
				&fuse.CreateIn{
					InHeader: fuse.InHeader{NodeId: parent.NodeID()},
					Mode:     tc.secondMode,
				},
				tc.filename,
				&fuse.CreateOut{},
			)

			// Get the child after second creation
			child, err = fs.GetChild(parent.ID(), tc.filename, auth)
			require.NoError(t, err, "Failed to get child after second creation")
			require.NotNil(t, child, "Child not found after second creation")

			// Verify the ID is the same
			assert.Equal(t, childID, child.ID(),
				"IDs did not match when create run twice on same file: %s != %s",
				childID, child.ID())

			// For the "CreateAfterWrite" test case, verify the file was truncated
			if tc.name == "CreateAfterWrite_ShouldTruncateAndReturnSameInode" {
				filePath := filepath.Join("tmp", "mount/onedriver_tests", tc.filename)
				content, err := os.ReadFile(filePath)
				require.NoError(t, err, "Failed to read file content after second creation")
				assert.Empty(t, string(content), "File was not truncated after second creation")
			}
		})
	}
}
