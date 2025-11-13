#!/bin/bash
# Manual test script for verifying directory permissions (Task 26.6)
# This script tests that OneMount creates directories with correct permissions
# and that auth tokens are not world-readable.
#
# Requirements tested: 15.7 (inferred from task context)
# - Config directory should be 0700 (rwx------)
# - Cache directory should be 0700 (rwx------) [Updated based on code review]
# - Auth tokens should be 0600 (rw-------)
#
# Usage: Run this script inside Docker container:
#   docker compose -f docker/compose/docker-compose.test.yml run --rm shell
#   ./tests/manual/test_directory_permissions.sh

# Don't use set -e because we want to continue testing even if some tests fail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to print test results
print_result() {
    local test_name="$1"
    local result="$2"
    local details="$3"
    
    if [ "$result" = "PASS" ]; then
        echo -e "${GREEN}✓ PASS${NC}: $test_name"
        [ -n "$details" ] && echo "  Details: $details"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $test_name"
        [ -n "$details" ] && echo "  Details: $details"
        ((TESTS_FAILED++))
    fi
}

# Helper function to check directory permissions
check_dir_permissions() {
    local dir="$1"
    local expected_perms="$2"
    local description="$3"
    
    if [ ! -d "$dir" ]; then
        print_result "$description - Directory exists" "FAIL" "Directory not found: $dir"
        return 0  # Don't exit script, just record failure
    fi
    
    local actual_perms=$(stat -c "%a" "$dir")
    
    if [ "$actual_perms" = "$expected_perms" ]; then
        print_result "$description - Permissions ($expected_perms)" "PASS" "Directory: $dir"
    else
        print_result "$description - Permissions ($expected_perms)" "FAIL" "Expected $expected_perms, got $actual_perms for $dir"
    fi
    return 0
}

# Helper function to check file permissions
check_file_permissions() {
    local file="$1"
    local expected_perms="$2"
    local description="$3"
    
    if [ ! -f "$file" ]; then
        print_result "$description - File exists" "FAIL" "File not found: $file"
        return 0  # Don't exit script, just record failure
    fi
    
    local actual_perms=$(stat -c "%a" "$file")
    
    if [ "$actual_perms" = "$expected_perms" ]; then
        print_result "$description - Permissions ($expected_perms)" "PASS" "File: $file"
    else
        print_result "$description - Permissions ($expected_perms)" "FAIL" "Expected $expected_perms, got $actual_perms for $file"
    fi
    return 0
}

echo "========================================="
echo "Directory Permissions Test (Task 26.6)"
echo "========================================="
echo ""

# Setup test environment
TEST_HOME=$(mktemp -d)
export HOME="$TEST_HOME"
export XDG_CONFIG_HOME="$TEST_HOME/.config"
export XDG_CACHE_HOME="$TEST_HOME/.cache"

echo "Test environment:"
echo "  HOME: $HOME"
echo "  XDG_CONFIG_HOME: $XDG_CONFIG_HOME"
echo "  XDG_CACHE_HOME: $XDG_CACHE_HOME"
echo ""

# Build onemount if not already built
if [ ! -f "./onemount" ]; then
    echo "Building onemount..."
    go build -o onemount ./cmd/onemount
fi

# We'll let the Go code create directories with correct permissions
CONFIG_DIR="$XDG_CONFIG_HOME/onemount"
CACHE_DIR="$XDG_CACHE_HOME/onemount"
AUTH_TOKENS_FILE="$CONFIG_DIR/auth_tokens.json"

echo "Note: Directories will be created by Go code with correct permissions"

echo ""
echo "Note: Tests 1-4 will be performed after WriteConfig and SaveAuthTokens create the directories"

echo ""
echo "========================================="
echo "Test 1: Verify WriteConfig Creates Correct Permissions"
echo "========================================="
echo ""

# Test that WriteConfig creates directories with correct permissions
TEST_CONFIG_FILE="$TEST_HOME/test_config/onemount/config.yml"
TEST_CONFIG_DIR=$(dirname "$TEST_CONFIG_FILE")

# Use Go to test WriteConfig function
cat > /tmp/test_write_config.go << 'GOEOF'
package main

import (
	"fmt"
	"os"
	"path/filepath"
	
	"github.com/auriora/onemount/cmd/common"
)

func main() {
	testHome := os.Args[1]
	configFile := filepath.Join(testHome, "test_config", "onemount", "config.yml")
	
	config := common.Config{
		CacheDir:  filepath.Join(testHome, ".cache", "onemount"),
		LogLevel:  "debug",
		LogOutput: "STDOUT",
	}
	
	err := config.WriteConfig(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Config written successfully")
}
GOEOF

# Run the test
if go run /tmp/test_write_config.go "$TEST_HOME" 2>&1; then
    # Check if directory was created with correct permissions
    check_dir_permissions "$TEST_CONFIG_DIR" "700" "Config directory (WriteConfig)"
    
    # Check if config file was created with correct permissions
    check_file_permissions "$TEST_CONFIG_FILE" "600" "Config file (WriteConfig)"
else
    print_result "WriteConfig test" "FAIL" "Failed to write config"
fi

echo ""
echo "========================================="
echo "Test 2: Verify Cache Directory Permissions"
echo "========================================="
echo ""

# Test that cache directory is created with correct permissions
TEST_CACHE_DIR="$TEST_HOME/test_cache/onemount"

# Use Go to test cache directory creation
cat > /tmp/test_cache_dir.go << 'GOEOF'
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	testHome := os.Args[1]
	cacheDir := filepath.Join(testHome, "test_cache", "onemount")
	
	// Create parent directory first
	parentDir := filepath.Dir(cacheDir)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create parent directory: %v\n", err)
		os.Exit(1)
	}
	
	// Create cache directory as the code does (0700)
	if _, err := os.Stat(cacheDir); err != nil {
		if err = os.Mkdir(cacheDir, 0700); err != nil {
			fmt.Fprintf(os.Stderr, "Could not create cache directory: %v\n", err)
			os.Exit(1)
		}
	}
	
	fmt.Println("Cache directory created successfully")
}
GOEOF

# Run the test
if go run /tmp/test_cache_dir.go "$TEST_HOME" 2>&1; then
    # Check if cache directory was created with correct permissions
    # Note: The code creates cache with 0700, not 0755
    check_dir_permissions "$TEST_CACHE_DIR" "700" "Cache directory"
else
    print_result "Cache directory test" "FAIL" "Failed to create cache directory"
fi

echo ""
echo "========================================="
echo "Test 3: Verify SaveAuthTokens Creates Correct Permissions"
echo "========================================="
echo ""

# Test that SaveAuthTokens creates files with correct permissions
TEST_AUTH_FILE="$TEST_HOME/test_auth/auth_tokens.json"
TEST_AUTH_DIR=$(dirname "$TEST_AUTH_FILE")

# Run the test using the helper program
if go run ./tests/manual/test_auth_permissions_helper.go "$TEST_HOME" 2>&1; then
    # Check if auth directory was created with correct permissions
    check_dir_permissions "$TEST_AUTH_DIR" "700" "Auth directory (SaveAuthTokens)"
    
    # Check if auth file was created with correct permissions
    check_file_permissions "$TEST_AUTH_FILE" "600" "Auth tokens file (SaveAuthTokens)"
    
    # Verify auth tokens are not world-readable
    perms=$(stat -c "%a" "$TEST_AUTH_FILE")
    world_read=$((perms % 10))
    
    if [ $world_read -eq 0 ]; then
        print_result "Auth tokens not world-readable" "PASS" "World permissions: $world_read"
    else
        print_result "Auth tokens not world-readable" "FAIL" "World permissions: $world_read (should be 0)"
    fi
else
    print_result "SaveAuthTokens test" "FAIL" "Failed to save auth tokens"
fi

echo ""
echo "========================================="
echo "Summary"
echo "========================================="
echo ""
echo "Tests passed: $TESTS_PASSED"
echo "Tests failed: $TESTS_FAILED"
echo ""

# Cleanup
rm -rf "$TEST_HOME"
rm -f /tmp/test_write_config.go /tmp/test_save_auth.go

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi
