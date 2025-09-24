// Package helpers provides testing utilities for the OneMount project.
package helpers

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/auriora/onemount/internal/logging"
)

// FileTestHelper provides utilities for file operations in tests
type FileTestHelper struct {
	t            *testing.T
	createdFiles []string
	createdDirs  []string
	mu           sync.Mutex
}

// NewFileTestHelper creates a new file test helper
func NewFileTestHelper(t *testing.T) *FileTestHelper {
	helper := &FileTestHelper{
		t:            t,
		createdFiles: make([]string, 0),
		createdDirs:  make([]string, 0),
	}

	// Register cleanup function
	t.Cleanup(helper.Cleanup)

	return helper
}

// CreateTestFile creates a file with the given content and ensures it's cleaned up after the test
func (h *FileTestHelper) CreateTestFile(path, content string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory %s: %v", dir, err)
	}

	// Create the file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create test file %s: %v", path, err)
	}

	// Track for cleanup
	h.createdFiles = append(h.createdFiles, path)

	logging.Debug().Str("path", path).Int("size", len(content)).Msg("Created test file")
	return nil
}

// CreateTestDir creates a directory and ensures it's cleaned up after the test
func (h *FileTestHelper) CreateTestDir(path string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create test directory %s: %v", path, err)
	}

	// Track for cleanup
	h.createdDirs = append(h.createdDirs, path)

	logging.Debug().Str("path", path).Msg("Created test directory")
	return nil
}

// CreateTempDir creates a temporary directory and ensures it's cleaned up after the test
func (h *FileTestHelper) CreateTempDir(prefix string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	tempDir, err := os.MkdirTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %v", err)
	}

	// Track for cleanup
	h.createdDirs = append(h.createdDirs, tempDir)

	logging.Debug().Str("path", tempDir).Str("prefix", prefix).Msg("Created temporary directory")
	return tempDir, nil
}

// CreateTempFile creates a temporary file with the given content and ensures it's cleaned up after the test
func (h *FileTestHelper) CreateTempFile(prefix, content string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	tempFile, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	// Write content
	if _, err := tempFile.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to write content to temporary file: %v", err)
	}

	path := tempFile.Name()

	// Track for cleanup
	h.createdFiles = append(h.createdFiles, path)

	logging.Debug().Str("path", path).Str("prefix", prefix).Int("size", len(content)).Msg("Created temporary file")
	return path, nil
}

// FileExists checks if a file exists at the given path
func (h *FileTestHelper) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileContains checks if a file contains the expected content
func (h *FileTestHelper) FileContains(path, expectedContent string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("failed to read file %s: %v", path, err)
	}

	return strings.Contains(string(content), expectedContent), nil
}

// AssertFileExists asserts that a file exists at the given path
func (h *FileTestHelper) AssertFileExists(path string) {
	if !h.FileExists(path) {
		h.t.Errorf("Expected file to exist at path: %s", path)
	}
}

// AssertFileNotExists asserts that a file does not exist at the given path
func (h *FileTestHelper) AssertFileNotExists(path string) {
	if h.FileExists(path) {
		h.t.Errorf("Expected file to not exist at path: %s", path)
	}
}

// AssertFileContains asserts that a file contains the expected content
func (h *FileTestHelper) AssertFileContains(path, expectedContent string) {
	contains, err := h.FileContains(path, expectedContent)
	if err != nil {
		h.t.Errorf("Failed to check file content: %v", err)
		return
	}

	if !contains {
		h.t.Errorf("Expected file %s to contain: %s", path, expectedContent)
	}
}

// AssertFileContent asserts that a file has exactly the expected content
func (h *FileTestHelper) AssertFileContent(path, expectedContent string) {
	content, err := os.ReadFile(path)
	if err != nil {
		h.t.Errorf("Failed to read file %s: %v", path, err)
		return
	}

	if string(content) != expectedContent {
		h.t.Errorf("File content mismatch.\nExpected: %s\nActual: %s", expectedContent, string(content))
	}
}

// FileSystemState represents the state of a filesystem
type FileSystemState struct {
	Files       map[string]FileInfo `json:"files"`
	Directories []string            `json:"directories"`
}

// FileInfo represents information about a file
type FileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
}

// CaptureFileSystemState captures the current state of the filesystem by listing all files and directories
func (h *FileTestHelper) CaptureFileSystemState(rootPath string) (*FileSystemState, error) {
	state := &FileSystemState{
		Files:       make(map[string]FileInfo),
		Directories: make([]string, 0),
	}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		if d.IsDir() {
			state.Directories = append(state.Directories, relPath)
		} else {
			info, err := d.Info()
			if err != nil {
				return err
			}

			state.Files[relPath] = FileInfo{
				Path:    relPath,
				Size:    info.Size(),
				Mode:    info.Mode().String(),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to capture filesystem state: %v", err)
	}

	logging.Debug().Str("root", rootPath).Int("files", len(state.Files)).Int("dirs", len(state.Directories)).Msg("Captured filesystem state")
	return state, nil
}

// Cleanup removes all created files and directories
func (h *FileTestHelper) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Remove files first
	for _, file := range h.createdFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			logging.Warn().Err(err).Str("path", file).Msg("Failed to remove test file")
		}
	}

	// Remove directories (in reverse order to handle nested directories)
	for i := len(h.createdDirs) - 1; i >= 0; i-- {
		dir := h.createdDirs[i]
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			logging.Warn().Err(err).Str("path", dir).Msg("Failed to remove test directory")
		}
	}

	logging.Debug().Int("files", len(h.createdFiles)).Int("dirs", len(h.createdDirs)).Msg("Cleaned up test files and directories")
}

// Global helper functions for convenience

// CreateTestFile creates a file with the given content (convenience function)
func CreateTestFile(t *testing.T, path, content string) error {
	helper := NewFileTestHelper(t)
	return helper.CreateTestFile(path, content)
}

// CreateTestDir creates a directory (convenience function)
func CreateTestDir(t *testing.T, path string) error {
	helper := NewFileTestHelper(t)
	return helper.CreateTestDir(path)
}

// CreateTempDir creates a temporary directory (convenience function)
func CreateTempDir(t *testing.T, prefix string) (string, error) {
	helper := NewFileTestHelper(t)
	return helper.CreateTempDir(prefix)
}

// CreateTempFile creates a temporary file with content (convenience function)
func CreateTempFile(t *testing.T, prefix, content string) (string, error) {
	helper := NewFileTestHelper(t)
	return helper.CreateTempFile(prefix, content)
}

// FileExists checks if a file exists (convenience function)
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileContains checks if a file contains expected content (convenience function)
func FileContains(path, expectedContent string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(content), expectedContent), nil
}

// AssertFileExists asserts that a file exists (convenience function)
func AssertFileExists(t *testing.T, path string) {
	if !FileExists(path) {
		t.Errorf("Expected file to exist at path: %s", path)
	}
}

// AssertFileNotExists asserts that a file does not exist (convenience function)
func AssertFileNotExists(t *testing.T, path string) {
	if FileExists(path) {
		t.Errorf("Expected file to not exist at path: %s", path)
	}
}

// AssertFileContains asserts that a file contains expected content (convenience function)
func AssertFileContains(t *testing.T, path, expectedContent string) {
	contains, err := FileContains(path, expectedContent)
	if err != nil {
		t.Errorf("Failed to check file content: %v", err)
		return
	}

	if !contains {
		t.Errorf("Expected file %s to contain: %s", path, expectedContent)
	}
}
