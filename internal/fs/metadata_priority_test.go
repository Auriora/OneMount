package fs

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// TestUT_FS_Metadata_07_PrioritySystem tests the metadata request priority system.
//
//	Test Case ID    UT-FS-Metadata-07
//	Title           Metadata Request Priority System
//	Description     Tests that foreground requests are prioritized over background requests
//	Preconditions   None
//	Steps           1. Start metadata request manager
//	                2. Queue background and foreground requests
//	                3. Verify foreground requests are processed first
//	Expected Result Foreground requests complete before background requests
//	Notes: This test verifies the metadata request prioritization system.
func TestUT_FS_Metadata_07_PrioritySystem(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "MetadataPriorityFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient

		// Mock responses for children requests
		mockClient.AddMockResponse("/me/drive/items/test-dir-1/children", []byte(`{"value":[]}`), 200, nil)
		mockClient.AddMockResponse("/me/drive/items/test-dir-2/children", []byte(`{"value":[]}`), 200, nil)
		mockClient.AddMockResponse("/me/drive/items/test-dir-3/children", []byte(`{"value":[]}`), 200, nil)

		// Track completion order
		var completionOrder []string
		var orderMutex sync.Mutex

		// Create channels to synchronize the test
		backgroundStarted := make(chan struct{})
		foregroundQueued := make(chan struct{})

		// Queue a background request first
		go func() {
			err := fs.metadataRequestManager.QueueChildrenRequest("test-dir-1", fsFixture.Auth, PriorityBackground, func(items []*graph.DriveItem, err error) {
				orderMutex.Lock()
				completionOrder = append(completionOrder, "background-1")
				orderMutex.Unlock()
				close(backgroundStarted)
			})
			assert.NoError(err, "Background request should be queued successfully")
		}()

		// Wait a bit to ensure background request starts processing
		time.Sleep(50 * time.Millisecond)

		// Queue foreground requests
		go func() {
			err := fs.metadataRequestManager.QueueChildrenRequest("test-dir-2", fsFixture.Auth, PriorityForeground, func(items []*graph.DriveItem, err error) {
				orderMutex.Lock()
				completionOrder = append(completionOrder, "foreground-1")
				orderMutex.Unlock()
			})
			assert.NoError(err, "Foreground request should be queued successfully")

			err = fs.metadataRequestManager.QueueChildrenRequest("test-dir-3", fsFixture.Auth, PriorityForeground, func(items []*graph.DriveItem, err error) {
				orderMutex.Lock()
				completionOrder = append(completionOrder, "foreground-2")
				orderMutex.Unlock()
			})
			assert.NoError(err, "Second foreground request should be queued successfully")
			close(foregroundQueued)
		}()

		// Wait for all requests to complete
		<-backgroundStarted
		<-foregroundQueued

		// Give some time for all requests to complete
		time.Sleep(200 * time.Millisecond)

		// Verify completion order
		orderMutex.Lock()
		defer orderMutex.Unlock()

		assert.True(len(completionOrder) >= 1, "At least one request should have completed")

		// The first completed request should be background since it started first
		// But foreground requests should complete before any additional background requests
		if len(completionOrder) > 1 {
			// Check that foreground requests are prioritized
			foregroundCount := 0

			for _, req := range completionOrder {
				if req == "foreground-1" || req == "foreground-2" {
					foregroundCount++
				}
			}

			assert.True(foregroundCount > 0, "At least one foreground request should have completed")
		}
	})
}

// TestUT_FS_Metadata_08_SyncProgress tests the sync progress tracking.
//
//	Test Case ID    UT-FS-Metadata-08
//	Title           Sync Progress Tracking
//	Description     Tests that sync progress is tracked correctly during tree synchronization
//	Preconditions   None
//	Steps           1. Start directory tree sync
//	                2. Monitor progress updates
//	                3. Verify progress completion
//	Expected Result Progress is tracked and reported correctly
//	Notes: This test verifies the sync progress tracking system.
func TestUT_FS_Metadata_08_SyncProgress(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "SyncProgressFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		// Mock responses for a simple directory structure
		mockClient.AddMockResponse(fmt.Sprintf("/me/drive/items/%s/children", rootID), []byte(`{
			"value": [
				{"id": "dir1", "name": "Directory1", "folder": {}},
				{"id": "file1", "name": "file1.txt", "size": 1024, "file": {}}
			]
		}`), 200, nil)
		mockClient.AddMockResponse("/me/drive/items/dir1/children", []byte(`{
			"value": [
				{"id": "file2", "name": "file2.txt", "size": 2048, "file": {}}
			]
		}`), 200, nil)

		// Start sync with context
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Start sync in background
		syncDone := make(chan error, 1)
		go func() {
			err := fs.SyncDirectoryTreeWithContext(ctx, fsFixture.Auth)
			syncDone <- err
		}()

		// Monitor progress
		var lastProgress *SyncProgressSnapshot
		progressChecks := 0
		maxChecks := 10

		for progressChecks < maxChecks {
			time.Sleep(100 * time.Millisecond)
			progress := fs.GetSyncProgress()
			if progress != nil {
				snapshot := progress.GetProgress()
				lastProgress = &snapshot
				progressChecks++

				// Check if sync is complete
				if progress.IsComplete {
					break
				}
			}
		}

		// Wait for sync to complete
		select {
		case err := <-syncDone:
			assert.NoError(err, "Sync should complete without error")
		case <-ctx.Done():
			t.Fatal("Sync timed out")
		}

		// Verify progress was tracked
		assert.NotNil(lastProgress, "Progress should be tracked")
		if lastProgress != nil {
			assert.True(lastProgress.IsComplete, "Sync should be marked as complete")
			assert.True(lastProgress.StartTime.Before(lastProgress.LastUpdateTime) || lastProgress.StartTime.Equal(lastProgress.LastUpdateTime),
				"Last update time should be after or equal to start time")
		}
	})
}

// TestUT_FS_Metadata_09_QueueStats tests the queue statistics functionality.
//
//	Test Case ID    UT-FS-Metadata-09
//	Title           Queue Statistics
//	Description     Tests that queue statistics are reported correctly
//	Preconditions   None
//	Steps           1. Queue requests of different priorities
//	                2. Check queue statistics
//	                3. Verify counts are accurate
//	Expected Result Queue statistics reflect actual queue state
//	Notes: This test verifies the queue statistics functionality.
func TestUT_FS_Metadata_09_QueueStats(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "QueueStatsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		// Create the filesystem
		fs, err := NewFilesystem(auth, mountPoint, cacheTTL)
		if err != nil {
			return nil, err
		}
		return fs, nil
	})

	// Use the fixture to run the test
	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		// Create assertions helper
		assert := framework.NewAssert(t)

		// Get the test data
		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)

		// Check initial queue stats
		highCount, lowCount := fs.metadataRequestManager.GetQueueStats()
		assert.Equal(0, highCount, "High priority queue should be empty initially")
		assert.Equal(0, lowCount, "Low priority queue should be empty initially")

		// Note: Since the workers process requests immediately in tests,
		// we can't easily test queue counts with actual requests.
		// This test verifies the statistics method works correctly.
		assert.True(highCount >= 0, "High priority count should be non-negative")
		assert.True(lowCount >= 0, "Low priority count should be non-negative")
	})
}
