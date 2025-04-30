// Run tests to verify that we are syncing changes from the server.
package fs

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bcherrington/onemount/internal/fs/graph"
	"github.com/bcherrington/onemount/internal/testutil/common"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// a helper function for use with tests
func (i *Inode) setContent(f *Filesystem, newContent []byte) error {
	i.DriveItem.Size = uint64(len(newContent))
	now := time.Now()
	i.DriveItem.ModTime = &now

	err := f.content.Insert(i.ID(), newContent)
	if err != nil {
		return err
	}

	if i.DriveItem.File == nil {
		i.DriveItem.File = &graph.File{}
	}

	i.DriveItem.File.Hashes.QuickXorHash = graph.QuickXORHash(&newContent)
	return nil
}

// TestDeltaOperations tests various delta operations using a table-driven approach
func TestDeltaOperations(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name            string
		setup           func(t *testing.T) (string, *graph.DriveItem, error)
		operation       func(t *testing.T, item *graph.DriveItem) error
		expectedPath    string
		verifyContent   bool
		expectedContent []byte
		verifyExists    bool
	}{
		{
			name: "CreateDirectoryOnServer_ShouldSyncToClient",
			setup: func(t *testing.T) (string, *graph.DriveItem, error) {
				parent, err := graph.GetItemPath("/onemount_tests/delta", auth)
				if err != nil {
					return "", nil, err
				}
				return filepath.Join(DeltaDir, "first"), parent, nil
			},
			operation: func(t *testing.T, item *graph.DriveItem) error {
				_, err := graph.Mkdir("first", item.ID, auth)
				return err
			},
			expectedPath: filepath.Join(DeltaDir, "first"),
			verifyExists: true,
		},
		{
			name: "DeleteDirectoryOnServer_ShouldSyncToClient",
			setup: func(t *testing.T) (string, *graph.DriveItem, error) {
				fname := filepath.Join(DeltaDir, "delete_me")
				err := os.Mkdir(fname, 0755)
				if err != nil {
					return "", nil, err
				}

				// Wait for the directory to be recognized by the server
				var item *graph.DriveItem
				common.WaitForCondition(t, func() bool {
					item, err = graph.GetItemPath("/onemount_tests/delta/delete_me", auth)
					return err == nil && item != nil
				}, 10*time.Second, time.Second, "Directory was not recognized by server")

				return fname, item, nil
			},
			operation: func(t *testing.T, item *graph.DriveItem) error {
				return graph.Remove(item.ID, auth)
			},
			expectedPath: filepath.Join(DeltaDir, "delete_me"),
			verifyExists: false,
		},
		{
			name: "RenameFileOnServer_ShouldSyncToClient",
			setup: func(t *testing.T) (string, *graph.DriveItem, error) {
				filePath := filepath.Join(DeltaDir, "delta_rename_start")
				err := os.WriteFile(filePath, []byte("cheesecake"), 0644)
				if err != nil {
					return "", nil, err
				}

				// Wait for the file to be recognized by the server
				var item *graph.DriveItem
				common.WaitForCondition(t, func() bool {
					item, err = graph.GetItemPath("/onemount_tests/delta/delta_rename_start", auth)
					return err == nil && item != nil
				}, 10*time.Second, time.Second, "File was not recognized by server")

				return filepath.Join(DeltaDir, "delta_rename_end"), item, nil
			},
			operation: func(t *testing.T, item *graph.DriveItem) error {
				return graph.Rename(item.ID, "delta_rename_end", item.Parent.ID, auth)
			},
			expectedPath:    filepath.Join(DeltaDir, "delta_rename_end"),
			verifyContent:   true,
			expectedContent: []byte("cheesecake"),
			verifyExists:    true,
		},
		{
			name: "MoveFileToNewParentOnServer_ShouldSyncToClient",
			setup: func(t *testing.T) (string, *graph.DriveItem, error) {
				filePath := filepath.Join(DeltaDir, "delta_move_start")
				err := os.WriteFile(filePath, []byte("carrotcake"), 0644)
				if err != nil {
					return "", nil, err
				}

				// Wait for the file to be recognized by the server
				var item *graph.DriveItem
				common.WaitForCondition(t, func() bool {
					item, err = graph.GetItemPath("/onemount_tests/delta/delta_move_start", auth)
					return err == nil && item != nil
				}, 10*time.Second, time.Second, "File was not recognized by server")

				return filepath.Join(TestDir, "delta_rename_end"), item, nil
			},
			operation: func(t *testing.T, item *graph.DriveItem) error {
				newParent, err := graph.GetItemPath("/onemount_tests/", auth)
				if err != nil {
					return err
				}
				return graph.Rename(item.ID, "delta_rename_end", newParent.ID, auth)
			},
			expectedPath:    filepath.Join(TestDir, "delta_rename_end"),
			verifyContent:   true,
			expectedContent: []byte("carrotcake"),
			verifyExists:    true,
		},
	}

	// Run each test case
	for _, tc := range testCases {
		tc := tc // Capture range variable for parallel execution
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			expectedPath, item, err := tc.setup(t)
			require.NoError(t, err, "Setup failed")
			require.NotNil(t, item, "Setup did not return a valid item")

			// Perform the operation
			err = tc.operation(t, item)
			require.NoError(t, err, "Operation failed")

			// Verify the result
			if tc.verifyExists {
				// Wait for the item to exist and verify its content if needed
				common.WaitForCondition(t, func() bool {
					_, err := os.Stat(expectedPath)
					if err != nil {
						return false
					}

					if tc.verifyContent {
						content, err := os.ReadFile(expectedPath)
						if err != nil {
							return false
						}
						return bytes.Contains(content, tc.expectedContent)
					}

					// For directories, verify it's a directory
					if !tc.verifyContent {
						info, err := os.Stat(expectedPath)
						if err != nil {
							return false
						}
						return info.IsDir()
					}

					return true
				}, retry, time.Second, "Expected item was not found or had incorrect content")
			} else {
				// Wait for the item to not exist
				common.WaitForCondition(t, func() bool {
					_, err := os.Stat(expectedPath)
					return os.IsNotExist(err)
				}, retry, time.Second, "Item still exists but should have been deleted")
			}
		})
	}
}

// Change the content remotely on the server, and verify it gets propagated to
// to the client.
func TestDeltaContentChangeRemote(t *testing.T) {
	require.NoError(t, os.WriteFile(
		filepath.Join(DeltaDir, "remote_content"),
		[]byte("the cake is a lie"),
		0644,
	))

	// change and upload it via the API
	var item *graph.DriveItem
	var err error
	require.Eventually(t, func() bool {
		item, err = graph.GetItemPath("/onemount_tests/delta/remote_content", auth)
		return err == nil && item != nil
	}, 30*time.Second, time.Second, "Could not find remote_content file")
	inode := NewInodeDriveItem(item)
	newContent := []byte("because it has been changed remotely!")
	err = inode.setContent(fs, newContent)
	require.NoError(t, err)
	data := fs.content.Get(inode.ID())
	session, err := NewUploadSession(inode, &data)
	require.NoError(t, err)
	require.NoError(t, session.Upload(auth))

	var body []byte
	var getErr error
	require.Eventually(t, func() bool {
		body, _, getErr = graph.GetItemContent(inode.ID(), auth)
		return getErr == nil && bytes.Equal(body, newContent)
	}, 30*time.Second, time.Second, "Failed to upload test file or content mismatch")

	// Wait for the DeltaLoop to detect the change and update the local file
	// The DeltaLoop polls every 5 seconds, so we need to wait long enough for it to detect changes
	var content []byte
	assert.Eventuallyf(t, func() bool {
		content, err = os.ReadFile(filepath.Join(DeltaDir, "remote_content"))
		require.NoError(t, err)
		return bytes.Equal(content, newContent)
	}, retry, time.Second,
		"Failed to sync content to local machine. Got content: \"%s\". "+
			"Wanted: \"because it has been changed remotely!\". "+
			"Remote content: \"%s\".",
		string(content), string(body),
	)
}

// Change the content both on the server and the client and verify that the
// client data is preserved.
func TestDeltaContentChangeBoth(t *testing.T) {

	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_delta_content_change_both"), 30)
	require.NoError(t, err)
	inode := NewInode("both_content_changed.txt", 0644|fuse.S_IFREG, nil)
	_, err = cache.InsertPath("/both_content_changed.txt", nil, inode)
	require.NoError(t, err)
	original := []byte("initial content")
	err = inode.setContent(cache, original)
	require.NoError(t, err)

	// write to, but do not close the file to simulate an in-use local file
	local := []byte("local write content")
	_, status := cache.Write(
		context.Background().Done(),
		&fuse.WriteIn{
			InHeader: fuse.InHeader{NodeId: inode.NodeID()},
			Offset:   0,
			Size:     uint32(len(local)),
		},
		local,
	)
	require.Equal(t, fuse.OK, status, "Write failed")

	// apply a fake delta to the local item
	fakeDelta := inode.DriveItem
	now := time.Now().Add(time.Second * 10)
	fakeDelta.ModTime = &now
	fakeDelta.Size = uint64(len(original))
	fakeDelta.ETag = "sldfjlsdjflkdj"
	fakeDelta.File.Hashes = graph.Hashes{
		QuickXorHash: graph.QuickXORHash(&original),
	}

	// should do nothing
	require.NoError(t, cache.applyDelta(&fakeDelta))
	require.Equal(t, uint64(len(local)), inode.Size(), "Contents of open local file changed!")

	// act as if the file is now flushed (these are the ops that would happen during
	// a flush)
	inode.DriveItem.File = &graph.File{}
	fd, err := fs.content.Open(inode.ID())
	require.NoError(t, err)
	inode.DriveItem.File.Hashes.QuickXorHash = graph.QuickXORHashStream(fd)
	err = cache.content.Close(inode.DriveItem.ID)
	require.NoError(t, err)
	inode.hasChanges = false

	// should now change the file
	require.NoError(t, cache.applyDelta(&fakeDelta))
	require.Equal(t, fakeDelta.Size, inode.Size(),
		"Contents of local file was not changed after disabling local changes!")
}

// If we have local content in the local disk cache that doesn't match what the
// server has, Open() should pick this up and wipe it. Otherwise Open() could
// pick up an old version of a file from previous program startups and think
// it's current, which would erase the real, up-to-date server copy.
func TestDeltaBadContentInCache(t *testing.T) {
	// write a file to the server and poll until it exists
	require.NoError(t, os.WriteFile(
		filepath.Join(DeltaDir, "corrupted"),
		[]byte("correct contents"),
		0644,
	))
	var id string
	require.Eventually(t, func() bool {
		item, err := graph.GetItemPath("/onemount_tests/delta/corrupted", auth)
		if err == nil {
			id = item.ID
			return true
		}
		return false
	}, retry, time.Second)

	insertErr := fs.content.Insert(id, []byte("wrong contents"))
	require.NoError(t, insertErr)

	// Use Eventually to wait for the file to be redownloaded with correct contents
	// This gives the system time to detect the corrupted cache and download the correct content
	var contents []byte
	var err error
	require.Eventually(t, func() bool {
		contents, err = os.ReadFile(filepath.Join(DeltaDir, "corrupted"))
		return err == nil && !bytes.HasPrefix(contents, []byte("wrong"))
	}, retry, time.Second, "File contents were wrong! Got \"%s\", wanted \"correct contents\"",
		func() string {
			if err != nil {
				return err.Error()
			}
			return string(contents)
		}())
}

// Check that folders are deleted only when empty after syncing the complete set of
// changes.
func TestDeltaFolderDeletion(t *testing.T) {
	require.NoError(t, os.MkdirAll(filepath.Join(DeltaDir, "nested/directory"), 0755))
	nested, err := graph.GetItemPath("/onemount_tests/delta/nested", auth)
	require.NoError(t, err)
	require.NoError(t, graph.Remove(nested.ID, auth))

	// now poll and wait for deletion
	assert.Eventually(t, func() bool {
		entries, _ := os.ReadDir(DeltaDir)
		for _, entry := range entries {
			if entry.Name() == "nested" {
				return false
			}
		}
		return true
	}, retry, time.Second, "\"nested/\" directory was not deleted.")
}

// We should only perform a delta deletion of a folder if it was nonempty
func TestDeltaFolderDeletionNonEmpty(t *testing.T) {
	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_delta_folder_deletion_nonempty"), 30)
	require.NoError(t, err)
	dir := NewInode("folder", 0755|fuse.S_IFDIR, nil)
	file := NewInode("file", 0644|fuse.S_IFREG, nil)
	_, err = cache.InsertPath("/folder", nil, dir)
	require.NoError(t, err)
	_, err = cache.InsertPath("/folder/file", nil, file)
	require.NoError(t, err)

	delta := &graph.DriveItem{
		ID:      dir.ID(),
		Parent:  &graph.DriveItemParent{ID: dir.ParentID()},
		Deleted: &graph.Deleted{State: "softdeleted"},
		Folder:  &graph.Folder{},
	}
	deltaErr := cache.applyDelta(delta)
	require.NotNil(t, cache.GetID(delta.ID), "Folder should still be present")
	require.Error(t, deltaErr, "A delta deletion of a non-empty folder was not an error")

	cache.DeletePath("/folder/file")
	err = cache.applyDelta(delta)
	if err != nil {
		t.Fatalf("Failed to apply delta after emptying folder: %v", err)
	}
	assert.Nil(t, cache.GetID(delta.ID),
		"Still found folder after emptying it first (the correct way).")
}

// Some programs like LibreOffice and WPS Office will have a fit if the
// modification times on their lockfiles is updated after they are written. This
// test verifies that the delta thread does not modify modification times if the
// content is unchanged.
func TestDeltaNoModTimeUpdate(t *testing.T) {
	fname := filepath.Join(DeltaDir, "mod_time_update.txt")
	require.NoError(t, os.WriteFile(fname, []byte("a pretend lockfile"), 0644))
	finfo, err := os.Stat(fname)
	require.NoError(t, err)
	mtimeOriginal := finfo.ModTime()

	// Wait for enough time to ensure the DeltaLoop has run multiple times
	// The DeltaLoop polls every 5 seconds, so we'll wait for 15 seconds
	// While waiting, periodically check that the modification time hasn't changed
	var mtimeNew time.Time
	common.WaitForCondition(t, func() bool {
		currentInfo, err := os.Stat(fname)
		if err != nil {
			t.Logf("Error stating file: %v", err)
			return false
		}

		// Store the current modification time for later comparison
		mtimeNew = currentInfo.ModTime()

		// Check if enough time has passed (at least 15 seconds)
		return time.Since(mtimeOriginal) >= 15*time.Second
	}, 20*time.Second, 500*time.Millisecond, "Failed to wait long enough for DeltaLoop to run multiple times")
	require.True(t, mtimeNew.Equal(mtimeOriginal),
		"Modification time was updated even though the file did not change.\n"+
			"Old mtime: %d, New mtime: %d\n", mtimeOriginal.Unix(), mtimeNew.Unix())
}

// deltas can come back missing from the server
// https://github.com/bcherrington/onemount/issues/111
func TestDeltaMissingHash(t *testing.T) {
	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_delta_missing_hash"), 30)
	require.NoError(t, err)
	file := NewInode("file", 0644|fuse.S_IFREG, nil)
	_, err = cache.InsertPath("/folder", nil, file)
	require.NoError(t, err)

	// Wait for the filesystem to process the insertion
	common.WaitForCondition(t, func() bool {
		// Check if the file exists in the filesystem
		return cache.GetID(file.ID()) != nil
	}, 5*time.Second, 100*time.Millisecond, "File was not inserted into filesystem within timeout")
	now := time.Now()
	delta := &graph.DriveItem{
		ID:      file.ID(),
		Parent:  &graph.DriveItemParent{ID: file.ParentID()},
		ModTime: &now,
		Size:    12345,
	}
	err = cache.applyDelta(delta)
	require.NoError(t, err)
	// if we survive to here without a segfault, test passed
}
