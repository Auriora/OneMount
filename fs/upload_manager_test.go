package fs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jstaf/onedriver/fs/graph"
	testutil "github.com/jstaf/onedriver/testutil/common"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

// Test that new uploads are written to disk to support resuming them later if
// the user shuts down their computer.
func TestUploadDiskSerialization(t *testing.T) {
	// Create a file path for our test file
	filePath := filepath.Join(TestDir, "upload_to_disk.fa")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
		}
	})

	// Copy the file synchronously to avoid race conditions
	cmd := exec.Command("cp", "dmel.fa", filePath)
	require.NoError(t, cmd.Run(), "Failed to copy test file")

	// Wait for the file to be recognized by the filesystem
	var inode *Inode
	testutil.WaitForCondition(t, func() bool {
		var err error
		inode, err = fs.GetPath("/onedriver_tests/upload_to_disk.fa", nil)
		return err == nil && inode != nil
	}, 10*time.Second, 500*time.Millisecond, "File was not recognized by filesystem")

	// Wait for the upload session to be created and serialized to disk
	var session UploadSession
	var found bool
	testutil.WaitForCondition(t, func() bool {
		session, found = findUploadSession(fs.db, inode.ID())
		return found
	}, 10*time.Second, 500*time.Millisecond, "Upload session was not created")

	// Now that we have a valid upload session, cancel it before it gets uploaded
	fs.uploads.CancelUpload(session.ID)

	// Confirm that the file didn't get uploaded yet
	// Use WaitForCondition with a short timeout to give any in-progress upload a chance to complete
	var driveItem *graph.DriveItem
	var err error
	testutil.WaitForCondition(t, func() bool {
		driveItem, err = graph.GetItemPath("/onedriver_tests/upload_to_disk.fa", auth)
		// If we can't find the item or it has no content, the test can proceed
		return err != nil || driveItem == nil || driveItem.Size == 0
	}, 5*time.Second, 500*time.Millisecond, "File was uploaded before the upload could be canceled")

	// Now create a new UploadManager from scratch with the file injected into its db
	dbPath := filepath.Join(testDBLoc, "test_upload_disk_serialization.db")

	// Setup cleanup to remove the database file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up database file %s: %v", dbPath, err)
		}
	})

	db, err := bolt.Open(dbPath, 0644, nil)
	require.NoError(t, err, "Failed to open database")

	// Setup cleanup to close the database after test completes or fails
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("Warning: Failed to close database: %v", err)
		}
	})

	// Create a bucket and store the upload session
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket(bucketUploads)
		if err != nil {
			return err
		}
		payload, err := json.Marshal(&session)
		if err != nil {
			return err
		}
		return b.Put([]byte(session.ID), payload)
	})
	require.NoError(t, err, "Failed to store upload session in database")

	// Create a new upload manager and wait for it to upload the file
	NewUploadManager(time.Second, db, fs, auth)

	// Wait for the file to be uploaded
	testutil.WaitForCondition(t, func() bool {
		driveItem, err = graph.GetItemPath("/onedriver_tests/upload_to_disk.fa", auth)
		return err == nil && driveItem != nil && driveItem.Size > 0
	}, 30*time.Second, 1*time.Second, "Could not find uploaded file after unserializing from disk and resuming upload")
}

// Helper function to find an upload session in the database
func findUploadSession(db *bolt.DB, inodeID string) (UploadSession, bool) {
	var session UploadSession
	var found bool

	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketUploads)
		if b == nil {
			return nil
		}

		diskSession := b.Get([]byte(inodeID))
		if diskSession == nil {
			return nil
		}

		if err := json.Unmarshal(diskSession, &session); err != nil {
			return err
		}

		found = true
		return nil
	})

	return session, found
}

// Make sure that uploading the same file multiple times works exactly as it should.
func TestRepeatedUploads(t *testing.T) {
	// test setup
	fname := filepath.Join(TestDir, "repeated_upload.txt")

	// Setup cleanup to remove the file after test completes or fails
	t.Cleanup(func() {
		if err := os.Remove(fname); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", fname, err)
		}
	})

	// Create initial file
	require.NoError(t, os.WriteFile(fname, []byte("initial content"), 0644), "Failed to write initial content")

	// Wait for the file to be recognized and uploaded
	var inode *Inode
	testutil.WaitForCondition(t, func() bool {
		var err error
		inode, err = fs.GetPath("/onedriver_tests/repeated_upload.txt", auth)
		if err != nil || inode == nil {
			return false
		}
		return !isLocalID(inode.ID())
	}, retrySeconds, 2*time.Second, "ID was local after upload")

	// Test multiple uploads of the same file
	for i := 0; i < 5; i++ {
		// Create new content for this iteration
		uploadme := []byte(fmt.Sprintf("iteration: %d", i))
		require.NoError(t, os.WriteFile(fname, uploadme, 0644), "Failed to write iteration content")

		// Wait for the file to be uploaded
		testutil.WaitForCondition(t, func() bool {
			// Get the item from the server
			item, err := graph.GetItemPath("/onedriver_tests/repeated_upload.txt", auth)
			if err != nil || item == nil {
				return false
			}

			// Get the content and verify it matches what we uploaded
			content, _, err := graph.GetItemContent(item.ID, auth)
			if err != nil {
				return false
			}

			return bytes.Equal(content, uploadme)
		}, 30*time.Second, 1*time.Second, fmt.Sprintf("Upload failed for iteration %d", i))
	}
}
