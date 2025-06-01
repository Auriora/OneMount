#!/bin/bash

# OneMount System Test Runner
# This script runs comprehensive system tests using a real OneDrive account

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
AUTH_TOKENS_PATH="$HOME/.onemount-tests/.auth_tokens.json"
TEST_LOG_PATH="$HOME/.onemount-tests/logs/system_tests.log"
TIMEOUT="30m"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."

    # Check if auth tokens exist
    if [[ ! -f "$AUTH_TOKENS_PATH" ]]; then
        print_error "Authentication tokens not found at $AUTH_TOKENS_PATH"

        # Provide different instructions based on environment
        if [[ -n "${CI:-}" || -n "${GITHUB_ACTIONS:-}" ]]; then
            print_error "In CI environment, ensure secrets are properly configured:"
            print_error "  - For service principal: AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID"
            print_error "  - For test account: ONEDRIVE_TEST_TOKENS"
            print_error "See docs/testing/ci-system-tests-setup.md for details"
        else
            print_error "Please run OneMount authentication first:"
            print_error "  make onemount"
            print_error "  ./build/onemount --auth-only"
            print_error "  mkdir -p ~/.onemount-tests"
            print_error "  cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json"
        fi
        exit 1
    fi
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we're in the OneMount project directory
    if [[ ! -f "go.mod" ]] || ! grep -q "github.com/auriora/onemount" go.mod; then
        print_error "This script must be run from the OneMount project root directory"
        exit 1
    fi
    
    # Create log directory if it doesn't exist
    mkdir -p "$(dirname "$TEST_LOG_PATH")"
    
    # Check for CI environment and provide additional info
    if [[ -n "${CI:-}" || -n "${GITHUB_ACTIONS:-}" ]]; then
        print_status "Running in CI environment"

        # Validate token file format
        if ! jq empty "$AUTH_TOKENS_PATH" 2>/dev/null; then
            print_error "Auth tokens file is not valid JSON"
            exit 1
        fi

        # Check token expiration
        EXPIRES_AT=$(jq -r '.expires_at // 0' "$AUTH_TOKENS_PATH")
        CURRENT_TIME=$(date +%s)

        if [[ "$EXPIRES_AT" -le "$CURRENT_TIME" ]]; then
            print_warning "Auth tokens appear to be expired (expires_at: $EXPIRES_AT, current: $CURRENT_TIME)"
            print_warning "Tests may fail due to expired tokens"
        else
            print_status "Auth tokens are valid (expires in $((EXPIRES_AT - CURRENT_TIME)) seconds)"
        fi
    fi

    print_success "Prerequisites check passed"
}

# Function to display usage
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help              Show this help message"
    echo "  -v, --verbose           Enable verbose output"
    echo "  -t, --timeout DURATION  Set test timeout (default: 30m)"
    echo "  --comprehensive         Run comprehensive system tests (default)"
    echo "  --performance           Run performance tests only"
    echo "  --reliability           Run reliability tests only"
    echo "  --integration           Run integration tests only"
    echo "  --stress                Run stress tests only"
    echo "  --all                   Run all test categories"
    echo ""
    echo "Examples:"
    echo "  $0                      # Run comprehensive system tests"
    echo "  $0 --performance        # Run performance tests only"
    echo "  $0 --all --verbose      # Run all tests with verbose output"
    echo "  $0 --timeout 60m        # Run tests with 60 minute timeout"
}

# Function to run specific test category
run_test_category() {
    local category="$1"
    local test_pattern="$2"
    
    print_status "Running $category tests..."
    
    # Build the go test command
    local cmd="go test -v -timeout $TIMEOUT"
    
    if [[ -n "$test_pattern" ]]; then
        cmd="$cmd -run $test_pattern"
    fi
    
    cmd="$cmd ./tests/system"
    
    # Run the tests
    if eval "$cmd"; then
        print_success "$category tests completed successfully"
        return 0
    else
        print_error "$category tests failed"
        return 1
    fi
}

# Function to run all tests
run_all_tests() {
    local failed_tests=()
    
    print_status "Running all system test categories..."
    
    # Comprehensive tests
    if ! run_test_category "Comprehensive" "TestSystemST_COMPREHENSIVE_01_AllOperations"; then
        failed_tests+=("Comprehensive")
    fi
    
    # Performance tests
    if ! run_test_category "Performance" "TestSystemST_PERFORMANCE_01_UploadDownloadSpeed"; then
        failed_tests+=("Performance")
    fi
    
    # Reliability tests
    if ! run_test_category "Reliability" "TestSystemST_RELIABILITY_01_ErrorRecovery"; then
        failed_tests+=("Reliability")
    fi
    
    # Integration tests
    if ! run_test_category "Integration" "TestSystemST_INTEGRATION_01_MountUnmount"; then
        failed_tests+=("Integration")
    fi
    
    # Stress tests
    if ! run_test_category "Stress" "TestSystemST_STRESS_01_HighLoad"; then
        failed_tests+=("Stress")
    fi
    
    # Report results
    if [[ ${#failed_tests[@]} -eq 0 ]]; then
        print_success "All system tests completed successfully!"
        return 0
    else
        print_error "The following test categories failed: ${failed_tests[*]}"
        return 1
    fi
}

# Parse command line arguments
VERBOSE=false
TEST_CATEGORY=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -t|--timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --comprehensive)
            TEST_CATEGORY="comprehensive"
            shift
            ;;
        --performance)
            TEST_CATEGORY="performance"
            shift
            ;;
        --reliability)
            TEST_CATEGORY="reliability"
            shift
            ;;
        --integration)
            TEST_CATEGORY="integration"
            shift
            ;;
        --stress)
            TEST_CATEGORY="stress"
            shift
            ;;
        --all)
            TEST_CATEGORY="all"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Set default test category if none specified
if [[ -z "$TEST_CATEGORY" ]]; then
    TEST_CATEGORY="comprehensive"
fi

# Main execution
main() {
    print_status "OneMount System Test Runner"
    print_status "Test category: $TEST_CATEGORY"
    print_status "Timeout: $TIMEOUT"
    print_status "Log file: $TEST_LOG_PATH"
    echo ""
    
    # Check prerequisites
    check_prerequisites
    
    # Set verbose mode if requested
    if [[ "$VERBOSE" == "true" ]]; then
        set -x
    fi
    
    # Run tests based on category
    case "$TEST_CATEGORY" in
        comprehensive)
            run_test_category "Comprehensive" "TestSystemST_COMPREHENSIVE_01_AllOperations"
            ;;
        performance)
            run_test_category "Performance" "TestSystemST_PERFORMANCE_01_UploadDownloadSpeed"
            ;;
        reliability)
            run_test_category "Reliability" "TestSystemST_RELIABILITY_01_ErrorRecovery"
            ;;
        integration)
            run_test_category "Integration" "TestSystemST_INTEGRATION_01_MountUnmount"
            ;;
        stress)
            run_test_category "Stress" "TestSystemST_STRESS_01_HighLoad"
            ;;
        all)
            run_all_tests
            ;;
        *)
            print_error "Unknown test category: $TEST_CATEGORY"
            exit 1
            ;;
    esac
    
    local exit_code=$?
    
    echo ""
    if [[ $exit_code -eq 0 ]]; then
        print_success "System tests completed successfully!"
        print_status "Check the log file for detailed output: $TEST_LOG_PATH"
    else
        print_error "System tests failed!"
        print_status "Check the log file for error details: $TEST_LOG_PATH"
    fi
    
    exit $exit_code
}

# Run main function
main "$@"
