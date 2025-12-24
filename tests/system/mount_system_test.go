package system

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/auriora/onemount/internal/fs"
	"github.com/auriora/onemount/internal/graph"
	"github.com/auriora/onemount/internal/testutil"
)

// TestSystemST_Mount_01_01_RealOneDriveMount tests mounting with real OneDrive integration.
//
// Test Case ID: ST-MOUNT-01-01
// Description: Verify filesystem can mount and access real OneDrive content
// Prerequisites: Valid authentication tokens and network connectivity
// Expected Result: Mount succeeds and real OneDrive files are accessible
// Requirements: 2.1, 2.2, 3.1
func TestSystemST_Mount_01_01_RealOneDriveMount(t *testing.T) {
	// Load existing auth tokens
	authPath := testutil.AuthTokensPath
	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Skipf("Skipping real OneDrive mount test - no existing auth tokens: %v", err)
	}

	// Create temporary mount point
	mountPoint, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "system-mount-*")
	if err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}
	defer os.RemoveAll(mountPoint)

	// Create filesystem
	filesystem, err := fs.NewFilesystem(auth, mountPoint, 300) // 5 minute cache TTL
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer filesystem.Stop()

	// Wait a moment for initial sync
	time.Sleep(2 * time.Second)

	// Test basic directory listing
	entries, err := os.ReadDir(mountPoint)
	if err != nil {
		t.Fatalf("Failed to read mount point directory: %v", err)
	}

	t.Logf("Found %d entries in OneDrive root", len(entries))
	for i, entry := range entries {
		if i < 5 { // Log first 5 entries
			t.Logf("  - %s (dir: %v)", entry.Name(), entry.IsDir())
		}
	}

	// Test file stat operations
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip directories for this test
		}

		filePath := filepath.Join(mountPoint, entry.Name())
		info, err := os.Stat(filePath)
		if err != nil {
			t.Errorf("Failed to stat file %s: %v", entry.Name(), err)
			continue
		}

		if info.Size() < 0 {
			t.Errorf("File %s has negative size: %d", entry.Name(), info.Size())
		}

		if info.ModTime().IsZero() {
			t.Errorf("File %s has zero modification time", entry.Name())
		}

		// Only test first file to avoid long test times
		break
	}

	t.Log("Real OneDrive mount test completed successfully")
}

// TestSystemST_Mount_02_01_FileOperations tests file operations with real OneDrive.
//
// Test Case ID: ST-MOUNT-02-01
// Description: Verify file read/write operations work with real OneDrive
// Prerequisites: Valid authentication tokens and mounted filesystem
// Expected Result: File operations succeed and sync to OneDrive
// Requirements: 3.2, 4.1, 4.2
func TestSystemST_Mount_02_01_FileOperations(t *testing.T) {
	// Load existing auth tokens
	authPath := testutil.AuthTokensPath
	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Skipf("Skipping file operations test - no existing auth tokens: %v", err)
	}

	// Create temporary mount point
	mountPoint, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "system-fileops-*")
	if err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}
	defer os.RemoveAll(mountPoint)

	// Create filesystem
	filesystem, err := fs.NewFilesystem(auth, mountPoint, 300)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer filesystem.Stop()

	// Wait for initial sync
	time.Sleep(2 * time.Second)

	// Test file creation
	testFileName := "onemount-test-" + time.Now().Format("20060102-150405") + ".txt"
	testFilePath := filepath.Join(mountPoint, testFileName)
	testContent := "This is a test file created by OneMount system tests.\nTimestamp: " + time.Now().String()

	// Write test file
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Verify file was created
	info, err := os.Stat(testFilePath)
	if err != nil {
		t.Fatalf("Failed to stat created file: %v", err)
	}

	if info.Size() != int64(len(testContent)) {
		t.Errorf("File size mismatch: expected %d, got %d", len(testContent), info.Size())
	}

	// Read file back
	readContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(readContent) != testContent {
		t.Errorf("File content mismatch:\nExpected: %s\nGot: %s", testContent, string(readContent))
	}

	// Wait for upload to complete
	t.Log("Waiting for file upload to complete...")
	time.Sleep(5 * time.Second)

	// Test file modification
	modifiedContent := testContent + "\nModified at: " + time.Now().String()
	err = os.WriteFile(testFilePath, []byte(modifiedContent), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Verify modification
	readModified, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	if string(readModified) != modifiedContent {
		t.Errorf("Modified file content mismatch:\nExpected: %s\nGot: %s", modifiedContent, string(readModified))
	}

	// Clean up test file
	err = os.Remove(testFilePath)
	if err != nil {
		t.Errorf("Failed to remove test file: %v", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	// Verify file was deleted
	if _, err := os.Stat(testFilePath); !os.IsNotExist(err) {
		t.Error("Test file still exists after deletion")
	}

	t.Log("File operations test completed successfully")
}

// TestSystemST_Mount_03_01_DirectoryOperations tests directory operations with real OneDrive.
//
// Test Case ID: ST-MOUNT-03-01
// Description: Verify directory create/delete operations work with real OneDrive
// Prerequisites: Valid authentication tokens and mounted filesystem
// Expected Result: Directory operations succeed and sync to OneDrive
// Requirements: 4.1
func TestSystemST_Mount_03_01_DirectoryOperations(t *testing.T) {
	// Load existing auth tokens
	authPath := testutil.AuthTokensPath
	auth, err := graph.LoadAuthTokens(authPath)
	if err != nil {
		t.Skipf("Skipping directory operations test - no existing auth tokens: %v", err)
	}

	// Create temporary mount point
	mountPoint, err := os.MkdirTemp(testutil.TestSandboxTmpDir, "system-dirops-*")
	if err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}
	defer os.RemoveAll(mountPoint)

	// Create filesystem
	filesystem, err := fs.NewFilesystem(auth, mountPoint, 300)
	if err != nil {
		t.Fatalf("Failed to create filesystem: %v", err)
	}
	defer filesystem.Stop()

	// Wait for initial sync
	time.Sleep(2 * time.Second)

	// Test directory creation
	testDirName := "onemount-test-dir-" + time.Now().Format("20060102-150405")
	testDirPath := filepath.Join(mountPoint, testDirName)

	err = os.Mkdir(testDirPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(testDirPath)
	if err != nil {
		t.Fatalf("Failed to stat created directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}

	// Create a file inside the directory
	testFileName := "test-file.txt"
	testFilePath := filepath.Join(testDirPath, testFileName)
	testContent := "File inside test directory"

	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create file in test directory: %v", err)
	}

	// Wait for operations to sync
	t.Log("Waiting for directory operations to sync...")
	time.Sleep(5 * time.Second)

	// List directory contents
	entries, err := os.ReadDir(testDirPath)
	if err != nil {
		t.Fatalf("Failed to read test directory: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("Expected 1 file in directory, got %d", len(entries))
	}

	if len(entries) > 0 && entries[0].Name() != testFileName {
		t.Errorf("Expected file name %s, got %s", testFileName, entries[0].Name())
	}

	// Clean up - remove file first, then directory
	err = os.Remove(testFilePath)
	if err != nil {
		t.Errorf("Failed to remove test file: %v", err)
	}

	err = os.Remove(testDirPath)
	if err != nil {
		t.Errorf("Failed to remove test directory: %v", err)
	}

	// Wait for deletion to sync
	time.Sleep(3 * time.Second)

	// Verify directory was deleted
	if _, err := os.Stat(testDirPath); !os.IsNotExist(err) {
		t.Error("Test directory still exists after deletion")
	}

	t.Log("Directory operations test completed successfully")
}
