# Tasks: Delta Sync Realtime

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/tasks.md`. Sections below are verbatim.

## Phase 8: Delta Synchronization Verification

- [x] 10. Verify delta synchronization
- [x] 10.1 Review delta sync code
  - Read and analyze `internal/fs/delta.go`
  - Review `internal/fs/sync.go`
  - Check delta loop implementation
  - Review delta link persistence
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 10.2 Test initial delta sync
  - Start with empty cache
  - Mount filesystem
  - Verify initial sync fetches all metadata
  - Check that delta link is stored
  - _Requirements: 5.1, 5.5_

- [x] 10.3 Test incremental delta sync
  - Create a file on OneDrive web interface
  - Wait for delta sync to run
  - Verify new file appears in mounted filesystem
  - Check that only changes were fetched
  - _Requirements: 5.1, 5.2_

- [x] 10.4 Test remote file modification
  - Modify a file on OneDrive web interface
  - Wait for delta sync
  - Access the file locally
  - Verify new version is downloaded
  - _Requirements: 5.3_

- [x] 10.5 Test conflict detection and resolution with real OneDrive
  - Modify a file locally (don't let it upload yet)
  - Modify same file on OneDrive web interface
  - Trigger delta sync
  - Verify conflict is detected
  - Check that conflict copy is created
  - Verify local version is preserved
  - **Retest with real OneDrive**: `docker compose -f docker/compose/docker-compose.test.yml run --rm integration-tests go test -v -run TestIT_FS.*Conflict ./internal/fs`
  - Verify delta sync conflict detection works with real API
  - Verify local changes preserved when remote changes detected
  - Verify ETag comparison mechanism for conflict detection
  - Document results in `docs/verification-tracking.md` Phase 7 section
  - _Requirements: 5.4, 8.1, 8.2, 8.3_

- [x] 10.6 Test delta sync persistence
  - Run delta sync
  - Unmount filesystem
  - Remount filesystem
  - Verify delta sync resumes from last position
  - _Requirements: 5.5_

- [x] 10.7 Create delta sync integration tests
  - Write test for initial sync
  - Write test for incremental sync
  - Write test for conflict detection
  - Write test for delta link persistence
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 10.8 Implement delta synchronization property-based tests
- [x] 10.8.1 Implement Property 20: Initial Delta Sync
  - **Property 20: Initial Delta Sync**
  - **Validates: Requirements 5.1**
  - Create `internal/fs/delta_property_test.go`
  - Generate random first filesystem mount scenarios
  - Verify complete directory structure fetch using delta API
  - Test with various OneDrive structures
  - _Requirements: 5.1_

- [x] 10.8.2 Implement Property 21: Metadata Cache Updates
  - **Property 21: Metadata Cache Updates**
  - **Validates: Requirements 5.8**
  - Generate random remote change scenarios via delta query
  - Verify local metadata cache is updated correctly
  - Test with various change types (create, modify, delete)
  - _Requirements: 5.8_

- [x] 10.8.3 Implement Property 22: Conflict Copy Creation
  - **Property 22: Conflict Copy Creation**
  - **Validates: Requirements 5.11**
  - Generate random files with both local and remote changes
  - Verify conflict copy is created correctly
  - Test conflict detection accuracy
  - _Requirements: 5.11_

- [x] 10.8.4 Implement Property 23: Delta Token Persistence
  - **Property 23: Delta Token Persistence**
  - **Validates: Requirements 5.12**
  - Generate random delta sync completion scenarios
  - Verify @odata.deltaLink token is stored for next cycle
  - Test token persistence across restarts
  - _Requirements: 5.12_

- [x] 10.8.5 Implement Property 30: ETag-Based Conflict Detection
  - **Property 30: ETag-Based Conflict Detection**
  - **Validates: Requirements 8.1**
  - Generate random files modified both locally and remotely
  - Verify conflict detection using ETag comparison
  - Test ETag comparison accuracy
  - _Requirements: 8.1_

- [x] 10.8.6 Implement Property 31: Local Version Preservation
  - **Property 31: Local Version Preservation**
  - **Validates: Requirements 8.4**
  - Generate random conflict scenarios
  - Verify local version is preserved with original name
  - Test version preservation integrity
  - _Requirements: 8.4_

- [x] 10.8.7 Implement Property 32: Conflict Copy Creation with Timestamp
  - **Property 32: Conflict Copy Creation with Timestamp**
  - **Validates: Requirements 8.5**
  - Generate random conflict scenarios
  - Verify conflict copy creation with timestamp suffix
  - Test timestamp format and uniqueness
  - _Requirements: 8.5_

- [x] 10.9 Document delta sync issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---

## Phase 16: Socket.IO Transport Implementation Verification

**Status**: ✅ **IMPLEMENTED** - Verification tasks for Requirement 20 compliance

- [x] 27. Verify Engine.IO/Socket.IO Transport Implementation (Requirement 20)
- [x] 27.1 Review Socket.IO transport implementation
  - Read and analyze `internal/socketio/` implementation
  - Review Engine.IO v4 WebSocket transport
  - Check EIO=4 and transport=websocket query parameters
  - Verify default namespace (/) joining
  - _Requirements: 20.1_

- [x] 27.2 Test OAuth token attachment and refresh
  - Test Authorization bearer header attachment
  - Test token refresh during connection
  - Verify additional Graph-required headers
  - Test connection refresh on token rotation
  - _Requirements: 20.2_

- [x] 27.3 Test Engine.IO handshake and heartbeat
  - Test Engine.IO handshake frame parsing
  - Verify ping interval/timeout value parsing
  - Test debug level logging of handshake data
  - Test heartbeat timer configuration
  - _Requirements: 20.3_

- [x] 27.4 Test ping/pong and failure detection
  - Test ping/pong frame sending per negotiated interval
  - Test two consecutive missed heartbeat detection
  - Verify unhealthy state surfacing to ChangeNotifier
  - Test fallback to polling on heartbeat failure
  - _Requirements: 20.4_

- [x] 27.5 Test reconnection and backoff logic
  - Test exponential backoff on connection close/error
  - Verify backoff parameters (1s start, 2x multiplier, 60s cap, ±10% jitter)
  - Test backoff reset after successful reconnect
  - Test connection retry behavior
  - _Requirements: 20.5_

- [x] 27.6 Test event streaming and health monitoring
  - Test Socket.IO event streaming (notification, error)
  - Verify strongly typed callback handling
  - Test health indicator constant-time queries
  - Test ChangeNotifier integration
  - _Requirements: 20.6_

- [x] 27.7 Test verbose logging and tracing
  - Test structured trace logs for handshake data
  - Test ping/pong timing logs
  - Test packet read/write summary logs
  - Test payload truncation to configurable limit
  - Test close/error code logging
  - _Requirements: 20.7_

- [x] 27.8 Test automated transport tests
  - Run packet encode/decode tests
  - Run heartbeat scheduling tests
  - Run reconnection backoff tests
  - Run error propagation tests
  - Verify tests work without live Graph access
  - _Requirements: 20.8_

- [x] 27.9 Verify self-contained implementation
  - Verify no third-party Socket.IO client libraries
  - Verify no external proxies or managed relays
  - Check configuration whitelist for troubleshooting tools
  - Verify implementation is within OneMount codebase
  - _Requirements: 20.9_

- [x] 27.10 Create Socket.IO transport integration tests
  - Write test for complete transport lifecycle
  - Write test for OAuth integration
  - Write test for heartbeat and reconnection
  - Write test for event streaming
  - _Requirements: 20.1-20.9_

---
