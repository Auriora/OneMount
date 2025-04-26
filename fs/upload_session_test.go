package fs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jstaf/onedriver/fs/graph"
	"github.com/jstaf/onedriver/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUploadSessionOperations tests various upload session operations
func TestUploadSessionOperations(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name        string
		description string
		setupFunc   func(t *testing.T) (string, []byte, func())
		testFunc    func(t *testing.T, filePath string, initialData []byte)
		skipCheck   func() bool
	}{
		{
			name:        "DirectUpload_ShouldSucceed",
			description: "Tests the basic functionality of uploads using internal functions directly",
			setupFunc: func(t *testing.T) (string, []byte, func()) {
				fileName := "uploadSessionSmall_" + t.Name() + ".txt"
				data := []byte("our super special data for " + t.Name())

				// Create a cleanup function
				cleanup := func() {
					// No cleanup needed for this test as it doesn't create files on disk
				}

				return fileName, data, cleanup
			},
			testFunc: func(t *testing.T, fileName string, data []byte) {
				testDir, err := fs.GetPath("/onedriver_tests", auth)
				require.NoError(t, err, "Failed to get test directory")

				inode := NewInode(fileName, 0644, testDir)
				inode.setContent(fs, data)
				mtime := inode.ModTime()

				// Create and upload the session
				session, err := NewUploadSession(inode, &data)
				require.NoError(t, err, "Failed to create upload session")
				err = session.Upload(auth)
				require.NoError(t, err, "Failed to upload session")

				// Verify the upload was successful
				require.False(t, isLocalID(session.ID),
					"The session's ID was somehow still local following an upload: %s",
					session.ID)
				sessionMtime := uint64(session.ModTime.Unix())
				assert.Equal(t, mtime, sessionMtime, 
					"Session modtime changed - before: %d - after: %d", mtime, sessionMtime)

				// Verify the content was uploaded correctly
				resp, _, err := graph.GetItemContent(session.ID, auth)
				require.NoError(t, err, "Failed to get item content")
				require.True(t, bytes.Equal(data, resp),
					"Data mismatch. Original content: %s\nRemote content: %s", data, resp)

				// Update the inode ID to the new remote ID
				inode.DriveItem.ID = session.ID

				// Test overwriting with new data
				newData := []byte("new data is extra long so it covers the old one completely - " + t.Name())
				inode.setContent(fs, newData)

				// Create and upload a new session
				session2, err := NewUploadSession(inode, &newData)
				require.NoError(t, err, "Failed to create second upload session")
				err = session2.Upload(auth)
				require.NoError(t, err, "Failed to upload second session")

				// Verify the content was updated correctly
				resp, _, err = graph.GetItemContent(session.ID, auth)
				require.NoError(t, err, "Failed to get updated item content")
				require.True(t, bytes.Equal(newData, resp),
					"Data mismatch after update. Original content: %s\nRemote content: %s", newData, resp)
			},
			skipCheck: func() bool {
				return false
			},
		},
		{
			name:        "SmallFileUpload_ShouldSucceed",
			description: "Tests small file uploads using the filesystem interface",
			setupFunc: func(t *testing.T) (string, []byte, func()) {
				fileName := "uploadSessionSmallFS_" + t.Name() + ".txt"
				filePath := filepath.Join(TestDir, fileName)
				data := []byte("super special data for upload test - " + t.Name())

				// Create a cleanup function
				cleanup := func() {
					if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
					}
				}

				return filePath, data, cleanup
			},
			testFunc: func(t *testing.T, filePath string, data []byte) {
				// Write the initial data to the file
				err := os.WriteFile(filePath, data, 0644)
				require.NoError(t, err, "Failed to write initial data to file")

				// Get the file name from the path
				fileName := filepath.Base(filePath)
				remotePath := "/onedriver_tests/" + fileName

				// Wait for the file to be uploaded and available on the server
				testutil.WaitForCondition(t, func() bool {
					item, err := graph.GetItemPath(remotePath, auth)
					return err == nil && item != nil
				}, 30*time.Second, time.Second, "File was not uploaded to server within timeout")

				// Now get the item for content verification
				item, err := graph.GetItemPath(remotePath, auth)
				require.NoError(t, err, "Failed to get item from server")
				require.NotNil(t, item, "Item not found on server")

				// Verify the content was uploaded correctly
				content, _, err := graph.GetItemContent(item.ID, auth)
				require.NoError(t, err, "Failed to get item content")
				require.True(t, bytes.Equal(content, data),
					"Data mismatch. Original content: %s\nRemote content: %s", data, content)

				// Test uploading again with new data
				newData := []byte("more super special data for - " + t.Name())
				err = os.WriteFile(filePath, newData, 0644)
				require.NoError(t, err, "Failed to write updated data to file")

				// Wait for the file to be uploaded again and available on the server with updated content
				testutil.WaitForCondition(t, func() bool {
					updatedItem, err := graph.GetItemPath(remotePath, auth)
					if err != nil || updatedItem == nil {
						return false
					}

					// Check if the content has been updated
					content, _, err := graph.GetItemContent(updatedItem.ID, auth)
					return err == nil && bytes.Equal(content, newData)
				}, 30*time.Second, time.Second, "File was not re-uploaded to server with updated content within timeout")

				// Now get the item for final verification
				item2, err := graph.GetItemPath(remotePath, auth)
				require.NoError(t, err, "Failed to get updated item from server")
				require.NotNil(t, item2, "Updated item not found on server")

				// Verify the content was updated correctly
				content, _, err = graph.GetItemContent(item2.ID, auth)
				require.NoError(t, err, "Failed to get updated item content")
				require.True(t, bytes.Equal(content, newData),
					"Data mismatch after update. Original content: %s\nRemote content: %s", newData, content)
			},
			skipCheck: func() bool {
				return false
			},
		},
		{
			name:        "LargeFileUpload_ShouldSucceed",
			description: "Tests large file uploads using the filesystem interface",
			setupFunc: func(t *testing.T) (string, []byte, func()) {
				// Check if the source file exists
				_, err := os.Stat("dmel.fa")
				if err != nil {
					return "", nil, func() {}
				}

				// Read the source file
				sourceData, err := os.ReadFile("dmel.fa")
				require.NoError(t, err, "Failed to read source file")

				fileName := "dmel_" + t.Name() + ".fa"
				filePath := filepath.Join(TestDir, fileName)

				// Create a cleanup function
				cleanup := func() {
					if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
					}
				}

				return filePath, sourceData, cleanup
			},
			testFunc: func(t *testing.T, filePath string, sourceData []byte) {
				// Write the data to the file
				err := os.WriteFile(filePath, sourceData, 0644)
				require.NoError(t, err, "Failed to write to destination file")

				// Read the file to verify it was written correctly
				contents, err := os.ReadFile(filePath)
				require.NoError(t, err, "Failed to read file after writing")

				// Verify the file header
				header := ">X dna:chromosome chromosome:BDGP6.22:X:1:23542271:1 REF"
				require.Equal(t, header, string(contents[:len(header)]),
					"Could not read FASTA header. Wanted \"%s\", got \"%s\"",
					header, string(contents[:len(header)]))

				// Verify the file footer
				final := "AAATAAAATAC\n" // makes yucky test output, but is the final line
				match := string(contents[len(contents)-len(final):])
				require.Equal(t, final, match,
					"Could not read final line of FASTA. Wanted \"%s\", got \"%s\"",
					final, match)

				// Verify the file size
				st, _ := os.Stat(filePath)
				require.NotZero(t, st.Size(), "File size cannot be 0.")

				// Get the file name from the path
				fileName := filepath.Base(filePath)
				remotePath := "/onedriver_tests/" + fileName

				// Poll endpoint to make sure it has a size greater than 0
				size := uint64(len(contents))
				var item *graph.DriveItem
				assert.Eventually(t, func() bool {
					item, _ = graph.GetItemPath(remotePath, auth)
					if item == nil {
						return false
					}
					inode := NewInodeDriveItem(item)
					return inode.Size() == size
				}, 120*time.Second, time.Second, "Upload session did not complete successfully!")

				// Test multipart downloads
				downloaded, _, err := graph.GetItemContent(item.ID, auth)
				assert.NoError(t, err, "Failed to download content")
				assert.Equal(t, graph.QuickXORHash(&contents), graph.QuickXORHash(&downloaded),
					"Downloaded content did not match original content.")
			},
			skipCheck: func() bool {
				_, err := os.Stat("dmel.fa")
				return err != nil
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Check if the test should be skipped
			if tc.skipCheck() {
				t.Skip("Skipping test: " + tc.description)
			}

			// Setup test resources
			filePath, initialData, cleanup := tc.setupFunc(t)

			// Register cleanup
			t.Cleanup(cleanup)

			// Run the test
			tc.testFunc(t, filePath, initialData)
		})
	}
}
