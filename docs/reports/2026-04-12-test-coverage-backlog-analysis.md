# Test Coverage Backlog Analysis

**Date**: 2026-04-12  
**Type**: Coverage Report  
**Scope**: All `*_test.go` files across the codebase  
**Target Release**: v1.1

---

## Executive Summary

There are **46 unimplemented test cases** across the codebase (all containing `t.Skip("Test not implemented yet")`). The existing TODO summary document (`docs/0-project-management/todo_comments_summary.md`) overstates the count at 50+ and incorrectly lists several test areas (D-Bus, systemd, UI) as unimplemented when they have since been fully implemented.

---

## Accurate Inventory of Unimplemented Tests

### 1. Hash Functions — `internal/graph/hash_functions_test.go` (4 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| UT-GR-13-01 | `TestUT_GR_13_01_SHA1Hash_VariousInputs_ReturnsCorrectHash` | High |
| UT-GR-14-01 | `TestUT_GR_14_01_SHA1HashStream_VariousInputs_ReturnsCorrectHash` | High |
| UT-GR-15-01 | `TestUT_GR_15_01_QuickXORHash_VariousInputs_ReturnsCorrectHash` | **Critical** |
| UT-GR-16-01 | `TestUT_GR_16_01_QuickXORHashStream_VariousInputs_ReturnsCorrectHash` | **Critical** |

**Notes**: SHA256Hash and SHA256HashStream tests (UT-GR-11-01, UT-GR-12-01) are already fully implemented and serve as a pattern template. The QuickXORHash tests are the highest priority — OneDrive file integrity verification depends on this algorithm. The source is in `internal/graph/hashes.go` and uses `internal/quickxorhash`.

### 2. Filesystem Integration Tests — `internal/fs/fs_integration_test.go` (12 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| IT-FS-15-01 | `TestIT_FS_15_01_File_Permissions_CorrectlyApplied` | Medium |
| IT-FS-16-01 | `TestIT_FS_16_01_Directory_CreateAndModify_OperationsSucceed` | Medium |
| IT-FS-19-01 | `TestIT_FS_19_01_File_WriteAtOffset_DataCorrectlyPositioned` | Medium |
| IT-FS-20-01 | `TestIT_FS_20_01_File_MoveAndRename_FileCorrectlyRelocated` | Medium |
| IT-FS-23-01 | `TestIT_FS_23_01_Filename_CaseSensitivity_HandledCorrectly` | Medium |
| IT-FS-24-01 | `TestIT_FS_24_01_Filename_Case_PreservedCorrectly` | Medium |
| IT-FS-25-01 | `TestIT_FS_25_01_Shell_FileOperations_WorkCorrectly` | Medium |
| IT-FS-26-01 | `TestIT_FS_26_01_File_GetInfo_AttributesCorrectlyRetrieved` | Medium |
| IT-FS-27-01 | `TestIT_FS_27_01_Filename_QuestionMarks_HandledCorrectly` | Medium |
| IT-FS-29-01 | `TestIT_FS_29_01_ListChildren_Paging_AllChildrenReturned` | Medium |
| IT-FS-30-01 | `TestIT_FS_30_01_LibreOffice_SavePattern_HandledCorrectly` | Medium |
| IT-FS-31-01 | `TestIT_FS_31_01_Filename_DisallowedCharacters_HandledCorrectly` | Medium |

**Notes**: Several FS integration tests ARE implemented: IT-FS-17-01 (Directory Remove), IT-FS-18-01 (File Basic Operations), IT-FS-21-01 (File Positional Operations), IT-FS-22-01 (FileSystem Basic Operations), IT-FS-28-01 (GIO Trash Integration). These serve as implementation patterns.

### 3. Delta Sync Tests — `internal/fs/delta_test.go` (7 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| IT-FS-03-01 | `TestIT_FS_03_01_Delta_SyncOperations_ChangesAreSynced` | Medium-High |
| IT-FS-04-01 | `TestIT_FS_04_01_Delta_RemoteContentChange_ClientIsUpdated` | Medium-High |
| IT-FS-06-01 | `TestIT_FS_06_01_Delta_CorruptedCache_ContentIsRestored` | Medium |
| IT-FS-07-01 | `TestIT_FS_07_01_Delta_FolderDeletion_EmptyFoldersAreDeleted` | Medium |
| IT-FS-08-01 | `TestIT_FS_08_01_Delta_NonEmptyFolderDeletion_FolderIsPreserved` | Medium |
| IT-FS-09-01 | `TestIT_FS_09_01_Delta_UnchangedContent_ModTimeIsPreserved` | Medium |
| IT-FS-10-01 | `TestIT_FS_10_01_Delta_MissingHash_HandledCorrectly` | Medium |

**Notes**: IT-FS-05-01 (Conflicting Changes) is fully implemented and serves as a pattern. Additionally, 4 newer delta tests are implemented: `ApplyDeltaPersistsMetadata`, `RemoteInvalidationTransitions`, `MoveUpdatesMetadata`, `PinnedFileQueuesHydration`, plus 5 `DesiredDeltaInterval*` tests.

### 4. Inode Tests — `internal/fs/inode_test.go` (4 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| IT-FS-01-01 | `TestIT_FS_01_01_Inode_Creation_HasCorrectProperties` | Medium |
| IT-FS-02-01 | `TestIT_FS_02_01_Inode_Properties_ModeAndDirectoryDetection` | Medium |
| IT-FS-03-01 | `TestIT_FS_03_01_Filename_SpecialCharacters_ProperlyEscaped` | Medium |
| IT-FS-04-01 | `TestIT_FS_04_01_FileCreation_VariousScenarios_BehavesCorrectly` | Medium |

### 5. XAttr Operations Tests — `internal/fs/xattr_operations_test.go` (3 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| IT-FS-32-01 | `TestIT_FS_32_01_XAttr_BasicOperations_WorkCorrectly` | Medium |
| IT-FS-33-01 | `TestIT_FS_33_01_FileStatus_XAttr_StatusCorrectlyReported` | Medium |
| IT-FS-34-01 | `TestIT_FS_34_01_Filesystem_XAttrOperations_WorkCorrectly` | Medium |

### 6. Thumbnail Tests — `internal/fs/thumbnail_test.go` (3 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| UT-FS-08-01 | `TestIT_FS_08_01_ThumbnailCache_BasicOperations_WorkCorrectly` | Medium |
| IT-FS-09-01 | `TestIT_FS_09_01_ThumbnailCache_Cleanup_RemovesExpiredThumbnails` | Medium |
| IT-FS-10-01 | `TestIT_FS_10_01_Thumbnails_FileSystemOperations_WorkCorrectly` | Medium |

### 7. Upload Manager Tests — `internal/fs/upload_manager_test.go` (2 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| IT-FS-35-01 | `TestIT_FS_35_01_UploadDisk_Serialization_StatePreserved` | Medium |
| IT-FS-36-01 | `TestIT_FS_36_01_Upload_RepeatedUploads_HandledCorrectly` | Medium |

**Notes**: There is also a partial TODO in `TestIT_FS_06_UploadDiskSerialization_LargeFile_SuccessfulUpload` (line 889) but the test function has surrounding implementation — it's incomplete rather than fully stubbed.

### 8. Upload Session Tests — `internal/fs/upload_session_test.go` (1 test)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| IT-FS-05 | `TestIT_FS_05_UploadSessionManagement_SessionsCorrectlyManaged` | Medium |

### 9. Offline Tests — `internal/graph/offline_test.go` (3 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| UT-GR-23-01 | `TestUT_GR_23_01_OfflineState_SetAndGet_StateCorrectlyManaged` | Medium-High |
| UT-GR-24-01 | `TestUT_GR_24_01_IsOffline_OperationalStateSet_ReturnsTrue` | Medium-High |
| UT-GR-25-01 | `TestUT_GR_25_01_IsOffline_VariousErrors_IdentifiesNetworkErrors` | Medium-High |

### 10. OAuth2 GTK Test — `internal/graph/oauth2_gtk_test.go` (1 test)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| UT-GR-17-02 | `TestUT_GR_17_02_URIGetHost_VariousURIs_ReturnsCorrectHost` | Low-Medium |

### 11. Config Tests — `cmd/common/config_test.go` (4 tests)

| Test ID | Test Name | Priority |
|---------|-----------|----------|
| UT-CMD-02-01 | `TestUT_CMD_02_01_Config_ValidConfigFile_LoadsCorrectValues` | Medium |
| UT-CMD-03-01 | `TestUT_CMD_03_01_Config_MergedSettings_ContainsMergedValues` | Medium |
| UT-CMD-04-01 | `TestUT_CMD_04_01_Config_NonexistentFile_LoadsDefaultValues` | Medium |
| UT-CMD-05-01 | `TestUT_CMD_05_01_Config_ValidSettings_WritesSuccessfully` | Medium |

**Notes**: 8 other config tests in this file ARE implemented (delta interval, overlay policy, realtime config, active delta tuning, hydration, metadata queue, fallback bounds).

---

## Corrections to Existing TODO Summary

The document `docs/0-project-management/todo_comments_summary.md` has several inaccuracies:

| Area | Summary Claims | Actual Status |
|------|---------------|---------------|
| D-Bus tests | 2 TODOs | **0 TODOs** — All D-Bus tests fully implemented (8 test functions) |
| Systemd tests | 2 TODOs | **0 TODOs** — Both systemd tests fully implemented |
| UI tests | 3 TODOs | **0 TODOs** — All 3 UI tests fully implemented |
| Common tests | 1 TODO | **0 TODOs** — `TestUT_CMD_01_01` is fully implemented |
| FS integration | 13 TODOs | **12 TODOs** — IT-FS-28-01 (GIO Trash) is implemented |
| Upload manager | 3 TODOs | **2 TODOs** — One is partial (incomplete, not fully stubbed) |
| Total | "50+" | **46** actual unimplemented tests |

---

## Recommended Implementation Order

### Phase 1: Critical (blocks release confidence)
1. **QuickXORHash tests** (UT-GR-15-01, UT-GR-16-01) — OneDrive file integrity depends on this. Pattern: copy SHA256 test structure, use known QuickXOR values.
2. **SHA1Hash tests** (UT-GR-13-01, UT-GR-14-01) — Straightforward; mirror SHA256 tests with SHA1 known values.

### Phase 2: High (core functionality coverage)
3. **Offline mode tests** (UT-GR-23-01, UT-GR-24-01, UT-GR-25-01) — Unit tests, no external dependencies.
4. **Delta sync tests** (7 tests) — IT-FS-05-01 provides the implementation pattern.

### Phase 3: Medium (feature completeness)
5. **Inode tests** (4 tests) — Pure unit tests, straightforward.
6. **Config tests** (4 tests) — Pure unit tests, 8 sibling tests provide patterns.
7. **FS integration tests** (12 tests) — IT-FS-17-01, IT-FS-18-01, IT-FS-22-01 provide patterns.
8. **XAttr, thumbnail, upload tests** (9 tests) — Feature-specific coverage.

### Phase 4: Low
9. **OAuth2 GTK test** (1 test) — May require GTK build dependencies.

---

## Summary by Count

| Category | File | Unimplemented | Implemented |
|----------|------|:---:|:---:|
| Hash Functions | `internal/graph/hash_functions_test.go` | 4 | 2 |
| FS Integration | `internal/fs/fs_integration_test.go` | 12 | 6 |
| Delta Sync | `internal/fs/delta_test.go` | 7 | 10 |
| Inode | `internal/fs/inode_test.go` | 4 | 0 |
| XAttr | `internal/fs/xattr_operations_test.go` | 3 | 0 |
| Thumbnails | `internal/fs/thumbnail_test.go` | 3 | 0 |
| Upload Manager | `internal/fs/upload_manager_test.go` | 2 | 6 |
| Upload Session | `internal/fs/upload_session_test.go` | 1 | 0 |
| Offline | `internal/graph/offline_test.go` | 3 | 0 |
| OAuth2 GTK | `internal/graph/oauth2_gtk_test.go` | 1 | 0 |
| Config | `cmd/common/config_test.go` | 4 | 8 |
| D-Bus | `internal/fs/dbus_test.go` | 0 | 8 |
| Systemd | `internal/ui/systemd/systemd_test.go` | 0 | 2 |
| UI | `internal/ui/onemount_test.go` | 0 | 3 |
| Common | `cmd/common/common_test.go` | 0 | 1 |
| **Total** | | **44** | **46** |

*Note: 2 additional partial TODOs exist in upload_manager_test.go (incomplete implementation rather than full stubs), bringing the effective total to 46.*
