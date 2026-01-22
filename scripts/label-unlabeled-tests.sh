#!/bin/bash
# Script to label all unlabeled tests based on test audit analysis
# Task 46.1.2: Label all unlabeled tests

set -e

echo "=== Labeling Unlabeled Tests ==="
echo "This script will add appropriate prefixes to all unlabeled tests"
echo ""

# Function to add prefix to test function
add_test_prefix() {
    local file="$1"
    local old_name="$2"
    local new_name="$3"
    
    echo "  $file: $old_name -> $new_name"
    
    # Use sed to replace the function name
    sed -i "s/^func ${old_name}(/func ${new_name}(/" "$file"
}

# Priority 1: internal/fs/cache_test.go (5 unlabeled tests - all require auth)
echo "Processing internal/fs/cache_test.go..."
add_test_prefix "internal/fs/cache_test.go" "TestGetChildrenIDUsesMetadataStoreWhenOffline" "TestIT_FS_Cache_GetChildrenIDUsesMetadataStoreWhenOffline"
add_test_prefix "internal/fs/cache_test.go" "TestGetPathUsesMetadataStoreWhenOffline" "TestIT_FS_Cache_GetPathUsesMetadataStoreWhenOffline"
add_test_prefix "internal/fs/cache_test.go" "TestGetChildrenIDReturnsQuicklyWhenUncached" "TestIT_FS_Cache_GetChildrenIDReturnsQuicklyWhenUncached"
add_test_prefix "internal/fs/cache_test.go" "TestGetChildrenIDDoesNotCallGraphWhenMetadataPresent" "TestIT_FS_Cache_GetChildrenIDDoesNotCallGraphWhenMetadataPresent"
add_test_prefix "internal/fs/cache_test.go" "TestFallbackRootFromMetadata" "TestIT_FS_Cache_FallbackRootFromMetadata"

# Priority 2: internal/fs/concurrency_test.go (6 unlabeled tests - all require auth)
echo "Processing internal/fs/concurrency_test.go..."
add_test_prefix "internal/fs/concurrency_test.go" "TestConcurrentFileAccess" "TestIT_FS_Concurrency_ConcurrentFileAccess"
add_test_prefix "internal/fs/concurrency_test.go" "TestConcurrentCacheOperations" "TestIT_FS_Concurrency_ConcurrentCacheOperations"
add_test_prefix "internal/fs/concurrency_test.go" "TestDeadlockPrevention" "TestIT_FS_Concurrency_DeadlockPrevention"
add_test_prefix "internal/fs/concurrency_test.go" "TestDirectoryEnumerationWhileRefreshing" "TestIT_FS_Concurrency_DirectoryEnumerationWhileRefreshing"
add_test_prefix "internal/fs/concurrency_test.go" "TestHighConcurrencyStress" "TestIT_FS_Concurrency_HighConcurrencyStress"
add_test_prefix "internal/fs/concurrency_test.go" "TestConcurrentDirectoryOperations" "TestIT_FS_Concurrency_ConcurrentDirectoryOperations"

# Priority 3: internal/fs/dbus_test.go (8 unlabeled tests - all require auth)
echo "Processing internal/fs/dbus_test.go..."
add_test_prefix "internal/fs/dbus_test.go" "TestDBusServer_GetFileStatus" "TestIT_FS_DBus_GetFileStatus"
add_test_prefix "internal/fs/dbus_test.go" "TestDBusServer_GetFileStatus_WithRealFiles" "TestIT_FS_DBus_GetFileStatus_WithRealFiles"
add_test_prefix "internal/fs/dbus_test.go" "TestDBusServer_SendFileStatusUpdate" "TestIT_FS_DBus_SendFileStatusUpdate"
add_test_prefix "internal/fs/dbus_test.go" "TestDBusServiceNameGeneration" "TestIT_FS_DBus_ServiceNameGeneration"
add_test_prefix "internal/fs/dbus_test.go" "TestSetDBusServiceNameForMount" "TestIT_FS_DBus_SetServiceNameForMount"
add_test_prefix "internal/fs/dbus_test.go" "TestDBusServer_MultipleInstances" "TestIT_FS_DBus_MultipleInstances"
add_test_prefix "internal/fs/dbus_test.go" "TestSplitPathComponents" "TestIT_FS_DBus_SplitPathComponents"
add_test_prefix "internal/fs/dbus_test.go" "TestFindInodeByPath_PathTraversal" "TestIT_FS_DBus_FindInodeByPath_PathTraversal"

# Priority 4: internal/fs/delta_test.go (10 unlabeled tests - all require auth)
echo "Processing internal/fs/delta_test.go..."
add_test_prefix "internal/fs/delta_test.go" "TestApplyDeltaPersistsMetadataOnMetadataOnlyChange" "TestIT_FS_Delta_ApplyDeltaPersistsMetadataOnMetadataOnlyChange"
add_test_prefix "internal/fs/delta_test.go" "TestApplyDeltaRemoteInvalidationTransitionsMetadata" "TestIT_FS_Delta_ApplyDeltaRemoteInvalidationTransitionsMetadata"
add_test_prefix "internal/fs/delta_test.go" "TestApplyDeltaMoveUpdatesMetadataEntry" "TestIT_FS_Delta_ApplyDeltaMoveUpdatesMetadataEntry"
add_test_prefix "internal/fs/delta_test.go" "TestApplyDeltaPinnedFileQueuesHydration" "TestIT_FS_Delta_ApplyDeltaPinnedFileQueuesHydration"
add_test_prefix "internal/fs/delta_test.go" "TestDesiredDeltaIntervalUsesActiveWindow" "TestIT_FS_Delta_DesiredDeltaIntervalUsesActiveWindow"
add_test_prefix "internal/fs/delta_test.go" "TestDesiredDeltaIntervalFallsBackAfterWindow" "TestIT_FS_Delta_DesiredDeltaIntervalFallsBackAfterWindow"
add_test_prefix "internal/fs/delta_test.go" "TestDesiredDeltaIntervalUsesNotifierHealthHealthy" "TestIT_FS_Delta_DesiredDeltaIntervalUsesNotifierHealthHealthy"
add_test_prefix "internal/fs/delta_test.go" "TestDesiredDeltaIntervalUsesNotifierHealthDegraded" "TestIT_FS_Delta_DesiredDeltaIntervalUsesNotifierHealthDegraded"
add_test_prefix "internal/fs/delta_test.go" "TestDesiredDeltaIntervalUsesNotifierHealthFailedRecovery" "TestIT_FS_Delta_DesiredDeltaIntervalUsesNotifierHealthFailedRecovery"
add_test_prefix "internal/fs/delta_test.go" "TestApplyDeltaTransitionsStateOnRemoteInvalidation" "TestIT_FS_Delta_ApplyDeltaTransitionsStateOnRemoteInvalidation"

# Additional files with unlabeled tests requiring auth
echo "Processing internal/fs/dbus_getfilestatus_test.go..."
add_test_prefix "internal/fs/dbus_getfilestatus_test.go" "TestDBusServer_GetFileStatus_ValidPaths" "TestIT_FS_DBus_GetFileStatus_ValidPaths"
add_test_prefix "internal/fs/dbus_getfilestatus_test.go" "TestDBusServer_GetFileStatus_InvalidPaths" "TestIT_FS_DBus_GetFileStatus_InvalidPaths"
add_test_prefix "internal/fs/dbus_getfilestatus_test.go" "TestDBusServer_GetFileStatus_StatusChanges" "TestIT_FS_DBus_GetFileStatus_StatusChanges"
add_test_prefix "internal/fs/dbus_getfilestatus_test.go" "TestDBusServer_GetFileStatus_SpecialCharacters" "TestIT_FS_DBus_GetFileStatus_SpecialCharacters"

echo "Processing internal/fs/file_operations_test.go..."
add_test_prefix "internal/fs/file_operations_test.go" "TestFileCreationMarksMetadataDirty" "TestIT_FS_FileOps_FileCreationMarksMetadataDirty"
add_test_prefix "internal/fs/file_operations_test.go" "TestMkdirStateReflectsConnectivity" "TestIT_FS_FileOps_MkdirStateReflectsConnectivity"

echo "Processing internal/fs/metadata_operations_test.go..."
add_test_prefix "internal/fs/metadata_operations_test.go" "TestRenameRecordsOfflineChange" "TestIT_FS_Metadata_RenameRecordsOfflineChange"

echo "Processing internal/fs/dir_pending_test.go..."
add_test_prefix "internal/fs/dir_pending_test.go" "TestDirectoryPendingRemoteVisibilitySurvivesRefresh" "TestIT_FS_DirPending_RemoteVisibilitySurvivesRefresh"

echo "Processing internal/fs/file_status_profile_test.go..."
add_test_prefix "internal/fs/file_status_profile_test.go" "TestDocumentPerformanceCharacteristics" "TestIT_FS_FileStatus_DocumentPerformanceCharacteristics"
add_test_prefix "internal/fs/file_status_profile_test.go" "TestIdentifyBottlenecks" "TestIT_FS_FileStatus_IdentifyBottlenecks"
add_test_prefix "internal/fs/file_status_profile_test.go" "TestProfileMemoryUsage" "TestIT_FS_FileStatus_ProfileMemoryUsage"

echo "Processing internal/fs/upload_manager_test.go..."
add_test_prefix "internal/fs/upload_manager_test.go" "TestUploadManagerQueuesDirtyStateInMetadata" "TestIT_FS_Upload_ManagerQueuesDirtyStateInMetadata"
add_test_prefix "internal/fs/upload_manager_test.go" "TestUploadManagerHydratesMetadataOnCompletion" "TestIT_FS_Upload_ManagerHydratesMetadataOnCompletion"
add_test_prefix "internal/fs/upload_manager_test.go" "TestUploadManagerSetsErrorStateOnFailure" "TestIT_FS_Upload_ManagerSetsErrorStateOnFailure"

# Configuration tests (unit tests - no auth required)
echo "Processing cmd/common/config_test.go..."
add_test_prefix "cmd/common/config_test.go" "TestDefaultDeltaIntervalIsFiveMinutes" "TestUT_CMD_Config_DefaultDeltaIntervalIsFiveMinutes"
add_test_prefix "cmd/common/config_test.go" "TestValidateConfigOverlayPolicy" "TestUT_CMD_Config_ValidateConfigOverlayPolicy"
add_test_prefix "cmd/common/config_test.go" "TestValidateRealtimeConfigDefaults" "TestUT_CMD_Config_ValidateRealtimeConfigDefaults"
add_test_prefix "cmd/common/config_test.go" "TestDefaultActiveDeltaTuning" "TestUT_CMD_Config_DefaultActiveDeltaTuning"
add_test_prefix "cmd/common/config_test.go" "TestValidateConfigResetsInvalidActiveDeltaTuning" "TestUT_CMD_Config_ValidateConfigResetsInvalidActiveDeltaTuning"
add_test_prefix "cmd/common/config_test.go" "TestHydrationConfigDefaultsAndValidation" "TestUT_CMD_Config_HydrationConfigDefaultsAndValidation"
add_test_prefix "cmd/common/config_test.go" "TestMetadataQueueDefaultsAndValidation" "TestUT_CMD_Config_MetadataQueueDefaultsAndValidation"
add_test_prefix "cmd/common/config_test.go" "TestRealtimeFallbackValidationBounds" "TestUT_CMD_Config_RealtimeFallbackValidationBounds"

echo "Processing cmd/common/setup_test.go..."
add_test_prefix "cmd/common/setup_test.go" "TestMain" "TestUT_CMD_Setup_Main"

echo "Processing cmd/onemount/main_test.go..."
add_test_prefix "cmd/onemount/main_test.go" "TestToRealtimeOptionsCopiesPollingOnly" "TestUT_CMD_Main_ToRealtimeOptionsCopiesPollingOnly"

# XDG config tests (unit tests)
echo "Processing internal/config/xdg_property_test.go..."
add_test_prefix "internal/config/xdg_property_test.go" "TestXDGConfigHomeRespected" "TestUT_Config_XDGConfigHomeRespected"
add_test_prefix "internal/config/xdg_property_test.go" "TestXDGCacheHomeRespected" "TestUT_Config_XDGCacheHomeRespected"

# Metadata tests (unit tests - use mocks)
echo "Processing internal/metadata/entry_test.go..."
add_test_prefix "internal/metadata/entry_test.go" "TestEntryValidateDefaults" "TestUT_Metadata_EntryValidateDefaults"
add_test_prefix "internal/metadata/entry_test.go" "TestEntryValidateRejectsBadState" "TestUT_Metadata_EntryValidateRejectsBadState"
add_test_prefix "internal/metadata/entry_test.go" "TestEntryValidateRejectsOverlayPolicy" "TestUT_Metadata_EntryValidateRejectsOverlayPolicy"

echo "Processing internal/metadata/manager_test.go..."
add_test_prefix "internal/metadata/manager_test.go" "TestStateManagerHydrationLifecycle" "TestUT_Metadata_StateManagerHydrationLifecycle"
add_test_prefix "internal/metadata/manager_test.go" "TestStateManagerRejectsInvalidTransition" "TestUT_Metadata_StateManagerRejectsInvalidTransition"
add_test_prefix "internal/metadata/manager_test.go" "TestStateManagerErrorTransition" "TestUT_Metadata_StateManagerErrorTransition"
add_test_prefix "internal/metadata/manager_test.go" "TestStateManagerTransitionTable" "TestUT_Metadata_StateManagerTransitionTable"

echo "Processing internal/metadata/store_test.go..."
add_test_prefix "internal/metadata/store_test.go" "TestBoltStoreSaveAndGet" "TestUT_Metadata_BoltStoreSaveAndGet"
add_test_prefix "internal/metadata/store_test.go" "TestBoltStoreUpdate" "TestUT_Metadata_BoltStoreUpdate"

# Test framework tests (unit tests)
echo "Processing internal/testutil/framework/load_test_scenarios_test.go..."
add_test_prefix "internal/testutil/framework/load_test_scenarios_test.go" "TestLoadTestScenarios" "TestUT_Framework_LoadTestScenarios"

echo "Processing internal/testutil/framework/network_simulator_test.go..."
add_test_prefix "internal/testutil/framework/network_simulator_test.go" "TestNetworkSimulator_SetConditions" "TestUT_Framework_NetworkSimulator_SetConditions"
add_test_prefix "internal/testutil/framework/network_simulator_test.go" "TestNetworkSimulator_ApplyPreset" "TestUT_Framework_NetworkSimulator_ApplyPreset"
add_test_prefix "internal/testutil/framework/network_simulator_test.go" "TestNetworkSimulator_DisconnectReconnect" "TestUT_Framework_NetworkSimulator_DisconnectReconnect"
add_test_prefix "internal/testutil/framework/network_simulator_test.go" "TestNetworkSimulator_SimulateNetworkDelay" "TestUT_Framework_NetworkSimulator_SimulateNetworkDelay"
add_test_prefix "internal/testutil/framework/network_simulator_test.go" "TestNetworkSimulator_SimulatePacketLoss" "TestUT_Framework_NetworkSimulator_SimulatePacketLoss"
add_test_prefix "internal/testutil/framework/network_simulator_test.go" "TestNetworkSimulator_SimulateNetworkError" "TestUT_Framework_NetworkSimulator_SimulateNetworkError"
add_test_prefix "internal/testutil/framework/network_simulator_test.go" "TestNetworkSimulator_RegisterProvider" "TestUT_Framework_NetworkSimulator_RegisterProvider"
add_test_prefix "internal/testutil/framework/network_simulator_test.go" "TestWithNetworkSimulator" "TestUT_Framework_WithNetworkSimulator"

# Graph debug tests (unit tests)
echo "Processing internal/graph/debug/mock_test.go..."
add_test_prefix "internal/graph/debug/mock_test.go" "TestMockPackage" "TestUT_Graph_Debug_MockPackage"

echo ""
echo "=== Labeling Complete ==="
echo "All unlabeled tests have been labeled with appropriate prefixes"
echo ""
echo "Summary:"
echo "  - TestIT_: Integration tests (require auth)"
echo "  - TestUT_: Unit tests (no auth required)"
echo ""
echo "Next steps:"
echo "  1. Verify tests compile: go build ./..."
echo "  2. Run unit tests: docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests"
echo "  3. Run integration tests: docker compose -f docker/compose/docker-compose.test.yml -f docker/compose/docker-compose.auth.yml run --rm integration-tests"
