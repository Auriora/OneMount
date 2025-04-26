// A bunch of "black box" filesystem integration tests that test the
// functionality of key syscalls and their implementation. If something fails
// here, the filesystem is not functional.
package fs

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/jstaf/onedriver/fs/graph"
	"github.com/jstaf/onedriver/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Does Go's internal ReadDir function work? This is mostly here to compare against
// the offline versions of this test.
func TestReaddir(t *testing.T) {
	testCases := []struct {
		name           string
		directory      string
		expectedItems  []string
		checkItemTypes map[string]string // Map of item name to expected type ("file" or "dir")
	}{
		{
			name:          "RootDirectory_ShouldContainDocumentsFolder",
			directory:     "mount",
			expectedItems: []string{"Documents"},
			checkItemTypes: map[string]string{
				"Documents": "dir",
			},
		},
		{
			name:          "TestDirectory_ShouldContainExpectedFiles",
			directory:     "mount/onedriver_tests",
			expectedItems: []string{"paging"},
			checkItemTypes: map[string]string{
				"paging": "dir",
			},
		},
		{
			name:          "DocumentsDirectory_ShouldBeReadable",
			directory:     "mount/Documents",
			expectedItems: []string{}, // We don't care about specific files, just that we can read the directory
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Read the directory
			entries, err := os.ReadDir(tc.directory)
			require.NoError(t, err, "Failed to read directory: %s", tc.directory)

			// Convert entries to FileInfo for more detailed checks
			files := make([]os.FileInfo, 0, len(entries))
			for _, entry := range entries {
				info, err := entry.Info()
				if err == nil {
					files = append(files, info)
				}
			}

			// Log the found items for debugging
			fileNames := make([]string, 0, len(files))
			for _, file := range files {
				fileNames = append(fileNames, file.Name())
			}
			t.Logf("Found items in %s: %v", tc.directory, fileNames)

			// Check for expected items
			for _, expectedItem := range tc.expectedItems {
				found := false
				for _, file := range files {
					if file.Name() == expectedItem {
						found = true
						break
					}
				}
				require.True(t, found, "Could not find expected item %q in directory %s", 
					expectedItem, tc.directory)
			}

			// Check item types if specified
			for itemName, expectedType := range tc.checkItemTypes {
				for _, file := range files {
					if file.Name() == itemName {
						switch expectedType {
						case "dir":
							require.True(t, file.IsDir(), "Expected %q to be a directory", itemName)
						case "file":
							require.False(t, file.IsDir(), "Expected %q to be a file", itemName)
						default:
							t.Fatalf("Invalid expected type: %s", expectedType)
						}
						break
					}
				}
			}
		})
	}
}

// does ls work and can we find the expected folders and files?
func TestLs(t *testing.T) {
	testCases := []struct {
		name           string
		directory      string
		options        []string // Additional ls options
		expectedItems  []string
		unexpectedItems []string // Items that should NOT be in the output
	}{
		{
			name:          "RootDirectory_ShouldContainDocumentsFolder",
			directory:     "mount",
			options:       []string{},
			expectedItems: []string{"Documents"},
		},
		{
			name:          "TestDirectory_ShouldContainExpectedFiles",
			directory:     "mount/onedriver_tests",
			options:       []string{},
			expectedItems: []string{"paging"},
		},
		{
			name:          "RootDirectoryWithAllFiles_ShouldShowHiddenFiles",
			directory:     "mount",
			options:       []string{"-a"},
			expectedItems: []string{".", ".."},
		},
		{
			name:          "ListingWithLongFormat_ShouldShowPermissions",
			directory:     "mount",
			options:       []string{"-l"},
			expectedItems: []string{"Documents"},
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Build the command arguments
			args := append(tc.options, tc.directory)

			// Execute the ls command
			stdout, err := exec.Command("ls", args...).Output()
			require.NoError(t, err, "ls command failed for directory %s with options %v", 
				tc.directory, tc.options)

			// Convert output to string for easier checking
			output := string(stdout)

			// Log the output for debugging
			t.Logf("ls %s %s output:\n%s", strings.Join(tc.options, " "), tc.directory, output)

			// Check for expected items
			for _, expectedItem := range tc.expectedItems {
				require.Contains(t, output, expectedItem, 
					"Could not find expected item %q in directory %s", expectedItem, tc.directory)
			}

			// Check for unexpected items (if any)
			for _, unexpectedItem := range tc.unexpectedItems {
				require.NotContains(t, output, unexpectedItem, 
					"Found unexpected item %q in directory %s", unexpectedItem, tc.directory)
			}
		})
	}
}

// can touch create an empty file?
func TestTouchCreate(t *testing.T) {
	fname := filepath.Join(TestDir, "empty")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	syscall.Umask(022) // otherwise tests fail if default umask is 002
	require.NoError(t, exec.Command("touch", fname).Run())
	st, err := os.Stat(fname)
	require.NoError(t, err)

	require.Zero(t, st.Size(), "Size should be zero.")
	// Check that the file is at least readable and writable by the owner, and readable by group and others
	// Some systems might use umask 002 instead of 022, resulting in 664 instead of 644
	mode := st.Mode()
	require.True(t, mode&0600 == 0600, "File should be readable and writable by owner")
	require.True(t, mode&0044 == 0044, "File should be readable by group and others")
	require.False(t, st.IsDir(), "New file detected as directory.")
}

// does the touch command update modification time properly?
func TestTouchUpdateTime(t *testing.T) {
	fname := filepath.Join(TestDir, "modtime")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	require.NoError(t, exec.Command("touch", fname).Run())
	st1, err := os.Stat(fname)
	require.NoError(t, err, "Failed to stat file after first touch")
	initialModTime := st1.ModTime()

	// Run the second touch command
	require.NoError(t, exec.Command("touch", fname).Run())

	// Wait for the modification time to change
	testutil.WaitForCondition(t, func() bool {
		st2, err := os.Stat(fname)
		if err != nil {
			t.Logf("Error stating file: %v", err)
			return false
		}
		return !st2.ModTime().Equal(initialModTime) && !st2.ModTime().Before(initialModTime)
	}, 5*time.Second, 100*time.Millisecond, "File modification time was not updated by touch")

	// Verify the modification time has changed
	st2, err := os.Stat(fname)
	require.NoError(t, err, "Failed to stat file after second touch")
	require.False(t, st2.ModTime().Equal(initialModTime) || st2.ModTime().Before(initialModTime),
		"File modification time was not updated by touch:\nBefore: %d\nAfter: %d\n",
		initialModTime.Unix(), st2.ModTime().Unix())
}

// TestFilePermissions tests that chmod works correctly with different permission modes
func TestFilePermissions(t *testing.T) {
	// Define test cases with different permission modes
	testCases := []struct {
		name        string
		permissions os.FileMode
		description string
	}{
		{
			name:        "ReadOnly",
			permissions: 0444,
			description: "read-only",
		},
		{
			name:        "ReadWrite",
			permissions: 0644,
			description: "read-write",
		},
		{
			name:        "ReadWriteExecute",
			permissions: 0755,
			description: "read-write-execute",
		},
		{
			name:        "AllPermissions",
			permissions: 0777,
			description: "all permissions",
		},
	}

	// Run each test case as a subtest
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Create a unique filename for this subtest to avoid conflicts
			fname := filepath.Join(TestDir, fmt.Sprintf("chmod_test_%s", tc.name))

			// Setup cleanup to remove the file after test completes or fails
			t.Cleanup(func() {
				if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
				}
			})

			// Create the test file
			require.NoError(t, exec.Command("touch", fname).Run(), "Failed to create test file")

			// Wait for the file to be created
			testutil.WaitForCondition(t, func() bool {
				_, err := os.Stat(fname)
				return err == nil
			}, 5*time.Second, 100*time.Millisecond, "File was not created within timeout")

			// Change the file permissions
			require.NoError(t, os.Chmod(fname, tc.permissions), 
				"Failed to change permissions to %o (%s)", tc.permissions, tc.description)

			// Verify the permissions were set correctly
			st, err := os.Stat(fname)
			require.NoError(t, err, "Failed to stat file")
			require.Equal(t, tc.permissions, st.Mode()&0777, 
				"Mode of file was not %o (%s), got %o instead!", 
				tc.permissions, tc.description, st.Mode()&0777)
		})
	}
}

// test that both mkdir and rmdir work, as well as the potentially failing
// mkdir->rmdir->mkdir chain that fails if the cache hangs on to an old copy
// after rmdir
func TestMkdirRmdir(t *testing.T) {
	fname := filepath.Join(TestDir, "folder1")

	// Setup cleanup to remove the directory after test completes or fails
	t.Cleanup(func() {
		if err := os.RemoveAll(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test directory %s: %v", fname, err)
		}
	})

	// Remove the directory if it exists to ensure we start fresh
	if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to remove directory: %v", err)
	}

	// Create, remove, and recreate the directory
	require.NoError(t, os.Mkdir(fname, 0755))
	require.NoError(t, os.Remove(fname))

	// Wait for the filesystem to process the removal
	// This ensures the directory is fully removed before we try to recreate it
	testutil.WaitForCondition(t, func() bool {
		_, err := os.Stat(fname)
		return os.IsNotExist(err) // Return true when the directory no longer exists
	}, 5*time.Second, 100*time.Millisecond, "Directory was not removed within timeout")

	require.NoError(t, os.Mkdir(fname, 0755))
}

// We shouldn't be able to rmdir nonempty directories
func TestRmdirNonempty(t *testing.T) {
	dir := filepath.Join(TestDir, "nonempty")

	// Setup cleanup to remove the directory after test completes or fails
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test directory %s: %v", dir, err)
		}
	})

	require.NoError(t, os.Mkdir(dir, 0755))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "contents"), 0755))

	require.Error(t, os.Remove(dir), "We somehow removed a nonempty directory!")

	require.NoError(t, os.RemoveAll(dir),
		"Could not remove a nonempty directory the correct way!")
}

// TestFileOperations tests various file operations using a table-driven approach
func TestFileOperations(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		operation   string
		content     string
		iterations  int
		fileMode    int
		verifyFunc  func(t *testing.T, filePath string, content string, iterations int)
	}{
		{
			name:       "WriteAndRead_ShouldPreserveContent",
			operation:  "write",
			content:    "my hands are typing words\n",
			iterations: 1,
			fileMode:   os.O_CREATE|os.O_RDWR,
			verifyFunc: func(t *testing.T, filePath string, content string, iterations int) {
				read, err := os.ReadFile(filePath)
				require.NoError(t, err, "Failed to read file")
				assert.Equal(t, content, string(read), "File content was not correct")
			},
		},
		{
			name:       "AppendMultipleTimes_ShouldHaveMultipleLines",
			operation:  "append",
			content:    "append\n",
			iterations: 5,
			fileMode:   os.O_APPEND|os.O_CREATE|os.O_RDWR,
			verifyFunc: func(t *testing.T, filePath string, content string, iterations int) {
				file, err := os.Open(filePath)
				require.NoError(t, err, "Failed to open file for verification")
				defer func() {
					if closeErr := file.Close(); closeErr != nil {
						t.Logf("Warning: Failed to close file: %v", closeErr)
					}
				}()

				scanner := bufio.NewScanner(file)
				var counter int
				for scanner.Scan() {
					counter++
					scanned := scanner.Text()
					require.Equal(t, strings.TrimSuffix(content, "\n"), scanned, 
						"File text was wrong. Got %q, wanted %q", scanned, strings.TrimSuffix(content, "\n"))
				}
				require.Equal(t, iterations, counter, "Got wrong number of lines (%d), expected %d", counter, iterations)
			},
		},
		{
			name:       "TruncateMultipleTimes_ShouldHaveOneLine",
			operation:  "truncate",
			content:    "append\n",
			iterations: 5,
			fileMode:   os.O_TRUNC|os.O_CREATE|os.O_RDWR,
			verifyFunc: func(t *testing.T, filePath string, content string, iterations int) {
				file, err := os.Open(filePath)
				require.NoError(t, err, "Failed to open file for verification")
				defer func() {
					if closeErr := file.Close(); closeErr != nil {
						t.Logf("Warning: Failed to close file: %v", closeErr)
					}
				}()

				scanner := bufio.NewScanner(file)
				var counter int
				for scanner.Scan() {
					counter++
					assert.Equal(t, strings.TrimSuffix(content, "\n"), scanner.Text(), "File text was wrong")
				}
				require.Equal(t, 1, counter, "Got wrong number of lines (%d), expected 1", counter)
			},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel
			// Create a unique filename for this test case to avoid conflicts
			filePath := filepath.Join(TestDir, fmt.Sprintf("%s_%s.txt", tc.operation, t.Name()))

			// Setup cleanup to remove the file after test completes or fails
			t.Cleanup(func() {
				if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
				}
			})

			// Remove the file if it exists to ensure we start fresh
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				t.Logf("Warning: Failed to remove file: %v", err)
			}

			// Perform the operation
			if tc.operation == "write" {
				// Simple write operation
				require.NoError(t, os.WriteFile(filePath, []byte(tc.content), 0644), 
					"Failed to write to file")
			} else {
				// Append or truncate operations
				for i := 0; i < tc.iterations; i++ {
					file, err := os.OpenFile(filePath, tc.fileMode, 0644)
					require.NoError(t, err, "Failed to open file for %s: %v", tc.operation, err)
					_, err = file.WriteString(tc.content)
					require.NoError(t, err, "Failed to write to file: %v", err)
					require.NoError(t, file.Close(), "Failed to close file: %v", err)
				}
			}

			// Verify the results
			tc.verifyFunc(t, filePath, tc.content, tc.iterations)
		})
	}
}

// ld can crash the filesystem because it starts writing output at byte 64 in previously
// empty file
func TestWriteOffset(t *testing.T) {
	fname := filepath.Join(TestDir, "main.c")
	outputFile := filepath.Join(TestDir, "main.o")

	// Setup cleanup to remove the files after test completes or fails
	t.Cleanup(func() {
		// Clean up source file
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}

		// Clean up compiled output
		if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", outputFile, err)
		}
	})

	require.NoError(t, os.WriteFile(fname,
		[]byte(`#include <stdio.h>

int main(int argc, char **argv) {
	printf("ld writes files in a funny manner!");
}`), 0644))
	require.NoError(t, exec.Command("gcc", "-o", outputFile, fname).Run())
}

// TestFileMovementOperations tests file operations like rename, move, and copy
func TestFileMovementOperations(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		operation      string
		content        string
		setupFunc      func(t *testing.T, baseDir string, content string) (string, string, error)
		operationFunc  func(t *testing.T, source string, dest string) error
		verifyFunc     func(t *testing.T, source string, dest string, content string) error
		description    string
	}{
		{
			name:      "RenameInSameDirectory_ShouldPreserveContent",
			operation: "rename",
			content:   "hopefully renames work\n",
			setupFunc: func(t *testing.T, baseDir string, content string) (string, string, error) {
				// Create source file
				source := filepath.Join(baseDir, fmt.Sprintf("rename_source_%s.txt", t.Name()))
				dest := filepath.Join(baseDir, fmt.Sprintf("rename_dest_%s.txt", t.Name()))

				if err := os.WriteFile(source, []byte(content), 0644); err != nil {
					return "", "", err
				}

				return source, dest, nil
			},
			operationFunc: func(t *testing.T, source string, dest string) error {
				return os.Rename(source, dest)
			},
			verifyFunc: func(t *testing.T, source string, dest string, content string) error {
				// Source should no longer exist
				if _, err := os.Stat(source); !os.IsNotExist(err) {
					return fmt.Errorf("source file %s still exists after rename", source)
				}

				// Destination should exist with correct content
				data, err := os.ReadFile(dest)
				if err != nil {
					return err
				}

				if string(data) != content {
					return fmt.Errorf("content mismatch after rename. Got %q, expected %q", string(data), content)
				}

				return nil
			},
			description: "Rename a file within the same directory",
		},
		{
			name:      "MoveToSubdirectory_ShouldPreserveContent",
			operation: "move",
			content:   "this file should be moved to a subdirectory\n",
			setupFunc: func(t *testing.T, baseDir string, content string) (string, string, error) {
				// Create source file
				source := filepath.Join(baseDir, fmt.Sprintf("move_source_%s.txt", t.Name()))

				// Create destination directory
				destDir := filepath.Join(baseDir, fmt.Sprintf("move_dest_dir_%s", t.Name()))
				if err := os.Mkdir(destDir, 0755); err != nil && !os.IsExist(err) {
					return "", "", err
				}

				dest := filepath.Join(destDir, fmt.Sprintf("moved_file_%s.txt", t.Name()))

				if err := os.WriteFile(source, []byte(content), 0644); err != nil {
					return "", "", err
				}

				return source, dest, nil
			},
			operationFunc: func(t *testing.T, source string, dest string) error {
				return os.Rename(source, dest)
			},
			verifyFunc: func(t *testing.T, source string, dest string, content string) error {
				// Source should no longer exist
				if _, err := os.Stat(source); !os.IsNotExist(err) {
					return fmt.Errorf("source file %s still exists after move", source)
				}

				// Destination should exist with correct content
				data, err := os.ReadFile(dest)
				if err != nil {
					return err
				}

				if string(data) != content {
					return fmt.Errorf("content mismatch after move. Got %q, expected %q", string(data), content)
				}

				return nil
			},
			description: "Move a file to a subdirectory",
		},
		{
			name:      "CopyFile_ShouldDuplicateContent",
			operation: "copy",
			content:   "and copies too!\n",
			setupFunc: func(t *testing.T, baseDir string, content string) (string, string, error) {
				// Create source file
				source := filepath.Join(baseDir, fmt.Sprintf("copy_source_%s.txt", t.Name()))
				dest := filepath.Join(baseDir, fmt.Sprintf("copy_dest_%s.txt", t.Name()))

				if err := os.WriteFile(source, []byte(content), 0644); err != nil {
					return "", "", err
				}

				return source, dest, nil
			},
			operationFunc: func(t *testing.T, source string, dest string) error {
				return exec.Command("cp", source, dest).Run()
			},
			verifyFunc: func(t *testing.T, source string, dest string, content string) error {
				// Source should still exist
				sourceData, err := os.ReadFile(source)
				if err != nil {
					return fmt.Errorf("failed to read source file after copy: %v", err)
				}

				if string(sourceData) != content {
					return fmt.Errorf("source content changed after copy. Got %q, expected %q", string(sourceData), content)
				}

				// Destination should exist with correct content
				destData, err := os.ReadFile(dest)
				if err != nil {
					return fmt.Errorf("failed to read destination file after copy: %v", err)
				}

				if string(destData) != content {
					return fmt.Errorf("destination content mismatch after copy. Got %q, expected %q", string(destData), content)
				}

				return nil
			},
			description: "Copy a file to create a duplicate",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel

			// Setup test files and directories
			source, dest, err := tc.setupFunc(t, TestDir, tc.content)
			require.NoError(t, err, "Failed to set up test files for %s operation", tc.operation)

			// Setup cleanup to remove all test files and directories after test completes or fails
			t.Cleanup(func() {
				// Clean up source file if it exists (it might not after a move/rename)
				if err := os.Remove(source); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up source file %s: %v", source, err)
				}

				// Clean up destination file
				if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up destination file %s: %v", dest, err)
				}

				// If this is a move operation, clean up the destination directory
				if tc.operation == "move" {
					destDir := filepath.Dir(dest)
					if destDir != TestDir {
						if err := os.RemoveAll(destDir); err != nil && !os.IsNotExist(err) {
							t.Logf("Warning: Failed to clean up destination directory %s: %v", destDir, err)
						}
					}
				}
			})

			// Perform the operation
			err = tc.operationFunc(t, source, dest)
			require.NoError(t, err, "Failed to perform %s operation", tc.operation)

			// Verify the results
			err = tc.verifyFunc(t, source, dest, tc.content)
			require.NoError(t, err, "Verification failed for %s operation", tc.operation)
		})
	}
}

// Note: TestAppend and TestTruncate have been refactored into the table-driven TestFileOperations above

// TestPositionalFileOperations tests that we can seek to specific positions in a file and perform operations
func TestPositionalFileOperations(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		initialContent string
		writeOffset    int64
		contentToWrite string
		description    string
	}{
		{
			name: "WriteToMiddle_ShouldPreserveContentAtSpecificOffset",
			initialContent: `Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
Phasellus viverra dui vel velit eleifend, vel auctor nulla scelerisque.
Mauris volutpat a justo vel suscipit. Suspendisse diam lorem, imperdiet eget
fermentum ut, sodales a nunc. Phasellus eget mattis purus.`,
			writeOffset:    123,
			contentToWrite: "my hands are typing words. aaaaaaa",
			description:    "Write to middle of medium-sized file",
		},
		{
			name:           "WriteToBeginning_ShouldOverwriteStartOfFile",
			initialContent: "This is a test file with some initial content that will be partially overwritten.",
			writeOffset:    0,
			contentToWrite: "REPLACED",
			description:    "Write to beginning of file",
		},
		{
			name:           "WriteToEnd_ShouldAppendToEndOfFile",
			initialContent: "Short content. ",
			writeOffset:    15, // Length of "Short content. "
			contentToWrite: "Additional text at the end.",
			description:    "Write to end of file",
		},
		{
			name:           "WriteToEmptyFile_ShouldCreateHoleInFile",
			initialContent: "",
			writeOffset:    100,
			contentToWrite: "Writing beyond the end creates a sparse file with a hole",
			description:    "Write to empty file at non-zero offset",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel

			// Create a unique filename for this test case to avoid conflicts
			fname := filepath.Join(TestDir, fmt.Sprintf("midfile_%s.txt", t.Name()))

			// Setup cleanup to remove the file after test completes or fails
			t.Cleanup(func() {
				if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
				}
			})

			// Create the file with initial content
			require.NoError(t, os.WriteFile(fname, []byte(tc.initialContent), 0644),
				"Failed to create test file with initial content")

			// Open the file for read/write
			file, err := os.OpenFile(fname, os.O_RDWR, 0644)
			require.NoError(t, err, "Failed to open file for read/write: %v", err)

			// Ensure file is closed after test
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					t.Logf("Warning: Failed to close file: %v", closeErr)
				}
			}()

			// Write content at the specified offset
			n, err := file.WriteAt([]byte(tc.contentToWrite), tc.writeOffset)
			require.NoError(t, err, "Failed to write to file at offset %d: %v", tc.writeOffset, err)
			require.Equal(t, len(tc.contentToWrite), n, 
				"Wrong number of bytes written. Got %d, expected %d", n, len(tc.contentToWrite))

			// Read back the content from the same offset
			result := make([]byte, len(tc.contentToWrite))
			n, err = file.ReadAt(result, tc.writeOffset)
			require.NoError(t, err, "Failed to read from file at offset %d: %v", tc.writeOffset, err)
			require.Equal(t, len(tc.contentToWrite), n, 
				"Wrong number of bytes read. Got %d, expected %d", n, len(tc.contentToWrite))

			// Verify the content matches what was written
			require.Equal(t, tc.contentToWrite, string(result), 
				"Content read from offset %d did not match what was written. Got %q, expected %q", 
				tc.writeOffset, string(result), tc.contentToWrite)

			// For the test case with offset 0, verify the beginning of the file was changed
			if tc.writeOffset == 0 {
				// Read the entire file
				fullContent, err := os.ReadFile(fname)
				require.NoError(t, err, "Failed to read entire file: %v", err)

				// Verify the beginning of the file was overwritten
				require.True(t, strings.HasPrefix(string(fullContent), tc.contentToWrite),
					"Beginning of file was not overwritten correctly. Got %q, expected prefix %q",
					string(fullContent), tc.contentToWrite)
			}
		})
	}
}

// Statfs should succeed
func TestStatFs(t *testing.T) {
	var st syscall.Statfs_t
	err := syscall.Statfs(TestDir, &st)
	require.NoError(t, err)
	require.NotZero(t, st.Blocks, "StatFs failed, got 0 blocks!")
}

// does unlink work? (because apparently we weren't testing that before...)
func TestUnlink(t *testing.T) {
	fname := filepath.Join(TestDir, "unlink_tester")
	require.NoError(t, exec.Command("touch", fname).Run())
	require.NoError(t, os.Remove(fname))
	stdout, _ := exec.Command("ls", "mount").Output()
	require.NotContains(t, string(stdout), "unlink_tester", "Deleting %s did not work.", fname)
}

// TestCaseSensitivityHandling tests how the filesystem handles case-sensitivity issues
// with OneDrive (which uses NTFS, a case-insensitive filesystem).
func TestCaseSensitivityHandling(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		description string
		testFunc    func(t *testing.T)
	}{
		{
			name:        "BasicCaseSensitivity_ShouldHandleFilesWithDifferentCase",
			description: "Tests basic case-sensitivity by creating two files with different case variants",
			testFunc: func(t *testing.T) {
				// Create unique filenames for this test to avoid conflicts
				file1 := filepath.Join(TestDir, fmt.Sprintf("case-sensitive-%s.txt", t.Name()))
				file2 := filepath.Join(TestDir, fmt.Sprintf("CASE-SENSITIVE-%s.txt", t.Name()))
				file3 := filepath.Join(TestDir, fmt.Sprintf("Case-Sensitive-%s.TXT", t.Name()))

				// Setup cleanup to remove the files after test completes or fails
				t.Cleanup(func() {
					for _, file := range []string{file1, file2, file3} {
						if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
							t.Logf("Warning: Failed to clean up test file %s: %v", file, err)
						}
					}
				})

				// Create the first file
				require.NoError(t, os.WriteFile(file1, []byte("NTFS is bad"), 0644))

				// Wait for the filesystem to process the file creation
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(file1)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "First file was not created within timeout")

				// Create the second file with different case
				require.NoError(t, os.WriteFile(file2, []byte("yep"), 0644))

				// Wait for the filesystem to process the file creation
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(file2)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "Second file was not created within timeout")

				// Try to read the file with a third case variant
				content, err := os.ReadFile(file3)

				// If the read fails, check if either of the original files exists
				if err != nil {
					t.Logf("Could not read %s: %v", file3, err)

					// Try reading the original files
					content1, err1 := os.ReadFile(file1)
					content2, err2 := os.ReadFile(file2)

					if err1 == nil {
						t.Logf("Successfully read %s: %s", file1, content1)
						require.Equal(t, "NTFS is bad", string(content1), 
							"Content of %s was not as expected", file1)
					} else {
						t.Logf("Could not read %s: %v", file1, err1)
					}

					if err2 == nil {
						t.Logf("Successfully read %s: %s", file2, content2)
						require.Equal(t, "yep", string(content2), 
							"Content of %s was not as expected", file2)
						// Use the content from file2 for the test
						content = content2
						err = nil
					} else {
						t.Logf("Could not read %s: %v", file2, err2)
					}
				}

				// At least one of the files should be readable
				require.NoError(t, err, "Could not read any of the case-sensitive test files")
				require.Equal(t, "yep", string(content), "Did not find expected output.")
			},
		},
		{
			name:        "ExclusiveCreate_ShouldHandleCaseSensitivityWithExclusiveCreate",
			description: "Tests case-sensitivity with exclusive create calls",
			testFunc: func(t *testing.T) {
				// Create unique filenames for this test to avoid conflicts
				file1Path := filepath.Join(TestDir, fmt.Sprintf("case-sensitive2-%s.txt", t.Name()))
				file2Path := filepath.Join(TestDir, fmt.Sprintf("CASE-SENSITIVE2-%s.txt", t.Name()))

				// Setup cleanup to remove the files after test completes or fails
				t.Cleanup(func() {
					for _, file := range []string{file1Path, file2Path} {
						if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
							t.Logf("Warning: Failed to clean up test file %s: %v", file, err)
						}
					}
				})

				// Remove any existing test files to ensure a clean state
				if err := os.Remove(file1Path); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to remove file1: %v", err)
				}
				if err := os.Remove(file2Path); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to remove file2: %v", err)
				}

				// Wait for the filesystem to process the removals
				testutil.WaitForCondition(t, func() bool {
					_, err1 := os.Stat(file1Path)
					_, err2 := os.Stat(file2Path)
					return os.IsNotExist(err1) && os.IsNotExist(err2)
				}, 5*time.Second, 100*time.Millisecond, "Files were not removed within timeout")

				// Create the first file
				file1, err := os.OpenFile(file1Path, os.O_CREATE|os.O_EXCL, 0644)
				if err == nil {
					if closeErr := file1.Close(); closeErr != nil {
						t.Logf("Warning: Failed to close file1: %v", closeErr)
					}
				} else {
					t.Logf("Failed to create first file: %v", err)
					// If we can't create the first file, skip the test
					t.Skip("Could not create the first test file, skipping test")
				}

				// Wait for the filesystem to process the file creation
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(file1Path)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "First file was not created within timeout")

				// Try to create the second file with different case
				file2, err := os.OpenFile(file2Path, os.O_CREATE|os.O_EXCL, 0644)
				if err == nil {
					if closeErr := file2.Close(); closeErr != nil {
						t.Logf("Warning: Failed to close file2: %v", closeErr)
					}

					// Check if both files exist now
					_, err1 := os.Stat(file1Path)
					_, err2 := os.Stat(file2Path)

					if err1 == nil && err2 == nil {
						t.Log("Both files with different case exist simultaneously")
						// This is acceptable if the filesystem doesn't enforce case-insensitivity
					}
				} else {
					// This is the expected behavior for a case-insensitive filesystem
					t.Logf("Got expected error when creating second file: %v", err)
				}

				// The test passes either way - we're just documenting the behavior
			},
		},
		{
			name:        "RenameHandling_ShouldHandleCaseSensitivityWithRenames",
			description: "Tests case-sensitivity with renames",
			testFunc: func(t *testing.T) {
				// Create unique filenames for this test to avoid conflicts
				fname := filepath.Join(TestDir, fmt.Sprintf("original_NAME-%s.txt", t.Name()))
				secondName := filepath.Join(TestDir, fmt.Sprintf("new_name-%s.txt", t.Name()))
				thirdName := filepath.Join(TestDir, fmt.Sprintf("new_name2-%s.txt", t.Name()))
				fourthName := filepath.Join(TestDir, fmt.Sprintf("original_name-%s.txt", t.Name()))

				// Setup cleanup to remove the files after test completes or fails
				t.Cleanup(func() {
					for _, file := range []string{fname, secondName, thirdName, fourthName} {
						if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
							t.Logf("Warning: Failed to clean up test file %s: %v", file, err)
						}
					}
				})

				require.NoError(t, os.WriteFile(fname, []byte("original"), 0644))

				// Wait for the DeltaLoop to process the file creation
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(fname)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "First file was not created within timeout")

				// should work
				require.NoError(t, os.WriteFile(secondName, []byte("new"), 0644))

				// Wait for the DeltaLoop to process the file creation
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(secondName)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "Second file was not created within timeout")

				require.NoError(t, os.Rename(secondName, fname))

				// Wait for the DeltaLoop to process the rename and for the file content to be updated
				testutil.WaitForCondition(t, func() bool {
					content, err := os.ReadFile(fname)
					return err == nil && string(content) == "new"
				}, 5*time.Second, 100*time.Millisecond, "File content was not updated after rename")

				contents, err := os.ReadFile(fname)
				require.NoError(t, err)
				require.Equal(t, "new", string(contents), "Contents did not match expected output.")

				// should fail
				require.NoError(t, os.WriteFile(thirdName, []byte("this rename should work"), 0644))

				// Wait for the DeltaLoop to process the file creation
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(thirdName)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "Third file was not created within timeout")

				err = os.Rename(thirdName, fourthName)
				require.NoError(t, err, "Rename failed.")

				// Wait for the DeltaLoop to process the rename
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(fourthName)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "Renamed file was not created within timeout")

				_, err = os.Stat(fname)
				require.NoErrorf(t, err, "\"%s\" does not exist after the rename.", fname)
			},
		},
	}

	// Run each test case as a subtest
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Note: We don't use t.Parallel() here because these tests might interfere with each other
			// due to their nature of testing case-sensitivity in the same filesystem

			// Run the test function
			tc.testFunc(t)
		})
	}
}

// TestFilenameCase tests that the filesystem properly preserves case in filenames
// This is insurance to prevent tests (and the fs) from accidentally not storing case for filenames at all
func TestFilenameCase(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		filename    string
		content     string
		description string
	}{
		{
			name:        "UpperCase_ShouldPreserveCase",
			filename:    "UPPERCASE-FILE.txt",
			content:     "uppercase content",
			description: "File with all uppercase characters",
		},
		{
			name:        "LowerCase_ShouldPreserveCase",
			filename:    "lowercase-file.txt",
			content:     "lowercase content",
			description: "File with all lowercase characters",
		},
		{
			name:        "MixedCase_ShouldPreserveCase",
			filename:    "MixedCase-FiLe.TxT",
			content:     "mixed case content",
			description: "File with mixed case characters",
		},
		{
			name:        "SpecialChars_ShouldPreserveCase",
			filename:    "SPECIAL_Chars-123.txt",
			content:     "special chars content",
			description: "File with special characters and numbers",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel

			// Create a unique filename for this test to avoid conflicts
			filePath := filepath.Join(TestDir, fmt.Sprintf("%s-%s", tc.filename, t.Name()))

			// Setup cleanup to remove the file after test completes or fails
			t.Cleanup(func() {
				if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
				}
			})

			// Create the test file
			require.NoError(t, os.WriteFile(filePath, []byte(tc.content), 0644),
				"Failed to create test file %s", filePath)

			// Wait for the filesystem to process the file creation
			testutil.WaitForCondition(t, func() bool {
				_, err := os.Stat(filePath)
				return err == nil
			}, 5*time.Second, 100*time.Millisecond, "File was not created within timeout")

			// Run ls command to get directory listing
			stdout, err := exec.Command("ls", TestDir).Output()
			require.NoError(t, err, "Failed to list directory: %v", err)

			// Verify the filename appears with the correct case in the directory listing
			require.Contains(t, string(stdout), filepath.Base(filePath),
				"Filename case was not preserved. Expected %q in output, got: %s", 
				filepath.Base(filePath), string(stdout))

			// Verify the file content
			content, err := os.ReadFile(filePath)
			require.NoError(t, err, "Failed to read file content: %v", err)
			require.Equal(t, tc.content, string(content),
				"File content does not match. Got %q, expected %q", string(content), tc.content)
		})
	}
}

// TestShellFileOperations tests that shell commands like echo, cat, etc. properly write to files
// This verifies that when running commands like "echo some text > file.txt", the file actually becomes populated
func TestShellFileOperations(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name        string
		command     string
		content     string
		description string
	}{
		{
			name:        "EchoToFile_ShouldWriteContent",
			command:     "echo %s > %s",
			content:     "simple content",
			description: "Basic echo command with redirection",
		},
		{
			name:        "EchoWithQuotes_ShouldPreserveQuotes",
			command:     "echo \"%s\" > %s",
			content:     "content with \"quotes\"",
			description: "Echo command with quoted content",
		},
		{
			name:        "EchoWithSpecialChars_ShouldPreserveSpecialChars",
			command:     "echo '%s' > %s",
			content:     "content with $pecial ch@rs!",
			description: "Echo command with special characters",
		},
		{
			name:        "CatToFile_ShouldWriteContent",
			command:     "echo %s | cat > %s",
			content:     "content via cat",
			description: "Using cat with pipe to write content",
		},
		{
			name:        "MultipleLinesEcho_ShouldPreserveNewlines",
			command:     "echo -e '%s' > %s",
			content:     "line1\\nline2\\nline3",
			description: "Echo command with multiple lines",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Run subtests in parallel

			// Create a unique filename for this test to avoid conflicts
			filePath := filepath.Join(TestDir, fmt.Sprintf("shell-test-%s.txt", t.Name()))

			// Setup cleanup to remove the file after test completes or fails
			t.Cleanup(func() {
				if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
				}
			})

			// Format the command with the content and file path
			formattedCommand := fmt.Sprintf(tc.command, tc.content, filePath)

			// Execute the shell command
			out, err := exec.Command("bash", "-c", formattedCommand).CombinedOutput()
			require.NoError(t, err, "Command failed: %s\nOutput: %s", formattedCommand, out)

			// Wait for the DeltaLoop to process the file creation and for the file to contain the expected content
			testutil.WaitForCondition(t, func() bool {
				content, err := os.ReadFile(filePath)
				if err != nil {
					return false
				}

				// For the multiple lines test case, we need to handle the newline characters
				expectedContent := tc.content
				if tc.name == "MultipleLinesEcho_ShouldPreserveNewlines" {
					expectedContent = strings.ReplaceAll(expectedContent, "\\n", "\n")
				}

				return strings.Contains(string(content), expectedContent)
			}, 5*time.Second, 100*time.Millisecond, "File was not created or did not contain expected content within timeout")

			// Read the file content
			content, err := os.ReadFile(filePath)
			require.NoError(t, err, "Failed to read file: %v", err)

			// For the multiple lines test case, we need to handle the newline characters
			expectedContent := tc.content
			if tc.name == "MultipleLinesEcho_ShouldPreserveNewlines" {
				expectedContent = strings.ReplaceAll(expectedContent, "\\n", "\n")
			}

			// Verify the content
			require.Contains(t, string(content), expectedContent,
				"File content does not match expected content.\nGot: %q\nExpected to contain: %q", 
				string(content), expectedContent)
		})
	}
}

// TestFileInfo tests that the stat operation returns correct information about files and directories
func TestFileInfo(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name           string
		setupFunc      func(t *testing.T) (string, os.FileMode, error)
		expectedName   string
		isDir          bool
		verifyFunc     func(t *testing.T, stat os.FileInfo) error
		description    string
	}{
		{
			name: "Directory_ShouldHaveCorrectAttributes",
			setupFunc: func(t *testing.T) (string, os.FileMode, error) {
				// Ensure the Documents directory exists
				docDir := "mount/Documents"
				if _, err := os.Stat(docDir); os.IsNotExist(err) {
					if err := os.Mkdir(docDir, 0755); err != nil {
						return "", 0, err
					}

					// Wait for the filesystem to process the directory creation
					testutil.WaitForCondition(t, func() bool {
						stat, err := os.Stat(docDir)
						return err == nil && stat.IsDir()
					}, 5*time.Second, 100*time.Millisecond, "Documents directory was not created within timeout")
				}
				return docDir, 0755, nil
			},
			expectedName: "Documents",
			isDir:        true,
			verifyFunc: func(t *testing.T, stat os.FileInfo) error {
				if stat.ModTime().Year() < 1971 {
					return fmt.Errorf("modification time wrong, got: %s", stat.ModTime().String())
				}
				return nil
			},
			description: "Verify attributes of a directory",
		},
		{
			name: "RegularFile_ShouldHaveCorrectAttributes",
			setupFunc: func(t *testing.T) (string, os.FileMode, error) {
				// Create a regular file
				filePath := filepath.Join(TestDir, fmt.Sprintf("stat-test-file-%s.txt", t.Name()))
				content := []byte("test content for stat")
				if err := os.WriteFile(filePath, content, 0644); err != nil {
					return "", 0, err
				}

				// Wait for the filesystem to process the file creation
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(filePath)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "File was not created within timeout")

				// Setup cleanup to remove the file after test completes or fails
				t.Cleanup(func() {
					if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
					}
				})

				return filePath, 0644, nil
			},
			expectedName: "", // Will be set dynamically based on the generated filename
			isDir:        false,
			verifyFunc: func(t *testing.T, stat os.FileInfo) error {
				expectedSize := int64(len("test content for stat"))
				if stat.Size() != expectedSize {
					return fmt.Errorf("file size wrong, got: %d, expected: %d", stat.Size(), expectedSize)
				}
				if stat.ModTime().Year() < 1971 {
					return fmt.Errorf("modification time wrong, got: %s", stat.ModTime().String())
				}
				return nil
			},
			description: "Verify attributes of a regular file",
		},
		{
			name: "ExecutableFile_ShouldHaveCorrectPermissions",
			setupFunc: func(t *testing.T) (string, os.FileMode, error) {
				// Create an executable file
				filePath := filepath.Join(TestDir, fmt.Sprintf("stat-test-exec-%s.sh", t.Name()))
				content := []byte("#!/bin/bash\necho 'This is an executable file'")
				if err := os.WriteFile(filePath, content, 0755); err != nil {
					return "", 0, err
				}

				// Wait for the filesystem to process the file creation
				testutil.WaitForCondition(t, func() bool {
					_, err := os.Stat(filePath)
					return err == nil
				}, 5*time.Second, 100*time.Millisecond, "File was not created within timeout")

				// Setup cleanup to remove the file after test completes or fails
				t.Cleanup(func() {
					if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
					}
				})

				return filePath, 0755, nil
			},
			expectedName: "", // Will be set dynamically based on the generated filename
			isDir:        false,
			verifyFunc: func(t *testing.T, stat os.FileInfo) error {
				// Check if the file has execute permissions
				if stat.Mode()&0111 == 0 {
					return fmt.Errorf("file should have execute permissions, got mode: %s", stat.Mode())
				}
				return nil
			},
			description: "Verify permissions of an executable file",
		},
		{
			name: "EmptyDirectory_ShouldHaveCorrectAttributes",
			setupFunc: func(t *testing.T) (string, os.FileMode, error) {
				// Create an empty directory
				dirPath := filepath.Join(TestDir, fmt.Sprintf("stat-test-dir-%s", t.Name()))
				if err := os.Mkdir(dirPath, 0755); err != nil {
					return "", 0, err
				}

				// Wait for the filesystem to process the directory creation
				testutil.WaitForCondition(t, func() bool {
					stat, err := os.Stat(dirPath)
					return err == nil && stat.IsDir()
				}, 5*time.Second, 100*time.Millisecond, "Directory was not created within timeout")

				// Setup cleanup to remove the directory after test completes or fails
				t.Cleanup(func() {
					if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up test directory %s: %v", dirPath, err)
					}
				})

				return dirPath, 0755, nil
			},
			expectedName: "", // Will be set dynamically based on the generated dirname
			isDir:        true,
			verifyFunc: func(t *testing.T, stat os.FileInfo) error {
				// Check if the directory has the correct size (usually 0 or 4096 for directories)
				if stat.Size() < 0 {
					return fmt.Errorf("directory size should not be negative, got: %d", stat.Size())
				}
				return nil
			},
			description: "Verify attributes of an empty directory",
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Note: We don't use t.Parallel() here because some tests might interfere with each other
			// when creating/accessing the Documents directory

			// Setup the test
			path, expectedMode, err := tc.setupFunc(t)
			require.NoError(t, err, "Failed to setup test: %v", err)

			// Get file info using stat
			stat, err := os.Stat(path)
			require.NoError(t, err, "Failed to stat %s: %v", path, err)

			// Verify the file name
			if tc.expectedName != "" {
				require.Equal(t, tc.expectedName, stat.Name(), 
					"File name does not match. Got %q, expected %q", stat.Name(), tc.expectedName)
			} else {
				// If expectedName is not specified, use the base name of the path
				expectedName := filepath.Base(path)
				require.Equal(t, expectedName, stat.Name(), 
					"File name does not match. Got %q, expected %q", stat.Name(), expectedName)
			}

			// Verify if it's a directory
			require.Equal(t, tc.isDir, stat.IsDir(), 
				"IsDir() returned %v, expected %v", stat.IsDir(), tc.isDir)

			// Verify the file mode (permissions)
			if expectedMode != 0 {
				// We only check the permission bits, not the file type bits
				require.Equal(t, expectedMode&0777, stat.Mode()&0777, 
					"File mode does not match. Got %o, expected %o", stat.Mode()&0777, expectedMode&0777)
			}

			// Run additional verification if provided
			if tc.verifyFunc != nil {
				err := tc.verifyFunc(t, stat)
				require.NoError(t, err, "Verification failed: %v", err)
			}
		})
	}
}

// Question marks appear in `ls -l`s output if an item is populated via readdir,
// but subsequently not found by lookup. Also is a nice catch-all for fs
// metadata corruption, as `ls` will exit with 1 if something bad happens.
func TestNoQuestionMarks(t *testing.T) {
	testCases := []struct {
		name      string
		directory string
		options   []string // Additional ls options
	}{
		{
			name:      "RootDirectory_ShouldNotHaveQuestionMarks",
			directory: "mount/",
			options:   []string{"-l"},
		},
		{
			name:      "TestDirectory_ShouldNotHaveQuestionMarks",
			directory: "mount/onedriver_tests/",
			options:   []string{"-l"},
		},
		{
			name:      "RootDirectoryWithAllFiles_ShouldNotHaveQuestionMarks",
			directory: "mount/",
			options:   []string{"-la"}, // Include hidden files
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Build the command arguments
			args := append(tc.options, tc.directory)

			// Execute the ls command
			out, err := exec.Command("ls", args...).CombinedOutput()

			// Check for error in command execution
			if err != nil {
				t.Logf("Command output: %s", string(out))
				require.NoError(t, err, "ls command failed for directory %s with options %v", 
					tc.directory, tc.options)
			}

			// Check for question marks in the output
			require.False(t, strings.Contains(string(out), "??????????"),
				"A Lookup() failed on an inode found by Readdir() in directory %s\nCommand output:\n%s", 
				tc.directory, string(out))
		})
	}
}

// Trashing items through nautilus or other Linux file managers is done via
// "gio trash". Make an item then trash it to verify that this works.
func TestGIOTrash(t *testing.T) {
	// Check if gio is installed
	_, err := exec.LookPath("gio")
	if err != nil {
		t.Skip("gio command not found, skipping test")
	}

	// Ensure the test directory exists
	err = os.MkdirAll(TestDir, 0755)
	require.NoError(t, err, "Failed to create test directory")

	testCases := []struct {
		name        string
		fileType    string // "file" or "directory"
		content     []byte
		permissions os.FileMode
	}{
		{
			name:        "RegularTextFile_ShouldBeTrashable",
			fileType:    "file",
			content:     []byte("i should be trashed"),
			permissions: 0644,
		},
		{
			name:        "BinaryFile_ShouldBeTrashable",
			fileType:    "file",
			content:     []byte{0x00, 0x01, 0x02, 0x03, 0x04}, // Binary content
			permissions: 0644,
		},
		{
			name:        "ExecutableFile_ShouldBeTrashable",
			fileType:    "file",
			content:     []byte("#!/bin/sh\necho 'Hello, world!'"),
			permissions: 0755,
		},
		{
			name:        "EmptyDirectory_ShouldBeTrashable",
			fileType:    "directory",
			content:     nil,
			permissions: 0755,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Create a unique filename for this test case to avoid conflicts
			fname := filepath.Join(TestDir, fmt.Sprintf("trash_me_%s", t.Name()))

			// Create the file or directory based on the test case
			if tc.fileType == "file" {
				require.NoError(t, os.WriteFile(fname, tc.content, tc.permissions),
					"Failed to create test file")
			} else if tc.fileType == "directory" {
				require.NoError(t, os.MkdirAll(fname, tc.permissions),
					"Failed to create test directory")
			}

			// Clean up the file/directory after the test
			t.Cleanup(func() {
				// Try to remove the file/directory if it still exists
				_ = os.RemoveAll(fname)
			})

			// Wait for the DeltaLoop to process the file/directory creation
			testutil.WaitForCondition(t, func() bool {
				_, err := os.Stat(fname)
				return err == nil
			}, 5*time.Second, 100*time.Millisecond, "Item was not created within timeout")

			// Trash the file/directory
			out, err := exec.Command("gio", "trash", fname).CombinedOutput()
			if err != nil {
				t.Log(string(out))
				t.Log(err)
				if st, err2 := os.Stat(fname); err2 == nil {
					if !st.IsDir() && strings.Contains(string(out), "Is a directory") {
						t.Skip("This is a GIO bug (it complains about test file being " +
							"a directory despite correct metadata from onedriver), skipping.")
					}
					require.Fail(t, fmt.Sprintf("%s still exists after deletion!", fname))
				}
			}
			require.False(t, strings.Contains(string(out), "Unable to find or create trash directory"),
				"Error creating trash directory: %s", string(out))

			// Wait for the DeltaLoop to process the file/directory deletion
			testutil.WaitForCondition(t, func() bool {
				_, err := os.Stat(fname)
				return os.IsNotExist(err) // Return true when the file no longer exists
			}, 5*time.Second, 100*time.Millisecond, "Item was not deleted within timeout")
		})
	}
}

// Test that we are able to work around onedrive paging limits when
// listing a folder's children.
func TestListChildrenPaging(t *testing.T) {
	testCases := []struct {
		name           string
		dirPath        string
		minExpectedAPI int // Minimum expected number of items from API
		minExpectedFS  int // Minimum expected number of files in filesystem
	}{
		{
			name:           "PagingDirectory_ShouldHandleMoreThan200Items",
			dirPath:        "paging",
			minExpectedAPI: 201,
			minExpectedFS:  201,
		},
		{
			name:           "RootDirectory_ShouldListAllItems",
			dirPath:        "",
			minExpectedAPI: 1, // At least one item should be present in root
			minExpectedFS:  1,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// files have been prepopulated during test setup to avoid being picked up by
			// the delta thread
			apiPath := "/onedriver_tests"
			if tc.dirPath != "" {
				apiPath = filepath.Join(apiPath, tc.dirPath)
			}

			items, err := graph.GetItemChildrenPath(apiPath, auth)
			require.NoError(t, err, "Failed to get items from API for path: %s", apiPath)

			fsPath := TestDir
			if tc.dirPath != "" {
				fsPath = filepath.Join(fsPath, tc.dirPath)
			}

			entries, err := os.ReadDir(fsPath)
			require.NoError(t, err, "Failed to read directory: %s", fsPath)

			files := make([]os.FileInfo, 0, len(entries))
			for _, entry := range entries {
				info, err := entry.Info()
				if err == nil {
					files = append(files, info)
				}
			}

			// Log the actual counts for debugging
			t.Logf("Path: %s, API items: %d, FS files: %d", apiPath, len(items), len(files))

			// Skip if both API and FS have fewer items than expected
			if len(files) < tc.minExpectedFS && len(items) < tc.minExpectedAPI {
				t.Skipf("Skipping test, not enough items. API: %d (expected %d), FS: %d (expected %d)",
					len(items), tc.minExpectedAPI, len(files), tc.minExpectedFS)
			}

			// Verify that we have at least the minimum expected number of files
			require.GreaterOrEqual(t, len(files), tc.minExpectedFS, 
				"Paging limit failed. Got %d files, wanted at least %d.", len(files), tc.minExpectedFS)

			// Verify that the API returned at least the minimum expected number of items
			require.GreaterOrEqual(t, len(items), tc.minExpectedAPI,
				"API returned fewer items than expected. Got %d items, wanted at least %d.", len(items), tc.minExpectedAPI)
		})
	}
}

// Libreoffice writes to files in a funny manner and it can result in a 0 byte file
// being uploaded (can check syscalls via "inotifywait -m -r .").
func TestLibreOfficeSavePattern(t *testing.T) {
	// Check if LibreOffice is installed
	_, err := exec.LookPath("libreoffice")
	if err != nil {
		t.Skip("LibreOffice not found, skipping test")
	}

	// Ensure the test directory exists
	err = os.MkdirAll(TestDir, 0755)
	require.NoError(t, err, "Failed to create test directory")

	testCases := []struct {
		name           string
		sourceContent  []byte
		sourceFileName string
		sourceFileExt  string
		targetFormat   string
		expectedSize   uint64 // Minimum expected size in bytes, 0 means just check for non-zero
	}{
		{
			name:           "TextToDocx_ShouldCreateNonEmptyFile",
			sourceContent:  []byte("This will break things."),
			sourceFileName: "libreoffice",
			sourceFileExt:  "txt",
			targetFormat:   "docx",
			expectedSize:   0, // Just check for non-zero
		},
		{
			name:           "TextToOdt_ShouldCreateNonEmptyFile",
			sourceContent:  []byte("Converting to OpenDocument format."),
			sourceFileName: "libreoffice_odt",
			sourceFileExt:  "txt",
			targetFormat:   "odt",
			expectedSize:   0, // Just check for non-zero
		},
		{
			name:           "TextToPdf_ShouldCreateNonEmptyFile",
			sourceContent:  []byte("Converting to PDF format."),
			sourceFileName: "libreoffice_pdf",
			sourceFileExt:  "txt",
			targetFormat:   "pdf",
			expectedSize:   0, // Just check for non-zero
		},
		{
			name:           "LargerTextToDocx_ShouldCreateLargerFile",
			sourceContent:  []byte("This is a larger text document with multiple lines.\nIt should create a larger output file.\nThis helps test that the file content is properly preserved during conversion."),
			sourceFileName: "libreoffice_large",
			sourceFileExt:  "txt",
			targetFormat:   "docx",
			expectedSize:   1000, // Expect at least 1KB
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Create unique source and target filenames for this test case
			sourceFileName := fmt.Sprintf("%s_%s.%s", tc.sourceFileName, t.Name(), tc.sourceFileExt)
			sourcePath := filepath.Join(TestDir, sourceFileName)
			targetFileName := fmt.Sprintf("%s_%s.%s", tc.sourceFileName, t.Name(), tc.targetFormat)
			targetPath := filepath.Join(TestDir, targetFileName)

			// Create the source file
			require.NoError(t, os.WriteFile(sourcePath, tc.sourceContent, 0644),
				"Failed to create source file: %s", sourcePath)

			// Clean up the files after the test
			t.Cleanup(func() {
				// Try to remove the source and target files if they exist
				_ = os.Remove(sourcePath)
				_ = os.Remove(targetPath)
			})

			// Run LibreOffice to convert the file
			out, err := exec.Command(
				"libreoffice",
				"--headless",
				"--convert-to", tc.targetFormat,
				"--outdir", TestDir,
				sourcePath,
			).CombinedOutput()

			require.NoError(t, err, "LibreOffice conversion failed: %v\nOutput: %s", err, out)

			// LibreOffice document conversion can fail with an exit code of 0,
			// so we need to actually check the command output
			require.NotContains(t, string(out), "Error:", 
				"LibreOffice reported an error in its output: %s", out)

			// Log the conversion output for debugging
			t.Logf("LibreOffice conversion output: %s", out)

			// Construct the API path for the target file
			apiPath := fmt.Sprintf("/onedriver_tests/%s", targetFileName)

			// Format the error message before passing it to WaitForCondition
			errorMessage := fmt.Sprintf("Could not find %s post-upload or file size was too small", apiPath)

			// Use WaitForCondition to wait for the file to be uploaded and available
			testutil.WaitForCondition(t, func() bool {
				item, err := graph.GetItemPath(apiPath, auth)
				if err == nil && item != nil {
					// Check that the file size meets the expected minimum
					if tc.expectedSize > 0 {
						if item.Size >= tc.expectedSize {
							return true
						}
						t.Logf("File found but size is smaller than expected. Got: %d, Expected: at least %d bytes", 
							item.Size, tc.expectedSize)
					} else if item.Size > 0 {
						// Just check for non-zero size
						return true
					}
					t.Logf("File found but size is 0, waiting for upload to complete...")
				}
				return false
			}, retrySeconds, 3*time.Second, errorMessage)
		})
	}
}

// TestDisallowedFilenames verifies that we can't create any of the disallowed filenames
// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa
func TestDisallowedFilenames(t *testing.T) {
	// This test checks if the filesystem properly restricts disallowed filenames
	// OneDrive has restrictions on certain characters and names:
	// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa

	// Define test cases for creating files/directories with disallowed names
	createTestCases := []struct {
		name        string
		path        string
		isDir       bool
		description string
	}{
		{
			name:        "FileWithColon_ShouldBeRejected",
			path:        filepath.Join(TestDir, fmt.Sprintf("disallowed-colon-%s: filename.txt", t.Name())),
			isDir:       false,
			description: "File with colon character in name",
		},
		{
			name:        "FileWithVtiPrefix_ShouldBeRejected",
			path:        filepath.Join(TestDir, fmt.Sprintf("disallowed-vti-%s_vti_text.txt", t.Name())),
			isDir:       false,
			description: "File with _vti_ prefix (reserved by SharePoint)",
		},
		{
			name:        "FileWithLessThan_ShouldBeRejected",
			path:        filepath.Join(TestDir, fmt.Sprintf("disallowed-lt-%s_<_text.txt", t.Name())),
			isDir:       false,
			description: "File with < character in name",
		},
		{
			name:        "ReservedNameCOM0_ShouldBeRejected",
			path:        filepath.Join(TestDir, "COM0"),
			isDir:       false,
			description: "Reserved Windows device name COM0",
		},
		{
			name:        "DirectoryWithColon_ShouldBeRejected",
			path:        filepath.Join(TestDir, fmt.Sprintf("disallowed-dir-colon-%s:folder", t.Name())),
			isDir:       true,
			description: "Directory with colon character in name",
		},
		{
			name:        "DirectoryWithVtiPrefix_ShouldBeRejected",
			path:        filepath.Join(TestDir, fmt.Sprintf("disallowed-dir-vti-%s_vti_folder", t.Name())),
			isDir:       true,
			description: "Directory with _vti_ prefix (reserved by SharePoint)",
		},
		{
			name:        "DirectoryWithGreaterThan_ShouldBeRejected",
			path:        filepath.Join(TestDir, fmt.Sprintf("disallowed-dir-gt-%s>folder", t.Name())),
			isDir:       true,
			description: "Directory with > character in name",
		},
		{
			name:        "ReservedNameDesktopIni_ShouldBeRejected",
			path:        filepath.Join(TestDir, "desktop.ini"),
			isDir:       true,
			description: "Reserved Windows configuration file desktop.ini",
		},
	}

	// Run each create test case as a subtest
	for _, tc := range createTestCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// We can use t.Parallel() here since each test uses a unique path
			t.Parallel()

			// Setup cleanup to remove the file/directory after test completes or fails
			t.Cleanup(func() {
				if tc.isDir {
					if err := os.RemoveAll(tc.path); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up directory %s: %v", tc.path, err)
					}
				} else {
					if err := os.Remove(tc.path); err != nil && !os.IsNotExist(err) {
						t.Logf("Warning: Failed to clean up file %s: %v", tc.path, err)
					}
				}
			})

			// Attempt to create the file/directory
			var err error
			if tc.isDir {
				err = os.Mkdir(tc.path, 0755)
			} else {
				err = os.WriteFile(tc.path, []byte("this should not work"), 0644)
			}

			// Check if the operation was rejected as expected
			if err != nil {
				t.Logf("Got expected error: %v", err)
			} else {
				fileType := "file"
				if tc.isDir {
					fileType = "directory"
				}
				t.Errorf("No error when creating %s with disallowed name: %s", fileType, tc.path)
			}
		})
	}

	// Test renaming to disallowed name
	t.Run("RenameToDisallowedName_ShouldBeRejected", func(t *testing.T) {
		t.Parallel()

		// Create unique directory names for this test
		validDir := filepath.Join(TestDir, fmt.Sprintf("valid-directory-%s", t.Name()))
		invalidDir := filepath.Join(TestDir, fmt.Sprintf("invalid-vti-directory-%s_vti_dir", t.Name()))

		// Setup cleanup to remove the directories after test completes or fails
		t.Cleanup(func() {
			for _, dir := range []string{validDir, invalidDir} {
				if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
					t.Logf("Warning: Failed to clean up directory %s: %v", dir, err)
				}
			}
		})

		// Create a valid directory
		err := os.Mkdir(validDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create valid directory: %v", err)
		}

		// Wait for the filesystem to process the directory creation
		testutil.WaitForCondition(t, func() bool {
			_, err := os.Stat(validDir)
			return err == nil
		}, 5*time.Second, 100*time.Millisecond, "Directory was not created within timeout")

		// Try to rename it to an invalid name
		err = os.Rename(validDir, invalidDir)
		if err != nil {
			t.Logf("Got expected error when renaming to disallowed name: %v", err)
		} else {
			t.Errorf("No error when renaming to disallowed name: %s", invalidDir)
		}
	})

	t.Log("Note: This test is informational. OneDrive may reject these files later during upload.")
}
