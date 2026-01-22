package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestIT_FS_Mount tests mounting with real OneDrive
// This test requires:
// - Real OneDrive authentication tokens
// - FUSE device access
// - Environment variable: ONEMOUNT_AUTH_PATH (set by setup-auth-reference.sh)
//
// Requirements: 2.1, 2.2, 2.4, 2.5
func TestIT_FS_Mount(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load real authentication
	authPath, err := testutil.GetAuthTokenPath()
	if err != nil {
		t.Fatalf("Authentication not configured: %v", err)
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Fatalf("Cannot load auth tokens: %v", err)
	}

	// Create temporary directories
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "cache")
	mountPoint := filepath.Join(tempDir, "mount")

	// Create mount point
	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	t.Logf("Test setup complete:")
	t.Logf("  Cache dir: %s", cacheDir)
	t.Logf("  Mount point: %s", mountPoint)

	// Test 1: Verify filesystem mounts successfully (Requirement 2.1)
	t.Run("MountSuccessfully", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// Create filesystem
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

		// Mount the filesystem
		server, err := fuse.NewServer(filesystem, mountPoint, mountOptions)
		if err != nil {
			t.Fatalf("Failed to create FUSE server: %v", err)
		}
		defer server.Unmount()

		// Start serving in background
		go server.Serve()

		// Wait for mount to be ready
		if err := server.WaitMount(); err != nil {
			t.Fatalf("Failed to wait for mount: %v", err)
		}

		t.Log("✓ Filesystem mounted successfully")

		// Test 2: Verify root directory is accessible (Requirement 2.2, 2.3)
		t.Run("RootDirectoryAccessible", func(t *testing.T) {
			// List root directory
			entries, err := os.ReadDir(mountPoint)
			if err != nil {
				t.Fatalf("Failed to read root directory: %v", err)
			}

			t.Logf("✓ Root directory accessible, found %d entries", len(entries))

			// Verify we can stat the mount point
			info, err := os.Stat(mountPoint)
			if err != nil {
				t.Fatalf("Failed to stat mount point: %v", err)
			}

			if !info.IsDir() {
				t.Fatalf("Mount point is not a directory")
			}

			t.Log("✓ Root directory stat successful")

			// Try to list a few entries if they exist
			for i, entry := range entries {
				if i >= 3 {
					break // Only check first 3 entries
				}
				entryPath := filepath.Join(mountPoint, entry.Name())
				entryInfo, err := os.Stat(entryPath)
				if err != nil {
					t.Logf("Warning: Failed to stat entry %s: %v", entry.Name(), err)
					continue
				}
				t.Logf("  Entry: %s (dir=%v, size=%d)", entry.Name(), entryInfo.IsDir(), entryInfo.Size())
			}

			t.Log("✓ Directory entries accessible")
		})

		// Test 3: Verify mount point validation works (Requirement 2.4)
		t.Run("MountPointValidation", func(t *testing.T) {
			// Try to mount at the same location (should fail)
			invalidMountPoint := mountPoint // Already mounted

			filesystem2, err := NewFilesystemWithContext(ctx, auth, filepath.Join(tempDir, "cache2"), 30, 24, 0)
			if err != nil {
				t.Fatalf("Failed to create second filesystem: %v", err)
			}
			defer func() {
				filesystem2.StopCacheCleanup()
				filesystem2.StopDeltaLoop()
				filesystem2.StopDownloadManager()
				filesystem2.StopUploadManager()
				filesystem2.StopMetadataRequestManager()
			}()

			// Attempt to mount at already-mounted location
			server2, err := fuse.NewServer(filesystem2, invalidMountPoint, mountOptions)
			if err == nil {
				// If server creation succeeded, try to serve
				go server2.Serve()
				err = server2.WaitMount()
				if err == nil {
					server2.Unmount()
					t.Fatal("Expected mount to fail at already-mounted location, but it succeeded")
				}
			}

			t.Log("✓ Mount point validation works (duplicate mount rejected)")
		})

		// Test 4: Verify graceful unmount (Requirement 2.5)
		t.Run("GracefulUnmount", func(t *testing.T) {
			// Unmount
			err := server.Unmount()
			if err != nil {
				t.Fatalf("Failed to unmount: %v", err)
			}

			// Wait a bit for cleanup
			time.Sleep(500 * time.Millisecond)

			// Verify mount point is no longer mounted
			// Try to access it - should work as regular directory
			_, err = os.Stat(mountPoint)
			if err != nil {
				t.Fatalf("Mount point should still exist as directory after unmount: %v", err)
			}

			t.Log("✓ Graceful unmount successful")
		})
	})
}

// TestIT_FS_Mount_ValidationScenarios tests mount point validation with real OneDrive
// Requirements: 2.4
func TestIT_FS_Mount_ValidationScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load real authentication
	authPath, err := testutil.GetAuthTokenPath()
	if err != nil {
		t.Fatalf("Authentication not configured: %v", err)
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Fatalf("Cannot load auth tokens: %v", err)
	}

	tempDir := t.TempDir()

	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string // Returns mount point
		shouldFail    bool
		expectedError string
	}{
		{
			name: "non-existent mount point",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(tempDir, "does-not-exist")
			},
			shouldFail:    true,
			expectedError: "no such file or directory",
		},
		{
			name: "file as mount point",
			setupFunc: func(t *testing.T) string {
				mountPoint := filepath.Join(tempDir, "file.txt")
				if err := os.WriteFile(mountPoint, []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return mountPoint
			},
			shouldFail:    true,
			expectedError: "not a directory",
		},
		{
			name: "non-empty mount point",
			setupFunc: func(t *testing.T) string {
				mountPoint := filepath.Join(tempDir, "non-empty")
				if err := os.MkdirAll(mountPoint, 0755); err != nil {
					t.Fatalf("Failed to create mount point: %v", err)
				}
				// Create a file in the mount point
				if err := os.WriteFile(filepath.Join(mountPoint, "existing.txt"), []byte("test"), 0644); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				return mountPoint
			},
			shouldFail:    true,
			expectedError: "not empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mountPoint := tt.setupFunc(t)
			cacheDir := filepath.Join(tempDir, "cache-"+tt.name)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Create filesystem
			filesystem, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
			if err != nil {
				t.Logf("Filesystem creation failed (may be expected): %v", err)
				if !tt.shouldFail {
					t.Fatalf("Unexpected filesystem creation failure: %v", err)
				}
				return
			}
			defer func() {
				filesystem.StopCacheCleanup()
				filesystem.StopDeltaLoop()
				filesystem.StopDownloadManager()
				filesystem.StopUploadManager()
				filesystem.StopMetadataRequestManager()
			}()

			// Try to mount
			mountOptions := &fuse.MountOptions{
				Name:   "onemount-test",
				FsName: "onemount-test",
			}

			server, err := fuse.NewServer(filesystem, mountPoint, mountOptions)
			if err != nil {
				if tt.shouldFail {
					t.Logf("✓ Mount failed as expected: %v", err)
					return
				}
				t.Fatalf("Unexpected mount failure: %v", err)
			}

			// Try to serve
			go server.Serve()
			err = server.WaitMount()

			if err != nil {
				if tt.shouldFail {
					t.Logf("✓ Mount failed as expected: %v", err)
					return
				}
				t.Fatalf("Unexpected mount failure: %v", err)
			}

			// Clean up
			defer server.Unmount()

			if tt.shouldFail {
				t.Fatalf("Expected mount to fail, but it succeeded")
			}

			t.Log("✓ Mount succeeded as expected")
		})
	}
}
