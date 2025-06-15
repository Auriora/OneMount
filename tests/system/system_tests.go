// Package system provides comprehensive system testing utilities for OneMount
package system

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/fs"
	"github.com/auriora/onemount/pkg/graph"
	"github.com/auriora/onemount/pkg/logging"
	"github.com/auriora/onemount/pkg/testutil"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// SystemTestSuite provides comprehensive system testing for OneMount
type SystemTestSuite struct {
	t          *testing.T
	auth       *graph.Auth
	filesystem *fs.Filesystem
	server     *fuse.Server
	mountPoint string
	testDir    string
	cleanup    []func() error
	mu         sync.Mutex
}

// NewSystemTestSuite creates a new system test suite
func NewSystemTestSuite(t *testing.T) (*SystemTestSuite, error) {
	// Load real authentication tokens
	auth, err := graph.LoadAuthTokens(testutil.AuthTokensPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load auth tokens from %s: %w", testutil.AuthTokensPath, err)
	}

	// Refresh tokens if needed
	ctx := context.Background()
	if err := auth.Refresh(ctx); err != nil {
		logging.Warn().Err(err).Msg("Failed to refresh auth tokens, continuing with existing tokens")
	}

	// Create unique mount point and test directory for this test instance
	uniqueID := fmt.Sprintf("%d_%d", os.Getpid(), time.Now().UnixNano())
	uniqueMountPoint := filepath.Join(testutil.SystemTestDataDir, "mount", uniqueID)
	uniqueTestDir := filepath.Join(uniqueMountPoint, "system-test-"+uniqueID)

	suite := &SystemTestSuite{
		t:          t,
		auth:       auth,
		mountPoint: uniqueMountPoint,
		testDir:    uniqueTestDir,
		cleanup:    make([]func() error, 0),
	}

	return suite, nil
}

// Setup initializes the system test environment
func (s *SystemTestSuite) Setup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure test directories exist
	if err := testutil.EnsureDirectoriesExist(); err != nil {
		return fmt.Errorf("failed to ensure test directories exist: %w", err)
	}

	// Create system test specific directories
	for _, dir := range []string{testutil.SystemTestDataDir, filepath.Dir(testutil.SystemTestLogPath)} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Clean up any existing mount point first
	// Try to unmount if it's still mounted
	if _, err := os.Stat(s.mountPoint); err == nil {
		// Try to unmount using fusermount3
		if err := exec.Command("fusermount3", "-uz", s.mountPoint).Run(); err != nil {
			logging.Debug().Err(err).Msg("fusermount3 unmount failed, trying fusermount")
		}
		// Also try fusermount (older systems)
		if err := exec.Command("fusermount", "-uz", s.mountPoint).Run(); err != nil {
			logging.Debug().Err(err).Msg("fusermount unmount failed")
		}
		// Wait a moment for unmount to complete
		time.Sleep(500 * time.Millisecond)
	}

	if err := os.RemoveAll(s.mountPoint); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean up existing mount point %s: %w", s.mountPoint, err)
	}

	// Create mount point
	if err := os.MkdirAll(s.mountPoint, 0755); err != nil {
		return fmt.Errorf("failed to create mount point %s: %w", s.mountPoint, err)
	}

	// Create unique cache directory per test to avoid database conflicts
	// Use process ID and timestamp for uniqueness
	uniqueID := fmt.Sprintf("%d_%d", os.Getpid(), time.Now().UnixNano())
	cacheDir := filepath.Join(testutil.SystemTestDataDir, "cache", uniqueID)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	// Update auth token path to use the test-specific cache directory
	// This ensures token refresh operations use the correct path
	testAuthPath := filepath.Join(cacheDir, "auth_tokens.json")
	s.auth.Path = testAuthPath

	// Save auth tokens to the test-specific location
	if err := s.auth.ToFile(testAuthPath); err != nil {
		return fmt.Errorf("failed to save auth tokens to test location: %w", err)
	}

	// Create filesystem with unique cache directory
	filesystem, err := fs.NewFilesystem(s.auth, cacheDir, 30) // 30 second cache TTL
	if err != nil {
		return fmt.Errorf("failed to create filesystem: %w", err)
	}
	s.filesystem = filesystem

	// Create mount options
	mountOptions := &fuse.MountOptions{
		Name:          "onemount-system-test",
		FsName:        "onemount-system-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	// Create FUSE server
	server, err := fuse.NewServer(s.filesystem, s.mountPoint, mountOptions)
	if err != nil {
		return fmt.Errorf("failed to create FUSE server: %w", err)
	}
	s.server = server

	// Start the server in a goroutine
	go func() {
		s.server.Serve()
	}()

	// Add cleanup for unmounting
	s.cleanup = append(s.cleanup, func() error {
		return s.server.Unmount()
	})

	// Wait for mount to be ready
	time.Sleep(2 * time.Second)

	// Create test directory on OneDrive
	if err := os.MkdirAll(s.testDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory %s: %w", s.testDir, err)
	}

	// Wait for directory creation to sync
	time.Sleep(1 * time.Second)

	return nil
}

// Cleanup cleans up the system test environment
func (s *SystemTestSuite) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errors []error

	// Force unmount if still mounted
	if s.server != nil {
		if err := s.server.Unmount(); err != nil {
			errors = append(errors, fmt.Errorf("failed to unmount server: %w", err))
		}
		// Wait for unmount to complete
		time.Sleep(1 * time.Second)
	}

	// Force unmount using system commands as backup
	if s.mountPoint != "" {
		// Try fusermount3 first
		if err := exec.Command("fusermount3", "-uz", s.mountPoint).Run(); err != nil {
			logging.Debug().Err(err).Msg("fusermount3 unmount failed during cleanup")
		}
		// Try fusermount as fallback
		if err := exec.Command("fusermount", "-uz", s.mountPoint).Run(); err != nil {
			logging.Debug().Err(err).Msg("fusermount unmount failed during cleanup")
		}
		// Wait for unmount to complete
		time.Sleep(500 * time.Millisecond)
	}

	// Run cleanup functions in reverse order
	for i := len(s.cleanup) - 1; i >= 0; i-- {
		if err := s.cleanup[i](); err != nil {
			errors = append(errors, err)
		}
	}

	// Clean up test directory on OneDrive
	if s.testDir != "" {
		if err := os.RemoveAll(s.testDir); err != nil {
			errors = append(errors, fmt.Errorf("failed to remove test directory %s: %w", s.testDir, err))
		}
	}

	// Clean up mount point directory
	if s.mountPoint != "" {
		if err := os.RemoveAll(s.mountPoint); err != nil && !os.IsNotExist(err) {
			errors = append(errors, fmt.Errorf("failed to remove mount point %s: %w", s.mountPoint, err))
		}
	}

	// Wait for cleanup to sync
	time.Sleep(2 * time.Second)

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// TestBasicFileOperations tests basic file create, read, write, delete operations
func (s *SystemTestSuite) TestBasicFileOperations() error {
	testFile := filepath.Join(s.testDir, "basic_test.txt")
	testContent := "Hello, OneMount System Test!"

	// Clean up any existing file first to avoid conflicts
	if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean up existing test file: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Test file creation and writing
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		return fmt.Errorf("failed to create test file: %w", err)
	}

	// Wait for upload
	time.Sleep(3 * time.Second)

	// Test file reading
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file: %w", err)
	}

	if string(content) != testContent {
		return fmt.Errorf("file content mismatch: expected %q, got %q", testContent, string(content))
	}

	// Test file modification
	newContent := testContent + " - Modified"
	if err := os.WriteFile(testFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to modify test file: %w", err)
	}

	// Wait for upload
	time.Sleep(3 * time.Second)

	// Verify modification
	content, err = os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read modified test file: %w", err)
	}

	if string(content) != newContent {
		return fmt.Errorf("modified file content mismatch: expected %q, got %q", newContent, string(content))
	}

	// Test file deletion
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to delete test file: %w", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	// Verify deletion
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		return fmt.Errorf("file should not exist after deletion")
	}

	return nil
}

// TestDirectoryOperations tests directory creation, listing, and deletion
func (s *SystemTestSuite) TestDirectoryOperations() error {
	testSubDir := filepath.Join(s.testDir, "subdir_test")

	// Test directory creation
	if err := os.MkdirAll(testSubDir, 0755); err != nil {
		return fmt.Errorf("failed to create test subdirectory: %w", err)
	}

	// Wait for creation to sync
	time.Sleep(2 * time.Second)

	// Test directory listing
	entries, err := os.ReadDir(s.testDir)
	if err != nil {
		return fmt.Errorf("failed to list test directory: %w", err)
	}

	found := false
	for _, entry := range entries {
		if entry.Name() == "subdir_test" && entry.IsDir() {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("created subdirectory not found in directory listing")
	}

	// Create a file in the subdirectory
	testFile := filepath.Join(testSubDir, "nested_file.txt")
	if err := os.WriteFile(testFile, []byte("nested content"), 0644); err != nil {
		return fmt.Errorf("failed to create nested file: %w", err)
	}

	// Wait for upload
	time.Sleep(3 * time.Second)

	// Test nested directory deletion
	if err := os.RemoveAll(testSubDir); err != nil {
		return fmt.Errorf("failed to delete test subdirectory: %w", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	// Verify deletion
	if _, err := os.Stat(testSubDir); !os.IsNotExist(err) {
		return fmt.Errorf("subdirectory should not exist after deletion")
	}

	return nil
}

// TestLargeFileOperations tests operations with large files
func (s *SystemTestSuite) TestLargeFileOperations() error {
	testFile := filepath.Join(s.testDir, "large_file_test.bin")

	// Clean up any existing file first to avoid conflicts
	if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean up existing large test file: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Create a 10MB test file using streaming to avoid memory issues
	const fileSize = 10 * 1024 * 1024 // 10MB
	const chunkSize = 1024 * 1024     // 1MB chunks to reduce memory usage

	// Create file and write in chunks to avoid allocating 10MB in memory
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create large test file: %w", err)
	}
	defer file.Close()

	// Create a reusable 1MB chunk
	chunk := make([]byte, chunkSize)
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	// Write file in chunks
	bytesWritten := 0
	for bytesWritten < fileSize {
		remainingBytes := fileSize - bytesWritten
		writeSize := chunkSize
		if remainingBytes < chunkSize {
			writeSize = remainingBytes
		}

		if _, err := file.Write(chunk[:writeSize]); err != nil {
			return fmt.Errorf("failed to write chunk to large test file: %w", err)
		}
		bytesWritten += writeSize
	}

	// Ensure data is written to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync large test file: %w", err)
	}
	file.Close()

	// Wait for upload (longer for large files)
	time.Sleep(10 * time.Second)

	// Test large file reading
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read large test file: %w", err)
	}

	if len(content) != fileSize {
		return fmt.Errorf("large file size mismatch: expected %d, got %d", fileSize, len(content))
	}

	// Verify content integrity
	for i, b := range content {
		if b != byte(i%256) {
			return fmt.Errorf("large file content corruption at byte %d: expected %d, got %d", i, i%256, b)
		}
	}

	// Clean up large file
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to delete large test file: %w", err)
	}

	// Wait for deletion to sync
	time.Sleep(5 * time.Second)

	return nil
}

// TestSpecialCharacterFiles tests files with special characters in names
func (s *SystemTestSuite) TestSpecialCharacterFiles() error {
	specialNames := []string{
		"file with spaces.txt",
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file.with.dots.txt",
		"file(with)parentheses.txt",
		"file[with]brackets.txt",
		"file{with}braces.txt",
		"file'with'quotes.txt",
		"file&with&ampersands.txt",
		"file%with%percent.txt",
		"file#with#hash.txt",
		"file@with@at.txt",
		"file+with+plus.txt",
		"file=with=equals.txt",
	}

	// Clean up any existing files first to avoid conflicts
	for _, name := range specialNames {
		testFile := filepath.Join(s.testDir, name)
		if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to clean up existing special character file %q: %w", name, err)
		}
	}
	time.Sleep(2 * time.Second)

	for _, name := range specialNames {
		testFile := filepath.Join(s.testDir, name)
		testContent := fmt.Sprintf("Content for file: %s", name)

		// Create file with special characters
		if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
			return fmt.Errorf("failed to create file with special name %q: %w", name, err)
		}

		// Wait for upload
		time.Sleep(2 * time.Second)

		// Read and verify
		content, err := os.ReadFile(testFile)
		if err != nil {
			return fmt.Errorf("failed to read file with special name %q: %w", name, err)
		}

		if string(content) != testContent {
			return fmt.Errorf("content mismatch for file %q: expected %q, got %q", name, testContent, string(content))
		}

		// Clean up
		if err := os.Remove(testFile); err != nil {
			return fmt.Errorf("failed to delete file with special name %q: %w", name, err)
		}
	}

	// Wait for all deletions to sync
	time.Sleep(3 * time.Second)

	return nil
}

// TestConcurrentOperations tests concurrent file operations
func (s *SystemTestSuite) TestConcurrentOperations() error {
	const numFiles = 10

	var wg sync.WaitGroup
	errors := make(chan error, numFiles)

	// Create files concurrently
	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(fileNum int) {
			defer wg.Done()

			fileName := fmt.Sprintf("concurrent_file_%d.txt", fileNum)
			testFile := filepath.Join(s.testDir, fileName)
			testContent := fmt.Sprintf("Concurrent content for file %d", fileNum)

			// Create file
			if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
				errors <- fmt.Errorf("failed to create concurrent file %d: %w", fileNum, err)
				return
			}

			// Wait a bit
			time.Sleep(1 * time.Second)

			// Read and verify
			content, err := os.ReadFile(testFile)
			if err != nil {
				errors <- fmt.Errorf("failed to read concurrent file %d: %w", fileNum, err)
				return
			}

			if string(content) != testContent {
				errors <- fmt.Errorf("content mismatch for concurrent file %d: expected %q, got %q", fileNum, testContent, string(content))
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	// Wait for all uploads to complete
	time.Sleep(5 * time.Second)

	// Clean up all files
	for i := 0; i < numFiles; i++ {
		fileName := fmt.Sprintf("concurrent_file_%d.txt", i)
		testFile := filepath.Join(s.testDir, fileName)
		if err := os.Remove(testFile); err != nil {
			return fmt.Errorf("failed to delete concurrent file %d: %w", i, err)
		}
	}

	// Wait for deletions to sync
	time.Sleep(3 * time.Second)

	return nil
}

// TestFilePermissions tests file permission handling
func (s *SystemTestSuite) TestFilePermissions() error {
	testFile := filepath.Join(s.testDir, "permissions_test.txt")
	testContent := "Permission test content"

	// Clean up any existing file first to avoid conflicts
	if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean up existing permissions test file: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Create file with specific permissions
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		return fmt.Errorf("failed to create permissions test file: %w", err)
	}

	// Wait for upload
	time.Sleep(3 * time.Second)

	// Check file permissions
	info, err := os.Stat(testFile)
	if err != nil {
		return fmt.Errorf("failed to stat permissions test file: %w", err)
	}

	// Verify file mode (OneDrive may not preserve exact permissions, but should be readable)
	if !info.Mode().IsRegular() {
		return fmt.Errorf("file should be a regular file")
	}

	// Test file is readable
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read permissions test file: %w", err)
	}

	if string(content) != testContent {
		return fmt.Errorf("permissions test file content mismatch: expected %q, got %q", testContent, string(content))
	}

	// Clean up
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to delete permissions test file: %w", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	return nil
}

// TestStreamingOperations tests streaming read/write operations
func (s *SystemTestSuite) TestStreamingOperations() error {
	testFile := filepath.Join(s.testDir, "streaming_test.txt")

	// Clean up any existing file first to avoid conflicts
	if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean up existing streaming test file: %w", err)
	}
	time.Sleep(1 * time.Second)

	// Test streaming write
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create streaming test file: %w", err)
	}

	// Write data in chunks
	chunks := []string{
		"First chunk of data\n",
		"Second chunk of data\n",
		"Third chunk of data\n",
		"Fourth chunk of data\n",
		"Final chunk of data\n",
	}

	// Write all chunks without syncing between them to avoid race conditions
	// where multiple upload sessions are created for the same file
	for _, chunk := range chunks {
		if _, err := file.WriteString(chunk); err != nil {
			file.Close()
			return fmt.Errorf("failed to write chunk to streaming test file: %w", err)
		}
		// Small delay to simulate streaming behavior without triggering sync
		time.Sleep(100 * time.Millisecond)
	}

	// Sync only once at the end to trigger upload
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("failed to sync streaming test file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close streaming test file: %w", err)
	}

	// Wait for upload to complete
	time.Sleep(8 * time.Second)

	// Test streaming read
	file, err = os.Open(testFile)
	if err != nil {
		return fmt.Errorf("failed to open streaming test file for reading: %w", err)
	}
	defer file.Close()

	// Read data in chunks
	buffer := make([]byte, 1024)
	var readContent strings.Builder

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read chunk from streaming test file: %w", err)
		}
		readContent.Write(buffer[:n])
	}

	expectedContent := strings.Join(chunks, "")
	if readContent.String() != expectedContent {
		return fmt.Errorf("streaming test file content mismatch: expected %q, got %q", expectedContent, readContent.String())
	}

	// Clean up
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to delete streaming test file: %w", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	return nil
}

// TestPerformance tests upload/download performance for a given file size
func (s *SystemTestSuite) TestPerformance(sizeName string, sizeBytes int) error {
	testFile := filepath.Join(s.testDir, fmt.Sprintf("performance_test_%s.bin", sizeName))

	// Create test data
	data := make([]byte, sizeBytes)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Measure upload time
	uploadStart := time.Now()
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		return fmt.Errorf("failed to create performance test file: %w", err)
	}

	// Wait for upload to complete
	uploadWaitStart := time.Now()
	for {
		if time.Since(uploadWaitStart) > 60*time.Second {
			return fmt.Errorf("upload timeout for %s file", sizeName)
		}

		// Check if file is fully uploaded by trying to read it
		if content, err := os.ReadFile(testFile); err == nil && len(content) == sizeBytes {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	uploadDuration := time.Since(uploadStart)

	// Measure download time (clear cache first by unmounting and remounting)
	// For simplicity, we'll just read the file again
	downloadStart := time.Now()
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read performance test file: %w", err)
	}
	downloadDuration := time.Since(downloadStart)

	// Verify content integrity
	if len(content) != sizeBytes {
		return fmt.Errorf("performance test file size mismatch: expected %d, got %d", sizeBytes, len(content))
	}

	// Calculate speeds (bytes per second)
	uploadSpeed := float64(sizeBytes) / uploadDuration.Seconds()
	downloadSpeed := float64(sizeBytes) / downloadDuration.Seconds()

	s.t.Logf("Performance results for %s (%d bytes):", sizeName, sizeBytes)
	s.t.Logf("  Upload: %v (%.2f KB/s)", uploadDuration, uploadSpeed/1024)
	s.t.Logf("  Download: %v (%.2f KB/s)", downloadDuration, downloadSpeed/1024)

	// Clean up
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to delete performance test file: %w", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	return nil
}

// TestInvalidFileNames tests handling of invalid file names
func (s *SystemTestSuite) TestInvalidFileNames() error {
	invalidNames := []string{
		"",             // Empty name
		".",            // Current directory
		"..",           // Parent directory
		"con",          // Windows reserved name
		"prn",          // Windows reserved name
		"aux",          // Windows reserved name
		"nul",          // Windows reserved name
		"file\x00name", // Null character
		"file\tname",   // Tab character
		"file\nname",   // Newline character
		"file\rname",   // Carriage return
		"file|name",    // Pipe character
		"file<name",    // Less than
		"file>name",    // Greater than
		"file\"name",   // Quote character
		"file*name",    // Asterisk
		"file?name",    // Question mark
		"file:name",    // Colon
		"file\\name",   // Backslash
		"file/name",    // Forward slash
	}

	for _, name := range invalidNames {
		testFile := filepath.Join(s.testDir, name)

		// Try to create file with invalid name
		err := os.WriteFile(testFile, []byte("test content"), 0644)

		// We expect this to either fail or be handled gracefully
		if err != nil {
			// This is expected for truly invalid names
			s.t.Logf("Invalid file name %q correctly rejected: %v", name, err)
		} else {
			// If it succeeded, clean up
			s.t.Logf("Invalid file name %q was accepted (may be normalized)", name)
			if err := os.Remove(testFile); err != nil {
				s.t.Logf("Warning: failed to clean up invalid filename test file %q: %v", name, err)
			}
		}
	}

	return nil
}

// TestAuthenticationRefresh tests authentication token refresh
func (s *SystemTestSuite) TestAuthenticationRefresh() error {
	// Force a token refresh by calling the refresh method
	ctx := context.Background()
	if err := s.auth.Refresh(ctx); err != nil {
		return fmt.Errorf("failed to refresh authentication tokens: %w", err)
	}

	// Test that operations still work after refresh
	testFile := filepath.Join(s.testDir, "auth_refresh_test.txt")
	testContent := "Authentication refresh test content"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		return fmt.Errorf("failed to create file after auth refresh: %w", err)
	}

	// Wait for upload
	time.Sleep(3 * time.Second)

	// Verify file content
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read file after auth refresh: %w", err)
	}

	if string(content) != testContent {
		return fmt.Errorf("file content mismatch after auth refresh: expected %q, got %q", testContent, string(content))
	}

	// Clean up
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to delete file after auth refresh: %w", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	return nil
}

// TestDiskSpaceHandling tests handling of disk space issues
func (s *SystemTestSuite) TestDiskSpaceHandling() error {
	// This test is more about ensuring the system doesn't crash
	// when dealing with large files or potential disk space issues

	// Try to create a moderately large file (50MB)
	testFile := filepath.Join(s.testDir, "disk_space_test.bin")
	const fileSize = 50 * 1024 * 1024 // 50MB

	// Create the file in chunks to avoid memory issues
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("failed to create disk space test file: %w", err)
	}
	defer file.Close()

	chunkSize := 1024 * 1024 // 1MB chunks
	chunk := make([]byte, chunkSize)
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	bytesWritten := 0
	for bytesWritten < fileSize {
		remainingBytes := fileSize - bytesWritten
		writeSize := chunkSize
		if remainingBytes < chunkSize {
			writeSize = remainingBytes
		}

		n, err := file.Write(chunk[:writeSize])
		if err != nil {
			// This might fail due to disk space or other issues
			s.t.Logf("Disk space test stopped at %d bytes: %v", bytesWritten, err)
			break
		}
		bytesWritten += n

		// Sync periodically
		if bytesWritten%(10*1024*1024) == 0 {
			if err := file.Sync(); err != nil {
				s.t.Logf("Warning: failed to sync disk space test file: %v", err)
			}
		}
	}

	if err := file.Close(); err != nil {
		s.t.Logf("Warning: failed to close disk space test file: %v", err)
	}

	// Wait for upload (or partial upload)
	time.Sleep(10 * time.Second)

	// Clean up
	if err := os.Remove(testFile); err != nil {
		s.t.Logf("Warning: failed to delete disk space test file: %v", err)
	}

	// Wait for deletion to sync
	time.Sleep(5 * time.Second)

	return nil
}

// TestMountUnmountCycle tests mount/unmount operations
func (s *SystemTestSuite) TestMountUnmountCycle() error {
	// Create a test file before unmounting
	testFile := filepath.Join(s.testDir, "mount_cycle_test.txt")
	testContent := "Mount cycle test content"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		return fmt.Errorf("failed to create test file before unmount: %w", err)
	}

	// Wait for upload
	time.Sleep(3 * time.Second)

	// Unmount the filesystem
	if err := s.server.Unmount(); err != nil {
		return fmt.Errorf("failed to unmount filesystem: %w", err)
	}

	// Wait for unmount to complete
	time.Sleep(2 * time.Second)

	// Verify mount point is no longer accessible
	if _, err := os.Stat(s.testDir); err == nil {
		return fmt.Errorf("test directory should not be accessible after unmount")
	}

	// Remount the filesystem
	// Create a new FUSE server for remounting
	mountOptions := &fuse.MountOptions{
		Name:          "onemount-system-test",
		FsName:        "onemount-system-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(s.filesystem, s.mountPoint, mountOptions)
	if err != nil {
		return fmt.Errorf("failed to create FUSE server for remount: %w", err)
	}
	s.server = server

	// Start the server in a goroutine
	go func() {
		s.server.Serve()
	}()

	// Wait for mount to be ready
	time.Sleep(3 * time.Second)

	// Verify test file is still present
	content, err := os.ReadFile(testFile)
	if err != nil {
		return fmt.Errorf("failed to read test file after remount: %w", err)
	}

	if string(content) != testContent {
		return fmt.Errorf("test file content mismatch after remount: expected %q, got %q", testContent, string(content))
	}

	// Clean up
	if err := os.Remove(testFile); err != nil {
		return fmt.Errorf("failed to delete test file after remount: %w", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	return nil
}

// TestHighLoadOperations tests system behavior under high load
func (s *SystemTestSuite) TestHighLoadOperations() error {
	const numFiles = 50

	var wg sync.WaitGroup
	errors := make(chan error, numFiles)

	// Create many files concurrently
	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(fileNum int) {
			defer wg.Done()

			fileName := fmt.Sprintf("high_load_file_%03d.txt", fileNum)
			testFile := filepath.Join(s.testDir, fileName)
			testContent := fmt.Sprintf("High load test content for file %d\nThis is a longer content to make the file larger.\nLine 3\nLine 4\nLine 5", fileNum)

			// Create file
			if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
				errors <- fmt.Errorf("failed to create high load file %d: %w", fileNum, err)
				return
			}

			// Wait a bit
			time.Sleep(100 * time.Millisecond)

			// Read and verify
			content, err := os.ReadFile(testFile)
			if err != nil {
				errors <- fmt.Errorf("failed to read high load file %d: %w", fileNum, err)
				return
			}

			if string(content) != testContent {
				errors <- fmt.Errorf("content mismatch for high load file %d", fileNum)
				return
			}

			// Modify the file
			modifiedContent := testContent + "\nModified line"
			if err := os.WriteFile(testFile, []byte(modifiedContent), 0644); err != nil {
				errors <- fmt.Errorf("failed to modify high load file %d: %w", fileNum, err)
				return
			}

			// Wait a bit more
			time.Sleep(100 * time.Millisecond)

			// Read and verify modification
			content, err = os.ReadFile(testFile)
			if err != nil {
				errors <- fmt.Errorf("failed to read modified high load file %d: %w", fileNum, err)
				return
			}

			if string(content) != modifiedContent {
				errors <- fmt.Errorf("modified content mismatch for high load file %d", fileNum)
				return
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	// Wait for all uploads to complete
	time.Sleep(10 * time.Second)

	// Clean up all files
	for i := 0; i < numFiles; i++ {
		fileName := fmt.Sprintf("high_load_file_%03d.txt", i)
		testFile := filepath.Join(s.testDir, fileName)
		if err := os.Remove(testFile); err != nil {
			s.t.Logf("Warning: failed to delete high load file %d: %v", i, err)
		}
	}

	// Wait for deletions to sync
	time.Sleep(5 * time.Second)

	return nil
}
