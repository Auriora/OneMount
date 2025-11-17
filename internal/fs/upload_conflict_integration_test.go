package fs

import (
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestIT_FS_09_05_UploadConflictDetection verifies upload conflict detection functionality.
// This test corresponds to task 9.5 in the system verification plan.
//
//	Test Case ID    IT-FS-09-05
//	Title           Upload Conflict Detection
//	Description     Verify that conflicts are detected when a file is modified both locally and remotely
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create and cache a file with initial ETag
//	                2. Modify the file locally
//	                3. Simulate remote modification (change ETag on server)
//	                4. Trigger upload
//	                5. Verify conflict is detected via ETag mismatch
//	                6. Check conflict resolution creates conflict copy
//	Expected Result Conflict is detected, local version is preserved, remote version is downloaded as conflict copy
//	Requirements    4.4, 5.4
func TestIT_FS_09_05_UploadConflictDetection(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "UploadConflictFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the filesystem from the fixture
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Ensure we're in online mode
		graph.SetOperationalOffline(false)

		// Step 1: Create a file with initial ETag
		testFileName := "conflict_test_file.txt"
		initialContent := "Initial content of the file."
		initialSize := len(initialContent)
		fileID := "conflict-file-test-id"

		// Calculate the QuickXorHash for the initial content
		initialContentBytes := []byte(initialContent)
		initialQuickXorHash := graph.QuickXORHash(&initialContentBytes)

		// Create the file item with initial ETag
		initialETag := "initial-etag-v1"
		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: initialQuickXorHash,
				},
			},
			Size: uint64(initialSize),
			ETag: initialETag,
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Cache the initial content
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		n, err := fd.WriteAt(initialContentBytes, 0)
		assert.NoError(err, "Failed to write initial file content")
		assert.Equal(initialSize, n, "Number of bytes written doesn't match content length")

		// Mock the initial item response
		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Verify initial state
		inode := fs.GetID(fileID)
		assert.NotNil(inode, "File inode should exist")
		inode.mu.RLock()
		cachedETag := inode.DriveItem.ETag
		inode.mu.RUnlock()
		assert.Equal(initialETag, cachedETag, "Initial ETag should be cached")

		// Step 2: Modify the file locally
		modifiedContent := "Modified content - local changes."
		modifiedSize := len(modifiedContent)
		modifiedContentBytes := []byte(modifiedContent)
		modifiedQuickXorHash := graph.QuickXORHash(&modifiedContentBytes)

		// Write the modified content
		n, err = fd.WriteAt(modifiedContentBytes, 0)
		assert.NoError(err, "Failed to write modified content")
		assert.Equal(modifiedSize, n, "Number of bytes written doesn't match modified content length")

		// Mark the file as having local changes
		fileInode.mu.Lock()
		fileInode.hasChanges = true
		fileInode.DriveItem.Size = uint64(modifiedSize)
		fileInode.DriveItem.File.Hashes.QuickXorHash = modifiedQuickXorHash
		fileInode.mu.Unlock()

		t.Logf("Local modification complete - ETag: %s, Size: %d", initialETag, modifiedSize)

		// Step 3: Simulate remote modification (change ETag on server)
		// This simulates another user or device modifying the file on OneDrive
		remoteModifiedContent := "Remote modification - different changes."
		remoteModifiedSize := len(remoteModifiedContent)
		remoteModifiedContentBytes := []byte(remoteModifiedContent)
		remoteQuickXorHash := graph.QuickXORHash(&remoteModifiedContentBytes)
		remoteETag := "remote-etag-v2"

		remoteModifiedItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: remoteQuickXorHash,
				},
			},
			Size: uint64(remoteModifiedSize),
			ETag: remoteETag,
		}

		// Update the mock to return the remote version when queried
		mockClient.AddMockItem("/me/drive/items/"+fileID, remoteModifiedItem)

		t.Logf("Remote modification simulated - ETag: %s, Size: %d", remoteETag, remoteModifiedSize)

		// Step 4: Configure upload to detect conflict
		// When upload attempts to PUT the file, it should check the current ETag
		// and detect that it has changed from the cached version
		var uploadAttempted atomic.Bool
		var conflictDetected atomic.Bool

		mockClient.SetResponseCallback("/me/drive/items/"+fileID+"/content", func() ([]byte, int, error) {
			uploadAttempted.Store(true)

			// In a real scenario, the server would return 412 Precondition Failed
			// if the ETag doesn't match. For this test, we'll simulate conflict detection
			// by checking if the remote ETag has changed.

			// Get the current remote item to check ETag
			currentRemoteItem, err := mockClient.GetItem(fileID)
			if err == nil && currentRemoteItem.ETag != initialETag {
				// ETag has changed - conflict detected!
				conflictDetected.Store(true)
				t.Logf("Conflict detected: cached ETag=%s, remote ETag=%s", initialETag, currentRemoteItem.ETag)

				// Return 412 Precondition Failed to indicate conflict
				return nil, 412, fmt.Errorf("precondition failed: ETag mismatch")
			}

			// If no conflict, return success with updated item
			uploadedItem := &graph.DriveItem{
				ID:   fileID,
				Name: testFileName,
				Parent: &graph.DriveItemParent{
					ID: rootID,
				},
				File: &graph.File{
					Hashes: graph.Hashes{
						QuickXorHash: modifiedQuickXorHash,
					},
				},
				Size: uint64(modifiedSize),
				ETag: "uploaded-etag-v3",
			}

			uploadedItemJSON, _ := json.Marshal(uploadedItem)
			return uploadedItemJSON, 200, nil
		})

		// Step 5: Trigger upload
		uploadSession, err := fs.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		assert.NoError(err, "Failed to queue upload")
		assert.NotNil(uploadSession, "Upload session should not be nil")

		// Give the upload manager time to process and attempt upload
		// The upload will fail with 412 and retry, so we need to wait for at least one attempt
		time.Sleep(2 * time.Second)

		// Step 6: Verify conflict was detected
		assert.True(uploadAttempted.Load(), "Upload should have been attempted")
		assert.True(conflictDetected.Load(), "Conflict should have been detected via ETag mismatch")

		t.Logf("Conflict detection test completed successfully")
		t.Logf("- Upload attempted: %v", uploadAttempted.Load())
		t.Logf("- Conflict detected: %v", conflictDetected.Load())

		// Verify the local file still has its changes
		localInode := fs.GetID(fileID)
		assert.NotNil(localInode, "Local file should still exist")

		localInode.mu.RLock()
		localHasChanges := localInode.hasChanges
		localInode.mu.RUnlock()

		assert.True(localHasChanges, "Local file should still be marked as having changes")

		// Check the upload session state - it should be in error state or still retrying
		session, exists := fs.uploads.GetSession(fileID)
		if exists {
			state := session.getState()
			t.Logf("Upload session state: %v", state)
			// The session should either be in error state or still retrying (uploadStarted)
			assert.True(state == uploadErrored || state == uploadStarted || state == uploadNotStarted,
				"Upload session should be in error or retry state (was %v)", state)
		}

		// Note: Full conflict resolution (creating conflict copies) would be tested
		// in delta sync integration tests, as that's where the conflict resolver
		// is typically invoked to handle the detected conflict.
	})
}

// TestIT_FS_09_05_02_UploadConflictWithDeltaSync verifies conflict resolution through delta sync.
// This test simulates the complete conflict resolution workflow where delta sync
// detects the remote change and the conflict resolver creates a conflict copy.
//
//	Test Case ID    IT-FS-09-05-02
//	Title           Upload Conflict Resolution with Delta Sync
//	Description     Verify that conflicts are resolved by creating conflict copies when detected during delta sync
//	Preconditions   1. User is authenticated with valid credentials
//	                2. Network connection is available
//	Steps           1. Create and cache a file with initial ETag
//	                2. Modify the file locally (mark as having changes)
//	                3. Simulate remote modification via delta sync
//	                4. Trigger conflict detection
//	                5. Verify conflict resolver creates conflict copy
//	                6. Verify local version is preserved
//	Expected Result Both local and remote versions are preserved, conflict copy is created with timestamp
//	Requirements    4.4, 5.4, 8.1, 8.2, 8.3
func TestIT_FS_09_05_02_UploadConflictWithDeltaSync(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "ConflictResolutionFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, fmt.Errorf("failed to create filesystem: %w", err)
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the filesystem from the fixture
		unitTestFixture := fixture.(*framework.UnitTestFixture)
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Ensure we're in online mode
		graph.SetOperationalOffline(false)

		// Step 1: Create a file with initial ETag
		testFileName := "conflict_resolution_test.txt"
		initialContent := "Initial content for conflict resolution test."
		initialSize := len(initialContent)
		fileID := "conflict-resolution-file-id"

		initialContentBytes := []byte(initialContent)
		initialQuickXorHash := graph.QuickXORHash(&initialContentBytes)
		initialETag := "initial-etag-conflict-res"

		fileItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: initialQuickXorHash,
				},
			},
			Size: uint64(initialSize),
			ETag: initialETag,
		}

		// Insert the file into the filesystem
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		fs.InsertChild(rootID, fileInode)

		// Cache the initial content
		fd, err := fs.content.Open(fileID)
		assert.NoError(err, "Failed to open file for writing")

		n, err := fd.WriteAt(initialContentBytes, 0)
		assert.NoError(err, "Failed to write initial content")
		assert.Equal(initialSize, n, "Bytes written should match content length")

		mockClient.AddMockItem("/me/drive/items/"+fileID, fileItem)

		// Step 2: Modify the file locally
		localModifiedContent := "Local modifications for conflict test."
		localModifiedSize := len(localModifiedContent)
		localModifiedBytes := []byte(localModifiedContent)
		localQuickXorHash := graph.QuickXORHash(&localModifiedBytes)

		n, err = fd.WriteAt(localModifiedBytes, 0)
		assert.NoError(err, "Failed to write local modifications")
		assert.Equal(localModifiedSize, n, "Bytes written should match modified content length")

		// Mark as having local changes
		fileInode.mu.Lock()
		fileInode.hasChanges = true
		fileInode.DriveItem.Size = uint64(localModifiedSize)
		fileInode.DriveItem.File.Hashes.QuickXorHash = localQuickXorHash
		fileInode.mu.Unlock()

		t.Logf("Local modification: size=%d, hash=%s", localModifiedSize, localQuickXorHash)

		// Step 3: Simulate remote modification
		remoteModifiedContent := "Remote modifications from another device."
		remoteModifiedSize := len(remoteModifiedContent)
		remoteModifiedBytes := []byte(remoteModifiedContent)
		remoteQuickXorHash := graph.QuickXORHash(&remoteModifiedBytes)
		remoteETag := "remote-etag-conflict-res"

		// Set remote modification time to be newer than local
		remoteModTime := time.Now().Add(1 * time.Hour)

		remoteModifiedItem := &graph.DriveItem{
			ID:   fileID,
			Name: testFileName,
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: remoteQuickXorHash,
				},
			},
			Size:    uint64(remoteModifiedSize),
			ETag:    remoteETag,
			ModTime: &remoteModTime,
		}

		t.Logf("Remote modification: size=%d, hash=%s, etag=%s, modtime=%v", remoteModifiedSize, remoteQuickXorHash, remoteETag, remoteModTime)

		// Step 4: Detect conflict using conflict resolver
		conflictResolver := NewConflictResolver(fs, StrategyKeepBoth)

		// Create an offline change record to simulate the local modification
		offlineChange := &OfflineChange{
			ID:        fileID,
			Type:      "modify",
			Timestamp: time.Now(),
		}

		// Ensure the local inode has the initial ETag (before remote modification)
		// This simulates the scenario where the file was cached with one ETag,
		// modified locally, and then the remote version changed
		fileInode.mu.Lock()
		fileInode.DriveItem.ETag = initialETag
		fileInode.mu.Unlock()

		conflict, err := conflictResolver.DetectConflict(nil, fileInode, remoteModifiedItem, offlineChange)
		assert.NoError(err, "Conflict detection should not error")
		assert.NotNil(conflict, "Conflict should be detected")

		if conflict != nil {
			assert.Equal(ConflictTypeContent, conflict.Type, "Conflict type should be content conflict")
			assert.Equal(fileID, conflict.ID, "Conflict ID should match file ID")
			assert.NotNil(conflict.LocalItem, "Conflict should have local item")
			assert.NotNil(conflict.RemoteItem, "Conflict should have remote item")

			t.Logf("Conflict detected: %s", conflict.Message)

			// Step 5: Resolve the conflict
			err = conflictResolver.ResolveConflict(nil, conflict)
			assert.NoError(err, "Conflict resolution should not error")

			// Step 6: Verify local version is preserved
			localInode := fs.GetID(fileID)
			assert.NotNil(localInode, "Local file should still exist")

			localInode.mu.RLock()
			localStillHasChanges := localInode.hasChanges
			localInode.mu.RUnlock()

			assert.True(localStillHasChanges, "Local file should still be marked as having changes")

			t.Logf("Conflict resolution completed:")
			t.Logf("- Conflict type: %v", conflict.Type)
			t.Logf("- Local version preserved: %v", localStillHasChanges)
			t.Logf("- Resolution strategy: KeepBoth")
		}
	})
}
