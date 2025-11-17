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
- **Docker Container**: Isolated environment for running tests without affecting the host system
- **Test Runner**: Docker container configured with all dependencies needed to run OneMount tests
- **BBolt**: Embedded key/value database used for persistent storage of metadata and state

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
2. WHEN the filesystem is mounted for the first time, THE OneMount System SHALL fetch and cache the complete directory structure from OneDrive without blocking interactive operations; commands SHALL use whatever metadata is already cached while the remaining tree sync runs in the background
3. WHEN the filesystem is mounted, THE OneMount System SHALL display the root directory contents
4. WHILE the filesystem is mounted, THE OneMount System SHALL respond to standard file operations (ls, cat, cp, etc.)
5. WHEN the user navigates directories, THE OneMount System SHALL serve directory listings from the cached metadata without network requests; if cached metadata exists but is older than the refresh threshold, THE OneMount System SHALL return the cached data immediately and trigger a refresh asynchronously
6. WHEN a directory lookup fails (including typos, case mismatches, or maintenance of virtual files such as `.xdg-volume-info`), THE OneMount System SHALL scope cache invalidation to the affected entry rather than clearing the entire parent directory cache
7. IF the mount point is already in use, THEN THE OneMount System SHALL display an error message with the conflicting process
8. WHEN the user unmounts the filesystem, THE OneMount System SHALL cleanly release all resources
9. WHERE the user specifies daemon mode, THE OneMount System SHALL fork the process and detach from the terminal for background operation
10. WHEN the user specifies a mount timeout, THE OneMount System SHALL wait up to the specified duration for the mount operation to complete
11. IF the mount timeout is not specified, THEN THE OneMount System SHALL use a default timeout of 60 seconds
12. WHEN opening the metadata database, THE OneMount System SHALL detect stale lock files older than 5 minutes and attempt to remove them
13.IF a database lock file is detected and is not stale, THEN THE OneMount System SHALL retry with exponential backoff up to 10 attempts

### Requirement 3: On-Demand File Download Verification

**User Story:** As a user with limited disk space, I want files to download only when I access them so that I don't need to sync my entire OneDrive.

#### Acceptance Criteria

1. WHEN the user lists a directory, THE OneMount System SHALL display all files using cached metadata without downloading file content
2. WHEN the user opens a file that is not cached, THE OneMount System SHALL request the file content using GET `/items/{id}/content` API
3. WHEN the API returns a 302 redirect, THE OneMount System SHALL follow the redirect to download from the preauthenticated URL
4. WHEN the user opens a cached file, THE OneMount System SHALL validate the cache using ETag comparison from delta sync metadata
5. IF the cached file's ETag matches the current metadata ETag, THEN THE OneMount System SHALL serve the content from local cache
6. IF the cached file's ETag differs from the current metadata ETag, THEN THE OneMount System SHALL invalidate the cache entry and download the new content
7. WHILE a file is downloading, THE OneMount System SHALL update the file status to "downloading"
8. IF a download fails, THEN THE OneMount System SHALL mark the file with an error status and log the failure
9. WHERE the user specifies download worker pool size, THE OneMount System SHALL use the specified number of concurrent download workers
10. IF the download worker pool size is not specified, THEN THE OneMount System SHALL use a default of 3 concurrent workers

**Note on ETag Validation Implementation**:
Requirements 3.4, 3.5, and 3.6 specify ETag-based cache validation. The implementation achieves this through delta sync rather than HTTP `if-none-match` headers because Microsoft Graph API's pre-authenticated download URLs (from `@microsoft.graph.downloadUrl`) do not support conditional GET requests. The delta sync approach:
- Proactively fetches metadata changes including updated ETags
- Invalidates cache entries when ETags change
- Triggers re-download on next file access
- Provides equivalent or better behavior than conditional GET (batch updates, proactive detection)
- Satisfies the intent of requirements 3.4, 3.5, and 3.6

11. WHEN configuring download worker pool size, THE OneMount System SHALL validate the value is between 1 and 10 workers
12. WHERE the user specifies download retry attempts limit, THE OneMount System SHALL retry failed downloads up to the specified number of attempts
13. IF the download retry attempts limit is not specified, THEN THE OneMount System SHALL use a default of 3 retry attempts
14. WHEN configuring download retry attempts, THE OneMount System SHALL validate the value is between 1 and 10 attempts
15. WHERE the user specifies download queue size, THE OneMount System SHALL buffer up to the specified number of pending download requests
16. IF the download queue size is not specified, THEN THE OneMount System SHALL use a default queue size of 500 requests
17. WHEN configuring download queue size, THE OneMount System SHALL validate the value is between 100 and 5000 requests
18. WHERE the user specifies download chunk size for large files, THE OneMount System SHALL download files in chunks of the specified size
19. IF the download chunk size is not specified, THEN THE OneMount System SHALL use a default chunk size of 10 MB
20. WHEN configuring download chunk size, THE OneMount System SHALL validate the value is between 1 MB and 100 MB
21. WHEN download manager configuration is invalid, THE OneMount System SHALL display a clear error message with valid ranges

### Requirement 4: File Modification and Upload Verification

**User Story:** As a user, I want to edit files locally and have changes automatically uploaded to OneDrive so that my work is synchronized.

#### Acceptance Criteria

1. WHEN the user modifies a file, THE OneMount System SHALL mark the file as having local changes
2. WHEN the user saves a modified file, THE OneMount System SHALL queue the file for upload
3. WHEN uploading a file smaller than 250 MB, THE OneMount System SHALL use PUT `/items/{id}/content` with the file content
4. WHEN uploading a file larger than 250 MB, THE OneMount System SHALL create an upload session using POST `/createUploadSession`
5. WHEN using an upload session, THE OneMount System SHALL upload the file in chunks to the session URL
6. IF an upload fails due to network issues, THEN THE OneMount System SHALL retry with exponential backoff
7. WHEN an upload completes successfully, THE OneMount System SHALL update the file's ETag from the response
8. WHEN an upload completes successfully, THE OneMount System SHALL clear the modified flag
9. WHEN the user creates a directory, THE OneMount System SHALL create the directory on the server and assign it a unique ID
10. WHEN the user deletes an empty directory using Rmdir, THE OneMount System SHALL remove the directory from the server
11. IF the user attempts to delete a non-empty directory, THEN THE OneMount System SHALL return ENOTEMPTY error
12. WHEN a directory is deleted, THE OneMount System SHALL remove the directory from the parent's children list
13. WHEN a directory is deleted, THE OneMount System SHALL remove the directory inode from the filesystem's internal tracking

### Requirement 5: Delta Synchronization Verification

**User Story:** As a user, I want local changes from OneDrive to be reflected automatically so that I always see the latest version of files.

#### Acceptance Criteria

1. WHEN the filesystem is first mounted, THE OneMount System SHALL fetch the complete directory structure from OneDrive using the delta API
2. WHEN the filesystem is mounted, THE OneMount System SHALL attempt to establish a Microsoft Graph Socket.IO subscription for the mounted drive
3. WHEN establishing a subscription for personal OneDrive, THE OneMount System SHALL target the root folder or selected subfolders, matching Graph's supported resources
4. WHEN establishing a subscription for OneDrive for Business, THE OneMount System SHALL limit the subscription scope to the drive root as required by Graph
5. WHEN a Socket.IO subscription is healthy, THE OneMount System SHALL run delta polling no more frequently than every 30 minutes (configurable but never lower than 5 minutes) and SHALL log any deviation from that cadence
6. WHEN a Socket.IO notification is received, THE OneMount System SHALL immediately trigger a delta query to fetch changes and SHALL preempt lower-priority metadata work so user-facing operations do not stall
7. WHEN the Socket.IO subscription is unavailable or unhealthy, THE OneMount System SHALL automatically fall back to delta polling every 5 minutes by default and SHALL log the degraded state
8. IF the Socket.IO channel continues to fail or error, THEN THE OneMount System MAY temporarily shorten the polling interval down to 10 seconds to recover, but MUST return to the configured fallback cadence within one interval after the channel is restored and SHALL log the entire degraded period
9. WHEN remote changes are detected via delta query, THE OneMount System SHALL update the local metadata cache
10. WHEN a remotely modified file is accessed, THE OneMount System SHALL download the new version
11. WHEN a cached file has been modified remotely, THE OneMount System SHALL invalidate the local cache entry using ETag comparison
12. IF a file has both local and remote changes, THEN THE OneMount System SHALL create a conflict copy
13. WHEN delta sync completes, THE OneMount System SHALL store the @odata.deltaLink token for the next sync cycle
14. WHEN a subscription approaches expiration (per Graph limits), THE OneMount System SHALL renew the Socket.IO subscription proactively
15. IF subscription renewal or reconnection fails, THEN THE OneMount System SHALL continue using the shorter polling interval until the subscription is restored and SHALL raise diagnostics for the operator

### Requirement 6: Offline Mode Verification

**User Story:** As a user with unreliable internet, I want to access previously downloaded files when offline so that I can continue working.

#### Acceptance Criteria

1. WHEN network connectivity is lost, THE OneMount System SHALL detect the offline state using passive monitoring of API call failures
2. WHEN the system is online, THE OneMount System SHALL perform periodic active connectivity checks to Microsoft Graph endpoints
3. WHEN a network error matches known offline patterns, THE OneMount System SHALL transition to offline mode
4. WHILE offline, THE OneMount System SHALL serve cached files for read operations
5. WHILE offline, THE OneMount System SHALL allow read and write operations with changes queued for synchronization when connectivity is restored
6. WHEN a file is modified offline, THE OneMount System SHALL track the change in persistent storage for later upload
7. WHEN multiple changes are made to the same file offline, THE OneMount System SHALL preserve the most recent version for upload
8. WHEN a file is created offline, THE OneMount System SHALL queue the creation operation for synchronization when connectivity is restored
9. WHEN a file is deleted offline, THE OneMount System SHALL queue the deletion operation for synchronization when connectivity is restored
10. WHEN network connectivity is restored, THE OneMount System SHALL process queued uploads in batches
11. WHEN processing offline changes, THE OneMount System SHALL verify each change was successfully synchronized before removing it from the queue
12. WHEN processing offline changes, THE OneMount System SHALL detect conflicts between local and remote versions using ETag comparison
13. IF a conflict is detected during offline-to-online synchronization, THEN THE OneMount System SHALL apply the configured conflict resolution strategy
14. WHERE the user configures connectivity check interval, THE OneMount System SHALL use the specified interval for active connectivity checks
15. IF the connectivity check interval is not specified, THEN THE OneMount System SHALL use a default interval of 15 seconds
16. WHERE the user configures connectivity timeout, THE OneMount System SHALL use the specified timeout for connectivity checks
17. IF the connectivity timeout is not specified, THEN THE OneMount System SHALL use a default timeout of 10 seconds
18. WHERE the user configures maximum pending changes limit, THE OneMount System SHALL enforce the specified limit for offline change tracking
19. IF the maximum pending changes limit is not specified, THEN THE OneMount System SHALL use a default limit of 1000 changes
20. WHEN network connectivity is restored, THE OneMount System SHALL resume delta sync operations

### Requirement 7: Cache Management Verification

**User Story:** As a user, I want the cache to be managed efficiently so that it doesn't consume excessive disk space and always reflects the latest remote state.

#### Acceptance Criteria

1. WHEN files are downloaded, THE OneMount System SHALL store content in the cache directory with the file's ETag
2. WHEN files are accessed, THE OneMount System SHALL update the last access time in the cache
3. WHEN a cached file's ETag differs from the remote ETag, THE OneMount System SHALL invalidate the cache entry and download the new version
4. WHEN delta sync detects remote changes, THE OneMount System SHALL invalidate affected cache entries to prevent stale data
5. WHILE the cache cleanup process runs, THE OneMount System SHALL remove files older than the expiration threshold
6. WHERE cache expiration is configured, THE OneMount System SHALL respect the configured number of days
7. WHEN the user requests cache statistics, THE OneMount System SHALL display cache size, file count, and hit rate
8. WHEN a file is deleted from the filesystem, THE OneMount System SHALL remove the corresponding cache entry to free disk space
9. WHEN cache cleanup runs, THE OneMount System SHALL identify and remove cache entries for files that no longer exist in the filesystem metadata

### Requirement 8: Conflict Resolution Verification

**User Story:** As a user, I want conflicts between local and remote changes to be handled gracefully so that I don't lose any work.

#### Acceptance Criteria

1. WHEN a file has been modified both locally and remotely, THE OneMount System SHALL detect the conflict by comparing ETags
2. WHEN uploading a file with local changes, THE OneMount System SHALL check if the remote ETag has changed since last sync
3. IF the remote ETag differs from the cached ETag, THEN THE OneMount System SHALL detect a conflict
4. WHEN a conflict is detected, THE OneMount System SHALL preserve the local version with its original name
5. WHEN a conflict is detected, THE OneMount System SHALL create a conflict copy with a timestamp suffix
6. WHEN a conflict is detected, THE OneMount System SHALL download the remote version as the conflict copy
7. WHEN a conflict is resolved, THE OneMount System SHALL log the conflict details including file path, ETags, and timestamps
8. WHERE multiple conflict resolution strategies are available, THE OneMount System SHALL use the configured strategy (last-writer-wins, keep-both, user-choice, merge, or rename)
9. WHEN the user accesses a file with unresolved conflicts, THE OneMount System SHALL display both versions
10. WHERE the user configures a default conflict resolution strategy, THE OneMount System SHALL use the specified strategy for automatic conflict resolution
11. IF no conflict resolution strategy is configured, THEN THE OneMount System SHALL use the keep-both strategy as default
12. WHEN using the last-writer-wins strategy, THE OneMount System SHALL compare modification timestamps and preserve the most recent version
13. WHEN using the user-choice strategy, THE OneMount System SHALL present resolution options to the user
14. WHEN using the merge strategy, THE OneMount System SHALL attempt automatic merging for compatible changes
15. WHEN using the rename strategy, THE OneMount System SHALL create separate versions with conflict indicators
16. WHEN using the keep-both strategy, THE OneMount System SHALL create separate versions for both local and remote changes

### Requirement 9: User Notifications and Feedback

**User Story:** As a user, I want to be notified of network state changes and synchronization status so that I understand the current state of my files.

#### Acceptance Criteria

1. WHERE the user configures feedback level, THE OneMount System SHALL provide notifications according to the specified level (none, basic, or detailed)
2. IF no feedback level is configured, THEN THE OneMount System SHALL use basic feedback level as default
3. WHEN network connectivity is lost, THE OneMount System SHALL emit a network disconnected notification
4. WHEN network connectivity is restored, THE OneMount System SHALL emit a network connected notification
5. WHEN offline-to-online synchronization starts, THE OneMount System SHALL emit a sync started notification
6. WHEN offline-to-online synchronization completes successfully, THE OneMount System SHALL emit a sync completed notification
7. WHEN conflicts are detected during synchronization, THE OneMount System SHALL emit a conflicts detected notification
8. WHEN synchronization fails, THE OneMount System SHALL emit a sync failed notification with error details
9. WHEN using basic feedback level, THE OneMount System SHALL provide simple connectivity status messages
10. WHEN using detailed feedback level, THE OneMount System SHALL provide comprehensive network and sync information
11. WHEN using none feedback level, THE OneMount System SHALL suppress user notifications but continue logging
12. WHERE D-Bus is available, THE OneMount System SHALL emit notifications via D-Bus signals
13. WHEN the user queries offline status, THE OneMount System SHALL provide current network connectivity state
14. WHEN the user queries cache status, THE OneMount System SHALL provide information about cached files for offline planning
15. WHERE the user enables manual offline mode, THE OneMount System SHALL allow explicit offline mode activation via command-line or configuration

### Requirement 9: User Notifications and Feedback

**User Story:** As a user, I want to be notified of network state changes and synchronization status so that I understand the current state of my files.

#### Acceptance Criteria

1. WHERE the user configures feedback level, THE OneMount System SHALL provide notifications according to the specified level (none, basic, or detailed)
2. IF no feedback level is configured, THEN THE OneMount System SHALL use basic feedback level as default
3. WHEN network connectivity is lost, THE OneMount System SHALL emit a network disconnected notification
4. WHEN network connectivity is restored, THE OneMount System SHALL emit a network connected notification
5. WHEN offline-to-online synchronization starts, THE OneMount System SHALL emit a sync started notification
6. WHEN offline-to-online synchronization completes successfully, THE OneMount System SHALL emit a sync completed notification
7. WHEN conflicts are detected during synchronization, THE OneMount System SHALL emit a conflicts detected notification
8. WHEN synchronization fails, THE OneMount System SHALL emit a sync failed notification with error details
9. WHEN using basic feedback level, THE OneMount System SHALL provide simple connectivity status messages
10. WHEN using detailed feedback level, THE OneMount System SHALL provide comprehensive network and sync information
11. WHEN using none feedback level, THE OneMount System SHALL suppress user notifications but continue logging
12. WHERE D-Bus is available, THE OneMount System SHALL emit notifications via D-Bus signals
13. WHEN the user queries offline status, THE OneMount System SHALL provide current network connectivity state
14. WHEN the user queries cache status, THE OneMount System SHALL provide information about cached files for offline planning
15. WHERE the user enables manual offline mode, THE OneMount System SHALL allow explicit offline mode activation via command-line or configuration

### Requirement 10: File Status and D-Bus Integration Verification

**User Story:** As a user of Nemo/Nautilus file manager, I want to see file sync status icons so that I know which files are synced, downloading, or have errors.

#### Acceptance Criteria

1. WHEN a file status changes, THE OneMount System SHALL update the extended attributes on the file
2. WHERE D-Bus is available, THE OneMount System SHALL send status update signals via D-Bus
3. WHEN the Nemo extension queries file status, THE OneMount System SHALL provide current status information
4. IF D-Bus is unavailable, THEN THE OneMount System SHALL continue operating using extended attributes only
5. WHILE files are downloading, THE OneMount System SHALL update status to show download progress

### Requirement 11: Error Handling and Recovery Verification

**User Story:** As a user, I want the system to handle errors gracefully so that temporary issues don't cause data loss or crashes.

#### Acceptance Criteria

1. WHEN a network error occurs, THE OneMount System SHALL log the error with context
2. WHEN an API rate limit is encountered, THE OneMount System SHALL implement exponential backoff
3. IF the filesystem crashes, THEN THE OneMount System SHALL preserve state in the persistent database
4. WHEN the system restarts after a crash, THE OneMount System SHALL recover incomplete uploads and resume operations
5. WHERE errors are user-facing, THE OneMount System SHALL display helpful error messages

### Requirement 12: Performance and Concurrency Verification

**User Story:** As a user, I want the filesystem to be responsive so that file operations don't block or hang.

#### Acceptance Criteria

1. WHEN multiple files are accessed simultaneously, THE OneMount System SHALL handle concurrent operations safely
2. WHILE downloads are in progress, THE OneMount System SHALL allow other file operations to proceed
3. WHEN the user lists a large directory, THE OneMount System SHALL respond within 2 seconds
4. WHERE file operations require locks, THE OneMount System SHALL use appropriate locking granularity
5. WHEN goroutines are spawned, THE OneMount System SHALL track them with wait groups for clean shutdown

### Requirement 13: Integration Test Coverage

**User Story:** As a developer, I want comprehensive integration tests so that I can verify the system works end-to-end.

#### Acceptance Criteria

1. THE OneMount System SHALL have integration tests for the complete authentication flow
2. THE OneMount System SHALL have integration tests for file upload and download workflows
3. THE OneMount System SHALL have integration tests for offline mode transitions
4. THE OneMount System SHALL have integration tests for conflict resolution
5. THE OneMount System SHALL have integration tests for cache cleanup and expiration

### Requirement 14: Multiple Account and Drive Support

**User Story:** As a user with multiple OneDrive accounts, I want to mount my personal OneDrive, work OneDrive, and shared drives simultaneously so that I can access all my files.

#### Acceptance Criteria

1. THE OneMount System SHALL support mounting multiple OneDrive accounts simultaneously at different mount points
2. WHEN mounting a personal OneDrive account, THE OneMount System SHALL access the user's personal drive using `/me/drive`
3. WHEN mounting a OneDrive for Business account, THE OneMount System SHALL access the user's work drive using `/me/drive`
4. THE OneMount System SHALL support mounting shared drives using `/drives/{drive-id}`
5. THE OneMount System SHALL support accessing "Shared with me" items using `/me/drive/sharedWithMe`
6. WHEN multiple accounts are mounted, THE OneMount System SHALL maintain separate authentication tokens for each account
7. WHEN multiple accounts are mounted, THE OneMount System SHALL maintain separate caches for each account
8. WHEN multiple accounts are mounted, THE OneMount System SHALL maintain separate delta sync loops for each account

### Requirement 15: XDG Base Directory Compliance

**User Story:** As a Linux user, I want OneMount to follow XDG Base Directory standards so that my configuration and cache files are stored in standard locations.

#### Acceptance Criteria

1. THE OneMount System SHALL use `os.UserConfigDir()` to determine the configuration directory
2. WHEN `XDG_CONFIG_HOME` is set, THE OneMount System SHALL store configuration in `$XDG_CONFIG_HOME/onemount/`
3. WHEN `XDG_CONFIG_HOME` is not set, THE OneMount System SHALL store configuration in `$HOME/.config/onemount/`
4. THE OneMount System SHALL use `os.UserCacheDir()` to determine the cache directory
5. WHEN `XDG_CACHE_HOME` is set, THE OneMount System SHALL store cache in `$XDG_CACHE_HOME/onemount/`
6. WHEN `XDG_CACHE_HOME` is not set, THE OneMount System SHALL store cache in `$HOME/.cache/onemount/`
7. THE OneMount System SHALL store authentication tokens in the configuration directory
8. THE OneMount System SHALL store file content cache in the cache directory
9. THE OneMount System SHALL store metadata database (bbolt) in the cache directory
10. WHERE the user specifies custom paths via command-line flags, THE OneMount System SHALL use the specified paths instead of XDG defaults
11. THE OneMount System SHALL create `.xdg-volume-info` files as local-only virtual files that are NOT synced to OneDrive
12. WHEN creating `.xdg-volume-info` files, THE OneMount System SHALL assign them a local-only ID (prefixed with "local-")
13. WHEN accessing `.xdg-volume-info` files, THE OneMount System SHALL serve content from the local cache without attempting to sync to OneDrive

### Requirement 16: Docker-Based Test Environment

**User Story:** As a developer, I want to run all tests in isolated Docker containers so that my local environment is not affected by test execution.

#### Acceptance Criteria

1. THE OneMount System SHALL provide Docker containers for running unit tests
2. THE OneMount System SHALL provide Docker containers for running integration tests
3. THE OneMount System SHALL provide Docker containers for running system tests
4. WHEN tests are executed in Docker, THE OneMount System SHALL mount the workspace as a volume to access source code
5. WHEN tests complete, THE OneMount System SHALL write test artifacts to a mounted volume accessible from the host
6. WHERE FUSE operations are required, THE OneMount System SHALL configure containers with appropriate capabilities and devices
7. THE OneMount System SHALL provide a test runner container with all required dependencies pre-installed

### Requirement 17: Socket.IO Subscription Management

**User Story:** As a system, I want to consume Microsoft Graph Socket.IO change notifications directly so that I can deliver real-time updates without relying on external services or webhooks.

#### Acceptance Criteria

1. WHEN mounting a drive, THE OneMount System SHALL request a Socket.IO notification endpoint from Microsoft Graph for the selected resource (e.g., `/me/drive/root`).
2. THE OneMount System SHALL establish and maintain the Engine.IO v4 WebSocket connection directly, without delegating to Azure Web PubSub or any other managed relay; the application SHALL remain fully standalone.
3. WHEN establishing the Socket.IO connection, THE OneMount System SHALL persist the subscription ID, notification URL, and expiration so they can be renewed before expiry.
4. WHILE the connection is active, THE OneMount System SHALL send and monitor ping/pong heartbeats according to Engine.IO timing and SHALL treat missed heartbeats as failures requiring reconnection.
5. WHEN the Socket.IO connection drops, THE OneMount System SHALL attempt reconnection with exponential backoff (capped to 60 seconds), and SHALL log each failure with enough context to diagnose authentication or protocol issues.
6. IF reconnection fails repeatedly, THEN THE OneMount System SHALL declare the subscription unhealthy, fall back to the delta polling behavior defined in Requirement 5, and continue retrying the Socket.IO channel in the background.
7. WHEN the subscription approaches expiration (per Graph-imposed limits), THE OneMount System SHALL renew it proactively; renewal attempts SHALL be logged and retried on failure.
8. WHEN unmounting a drive or shutting down, THE OneMount System SHALL gracefully close the Socket.IO connection and delete/cleanup the associated subscription metadata.
9. THE OneMount System SHALL NOT expose or require inbound HTTP webhook endpoints; all real-time updates SHALL flow over the Socket.IO subscription to avoid terminology confusion with traditional webhooks.

### Requirement 18: Documentation Alignment

**User Story:** As a developer, I want documentation to match the actual implementation so that I can understand and maintain the code.

#### Acceptance Criteria

1. THE OneMount System SHALL have architecture documentation that accurately describes component interactions
2. THE OneMount System SHALL have design documentation that matches the implemented data models
3. THE OneMount System SHALL have API documentation that reflects actual function signatures
4. WHERE implementation differs from design, THE OneMount System SHALL document the rationale
5. WHEN code changes are made, THE OneMount System SHALL update corresponding documentation

### Requirement 19: Network Error Pattern Recognition

**User Story:** As a system, I want to recognize specific network error patterns so that I can accurately detect offline conditions.

#### Acceptance Criteria

1. WHEN a network error contains "no such host", THE OneMount System SHALL classify it as an offline condition
2. WHEN a network error contains "network is unreachable", THE OneMount System SHALL classify it as an offline condition
3. WHEN a network error contains "connection refused", THE OneMount System SHALL classify it as an offline condition
4. WHEN a network error contains "connection timed out", THE OneMount System SHALL classify it as an offline condition
5. WHEN a network error contains "dial tcp", THE OneMount System SHALL classify it as an offline condition
6. WHEN a network error contains "context deadline exceeded", THE OneMount System SHALL classify it as an offline condition
7. WHEN a network error contains "no route to host", THE OneMount System SHALL classify it as an offline condition
8. WHEN a network error contains "network is down", THE OneMount System SHALL classify it as an offline condition
9. WHEN a network error contains "temporary failure in name resolution", THE OneMount System SHALL classify it as an offline condition
10. WHEN a network error contains "operation timed out", THE OneMount System SHALL classify it as an offline condition
11. WHEN an offline condition is detected, THE OneMount System SHALL log the specific error pattern that triggered the detection

### Requirement 20: Engine.IO / Socket.IO Transport Implementation

**User Story:** As a OneMount developer, I want clear requirements for the built-in Engine.IO/Socket.IO transport so that the realtime channel can be implemented and tested without relying on unmaintained third-party libraries or external services.

#### Acceptance Criteria

1. THE OneMount System SHALL implement the Microsoft Graph notification channel using Engine.IO v4 over WebSocket only, setting `EIO=4` and `transport=websocket` query parameters and joining the default namespace (`/`).
2. THE transport SHALL attach the current OAuth access token (Authorization bearer header) and any additional headers Graph requires, and SHALL refresh the connection whenever the token is rotated.
3. WHEN an Engine.IO handshake frame is received, THE transport SHALL parse the ping interval/timeout values, log them at debug level, and configure its heartbeat timers accordingly.
4. THE transport SHALL send ping/pong frames per the negotiated interval, detect two consecutive missed heartbeats as a failure, and immediately surface the unhealthy state to the delta loop for fallback polling.
5. WHEN the connection closes or errors, THE transport SHALL attempt reconnection with exponential backoff (starting at 1 s, doubling each attempt, capped at 60 s, with ±10 % jitter) and SHALL reset the backoff after a successful reconnect.
6. THE implementation SHALL stream decoded Socket.IO events (e.g., `notification`, `error`) through strongly typed callbacks, and SHALL expose a health indicator that the delta sync loop can query in constant time.
7. THE transport SHALL emit structured trace logs for handshake data, ping/pong timing, packet read/write summaries (payload truncated to a configurable limit), and close/error codes sufficient for supportability.
8. THE transport SHALL include automated tests covering packet encode/decode, heartbeat scheduling, reconnection backoff, and error propagation so regressions can be caught without live Graph access.
9. THE transport SHALL remain self-contained within the OneMount codebase—no third-party Socket.IO client libraries, proxies, or managed relays (e.g., Azure Web PubSub) are permitted.
