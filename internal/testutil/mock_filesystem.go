// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"os"
	"time"
)

// MockFile represents a file in the mock filesystem
type MockFile struct {
	Name     string
	Content  []byte
	Metadata map[string]interface{}
	Mode     os.FileMode
	ModTime  time.Time
	IsDir    bool
}

// FSOperation represents a filesystem operation
type FSOperation struct {
	Type      string // e.g., "read", "write", "list", "delete"
	Path      string
	Timestamp time.Time
	Data      map[string]interface{}
}

// ErrorConditions simulates different error scenarios
type ErrorConditions struct {
	ReadErrors  map[string]error
	WriteErrors map[string]error
	ListErrors  map[string]error
	StatErrors  map[string]error
}

// MockFileSystemProvider implements the MockProvider interface for simulating filesystem operations
type MockFileSystemProvider struct {
	// Virtual filesystem state
	files map[string]*MockFile

	// Record of operations
	operations []FSOperation

	// Simulated error conditions
	errorConditions ErrorConditions

	// Mock recorder for verification
	recorder MockRecorder

	// Configuration for mock behavior
	config MockConfig
}

// NewMockFileSystemProvider creates a new MockFileSystemProvider
func NewMockFileSystemProvider() *MockFileSystemProvider {
	return &MockFileSystemProvider{
		files:      make(map[string]*MockFile),
		operations: make([]FSOperation, 0),
		errorConditions: ErrorConditions{
			ReadErrors:  make(map[string]error),
			WriteErrors: make(map[string]error),
			ListErrors:  make(map[string]error),
			StatErrors:  make(map[string]error),
		},
		recorder: NewBasicMockRecorder(),
		config:   MockConfig{},
	}
}

// Setup initializes the mock provider
func (m *MockFileSystemProvider) Setup() error {
	// Nothing to do for basic setup
	return nil
}

// Teardown cleans up the mock provider
func (m *MockFileSystemProvider) Teardown() error {
	// Nothing to do for basic teardown
	return nil
}

// Reset resets the mock provider to its initial state
func (m *MockFileSystemProvider) Reset() error {
	m.files = make(map[string]*MockFile)
	m.operations = make([]FSOperation, 0)
	m.errorConditions = ErrorConditions{
		ReadErrors:  make(map[string]error),
		WriteErrors: make(map[string]error),
		ListErrors:  make(map[string]error),
		StatErrors:  make(map[string]error),
	}
	m.recorder = NewBasicMockRecorder()
	m.config = MockConfig{}
	return nil
}

// SetConfig sets the mock configuration
func (m *MockFileSystemProvider) SetConfig(config MockConfig) {
	m.config = config
}

// GetRecorder returns the mock recorder
func (m *MockFileSystemProvider) GetRecorder() MockRecorder {
	return m.recorder
}

// AddFile adds a file to the mock filesystem
func (m *MockFileSystemProvider) AddFile(path string, content []byte, mode os.FileMode, isDir bool) {
	m.files[path] = &MockFile{
		Name:     path,
		Content:  content,
		Metadata: make(map[string]interface{}),
		Mode:     mode,
		ModTime:  time.Now(),
		IsDir:    isDir,
	}
	m.recorder.RecordCall("AddFile", path, content, mode, isDir)
}

// AddFileWithMetadata adds a file with metadata to the mock filesystem
func (m *MockFileSystemProvider) AddFileWithMetadata(path string, content []byte, metadata map[string]interface{}, mode os.FileMode, isDir bool) {
	m.files[path] = &MockFile{
		Name:     path,
		Content:  content,
		Metadata: metadata,
		Mode:     mode,
		ModTime:  time.Now(),
		IsDir:    isDir,
	}
	m.recorder.RecordCall("AddFileWithMetadata", path, content, metadata, mode, isDir)
}

// RemoveFile removes a file from the mock filesystem
func (m *MockFileSystemProvider) RemoveFile(path string) error {
	m.recorder.RecordCall("RemoveFile", path)

	if _, exists := m.files[path]; !exists {
		err := os.ErrNotExist
		m.recorder.RecordCallWithResult("RemoveFile", nil, err, path)
		return err
	}

	delete(m.files, path)
	m.operations = append(m.operations, FSOperation{
		Type:      "delete",
		Path:      path,
		Timestamp: time.Now(),
	})

	m.recorder.RecordCallWithResult("RemoveFile", nil, nil, path)
	return nil
}

// AddErrorCondition adds an error condition for a specific operation and path
func (m *MockFileSystemProvider) AddErrorCondition(operation string, path string, err error) {
	m.recorder.RecordCall("AddErrorCondition", operation, path, err)

	switch operation {
	case "read":
		m.errorConditions.ReadErrors[path] = err
	case "write":
		m.errorConditions.WriteErrors[path] = err
	case "list":
		m.errorConditions.ListErrors[path] = err
	case "stat":
		m.errorConditions.StatErrors[path] = err
	}
}

// GetOperations returns all recorded filesystem operations
func (m *MockFileSystemProvider) GetOperations() []FSOperation {
	return m.operations
}

// ReadFile reads a file from the mock filesystem
func (m *MockFileSystemProvider) ReadFile(path string) ([]byte, error) {
	m.recorder.RecordCall("ReadFile", path)

	// Check for error condition
	if err, exists := m.errorConditions.ReadErrors[path]; exists {
		m.recorder.RecordCallWithResult("ReadFile", nil, err, path)
		return nil, err
	}

	// Check if file exists
	file, exists := m.files[path]
	if !exists {
		err := os.ErrNotExist
		m.recorder.RecordCallWithResult("ReadFile", nil, err, path)
		return nil, err
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "read",
		Path:      path,
		Timestamp: time.Now(),
	})

	m.recorder.RecordCallWithResult("ReadFile", file.Content, nil, path)
	return file.Content, nil
}

// WriteFile writes a file to the mock filesystem
func (m *MockFileSystemProvider) WriteFile(path string, content []byte, mode os.FileMode) error {
	m.recorder.RecordCall("WriteFile", path, content, mode)

	// Check for error condition
	if err, exists := m.errorConditions.WriteErrors[path]; exists {
		m.recorder.RecordCallWithResult("WriteFile", nil, err, path, content, mode)
		return err
	}

	// Create or update file
	if file, exists := m.files[path]; exists {
		file.Content = content
		file.Mode = mode
		file.ModTime = time.Now()
	} else {
		m.files[path] = &MockFile{
			Name:     path,
			Content:  content,
			Metadata: make(map[string]interface{}),
			Mode:     mode,
			ModTime:  time.Now(),
			IsDir:    false,
		}
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "write",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"size": len(content),
			"mode": mode,
		},
	})

	m.recorder.RecordCallWithResult("WriteFile", nil, nil, path, content, mode)
	return nil
}

// ReadDir reads a directory from the mock filesystem
func (m *MockFileSystemProvider) ReadDir(path string) ([]os.FileInfo, error) {
	m.recorder.RecordCall("ReadDir", path)

	// Check for error condition
	if err, exists := m.errorConditions.ListErrors[path]; exists {
		m.recorder.RecordCallWithResult("ReadDir", nil, err, path)
		return nil, err
	}

	// Check if directory exists
	if _, exists := m.files[path]; !exists || !m.files[path].IsDir {
		err := os.ErrNotExist
		m.recorder.RecordCallWithResult("ReadDir", nil, err, path)
		return nil, err
	}

	// Find all files in this directory
	var fileInfos []os.FileInfo
	for filePath, file := range m.files {
		if filePath != path && filePath[:len(path)] == path {
			// This is a file in the directory
			fileInfos = append(fileInfos, &mockFileInfo{
				name:    file.Name,
				size:    int64(len(file.Content)),
				mode:    file.Mode,
				modTime: file.ModTime,
				isDir:   file.IsDir,
			})
		}
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "list",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"count": len(fileInfos),
		},
	})

	m.recorder.RecordCallWithResult("ReadDir", fileInfos, nil, path)
	return fileInfos, nil
}

// Stat returns file info for a path in the mock filesystem
func (m *MockFileSystemProvider) Stat(path string) (os.FileInfo, error) {
	m.recorder.RecordCall("Stat", path)

	// Check for error condition
	if err, exists := m.errorConditions.StatErrors[path]; exists {
		m.recorder.RecordCallWithResult("Stat", nil, err, path)
		return nil, err
	}

	// Check if file exists
	file, exists := m.files[path]
	if !exists {
		err := os.ErrNotExist
		m.recorder.RecordCallWithResult("Stat", nil, err, path)
		return nil, err
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "stat",
		Path:      path,
		Timestamp: time.Now(),
	})

	fileInfo := &mockFileInfo{
		name:    file.Name,
		size:    int64(len(file.Content)),
		mode:    file.Mode,
		modTime: file.ModTime,
		isDir:   file.IsDir,
	}

	m.recorder.RecordCallWithResult("Stat", fileInfo, nil, path)
	return fileInfo, nil
}

// mockFileInfo implements os.FileInfo for mock files
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *mockFileInfo) Name() string       { return fi.name }
func (fi *mockFileInfo) Size() int64        { return fi.size }
func (fi *mockFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *mockFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *mockFileInfo) IsDir() bool        { return fi.isDir }
func (fi *mockFileInfo) Sys() interface{}   { return nil }
