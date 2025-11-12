#!/bin/bash
# Cache Management Test Script
# Tests content caching, cache hit/miss, expiration, and statistics

set -e

echo "========================================="
echo "Cache Management Verification Tests"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0
SKIPPED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local test_pattern="$2"
    
    echo "----------------------------------------"
    echo "Running: $test_name"
    echo "Pattern: $test_pattern"
    echo "----------------------------------------"
    
    if go test -v -run "$test_pattern" ./internal/fs/ -timeout 2m; then
        echo -e "${GREEN}✓ PASSED${NC}: $test_name"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAILED${NC}: $test_name"
        ((FAILED++))
    fi
    echo ""
}

# Test 11.2: Content Caching
echo "=== Task 11.2: Test Content Caching ==="
run_test "Content Cache Operations" "TestUT_FS_Cache_02_ContentCache_Operations"

# Test 11.3: Cache Hit/Miss (via cache consistency tests)
echo "=== Task 11.3: Test Cache Hit/Miss ==="
run_test "Cache Consistency" "TestUT_FS_Cache_03_CacheConsistency_MultipleOperations"

# Test 11.4: Cache Expiration
echo "=== Task 11.4: Test Cache Expiration ==="
run_test "Cache Invalidation" "TestUT_FS_Cache_01_CacheInvalidation_WorksCorrectly"
run_test "Comprehensive Cache Invalidation" "TestUT_FS_Cache_04_CacheInvalidation_Comprehensive"

# Test 11.5: Cache Statistics (via performance test)
echo "=== Task 11.5: Test Cache Statistics ==="
run_test "Cache Performance" "TestUT_FS_Cache_05_CachePerformance_Operations"

# Summary
echo "========================================="
echo "Test Summary"
echo "========================================="
echo -e "Passed:  ${GREEN}$PASSED${NC}"
echo -e "Failed:  ${RED}$FAILED${NC}"
echo -e "Skipped: ${YELLOW}$SKIPPED${NC}"
echo "========================================="

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
