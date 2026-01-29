package fs

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/auriora/onemount/internal/testutil/framework"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/stretchr/testify/require"

	bolt "go.etcd.io/bbolt"
)

// TestUT_FS_01_01_Cache_BasicOperations_WorkCorrectly tests various cache operations.
//
//	Test Case ID    UT-FS-01-01
//	Title           Cache Operations
//	Description     Tests various cache operations
//	Preconditions   None
//	Steps           1. Create a filesystem cache
//	                2. Perform operations on the cache (get path, get children, check pointers)
//	                3. Verify the results of each operation
//	Expected Result Cache operations work correctly
//	Notes: This test verifies that the cache operations work correctly.
func TestUT_FS_01_01_Cache_BasicOperations_WorkCorrectly(t *testing.T) {
	// Create a test fixture using the common setup
	fixture := helpers.SetupFSTestFixture(t, "CacheOperationsFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
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

		// Ensure root is properly set up in mock mode
		if rootID == "" {
			t.Fatal("Root ID is empty - mock fixture not properly initialized")
		}

		// Step 1: Test basic cache operations

		// Test GetPath operation
		rootInodeByPath, err := fs.GetPath("/", fs.auth)
		assert.NoError(err, "GetPath should not return error")
		assert.NotNil(rootInodeByPath, "Root inode should exist")
		if rootInodeByPath != nil {
			assert.Equal("/", rootInodeByPath.Path(), "Root path should be /")
		}

		// Test GetID operation
		rootInode := fs.GetID(rootID)
		assert.NotNil(rootInode, "Root inode should exist")
		if rootInode != nil {
			assert.Equal(rootID, rootInode.ID(), "Root inode ID should match")
		}

		// Step 2: Test cache insertion and retrieval

		// Create a test file item
		testFileID := "test-cache-file-id"
		testFileName := "cache_test_file.txt"
		fileItem := &graph.DriveItem{
			ID:   testFileID,
			Name: testFileName,
			Size: 1024,
			File: &graph.File{
				Hashes: graph.Hashes{
					QuickXorHash: "test-hash",
				},
			},
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
		}

		// Add mock response for the file
		mockClient.AddMockItem("/me/drive/items/"+testFileID, fileItem)

		// Insert the file into the cache
		fileInode := NewInodeDriveItem(fileItem)
		fs.InsertNodeID(fileInode)
		nodeID := fs.InsertChild(rootID, fileInode)

		// Verify insertion
		assert.NotEqual(uint64(0), nodeID, "Node ID should be assigned")
		assert.Equal(nodeID, fileInode.NodeID(), "Node ID should match inode")

		// Test retrieval by ID
		retrievedInode := fs.GetID(testFileID)
		assert.NotNil(retrievedInode, "File should be retrievable by ID")
		if retrievedInode != nil {
			assert.Equal(testFileID, retrievedInode.ID(), "Retrieved inode ID should match")
			assert.Equal(testFileName, retrievedInode.Name(), "Retrieved inode name should match")
		}

		// Test retrieval by NodeID
		retrievedByNodeID := fs.GetNodeID(nodeID)
		assert.NotNil(retrievedByNodeID, "File should be retrievable by NodeID")
		if retrievedByNodeID != nil {
			assert.Equal(testFileID, retrievedByNodeID.ID(), "Retrieved inode ID should match")
		}

		// Step 3: Test GetChild operation
		childInode, err := fs.GetChild(rootID, testFileName, fs.auth)
		assert.NoError(err, "GetChild should not return error")
		assert.NotNil(childInode, "Child should be found")
		if childInode != nil {
			assert.Equal(testFileID, childInode.ID(), "Child ID should match")
		}

		// Step 4: Test GetChildrenID operation
		// Add mock response for children listing
		mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{fileItem})

		children, err := fs.GetChildrenID(rootID, fs.auth)
		assert.NoError(err, "GetChildrenID should not return error")
		assert.NotNil(children, "Children map should not be nil")
		if children != nil {
			childKey := strings.ToLower(testFileName)
			childInode, exists := children[childKey]
			assert.True(exists, "Children should contain our test file")
			if exists {
				assert.Equal(testFileID, childInode.ID(), "Child in map should have correct ID")
			}
		}

		// Step 5: Test path operations
		expectedPath := "/" + testFileName
		assert.Equal(expectedPath, fileInode.Path(), "File path should be correct")

		// Test that the file can be found by path
		foundInode, err := fs.GetPath(expectedPath, fs.auth)
		assert.NoError(err, "GetPath should find the file")
		assert.NotNil(foundInode, "Found inode should not be nil")
		if foundInode != nil {
			assert.Equal(testFileID, foundInode.ID(), "Found inode ID should match")
		}

		// Step 6: Test cache cleanup and management

		// Test that cache pointers are working correctly
		assert.Equal(fileInode, fs.GetNodeID(nodeID), "Node ID pointer should be consistent")
		assert.Equal(fileInode, fs.GetID(testFileID), "ID pointer should be consistent")
	})
}

// TestUT_FS_01_02_Cache_SkipsXDGVolumeInfoFromServer verifies that remote
// .xdg-volume-info entries returned by the Graph API are ignored so the local
// virtual file can be used instead.
func TestUT_FS_01_02_Cache_SkipsXDGVolumeInfoFromServer(t *testing.T) {
	fixture := helpers.SetupFSTestFixture(t, "SkipXDGVolumeInfoFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, fixture interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := fixture.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", fixture)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		mockClient := fsFixture.MockClient
		rootID := fsFixture.RootID

		xdgItem := &graph.DriveItem{
			ID:   "remote-xdg-id",
			Name: ".xdg-volume-info",
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{QuickXorHash: "fakehash=="},
			},
		}
		regularItem := &graph.DriveItem{
			ID:   "regular-file-id",
			Name: "regular.txt",
			Parent: &graph.DriveItemParent{
				ID: rootID,
			},
			File: &graph.File{
				Hashes: graph.Hashes{QuickXorHash: "anotherhash=="},
			},
			Size: 42,
		}

		mockClient.AddMockItem("/me/drive/items/"+rootID, &graph.DriveItem{ID: rootID, Name: "root", Folder: &graph.Folder{}})
		mockClient.AddMockItem("/me/drive/items/"+regularItem.ID, regularItem)
		mockClient.AddMockItem("/me/drive/items/"+xdgItem.ID, xdgItem)
		mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{xdgItem, regularItem})

		children, err := fs.GetChildrenID(rootID, fsFixture.Auth)
		assert.NoError(err, "GetChildrenID should not error")
		assert.NotNil(children, "Children map should not be nil")
		if _, exists := children[strings.ToLower(xdgItem.Name)]; exists {
			t.Fatalf("Remote .xdg-volume-info should be ignored but was present in children map")
		}
		assert.Nil(fs.GetID(xdgItem.ID), "Remote .xdg-volume-info should not be cached")
	})
}

// TestGetChildrenIDUsesMetadataStoreWhenOffline ensures cached metadata can satisfy directory listings without Graph access.
func TestIT_FS_Cache_GetChildrenIDUsesMetadataStoreWhenOffline(t *testing.T) {
	tempSandbox := filepath.Join(os.TempDir(), "onemount-tests")
	originalSandbox := testutil.TestSandboxDir
	originalTmp := testutil.TestSandboxTmpDir
	originalAuth := testutil.AuthTokensPath
	originalLog := testutil.TestLogPath
	originalGraph := testutil.GraphTestDir
	originalMount := testutil.TestMountPoint
	originalDir := testutil.TestDir
	originalSystemMount := testutil.SystemTestMountPoint
	originalSystemData := testutil.SystemTestDataDir
	originalSystemLog := testutil.SystemTestLogPath

	testutil.TestSandboxDir = tempSandbox
	testutil.TestSandboxTmpDir = filepath.Join(tempSandbox, "tmp")
	testutil.AuthTokensPath = filepath.Join(tempSandbox, ".auth_tokens.json")
	testutil.TestLogPath = filepath.Join(tempSandbox, "logs", "fusefs_tests.log")
	testutil.GraphTestDir = filepath.Join(tempSandbox, "graph_test_dir")
	testutil.TestMountPoint = filepath.Join(testutil.TestSandboxTmpDir, "mount")
	testutil.TestDir = filepath.Join(testutil.TestMountPoint, "onemount_tests")
	testutil.SystemTestMountPoint = filepath.Join(testutil.TestSandboxTmpDir, "system-test-mount")
	testutil.SystemTestDataDir = filepath.Join(tempSandbox, "system-test-data")
	testutil.SystemTestLogPath = filepath.Join(tempSandbox, "logs", "system_tests.log")

	t.Cleanup(func() {
		testutil.TestSandboxDir = originalSandbox
		testutil.TestSandboxTmpDir = originalTmp
		testutil.AuthTokensPath = originalAuth
		testutil.TestLogPath = originalLog
		testutil.GraphTestDir = originalGraph
		testutil.TestMountPoint = originalMount
		testutil.TestDir = originalDir
		testutil.SystemTestMountPoint = originalSystemMount
		testutil.SystemTestDataDir = originalSystemData
		testutil.SystemTestLogPath = originalSystemLog
	})

	if err := helpers.EnsureTestDirectories(); err != nil {
		t.Fatalf("Failed to prepare test directories: %v", err)
	}
	fixture := helpers.SetupFSTestFixture(t, "MetadataChildrenRecoveryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
		return NewFilesystem(auth, mountPoint, cacheTTL)
	})

	fixture.Use(t, func(t *testing.T, f interface{}) {
		assert := framework.NewAssert(t)

		unitTestFixture, ok := f.(*framework.UnitTestFixture)
		if !ok {
			t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", f)
		}
		fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
		fs := fsFixture.FS.(*Filesystem)
		rootID := fsFixture.RootID

		file := helpers.CreateMockFile(fsFixture.MockClient, rootID, "metadata-recovery.txt", "metadata-recovery-id", "hello metadata")
		if fs.metadataStore != nil {
			_ = fs.metadataStore.Save(context.Background(), &metadata.Entry{
				ID:        file.ID,
				ParentID:  rootID,
				Name:      file.Name,
				ItemType:  metadata.ItemKindFile,
				State:     metadata.ItemStateHydrated,
				ETag:      "etag-recovery",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			})
			_, _ = fs.metadataStore.Update(context.Background(), rootID, func(entry *metadata.Entry) error {
				if entry == nil {
					return metadata.ErrNotFound
				}
				entry.Children = []string{file.ID}
				return nil
			})
		}
		children, err := fs.GetChildrenID(rootID, fs.auth)
		assert.NoError(err, "Initial metadata fetch should succeed")
		childInode, exists := children[strings.ToLower(file.Name)]
		if !assert.True(exists, "Fetched children should include the mock file") {
			return
		}

		fs.metadata.Delete(childInode.ID())
		parent := fs.GetID(rootID)
		assert.NotNil(parent, "Root inode should still exist")
		parent.mu.Lock()
		parent.children = nil
		parent.subdir = 0
		parent.mu.Unlock()

		graph.SetOperationalOffline(true)
		defer graph.SetOperationalOffline(false)

		restoredChildren, err := fs.GetChildrenID(rootID, fs.auth)
		assert.NoError(err, "GetChildrenID should read from structured metadata while offline")
		restored, ok := restoredChildren[strings.ToLower(file.Name)]
		if !assert.True(ok, "Children map should include entry restored from metadata") {
			return
		}
		assert.Equal(childInode.ID(), restored.ID(), "Restored inode should match original child ID")
	})
}

func TestIT_FS_Cache_GetPathUsesMetadataStoreWhenOffline(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "MetadataPathRecoveryFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, data interface{}) {
			unitTestFixture, ok := data.(*framework.UnitTestFixture)
			if !ok {
				t.Fatalf("Expected fixture to be of type *framework.UnitTestFixture, but got %T", data)
			}
			fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
			fs := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID

			fileItem := helpers.CreateMockFile(fsFixture.MockClient, rootID, "metadata-path.txt", "metadata-path-id", "path-metadata-content")
			if fs.metadataStore != nil {
				_ = fs.metadataStore.Save(context.Background(), &metadata.Entry{
					ID:        fileItem.ID,
					ParentID:  rootID,
					Name:      fileItem.Name,
					ItemType:  metadata.ItemKindFile,
					State:     metadata.ItemStateHydrated,
					ETag:      "etag-path",
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				})
				_, _ = fs.metadataStore.Update(context.Background(), rootID, func(entry *metadata.Entry) error {
					if entry == nil {
						return metadata.ErrNotFound
					}
					entry.Children = []string{fileItem.ID}
					return nil
				})
			}

			_, err := fs.GetPath("/"+fileItem.Name, fs.auth)
			require.NoError(t, err, "Initial GetPath should succeed online")

			fs.metadata.Delete(fileItem.ID)
			parent := fs.GetID(rootID)
			require.NotNil(t, parent, "Root inode should exist in memory")
			parent.mu.Lock()
			parent.children = nil
			parent.subdir = 0
			parent.mu.Unlock()

			graph.SetOperationalOffline(true)
			defer graph.SetOperationalOffline(false)

			inode, err := fs.GetPath("/"+fileItem.Name, fs.auth)
			require.NoError(t, err, "GetPath should resolve from metadata store while offline")
			require.NotNil(t, inode)
			require.Equal(t, fileItem.ID, inode.ID(), "Offline GetPath should return the same inode")
		})
	})
}

// TestUT_FS_Cache_GetChildrenIDReturnsPrefetchedDataQuickly ensures prefetched data returns immediately.
func TestUT_FS_Cache_GetChildrenIDReturnsPrefetchedDataQuickly(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "PrefetchImmediateFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, data interface{}) {
			unitTestFixture := data.(*framework.UnitTestFixture)
			fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
			fs := fsFixture.FS.(*Filesystem)
			mockClient := fsFixture.MockClient
			rootID := fsFixture.RootID

			if mockClient == nil {
				t.Skip("Skipping prefetch test: requires mock graph client")
				return
			}

			waitForPrefetchIdle(t, fs)

			dirItem := helpers.CreateMockDirectory(mockClient, rootID, "prefetch-dir", "prefetch-dir-id")
			require.NotNil(t, dirItem)
			fileItem := helpers.CreateMockFile(mockClient, dirItem.ID, "prefetch.txt", "prefetch-file-id", "prefetch")
			require.NotNil(t, fileItem)

			dirInode := NewInodeDriveItem(dirItem)
			fs.InsertChild(rootID, dirInode)
			fs.persistMetadataEntry(dirItem.ID, dirInode)
			fs.persistMetadataEntry(rootID, fs.GetID(rootID))

			_, err := fs.prefetchDirectory(context.Background(), dirItem.ID, 0)
			require.NoError(t, err)

			start := time.Now()
			children, err := fs.GetChildrenID(dirItem.ID, fs.auth)
			elapsed := time.Since(start)

			require.NoError(t, err)
			require.Len(t, children, 1, "Prefetched directory should return children immediately")
			require.Less(t, elapsed, 50*time.Millisecond, "Prefetched access should be < 50ms")
		})
	})
}

// TestUT_FS_Cache_GetChildrenIDWaitsForPrefetch ensures GetChildrenID waits for HYDRATING prefetch.
func TestUT_FS_Cache_GetChildrenIDWaitsForPrefetch(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "PrefetchWaitFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, data interface{}) {
			unitTestFixture := data.(*framework.UnitTestFixture)
			fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
			fs := fsFixture.FS.(*Filesystem)
			mockClient := fsFixture.MockClient
			rootID := fsFixture.RootID

			if mockClient == nil {
				t.Skip("Skipping prefetch test: requires mock graph client")
				return
			}

			waitForPrefetchIdle(t, fs)

			testFile := helpers.CreateMockFile(mockClient, rootID, "wait-file.txt", "prefetch-wait-file", "wait")
			require.NotNil(t, testFile)
			mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{testFile})

			if root := fs.GetID(rootID); root != nil {
				root.mu.Lock()
				root.children = nil
				root.subdir = 0
				root.mu.Unlock()
			}

			_, err := fs.UpdateMetadataEntry(rootID, func(entry *metadata.Entry) error {
				if entry == nil {
					return metadata.ErrNotFound
				}
				entry.State = metadata.ItemStateHydrating
				entry.Children = nil
				entry.SubdirCount = 0
				return nil
			})
			require.NoError(t, err)

			delay := 150 * time.Millisecond
			go func() {
				time.Sleep(delay)
				childInode := NewInodeDriveItem(testFile)
				fs.InsertChild(rootID, childInode)
				fs.cacheChildrenFromMap(rootID, map[string]*Inode{strings.ToLower(testFile.Name): childInode})
				_, _ = fs.UpdateMetadataEntry(rootID, func(entry *metadata.Entry) error {
					if entry != nil {
						entry.State = metadata.ItemStateHydrated
						entry.Children = []string{testFile.ID}
						entry.SubdirCount = 0
					}
					return nil
				})
			}()

			start := time.Now()
			children, err := fs.GetChildrenID(rootID, fs.auth)
			elapsed := time.Since(start)

			require.NoError(t, err)
			require.Len(t, children, 1, "Prefetch wait should return populated children")
			require.GreaterOrEqual(t, elapsed, delay, "GetChildrenID should wait for prefetch completion")
			require.Less(t, elapsed, 5*time.Second, "Prefetch wait should respect 5-second timeout")
		})
	})
}

func TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "MetadataAsyncRefreshFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, data interface{}) {
			unitTestFixture := data.(*framework.UnitTestFixture)
			fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
			fs := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID
			mockClient := fsFixture.MockClient

			// Create a test file in the mock
			testFile := helpers.CreateMockFile(mockClient, rootID, "test-file.txt", "test-file-id", "test content")
			mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{testFile})

			root := fs.GetID(rootID)
			require.NotNil(t, root)
			root.mu.Lock()
			root.children = nil
			root.subdir = 0
			root.mu.Unlock()

			if fs.metadataStore != nil {
				_, _ = fs.metadataStore.Update(context.Background(), rootID, func(entry *metadata.Entry) error {
					if entry == nil {
						return metadata.ErrNotFound
					}
					entry.Children = nil
					entry.SubdirCount = 0
					return nil
				})
			}

			resource := "/me/drive/items/" + rootID + "/children"
			if response, ok := mockClient.RequestResponses[resource]; ok {
				delay := 200 * time.Millisecond
				mockClient.SetResponseCallback(resource, func() ([]byte, int, error) {
					time.Sleep(delay)
					return response.Body, response.StatusCode, response.Error
				})
			}

			// UPDATED TEST: Now expects blocking behavior (Requirement 7.1)
			// GetChildrenID should block and fetch synchronously when cache is empty
			// It should NEVER return empty when data exists (Requirement 1.2, 7.2)
			start := time.Now()
			children, err := fs.GetChildrenID(rootID, fs.auth)
			elapsed := time.Since(start)

			require.NoError(t, err, "GetChildrenID should not return error")
			require.NotNil(t, children, "Children map should not be nil")
			require.Len(t, children, 1, "Should return complete data (not empty) on first access")

			// Verify the test file is in the results
			_, exists := children[strings.ToLower(testFile.Name)]
			require.True(t, exists, "Test file should be present in children")

			// Should block until data is fetched (up to 10 seconds timeout per Requirement 1.3, 7.3)
			// In practice, with mock, this should complete quickly but not instantly
			require.Less(t, elapsed, 10*time.Second, "Should complete within 10-second timeout")
			require.GreaterOrEqual(t, elapsed, 200*time.Millisecond, "Should block until fetch completes")

			t.Logf("GetChildrenID completed in %v (blocking fetch as expected)", elapsed)
		})
	})
}

func TestIT_FS_Cache_GetChildrenIDDoesNotCallGraphWhenMetadataPresent(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "MetadataLocalOnlyFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, data interface{}) {
			unit := data.(*framework.UnitTestFixture)
			fsFixture := unit.SetupData.(*helpers.FSTestFixture)
			fs := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID

			// Seed metadata store with a child entry but make Graph return nothing.
			child := helpers.CreateMockFile(fsFixture.MockClient, rootID, "local-only.txt", "local-only-id", "payload")
			if fs.metadataStore != nil {
				_ = fs.metadataStore.Save(context.Background(), &metadata.Entry{
					ID:        child.ID,
					ParentID:  rootID,
					Name:      child.Name,
					ItemType:  metadata.ItemKindFile,
					State:     metadata.ItemStateHydrated,
					ETag:      "etag",
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
				})
				_, _ = fs.metadataStore.Update(context.Background(), rootID, func(entry *metadata.Entry) error {
					if entry == nil {
						return metadata.ErrNotFound
					}
					entry.Children = []string{child.ID}
					entry.SubdirCount = 0
					entry.State = metadata.ItemStateHydrated
					return nil
				})
			}
			graph.SetOperationalOffline(true)
			defer graph.SetOperationalOffline(false)

			// Invalidate in-memory cache to force GetChildrenID to rely on metadata, not Graph.
			if root := fs.GetID(rootID); root != nil {
				root.mu.Lock()
				root.children = nil
				root.subdir = 0
				root.mu.Unlock()
			}

			children, err := fs.GetChildrenID(rootID, fs.auth)
			require.NoError(t, err)
			if _, ok := children[strings.ToLower(child.Name)]; !ok {
				t.Fatalf("expected child %s to be returned from metadata without Graph", child.Name)
			}
		})
	})
}

func TestIT_FS_Cache_FallbackRootFromMetadata(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "meta.db")
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		t.Fatalf("failed to open bolt db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(bucketMetadataV2)
		if err != nil {
			return err
		}
		entry := &metadata.Entry{
			ID:        "root-entry",
			Name:      "Root",
			ItemType:  metadata.ItemKindDirectory,
			State:     metadata.ItemStateHydrated,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		blob, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			return marshalErr
		}
		return bucket.Put([]byte(entry.ID), blob)
	}); err != nil {
		t.Fatalf("failed to seed metadata_v2 bucket: %v", err)
	}

	fs := &Filesystem{db: db}
	root := fs.fallbackRootFromMetadata()
	if root == nil {
		t.Fatalf("expected fallback root inode")
	}
	if root.ID() != "root-entry" {
		t.Fatalf("expected root-entry got %s", root.ID())
	}
	if val, ok := fs.metadata.Load("root"); !ok || val == nil {
		t.Fatalf("expected synthetic root cached in metadata map")
	}
}

// TestIT_FS_Cache_StaleCacheRefreshWithTimeout tests the stale cache refresh policy
// Requirement 4.2, 7.6: Test stale cache refresh with 2-second timeout
// Requirement 4.4: Verify serves stale data if refresh times out
// Requirement 4.7: Verify background refresh continues
func TestIT_FS_Cache_StaleCacheRefreshWithTimeout(t *testing.T) {
	withTempSandbox(t, func() {
		fixture := helpers.SetupFSTestFixture(t, "StaleCacheRefreshFixture", func(auth *graph.Auth, mountPoint string, cacheTTL int) (interface{}, error) {
			return NewFilesystem(auth, mountPoint, cacheTTL)
		})

		fixture.Use(t, func(t *testing.T, data interface{}) {
			unitTestFixture := data.(*framework.UnitTestFixture)
			fsFixture := unitTestFixture.SetupData.(*helpers.FSTestFixture)
			fs := fsFixture.FS.(*Filesystem)
			rootID := fsFixture.RootID
			mockClient := fsFixture.MockClient

			// Create initial test files in the mock
			testFile1 := helpers.CreateMockFile(mockClient, rootID, "file1.txt", "file1-id", "content1")
			testFile2 := helpers.CreateMockFile(mockClient, rootID, "file2.txt", "file2-id", "content2")
			mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{testFile1, testFile2})

			// First access - populate cache
			children1, err := fs.GetChildrenID(rootID, fs.auth)
			require.NoError(t, err)
			require.Len(t, children1, 2, "Should have 2 files initially")

			// Make the cache stale by updating the metadata entry's UpdatedAt timestamp
			// to be older than the TTL (5 minutes)
			if fs.metadataStore != nil {
				_, err := fs.metadataStore.Update(context.Background(), rootID, func(entry *metadata.Entry) error {
					if entry == nil {
						return metadata.ErrNotFound
					}
					// Set UpdatedAt to 10 minutes ago (older than 5-minute TTL)
					entry.UpdatedAt = time.Now().Add(-10 * time.Minute)
					return nil
				})
				require.NoError(t, err, "Should be able to update metadata timestamp")
			}

			// Verify cache is now stale
			isFresh := fs.isCacheFresh(rootID)
			require.False(t, isFresh, "Cache should be stale after timestamp manipulation")

			// Add a new file to the mock (simulating remote changes)
			testFile3 := helpers.CreateMockFile(mockClient, rootID, "file3.txt", "file3-id", "content3")
			mockClient.AddMockItems("/me/drive/items/"+rootID+"/children", []*graph.DriveItem{testFile1, testFile2, testFile3})

			// Second access with stale cache - should attempt refresh
			// If refresh succeeds within 2 seconds, should return fresh data (3 files)
			// If refresh times out, should return stale data (2 files) and continue in background
			start := time.Now()
			children2, err := fs.GetChildrenID(rootID, fs.auth)
			elapsed := time.Since(start)

			require.NoError(t, err, "Should not return error even if refresh times out")
			require.NotNil(t, children2, "Should never return nil")

			// The result depends on whether refresh completed within 2 seconds
			// With mock client, it should complete quickly and return fresh data
			if len(children2) == 3 {
				// Refresh succeeded within timeout - got fresh data (Requirement 4.3)
				t.Logf("Stale cache refresh succeeded within timeout (%v), returned fresh data with 3 files", elapsed)
				_, exists := children2[strings.ToLower(testFile3.Name)]
				require.True(t, exists, "New file should be present in refreshed data")
			} else if len(children2) == 2 {
				// Refresh timed out - got stale data (Requirement 4.4)
				t.Logf("Stale cache refresh timed out (%v), served stale data with 2 files", elapsed)
				require.Less(t, elapsed, 3*time.Second, "Should return stale data within ~2 seconds")

				// Background refresh should continue (Requirement 4.7)
				// Wait a bit for background refresh to complete
				time.Sleep(3 * time.Second)

				// Third access should now have fresh data from background refresh
				children3, err := fs.GetChildrenID(rootID, fs.auth)
				require.NoError(t, err)
				require.Len(t, children3, 3, "Background refresh should have updated cache")
			} else {
				t.Fatalf("Unexpected number of children: %d (expected 2 or 3)", len(children2))
			}

			// Verify cache is fresh after refresh
			isFresh = fs.isCacheFresh(rootID)
			require.True(t, isFresh, "Cache should be fresh after refresh")
		})
	})
}
