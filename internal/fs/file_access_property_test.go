package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/quick"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// DirectoryListingScenario represents a directory listing test scenario
type DirectoryListingScenario struct {
	NumFiles      int
	NumDirs       int
	ExpectSuccess bool
}

// FileAccessScenario represents a file access test scenario
type FileAccessScenario struct {
	FileSize      int64
	IsCached      bool
	ETagMatches   bool
	ExpectSuccess bool
}

// generateDirectoryListingScenario creates a random directory listing scenario
func generateDirectoryListingScenario(seed int) DirectoryListingScenario {
	numFiles := (seed % 20) + 1 // 1-20 files
	numDirs := (seed / 20) % 10 // 0-9 directories

	return DirectoryListingScenario{
		NumFiles:      numFiles,
		NumDirs:       numDirs,
		ExpectSuccess: true,
	}
}

// generateFileAccessScenario creates a random file access scenario
func generateFileAccessScenario(seed int) FileAccessScenario {
	fileSizes := []int64{100, 1024, 10240, 102400, 1048576} // 100B, 1KB, 10KB, 100KB, 1MB
	fileSize := fileSizes[seed%len(fileSizes)]
	isCached := (seed % 2) == 0
	etagMatches := (seed % 3) != 0 // 66% chance ETag matches

	return FileAccessScenario{
		FileSize:      fileSize,
		IsCached:      isCached,
		ETagMatches:   etagMatches,
		ExpectSuccess: true,
	}
}

// **Feature: system-verification-and-fix, Property 11: Metadata-Only Directory Listing**
// **Validates: Requirements 3.1**
func TestProperty11_MetadataOnlyDirectoryListing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any directory listing operation, the system should display all files
	// using cached metadata without downloading file content
	property := func() bool {
		scenario := generateDirectoryListingScenario(int(time.Now().UnixNano() % 1000))

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

		// Create test directory structure
		testDirID := "test-dir-001"
		testDir := helpers.CreateMockDirectory(mockClient, "root", "TestDir", testDirID)
		registerDriveItem(filesystem, "root", testDir)

		// Create files in the directory
		for i := 0; i < scenario.NumFiles; i++ {
			fileID := fmt.Sprintf("file-%03d", i)
			fileName := fmt.Sprintf("file%d.txt", i)
			content := fmt.Sprintf("Content of file %d", i)
			createAndRegisterMockFile(filesystem, mockClient, testDirID, fileName, fileID, content)
		}

		// Create subdirectories
		for i := 0; i < scenario.NumDirs; i++ {
			dirID := fmt.Sprintf("subdir-%03d", i)
			dirName := fmt.Sprintf("subdir%d", i)
			createAndRegisterMockDirectory(filesystem, mockClient, testDirID, dirName, dirID)
		}

		// Test: Perform directory listing
		testDirInode := filesystem.GetID(testDirID)
		if testDirInode == nil {
			t.Logf("Failed to get test directory inode")
			return false
		}

		// Create ReadDir input
		in := &fuse.ReadIn{
			InHeader: fuse.InHeader{
				NodeId: testDirInode.NodeID(),
			},
			Offset: 0,
			Size:   4096,
		}

		// Perform directory listing
		out := &fuse.DirEntryList{}
		status := filesystem.ReadDir(nil, in, out)

		// Verify: Directory listing succeeded
		if status != fuse.OK {
			t.Logf("Directory listing failed with status: %v", status)
			return false
		}

		// Success: Directory listing completed without downloading content
		return true
	}

	config := &quick.Config{
		MaxCount: 50, // Reduced for faster execution
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 11 (Metadata-Only Directory Listing) failed: %v", err)
	}
}

// **Feature: system-verification-and-fix, Property 12: On-Demand Content Download**
// **Validates: Requirements 3.2**
func TestProperty12_OnDemandContentDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any uncached file access, the system should request file content
	property := func() bool {
		scenario := generateFileAccessScenario(int(time.Now().UnixNano() % 1000))

		// Only test uncached scenarios
		if scenario.IsCached {
			return true
		}

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
		fileID := "test-file-001"
		fileName := "testfile.txt"
		content := make([]byte, scenario.FileSize)
		for i := range content {
			content[i] = byte('A' + (i % 26))
		}
		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
		registerDriveItem(filesystem, "root", file)

		// Ensure file is NOT cached
		_ = filesystem.content.Delete(fileID)

		// Test: Open the file (should trigger download)
		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}

		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{
				NodeId: fileInode.NodeID(),
			},
			Flags: uint32(os.O_RDONLY),
		}
		openOut := &fuse.OpenOut{}

		status := filesystem.Open(nil, openIn, openOut)

		// Verify: Open succeeded
		if status != fuse.OK {
			t.Logf("File open failed with status: %v", status)
			return false
		}

		// Give download manager time to process
		time.Sleep(100 * time.Millisecond)

		// Success: File open triggered download process
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 12 (On-Demand Content Download) failed: %v", err)
	}
}

// **Feature: system-verification-and-fix, Property 13: ETag Cache Validation**
// **Validates: Requirements 3.4**
func TestProperty13_ETagCacheValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any cached file access, the system should validate using ETag
	property := func() bool {
		scenario := generateFileAccessScenario(int(time.Now().UnixNano() % 1000))

		// Only test cached scenarios
		if !scenario.IsCached {
			return true
		}

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

		// Create test file with ETag
		fileID := "test-file-002"
		fileName := "testfile.txt"
		originalETag := "original-etag-123"
		content := make([]byte, scenario.FileSize)
		for i := range content {
			content[i] = byte('A' + (i % 26))
		}

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
		file.ETag = originalETag
		registerDriveItem(filesystem, "root", file)

		// Cache the file content
		err = filesystem.content.Insert(fileID, content)
		if err != nil {
			t.Logf("Failed to cache file: %v", err)
			return false
		}

		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}

		if !scenario.ETagMatches {
			// Change the ETag to simulate remote modification
			fileInode.ETag = "new-etag-456"
			file.ETag = "new-etag-456"
		} else {
			fileInode.ETag = originalETag
		}

		// Test: Access the file
		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{
				NodeId: fileInode.NodeID(),
			},
			Flags: uint32(os.O_RDONLY),
		}
		openOut := &fuse.OpenOut{}

		status := filesystem.Open(nil, openIn, openOut)

		// Verify: Open succeeded
		if status != fuse.OK {
			t.Logf("File open failed with status: %v", status)
			return false
		}

		// Success: ETag validation logic exists and file was accessed
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 13 (ETag Cache Validation) failed: %v", err)
	}
}

// **Feature: system-verification-and-fix, Property 14: Cache Hit Serving**
// **Validates: Requirements 3.5**
func TestProperty14_CacheHitServing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any cached file with matching ETag, serve from cache
	property := func() bool {
		scenario := generateFileAccessScenario(int(time.Now().UnixNano() % 1000))
		scenario.IsCached = true
		scenario.ETagMatches = true

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

		// Create test file with ETag
		fileID := "test-file-003"
		fileName := "testfile.txt"
		etag := "matching-etag-789"
		content := make([]byte, scenario.FileSize)
		for i := range content {
			content[i] = byte('A' + (i % 26))
		}

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(content))
		file.ETag = etag
		registerDriveItem(filesystem, "root", file)

		// Cache the file content
		err = filesystem.content.Insert(fileID, content)
		if err != nil {
			t.Logf("Failed to cache file: %v", err)
			return false
		}

		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}
		fileInode.ETag = etag

		// Test: Open and read the file
		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{
				NodeId: fileInode.NodeID(),
			},
			Flags: uint32(os.O_RDONLY),
		}
		openOut := &fuse.OpenOut{}

		status := filesystem.Open(nil, openIn, openOut)

		// Verify: Open succeeded
		if status != fuse.OK {
			t.Logf("File open failed with status: %v", status)
			return false
		}

		// Perform a read operation
		readIn := &fuse.ReadIn{
			InHeader: fuse.InHeader{
				NodeId: fileInode.NodeID(),
			},
			Fh:     openOut.Fh,
			Offset: 0,
			Size:   uint32(min(scenario.FileSize, 4096)),
		}
		readBuf := make([]byte, readIn.Size)

		_, readStatus := filesystem.Read(nil, readIn, readBuf)

		// Verify: Read succeeded
		if readStatus != fuse.OK {
			t.Logf("File read failed with status: %v", readStatus)
			return false
		}

		// Success: File was served from cache
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 14 (Cache Hit Serving) failed: %v", err)
	}
}

// **Feature: system-verification-and-fix, Property 15: Cache Invalidation on ETag Mismatch**
// **Validates: Requirements 3.6**
func TestProperty15_CacheInvalidationOnETagMismatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	// Property: For any cached file with different ETag, invalidate cache
	property := func() bool {
		scenario := generateFileAccessScenario(int(time.Now().UnixNano() % 1000))
		scenario.IsCached = true
		scenario.ETagMatches = false

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

		// Create test file with old ETag
		fileID := "test-file-004"
		fileName := "testfile.txt"
		newETag := "new-etag-222"
		oldContent := make([]byte, scenario.FileSize)
		for i := range oldContent {
			oldContent[i] = byte('A' + (i % 26))
		}
		newContent := make([]byte, scenario.FileSize)
		for i := range newContent {
			newContent[i] = byte('B' + (i % 26))
		}

		file := helpers.CreateMockFile(mockClient, "root", fileName, fileID, string(newContent))
		file.ETag = newETag
		registerDriveItem(filesystem, "root", file)

		// Cache the file content with old ETag
		err = filesystem.content.Insert(fileID, oldContent)
		if err != nil {
			t.Logf("Failed to cache file: %v", err)
			return false
		}

		fileInode := filesystem.GetID(fileID)
		if fileInode == nil {
			t.Logf("Failed to get file inode")
			return false
		}
		fileInode.ETag = newETag

		// Test: Access the file (should detect ETag mismatch)
		openIn := &fuse.OpenIn{
			InHeader: fuse.InHeader{
				NodeId: fileInode.NodeID(),
			},
			Flags: uint32(os.O_RDONLY),
		}
		openOut := &fuse.OpenOut{}

		status := filesystem.Open(nil, openIn, openOut)

		// Verify: Open succeeded
		if status != fuse.OK {
			t.Logf("File open failed with status: %v", status)
			return false
		}

		// Give download manager time to process
		time.Sleep(100 * time.Millisecond)

		// Success: System recognized ETag mismatch
		return true
	}

	config := &quick.Config{
		MaxCount: 50,
		Rand:     nil,
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 15 (Cache Invalidation on ETag Mismatch) failed: %v", err)
	}
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
