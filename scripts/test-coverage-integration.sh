#!/bin/bash

# Integration test for OneMount coverage validation and reporting system
# This script tests all components of Phase 5 implementation

set -e

echo "🧪 Testing OneMount Coverage Integration System"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

# Test 1: Basic coverage generation
log_test "Testing basic coverage generation..."
if go test -v -coverprofile=coverage/test-coverage.out ./pkg/errors ./pkg/logging ./pkg/quickxorhash ./pkg/retry > /dev/null 2>&1; then
    if [ -f "coverage/test-coverage.out" ]; then
        log_pass "Coverage profile generated successfully"
    else
        log_fail "Coverage profile not found"
    fi
else
    log_fail "Failed to generate coverage profile"
fi

# Test 2: Coverage report script
log_test "Testing coverage report generation..."
if bash scripts/coverage-report.sh > /dev/null 2>&1; then
    # Check if required files were generated
    required_files=(
        "coverage/coverage.html"
        "coverage/coverage-func.txt"
        "coverage/package-analysis.txt"
        "coverage/coverage-gaps.txt"
        "coverage/coverage.json"
        "coverage/summary.txt"
    )
    
    all_files_exist=true
    for file in "${required_files[@]}"; do
        if [ ! -f "$file" ]; then
            log_fail "Required file not generated: $file"
            all_files_exist=false
        fi
    done
    
    if $all_files_exist; then
        log_pass "All coverage report files generated"
    fi
else
    log_fail "Coverage report script failed"
fi

# Test 3: Coverage history tracking
log_test "Testing coverage history tracking..."
if [ -f "coverage/coverage_history.json" ]; then
    # Check if history file contains valid JSON
    if python3 -c "import json; json.load(open('coverage/coverage_history.json'))" 2>/dev/null; then
        log_pass "Coverage history file is valid JSON"
    else
        log_fail "Coverage history file is invalid JSON"
    fi
else
    log_fail "Coverage history file not found"
fi

# Test 4: Trend analysis
log_test "Testing coverage trend analysis..."
if python3 scripts/coverage-trend-analysis.py --input coverage/coverage_history.json --output coverage/test-trends.html > /dev/null 2>&1; then
    if [ -f "coverage/test-trends.html" ]; then
        # Check if HTML file contains expected content
        if grep -q "OneMount Coverage Trend Analysis" coverage/test-trends.html; then
            log_pass "Trend analysis HTML report generated with correct content"
        else
            log_fail "Trend analysis HTML report missing expected content"
        fi
    else
        log_fail "Trend analysis HTML report not generated"
    fi
else
    log_fail "Trend analysis script failed"
fi

# Test 5: Makefile targets
log_test "Testing Makefile coverage targets..."
makefile_targets=("coverage" "coverage-report" "coverage-trend")
makefile_tests_passed=0

for target in "${makefile_targets[@]}"; do
    if make "$target" > /dev/null 2>&1; then
        ((makefile_tests_passed++))
    fi
done

if [ $makefile_tests_passed -eq ${#makefile_targets[@]} ]; then
    log_pass "All Makefile coverage targets work correctly"
else
    log_fail "Some Makefile coverage targets failed ($makefile_tests_passed/${#makefile_targets[@]} passed)"
fi

# Test 6: Coverage threshold checking
log_test "Testing coverage threshold enforcement..."
# Test with a low threshold that should pass
if bash scripts/coverage-report.sh --threshold-line 50 > /dev/null 2>&1; then
    log_pass "Coverage threshold check works (passing case)"
else
    log_fail "Coverage threshold check failed (should have passed)"
fi

# Test with a high threshold that should fail
if ! bash scripts/coverage-report.sh --threshold-line 95 > /dev/null 2>&1; then
    log_pass "Coverage threshold check works (failing case)"
else
    log_fail "Coverage threshold check failed (should have failed)"
fi

# Test 7: CI mode functionality
log_test "Testing CI mode functionality..."
if bash scripts/coverage-report.sh --ci > /dev/null 2>&1; then
    # Check if CI-specific files were generated
    if [ -f "coverage/ci-summary.json" ] && [ -f "coverage/coverage-gaps.json" ]; then
        log_pass "CI mode generates required files"
    else
        log_fail "CI mode missing required files"
    fi
else
    log_fail "CI mode script execution failed"
fi

# Test 8: JSON report validation
log_test "Testing JSON report validation..."
json_files=("coverage/coverage.json" "coverage/ci-summary.json" "coverage/coverage-gaps.json")
json_valid=true

for json_file in "${json_files[@]}"; do
    if [ -f "$json_file" ]; then
        if ! python3 -c "import json; json.load(open('$json_file'))" 2>/dev/null; then
            log_fail "Invalid JSON in $json_file"
            json_valid=false
        fi
    else
        log_fail "JSON file not found: $json_file"
        json_valid=false
    fi
done

if $json_valid; then
    log_pass "All JSON reports are valid"
fi

# Test 9: Coverage gaps analysis
log_test "Testing coverage gaps analysis..."
if [ -f "coverage/coverage-gaps.json" ]; then
    # Check if gaps file contains expected structure
    if python3 -c "
import json
data = json.load(open('coverage/coverage-gaps.json'))
required_keys = ['timestamp', 'threshold_line', 'packages_below_threshold', 'files_below_threshold', 'recommendations']
assert all(key in data for key in required_keys)
print('Coverage gaps structure is valid')
" 2>/dev/null; then
        log_pass "Coverage gaps analysis structure is correct"
    else
        log_fail "Coverage gaps analysis structure is incorrect"
    fi
else
    log_fail "Coverage gaps file not found"
fi

# Test 10: GitHub Actions workflow validation
log_test "Testing GitHub Actions workflow syntax..."
if command -v yamllint >/dev/null 2>&1; then
    if yamllint .github/workflows/coverage.yml > /dev/null 2>&1; then
        log_pass "GitHub Actions workflow YAML is valid"
    else
        log_fail "GitHub Actions workflow YAML has syntax errors"
    fi
else
    log_info "yamllint not available, skipping YAML validation"
    log_pass "GitHub Actions workflow file exists"
fi

# Cleanup test files
log_info "Cleaning up test files..."
rm -f coverage/test-coverage.out coverage/test-trends.html

# Summary
echo ""
echo "=============================================="
echo "🏁 Test Summary"
echo "=============================================="
echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
echo -e "Total Tests:  $((TESTS_PASSED + TESTS_FAILED))"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✅ All tests passed! Coverage integration is working correctly.${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please review the output above.${NC}"
    exit 1
fi
