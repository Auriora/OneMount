// Package testutil provides testing utilities for the OneMount project.
package testutil

import (
	"os"
	"path/filepath"
	"strings"
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
	ReadErrors     map[string]error
	WriteErrors    map[string]error
	ListErrors     map[string]error
	StatErrors     map[string]error
	MkdirErrors    map[string]error
	RenameErrors   map[string]error
	ChmodErrors    map[string]error
	RemoveErrors   map[string]error
	TruncateErrors map[string]error
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
	provider := &MockFileSystemProvider{
		files:      make(map[string]*MockFile),
		operations: make([]FSOperation, 0),
		errorConditions: ErrorConditions{
			ReadErrors:     make(map[string]error),
			WriteErrors:    make(map[string]error),
			ListErrors:     make(map[string]error),
			StatErrors:     make(map[string]error),
			MkdirErrors:    make(map[string]error),
			RenameErrors:   make(map[string]error),
			ChmodErrors:    make(map[string]error),
			RemoveErrors:   make(map[string]error),
			TruncateErrors: make(map[string]error),
		},
		recorder: NewBasicMockRecorder(),
		config: MockConfig{
			CustomBehavior: make(map[string]interface{}),
		},
	}

	// Add root directory by default
	provider.AddFile("/", nil, 0755, true)

	return provider
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
		ReadErrors:     make(map[string]error),
		WriteErrors:    make(map[string]error),
		ListErrors:     make(map[string]error),
		StatErrors:     make(map[string]error),
		MkdirErrors:    make(map[string]error),
		RenameErrors:   make(map[string]error),
		ChmodErrors:    make(map[string]error),
		RemoveErrors:   make(map[string]error),
		TruncateErrors: make(map[string]error),
	}
	m.recorder = NewBasicMockRecorder()
	m.config = MockConfig{
		CustomBehavior: make(map[string]interface{}),
	}

	// Add root directory by default
	m.AddFile("/", nil, 0755, true)

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
	m.recorder.RecordCall("AddFile", path, content, mode, isDir)

	// Normalize path
	path = filepath.Clean(path)

	// Create parent directories if they don't exist
	if path != "/" {
		dir := filepath.Dir(path)
		if _, exists := m.files[dir]; !exists {
			m.AddFile(dir, nil, 0755, true)
		}
	}

	m.files[path] = &MockFile{
		Name:     path,
		Content:  content,
		Metadata: make(map[string]interface{}),
		Mode:     mode,
		ModTime:  time.Now(),
		IsDir:    isDir,
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "create",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"isDir": isDir,
			"mode":  mode,
		},
	})
}

// AddFileWithMetadata adds a file with metadata to the mock filesystem
func (m *MockFileSystemProvider) AddFileWithMetadata(path string, content []byte, metadata map[string]interface{}, mode os.FileMode, isDir bool) {
	m.recorder.RecordCall("AddFileWithMetadata", path, content, metadata, mode, isDir)

	// Normalize path
	path = filepath.Clean(path)

	// Create parent directories if they don't exist
	if path != "/" {
		dir := filepath.Dir(path)
		if _, exists := m.files[dir]; !exists {
			m.AddFile(dir, nil, 0755, true)
		}
	}

	m.files[path] = &MockFile{
		Name:     path,
		Content:  content,
		Metadata: metadata,
		Mode:     mode,
		ModTime:  time.Now(),
		IsDir:    isDir,
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "create",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"isDir":    isDir,
			"mode":     mode,
			"metadata": metadata,
		},
	})
}

// RemoveFile removes a file from the mock filesystem
func (m *MockFileSystemProvider) RemoveFile(path string) error {
	m.recorder.RecordCall("RemoveFile", path)

	// Normalize path
	path = filepath.Clean(path)

	// Check for error condition
	if err, exists := m.errorConditions.RemoveErrors[path]; exists {
		m.recorder.RecordCallWithResult("RemoveFile", nil, err, path)
		return err
	}

	// Check if file exists
	file, exists := m.files[path]
	if !exists {
		err := os.ErrNotExist
		m.recorder.RecordCallWithResult("RemoveFile", nil, err, path)
		return err
	}

	// Check if it's a directory and not empty
	if file.IsDir {
		// Check if directory is empty
		for p := range m.files {
			if p != path && strings.HasPrefix(p, path+"/") {
				err := os.ErrInvalid // Directory not empty
				m.recorder.RecordCallWithResult("RemoveFile", nil, err, path)
				return err
			}
		}
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

	// Normalize path
	path = filepath.Clean(path)

	switch operation {
	case "read":
		m.errorConditions.ReadErrors[path] = err
	case "write":
		m.errorConditions.WriteErrors[path] = err
	case "list":
		m.errorConditions.ListErrors[path] = err
	case "stat":
		m.errorConditions.StatErrors[path] = err
	case "mkdir":
		m.errorConditions.MkdirErrors[path] = err
	case "rename":
		m.errorConditions.RenameErrors[path] = err
	case "chmod":
		m.errorConditions.ChmodErrors[path] = err
	case "remove":
		m.errorConditions.RemoveErrors[path] = err
	case "truncate":
		m.errorConditions.TruncateErrors[path] = err
	}
}

// GetOperations returns all recorded filesystem operations
func (m *MockFileSystemProvider) GetOperations() []FSOperation {
	return m.operations
}

// ReadFile reads a file from the mock filesystem
func (m *MockFileSystemProvider) ReadFile(path string) ([]byte, error) {
	m.recorder.RecordCall("ReadFile", path)

	// Normalize path
	path = filepath.Clean(path)

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

	// Check if it's a directory
	if file.IsDir {
		err := os.ErrInvalid // Can't read a directory
		m.recorder.RecordCallWithResult("ReadFile", nil, err, path)
		return nil, err
	}

	// Apply latency if configured
	if m.config.Latency > 0 {
		time.Sleep(m.config.Latency)
	}

	// Simulate error based on error rate
	if m.config.ErrorRate > 0 {
		if float64(time.Now().UnixNano()%100)/100 < m.config.ErrorRate {
			err := os.ErrInvalid // Random error
			m.recorder.RecordCallWithResult("ReadFile", nil, err, path)
			return nil, err
		}
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "read",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"size": len(file.Content),
		},
	})

	m.recorder.RecordCallWithResult("ReadFile", file.Content, nil, path)
	return file.Content, nil
}

// WriteFile writes a file to the mock filesystem
func (m *MockFileSystemProvider) WriteFile(path string, content []byte, mode os.FileMode) error {
	m.recorder.RecordCall("WriteFile", path, content, mode)

	// Normalize path
	path = filepath.Clean(path)

	// Check for error condition
	if err, exists := m.errorConditions.WriteErrors[path]; exists {
		m.recorder.RecordCallWithResult("WriteFile", nil, err, path, content, mode)
		return err
	}

	// Check if parent directory exists
	dir := filepath.Dir(path)
	if dir != path { // Skip this check for root directory
		parentDir, exists := m.files[dir]
		if !exists {
			// Create parent directories
			m.AddFile(dir, nil, 0755, true)
		} else if !parentDir.IsDir {
			// Parent exists but is not a directory
			err := os.ErrInvalid
			m.recorder.RecordCallWithResult("WriteFile", nil, err, path, content, mode)
			return err
		}
	}

	// Apply latency if configured
	if m.config.Latency > 0 {
		time.Sleep(m.config.Latency)
	}

	// Simulate error based on error rate
	if m.config.ErrorRate > 0 {
		if float64(time.Now().UnixNano()%100)/100 < m.config.ErrorRate {
			err := os.ErrInvalid // Random error
			m.recorder.RecordCallWithResult("WriteFile", nil, err, path, content, mode)
			return err
		}
	}

	// Create or update file
	if file, exists := m.files[path]; exists {
		// Check if it's a directory
		if file.IsDir {
			err := os.ErrInvalid // Can't write to a directory
			m.recorder.RecordCallWithResult("WriteFile", nil, err, path, content, mode)
			return err
		}

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

	// Normalize path
	path = filepath.Clean(path)

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

	// Find all files in this directory (direct children only)
	var fileInfos []os.FileInfo
	for filePath, file := range m.files {
		// Skip the directory itself
		if filePath == path {
			continue
		}

		// Check if this is a direct child of the directory
		dir := filepath.Dir(filePath)
		if dir == path {
			// This is a direct child of the directory
			fileInfos = append(fileInfos, &mockFileInfo{
				name:    filepath.Base(file.Name),
				size:    int64(len(file.Content)),
				mode:    file.Mode,
				modTime: file.ModTime,
				isDir:   file.IsDir,
			})
		}
	}

	// Apply latency if configured
	if m.config.Latency > 0 {
		time.Sleep(m.config.Latency)
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

	// Normalize path
	path = filepath.Clean(path)

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

	// Apply latency if configured
	if m.config.Latency > 0 {
		time.Sleep(m.config.Latency)
	}

	// Simulate error based on error rate
	if m.config.ErrorRate > 0 {
		if float64(time.Now().UnixNano()%100)/100 < m.config.ErrorRate {
			err := os.ErrInvalid // Random error
			m.recorder.RecordCallWithResult("Stat", nil, err, path)
			return nil, err
		}
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "stat",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"isDir": file.IsDir,
			"size":  len(file.Content),
			"mode":  file.Mode,
		},
	})

	fileInfo := &mockFileInfo{
		name:    filepath.Base(file.Name),
		size:    int64(len(file.Content)),
		mode:    file.Mode,
		modTime: file.ModTime,
		isDir:   file.IsDir,
	}

	m.recorder.RecordCallWithResult("Stat", fileInfo, nil, path)
	return fileInfo, nil
}

// Mkdir creates a directory in the mock filesystem
func (m *MockFileSystemProvider) Mkdir(path string, mode os.FileMode) error {
	m.recorder.RecordCall("Mkdir", path, mode)

	// Normalize path
	path = filepath.Clean(path)

	// Check for error condition
	if err, exists := m.errorConditions.MkdirErrors[path]; exists {
		m.recorder.RecordCallWithResult("Mkdir", nil, err, path, mode)
		return err
	}

	// Check if file already exists
	if file, exists := m.files[path]; exists {
		if file.IsDir {
			// Directory already exists
			err := os.ErrExist
			m.recorder.RecordCallWithResult("Mkdir", nil, err, path, mode)
			return err
		}
		// Path exists but is a file
		err := os.ErrInvalid
		m.recorder.RecordCallWithResult("Mkdir", nil, err, path, mode)
		return err
	}

	// Create parent directories if they don't exist
	if path != "/" {
		dir := filepath.Dir(path)
		if _, exists := m.files[dir]; !exists {
			if err := m.Mkdir(dir, 0755); err != nil {
				m.recorder.RecordCallWithResult("Mkdir", nil, err, path, mode)
				return err
			}
		}
	}

	// Apply latency if configured
	if m.config.Latency > 0 {
		time.Sleep(m.config.Latency)
	}

	// Create directory
	m.files[path] = &MockFile{
		Name:     path,
		Content:  nil,
		Metadata: make(map[string]interface{}),
		Mode:     mode | os.ModeDir,
		ModTime:  time.Now(),
		IsDir:    true,
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "mkdir",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"mode": mode,
		},
	})

	m.recorder.RecordCallWithResult("Mkdir", nil, nil, path, mode)
	return nil
}

// Rename renames a file or directory in the mock filesystem
func (m *MockFileSystemProvider) Rename(oldPath, newPath string) error {
	m.recorder.RecordCall("Rename", oldPath, newPath)

	// Normalize paths
	oldPath = filepath.Clean(oldPath)
	newPath = filepath.Clean(newPath)

	// Check for error condition
	if err, exists := m.errorConditions.RenameErrors[oldPath]; exists {
		m.recorder.RecordCallWithResult("Rename", nil, err, oldPath, newPath)
		return err
	}

	// Check if source file exists
	file, exists := m.files[oldPath]
	if !exists {
		err := os.ErrNotExist
		m.recorder.RecordCallWithResult("Rename", nil, err, oldPath, newPath)
		return err
	}

	// Check if destination already exists
	if _, exists := m.files[newPath]; exists {
		err := os.ErrExist
		m.recorder.RecordCallWithResult("Rename", nil, err, oldPath, newPath)
		return err
	}

	// Create parent directories for destination if they don't exist
	if newPath != "/" {
		dir := filepath.Dir(newPath)
		if _, exists := m.files[dir]; !exists {
			if err := m.Mkdir(dir, 0755); err != nil {
				m.recorder.RecordCallWithResult("Rename", nil, err, oldPath, newPath)
				return err
			}
		}
	}

	// Apply latency if configured
	if m.config.Latency > 0 {
		time.Sleep(m.config.Latency)
	}

	// If it's a directory, we need to rename all children too
	if file.IsDir {
		// Find all children
		childPaths := []string{}
		for path := range m.files {
			if path != oldPath && strings.HasPrefix(path, oldPath+"/") {
				childPaths = append(childPaths, path)
			}
		}

		// Rename all children
		for _, childPath := range childPaths {
			relativePath := strings.TrimPrefix(childPath, oldPath)
			childNewPath := filepath.Join(newPath, relativePath)
			childFile := m.files[childPath]

			// Create the new file
			m.files[childNewPath] = &MockFile{
				Name:     childNewPath,
				Content:  childFile.Content,
				Metadata: childFile.Metadata,
				Mode:     childFile.Mode,
				ModTime:  childFile.ModTime,
				IsDir:    childFile.IsDir,
			}

			// Delete the old file
			delete(m.files, childPath)
		}
	}

	// Create the new file
	m.files[newPath] = &MockFile{
		Name:     newPath,
		Content:  file.Content,
		Metadata: file.Metadata,
		Mode:     file.Mode,
		ModTime:  time.Now(),
		IsDir:    file.IsDir,
	}

	// Delete the old file
	delete(m.files, oldPath)

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "rename",
		Path:      oldPath,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"newPath": newPath,
		},
	})

	m.recorder.RecordCallWithResult("Rename", nil, nil, oldPath, newPath)
	return nil
}

// Chmod changes the mode of a file in the mock filesystem
func (m *MockFileSystemProvider) Chmod(path string, mode os.FileMode) error {
	m.recorder.RecordCall("Chmod", path, mode)

	// Normalize path
	path = filepath.Clean(path)

	// Check for error condition
	if err, exists := m.errorConditions.ChmodErrors[path]; exists {
		m.recorder.RecordCallWithResult("Chmod", nil, err, path, mode)
		return err
	}

	// Check if file exists
	file, exists := m.files[path]
	if !exists {
		err := os.ErrNotExist
		m.recorder.RecordCallWithResult("Chmod", nil, err, path, mode)
		return err
	}

	// Apply latency if configured
	if m.config.Latency > 0 {
		time.Sleep(m.config.Latency)
	}

	// Change mode
	if file.IsDir {
		file.Mode = mode | os.ModeDir
	} else {
		file.Mode = mode
	}

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "chmod",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"mode": mode,
		},
	})

	m.recorder.RecordCallWithResult("Chmod", nil, nil, path, mode)
	return nil
}

// Truncate truncates a file in the mock filesystem
func (m *MockFileSystemProvider) Truncate(path string, size int64) error {
	m.recorder.RecordCall("Truncate", path, size)

	// Normalize path
	path = filepath.Clean(path)

	// Check for error condition
	if err, exists := m.errorConditions.TruncateErrors[path]; exists {
		m.recorder.RecordCallWithResult("Truncate", nil, err, path, size)
		return err
	}

	// Check if file exists
	file, exists := m.files[path]
	if !exists {
		err := os.ErrNotExist
		m.recorder.RecordCallWithResult("Truncate", nil, err, path, size)
		return err
	}

	// Check if it's a directory
	if file.IsDir {
		err := os.ErrInvalid // Can't truncate a directory
		m.recorder.RecordCallWithResult("Truncate", nil, err, path, size)
		return err
	}

	// Apply latency if configured
	if m.config.Latency > 0 {
		time.Sleep(m.config.Latency)
	}

	// Truncate file
	currentSize := int64(len(file.Content))
	if size < currentSize {
		file.Content = file.Content[:size]
	} else if size > currentSize {
		newContent := make([]byte, size)
		copy(newContent, file.Content)
		file.Content = newContent
	}
	file.ModTime = time.Now()

	// Record operation
	m.operations = append(m.operations, FSOperation{
		Type:      "truncate",
		Path:      path,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"size": size,
		},
	})

	m.recorder.RecordCallWithResult("Truncate", nil, nil, path, size)
	return nil
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
