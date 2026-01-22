#!/bin/bash
# Script to label final remaining unlabeled tests
# Task 46.1.2: Label all unlabeled tests (Part 3 - Final)

set -e

echo "=== Labeling Final Unlabeled Tests ==="
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

# FUSE metadata local test (unit test)
echo "Processing internal/fs/fuse_metadata_local_test.go..."
add_test_prefix "internal/fs/fuse_metadata_local_test.go" "TestGetChildMissingDoesNotHitGraph" "TestUT_FS_FUSEMetadata_GetChildMissingDoesNotHitGraph"

# Test helper tests (unit tests)
echo "Processing internal/testutil/helpers/file_test.go..."
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_CreateTestFile" "TestUT_Helpers_FileTestHelper_CreateTestFile"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_CreateTestDir" "TestUT_Helpers_FileTestHelper_CreateTestDir"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_CreateTempDir" "TestUT_Helpers_FileTestHelper_CreateTempDir"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_CreateTempFile" "TestUT_Helpers_FileTestHelper_CreateTempFile"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_FileExists" "TestUT_Helpers_FileTestHelper_FileExists"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_FileContains" "TestUT_Helpers_FileTestHelper_FileContains"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_AssertFileExists" "TestUT_Helpers_FileTestHelper_AssertFileExists"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_AssertFileNotExists" "TestUT_Helpers_FileTestHelper_AssertFileNotExists"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_AssertFileContains" "TestUT_Helpers_FileTestHelper_AssertFileContains"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_AssertFileContent" "TestUT_Helpers_FileTestHelper_AssertFileContent"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestFileTestHelper_CaptureFileSystemState" "TestUT_Helpers_FileTestHelper_CaptureFileSystemState"
add_test_prefix "internal/testutil/helpers/file_test.go" "TestConvenienceFunctions" "TestUT_Helpers_ConvenienceFunctions"

# Unit test framework tests (unit tests)
echo "Processing internal/testutil/framework/unit_test_framework_test.go..."
add_test_prefix "internal/testutil/framework/unit_test_framework_test.go" "TestUnitTestFixture" "TestUT_Framework_UnitTestFixture"
add_test_prefix "internal/testutil/framework/unit_test_framework_test.go" "TestMock" "TestUT_Framework_Mock"
add_test_prefix "internal/testutil/framework/unit_test_framework_test.go" "TestTableTests" "TestUT_Framework_TableTests"
add_test_prefix "internal/testutil/framework/unit_test_framework_test.go" "TestAssert" "TestUT_Framework_Assert"
add_test_prefix "internal/testutil/framework/unit_test_framework_test.go" "TestEdgeCaseGenerator" "TestUT_Framework_EdgeCaseGenerator"
add_test_prefix "internal/testutil/framework/unit_test_framework_test.go" "TestErrorConditions" "TestUT_Framework_ErrorConditions"

# System test environment tests (unit tests)
echo "Processing internal/testutil/framework/system_test_env_test.go..."
add_test_prefix "internal/testutil/framework/system_test_env_test.go" "TestSystemTestEnvironment_Setup" "TestUT_Framework_SystemTestEnvironment_Setup"
add_test_prefix "internal/testutil/framework/system_test_env_test.go" "TestSystemTestEnvironment_DataGenerator" "TestUT_Framework_SystemTestEnvironment_DataGenerator"
add_test_prefix "internal/testutil/framework/system_test_env_test.go" "TestSystemTestEnvironment_ConfigManager" "TestUT_Framework_SystemTestEnvironment_ConfigManager"
add_test_prefix "internal/testutil/framework/system_test_env_test.go" "TestSystemTestEnvironment_Verifier" "TestUT_Framework_SystemTestEnvironment_Verifier"
add_test_prefix "internal/testutil/framework/system_test_env_test.go" "TestSystemTestEnvironment_Scenarios" "TestUT_Framework_SystemTestEnvironment_Scenarios"
add_test_prefix "internal/testutil/framework/system_test_env_test.go" "TestCommonSystemScenarios" "TestUT_Framework_CommonSystemScenarios"
add_test_prefix "internal/testutil/framework/system_test_env_test.go" "TestSystemTestEnvironment_SignalHandling" "TestUT_Framework_SystemTestEnvironment_SignalHandling"

# TestMain functions (unit test setup)
echo "Processing TestMain functions..."
add_test_prefix "internal/testutil/framework/setup_test.go" "TestMain" "TestUT_Framework_Main"
add_test_prefix "internal/graph/setup_test.go" "TestMain" "TestUT_Graph_Main"

# Throttler tests (unit tests)
echo "Processing internal/util/throttler_test.go..."
add_test_prefix "internal/util/throttler_test.go" "TestBandwidthThrottler_Disabled" "TestUT_Util_BandwidthThrottler_Disabled"
add_test_prefix "internal/util/throttler_test.go" "TestBandwidthThrottler_BasicThrottling" "TestUT_Util_BandwidthThrottler_BasicThrottling"
add_test_prefix "internal/util/throttler_test.go" "TestBandwidthThrottler_ContextCancellation" "TestUT_Util_BandwidthThrottler_ContextCancellation"
add_test_prefix "internal/util/throttler_test.go" "TestBandwidthThrottler_Reset" "TestUT_Util_BandwidthThrottler_Reset"
add_test_prefix "internal/util/throttler_test.go" "TestBandwidthThrottler_SetLimit" "TestUT_Util_BandwidthThrottler_SetLimit"
add_test_prefix "internal/util/throttler_test.go" "TestBandwidthThrottler_ConcurrentAccess" "TestUT_Util_BandwidthThrottler_ConcurrentAccess"
add_test_prefix "internal/util/throttler_test.go" "TestBandwidthThrottler_SmallTransfers" "TestUT_Util_BandwidthThrottler_SmallTransfers"

echo ""
echo "=== Labeling Complete (Final) ==="
echo "ALL unlabeled tests have been labeled!"
echo ""
echo "Total tests labeled in this run: 35"
echo ""
echo "Summary of all labeling:"
echo "  - Part 1: 59 tests"
echo "  - Part 2: ~150 tests"
echo "  - Part 3: 35 tests"
echo "  - TOTAL: ~244 tests labeled"
echo ""
echo "Next steps:"
echo "  1. Verify tests compile: go build ./..."
echo "  2. Run unit tests: docker compose -f docker/compose/docker-compose.test.yml run --rm unit-tests"
echo "  3. Run integration tests: docker compose -f docker/compose/docker-compose.test.yml -f docker/compose/docker-compose.auth.yml run --rm integration-tests"
