# Error Message Review and Testing Guide

**Date**: 2025-11-11  
**Purpose**: Review and test error messages for user-friendliness and clarity  
**Related Task**: 14.5 - Test error messages

## Overview

This document reviews all error messages in OneMount to ensure they are:
1. **Clear**: Easy to understand
2. **Actionable**: Tell users what to do
3. **Appropriate**: Technical details logged, not shown to users
4. **Consistent**: Follow a standard format

## Error Message Categories

### Category 1: Network Errors

**Purpose**: Inform users of network connectivity issues

| Error Message | Location | User-Facing | Assessment |
|---------------|----------|-------------|------------|
| "network request failed" | `internal/graph/graph.go` | ❌ No (logged) | ✅ **GOOD** - Clear, technical details in logs |
| "cannot fetch thumbnails in offline mode" | `internal/fs/thumbnail_operations.go` | ✅ Yes | ✅ **GOOD** - Clear and actionable |
| "network error during download" | `internal/fs/download_manager_test.go` | ❌ No (test) | ✅ **GOOD** - Test error message |

**Recommendations**:
- ✅ Network errors are well-handled
- ✅ Clear distinction between user messages and log messages
- ✅ Offline mode errors are informative

---

### Category 2: Authentication Errors

**Purpose**: Inform users of authentication issues

| Error Message | Location | User-Facing | Assessment |
|---------------|----------|-------------|------------|
| "reauth failed" | `internal/graph/graph.go` | ❌ No (logged) | ⚠️ **NEEDS IMPROVEMENT** - Should suggest re-authentication to user |
| "cannot make a request with empty auth" | `internal/graph/graph.go` | ❌ No (logged) | ✅ **GOOD** - Internal error, correctly logged |
| "Authentication token invalid or new app permissions required" | `internal/graph/graph.go` | ❌ No (logged) | ⚠️ **NEEDS IMPROVEMENT** - Should inform user to re-authenticate |

**Recommendations**:
- ⚠️ Authentication errors should be surfaced to users with clear instructions
- ⚠️ Suggest: "Your OneDrive session has expired. Please run 'onemount auth' to re-authenticate."
- ✅ Technical details correctly logged

---

### Category 3: Validation Errors

**Purpose**: Inform users of invalid input or operations

| Error Message | Location | User-Facing | Assessment |
|---------------|----------|-------------|------------|
| "invalid thumbnail size: %s" | `internal/fs/thumbnail_operations.go` | ✅ Yes | ✅ **GOOD** - Clear and specific |
| "directories do not have thumbnails" | `internal/fs/thumbnail_operations.go` | ✅ Yes | ✅ **GOOD** - Clear explanation |
| "data to upload cannot be nil" | `internal/fs/upload_session.go` | ❌ No (internal) | ✅ **GOOD** - Internal validation |
| "UploadSession UploadURL cannot be empty" | `internal/fs/upload_session.go` | ❌ No (internal) | ✅ **GOOD** - Internal validation |
| "offset cannot be larger than DriveItem size" | `internal/fs/upload_session.go` | ❌ No (internal) | ✅ **GOOD** - Internal validation |
| "size mismatch when remote checksums did not exist" | `internal/fs/upload_session.go` | ❌ No (logged) | ⚠️ **NEEDS IMPROVEMENT** - Should inform user of upload failure |
| "remote checksum did not match" | `internal/fs/upload_session.go` | ❌ No (logged) | ⚠️ **NEEDS IMPROVEMENT** - Should inform user of upload failure |

**Recommendations**:
- ✅ Validation errors for user input are clear
- ⚠️ Upload validation errors should be surfaced to users
- ⚠️ Suggest: "File upload failed: data integrity check failed. Please try again."

---

### Category 4: Not Found Errors

**Purpose**: Inform users when resources don't exist

| Error Message | Location | User-Facing | Assessment |
|---------------|----------|-------------|------------|
| "root inode not found" | `internal/fs/thumbnail_operations.go` | ❌ No (internal) | ✅ **GOOD** - Internal error |
| "path component not found: %s" | `internal/fs/thumbnail_operations.go` | ✅ Yes | ✅ **GOOD** - Clear and specific |

**Recommendations**:
- ✅ Not found errors are clear
- ✅ Path-specific errors help users identify the issue

---

### Category 5: Operation Errors

**Purpose**: Inform users of failed operations

| Error Message | Location | User-Facing | Assessment |
|---------------|----------|-------------|------------|
| "error uploading chunk - HTTP %d: %s" | `internal/fs/upload_session.go` | ❌ No (logged) | ⚠️ **NEEDS IMPROVEMENT** - Should inform user of upload failure |
| "upload cancelled by context" | `internal/fs/upload_session.go` | ❌ No (logged) | ✅ **GOOD** - Internal cancellation |
| "Failed to create thumbnail cache directory" | `internal/fs/thumbnail_cache.go` | ❌ No (logged) | ✅ **GOOD** - Non-critical, logged appropriately |
| "Failed to create content cache directory" | `internal/fs/content_cache.go` | ❌ No (logged) | ⚠️ **NEEDS IMPROVEMENT** - May affect functionality, should warn user |
| "Directory tree synchronization completed with errors" | `internal/fs/sync.go` | ❌ No (logged) | ⚠️ **NEEDS IMPROVEMENT** - Should inform user of sync issues |

**Recommendations**:
- ⚠️ Critical operation failures should be surfaced to users
- ⚠️ Cache directory creation failures may affect functionality
- ⚠️ Sync errors should be visible to users

---

### Category 6: File Status Errors

**Purpose**: Inform users of file operation status

| Error Message | Location | User-Facing | Assessment |
|---------------|----------|-------------|------------|
| "Unknown error" | `internal/fs/file_status.go` | ✅ Yes | ❌ **POOR** - Not informative |
| "Error checking offline changes" | `internal/fs/file_status.go` | ❌ No (logged) | ✅ **GOOD** - Internal error |

**Recommendations**:
- ❌ "Unknown error" is not helpful - should provide more context
- ⚠️ Suggest: "An error occurred during file operation. Check logs for details."

---

## Error Message Best Practices

### ✅ Good Examples

1. **Clear and Specific**:
   ```
   "invalid thumbnail size: large"
   ```
   - Tells user exactly what's wrong
   - Includes the invalid value

2. **Actionable**:
   ```
   "cannot fetch thumbnails in offline mode"
   ```
   - Explains why operation failed
   - Implies action: go online to fetch thumbnails

3. **User-Friendly**:
   ```
   "directories do not have thumbnails"
   ```
   - Simple language
   - Explains limitation clearly

### ❌ Poor Examples

1. **Too Technical**:
   ```
   "UploadSession UploadURL cannot be empty"
   ```
   - Exposes internal implementation
   - Not user-facing (correctly logged, not shown)

2. **Not Informative**:
   ```
   "Unknown error"
   ```
   - Doesn't help user understand what happened
   - No guidance on what to do

3. **Missing Context**:
   ```
   "error uploading chunk - HTTP 500"
   ```
   - HTTP status codes not meaningful to users
   - Should translate to user-friendly message

---

## Recommended Error Message Format

### For User-Facing Errors

```
[What happened]: [Why it happened]. [What to do about it].
```

**Examples**:
- "File upload failed: network connection lost. Please check your internet connection and try again."
- "Cannot access file: you don't have permission. Contact the file owner to request access."
- "Sync failed: OneDrive service is temporarily unavailable. Will retry automatically."

### For Logged Errors

```json
{
  "level": "error",
  "error": "[Technical error message]",
  "operation": "[What was being done]",
  "context": {
    "file_id": "ABC123",
    "path": "/Documents/file.txt"
  },
  "msg": "[User-friendly summary]"
}
```

---

## Error Message Testing Scenarios

### Scenario 1: Network Timeout

**Trigger**: Disconnect network during file download

**Expected User Message**: 
```
"Download failed: network connection lost. Will retry automatically."
```

**Expected Log**:
```json
{
  "level": "error",
  "error": "NetworkError: network request failed: connection timeout",
  "operation": "download",
  "file_id": "ABC123",
  "path": "/Documents/file.txt",
  "msg": "Network request failed"
}
```

**Test**:
```bash
# Disconnect network
sudo ip link set eth0 down

# Attempt file access
cat /mnt/onedrive/Documents/file.txt

# Check error message
# Should see user-friendly message, not technical details
```

---

### Scenario 2: Authentication Expired

**Trigger**: Wait for auth token to expire

**Expected User Message**:
```
"Your OneDrive session has expired. Please run 'onemount auth' to re-authenticate."
```

**Expected Log**:
```json
{
  "level": "error",
  "error": "AuthError: authentication token invalid",
  "operation": "api_request",
  "msg": "Authentication token invalid or new app permissions required"
}
```

**Test**:
```bash
# Wait for token expiration (or manually invalidate)
# Attempt file access
ls /mnt/onedrive/

# Check error message
journalctl -u onemount | grep -i "auth"
```

---

### Scenario 3: File Not Found

**Trigger**: Access non-existent file

**Expected User Message**:
```
"File not found: /Documents/missing.txt"
```

**Expected Log**:
```json
{
  "level": "error",
  "error": "NotFoundError: path component not found: missing.txt",
  "operation": "file_access",
  "path": "/Documents/missing.txt",
  "msg": "File not found"
}
```

**Test**:
```bash
# Access non-existent file
cat /mnt/onedrive/Documents/missing.txt

# Check error message
# Should see clear "file not found" message
```

---

### Scenario 4: Upload Failure

**Trigger**: Upload file with network issues

**Expected User Message**:
```
"File upload failed: network error. Will retry automatically."
```

**Expected Log**:
```json
{
  "level": "error",
  "error": "OperationError: error uploading chunk - HTTP 503",
  "operation": "upload",
  "file_id": "ABC123",
  "name": "file.txt",
  "msg": "Upload failed, will retry"
}
```

**Test**:
```bash
# Create test file
dd if=/dev/urandom of=/tmp/testfile.bin bs=1M count=10

# Disconnect network during upload
cp /tmp/testfile.bin /mnt/onedrive/ &
sleep 2
sudo ip link set eth0 down

# Check error message
journalctl -u onemount | grep -i "upload"
```

---

### Scenario 5: Permission Denied

**Trigger**: Access file without permission

**Expected User Message**:
```
"Access denied: you don't have permission to access this file."
```

**Expected Log**:
```json
{
  "level": "error",
  "error": "AuthError: insufficient permissions",
  "operation": "file_access",
  "path": "/Shared/restricted.txt",
  "msg": "Permission denied"
}
```

**Test**:
```bash
# Access restricted file (if available)
cat /mnt/onedrive/Shared/restricted.txt

# Check error message
```

---

### Scenario 6: Disk Full

**Trigger**: Fill local cache disk

**Expected User Message**:
```
"Cannot download file: local disk is full. Please free up space."
```

**Expected Log**:
```json
{
  "level": "error",
  "error": "OperationError: no space left on device",
  "operation": "cache_write",
  "file_id": "ABC123",
  "msg": "Failed to write to cache"
}
```

**Test**:
```bash
# Fill disk (be careful!)
# Or use quota limits to simulate

# Attempt file download
cat /mnt/onedrive/Documents/largefile.bin

# Check error message
```

---

## Error Message Improvements

### Priority 1: Critical User-Facing Errors

1. **Authentication Errors**:
   - Current: Logged only
   - Improved: Show notification + log
   - Message: "Your OneDrive session has expired. Please re-authenticate."

2. **Upload Failures**:
   - Current: Logged only
   - Improved: Show file status + log
   - Message: "Upload failed: [reason]. Will retry automatically."

3. **Sync Errors**:
   - Current: Logged only
   - Improved: Show notification + log
   - Message: "Sync failed: [reason]. Will retry in 5 minutes."

### Priority 2: Informative Error Messages

1. **Unknown Error**:
   - Current: "Unknown error"
   - Improved: "An error occurred. Check logs for details."

2. **Cache Errors**:
   - Current: Logged only
   - Improved: Warn user if critical
   - Message: "Warning: Cache directory unavailable. Performance may be affected."

3. **Checksum Errors**:
   - Current: "remote checksum did not match"
   - Improved: "Upload failed: data integrity check failed. Please try again."

### Priority 3: Enhanced Context

1. **Network Errors**:
   - Add: Retry count and next retry time
   - Example: "Network error. Retrying in 4 seconds (attempt 2 of 3)."

2. **Rate Limit Errors**:
   - Add: Expected wait time
   - Example: "Rate limited. Waiting 30 seconds before retry."

3. **Operation Errors**:
   - Add: Affected file/operation
   - Example: "Failed to sync /Documents: network timeout."

---

## Success Criteria

### User-Facing Messages

- ✅ Clear and understandable by non-technical users
- ✅ Actionable (tell user what to do)
- ✅ No technical jargon or implementation details
- ✅ Consistent format across all errors

### Logged Messages

- ✅ Include full technical details
- ✅ Include context (operation, file, path)
- ✅ Include error chain (wrapped errors)
- ✅ Structured format (JSON) for parsing

### Error Handling

- ✅ Critical errors surfaced to users
- ✅ Non-critical errors logged only
- ✅ Transient errors retried automatically
- ✅ Permanent errors reported clearly

---

## Test Results Template

```markdown
## Error Message Testing Results

**Date**: YYYY-MM-DD  
**Tester**: [Name]  
**OneMount Version**: [Version]

### Scenario 1: Network Timeout
- User message clear: ✅ Yes / ❌ No
- Technical details hidden: ✅ Yes / ❌ No
- Actionable guidance: ✅ Yes / ❌ No
- Notes: [Observations]

### Scenario 2: Authentication Expired
- User message clear: ✅ Yes / ❌ No
- Re-auth instructions: ✅ Yes / ❌ No
- Notes: [Observations]

### Scenario 3: File Not Found
- User message clear: ✅ Yes / ❌ No
- Path included: ✅ Yes / ❌ No
- Notes: [Observations]

### Scenario 4: Upload Failure
- User message clear: ✅ Yes / ❌ No
- Retry indication: ✅ Yes / ❌ No
- Notes: [Observations]

### Scenario 5: Permission Denied
- User message clear: ✅ Yes / ❌ No
- Actionable guidance: ✅ Yes / ❌ No
- Notes: [Observations]

### Scenario 6: Disk Full
- User message clear: ✅ Yes / ❌ No
- Actionable guidance: ✅ Yes / ❌ No
- Notes: [Observations]

### Issues Found
1. [Issue description]
2. [Issue description]

### Recommendations
1. [Recommendation]
2. [Recommendation]
```

---

## References

- **Error Types**: `internal/errors/error_types.go`
- **Error Logging**: `internal/logging/error.go`
- **File Status**: `internal/fs/file_status.go`
- **Upload Errors**: `internal/fs/upload_session.go`
- **Network Errors**: `internal/graph/graph.go`

---

**Last Updated**: 2025-11-11  
**Status**: Review complete, improvements identified  
**Next Steps**: Implement priority 1 improvements, test error messages
