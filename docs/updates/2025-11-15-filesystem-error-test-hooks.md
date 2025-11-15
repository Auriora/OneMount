# Filesystem Error Scenario Hook Support

**Date**: 2025-11-15  
**Type**: Bugfix / Test Infrastructure  
**Components**: Filesystem core, test utilities  
**Status**: Complete

## Summary

- Added optional `FilesystemTestHooks` so unit/integration tests can deterministically inject Open/Create/Write failures without relying on brittle Graph API mocks.
- Reworked the disk-space and permission error scenario tests to use the new hooks, aligning expectations with actual filesystem behavior.
- Fixed `fs_integration_test.go` to unwrap the `UnitTestFixture` correctly and refreshed mock directory listings so integration tests observe all created children.
- Updated `MountTestHelper` to provision a mock Graph client and exposed a helper to unwrap the mount fixtures, resolving the `TestIT_MU_01_01_MountUnmount_BasicCycle_WorksCorrectly` panic in the short suite.
- Taught the `path_operations_test` suite (Path_01 and Path_03) to seed mock directories/files via `CreateMockDirectory`/`CreateMockFile`, keeping DriveChildren responses consistent with recent helper changes.

## Testing

- `HOME=$(pwd)/.home GOCACHE=$(pwd)/.gocache LOG_LEVEL=error go test ./internal/fs -run 'TestUT_FS_ERR_|TestIT_FS_1[2348]_01' -count=1`
- `HOME=$(pwd)/.home GOCACHE=$(pwd)/.gocache LOG_LEVEL=error go test ./internal/fs -run 'TestUT_FS_Path_0' -count=1`

## Rules Consulted

- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-RULE-Testing-Conventions (priority 25)
- AGENT-RULE-Documentation-Conventions (priority 20)
