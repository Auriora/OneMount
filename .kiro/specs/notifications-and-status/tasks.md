# Tasks: Notifications And Status

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/tasks.md`. Sections below are verbatim.

## Phase 11: File Status and D-Bus Verification

- [-] 13. Verify file status tracking
- [x] 13.1 Review file status code
  - Read and analyze `internal/fs/file_status.go`
  - Review `internal/fs/dbus.go`
  - Check extended attribute implementation
  - Review Nemo extension code
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [x] 13.2 Test file status updates with manual verification
  - Monitor file status during various operations
  - Verify status changes appropriately (synced, downloading, error, etc.)
  - Check extended attributes are set correctly
  - Run: `./tests/manual/test_file_status_updates.sh`
  - Verify file status updates work correctly with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.1_

- [x] 13.3 Test D-Bus integration with manual verification
  - Verify D-Bus server starts successfully
  - Monitor D-Bus signals during file operations
  - Use `dbus-monitor` to observe signals
  - Verify signal format and content
  - Run outside of docker: `./tests/manual/test_dbus_integration.sh`
  - Verify D-Bus signals are emitted correctly with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.2_

- [x] 13.4 Test D-Bus fallback with manual verification
  - Disable D-Bus (or run in environment without D-Bus)
  - Verify system continues operating
  - Check that extended attributes still work
  - Run outside of docker: `./tests/manual/test_dbus_fallback.sh`
  - Verify fallback to extended attributes works with real OneDrive
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.4_

- [x] 13.5 Test Nemo extension with manual verification
  - Open Nemo file manager
  - Navigate to mounted OneDrive
  - Verify status icons appear on files
  - Trigger file operations and watch icons update
  - Test with real OneDrive mount outside Docker
  - Document results in `docs/verification-tracking.md` Phase 11 section
  - _Requirements: 8.3_

- [x] 13.6 Create file status integration tests
  - Write test for status tracking
  - Write test for D-Bus signal emission
  - Write test for extended attribute fallback
  - _Requirements: 8.1, 8.2, 8.4_

- [x] 13.7 Document file status issues and create fix plan
  - List all discovered issues
  - Identify root causes
  - Create prioritized fix plan
  - Update the relevant sections of the verification-tracking.md document
  - _Requirements: 12.1_

---
