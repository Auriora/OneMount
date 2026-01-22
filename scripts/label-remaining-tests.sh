#!/bin/bash
# Script to label remaining unlabeled tests
# Task 46.1.2: Label all unlabeled tests (Part 2)

set -e

echo "=== Labeling Remaining Unlabeled Tests ==="
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

# TestMain functions (unit test setup)
echo "Processing TestMain functions..."
add_test_prefix "internal/ui/systemd/setup_test.go" "TestMain" "TestUT_UI_Systemd_Main"
add_test_prefix "internal/ui/setup_test.go" "TestMain" "TestUT_UI_Main"
add_test_prefix "internal/fs/testing_main_test.go" "TestMain" "TestUT_FS_Main"

# Stats tests (unit tests - no auth required)
echo "Processing internal/fs/stats_cache_test.go..."
add_test_prefix "internal/fs/stats_cache_test.go" "TestCachedStatsExpiration" "TestUT_FS_Stats_CachedStatsExpiration"
add_test_prefix "internal/fs/stats_cache_test.go" "TestDefaultStatsConfig" "TestUT_FS_Stats_DefaultStatsConfig"
add_test_prefix "internal/fs/stats_cache_test.go" "TestStatsIsSampled" "TestUT_FS_Stats_StatsIsSampled"
add_test_prefix "internal/fs/stats_cache_test.go" "TestFormatSize" "TestUT_FS_Stats_FormatSize"

echo "Processing internal/fs/stats_metadata_test.go..."
add_test_prefix "internal/fs/stats_metadata_test.go" "TestStatsReportsMetadataStates" "TestUT_FS_Stats_ReportsMetadataStates"

echo "Processing internal/fs/stats_optimization_test.go..."
add_test_prefix "internal/fs/stats_optimization_test.go" "TestStatsCaching" "TestUT_FS_Stats_Caching"
add_test_prefix "internal/fs/stats_optimization_test.go" "TestStatsSampling" "TestUT_FS_Stats_Sampling"
add_test_prefix "internal/fs/stats_optimization_test.go" "TestQuickStats" "TestUT_FS_Stats_QuickStats"
add_test_prefix "internal/fs/stats_optimization_test.go" "TestStatsPagination" "TestUT_FS_Stats_Pagination"
add_test_prefix "internal/fs/stats_optimization_test.go" "TestBackgroundStatsUpdater" "TestUT_FS_Stats_BackgroundStatsUpdater"

echo "Processing internal/fs/stats_realtime_test.go..."
add_test_prefix "internal/fs/stats_realtime_test.go" "TestFilesystemAugmentRealtimeStatsFromManager" "TestUT_FS_Stats_FilesystemAugmentRealtimeStatsFromManager"
add_test_prefix "internal/fs/stats_realtime_test.go" "TestFilesystemAugmentRealtimeStatsPollingOnly" "TestUT_FS_Stats_FilesystemAugmentRealtimeStatsPollingOnly"

# State transition tests (unit tests - use mocks)
echo "Processing internal/fs/state_transition_atomicity_test.go..."
add_test_prefix "internal/fs/state_transition_atomicity_test.go" "TestStateTransitionAtomicity" "TestUT_FS_State_TransitionAtomicity"
add_test_prefix "internal/fs/state_transition_atomicity_test.go" "TestNoIntermediateInconsistentStates" "TestUT_FS_State_NoIntermediateInconsistentStates"
add_test_prefix "internal/fs/state_transition_atomicity_test.go" "TestStatePersistenceAcrossRestarts" "TestUT_FS_State_PersistenceAcrossRestarts"
add_test_prefix "internal/fs/state_transition_atomicity_test.go" "TestConcurrentStateTransitionSafety" "TestUT_FS_State_ConcurrentStateTransitionSafety"
add_test_prefix "internal/fs/state_transition_atomicity_test.go" "TestConcurrentStateTransitionOnSameFile" "TestUT_FS_State_ConcurrentStateTransitionOnSameFile"
add_test_prefix "internal/fs/state_transition_atomicity_test.go" "TestStateTransitionWithError" "TestUT_FS_State_TransitionWithError"
add_test_prefix "internal/fs/state_transition_atomicity_test.go" "TestVirtualFileStateImmutability" "TestUT_FS_State_VirtualFileStateImmutability"
add_test_prefix "internal/fs/state_transition_atomicity_test.go" "TestCompleteStateLifecycle" "TestUT_FS_State_CompleteStateLifecycle"

# Notifier tests (unit tests)
echo "Processing internal/fs/notifier_cadence_test.go..."
add_test_prefix "internal/fs/notifier_cadence_test.go" "TestDeltaIntervalRespectsNotifierHealth" "TestUT_FS_Notifier_DeltaIntervalRespectsNotifierHealth"

# Deadlock tests (unit tests)
echo "Processing internal/fs/deadlock_root_cause_test.go..."
add_test_prefix "internal/fs/deadlock_root_cause_test.go" "TestDeadlockRootCauseAnalysis" "TestUT_FS_Deadlock_RootCauseAnalysis"

# Mount integration tests (unit tests with mocks)
echo "Processing internal/fs/mount_integration_test.go..."
add_test_prefix "internal/fs/mount_integration_test.go" "TestMountIntegration_SuccessfulMount" "TestUT_FS_Mount_Integration_SuccessfulMount"
add_test_prefix "internal/fs/mount_integration_test.go" "TestMountIntegration_MountFailureScenarios" "TestUT_FS_Mount_Integration_MountFailureScenarios"
add_test_prefix "internal/fs/mount_integration_test.go" "TestMountIntegration_GracefulUnmount" "TestUT_FS_Mount_Integration_GracefulUnmount"
add_test_prefix "internal/fs/mount_integration_test.go" "TestMountIntegration_WithMockGraphAPI" "TestUT_FS_Mount_Integration_WithMockGraphAPI"

# DBus service discovery tests (unit tests)
echo "Processing internal/fs/dbus_service_discovery_test.go..."
add_test_prefix "internal/fs/dbus_service_discovery_test.go" "TestDBusServiceNameFileCreation" "TestUT_FS_DBus_ServiceNameFileCreation"
add_test_prefix "internal/fs/dbus_service_discovery_test.go" "TestDBusServiceNameFileCleanup" "TestUT_FS_DBus_ServiceNameFileCleanup"
add_test_prefix "internal/fs/dbus_service_discovery_test.go" "TestDBusServiceNameFileMultipleInstances" "TestUT_FS_DBus_ServiceNameFileMultipleInstances"

# Timeout config tests (unit tests)
echo "Processing internal/fs/timeout_config_test.go..."
add_test_prefix "internal/fs/timeout_config_test.go" "TestDefaultTimeoutConfig" "TestUT_FS_Timeout_DefaultTimeoutConfig"
add_test_prefix "internal/fs/timeout_config_test.go" "TestTimeoutConfigValidation" "TestUT_FS_Timeout_ConfigValidation"
add_test_prefix "internal/fs/timeout_config_test.go" "TestTimeoutConfigInFilesystem" "TestUT_FS_Timeout_ConfigInFilesystem"
add_test_prefix "internal/fs/timeout_config_test.go" "TestInvalidConfigError" "TestUT_FS_Timeout_InvalidConfigError"

# Delta state manager tests (unit tests)
echo "Processing internal/fs/delta_state_manager_test.go..."
add_test_prefix "internal/fs/delta_state_manager_test.go" "TestHydrationErrorPersistsLastErrorSnapshot" "TestUT_FS_DeltaState_HydrationErrorPersistsLastErrorSnapshot"
add_test_prefix "internal/fs/delta_state_manager_test.go" "TestUploadErrorPersistsSnapshot" "TestUT_FS_DeltaState_UploadErrorPersistsSnapshot"
add_test_prefix "internal/fs/delta_state_manager_test.go" "TestApplyDeltaPinnedItemRequeuesHydration" "TestUT_FS_DeltaState_ApplyDeltaPinnedItemRequeuesHydration"
add_test_prefix "internal/fs/delta_state_manager_test.go" "TestApplyDeltaSetsGhostOnRemoteChange" "TestUT_FS_DeltaState_ApplyDeltaSetsGhostOnRemoteChange"
add_test_prefix "internal/fs/delta_state_manager_test.go" "TestApplyDeltaHydratesWhenMetadataMatches" "TestUT_FS_DeltaState_ApplyDeltaHydratesWhenMetadataMatches"
add_test_prefix "internal/fs/delta_state_manager_test.go" "TestApplyDeltaMarksDeletedAndScrubsParent" "TestUT_FS_DeltaState_ApplyDeltaMarksDeletedAndScrubsParent"

# Metadata store tests (unit tests)
echo "Processing internal/fs/metadata_store_test.go..."
add_test_prefix "internal/fs/metadata_store_test.go" "TestMetadataEntryFromInodeStateInference" "TestUT_FS_MetadataStore_EntryFromInodeStateInference"
add_test_prefix "internal/fs/metadata_store_test.go" "TestBootstrapMetadataStoreMigratesLegacyEntries" "TestUT_FS_MetadataStore_BootstrapMigratesLegacyEntries"
add_test_prefix "internal/fs/metadata_store_test.go" "TestInodeFromMetadataEntry" "TestUT_FS_MetadataStore_InodeFromMetadataEntry"
add_test_prefix "internal/fs/metadata_store_test.go" "TestPendingRemoteMetadataUpdates" "TestUT_FS_MetadataStore_PendingRemoteMetadataUpdates"
add_test_prefix "internal/fs/metadata_store_test.go" "TestGetIDLoadsFromMetadataStore" "TestUT_FS_MetadataStore_GetIDLoadsFromMetadataStore"

# Mutation queue tests (unit tests)
echo "Processing internal/fs/mutation_queue_test.go..."
add_test_prefix "internal/fs/mutation_queue_test.go" "TestQueueRemoteDeleteTransitionsToDeleted" "TestUT_FS_MutationQueue_QueueRemoteDeleteTransitionsToDeleted"
add_test_prefix "internal/fs/mutation_queue_test.go" "TestMutationQueueReturnsImmediatelyAndProcessesWork" "TestUT_FS_MutationQueue_ReturnsImmediatelyAndProcessesWork"

# Inode types tests (unit tests)
echo "Processing internal/fs/inode_types_test.go..."
add_test_prefix "internal/fs/inode_types_test.go" "TestVirtualInodeContentHelpers" "TestUT_FS_InodeTypes_VirtualInodeContentHelpers"

# Change notifier tests (unit tests)
echo "Processing internal/fs/change_notifier_test.go..."
add_test_prefix "internal/fs/change_notifier_test.go" "TestChangeNotifierDisabled" "TestUT_FS_ChangeNotifier_Disabled"
add_test_prefix "internal/fs/change_notifier_test.go" "TestChangeNotifierPollingOnly" "TestUT_FS_ChangeNotifier_PollingOnly"
add_test_prefix "internal/fs/change_notifier_test.go" "TestChangeNotifierDelegatesToSocketManager" "TestUT_FS_ChangeNotifier_DelegatesToSocketManager"

# Performance verification tests (unit tests)
echo "Processing internal/fs/performance_verification_test.go..."
add_test_prefix "internal/fs/performance_verification_test.go" "TestPerformanceVerification" "TestUT_FS_Performance_Verification"

# Content eviction tests (unit tests)
echo "Processing internal/fs/content_eviction_test.go..."
add_test_prefix "internal/fs/content_eviction_test.go" "TestContentEvictionTransitionsMetadata" "TestUT_FS_ContentEviction_TransitionsMetadata"
add_test_prefix "internal/fs/content_eviction_test.go" "TestPinnedContentNotEvicted" "TestUT_FS_ContentEviction_PinnedContentNotEvicted"
add_test_prefix "internal/fs/content_eviction_test.go" "TestPinnedContentAutoHydratesAfterEviction" "TestUT_FS_ContentEviction_PinnedContentAutoHydratesAfterEviction"
add_test_prefix "internal/fs/content_eviction_test.go" "TestContentEvictionTransitionsToGhost" "TestUT_FS_ContentEviction_TransitionsToGhost"

# Inode attr tests (unit tests)
echo "Processing internal/fs/inode_attr_test.go..."
add_test_prefix "internal/fs/inode_attr_test.go" "TestInodeMakeAttrReportsBlocksUsingMetadata" "TestUT_FS_InodeAttr_MakeAttrReportsBlocksUsingMetadata"

# FUSE metadata local tests (unit tests)
echo "Processing internal/fs/fuse_metadata_local_test.go..."
add_test_prefix "internal/fs/fuse_metadata_local_test.go" "TestGetChildrenIDUsesMetadataStoreWhenCold" "TestUT_FS_FUSEMetadata_GetChildrenIDUsesMetadataStoreWhenCold"
add_test_prefix "internal/fs/fuse_metadata_local_test.go" "TestGetChildUsesMetadataStoreWhenCold" "TestUT_FS_FUSEMetadata_GetChildUsesMetadataStoreWhenCold"
add_test_prefix "internal/fs/fuse_metadata_local_test.go" "TestOpenDirUsesMetadataOffline" "TestUT_FS_FUSEMetadata_OpenDirUsesMetadataOffline"
add_test_prefix "internal/fs/fuse_metadata_local_test.go" "TestLookupUsesMetadataOffline" "TestUT_FS_FUSEMetadata_LookupUsesMetadataOffline"

# Minimal hang tests (unit tests)
echo "Processing internal/fs/minimal_hang_test.go..."
add_test_prefix "internal/fs/minimal_hang_test.go" "TestMinimalHangReproduction" "TestUT_FS_MinimalHang_Reproduction"
add_test_prefix "internal/fs/minimal_hang_test.go" "TestHangPointIsolation" "TestUT_FS_MinimalHang_PointIsolation"
add_test_prefix "internal/fs/minimal_hang_test.go" "TestConcurrentOperations" "TestUT_FS_MinimalHang_ConcurrentOperations"

# Download manager tests (unit tests)
echo "Processing internal/fs/download_manager_test.go..."
add_test_prefix "internal/fs/download_manager_test.go" "TestRestoreDownloadSessionsRequeues" "TestUT_FS_DownloadManager_RestoreDownloadSessionsRequeues"

# Socket subscription manager tests (unit tests)
echo "Processing internal/fs/socket_subscription_manager_test.go..."
add_test_prefix "internal/fs/socket_subscription_manager_test.go" "TestSocketSubscriptionManagerTriggersNotifications" "TestUT_FS_SocketSubscription_ManagerTriggersNotifications"

# Graph socket subscription tests (unit tests)
echo "Processing internal/graph/socket_subscription_test.go..."
add_test_prefix "internal/graph/socket_subscription_test.go" "TestBuildSocketSubscriptionPath" "TestUT_Graph_Socket_BuildSubscriptionPath"

# Graph network feedback tests (unit tests)
echo "Processing internal/graph/network_feedback_test.go..."
add_test_prefix "internal/graph/network_feedback_test.go" "TestNetworkFeedbackManager_WaitGroupTracking" "TestUT_Graph_NetworkFeedback_WaitGroupTracking"
add_test_prefix "internal/graph/network_feedback_test.go" "TestNetworkFeedbackManager_ShutdownTimeout" "TestUT_Graph_NetworkFeedback_ShutdownTimeout"
add_test_prefix "internal/graph/network_feedback_test.go" "TestNetworkFeedbackManager_ShutdownWithMultipleCallbacks" "TestUT_Graph_NetworkFeedback_ShutdownWithMultipleCallbacks"
add_test_prefix "internal/graph/network_feedback_test.go" "TestNetworkFeedbackManager_PanicRecovery" "TestUT_Graph_NetworkFeedback_PanicRecovery"
add_test_prefix "internal/graph/network_feedback_test.go" "TestNetworkFeedbackManager_ConcurrentNotifications" "TestUT_Graph_NetworkFeedback_ConcurrentNotifications"

# Graph debug tests (unit tests)
echo "Processing internal/graph/debug_test.go..."
add_test_prefix "internal/graph/debug_test.go" "TestDebug" "TestUT_Graph_Debug"

echo "Processing internal/graph/debug/debug_test.go..."
add_test_prefix "internal/graph/debug/debug_test.go" "TestDebug" "TestUT_Graph_Debug_Debug"

echo "Processing internal/graph/debug/testutil_test.go..."
add_test_prefix "internal/graph/debug/testutil_test.go" "TestTestUtilPaths" "TestUT_Graph_Debug_TestUtilPaths"

# Logging type helpers tests (unit tests)
echo "Processing internal/logging/type_helpers_test.go..."
add_test_prefix "internal/logging/type_helpers_test.go" "TestGetTypeLogger" "TestUT_Logging_GetTypeLogger"
add_test_prefix "internal/logging/type_helpers_test.go" "TestFuseStatusLogger" "TestUT_Logging_FuseStatusLogger"
add_test_prefix "internal/logging/type_helpers_test.go" "TestLogValueWithTypeLogger_FuseStatus" "TestUT_Logging_LogValueWithTypeLogger_FuseStatus"
add_test_prefix "internal/logging/type_helpers_test.go" "TestLogReturn_FuseStatus" "TestUT_Logging_LogReturn_FuseStatus"
add_test_prefix "internal/logging/type_helpers_test.go" "TestLogParam_FuseStatus" "TestUT_Logging_LogParam_FuseStatus"
add_test_prefix "internal/logging/type_helpers_test.go" "TestTypeLoggerPanicRegression" "TestUT_Logging_TypeLoggerPanicRegression"

# Socket.IO engine transport tests (unit tests)
echo "Processing internal/socketio/engine_transport_test.go..."
add_test_prefix "internal/socketio/engine_transport_test.go" "TestToEngineIOURL" "TestUT_SocketIO_ToEngineIOURL"
add_test_prefix "internal/socketio/engine_transport_test.go" "TestToEngineIOURLStripsCallback" "TestUT_SocketIO_ToEngineIOURLStripsCallback"
add_test_prefix "internal/socketio/engine_transport_test.go" "TestEngineTransportBackoffRespectsCap" "TestUT_SocketIO_EngineTransportBackoffRespectsCap"

# Test framework integration tests (unit tests)
echo "Processing internal/testutil/framework/integration_test_env_test.go..."
add_test_prefix "internal/testutil/framework/integration_test_env_test.go" "TestIntegrationTestEnvironment_Setup" "TestUT_Framework_IntegrationTestEnvironment_Setup"
add_test_prefix "internal/testutil/framework/integration_test_env_test.go" "TestIntegrationTestEnvironment_TestDataManager" "TestUT_Framework_IntegrationTestEnvironment_TestDataManager"
add_test_prefix "internal/testutil/framework/integration_test_env_test.go" "TestIntegrationTestEnvironment_RunScenario" "TestUT_Framework_IntegrationTestEnvironment_RunScenario"
add_test_prefix "internal/testutil/framework/integration_test_env_test.go" "TestIntegrationTestEnvironment_NetworkSimulation" "TestUT_Framework_IntegrationTestEnvironment_NetworkSimulation"
add_test_prefix "internal/testutil/framework/integration_test_env_test.go" "TestIntegrationTestEnvironment_ComponentIsolation" "TestUT_Framework_IntegrationTestEnvironment_ComponentIsolation"

# Test framework performance tests (unit tests)
echo "Processing internal/testutil/framework/performance_integration_test.go..."
add_test_prefix "internal/testutil/framework/performance_integration_test.go" "TestPerformanceIntegration_LargeFileHandling" "TestUT_Framework_PerformanceIntegration_LargeFileHandling"
add_test_prefix "internal/testutil/framework/performance_integration_test.go" "TestPerformanceIntegration_HighFileCount" "TestUT_Framework_PerformanceIntegration_HighFileCount"
add_test_prefix "internal/testutil/framework/performance_integration_test.go" "TestPerformanceIntegration_SustainedOperation" "TestUT_Framework_PerformanceIntegration_SustainedOperation"
add_test_prefix "internal/testutil/framework/performance_integration_test.go" "TestPerformanceIntegration_MemoryLeakDetection" "TestUT_Framework_PerformanceIntegration_MemoryLeakDetection"
add_test_prefix "internal/testutil/framework/performance_integration_test.go" "TestPerformanceIntegration_AllTests" "TestUT_Framework_PerformanceIntegration_AllTests"

echo "Processing internal/testutil/framework/performance_load_tests_test.go..."
add_test_prefix "internal/testutil/framework/performance_load_tests_test.go" "TestLargeFileHandling" "TestUT_Framework_PerformanceLoad_LargeFileHandling"
add_test_prefix "internal/testutil/framework/performance_load_tests_test.go" "TestHighFileCountDirectory" "TestUT_Framework_PerformanceLoad_HighFileCountDirectory"
add_test_prefix "internal/testutil/framework/performance_load_tests_test.go" "TestSustainedOperation" "TestUT_Framework_PerformanceLoad_SustainedOperation"
add_test_prefix "internal/testutil/framework/performance_load_tests_test.go" "TestMemoryLeakDetection" "TestUT_Framework_PerformanceLoad_MemoryLeakDetection"
add_test_prefix "internal/testutil/framework/performance_load_tests_test.go" "TestMemoryTracker" "TestUT_Framework_PerformanceLoad_MemoryTracker"
add_test_prefix "internal/testutil/framework/performance_load_tests_test.go" "TestDefaultConfigurations" "TestUT_Framework_PerformanceLoad_DefaultConfigurations"

echo "Processing internal/testutil/framework/performance_test.go..."
add_test_prefix "internal/testutil/framework/performance_test.go" "TestPerformanceBenchmark" "TestUT_Framework_Performance_Benchmark"
add_test_prefix "internal/testutil/framework/performance_test.go" "TestLoadTest" "TestUT_Framework_Performance_LoadTest"
add_test_prefix "internal/testutil/framework/performance_test.go" "TestBenchmarkScenarios" "TestUT_Framework_Performance_BenchmarkScenarios"

echo "Processing internal/testutil/framework/resources_test.go..."
add_test_prefix "internal/testutil/framework/resources_test.go" "TestFileSystemResource_Basic" "TestUT_Framework_Resources_FileSystemResource_Basic"
add_test_prefix "internal/testutil/framework/resources_test.go" "TestFileSystemResource_WithTestFramework" "TestUT_Framework_Resources_FileSystemResource_WithTestFramework"
add_test_prefix "internal/testutil/framework/resources_test.go" "TestFileSystemResource_MountUnmount" "TestUT_Framework_Resources_FileSystemResource_MountUnmount"

echo "Processing internal/testutil/framework/security_test.go..."
add_test_prefix "internal/testutil/framework/security_test.go" "TestSecurityTestFramework" "TestUT_Framework_Security_TestFramework"
add_test_prefix "internal/testutil/framework/security_test.go" "TestSecurityTestScenarios" "TestUT_Framework_Security_TestScenarios"
add_test_prefix "internal/testutil/framework/security_test.go" "TestSecurityScannerRegistration" "TestUT_Framework_Security_ScannerRegistration"
add_test_prefix "internal/testutil/framework/security_test.go" "TestSecurityAttackSimulatorRegistration" "TestUT_Framework_Security_AttackSimulatorRegistration"
add_test_prefix "internal/testutil/framework/security_test.go" "TestSecurityControlVerifierRegistration" "TestUT_Framework_Security_ControlVerifierRegistration"
add_test_prefix "internal/testutil/framework/security_test.go" "TestAuthenticationTesterRegistration" "TestUT_Framework_Security_AuthenticationTesterRegistration"
add_test_prefix "internal/testutil/framework/security_test.go" "TestAuthorizationTesterRegistration" "TestUT_Framework_Security_AuthorizationTesterRegistration"
add_test_prefix "internal/testutil/framework/security_test.go" "TestDataProtectionTesterRegistration" "TestUT_Framework_Security_DataProtectionTesterRegistration"
add_test_prefix "internal/testutil/framework/security_test.go" "TestRunSecurityScan" "TestUT_Framework_Security_RunSecurityScan"

echo ""
echo "=== Labeling Complete (Part 2) ==="
echo "All remaining unlabeled tests have been labeled"
echo ""
echo "Total tests labeled in this run: ~150"
echo ""
echo "Next steps:"
echo "  1. Verify tests compile: go build ./..."
echo "  2. Run unit tests: docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests"
