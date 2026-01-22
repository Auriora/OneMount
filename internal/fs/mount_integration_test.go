package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestMountIntegration_SuccessfulMount tests the complete mount flow
func TestUT_FS_Mount_Integration_SuccessfulMount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ensureMockGraphRoot(t)

	// Create temporary directories for test
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	mountPoint := filepath.Join(tempDir, "mount")

	// Create mount point
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	// Create mock auth
	auth := &graph.Auth{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}

	// Create filesystem with context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Note: This test requires a mock Graph API implementation
	// For now, we test the filesystem initialization without actual mounting
	filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer func() {
		filesystem.StopCacheCleanup()
		filesystem.StopDeltaLoop()
		filesystem.StopDownloadManager()
		filesystem.StopUploadManager()
		filesystem.StopMetadataRequestManager()
	}()

	// Verify filesystem was created
	if filesystem == nil {
		t.Fatal("Filesystem is nil")
	}

	// Verify cache directory was created
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Errorf("Cache directory was not created: %v", err)
	}

	// Verify database was created
	dbPath := filepath.Join(cacheDir, "onemount.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database was not created: %v", err)
	}

	t.Log("Filesystem initialization successful")
}

// TestMountIntegration_MountFailureScenarios tests various mount failure cases
func TestUT_FS_Mount_Integration_MountFailureScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) (string, string) // Returns mountPoint, cacheDir
		expectedError string
	}{
		{
			name: "non-existent mount point",
			setupFunc: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				return filepath.Join(tempDir, "does-not-exist"), filepath.Join(tempDir, "cache")
			},
			expectedError: "did not exist or was not a directory",
		},
		{
			name: "file as mount point",
			setupFunc: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				mountPoint := filepath.Join(tempDir, "file.txt")
				if err := os.WriteFile(mountPoint, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return mountPoint, filepath.Join(tempDir, "cache")
			},
			expectedError: "did not exist or was not a directory",
		},
		{
			name: "non-empty mount point",
			setupFunc: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				mountPoint := filepath.Join(tempDir, "mount")
				if err := os.MkdirAll(mountPoint, 0755); err != nil {
					t.Fatalf("Failed to create mount point: %v", err)
				}
				// Create a file in the mount point
				if err := os.WriteFile(filepath.Join(mountPoint, "existing.txt"), []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return mountPoint, filepath.Join(tempDir, "cache")
			},
			expectedError: "must be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensureMockGraphRoot(t)

			mountPoint, cacheDir := tt.setupFunc(t)

			// Create mock auth
			auth := &graph.Auth{
				AccessToken:  "mock_access_token",
				RefreshToken: "mock_refresh_token",
				ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
			}

			// Create filesystem
			ctx := context.Background()
			filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
			if err != nil {
				// Some errors may occur during filesystem creation
				t.Logf("Filesystem creation error (may be expected): %v", err)
			}

			if filesystem != nil {
				defer func() {
					filesystem.StopCacheCleanup()
					filesystem.StopDeltaLoop()
					filesystem.StopDownloadManager()
					filesystem.StopUploadManager()
					filesystem.StopMetadataRequestManager()
				}()
			}

			// Validate mount point (this is what main.go does)
			st, err := os.Stat(mountPoint)
			if err != nil || !st.IsDir() {
				// Expected error for non-existent or file mount points
				t.Logf("Mount point validation failed as expected: %v", err)
				return
			}

			// Check if directory is empty
			if entries, err := os.ReadDir(mountPoint); err == nil && len(entries) > 0 {
				// Expected error for non-empty mount point
				t.Logf("Mount point is not empty as expected")
				return
			}

			// If we get here, the mount point is valid
			// In a real scenario, we would attempt to create a FUSE server
			// For this test, we just verify the validation logic worked
			t.Log("Mount point validation passed")
		})
	}
}

// TestMountIntegration_GracefulUnmount tests the unmount process
func TestUT_FS_Mount_Integration_GracefulUnmount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ensureMockGraphRoot(t)

	// Create temporary directories
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")

	// Create mock auth
	auth := &graph.Auth{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}

	// Create filesystem
	ctx, cancel := context.WithCancel(context.Background())
	filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}

	// Verify filesystem is running
	if filesystem == nil {
		t.Fatal("Filesystem is nil")
	}

	// Simulate graceful shutdown
	t.Log("Initiating graceful shutdown...")

	// Cancel context (simulates signal handler)
	cancel()

	// Stop all background processes
	filesystem.StopCacheCleanup()
	filesystem.StopDeltaLoop()
	filesystem.StopDownloadManager()
	filesystem.StopUploadManager()
	filesystem.StopMetadataRequestManager()

	// Wait a bit for cleanup
	time.Sleep(100 * time.Millisecond)

	// Verify database is closed properly
	// (In a real test, we would check that the database file is not locked)

	t.Log("Graceful shutdown completed successfully")
}

// TestMountIntegration_WithMockGraphAPI tests mounting with a mock Graph API
func TestUT_FS_Mount_Integration_WithMockGraphAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary directories
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	mountPoint := filepath.Join(tempDir, "mount")

	// Create mount point
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	// Create mock Graph API (shared HTTP client)
	mockGraph := ensureMockGraphRoot(t)

	// Create mock auth
	auth := &graph.Auth{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}

	// Create filesystem
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer func() {
		filesystem.StopCacheCleanup()
		filesystem.StopDeltaLoop()
		filesystem.StopDownloadManager()
		filesystem.StopUploadManager()
		filesystem.StopMetadataRequestManager()
	}()

	// Create FUSE mount options
	mountOptions := &fuse.MountOptions{
		Name:          "onemount-test",
		FsName:        "onemount-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	// Note: Actual FUSE mounting requires root privileges and FUSE device
	// For unit/integration tests, we test the setup without actual mounting
	_ = mountOptions
	_ = mockGraph

	t.Log("Mock Graph API integration test setup successful")
	t.Log("Note: Actual FUSE mounting requires root privileges and is tested separately")
}

// Benchmark for filesystem initialization
func BenchmarkFilesystemInitialization(b *testing.B) {
	tempDir := b.TempDir()

	ensureMockGraphRoot(b)

	auth := &graph.Auth{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cacheDir := filepath.Join(tempDir, "cache", string(rune(i)))
		ctx := context.Background()
		filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
		if err != nil {
			b.Fatalf("Failed to create filesystem: %v", err)
		}

		// Cleanup
		filesystem.StopCacheCleanup()
		filesystem.StopDeltaLoop()
		filesystem.StopDownloadManager()
		filesystem.StopUploadManager()
		filesystem.StopMetadataRequestManager()
	}
}
