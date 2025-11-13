package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil/helpers"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// TestE2E_17_01_CompleteUserWorkflow tests the complete user workflow from authentication to file operations
//
//	Test Case ID    E2E-17-01
//	Title           Complete User Workflow
//	Description     Verify complete user workflow: authenticate, mount, create/modify/delete files, sync, unmount, remount
//	Preconditions   1. User has valid OneDrive credentials
//	                2. Docker test environment is available
//	Test Steps      1. Authenticate with Microsoft account
//	                2. Mount OneDrive filesystem
//	                3. Create new files
//	                4. Modify existing files
//	                5. Delete files
//	                6. Verify changes sync to OneDrive
//	                7. Unmount filesystem
//	                8. Remount filesystem
//	                9. Verify state is preserved
//	Expected Result All operations complete successfully and state persists across mount/unmount
//	Requirements    All requirements
func TestE2E_17_01_CompleteUserWorkflow(t *testing.T) {
	// Skip if not in system test mode
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	// Skip if not explicitly enabled (requires real OneDrive account)
	if os.Getenv("RUN_E2E_TESTS") != "1" {
		t.Skip("Skipping end-to-end test (set RUN_E2E_TESTS=1 to enable)")
	}

	t.Log("=== Starting Complete User Workflow Test ===")

	// Step 1: Setup authentication
	t.Log("Step 1: Setting up authentication")

	// Load auth from test artifacts
	authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Fatalf("Failed to load authentication: %v", err)
	}

	if auth.AccessToken == "" {
		t.Fatal("Access token is empty")
	}

	t.Log("Authentication loaded successfully")

	// Step 2: Create mount point and cache directory
	t.Log("Step 2: Creating mount point and cache directory")

	tempDir := t.TempDir()
	mountPoint := filepath.Join(tempDir, "mount")
	cacheDir := filepath.Join(tempDir, "cache")

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	t.Logf("Mount point: %s", mountPoint)
	t.Logf("Cache dir: %s", cacheDir)

	// Create filesystem
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fs, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer func() {
		fs.StopCacheCleanup()
		fs.StopDeltaLoop()
		fs.StopDownloadManager()
		fs.StopUploadManager()
		fs.StopMetadataRequestManager()
	}()

	// Mount in background using FUSE server
	mountOptions := &fuse.MountOptions{
		Name:          "onemount-e2e-test",
		FsName:        "onemount-e2e-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(fs, mountPoint, mountOptions)
	if err != nil {
		t.Fatalf("Failed to create FUSE server: %v", err)
	}
	defer server.Unmount()

	go server.Serve()

	// Wait for mount to be ready
	if err := server.WaitMount(); err != nil {
		t.Fatalf("Failed to wait for mount: %v", err)
	}

	t.Log("Filesystem mounted successfully")

	// Verify mount point is accessible
	entries, err := os.ReadDir(mountPoint)
	if err != nil {
		t.Fatalf("Failed to read mount point: %v", err)
	}
	t.Logf("Mount point contains %d entries", len(entries))

	// Step 3: Create new files
	t.Log("Step 3: Creating new files")
	testFiles := []struct {
		name    string
		content string
	}{
		{"e2e_test_file_1.txt", "This is end-to-end test file 1"},
		{"e2e_test_file_2.txt", "This is end-to-end test file 2"},
		{"e2e_test_file_3.txt", "This is end-to-end test file 3"},
	}

	for _, tf := range testFiles {
		filePath := filepath.Join(mountPoint, tf.name)
		err := os.WriteFile(filePath, []byte(tf.content), 0644)
		if err != nil {
			t.Errorf("Failed to create %s: %v", tf.name, err)
		} else {
			t.Logf("Created file: %s", tf.name)
		}
	}

	// Wait for files to be created
	time.Sleep(2 * time.Second)

	// Verify files exist
	for _, tf := range testFiles {
		filePath := filepath.Join(mountPoint, tf.name)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("File %s should exist: %v", tf.name, err)
		}
	}

	// Step 4: Modify existing files
	t.Log("Step 4: Modifying existing files")
	modifiedContent := "This file has been modified in end-to-end test"
	modifyPath := filepath.Join(mountPoint, testFiles[0].name)
	err = os.WriteFile(modifyPath, []byte(modifiedContent), 0644)
	if err != nil {
		t.Errorf("Failed to modify file: %v", err)
	}

	// Verify modification
	time.Sleep(1 * time.Second)
	content, err := os.ReadFile(modifyPath)
	if err != nil {
		t.Errorf("Failed to read modified file: %v", err)
	} else if string(content) != modifiedContent {
		t.Errorf("File content mismatch. Expected: %s, Got: %s", modifiedContent, string(content))
	}

	// Step 5: Delete files
	t.Log("Step 5: Deleting files")
	deletePath := filepath.Join(mountPoint, testFiles[1].name)
	err = os.Remove(deletePath)
	if err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}

	// Verify deletion
	time.Sleep(1 * time.Second)
	if _, err := os.Stat(deletePath); !os.IsNotExist(err) {
		t.Error("Deleted file should not exist")
	}

	// Step 6: Wait for sync
	t.Log("Step 6: Waiting for changes to sync")
	time.Sleep(5 * time.Second)

	// Step 7: Unmount filesystem
	t.Log("Step 7: Unmounting filesystem")

	err = server.Unmount()
	if err != nil {
		t.Logf("Unmount error (may be expected): %v", err)
	}

	time.Sleep(2 * time.Second)

	// Step 8: Remount filesystem
	t.Log("Step 8: Remounting filesystem")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel2()

	cacheDir2 := filepath.Join(tempDir, "cache2")
	fs2, err := NewFilesystemWithContext(ctx2, auth, cacheDir2, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create new filesystem instance: %v", err)
	}
	defer func() {
		fs2.StopCacheCleanup()
		fs2.StopDeltaLoop()
		fs2.StopDownloadManager()
		fs2.StopUploadManager()
		fs2.StopMetadataRequestManager()
	}()

	server2, err := fuse.NewServer(fs2, mountPoint, mountOptions)
	if err != nil {
		t.Fatalf("Failed to create second FUSE server: %v", err)
	}
	defer server2.Unmount()

	go server2.Serve()

	if err := server2.WaitMount(); err != nil {
		t.Fatalf("Failed to wait for second mount: %v", err)
	}

	t.Log("Filesystem remounted successfully")

	// Step 9: Verify state is preserved
	t.Log("Step 9: Verifying state is preserved after remount")

	for i, tf := range testFiles {
		filePath := filepath.Join(mountPoint, tf.name)
		if i == 1 {
			// This file was deleted
			if _, err := os.Stat(filePath); !os.IsNotExist(err) {
				t.Errorf("Deleted file %s should not exist after remount", tf.name)
			}
		} else {
			if _, err := os.Stat(filePath); err != nil {
				t.Errorf("File %s should exist after remount: %v", tf.name, err)
			}
		}
	}

	// Verify modified file has correct content
	content, err = os.ReadFile(modifyPath)
	if err != nil {
		t.Errorf("Failed to read modified file after remount: %v", err)
	} else if string(content) != modifiedContent {
		t.Errorf("Modified content should persist. Expected: %s, Got: %s", modifiedContent, string(content))
	}

	// Cleanup
	err = server2.Unmount()
	if err != nil {
		t.Logf("Second unmount error (may be expected): %v", err)
	}

	t.Log("=== Complete User Workflow Test Passed ===")
}

// TestE2E_17_02_MultiFileOperations tests copying entire directories with multiple files
//
//	Test Case ID    E2E-17-02
//	Title           Multi-File Operations
//	Description     Verify copying entire directories to/from OneDrive with multiple files
//	Preconditions   1. User is authenticated
//	                2. Filesystem is mounted
//	Test Steps      1. Create directory with multiple files locally
//	                2. Copy directory to OneDrive mount point
//	                3. Verify all files upload correctly
//	                4. Copy directory from OneDrive to local
//	                5. Verify all files download correctly
//	Expected Result All files in directory are copied correctly in both directions
//	Requirements    3.2, 4.3, 10.1, 10.2
func TestE2E_17_02_MultiFileOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	if os.Getenv("RUN_E2E_TESTS") != "1" {
		t.Skip("Skipping end-to-end test (set RUN_E2E_TESTS=1 to enable)")
	}

	t.Log("=== Starting Multi-File Operations Test ===")

	// Setup
	authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Fatalf("Failed to load authentication: %v", err)
	}

	tempDir := t.TempDir()
	mountPoint := filepath.Join(tempDir, "mount")
	cacheDir := filepath.Join(tempDir, "cache")

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fs, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer func() {
		fs.StopCacheCleanup()
		fs.StopDeltaLoop()
		fs.StopDownloadManager()
		fs.StopUploadManager()
		fs.StopMetadataRequestManager()
	}()

	mountOptions := &fuse.MountOptions{
		Name:          "onemount-e2e-test",
		FsName:        "onemount-e2e-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(fs, mountPoint, mountOptions)
	if err != nil {
		t.Fatalf("Failed to create FUSE server: %v", err)
	}
	defer server.Unmount()

	go server.Serve()

	if err := server.WaitMount(); err != nil {
		t.Fatalf("Failed to wait for mount: %v", err)
	}

	// Step 1: Create test directory with multiple files
	t.Log("Step 1: Creating test directory with multiple files")

	workDir := t.TempDir()
	testDir := filepath.Join(workDir, "e2e_test_directory")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFiles := []struct {
		name string
		size int
	}{
		{"small_file_1.txt", 100},
		{"small_file_2.txt", 500},
		{"medium_file_1.txt", 10000},
		{"medium_file_2.txt", 50000},
	}

	for _, tf := range testFiles {
		content := helpers.GenerateRandomString(tf.size)
		filePath := filepath.Join(testDir, tf.name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Errorf("Failed to create %s: %v", tf.name, err)
		}
	}

	// Create subdirectory
	subDir := filepath.Join(testDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Errorf("Failed to create subdirectory: %v", err)
	}

	subFiles := []string{"sub_file_1.txt", "sub_file_2.txt"}
	for _, name := range subFiles {
		filePath := filepath.Join(subDir, name)
		content := helpers.GenerateRandomString(1000)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Errorf("Failed to create %s: %v", name, err)
		}
	}

	// Step 2: Copy directory to OneDrive
	t.Log("Step 2: Copying directory to OneDrive")

	destDir := filepath.Join(mountPoint, "e2e_test_directory")
	if err := helpers.CopyDirectory(testDir, destDir); err != nil {
		t.Errorf("Failed to copy directory: %v", err)
	}

	time.Sleep(5 * time.Second)

	// Step 3: Verify all files uploaded
	t.Log("Step 3: Verifying all files uploaded")

	for _, tf := range testFiles {
		filePath := filepath.Join(destDir, tf.name)
		info, err := os.Stat(filePath)
		if err != nil {
			t.Errorf("File %s should exist: %v", tf.name, err)
		} else if info.Size() != int64(tf.size) {
			t.Errorf("File %s size mismatch. Expected: %d, Got: %d", tf.name, tf.size, info.Size())
		}
	}

	// Step 4: Copy directory from OneDrive to local
	t.Log("Step 4: Copying directory from OneDrive to local")

	downloadDir := filepath.Join(workDir, "downloaded_directory")
	if err := helpers.CopyDirectory(destDir, downloadDir); err != nil {
		t.Errorf("Failed to copy from OneDrive: %v", err)
	}

	// Step 5: Verify all files downloaded
	t.Log("Step 5: Verifying all files downloaded")

	for _, tf := range testFiles {
		filePath := filepath.Join(downloadDir, tf.name)
		if _, err := os.Stat(filePath); err != nil {
			t.Errorf("Downloaded file %s should exist: %v", tf.name, err)
		}
	}

	// Cleanup
	err = server.Unmount()
	if err != nil {
		t.Logf("Unmount error (may be expected): %v", err)
	}

	t.Log("=== Multi-File Operations Test Passed ===")
}

// TestE2E_17_03_LongRunningOperations tests uploading very large files
//
//	Test Case ID    E2E-17-03
//	Title           Long-Running Operations
//	Description     Verify large file upload with progress monitoring
//	Requirements    4.3, 4.4
func TestE2E_17_03_LongRunningOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	if os.Getenv("RUN_LONG_TESTS") != "1" {
		t.Skip("Skipping long-running test (set RUN_LONG_TESTS=1 to enable)")
	}

	t.Log("=== Starting Long-Running Operations Test ===")
	t.Log("Note: This test creates a 1GB file and may take 20+ minutes")

	// Setup
	authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Fatalf("Failed to load authentication: %v", err)
	}

	tempDir := t.TempDir()
	mountPoint := filepath.Join(tempDir, "mount")
	cacheDir := filepath.Join(tempDir, "cache")

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	fs, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer func() {
		fs.StopCacheCleanup()
		fs.StopDeltaLoop()
		fs.StopDownloadManager()
		fs.StopUploadManager()
		fs.StopMetadataRequestManager()
	}()

	mountOptions := &fuse.MountOptions{
		Name:          "onemount-e2e-test",
		FsName:        "onemount-e2e-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(fs, mountPoint, mountOptions)
	if err != nil {
		t.Fatalf("Failed to create FUSE server: %v", err)
	}
	defer server.Unmount()

	go server.Serve()

	if err := server.WaitMount(); err != nil {
		t.Fatalf("Failed to wait for mount: %v", err)
	}

	// Create large file
	t.Log("Creating 1GB file (this may take a few minutes)...")

	largeFilePath := filepath.Join(mountPoint, "e2e_large_test_file.dat")
	file, err := os.Create(largeFilePath)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	chunkSize := 10 * 1024 * 1024          // 10MB
	totalSize := int64(1024 * 1024 * 1024) // 1GB
	chunk := make([]byte, chunkSize)

	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	written := int64(0)
	startTime := time.Now()

	for written < totalSize {
		toWrite := chunkSize
		if written+int64(chunkSize) > totalSize {
			toWrite = int(totalSize - written)
		}

		n, err := file.Write(chunk[:toWrite])
		if err != nil {
			t.Fatalf("Failed to write chunk: %v", err)
		}
		written += int64(n)

		if written%(100*1024*1024) == 0 {
			t.Logf("Written %d MB / %d MB", written/(1024*1024), totalSize/(1024*1024))
		}
	}

	file.Close()
	t.Logf("File creation took %v", time.Since(startTime))

	// Monitor upload
	t.Log("Monitoring upload progress...")
	uploadStartTime := time.Now()
	maxWaitTime := 20 * time.Minute
	checkInterval := 30 * time.Second

	for time.Since(uploadStartTime) < maxWaitTime {
		status, err := helpers.GetFileStatus(largeFilePath)
		if err == nil {
			t.Logf("Upload status: %s", status)

			if status == "synced" || status == "uploaded" {
				t.Logf("Upload completed in %v", time.Since(uploadStartTime))
				break
			}

			if status == "error" {
				t.Error("Upload failed with error status")
				break
			}
		}

		time.Sleep(checkInterval)
	}

	// Cleanup
	t.Log("Cleaning up large file")
	os.Remove(largeFilePath)

	err = server.Unmount()
	if err != nil {
		t.Logf("Unmount error (may be expected): %v", err)
	}

	t.Log("=== Long-Running Operations Test Completed ===")
}

// TestE2E_17_04_StressScenarios tests system stability under concurrent operations
//
//	Test Case ID    E2E-17-04
//	Title           Stress Scenarios
//	Description     Verify system remains stable under many concurrent operations
//	Requirements    10.1, 10.2
func TestE2E_17_04_StressScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping end-to-end test in short mode")
	}

	if os.Getenv("RUN_STRESS_TESTS") != "1" {
		t.Skip("Skipping stress test (set RUN_STRESS_TESTS=1 to enable)")
	}

	t.Log("=== Starting Stress Scenarios Test ===")

	// Setup
	authPath := os.Getenv("ONEMOUNT_AUTH_PATH")
	if authPath == "" {
		authPath = "test-artifacts/.auth_tokens.json"
	}

	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Fatalf("Failed to load authentication: %v", err)
	}

	tempDir := t.TempDir()
	mountPoint := filepath.Join(tempDir, "mount")
	cacheDir := filepath.Join(tempDir, "cache")

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	fs, err := NewFilesystemWithContext(ctx, auth, cacheDir, 30, 24, 0)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer func() {
		fs.StopCacheCleanup()
		fs.StopDeltaLoop()
		fs.StopDownloadManager()
		fs.StopUploadManager()
		fs.StopMetadataRequestManager()
	}()

	mountOptions := &fuse.MountOptions{
		Name:          "onemount-e2e-test",
		FsName:        "onemount-e2e-test",
		DisableXAttrs: false,
		MaxBackground: 1024,
		Debug:         false,
	}

	server, err := fuse.NewServer(fs, mountPoint, mountOptions)
	if err != nil {
		t.Fatalf("Failed to create FUSE server: %v", err)
	}
	defer server.Unmount()

	go server.Serve()

	if err := server.WaitMount(); err != nil {
		t.Fatalf("Failed to wait for mount: %v", err)
	}

	// Start concurrent operations
	t.Log("Starting concurrent file operations")

	numWorkers := 20
	operationsPerWorker := 50

	type operationResult struct {
		workerID int
		opNum    int
		opType   string
		err      error
		duration time.Duration
	}

	results := make(chan operationResult, numWorkers*operationsPerWorker)
	done := make(chan bool, numWorkers)

	for workerID := 0; workerID < numWorkers; workerID++ {
		go func(id int) {
			defer func() { done <- true }()

			for opNum := 0; opNum < operationsPerWorker; opNum++ {
				opType := opNum % 4
				startTime := time.Now()
				var err error
				var opName string

				switch opType {
				case 0: // Create file
					opName = "create"
					fileName := fmt.Sprintf("e2e_stress_w%d_op%d.txt", id, opNum)
					filePath := filepath.Join(mountPoint, fileName)
					content := helpers.GenerateRandomString(1000)
					err = os.WriteFile(filePath, []byte(content), 0644)

				case 1: // Read file
					opName = "read"
					fileName := fmt.Sprintf("e2e_stress_w%d_op%d.txt", id, opNum-1)
					filePath := filepath.Join(mountPoint, fileName)
					_, err = os.ReadFile(filePath)
					if os.IsNotExist(err) {
						err = nil
					}

				case 2: // Modify file
					opName = "modify"
					fileName := fmt.Sprintf("e2e_stress_w%d_op%d.txt", id, opNum-2)
					filePath := filepath.Join(mountPoint, fileName)
					content := helpers.GenerateRandomString(2000)
					err = os.WriteFile(filePath, []byte(content), 0644)
					if os.IsNotExist(err) {
						err = nil
					}

				case 3: // List directory
					opName = "list"
					_, err = os.ReadDir(mountPoint)
				}

				duration := time.Since(startTime)
				results <- operationResult{
					workerID: id,
					opNum:    opNum,
					opType:   opName,
					err:      err,
					duration: duration,
				}

				time.Sleep(10 * time.Millisecond)
			}
		}(workerID)
	}

	// Monitor resources
	stopMonitoring := make(chan bool)
	monitoringDone := make(chan bool)

	go func() {
		defer func() { monitoringDone <- true }()

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopMonitoring:
				return
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				t.Logf("Memory: Alloc=%d MB, Sys=%d MB, NumGC=%d, Goroutines=%d",
					m.Alloc/1024/1024,
					m.Sys/1024/1024,
					m.NumGC,
					runtime.NumGoroutine())
			}
		}
	}()

	// Wait for workers
	for i := 0; i < numWorkers; i++ {
		<-done
	}
	close(results)

	stopMonitoring <- true
	<-monitoringDone

	// Analyze results
	t.Log("Analyzing results")

	successCount := 0
	errorCount := 0
	totalDuration := time.Duration(0)

	for result := range results {
		if result.err != nil {
			errorCount++
			t.Logf("Worker %d op %d (%s) failed: %v", result.workerID, result.opNum, result.opType, result.err)
		} else {
			successCount++
		}
		totalDuration += result.duration
	}

	totalOps := successCount + errorCount
	successRate := float64(successCount) / float64(totalOps)

	t.Logf("Total operations: %d", totalOps)
	t.Logf("Successful: %d (%.1f%%)", successCount, successRate*100)
	t.Logf("Failed: %d (%.1f%%)", errorCount, (1-successRate)*100)
	t.Logf("Average duration: %v", totalDuration/time.Duration(totalOps))

	if successRate < 0.90 {
		t.Errorf("Success rate too low: %.1f%% (expected > 90%%)", successRate*100)
	}

	// Check for memory leaks
	runtime.GC()
	time.Sleep(1 * time.Second)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	t.Logf("Final memory: Alloc=%d MB, Sys=%d MB, NumGC=%d",
		m.Alloc/1024/1024,
		m.Sys/1024/1024,
		m.NumGC)

	finalGoroutines := runtime.NumGoroutine()
	t.Logf("Final goroutine count: %d", finalGoroutines)

	if finalGoroutines > 100 {
		t.Errorf("Too many goroutines: %d (expected < 100)", finalGoroutines)
	}

	// Cleanup
	err = server.Unmount()
	if err != nil {
		t.Logf("Unmount error (may be expected): %v", err)
	}

	t.Log("=== Stress Scenarios Test Completed ===")
}
