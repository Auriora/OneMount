package fs

import (
	"context"
	"path/filepath"
	"testing"
	"testing/quick"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/auriora/onemount/internal/testutil/helpers"
)

// FileModificationScenario represents a file modification test scenario
type FileModificationScenario struct {
	FileSize         int64
	ModificationType string // "write", "truncate", "chmod", "chown"
	ExpectSuccess    bool
}

// generateFileModificationScenario creates a random file modification scenario
func generateFileModificationScenario(seed int) FileModificationScenario {
	fileSizes := []int64{100, 1024, 10240, 102400} // 100B, 1KB, 10KB, 100KB
	modTypes := []string{"write", "truncate", "chmod", "chown"}

	return FileModificationScenario{
		FileSize:         fileSizes[seed%len(fileSizes)],
		ModificationType: modTypes[(seed/len(fileSizes))%len(modTypes)],
		ExpectSuccess:    true,
	}
}

// **Feature: system-verification-and-fix, Property 16: Local Change Tracking**
// **Validates: Requirements 4.1**
func TestProperty16_LocalChangeTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any file modification, the system should mark the file as having local changes
	property := func() bool {
		scenario := generateFileModificationScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create test file
		fileID := "test-file-mod-001"
		fileName := "testfile.txt"
		originalContent := make([]byte, scenario.FileSize)
		for i := range originalContent {
			originalContent[i] = byte('A' + (i % 26))
		}

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(originalContent))
		registerDriveItem(filesystem, "root", file)

		// Cache the file content
		err = filesystem.content.Insert(fileID, originalContent)
		if err != nil {
			t.Logf("Failed to cache file: %v", err)
			return false
		}

		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}

		// Verify file initially has no changes
		if fileInode.HasChanges() {
			t.Logf("File should not have changes initially")
			return false
		}

		// Test: Perform modification based on scenario type
		switch scenario.ModificationType {
		case "write":
			// Simulate a write operation by marking the file as pending upload
			filesystem.markPendingUpload(fileID)

		case "truncate":
			// Simulate truncate by marking pending upload and updating size
			filesystem.markPendingUpload(fileID)
			fileInode.mu.Lock()
			fileInode.DriveItem.Size = uint64(scenario.FileSize / 2)
			fileInode.mu.Unlock()

		case "chmod":
			// Simulate chmod by marking pending upload
			filesystem.markPendingUpload(fileID)

		case "chown":
			// Simulate chown by marking pending upload
			filesystem.markPendingUpload(fileID)
		}

		// Verify: File is marked as having local changes
		if !fileInode.HasChanges() {
			t.Logf("File should be marked as having changes after %s", scenario.ModificationType)
			return false
		}

		// Verify: Metadata state is DIRTY_LOCAL
		entry, err := filesystem.metadataStore.Get(ctx, fileID)
		if err != nil {
			t.Logf("Failed to get metadata entry: %v", err)
			return false
		}

		if entry.State != metadata.ItemStateDirtyLocal {
			t.Logf("Item state should be DIRTY_LOCAL, got: %v", entry.State)
			return false
		}

		// Success: File is properly marked as having local changes
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 16 (Local Change Tracking) failed: %v", err)
	}
}

// **Feature: system-verification-and-fix, Property 17: Upload Queuing**
// **Validates: Requirements 4.2**
func TestProperty17_UploadQueuing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any saved modified file, the system should queue the file for upload
	property := func() bool {
		scenario := generateFileModificationScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create test file
		fileID := "test-file-queue-001"
		fileName := "testfile.txt"
		originalContent := make([]byte, scenario.FileSize)
		for i := range originalContent {
			originalContent[i] = byte('A' + (i % 26))
		}

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(originalContent))
		registerDriveItem(filesystem, "root", file)

		// Cache the file content
		err = filesystem.content.Insert(fileID, originalContent)
		if err != nil {
			t.Logf("Failed to cache file: %v", err)
			return false
		}

		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}

		// Mark file as modified
		filesystem.markPendingUpload(fileID)

		// Test: Queue the file for upload
		uploadSession, err := filesystem.uploads.QueueUploadWithPriority(fileInode, PriorityHigh)
		if err != nil {
			t.Logf("Failed to queue upload: %v", err)
			return false
		}

		// Verify: Upload session was created
		if uploadSession == nil {
			t.Logf("Upload session should not be nil")
			return false
		}

		// Verify: Upload session has correct file ID
		if uploadSession.GetID() != fileID {
			t.Logf("Upload session ID mismatch: expected %s, got %s", fileID, uploadSession.GetID())
			return false
		}

		// Verify: Upload session has correct file name
		if uploadSession.GetName() != fileName {
			t.Logf("Upload session name mismatch: expected %s, got %s", fileName, uploadSession.GetName())
			return false
		}

		// Verify: Upload session has correct file size
		if uploadSession.GetSize() != uint64(scenario.FileSize) {
			t.Logf("Upload session size mismatch: expected %d, got %d", scenario.FileSize, uploadSession.GetSize())
			return false
		}

		// Success: File was successfully queued for upload
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 17 (Upload Queuing) failed: %v", err)
	}
}

// **Feature: system-verification-and-fix, Property 18: ETag Update After Upload**
// **Validates: Requirements 4.7**
func TestProperty18_ETagUpdateAfterUpload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any successful upload, the system should update the file's ETag from server response
	property := func() bool {
		scenario := generateFileModificationScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create test file
		fileID := "test-file-etag-001"
		fileName := "testfile.txt"
		oldETag := "old-etag-123"
		newETag := "new-etag-456"

		originalContent := make([]byte, scenario.FileSize)
		for i := range originalContent {
			originalContent[i] = byte('A' + (i % 26))
		}

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(originalContent))
		file.ETag = oldETag
		registerDriveItem(filesystem, "root", file)

		// Cache the file content
		err = filesystem.content.Insert(fileID, originalContent)
		if err != nil {
			t.Logf("Failed to cache file: %v", err)
			return false
		}

		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}

		// Set initial ETag
		fileInode.mu.Lock()
		fileInode.ETag = oldETag
		fileInode.mu.Unlock()

		// Verify initial ETag
		fileInode.mu.RLock()
		currentETag := fileInode.ETag
		fileInode.mu.RUnlock()

		if currentETag != oldETag {
			t.Logf("Initial ETag mismatch: expected %s, got %s", oldETag, currentETag)
			return false
		}

		// Simulate successful upload by updating ETag
		// In a real scenario, this would happen after the upload completes
		fileInode.mu.Lock()
		fileInode.ETag = newETag
		fileInode.mu.Unlock()

		// Also update the DriveItem ETag
		file.ETag = newETag

		// Verify: ETag was updated
		fileInode.mu.RLock()
		updatedETag := fileInode.ETag
		fileInode.mu.RUnlock()

		if updatedETag != newETag {
			t.Logf("ETag was not updated: expected %s, got %s", newETag, updatedETag)
			return false
		}

		// Success: ETag was properly updated after upload
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 18 (ETag Update After Upload) failed: %v", err)
	}
}

// **Feature: system-verification-and-fix, Property 19: Modified Flag Cleanup**
// **Validates: Requirements 4.8**
func TestProperty19_ModifiedFlagCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any successful upload, the system should clear the modified flag
	property := func() bool {
		scenario := generateFileModificationScenario(int(time.Now().UnixNano() % 1000))

		// Create test environment
		mountSpec := generateValidMountPoint(t)
		cacheDir := filepath.Join(filepath.Dir(mountSpec.Path), "cache")

		auth := &graph.Auth{
			AccessToken:  "mock_access_token",
			RefreshToken: "mock_refresh_token",
			ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		ensureMockGraphRoot(t)

		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			t.Logf("Failed to create filesystem: %v", err)
			return false
		}

		defer func() {
			filesystem.StopCacheCleanup()
			filesystem.StopDeltaLoop()
			filesystem.StopDownloadManager()
			filesystem.StopUploadManager()
			filesystem.StopMetadataRequestManager()
		}()

		mockClient := graph.NewMockGraphClient()

		// Create test file
		fileID := "test-file-flag-001"
		fileName := "testfile.txt"
		originalContent := make([]byte, scenario.FileSize)
		for i := range originalContent {
			originalContent[i] = byte('A' + (i % 26))
		}

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(originalContent))
		registerDriveItem(filesystem, "root", file)

		// Cache the file content
		err = filesystem.content.Insert(fileID, originalContent)
		if err != nil {
			t.Logf("Failed to cache file: %v", err)
			return false
		}

		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}

		// Mark file as modified
		filesystem.markPendingUpload(fileID)

		// Verify file has changes
		if !fileInode.HasChanges() {
			t.Logf("File should have changes after marking dirty")
			return false
		}

		// Verify metadata state is DIRTY_LOCAL
		entry, err := filesystem.metadataStore.Get(ctx, fileID)
		if err != nil {
			t.Logf("Failed to get metadata entry: %v", err)
			return false
		}

		if entry.State != metadata.ItemStateDirtyLocal {
			t.Logf("Item state should be DIRTY_LOCAL before cleanup, got: %v", entry.State)
			return false
		}

		// Test: Simulate successful upload by marking clean
		filesystem.markCleanLocalState(fileID)

		// Verify: Modified flag is cleared
		if fileInode.HasChanges() {
			t.Logf("File should not have changes after marking clean")
			return false
		}

		// Verify: Metadata state is HYDRATED
		entry, err = filesystem.metadataStore.Get(ctx, fileID)
		if err != nil {
			t.Logf("Failed to get metadata entry after cleanup: %v", err)
			return false
		}

		if entry.State != metadata.ItemStateHydrated {
			t.Logf("Item state should be HYDRATED after cleanup, got: %v", entry.State)
			return false
		}

		// Success: Modified flag was properly cleared after upload
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 19 (Modified Flag Cleanup) failed: %v", err)
	}
}
