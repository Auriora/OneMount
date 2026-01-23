#!/bin/bash
# Script to audit all TestIT_ integration tests for proper auth usage
# Task 46.1.6: Verify integration tests are correctly labeled and use auth

set -euo pipefail

echo "=== Integration Test Audit ==="
echo "Checking all TestIT_* tests for proper auth usage..."
echo ""

# Find all test files with TestIT_ tests
test_files=$(find . -name "*_test.go" -type f -exec grep -l "^func TestIT_" {} \; | sort)

total_tests=0
tests_with_setup=0
tests_with_integration_fixture=0
tests_with_mock_fixture=0
tests_with_skip_logic=0
tests_needing_review=0

declare -a tests_to_review=()

for file in $test_files; do
    echo "Checking: $file"
    
    # Count TestIT_ functions in this file
    test_count=$(grep -c "^func TestIT_" "$file" || true)
    total_tests=$((total_tests + test_count))
    
    # Check if file uses SetupIntegrationFSTestFixture
    if grep -q "SetupIntegrationFSTestFixture" "$file"; then
        tests_with_integration_fixture=$((tests_with_integration_fixture + test_count))
        echo "  ✓ Uses SetupIntegrationFSTestFixture (proper auth handling)"
    # Check if file uses SetupFSTestFixture (auto-detects based on test name)
    elif grep -q "SetupFSTestFixture" "$file"; then
        tests_with_setup=$((tests_with_setup + test_count))
        echo "  ✓ Uses SetupFSTestFixture (auto-detects integration tests)"
    # Check if file uses SetupMockFSTestFixture (should not be used for TestIT_)
    elif grep -q "SetupMockFSTestFixture" "$file"; then
        tests_with_mock_fixture=$((tests_with_mock_fixture + test_count))
        echo "  ⚠️  Uses SetupMockFSTestFixture (should use integration fixture for TestIT_)"
        tests_to_review+=("$file: Uses mock fixture instead of integration fixture")
        tests_needing_review=$((tests_needing_review + test_count))
    # Check if file has explicit skip logic
    elif grep -q 't.Skip.*auth' "$file"; then
        tests_with_skip_logic=$((tests_with_skip_logic + test_count))
        echo "  ✓ Has explicit skip logic for missing auth"
    else
        echo "  ❌ No recognized auth setup pattern found"
        tests_to_review+=("$file: No recognized auth setup pattern")
        tests_needing_review=$((tests_needing_review + test_count))
    fi
    
    echo ""
done

echo "=== Summary ==="
echo "Total TestIT_ tests found: $total_tests"
echo "Tests using SetupFSTestFixture (auto-detect): $tests_with_setup"
echo "Tests using SetupIntegrationFSTestFixture: $tests_with_integration_fixture"
echo "Tests using SetupMockFSTestFixture (⚠️): $tests_with_mock_fixture"
echo "Tests with explicit skip logic: $tests_with_skip_logic"
echo "Tests needing review: $tests_needing_review"
echo ""

if [ ${#tests_to_review[@]} -gt 0 ]; then
    echo "=== Tests Needing Review ==="
    for item in "${tests_to_review[@]}"; do
        echo "  - $item"
    done
    echo ""
fi

# Calculate properly configured tests
properly_configured=$((tests_with_setup + tests_with_integration_fixture + tests_with_skip_logic))
echo "Properly configured: $properly_configured / $total_tests"

if [ $tests_needing_review -eq 0 ]; then
    echo "✅ All integration tests are properly configured!"
    exit 0
else
    echo "⚠️  Some tests need review"
    exit 1
fi
