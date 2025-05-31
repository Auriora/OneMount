package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileTestHelper_CreateTestFile(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a test file
	testPath := filepath.Join(os.TempDir(), "test_file.txt")
	testContent := "Hello, World!"

	err := helper.CreateTestFile(testPath, testContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify file exists
	if !helper.FileExists(testPath) {
		t.Error("Test file should exist")
	}

	// Verify content
	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content %q, got %q", testContent, string(content))
	}
}

func TestFileTestHelper_CreateTestDir(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a test directory
	testPath := filepath.Join(os.TempDir(), "test_dir")

	err := helper.CreateTestDir(testPath)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("Test directory should exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Path should be a directory")
	}
}

func TestFileTestHelper_CreateTempDir(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a temporary directory
	tempDir, err := helper.CreateTempDir("test_prefix_")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(tempDir)
	if err != nil {
		t.Fatalf("Temporary directory should exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("Path should be a directory")
	}

	// Verify prefix is in the name
	if !strings.Contains(filepath.Base(tempDir), "test_prefix_") {
		t.Errorf("Directory name should contain prefix, got: %s", tempDir)
	}
}

func TestFileTestHelper_CreateTempFile(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a temporary file
	testContent := "Temporary file content"
	tempFile, err := helper.CreateTempFile("test_prefix_", testContent)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	// Verify file exists
	if !helper.FileExists(tempFile) {
		t.Error("Temporary file should exist")
	}

	// Verify content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read temporary file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Expected content %q, got %q", testContent, string(content))
	}

	// Verify prefix is in the name
	if !strings.Contains(filepath.Base(tempFile), "test_prefix_") {
		t.Errorf("File name should contain prefix, got: %s", tempFile)
	}
}

func TestFileTestHelper_FileExists(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Test with non-existent file
	nonExistentPath := filepath.Join(os.TempDir(), "non_existent_file.txt")
	if helper.FileExists(nonExistentPath) {
		t.Error("Non-existent file should not exist")
	}

	// Create a file and test
	testPath := filepath.Join(os.TempDir(), "existing_file.txt")
	err := helper.CreateTestFile(testPath, "content")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !helper.FileExists(testPath) {
		t.Error("Created file should exist")
	}
}

func TestFileTestHelper_FileContains(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a test file
	testPath := filepath.Join(os.TempDir(), "content_test.txt")
	testContent := "This is a test file with specific content"

	err := helper.CreateTestFile(testPath, testContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test positive case
	contains, err := helper.FileContains(testPath, "specific content")
	if err != nil {
		t.Fatalf("Failed to check file content: %v", err)
	}

	if !contains {
		t.Error("File should contain the expected text")
	}

	// Test negative case
	contains, err = helper.FileContains(testPath, "not in file")
	if err != nil {
		t.Fatalf("Failed to check file content: %v", err)
	}

	if contains {
		t.Error("File should not contain the unexpected text")
	}
}

func TestFileTestHelper_AssertFileExists(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a test file
	testPath := filepath.Join(os.TempDir(), "assert_exists_test.txt")
	err := helper.CreateTestFile(testPath, "content")
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This should not fail
	helper.AssertFileExists(testPath)
}

func TestFileTestHelper_AssertFileNotExists(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Test with non-existent file
	nonExistentPath := filepath.Join(os.TempDir(), "should_not_exist.txt")

	// This should not fail
	helper.AssertFileNotExists(nonExistentPath)
}

func TestFileTestHelper_AssertFileContains(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a test file
	testPath := filepath.Join(os.TempDir(), "assert_contains_test.txt")
	testContent := "This file contains specific text"

	err := helper.CreateTestFile(testPath, testContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This should not fail
	helper.AssertFileContains(testPath, "specific text")
}

func TestFileTestHelper_AssertFileContent(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a test file
	testPath := filepath.Join(os.TempDir(), "assert_content_test.txt")
	testContent := "Exact content match"

	err := helper.CreateTestFile(testPath, testContent)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This should not fail
	helper.AssertFileContent(testPath, testContent)
}

func TestFileTestHelper_CaptureFileSystemState(t *testing.T) {
	helper := NewFileTestHelper(t)

	// Create a temporary directory structure
	tempDir, err := helper.CreateTempDir("fs_state_test_")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create some files and directories
	subDir := filepath.Join(tempDir, "subdir")
	err = helper.CreateTestDir(subDir)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	file1 := filepath.Join(tempDir, "file1.txt")
	err = helper.CreateTestFile(file1, "content1")
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	file2 := filepath.Join(subDir, "file2.txt")
	err = helper.CreateTestFile(file2, "content2")
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Capture filesystem state
	state, err := helper.CaptureFileSystemState(tempDir)
	if err != nil {
		t.Fatalf("Failed to capture filesystem state: %v", err)
	}

	// Verify captured state
	if len(state.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(state.Files))
	}

	if len(state.Directories) < 2 { // At least "." and "subdir"
		t.Errorf("Expected at least 2 directories, got %d", len(state.Directories))
	}

	// Check specific files
	if _, exists := state.Files["file1.txt"]; !exists {
		t.Error("file1.txt should be in captured state")
	}

	if _, exists := state.Files[filepath.Join("subdir", "file2.txt")]; !exists {
		t.Error("subdir/file2.txt should be in captured state")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test CreateTempFile convenience function
	tempFile, err := CreateTempFile(t, "convenience_test_", "test content")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test FileExists convenience function
	if !FileExists(tempFile) {
		t.Error("Temp file should exist")
	}

	// Test FileContains convenience function
	contains, err := FileContains(tempFile, "test content")
	if err != nil {
		t.Fatalf("Failed to check file content: %v", err)
	}

	if !contains {
		t.Error("File should contain expected content")
	}

	// Test AssertFileExists convenience function
	AssertFileExists(t, tempFile)

	// Test AssertFileContains convenience function
	AssertFileContains(t, tempFile, "test content")
}
