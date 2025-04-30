package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// CreateTestFile creates a file with the given content and ensures it's cleaned up after the test
func CreateTestFile(t *testing.T, dir, name string, content []byte) string {
	path := filepath.Join(dir, name)

	err := os.WriteFile(path, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file %s: %v", path, err)
	}

	t.Cleanup(func() {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up test file %s: %v", path, err)
		}
	})

	return path
}

// CreateTestDir creates a directory and ensures it's cleaned up after the test
func CreateTestDir(t *testing.T, parent, name string) string {
	path := filepath.Join(parent, name)

	err := os.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory %s: %v", path, err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(path); err != nil {
			t.Logf("Warning: Failed to clean up test directory %s: %v", path, err)
		}
	})

	return path
}

// CreateTempDir creates a temporary directory and ensures it's cleaned up after the test
func CreateTempDir(t *testing.T, prefix string) string {
	dir, err := os.MkdirTemp("", prefix)
	if err != nil {
		t.Fatalf("Failed to create temporary directory with prefix %s: %v", prefix, err)
	}

	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Logf("Warning: Failed to clean up temporary directory %s: %v", dir, err)
		}
	})

	return dir
}

// CreateTempFile creates a temporary file with the given content and ensures it's cleaned up after the test
func CreateTempFile(t *testing.T, dir, pattern string, content []byte) string {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		t.Fatalf("Failed to create temporary file with pattern %s: %v", pattern, err)
	}

	if content != nil {
		if _, err := file.Write(content); err != nil {
			err := file.Close()
			if err != nil {
				t.Logf("Warning: Failed to close temporary file %s: %v", file.Name(), err)
				return ""
			}
			t.Fatalf("Failed to write content to temporary file %s: %v", file.Name(), err)
		}
	}

	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close temporary file %s: %v", file.Name(), err)
	}

	t.Cleanup(func() {
		if err := os.Remove(file.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: Failed to clean up temporary file %s: %v", file.Name(), err)
		}
	})

	return file.Name()
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// FileContains checks if a file contains the expected content
func FileContains(t *testing.T, path string, expected []byte) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	return string(content) == string(expected)
}

// AssertFileExists asserts that a file exists at the given path
func AssertFileExists(t *testing.T, path string) {
	if !FileExists(path) {
		t.Fatalf("Expected file to exist at %s, but it doesn't", path)
	}
}

// AssertFileNotExists asserts that a file does not exist at the given path
func AssertFileNotExists(t *testing.T, path string) {
	if FileExists(path) {
		t.Fatalf("Expected file not to exist at %s, but it does", path)
	}
}

// AssertFileContains asserts that a file contains the expected content
func AssertFileContains(t *testing.T, path string, expected []byte) {
	if !FileContains(t, path, expected) {
		content, _ := os.ReadFile(path)
		t.Fatalf("Expected file %s to contain %q, but got %q", path, expected, content)
	}
}

// CaptureFileSystemState captures the current state of the filesystem
// by listing all files and directories in the specified directory
func CaptureFileSystemState(dir string) (map[string]os.FileInfo, error) {
	state := make(map[string]os.FileInfo)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the directory itself
		if path == dir {
			return nil
		}
		// Store the file info in the state map
		state[path] = info
		return nil
	})

	return state, err
}
