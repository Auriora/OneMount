# Requirements Document: OneMount System Verification and Fix

## Introduction

This specification defines the requirements for systematically verifying and fixing the OneMount application. The goal is to ensure that the implemented code matches the documented requirements and that the application works correctly in practice, not just in unit tests.

## Glossary

- **OneMount System**: The complete OneDrive filesystem client for Linux including FUSE filesystem, Graph API integration, caching, and UI components
- **FUSE**: Filesystem in Userspace - the interface used to mount OneDrive as a native filesystem
- **Graph API**: Microsoft's REST API for accessing OneDrive resources
- **Delta Sync**: Incremental synchronization mechanism that fetches only changes from OneDrive
- **Inode**: Internal representation of a file or directory in the filesystem
- **Cache**: Local storage of file metadata and content for offline access and performance
- **D-Bus**: Inter-process communication system used for file status updates
- **Integration Test**: Test that verifies multiple components working together
- **End-to-End Test**: Test that verifies complete user workflows from start to finish

## Requirements

### Requirement 1: Authentication Verification

**User Story:** As a Linux user, I want to authenticate with my Microsoft account so that I can access my OneDrive files.

#### Acceptance Criteria

1. WHEN the user launches OneMount for the first time, THE OneMount System SHALL display an authentication dialog
2. WHEN the user completes Microsoft OAuth2 authentication, THE OneMount System SHALL store authentication tokens securely
3. WHEN authentication tokens expire, THE OneMount System SHALL automatically refresh them using the refresh token
4. IF token refresh fails, THEN THE OneMount System SHALL prompt the user to re-authenticate
5. WHERE the system is running in headless mode, THE OneMount System SHALL use device code flow for authentication

### Requirement 2: Filesystem Mounting Verification

**User Story:** As a Linux user, I want to mount my OneDrive as a local directory so that I can access files using standard file operations.

#### Acceptance Criteria

1. WHEN the user specifies a mount point, THE OneMount System SHALL mount OneDrive at that location using FUSE
2. WHEN the filesystem is mounted, THE OneMount System SHALL display the root directory contents
3. WHILE the filesystem is mounted, THE OneMount System SHALL respond to standard file operations (ls, cat, cp, etc.)
4. IF the mount point is already in use, THEN THE OneMount System SHALL display an error message with the conflicting process
5. WHEN the user unmounts the filesystem, THE OneMount System SHALL cleanly release all resources

### Requirement 3: On-Demand File Download Verification

**User Story:** As a user with limited disk space, I want files to download only when I access them so that I don't need to sync my entire OneDrive.

#### Acceptance Criteria

1. WHEN the user lists a directory, THE OneMount System SHALL display all files without downloading their content
2. WHEN the user opens a file that is not cached, THE OneMount System SHALL download the file content from OneDrive
3. WHEN the user opens a cached file, THE OneMount System SHALL serve the content from local cache without network access
4. WHILE a file is downloading, THE OneMount System SHALL update the file status to "downloading"
5. IF a download fails, THEN THE OneMount System SHALL mark the file with an error status and log the failure

### Requirement 4: File Modification and Upload Verification

**User Story:** As a user, I want to edit files locally and have changes automatically uploaded to OneDrive so that my work is synchronized.

#### Acceptance Criteria

1. WHEN the user modifies a file, THE OneMount System SHALL mark the file as having local changes
2. WHEN the user saves a modified file, THE OneMount System SHALL queue the file for upload
3. WHEN the upload queue is processed, THE OneMount System SHALL upload modified files to OneDrive
4. IF an upload fails due to network issues, THEN THE OneMount System SHALL retry with exponential backoff
5. WHEN an upload completes successfully, THE OneMount System SHALL update the file's ETag and clear the modified flag

### Requirement 5: Delta Synchronization Verification

**User Story:** As a user, I want local changes from OneDrive to be reflected automatically so that I always see the latest version of files.

#### Acceptance Criteria

1. WHILE the filesystem is mounted, THE OneMount System SHALL periodically fetch changes from OneDrive using delta queries
2. WHEN remote changes are detected, THE OneMount System SHALL update the local metadata cache
3. WHEN a remotely modified file is accessed, THE OneMount System SHALL download the new version
4. IF a file has both local and remote changes, THEN THE OneMount System SHALL create a conflict copy
5. WHEN delta sync completes, THE OneMount System SHALL store the delta link for the next sync cycle

### Requirement 6: Offline Mode Verification

**User Story:** As a user with unreliable internet, I want to access previously downloaded files when offline so that I can continue working.

#### Acceptance Criteria

1. WHEN network connectivity is lost, THE OneMount System SHALL detect the offline state
2. WHILE offline, THE OneMount System SHALL serve cached files for read operations
3. WHILE offline, THE OneMount System SHALL make the filesystem read-only
4. WHEN the user modifies a file while offline, THE OneMount System SHALL queue the changes for upload
5. WHEN network connectivity is restored, THE OneMount System SHALL process queued uploads and resume delta sync

### Requirement 7: Cache Management Verification

**User Story:** As a user, I want the cache to be managed efficiently so that it doesn't consume excessive disk space.

#### Acceptance Criteria

1. WHEN files are downloaded, THE OneMount System SHALL store content in the cache directory
2. WHEN files are accessed, THE OneMount System SHALL update the last access time in the cache
3. WHILE the cache cleanup process runs, THE OneMount System SHALL remove files older than the expiration threshold
4. WHERE cache expiration is configured, THE OneMount System SHALL respect the configured number of days
5. WHEN the user requests cache statistics, THE OneMount System SHALL display cache size, file count, and hit rate

### Requirement 8: File Status and D-Bus Integration Verification

**User Story:** As a user of Nemo/Nautilus file manager, I want to see file sync status icons so that I know which files are synced, downloading, or have errors.

#### Acceptance Criteria

1. WHEN a file status changes, THE OneMount System SHALL update the extended attributes on the file
2. WHERE D-Bus is available, THE OneMount System SHALL send status update signals via D-Bus
3. WHEN the Nemo extension queries file status, THE OneMount System SHALL provide current status information
4. IF D-Bus is unavailable, THEN THE OneMount System SHALL continue operating using extended attributes only
5. WHILE files are downloading, THE OneMount System SHALL update status to show download progress

### Requirement 9: Error Handling and Recovery Verification

**User Story:** As a user, I want the system to handle errors gracefully so that temporary issues don't cause data loss or crashes.

#### Acceptance Criteria

1. WHEN a network error occurs, THE OneMount System SHALL log the error with context
2. WHEN an API rate limit is encountered, THE OneMount System SHALL implement exponential backoff
3. IF the filesystem crashes, THEN THE OneMount System SHALL preserve state in the persistent database
4. WHEN the system restarts after a crash, THE OneMount System SHALL recover incomplete uploads and resume operations
5. WHERE errors are user-facing, THE OneMount System SHALL display helpful error messages

### Requirement 10: Performance and Concurrency Verification

**User Story:** As a user, I want the filesystem to be responsive so that file operations don't block or hang.

#### Acceptance Criteria

1. WHEN multiple files are accessed simultaneously, THE OneMount System SHALL handle concurrent operations safely
2. WHILE downloads are in progress, THE OneMount System SHALL allow other file operations to proceed
3. WHEN the user lists a large directory, THE OneMount System SHALL respond within 2 seconds
4. WHERE file operations require locks, THE OneMount System SHALL use appropriate locking granularity
5. WHEN goroutines are spawned, THE OneMount System SHALL track them with wait groups for clean shutdown

### Requirement 11: Integration Test Coverage

**User Story:** As a developer, I want comprehensive integration tests so that I can verify the system works end-to-end.

#### Acceptance Criteria

1. THE OneMount System SHALL have integration tests for the complete authentication flow
2. THE OneMount System SHALL have integration tests for file upload and download workflows
3. THE OneMount System SHALL have integration tests for offline mode transitions
4. THE OneMount System SHALL have integration tests for conflict resolution
5. THE OneMount System SHALL have integration tests for cache cleanup and expiration

### Requirement 12: Documentation Alignment

**User Story:** As a developer, I want documentation to match the actual implementation so that I can understand and maintain the code.

#### Acceptance Criteria

1. THE OneMount System SHALL have architecture documentation that accurately describes component interactions
2. THE OneMount System SHALL have design documentation that matches the implemented data models
3. THE OneMount System SHALL have API documentation that reflects actual function signatures
4. WHERE implementation differs from design, THE OneMount System SHALL document the rationale
5. WHEN code changes are made, THE OneMount System SHALL update corresponding documentation
