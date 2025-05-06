

# Junie Prompt: Redesign Upload API for Robust Session Handling

## Task Overview
Implement Solution 5 from the race condition analysis to redesign the Upload API, making it more robust by enhancing the `WaitForUpload` method to handle cases where a session hasn't been added to the sessions map yet.

## Current Issue
There's a race condition in the `UploadManager` between queuing an upload and waiting for it. The `WaitForUpload` method checks if the upload session exists in the `sessions` map, but this map is only populated when the session is processed by the `uploadLoop`, which runs on a ticker. This causes test failures when `WaitForUpload` is called immediately after `QueueUploadWithPriority`.

## Proposed Changes

### 1. Requirements Impact Analysis
- No changes to the Software Requirements Specification (SRS) are needed
- This change improves the robustness of the API without changing its functional requirements
- The change maintains backward compatibility with existing code
- The change addresses a race condition that could affect reliability in production environments

### 2. Architecture Document Changes
Add the following to the architecture document:

```markdown
#### Upload Manager Session Handling
The UploadManager now includes enhanced session handling to prevent race conditions between queuing uploads and waiting for them to complete. The `WaitForUpload` method has been improved to handle cases where a session hasn't been processed by the upload loop yet, making the API more resilient to timing issues.
```

### 3. Design Documentation Changes
Update the design documentation with:

```markdown
#### Upload API Robustness Improvements
The Upload API has been enhanced to handle race conditions between session creation and waiting:

1. `WaitForUpload` now includes a waiting period for session creation
2. A new helper method `GetSession` provides thread-safe access to session information
3. Timeout mechanisms prevent indefinite waiting for sessions that may never be created
4. Error messages are more descriptive to help diagnose issues
```

### 4. Implementation Details

Modify `upload_manager.go` to include:

1. Add a new `GetSession` method:
```go
// GetSession returns the upload session for the given ID if it exists
func (u *UploadManager) GetSession(id string) (*UploadSession, bool) {
    u.mutex.RLock()
    defer u.mutex.RUnlock()
    session, exists := u.sessions[id]
    return session, exists
}
```

2. Enhance the `WaitForUpload` method:
```go
// WaitForUpload waits for an upload to complete
func (u *UploadManager) WaitForUpload(id string) error {
    // First, check if the session exists
    _, exists := u.GetSession(id)
    if !exists {
        // If not, wait for it to be created (with timeout)
        deadline := time.Now().Add(5 * time.Second)
        for time.Now().Before(deadline) {
            _, exists := u.GetSession(id)
            if exists {
                break
            }
            time.Sleep(10 * time.Millisecond)
        }
        
        // Final check after waiting
        _, exists := u.GetSession(id)
        if !exists {
            return errors.New("upload session not found: Failed to wait for upload")
        }
    }
    
    // Now wait for the upload to complete
    for {
        session, exists := u.GetSession(id)
        if !exists {
            return errors.New("upload session disappeared during wait")
        }
        
        state := session.getState()
        switch state {
        case uploadComplete:
            return nil
        case uploadErrored:
            return session.error
        default:
            // Still in progress, wait a bit
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```

### 5. Refactoring Dependent Code
No changes to dependent code are required as this implementation maintains the same API signature and behavior, only making it more robust.

### 6. Testing Strategy

#### Unit Tests
1. Test `WaitForUpload` with a session that already exists
2. Test `WaitForUpload` with a session that doesn't exist yet but is added shortly after
3. Test `WaitForUpload` with a session that is never added (should timeout)
4. Test `WaitForUpload` with a session that is removed during waiting

```go
func TestWaitForUpload_SessionAlreadyExists(t *testing.T) {
    // Setup test with a session already in the sessions map
    // Call WaitForUpload
    // Verify it waits correctly for completion
}

func TestWaitForUpload_SessionAddedLater(t *testing.T) {
    // Setup test
    // Start a goroutine that adds the session after a short delay
    // Call WaitForUpload
    // Verify it waits for the session to be added and then for completion
}

func TestWaitForUpload_SessionNeverAdded(t *testing.T) {
    // Setup test
    // Call WaitForUpload with an ID that will never be added
    // Verify it times out with the correct error message
}

func TestWaitForUpload_SessionDisappearsDuringWait(t *testing.T) {
    // Setup test with a session in the sessions map
    // Start a goroutine that removes the session after a short delay
    // Call WaitForUpload
    // Verify it returns the correct error
}
```

#### Integration Tests
1. Fix the existing `TestUT_FS_05_02_RepeatedUploads_OnlineMode_SuccessfulUpload` test without adding delays
2. Add a stress test that rapidly queues and waits for multiple uploads

```go
func TestStressUploadQueueAndWait(t *testing.T) {
    // Setup test
    // Queue multiple uploads in rapid succession
    // Wait for all of them to complete
    // Verify all completed successfully
}
```

### 7. Implementation Considerations
- The timeout value (5 seconds) should be configurable or at least carefully chosen
- Error messages should be descriptive to help diagnose issues
- Consider adding logging to help debug timing issues
- The implementation should be thread-safe and handle concurrent calls to `WaitForUpload`

### 8. Backward Compatibility
This change maintains backward compatibility with existing code that uses `WaitForUpload`, as it only adds functionality without changing the method signature or expected behavior.

## Expected Outcome
After implementing these changes:
1. The race condition in tests will be resolved without adding unreliable delays
2. The API will be more robust for all users, not just in tests
3. Error messages will be more descriptive
4. The code will handle edge cases more gracefully

## Acceptance Criteria
1. All existing tests pass without modifications
2. New tests for the enhanced functionality pass
3. No regression in performance or functionality
4. Code review confirms thread safety and proper error handling