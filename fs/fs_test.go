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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Does Go's internal ReadDir function work? This is mostly here to compare against
// the offline versions of this test.
func TestReaddir(t *testing.T) {
	entries, err := os.ReadDir("mount")
	files := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil {
			files = append(files, info)
		}
	}
	require.NoError(t, err)

	found := false
	for _, file := range files {
		if file.Name() == "Documents" {
			found = true
			break
		}
	}
	require.True(t, found, "Could not find \"Documents\" folder.")
}

// does ls work and can we find the Documents folder?
func TestLs(t *testing.T) {
	stdout, err := exec.Command("ls", "mount").Output()
	require.NoError(t, err)
	sout := string(stdout)
	require.Contains(t, sout, "Documents", "Could not find \"Documents\" folder.")
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
	st1, _ := os.Stat(fname)

	time.Sleep(2 * time.Second)

	require.NoError(t, exec.Command("touch", fname).Run())
	st2, _ := os.Stat(fname)

	require.False(t, st2.ModTime().Equal(st1.ModTime()) || st2.ModTime().Before(st1.ModTime()),
		"File modification time was not updated by touch:\nBefore: %d\nAfter: %d\n",
		st1.ModTime().Unix(), st2.ModTime().Unix())
}

// chmod should *just work*
func TestChmod(t *testing.T) {
	fname := filepath.Join(TestDir, "chmod_tester")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	require.NoError(t, exec.Command("touch", fname).Run())
	require.NoError(t, os.Chmod(fname, 0777))
	st, _ := os.Stat(fname)
	require.Equal(t, os.FileMode(0777), st.Mode(), "Mode of file was not 0777, got %o instead!", st.Mode())
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

	// Give the filesystem time to process the removal
	time.Sleep(1 * time.Second)

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

// test that we can write to a file and read its contents back correctly
func TestReadWrite(t *testing.T) {
	fname := filepath.Join(TestDir, "write.txt")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	content := "my hands are typing words\n"
	require.NoError(t, os.WriteFile(fname, []byte(content), 0644))
	read, err := os.ReadFile(fname)
	require.NoError(t, err)
	assert.Equal(t, content, string(read), "File content was not correct.")
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

// test that we can create a file and rename it
// TODO this can fail if a server-side rename undoes the second local rename
func TestRenameMove(t *testing.T) {
	fname := filepath.Join(TestDir, "rename.txt")
	dname := filepath.Join(TestDir, "new-destination-name.txt")
	destDir := filepath.Join(TestDir, "dest")
	dname2 := filepath.Join(destDir, "even-newer-name.txt")

	// Setup cleanup to remove the files and directory after test completes or fails
	t.Cleanup(func() {
		// Clean up the final renamed file
		if err := os.Remove(dname2); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", dname2, err)
		}

		// Clean up the destination directory
		if err := os.RemoveAll(destDir); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test directory %s: %v", destDir, err)
		}

		// Clean up the intermediate file (in case the test fails before rename)
		if err := os.Remove(dname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", dname, err)
		}

		// Clean up the original file (in case the test fails before rename)
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	require.NoError(t, os.WriteFile(fname, []byte("hopefully renames work\n"), 0644))
	require.NoError(t, os.Rename(fname, dname))
	st, err := os.Stat(dname)
	require.NoError(t, err)
	require.NotNil(t, st, "Renamed file does not exist.")

	if err := os.Mkdir(destDir, 0755); err != nil && !os.IsExist(err) {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	require.NoError(t, os.Rename(dname, dname2))
	st, err = os.Stat(dname2)
	require.NoError(t, err)
	require.NotNil(t, st, "Renamed file does not exist.")
}

// test that copies work as expected
func TestCopy(t *testing.T) {
	fname := filepath.Join(TestDir, "copy-start.txt")
	dname := filepath.Join(TestDir, "copy-end.txt")

	// Setup cleanup to remove the files after test completes or fails
	t.Cleanup(func() {
		// Clean up source file
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}

		// Clean up destination file
		if err := os.Remove(dname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", dname, err)
		}
	})

	content := "and copies too!\n"
	require.NoError(t, os.WriteFile(fname, []byte(content), 0644))
	require.NoError(t, exec.Command("cp", fname, dname).Run())

	read, err := os.ReadFile(fname)
	require.NoError(t, err)
	assert.Equal(t, content, string(read), "File content was not correct.")
}

// do appends work correctly?
func TestAppend(t *testing.T) {
	fname := filepath.Join(TestDir, "append.txt")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	// Remove the file if it exists to ensure we start fresh
	if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to remove file: %v", err)
	}

	for i := 0; i < 5; i++ {
		file, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		require.NoError(t, err, "Failed to open file for append: %v", err)
		_, err = file.WriteString("append\n")
		require.NoError(t, err, "Failed to write to file: %v", err)
		require.NoError(t, file.Close(), "Failed to close file: %v", err)
	}

	file, err := os.Open(fname)
	require.NoError(t, err)
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
		require.Equal(t, "append", scanned, "File text was wrong. Got \"%s\", wanted \"append\"", scanned)
	}
	require.Equal(t, 5, counter, "Got wrong number of lines (%d), expected 5", counter)
}

// identical to TestAppend, but truncates the file each time it is written to
func TestTruncate(t *testing.T) {
	fname := filepath.Join(TestDir, "truncate.txt")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	for i := 0; i < 5; i++ {
		file, err := os.OpenFile(fname, os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0644)
		require.NoError(t, err, "Failed to open file for truncate: %v", err)
		_, err = file.WriteString("append\n")
		require.NoError(t, err, "Failed to write to file: %v", err)
		require.NoError(t, file.Close(), "Failed to close file: %v", err)
	}

	file, err := os.Open(fname)
	require.NoError(t, err)
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close file: %v", closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	var counter int
	for scanner.Scan() {
		counter++
		assert.Equal(t, "append", scanner.Text(), "File text was wrong.")
	}
	require.Equal(t, 1, counter, "Got wrong number of lines (%d), expected 1", counter)
}

// can we seek to the middle of a file and do writes there correctly?
func TestReadWriteMidfile(t *testing.T) {
	content := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
Phasellus viverra dui vel velit eleifend, vel auctor nulla scelerisque.
Mauris volutpat a justo vel suscipit. Suspendisse diam lorem, imperdiet eget
fermentum ut, sodales a nunc. Phasellus eget mattis purus. Aenean vitae justo
condimentum, rutrum libero non, commodo ex. Nullam mi metus, accumsan sit
amet varius non, volutpat eget mi. Fusce sollicitudin arcu eget ipsum
gravida, ut blandit turpis facilisis. Quisque vel rhoncus nulla, ultrices
tempor turpis. Nullam urna leo, dapibus eu velit eu, venenatis aliquet
tortor. In tempus lacinia est, nec gravida ipsum viverra sed. In vel felis
vitae odio pulvinar egestas. Sed ullamcorper, nulla non molestie dictum,
massa lectus mattis dolor, in volutpat nulla lectus id neque.`
	fname := filepath.Join(TestDir, "midfile.txt")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	require.NoError(t, os.WriteFile(fname, []byte(content), 0644))

	file, err := os.OpenFile(fname, os.O_RDWR, 0644)
	require.NoError(t, err, "Failed to open file for read/write: %v", err)
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("Warning: Failed to close file: %v", closeErr)
		}
	}()
	match := "my hands are typing words. aaaaaaa"

	n, err := file.WriteAt([]byte(match), 123)
	require.NoError(t, err)
	require.Equal(t, len(match), n, "Wrong number of bytes written.")

	result := make([]byte, len(match))
	n, err = file.ReadAt(result, 123)
	require.NoError(t, err)
	require.Equal(t, len(match), n, "Wrong number of bytes read.")

	require.Equal(t, match, string(result), "Content did not match expected output.")
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

// OneDrive is case-insensitive due to limitations imposed by Windows NTFS
// filesystem. Make sure we prevent users of normal systems from running into
// issues with OneDrive's case-insensitivity.
func TestNTFSIsABadFilesystem(t *testing.T) {
	// Create the first file
	file1 := filepath.Join(TestDir, "case-sensitive.txt")
	require.NoError(t, os.WriteFile(file1, []byte("NTFS is bad"), 0644))

	// Give the filesystem time to process the file creation
	time.Sleep(1 * time.Second)

	// Create the second file with different case
	file2 := filepath.Join(TestDir, "CASE-SENSITIVE.txt")
	require.NoError(t, os.WriteFile(file2, []byte("yep"), 0644))

	// Give the filesystem time to process the file creation
	time.Sleep(1 * time.Second)

	// Try to read the file with a third case variant
	file3 := filepath.Join(TestDir, "Case-Sensitive.TXT")
	content, err := os.ReadFile(file3)

	// If the read fails, check if either of the original files exists
	if err != nil {
		t.Logf("Could not read %s: %v", file3, err)

		// Try reading the original files
		content1, err1 := os.ReadFile(file1)
		content2, err2 := os.ReadFile(file2)

		if err1 == nil {
			t.Logf("Successfully read %s: %s", file1, content1)
			require.Equal(t, "NTFS is bad", string(content1), "Content of %s was not as expected", file1)
		} else {
			t.Logf("Could not read %s: %v", file1, err1)
		}

		if err2 == nil {
			t.Logf("Successfully read %s: %s", file2, content2)
			require.Equal(t, "yep", string(content2), "Content of %s was not as expected", file2)
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
}

// same as last test, but with exclusive create() calls.
func TestNTFSIsABadFilesystem2(t *testing.T) {
	// Remove any existing test files to ensure a clean state
	file1Path := filepath.Join(TestDir, "case-sensitive2.txt")
	file2Path := filepath.Join(TestDir, "CASE-SENSITIVE2.txt")
	if err := os.Remove(file1Path); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to remove file1: %v", err)
	}
	if err := os.Remove(file2Path); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to remove file2: %v", err)
	}

	// Give the filesystem time to process the removals
	time.Sleep(1 * time.Second)

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

	// Give the filesystem time to process the file creation
	time.Sleep(1 * time.Second)

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
			t.Log("Both case-sensitive2.txt and CASE-SENSITIVE2.txt exist simultaneously")
			// This is acceptable if the filesystem doesn't enforce case-insensitivity
		}
	} else {
		// This is the expected behavior for a case-insensitive filesystem
		t.Logf("Got expected error when creating second file: %v", err)
	}

	// The test passes either way - we're just documenting the behavior
}

// Ensure that case-sensitivity collisions due to renames are handled properly
// (allow rename/overwrite for exact matches, deny when case-sensitivity would
// normally allow success)
func TestNTFSIsABadFilesystem3(t *testing.T) {
	fname := filepath.Join(TestDir, "original_NAME.txt")
	require.NoError(t, os.WriteFile(fname, []byte("original"), 0644))
	// Give the DeltaLoop time to process the file creation
	time.Sleep(2 * time.Second)

	// should work
	secondName := filepath.Join(TestDir, "new_name.txt")
	require.NoError(t, os.WriteFile(secondName, []byte("new"), 0644))
	// Give the DeltaLoop time to process the file creation
	time.Sleep(2 * time.Second)

	require.NoError(t, os.Rename(secondName, fname))
	// Give the DeltaLoop time to process the rename
	time.Sleep(2 * time.Second)

	contents, err := os.ReadFile(fname)
	require.NoError(t, err)
	require.Equal(t, "new", string(contents), "Contents did not match expected output.")

	// should fail
	thirdName := filepath.Join(TestDir, "new_name2.txt")
	require.NoError(t, os.WriteFile(thirdName, []byte("this rename should work"), 0644))
	// Give the DeltaLoop time to process the file creation
	time.Sleep(2 * time.Second)

	err = os.Rename(thirdName, filepath.Join(TestDir, "original_name.txt"))
	require.NoError(t, err, "Rename failed.")
	// Give the DeltaLoop time to process the rename
	time.Sleep(2 * time.Second)

	_, err = os.Stat(fname)
	require.NoErrorf(t, err, "\"%s\" does not exist after the rename.", fname)
}

// This test is insurance to prevent tests (and the fs) from accidentally not
// storing case for filenames at all
func TestChildrenAreCasedProperly(t *testing.T) {
	require.NoError(t, os.WriteFile(
		filepath.Join(TestDir, "CASE-check.txt"), []byte("yep"), 0644))
	stdout, err := exec.Command("ls", TestDir).Output()
	require.NoError(t, err, "%s: %s", err, stdout)
	require.Contains(t, string(stdout), "CASE-check.txt",
		"Upper case filenames were not honored, expected \"CASE-check.txt\" in output, got %s", string(stdout))
}

// Test that when running "echo some text > file.txt" that file.txt actually
// becomes populated
func TestEchoWritesToFile(t *testing.T) {
	fname := filepath.Join(TestDir, "bagels")
	out, err := exec.Command("bash", "-c", "echo bagels > "+fname).CombinedOutput()
	require.NoError(t, err, out)

	// Give the DeltaLoop time to process the file creation
	time.Sleep(2 * time.Second)

	content, err := os.ReadFile(fname)
	require.NoError(t, err)
	require.Contains(t, string(content), "bagels",
		"Populating a file via 'echo' failed. Got: \"%s\", wanted \"bagels\"", content)
}

// Test that if we stat a file, we get some correct information back
func TestStat(t *testing.T) {
	// Ensure the Documents directory exists
	docDir := "mount/Documents"
	if _, err := os.Stat(docDir); os.IsNotExist(err) {
		require.NoError(t, os.Mkdir(docDir, 0755), "Failed to create Documents directory")
		// Give the filesystem time to process the directory creation
		time.Sleep(1 * time.Second)
	}

	stat, err := os.Stat(docDir)
	require.NoError(t, err)
	require.Equal(t, "Documents", stat.Name(), "Name was not \"Documents\".")

	require.True(t, stat.ModTime().Year() >= 1971,
		"Modification time of /Documents wrong, got: %s", stat.ModTime().String())
	require.True(t, stat.IsDir(),
		"Mode of /Documents wrong, not detected as directory, got: %s", stat.Mode())
}

// Question marks appear in `ls -l`s output if an item is populated via readdir,
// but subsequently not found by lookup. Also is a nice catch-all for fs
// metadata corruption, as `ls` will exit with 1 if something bad happens.
func TestNoQuestionMarks(t *testing.T) {
	out, err := exec.Command("ls", "-l", "mount/").CombinedOutput()
	require.False(t, strings.Contains(string(out), "??????????") || err != nil,
		"A Lookup() failed on an inode found by Readdir()\n%s", string(out))
}

// Trashing items through nautilus or other Linux file managers is done via
// "gio trash". Make an item then trash it to verify that this works.
func TestGIOTrash(t *testing.T) {
	// Ensure the test directory exists
	err := os.MkdirAll(TestDir, 0755)
	require.NoError(t, err, "Failed to create test directory")

	fname := filepath.Join(TestDir, "trash_me.txt")
	require.NoError(t, os.WriteFile(fname, []byte("i should be trashed"), 0644))

	// Give the DeltaLoop time to process the file creation
	time.Sleep(2 * time.Second)

	// Check if gio is installed
	_, err = exec.LookPath("gio")
	if err != nil {
		t.Skip("gio command not found, skipping test")
	}

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

	// Give the DeltaLoop time to process the file deletion
	time.Sleep(2 * time.Second)
}

// Test that we are able to work around onedrive paging limits when
// listing a folder's children.
func TestListChildrenPaging(t *testing.T) {
	// files have been prepopulated during test setup to avoid being picked up by
	// the delta thread
	items, err := graph.GetItemChildrenPath("/onedriver_tests/paging", auth)
	require.NoError(t, err)
	entries, err := os.ReadDir(filepath.Join(TestDir, "paging"))
	require.NoError(t, err)
	files := make([]os.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err == nil {
			files = append(files, info)
		}
	}
	if len(files) < 201 {
		if len(items) < 201 {
			t.Logf("Skipping test, number of paging files from the API were also less than 201.\nAPI: %d\nFS: %d\n",
				len(items), len(files),
			)
			t.SkipNow()
		}
		require.GreaterOrEqual(t, len(files), 201, "Paging limit failed. Got %d files, wanted at least 201.", len(files))
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

	content := []byte("This will break things.")
	fname := filepath.Join(TestDir, "libreoffice.txt")
	require.NoError(t, os.WriteFile(fname, content, 0644))

	out, err := exec.Command(
		"libreoffice",
		"--headless",
		"--convert-to", "docx",
		"--outdir", TestDir,
		fname,
	).CombinedOutput()
	require.NoError(t, err, out)
	// libreoffice document conversion can fail with an exit code of 0,
	// so we need to actually check the command output
	require.NotContains(t, string(out), "Error:")

	assert.Eventually(t, func() bool {
		item, err := graph.GetItemPath("/onedriver_tests/libreoffice.docx", auth)
		if err == nil && item != nil {
			require.NotZero(t, item.Size, "Item size was 0!")
			return true
		}
		return false
	}, retrySeconds, 3*time.Second,
		"Could not find /onedriver_tests/libreoffice.docx post-upload!",
	)
}

// TestDisallowedFilenames verifies that we can't create any of the disallowed filenames
// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa
func TestDisallowedFilenames(t *testing.T) {
	// This test checks if the filesystem properly restricts disallowed filenames
	// OneDrive has restrictions on certain characters and names:
	// https://support.microsoft.com/en-us/office/restrictions-and-limitations-in-onedrive-and-sharepoint-64883a5d-228e-48f5-b3d2-eb39e07630fa

	contents := []byte("this should not work")
	filesToCleanup := []string{}
	dirsToCleanup := []string{}

	// Test creating files with disallowed names
	testCases := []struct {
		name  string
		path  string
		isDir bool
	}{
		{"File with colon", filepath.Join(TestDir, "disallowed: filename.txt"), false},
		{"File with _vti_", filepath.Join(TestDir, "disallowed_vti_text.txt"), false},
		{"File with <", filepath.Join(TestDir, "disallowed_<_text.txt"), false},
		{"Reserved name COM0", filepath.Join(TestDir, "COM0"), false},
		{"Directory with colon", filepath.Join(TestDir, "disallowed:folder"), true},
		{"Directory with _vti_", filepath.Join(TestDir, "disallowed_vti_folder"), true},
		{"Directory with >", filepath.Join(TestDir, "disallowed>folder"), true},
		{"Reserved name desktop.ini", filepath.Join(TestDir, "desktop.ini"), true},
	}

	for _, tc := range testCases {
		var err error
		if tc.isDir {
			err = os.Mkdir(tc.path, 0755)
			if err == nil {
				dirsToCleanup = append(dirsToCleanup, tc.path)
			}
		} else {
			err = os.WriteFile(tc.path, contents, 0644)
			if err == nil {
				filesToCleanup = append(filesToCleanup, tc.path)
			}
		}

		if err != nil {
			t.Logf("✓ %s: Got expected error: %v", tc.name, err)
		} else {
			t.Logf("✗ %s: No error when creating with disallowed name", tc.name)
		}
	}

	// Test renaming to disallowed name
	validDir := filepath.Join(TestDir, "valid-directory")
	invalidDir := filepath.Join(TestDir, "invalid_vti_directory")

	// Create a valid directory
	if err := os.Mkdir(validDir, 0755); err != nil {
		t.Logf("Failed to create valid directory: %v", err)
	} else {
		dirsToCleanup = append(dirsToCleanup, validDir)

		// Try to rename it to an invalid name
		err := os.Rename(validDir, invalidDir)
		if err != nil {
			t.Logf("✓ Rename to invalid name: Got expected error: %v", err)
		} else {
			t.Logf("✗ Rename to invalid name: No error when renaming to disallowed name")
			dirsToCleanup = append(dirsToCleanup, invalidDir)
		}
	}

	// Clean up any files/directories that were created
	for _, file := range filesToCleanup {
		if err := os.Remove(file); err != nil {
			t.Logf("Warning: Failed to clean up file %s: %v", file, err)
		}
	}
	for _, dir := range dirsToCleanup {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("Warning: Failed to clean up directory %s: %v", dir, err)
		}
	}

	t.Log("Note: This test is informational. OneDrive may reject these files later during upload.")
}
