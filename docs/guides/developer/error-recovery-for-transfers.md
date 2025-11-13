# Error Recovery for Uploads and Downloads

This document describes the comprehensive error recovery system implemented for file uploads and downloads in OneMount. The system provides automatic recovery from interruptions, network failures, and application crashes.

## Overview

The error recovery system includes:

- **Checkpointing**: Progress is saved at chunk boundaries for both uploads and downloads
- **Resumable Transfers**: Interrupted transfers can resume from the last successful checkpoint
- **Persistent State**: Transfer state survives application restarts and crashes
- **Automatic Retry**: Failed transfers are automatically retried with exponential backoff
- **User Notifications**: Clear status updates for interrupted and recovered transfers

## Upload Recovery

### Features

1. **Chunk-level Progress Tracking**: Large files (>4MB) are uploaded in 10MB chunks with progress saved after each successful chunk
2. **Session Persistence**: Upload sessions are stored in the database and restored on application restart
3. **Intelligent Recovery**: Failed uploads attempt to resume from the last successful chunk before falling back to full restart
4. **Retry Logic**: Up to 5 retry attempts with exponential backoff for transient failures

### Implementation Details

#### UploadSession Structure

The `UploadSession` struct has been enhanced with recovery fields:

```go
type UploadSession struct {
    // ... existing fields ...
    
    // Recovery and progress tracking fields
    LastSuccessfulChunk int       `json:"lastSuccessfulChunk"`
    TotalChunks         int       `json:"totalChunks"`
    BytesUploaded       uint64    `json:"bytesUploaded"`
    LastProgressTime    time.Time `json:"lastProgressTime"`
    RecoveryAttempts    int       `json:"recoveryAttempts"`
    CanResume           bool      `json:"canResume"`
}
```

#### Recovery Process

1. **Checkpoint Creation**: After each successful chunk upload, progress is updated and persisted
2. **Failure Detection**: Network errors, server errors, and timeouts trigger recovery logic
3. **Resume Attempt**: If possible, upload resumes from the next chunk after the last successful one
4. **Fallback**: If resume fails, the upload restarts from the beginning
5. **Persistence**: All state changes are saved to the database for crash recovery

### Usage Example

```go
// Upload a large file with automatic recovery
session, err := uploadManager.QueueUploadWithPriority(inode, PriorityHigh)
if err != nil {
    return err
}

// Wait for completion (handles recovery automatically)
err = uploadManager.WaitForUpload(session.ID)
if err != nil {
    log.Printf("Upload failed after recovery attempts: %v", err)
}
```

## Download Recovery

### Features

1. **Chunk-based Downloads**: Large files are downloaded in 1MB chunks with progress tracking
2. **Resume Capability**: Interrupted downloads can resume from the last successful chunk
3. **Session Persistence**: Download sessions survive application restarts
4. **Error Recovery**: Automatic retry with exponential backoff for network failures

### Implementation Details

#### DownloadSession Structure

The `DownloadSession` struct includes recovery capabilities:

```go
type DownloadSession struct {
    // ... existing fields ...
    
    // Recovery and progress tracking fields
    Size                uint64    `json:"size"`
    BytesDownloaded     uint64    `json:"bytesDownloaded"`
    LastSuccessfulChunk int       `json:"lastSuccessfulChunk"`
    TotalChunks         int       `json:"totalChunks"`
    ChunkSize           uint64    `json:"chunkSize"`
    LastProgressTime    time.Time `json:"lastProgressTime"`
    RecoveryAttempts    int       `json:"recoveryAttempts"`
    CanResume           bool      `json:"canResume"`
}
```

#### Recovery Process

1. **Session Restoration**: On startup, incomplete download sessions are restored from the database
2. **Progress Tracking**: Each successful chunk download updates the checkpoint
3. **Resume Logic**: Failed downloads attempt to resume from the last successful chunk
4. **Cleanup**: Completed downloads are removed from the database

### Usage Example

```go
// Queue a download with automatic recovery
session, err := downloadManager.QueueDownload(fileID)
if err != nil {
    return err
}

// Wait for completion (handles recovery automatically)
err = downloadManager.WaitForDownload(fileID)
if err != nil {
    log.Printf("Download failed after recovery attempts: %v", err)
}
```

## Database Schema

The recovery system uses two database buckets:

### Uploads Bucket
- **Key**: Upload session ID
- **Value**: JSON-serialized UploadSession with recovery state

### Downloads Bucket
- **Key**: Download session ID  
- **Value**: JSON-serialized DownloadSession with recovery state

## Configuration

### Upload Recovery Settings

- **Chunk Size**: 10MB (configurable via `uploadChunkSize` constant)
- **Max Retries**: 5 attempts before giving up
- **Recovery Attempts**: Up to 3 resume attempts before full restart
- **Large File Threshold**: 4MB (files larger than this use resumable uploads)

### Download Recovery Settings

- **Chunk Size**: 1MB (configurable via `downloadChunkSize` constant)
- **Max Recovery Attempts**: 3 attempts before giving up
- **Worker Threads**: 4 concurrent download workers

## Error Handling

### Retryable Errors

The system automatically retries the following error types:
- Network connectivity issues
- Server errors (5xx HTTP status codes)
- Timeout errors
- Rate limiting (429 HTTP status code)

### Non-Retryable Errors

The following errors cause immediate failure:
- Authentication errors (401, 403)
- File not found errors (404)
- Validation errors (400)
- Quota exceeded errors

## Monitoring and Logging

### Log Messages

The system provides detailed logging for recovery operations:

```
INFO  Resuming upload from last checkpoint: chunk=5, bytes=52428800
WARN  Upload session failed, attempting recovery: attempts=2
ERROR Upload session failed too many times, cancelling: retries=5
```

### Status Updates

File status is updated throughout the recovery process:
- `StatusUploading`: Transfer in progress
- `StatusDownloading`: Download in progress  
- `StatusError`: Transfer failed after all recovery attempts
- `StatusLocal`: Transfer completed successfully

## Best Practices

1. **Monitor Transfer Progress**: Use the session APIs to track transfer progress
2. **Handle Errors Gracefully**: Always check for errors and provide user feedback
3. **Avoid Concurrent Transfers**: Don't queue multiple transfers for the same file
4. **Clean Up Resources**: The system automatically cleans up completed transfers

## Testing

The recovery system includes comprehensive tests:

- Unit tests for recovery logic (`upload_recovery_test.go`)
- Integration tests for end-to-end scenarios (`error_recovery_integration_test.go`)
- Network interruption simulation tests
- Crash recovery tests

Run tests with:
```bash
go test ./internal/fs -run TestUT_UR  # Unit tests
go test ./internal/fs -run TestIT_ER  # Integration tests
```
