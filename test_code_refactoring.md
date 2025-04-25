# Test Code Refactoring Recommendations

This document outlines recommendations for refactoring the test code in the onedriver project. These recommendations aim to improve the consistency, reliability, and maintainability of the tests.

## Common Issues Identified

After reviewing the test code across the project, the following issues were identified:

1. **Inconsistent Test Patterns**:
   - Some tests use `t.Parallel()` while others don't
   - Some tests use `t.Cleanup()` while others manually clean up resources
   - Some tests use `require` assertions while others use `assert` or standard Go conditionals
   - Inconsistent error handling approaches

2. **Duplicate Code**:
   - Similar setup and teardown code repeated across test files
   - Common test utilities reimplemented in multiple places
   - Similar test patterns duplicated across test functions

3. **Lack of Test Helpers**:
   - Few reusable test helper functions for common operations
   - Limited use of test fixtures or test data generators
   - Manual setup of test preconditions in each test

4. **Brittle Tests**:
   - Many tests rely on fixed timeouts and sleeps
   - Race conditions in some tests (e.g., `TestUploadDiskSerialization`)
   - Tests that depend on specific file content or state from previous tests

5. **Insufficient Error Handling**:
   - Some tests don't properly check error conditions
   - Limited error context in failure messages
   - Inconsistent approach to handling expected errors

6. **Poor Test Organization**:
   - Long test functions with multiple assertions
   - Limited use of subtests for organizing related test cases
   - Unclear test naming conventions

## Refactoring Recommendations

### 1. Standardize Test Patterns

#### 1.1 Consistent Use of Test Parallelization

Example:
```
// Use t.Parallel() consistently for tests that can run in parallel
func TestSomething(t *testing.T) {
    t.Parallel()
    // Test logic
}

// Document when tests cannot run in parallel and why
func TestSomethingSequential(t *testing.T) {
    // Cannot use t.Parallel() because this test modifies global state
    // Test logic
}
```

#### 1.2 Consistent Resource Cleanup

Example:
```
// Use t.Cleanup() consistently for resource cleanup
func TestWithResourceCleanup(t *testing.T) {
    resource := createResource()
    
    t.Cleanup(func() {
        if err := cleanupResource(resource); err != nil {
            t.Logf("Warning: Failed to clean up resource: %v", err)
        }
    })
    
    // Test logic
}
```

#### 1.3 Consistent Assertion Style

Example:
```
// Use require for assertions that should terminate the test on failure
require.NoError(t, err, "Failed to create resource")
require.NotNil(t, result, "Result should not be nil")

// Use assert for assertions that should not terminate the test
assert.Equal(t, expected, actual, "Values should be equal")
assert.True(t, condition, "Condition should be true")
```

### 2. Extract Common Test Utilities

#### 2.1 Create a Common Test Utilities Package

Example:
```
// testutil/file.go
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
```

#### 2.2 Create Test Fixtures

Example:
```
// testutil/fixtures.go
package testutil

import (
    // Import the graph package
)

// StandardTestFile returns a standard test file with predictable content
func StandardTestFile() []byte {
    return []byte("This is a standard test file content")
}

// CreateDriveItemFixture creates a DriveItem fixture for testing
func CreateDriveItemFixture(name string, isFolder bool) *DriveItem {
    // Create and return a DriveItem with standard test values
    return &DriveItem{
        // Initialize with test values
    }
}
```

### 3. Implement Test Helpers

#### 3.1 Waiting for Asynchronous Operations

Example:
```
// testutil/async.go
package testutil

import (
    "testing"
    "time"
)

// WaitForCondition waits for a condition to be true with a configurable timeout and polling interval
func WaitForCondition(t *testing.T, condition func() bool, timeout, pollInterval time.Duration, message string) {
    deadline := time.Now().Add(timeout)
    
    for time.Now().Before(deadline) {
        if condition() {
            return
        }
        time.Sleep(pollInterval)
    }
    
    t.Fatalf("Timed out waiting for condition: %s", message)
}

// Example usage:
// WaitForCondition(t, func() bool {
//     _, err := os.Stat(path)
//     return err == nil
// }, 5*time.Second, 100*time.Millisecond, "File was not created within timeout")
```

#### 3.2 File System Operations

Example:
```
// testutil/fs.go
package testutil

import (
    "os"
    "path/filepath"
    "testing"
)

// EnsureDirectoryExists ensures a directory exists and is cleaned up after the test
func EnsureDirectoryExists(t *testing.T, path string) {
    err := os.MkdirAll(path, 0755)
    if err != nil {
        t.Fatalf("Failed to create directory %s: %v", path, err)
    }
    
    t.Cleanup(func() {
        if err := os.RemoveAll(path); err != nil {
            t.Logf("Warning: Failed to clean up directory %s: %v", path, err)
        }
    })
}

// CaptureFileSystemState captures the current state of a directory
func CaptureFileSystemState(t *testing.T, dir string) map[string]os.FileInfo {
    state := make(map[string]os.FileInfo)
    
    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        state[path] = info
        return nil
    })
    
    if err != nil {
        t.Fatalf("Failed to capture filesystem state: %v", err)
    }
    
    return state
}
```

### 4. Improve Test Reliability

#### 4.1 Replace Fixed Timeouts with Dynamic Waiting

Example:
```
// Instead of:
time.Sleep(2 * time.Second)
if err != nil {
    t.Fatal(err)
}

// Use:
WaitForCondition(t, func() bool {
    _, err := os.Stat(path)
    return err == nil
}, 5*time.Second, 100*time.Millisecond, "File was not created within timeout")
```

#### 4.2 Fix Race Conditions

Example:
```
// For TestUploadDiskSerialization, rewrite to avoid race conditions:
func TestUploadDiskSerialization(t *testing.T) {
    // Create a file with known content
    filePath := filepath.Join(TestDir, "upload_to_disk.fa")
    fileContent := []byte("Test content for upload serialization")
    err := os.WriteFile(filePath, fileContent, 0644)
    if err != nil {
        t.Fatalf("Failed to create test file: %v", err)
    }
    
    // Wait for the file to be recognized by the filesystem
    var inode *Inode
    WaitForCondition(t, func() bool {
        var err error
        inode, err = fs.GetPath("/onedriver_tests/upload_to_disk.fa", nil)
        return err == nil && inode != nil
    }, 10*time.Second, 500*time.Millisecond, "File was not recognized by filesystem")
    
    // Verify upload session was created
    var session UploadSession
    var found bool
    WaitForCondition(t, func() bool {
        session, found = findUploadSession(fs.db, inode.ID())
        return found
    }, 10*time.Second, 500*time.Millisecond, "Upload session was not created")
    
    // Rest of the test...
}

func findUploadSession(db *bolt.DB, inodeID string) (UploadSession, bool) {
    var session UploadSession
    var found bool
    
    _ = db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket(bucketUploads)
        if b == nil {
            return nil
        }
        
        data := b.Get([]byte(inodeID))
        if data == nil {
            return nil
        }
        
        if err := json.Unmarshal(data, &session); err != nil {
            return err
        }
        
        found = true
        return nil
    })
    
    return session, found
}
```

#### 4.3 Isolate Tests from Each Other

Example:
```
// Use subtests to isolate test cases
func TestFileOperations(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        t.Parallel()
        // Test file creation
    })
    
    t.Run("Read", func(t *testing.T) {
        t.Parallel()
        // Test file reading
    })
    
    t.Run("Update", func(t *testing.T) {
        t.Parallel()
        // Test file updating
    })
    
    t.Run("Delete", func(t *testing.T) {
        t.Parallel()
        // Test file deletion
    })
}
```

### 5. Improve Error Handling

#### 5.1 Provide Context in Error Messages

Example:
```
// Instead of:
if err != nil {
    t.Fatal(err)
}

// Use:
if err != nil {
    t.Fatalf("Failed to create file %s: %v", path, err)
}
```

#### 5.2 Test Error Conditions Explicitly

Example:
```
// Test that errors are handled correctly
func TestErrorHandling(t *testing.T) {
    t.Run("NonexistentFile", func(t *testing.T) {
        t.Parallel()
        
        // Attempt to read a nonexistent file
        _, err := os.ReadFile("/nonexistent/file")
        
        // Verify the error is of the expected type
        if err == nil {
            t.Fatal("Expected an error, got nil")
        }
        if !os.IsNotExist(err) {
            t.Fatalf("Expected IsNotExist error, got: %v", err)
        }
    })
    
    // More error condition tests...
}
```

### 6. Improve Test Organization

#### 6.1 Use Table-Driven Tests

Example:
```
func TestFilePermissions(t *testing.T) {
    testCases := []struct {
        name        string
        permissions os.FileMode
        readable    bool
        writable    bool
        executable  bool
    }{
        {
            name:        "ReadOnly",
            permissions: 0444,
            readable:    true,
            writable:    false,
            executable:  false,
        },
        {
            name:        "ReadWrite",
            permissions: 0644,
            readable:    true,
            writable:    true,
            executable:  false,
        },
        {
            name:        "ReadWriteExecute",
            permissions: 0755,
            readable:    true,
            writable:    true,
            executable:  true,
        },
    }
    
    for _, tc := range testCases {
        tc := tc // Capture range variable
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()
            
            // Create a file with the specified permissions
            filePath := filepath.Join(TestDir, "permissions_"+tc.name)
            err := os.WriteFile(filePath, []byte("test"), 0644)
            if err != nil {
                t.Fatalf("Failed to create test file: %v", err)
            }
            err = os.Chmod(filePath, tc.permissions)
            if err != nil {
                t.Fatalf("Failed to set permissions: %v", err)
            }
            
            t.Cleanup(func() {
                if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
                    t.Logf("Warning: Failed to clean up test file %s: %v", filePath, err)
                }
            })
            
            // Test the permissions
            info, err := os.Stat(filePath)
            if err != nil {
                t.Fatalf("Failed to stat file: %v", err)
            }
            
            mode := info.Mode()
            if (mode&0444 != 0) != tc.readable {
                t.Errorf("Readable permission mismatch: got %v, want %v", mode&0444 != 0, tc.readable)
            }
            if (mode&0222 != 0) != tc.writable {
                t.Errorf("Writable permission mismatch: got %v, want %v", mode&0222 != 0, tc.writable)
            }
            if (mode&0111 != 0) != tc.executable {
                t.Errorf("Executable permission mismatch: got %v, want %v", mode&0111 != 0, tc.executable)
            }
        })
    }
}
```

#### 6.2 Group Related Tests

Example:
```
// Group related tests in the same file with clear naming conventions
// file_operations_test.go
func TestFileCreate(t *testing.T) { /* ... */ }
func TestFileRead(t *testing.T) { /* ... */ }
func TestFileUpdate(t *testing.T) { /* ... */ }
func TestFileDelete(t *testing.T) { /* ... */ }

// directory_operations_test.go
func TestDirectoryCreate(t *testing.T) { /* ... */ }
func TestDirectoryRead(t *testing.T) { /* ... */ }
func TestDirectoryDelete(t *testing.T) { /* ... */ }
```

#### 6.3 Use Clear Test Names

Example:
```
// Instead of:
func TestDeltaMkdir(t *testing.T) { /* ... */ }

// Use:
func TestDelta_CreateDirectoryOnServer_ShouldSyncToClient(t *testing.T) { /* ... */ }
```

## Implementation Plan

1. **Create Test Utilities Package**:
   - Create a new package `testutil` with common test utilities
   - Move common test code to this package
   - Update existing tests to use the new utilities

2. **Standardize Test Patterns**:
   - Update tests to use consistent patterns for parallelization, cleanup, and assertions
   - Document when and why tests deviate from these patterns

3. **Improve Test Reliability**:
   - Replace fixed timeouts with dynamic waiting
   - Fix race conditions in tests
   - Isolate tests from each other

4. **Improve Error Handling**:
   - Add context to error messages
   - Test error conditions explicitly

5. **Improve Test Organization**:
   - Convert appropriate tests to table-driven tests
   - Group related tests
   - Use clear test names

## Conclusion

Implementing these refactoring recommendations will improve the consistency, reliability, and maintainability of the tests in the onedriver project. This will make it easier to add new tests, modify existing tests, and understand test failures.