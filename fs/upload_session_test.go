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

// TestUploadSession verifies that the basic functionality of uploads works correctly.
func TestUploadSession(t *testing.T) {
	testDir, err := fs.GetPath("/onedriver_tests", auth)
	require.NoError(t, err)

	inode := NewInode("uploadSessionSmall.txt", 0644, testDir)
	data := []byte("our super special data")
	inode.setContent(fs, data)
	mtime := inode.ModTime()

	session, err := NewUploadSession(inode, &data)
	require.NoError(t, err)
	err = session.Upload(auth)
	require.NoError(t, err)
	require.False(t, isLocalID(session.ID),
		"The session's ID was somehow still local following an upload: %s",
		session.ID)
	sessionMtime := uint64(session.ModTime.Unix())
	assert.Equal(t, mtime, sessionMtime, "session modtime changed - before: %d - after: %d", mtime, sessionMtime)

	resp, _, err := graph.GetItemContent(session.ID, auth)
	require.NoError(t, err)
	require.True(t, bytes.Equal(data, resp),
		"Data mismatch. Original content: %s\nRemote content: %s", data, resp)

	// item now has a new id following the upload. We just change the ID here
	// because thats part of the UploadManager functionality and gets tested elsewhere.
	inode.DriveItem.ID = session.ID

	// we overwrite and upload again to test uploading with the new remote id
	newData := []byte("new data is extra long so it covers the old one completely")
	inode.setContent(fs, newData)

	session2, err := NewUploadSession(inode, &newData)
	require.NoError(t, err)
	err = session2.Upload(auth)
	require.NoError(t, err)

	resp, _, err = graph.GetItemContent(session.ID, auth)
	require.NoError(t, err)
	require.True(t, bytes.Equal(newData, resp),
		"Data mismatch. Original content: %s\nRemote content: %s", newData, resp)
}

// TestUploadSessionSmallFS verifies is the same test as TestUploadSessionSmall, but uses
// the filesystem itself to perform the uploads instead of testing the internal upload
// functions directly
func TestUploadSessionSmallFS(t *testing.T) {
	filePath := filepath.Join(TestDir, "uploadSessionSmallFS.txt")
	data := []byte("super special data for upload test 2")
	err := os.WriteFile(filePath, data, 0644)
	require.NoError(t, err)

	// Wait for the file to be uploaded and available on the server
	testutil.WaitForCondition(t, func() bool {
		item, err := graph.GetItemPath("/onedriver_tests/uploadSessionSmallFS.txt", auth)
		return err == nil && item != nil
	}, 30*time.Second, time.Second, "File was not uploaded to server within timeout")

	// Now get the item for content verification
	item, err := graph.GetItemPath("/onedriver_tests/uploadSessionSmallFS.txt", auth)
	require.NoError(t, err)
	require.NotNil(t, item, "Item not found")

	content, _, err := graph.GetItemContent(item.ID, auth)
	require.NoError(t, err)
	require.True(t, bytes.Equal(content, data),
		"Data mismatch. Original content: %s\nRemote content: %s", data, content)

	// upload it again to ensure uploads with an existing remote id succeed
	data = []byte("more super special data")
	err = os.WriteFile(filePath, data, 0644)
	require.NoError(t, err)

	// Wait for the file to be uploaded again and available on the server with updated content
	testutil.WaitForCondition(t, func() bool {
		updatedItem, err := graph.GetItemPath("/onedriver_tests/uploadSessionSmallFS.txt", auth)
		if err != nil || updatedItem == nil {
			return false
		}

		// Check if the content has been updated
		content, _, err := graph.GetItemContent(updatedItem.ID, auth)
		return err == nil && bytes.Equal(content, data)
	}, 30*time.Second, time.Second, "File was not re-uploaded to server with updated content within timeout")

	// Now get the item for final verification
	item2, err := graph.GetItemPath("/onedriver_tests/uploadSessionSmallFS.txt", auth)
	require.NoError(t, err)
	require.NotNil(t, item2, "Item not found")

	content, _, err = graph.GetItemContent(item2.ID, auth)
	require.NoError(t, err)
	require.True(t, bytes.Equal(content, data),
		"Data mismatch. Original content: %s\nRemote content: %s", data, content)
}

// copy large file inside onedrive mount, then verify that we can still
// access selected lines
func TestUploadSessionLargeFS(t *testing.T) {
	// Check if the source file exists
	_, err := os.Stat("dmel.fa")
	if err != nil {
		t.Skip("dmel.fa file not found, skipping test")
	}

	// Use os.ReadFile and os.WriteFile instead of exec.Command("cp")
	sourceData, err := os.ReadFile("dmel.fa")
	require.NoError(t, err, "Failed to read source file")

	fname := filepath.Join(TestDir, "dmel.fa")
	err = os.WriteFile(fname, sourceData, 0644)
	require.NoError(t, err, "Failed to write to destination file")

	contents, err := os.ReadFile(fname)
	require.NoError(t, err)

	header := ">X dna:chromosome chromosome:BDGP6.22:X:1:23542271:1 REF"
	require.Equal(t, header, string(contents[:len(header)]),
		"Could not read FASTA header. Wanted \"%s\", got \"%s\"",
		header, string(contents[:len(header)]))

	final := "AAATAAAATAC\n" // makes yucky test output, but is the final line
	match := string(contents[len(contents)-len(final):])
	require.Equal(t, final, match,
		"Could not read final line of FASTA. Wanted \"%s\", got \"%s\"",
		final, match)

	st, _ := os.Stat(fname)
	require.NotZero(t, st.Size(), "File size cannot be 0.")

	// poll endpoint to make sure it has a size greater than 0
	size := uint64(len(contents))
	var item *graph.DriveItem
	assert.Eventually(t, func() bool {
		item, _ = graph.GetItemPath("/onedriver_tests/dmel.fa", auth)
		inode := NewInodeDriveItem(item)
		return item != nil && inode.Size() == size
	}, 120*time.Second, time.Second, "Upload session did not complete successfully!")

	// test multipart downloads as a bonus part of the test
	downloaded, _, err := graph.GetItemContent(item.ID, auth)
	assert.NoError(t, err)
	assert.Equal(t, graph.QuickXORHash(&contents), graph.QuickXORHash(&downloaded),
		"Downloaded content did not match original content.")
}
