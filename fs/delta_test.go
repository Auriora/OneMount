// Run tests to verify that we are syncing changes from the server.
package fs

import (
	"bytes"
	"context"
	"fmt"
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

// In this test, we create a directory through the API, and wait to see if
// the cache picks it up post-creation.
func TestDeltaMkdir(t *testing.T) {
	parent, err := graph.GetItemPath("/onedriver_tests/delta", auth)
	require.NoError(t, err)

	// create the directory directly through the API and bypass the cache
	_, err = graph.Mkdir("first", parent.ID, auth)
	require.NoError(t, err)
	fname := filepath.Join(DeltaDir, "first")

	// give the delta thread time to fetch the item
	assert.Eventuallyf(t, func() bool {
		st, err := os.Stat(fname)
		if err == nil {
			if st.Mode().IsDir() {
				return true
			}
			require.Fail(t, fmt.Sprintf("%s was not a directory", fname))
		}
		return false
	}, retrySeconds, time.Second, "%s not found", fname)
}

// We create a directory through the cache, then delete through the API and see
// if the cache picks it up.
func TestDeltaRmdir(t *testing.T) {
	fname := filepath.Join(DeltaDir, "delete_me")
	require.NoError(t, os.Mkdir(fname, 0755))

	item, err := graph.GetItemPath("/onedriver_tests/delta/delete_me", auth)
	require.NoError(t, err)
	require.NoError(t, graph.Remove(item.ID, auth))

	// wait for delta sync
	assert.Eventually(t, func() bool {
		_, err := os.Stat(fname)
		return os.IsNotExist(err)
	}, retrySeconds, time.Second, "File deletion not picked up by client")
}

// Create a file locally, then rename it remotely and verify that the renamed
// file still has the correct content under the new parent.
func TestDeltaRename(t *testing.T) {
	require.NoError(t, os.WriteFile(
		filepath.Join(DeltaDir, "delta_rename_start"),
		[]byte("cheesecake"),
		0644,
	))

	var item *graph.DriveItem
	var err error
	require.Eventually(t, func() bool {
		item, err = graph.GetItemPath("/onedriver_tests/delta/delta_rename_start", auth)
		return err == nil
	}, 10*time.Second, time.Second, "Could not prepare /onedriver_test/delta/delta_rename_start")
	inode := NewInodeDriveItem(item)

	require.NoError(t, graph.Rename(inode.ID(), "delta_rename_end", inode.ParentID(), auth))
	fpath := filepath.Join(DeltaDir, "delta_rename_end")
	assert.Eventually(t, func() bool {
		if _, err := os.Stat(fpath); err == nil {
			content, err := os.ReadFile(fpath)
			require.NoError(t, err)
			return bytes.Contains(content, []byte("cheesecake"))
		}
		return false
	}, retrySeconds, time.Second, "Rename not detected by client.")
}

// Create a file locally, then move it on the server to a new directory. Check
// to see if the cache picks it up.
func TestDeltaMoveParent(t *testing.T) {
	filePath := filepath.Join(DeltaDir, "delta_move_start")
	require.NoError(t, os.WriteFile(
		filePath,
		[]byte("carrotcake"),
		0644,
	))

	// Wait for the file to be recognized by the filesystem
	testutil.WaitForCondition(t, func() bool {
		_, err := os.Stat(filePath)
		return err == nil
	}, 5*time.Second, 100*time.Millisecond, "File was not created within timeout")

	var item *graph.DriveItem
	var err error
	require.Eventually(t, func() bool {
		item, err = graph.GetItemPath("/onedriver_tests/delta/delta_move_start", auth)
		return err == nil
	}, 10*time.Second, time.Second)

	newParent, err := graph.GetItemPath("/onedriver_tests/", auth)
	require.NoError(t, err)

	require.NoError(t, graph.Rename(item.ID, "delta_rename_end", newParent.ID, auth))
	fpath := filepath.Join(TestDir, "delta_rename_end")
	assert.Eventually(t, func() bool {
		if _, err := os.Stat(fpath); err == nil {
			content, err := os.ReadFile(fpath)
			require.NoError(t, err)
			return bytes.Contains(content, []byte("carrotcake"))
		}
		return false
	}, retrySeconds, time.Second, "Rename not detected by client")
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
		item, err = graph.GetItemPath("/onedriver_tests/delta/remote_content", auth)
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
	}, retrySeconds, time.Second,
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
		item, err := graph.GetItemPath("/onedriver_tests/delta/corrupted", auth)
		if err == nil {
			id = item.ID
			return true
		}
		return false
	}, retrySeconds, time.Second)

	insertErr := fs.content.Insert(id, []byte("wrong contents"))
	require.NoError(t, insertErr)

	// Use Eventually to wait for the file to be redownloaded with correct contents
	// This gives the system time to detect the corrupted cache and download the correct content
	var contents []byte
	var err error
	require.Eventually(t, func() bool {
		contents, err = os.ReadFile(filepath.Join(DeltaDir, "corrupted"))
		return err == nil && !bytes.HasPrefix(contents, []byte("wrong"))
	}, retrySeconds, time.Second, "File contents were wrong! Got \"%s\", wanted \"correct contents\"",
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
	nested, err := graph.GetItemPath("/onedriver_tests/delta/nested", auth)
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
	}, retrySeconds, time.Second, "\"nested/\" directory was not deleted.")
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
	cache.applyDelta(delta)
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
	testutil.WaitForCondition(t, func() bool {
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
// https://github.com/jstaf/onedriver/issues/111
func TestDeltaMissingHash(t *testing.T) {
	cache, err := NewFilesystem(auth, filepath.Join(testDBLoc, "test_delta_missing_hash"), 30)
	require.NoError(t, err)
	file := NewInode("file", 0644|fuse.S_IFREG, nil)
	_, err = cache.InsertPath("/folder", nil, file)
	require.NoError(t, err)

	// Wait for the filesystem to process the insertion
	testutil.WaitForCondition(t, func() bool {
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
