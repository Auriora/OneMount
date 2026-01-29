# Requirements Document: File Download Hydration

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/requirements.md`. Sections below are verbatim.

## Requirements

### Requirement 3: Basic On-Demand File Access

**User Story:** As a user with limited disk space, I want files to download only when I access them so that I don't need to sync my entire OneDrive.

#### Acceptance Criteria

1. WHEN the user lists a directory, THE OneMount System SHALL display all files using cached metadata without downloading file content
2. WHEN the user opens a file that is not cached, THE OneMount System SHALL request the file content using GET `/items/{id}/content` API
3. WHEN the API returns a 302 redirect, THE OneMount System SHALL follow the redirect to download from the preauthenticated URL
4. WHEN the user opens a cached file, THE OneMount System SHALL validate the cache using ETag comparison from delta sync metadata
5. IF the cached file's ETag matches the current metadata ETag, THEN THE OneMount System SHALL serve the content from local cache
6. IF the cached file's ETag differs from the current metadata ETag, THEN THE OneMount System SHALL invalidate the cache entry and download the new content

**Note on ETag Validation Implementation**:
Requirements 3.4, 3.5, and 3.6 specify ETag-based cache validation. The implementation achieves this through delta sync rather than HTTP `if-none-match` headers because Microsoft Graph API's pre-authenticated download URLs (from `@microsoft.graph.downloadUrl`) do not support conditional GET requests. The delta sync approach:
- Proactively fetches metadata changes including updated ETags
- Invalidates cache entries when ETags change
- Triggers re-download on next file access
- Provides equivalent or better behavior than conditional GET (batch updates, proactive detection)
- Satisfies the intent of requirements 3.4, 3.5, and 3.6

### Requirement 3A: Download Status and Progress Tracking

**User Story:** As a user, I want to see the status of file downloads so that I know when files are being downloaded and if any errors occur.

#### Acceptance Criteria

1. WHILE a file is downloading, THE OneMount System SHALL update the file status to "downloading"
2. IF a download fails, THEN THE OneMount System SHALL mark the file with an error status and log the failure

### Requirement 3B: Download Manager Configuration

**User Story:** As a power user, I want to configure download behavior so that I can optimize performance for my network conditions and usage patterns.

#### Acceptance Criteria

1. WHERE the user specifies download worker pool size, THE OneMount System SHALL use the specified number of concurrent download workers
2. IF the download worker pool size is not specified, THEN THE OneMount System SHALL use a default of 3 concurrent workers
3. WHEN configuring download worker pool size, THE OneMount System SHALL validate the value is between 1 and 10 workers
4. WHERE the user specifies download retry attempts limit, THE OneMount System SHALL retry failed downloads up to the specified number of attempts
5. IF the download retry attempts limit is not specified, THEN THE OneMount System SHALL use a default of 3 retry attempts
6. WHEN configuring download retry attempts, THE OneMount System SHALL validate the value is between 1 and 10 attempts
7. WHERE the user specifies download queue size, THE OneMount System SHALL buffer up to the specified number of pending download requests
8. IF the download queue size is not specified, THEN THE OneMount System SHALL use a default queue size of 500 requests
9. WHEN configuring download queue size, THE OneMount System SHALL validate the value is between 100 and 5000 requests
10. WHERE the user specifies download chunk size for large files, THE OneMount System SHALL download files in chunks of the specified size
11. IF the download chunk size is not specified, THEN THE OneMount System SHALL use a default chunk size of 10 MB
12. WHEN configuring download chunk size, THE OneMount System SHALL validate the value is between 1 MB and 100 MB
13. WHEN download manager configuration is invalid, THE OneMount System SHALL display a clear error message with valid ranges

### Requirement 3C: File Hydration State Management

**User Story:** As a user, I want the system to manage file availability states efficiently so that I can understand which files are available locally and which need to be downloaded.

#### Acceptance Criteria

1. WHEN the metadata database reports an item in the `GHOST` state (cloud-only), THE OneMount System SHALL block file access until hydration either completes successfully or is cancelled, at which point the state SHALL transition to `HYDRATED` (success) or `ERROR` (failure) and the cache SHALL reflect the outcome
2. WHEN a hydrated file is evicted to save space, THE OneMount System SHALL transition the item back to `GHOST` without removing its metadata so that future FUSE requests can immediately rehydrate it on demand
