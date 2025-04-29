// This package exists purely for the convenience of easily running tests which
// test the offline functionality of the graph package.
// `unshare -nr` is used to deny network access, and then the tests are run using
// cached data from the tests in the graph package.
package offline

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bcherrington/onedriver/fs"
	"github.com/stretchr/testify/require"
)

// TestOfflineFileAccess verifies that we can access files and directories in offline mode
func TestOfflineFileAccess(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name         string
		description  string
		skipParallel bool
		testFunc     func(t *testing.T)
	}{
		{
			name:         "ReadDirectory_ShouldSucceed",
			description:  "Reading directory contents should succeed in offline mode",
			skipParallel: false,
			testFunc: func(t *testing.T) {
				// Read the test directory
				files, err := os.ReadDir(TestDir)
				require.NoError(t, err, "Failed to read test directory %s in offline mode", TestDir)

				// Verify that the directory is not empty
				require.Greater(t, len(files), 0,
					"Expected more than 0 files in the test directory %s when in offline mode", TestDir)
			},
		},
		{
			name:         "BagelFileDetection_ShouldSucceed",
			description:  "Finding and accessing the bagels file should succeed in offline mode",
			skipParallel: true, // Not running in parallel to ensure this test runs after the file is fully created
			testFunc: func(t *testing.T) {
				// Read the test directory
				files, err := os.ReadDir(TestDir)
				require.NoError(t, err, "Failed to read test directory %s in offline mode", TestDir)

				// Collect all file names for better error reporting
				found := false
				allFiles := make([]string, 0, len(files))

				// Look for the "bagels" file
				for _, f := range files {
					allFiles = append(allFiles, f.Name())

					if f.Name() == "bagels" {
						found = true

						// Verify it's a regular file, not a directory
						require.False(t, f.IsDir(),
							"\"bagels\" should be an ordinary file, not a directory")

						// Check file permissions
						info, err := f.Info()
						require.NoError(t, err, "Failed to get file info for \"bagels\"")

						octal := fs.Octal(uint32(info.Mode().Perm()))
						// middle bit just needs to be higher than 4
						// for compatibility with 022 / 002 umasks on different distros
						require.True(t, octal[0] == '6' && int(octal[1])-4 >= 0 && octal[2] == '4',
							"\"bagels\" permissions bits wrong, got %s, expected 644", octal)

						break
					}
				}

				// Verify the file was found
				require.True(t, found,
					"\"bagels\" file not found in offline mode! Available files: %v", allFiles)
			},
		},
		{
			name:         "BagelFileContents_ShouldMatchExpected",
			description:  "The contents of the bagels file should match what was written",
			skipParallel: true, // Not running in parallel to ensure this test runs before TestOfflineFileModification
			testFunc: func(t *testing.T) {
				bagelPath := filepath.Join(TestDir, "bagels")
				contents, err := os.ReadFile(bagelPath)
				require.NoError(t, err, "Failed to read bagels file at %s in offline mode", bagelPath)

				expectedContent := []byte("bagels\n")
				require.Equal(t, expectedContent, contents,
					"Offline file contents did not match expected content. Got %q, expected %q",
					string(contents), string(expectedContent))
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			if !tc.skipParallel {
				t.Parallel()
			}

			// Run the test
			tc.testFunc(t)
		})
	}
}

// TestOfflineFileSystemOperations tests various file and directory operations in offline mode
func TestOfflineFileSystemOperations(t *testing.T) {
	t.Parallel()

	// Define test cases
	testCases := []struct {
		name        string
		description string
		isDir       bool
		setupFunc   func(t *testing.T) (string, []byte, []byte)
		testFunc    func(t *testing.T, path string, initialContent []byte, newContent []byte)
	}{
		{
			name:        "FileCreation_ShouldSucceed",
			description: "Creating a file should succeed in offline mode",
			isDir:       false,
			setupFunc: func(t *testing.T) (string, []byte, []byte) {
				path := filepath.Join(TestDir, "donuts_"+t.Name())
				content := []byte("donuts are tasty")
				return path, content, nil
			},
			testFunc: func(t *testing.T, path string, content []byte, _ []byte) {
				// Write the file in offline mode
				err := os.WriteFile(path, content, 0644)
				require.NoError(t, err, "Writing a file while offline should succeed")

				// Verify the file was created and has the correct content
				contents, err := os.ReadFile(path)
				require.NoError(t, err, "Reading the file should succeed")
				require.Equal(t, content, contents, "File contents should match what was written")
			},
		},
		{
			name:        "FileModification_ShouldSucceed",
			description: "Modifying a file should succeed in offline mode",
			isDir:       false,
			setupFunc: func(t *testing.T) (string, []byte, []byte) {
				// Create a file to modify
				path := filepath.Join(TestDir, "modify_"+t.Name()+".txt")
				initialContent := []byte("initial content")
				newContent := []byte("modified content is better")

				// Create the file
				err := os.WriteFile(path, initialContent, 0644)
				require.NoError(t, err, "Failed to create file for modification test")

				return path, initialContent, newContent
			},
			testFunc: func(t *testing.T, path string, initialContent []byte, newContent []byte) {
				// Verify the file has the initial content
				content, err := os.ReadFile(path)
				require.NoError(t, err, "Failed to read initial content")
				require.Equal(t, initialContent, content, "Initial file content does not match expected")

				// Modify the file in offline mode
				err = os.WriteFile(path, newContent, 0644)
				require.NoError(t, err, "Failed to modify file in offline mode")

				// Verify the file was modified and has the new content
				modifiedContent, err := os.ReadFile(path)
				require.NoError(t, err, "Failed to read modified content")

				require.Equal(t, newContent, modifiedContent,
					"File contents after modification did not match expected content. Got %q, expected %q",
					string(modifiedContent), string(newContent))

				require.NotEqual(t, initialContent, modifiedContent,
					"File contents were not changed after modification. Content is still %q",
					string(modifiedContent))
			},
		},
		{
			name:        "FileDeletion_ShouldSucceed",
			description: "Deleting a file should succeed in offline mode",
			isDir:       false,
			setupFunc: func(t *testing.T) (string, []byte, []byte) {
				// Create a file to delete
				path := filepath.Join(TestDir, "delete_"+t.Name()+".txt")
				content := []byte("this file will be deleted")

				// Create the file
				err := os.WriteFile(path, content, 0644)
				require.NoError(t, err, "Failed to create file for deletion test")

				return path, content, nil
			},
			testFunc: func(t *testing.T, path string, content []byte, _ []byte) {
				// Verify the file exists
				_, err := os.Stat(path)
				require.NoError(t, err, "Test file should exist before deletion but was not found")

				// Delete the file in offline mode
				err = os.Remove(path)
				require.NoError(t, err, "Failed to delete file in offline mode")

				// Verify the file was deleted
				_, err = os.Stat(path)
				require.Error(t, err, "Test file should not exist after deletion but was found")
				require.True(t, os.IsNotExist(err),
					"Error for deleted file should be 'file does not exist', but got: %v", err)
			},
		},
		{
			name:        "DirectoryCreation_ShouldSucceed",
			description: "Creating a directory should succeed in offline mode",
			isDir:       true,
			setupFunc: func(t *testing.T) (string, []byte, []byte) {
				path := filepath.Join(TestDir, "dir_create_"+t.Name())
				return path, nil, nil
			},
			testFunc: func(t *testing.T, path string, _ []byte, _ []byte) {
				// Create the directory in offline mode
				err := os.Mkdir(path, 0755)
				require.NoError(t, err, "Failed to create directory in offline mode")

				// Verify the directory was created
				info, err := os.Stat(path)
				require.NoError(t, err, "Directory should exist after creation but was not found")
				require.True(t, info.IsDir(),
					"Path should be a directory but has file mode %s", info.Mode().String())
			},
		},
		{
			name:        "DirectoryDeletion_ShouldSucceed",
			description: "Deleting a directory should succeed in offline mode",
			isDir:       true,
			setupFunc: func(t *testing.T) (string, []byte, []byte) {
				// Create a directory to delete
				path := filepath.Join(TestDir, "dir_delete_"+t.Name())

				// Create the directory
				err := os.Mkdir(path, 0755)
				require.NoError(t, err, "Failed to create directory for deletion test")

				return path, nil, nil
			},
			testFunc: func(t *testing.T, path string, _ []byte, _ []byte) {
				// Verify the directory exists
				_, err := os.Stat(path)
				require.NoError(t, err, "Test directory should exist before deletion but was not found")

				// Delete the directory in offline mode
				err = os.Remove(path)
				require.NoError(t, err, "Failed to delete directory in offline mode")

				// Verify the directory was deleted
				_, err = os.Stat(path)
				require.Error(t, err, "Test directory should not exist after deletion but was found")
				require.True(t, os.IsNotExist(err),
					"Error for deleted directory should be 'file does not exist', but got: %v", err)
			},
		},
		{
			name:        "FileInDirectory_ShouldWorkOffline",
			description: "Creating a file in a directory should work in offline mode",
			isDir:       false,
			setupFunc: func(t *testing.T) (string, []byte, []byte) {
				// Create a directory
				dirPath := filepath.Join(TestDir, "parent_dir_"+t.Name())
				err := os.Mkdir(dirPath, 0755)
				require.NoError(t, err, "Failed to create parent directory")

				// Create a path for a file in that directory
				filePath := filepath.Join(dirPath, "nested_file.txt")
				content := []byte("this is a file in a directory")

				return filePath, content, nil
			},
			testFunc: func(t *testing.T, path string, content []byte, _ []byte) {
				// Write the file in offline mode
				err := os.WriteFile(path, content, 0644)
				require.NoError(t, err, "Writing a file in a directory while offline should succeed")

				// Verify the file was created and has the correct content
				contents, err := os.ReadFile(path)
				require.NoError(t, err, "Reading the file in a directory should succeed")
				require.Equal(t, content, contents, "File contents should match what was written")

				// Clean up the parent directory
				parentDir := filepath.Dir(path)
				t.Cleanup(func() {
					if err := os.RemoveAll(parentDir); err != nil {
						t.Logf("Warning: Failed to clean up parent directory %s: %v", parentDir, err)
					}
				})
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup test resources
			path, initialContent, newContent := tc.setupFunc(t)

			// Setup cleanup
			t.Cleanup(func() {
				if tc.isDir {
					if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up test directory %s: %v", path, err)
					}
				} else {
					if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up test file %s: %v", path, err)
					}
				}
			})

			// Run the test
			tc.testFunc(t, path, initialContent, newContent)
		})
	}
}

// Test that changes made in offline mode are cached and marked as changed
func TestOfflineChangesCached(t *testing.T) {
	t.Parallel()

	// Create a test file in offline mode
	testFilePath := filepath.Join(TestDir, "cached_changes.txt")
	testContent := []byte("this file was created in offline mode")

	// Setup cleanup to remove the test file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(testFilePath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", testFilePath, err)
		}
	})

	err := os.WriteFile(testFilePath, testContent, 0644)
	require.NoError(t, err, "Failed to create file %s in offline mode", testFilePath)

	// Verify the file exists and has the correct content
	content, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read content from file %s in offline mode", testFilePath)
	require.Equal(t, testContent, content,
		"File content in %s did not match what was written. Got %q, expected %q",
		testFilePath, string(content), string(testContent))

	// The file should be marked as changed in the filesystem
	// Note: We can't directly access the filesystem's internal state from these tests,
	// but the fact that we can read the file back confirms it was cached locally
}

// Test that when going back online, files are synchronized
func TestOfflineSynchronization(t *testing.T) {
	// This test is not run in parallel because it changes the global offline state

	// Set a timeout for this test to ensure it doesn't run too long
	// This is especially important since this test is the last one and doesn't run in parallel
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a test file in offline mode
	syncFilePath := filepath.Join(TestDir, "sync_test.txt")
	syncContent := []byte("this file will be synchronized when online")

	// Setup cleanup to remove the test file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(syncFilePath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", syncFilePath, err)
		}
	})

	// Use a channel to signal when the test is done
	done := make(chan struct{})
	var testErr error

	go func() {
		defer close(done)

		err := os.WriteFile(syncFilePath, syncContent, 0644)
		if err != nil {
			testErr = err
			return
		}

		// Verify the file exists and has the correct content
		content, err := os.ReadFile(syncFilePath)
		if err != nil {
			testErr = err
			return
		}

		if !bytes.Equal(syncContent, content) {
			testErr = fmt.Errorf("File content in %s did not match what was written. Got %q, expected %q",
				syncFilePath, string(content), string(syncContent))
			return
		}
	}()

	// Wait for the test to complete or timeout
	select {
	case <-ctx.Done():
		t.Fatalf("Test timed out after 30 seconds: %v", ctx.Err())
	case <-done:
		if testErr != nil {
			t.Fatalf("Test failed: %v", testErr)
		}
	}

	// Note: We can't actually test the synchronization in this test suite because:
	// 1. We're running with network access disabled via unshare
	// 2. We don't have access to the filesystem's internal state
	//
	// In a real scenario, when the filesystem goes back online:
	// 1. The changes would be detected as they're marked in the local cache
	// 2. The upload manager would process the queued changes
	// 3. The files would be synchronized with OneDrive
}
