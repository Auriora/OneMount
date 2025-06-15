#!/bin/bash
# Test entrypoint script for OneMount Docker test container
# Provides a unified interface for running different types of tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
print_info() {
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

# Help function
show_help() {
    cat << EOF
OneMount Docker Test Runner

Usage: docker run [docker-options] onemount-test-runner [COMMAND] [OPTIONS]

Commands:
  help                    Show this help message
  unit                   Run unit tests only
  integration            Run integration tests
  system                 Run system tests (requires auth tokens)
  all                    Run all tests
  build                  Build OneMount binaries
  shell                  Start interactive shell
  coverage               Run tests with coverage analysis

Options:
  --verbose              Enable verbose output
  --timeout DURATION     Set test timeout (default: 5m)
  --sequential           Run tests sequentially (no parallel execution)

Examples:
  # Run unit tests
  docker run --rm onemount-test-runner unit

  # Run all tests with verbose output
  docker run --rm onemount-test-runner all --verbose

  # Run system tests with custom timeout
  docker run --rm onemount-test-runner system --timeout 30m

  # Start interactive shell for debugging
  docker run --rm -it onemount-test-runner shell

Environment Variables:
  ONEMOUNT_TEST_TIMEOUT   Test timeout duration (default: 5m)
  ONEMOUNT_TEST_VERBOSE   Enable verbose output (true/false)
  ONEMOUNT_AUTH_TOKENS    Path to OneDrive auth tokens for system tests

Notes:
  - System tests require OneDrive authentication tokens
  - Mount the project directory to /workspace
  - For system tests, copy auth tokens to workspace root as 'auth_tokens.json'
EOF
}

# Setup function
setup_environment() {
    print_info "Setting up test environment..."

    # Create home directory if running as different user
    if [[ ! -d "$HOME" ]]; then
        print_info "Creating home directory for user $(whoami)..."
        # Try to create in HOME first, fallback to /tmp if permission denied
        if ! mkdir -p "$HOME" 2>/dev/null; then
            print_warning "Cannot create $HOME, using /tmp/home-$(whoami) instead"
            export HOME="/tmp/home-$(whoami)"
            mkdir -p "$HOME"
        fi
        mkdir -p "$HOME/.cache/go-build"
        mkdir -p "$HOME/go"
    fi

    # Ensure we're in the workspace
    cd /workspace

    # Check if this looks like the OneMount project
    if [[ ! -f "go.mod" ]] || ! grep -q "github.com/auriora/onemount" go.mod; then
        print_error "This doesn't appear to be the OneMount project directory"
        print_error "Please mount the OneMount source code to /workspace"
        exit 1
    fi

    # Set up Go environment for current user (respect existing environment variables)
    export GOPATH="${GOPATH:-$HOME/go}"
    export GOCACHE="${GOCACHE:-$HOME/.cache/go-build}"
    export PATH="/usr/local/go/bin:$GOPATH/bin:$PATH"

    # Download dependencies
    print_info "Downloading Go dependencies..."
    go mod download

    # Verify FUSE is available
    if [[ ! -e /dev/fuse ]]; then
        print_warning "FUSE device not available - some tests may fail"
        print_warning "Run with --device /dev/fuse --cap-add SYS_ADMIN for full FUSE support"
    else
        print_info "FUSE device available for filesystem testing"
    fi

    # Setup auth tokens if available
    setup_auth_tokens

    print_success "Environment setup complete"
}

# Setup auth tokens function
setup_auth_tokens() {
    # Check multiple possible locations for auth tokens
    local auth_tokens_file=""

    # Check if auth tokens exist in the workspace (mounted from host)
    if [[ -f "/workspace/auth_tokens.json" ]]; then
        auth_tokens_file="/workspace/auth_tokens.json"
        print_info "Auth tokens found in workspace root"
    elif [[ -f "/workspace/test-artifacts/auth_tokens.json" ]]; then
        auth_tokens_file="/workspace/test-artifacts/auth_tokens.json"
        print_info "Auth tokens found in test-artifacts directory"
    elif [[ -f "/workspace/test-artifacts/.auth_tokens.json" ]]; then
        auth_tokens_file="/workspace/test-artifacts/.auth_tokens.json"
        print_info "Auth tokens found in test-artifacts directory (hidden file)"
    elif [[ -f "$HOME/.onemount-tests/.auth_tokens.json" ]]; then
        auth_tokens_file="$HOME/.onemount-tests/.auth_tokens.json"
        print_info "Auth tokens found in mounted directory"
    fi

    if [[ -n "$auth_tokens_file" ]]; then
        # Ensure the target directory exists
        mkdir -p "$HOME/.onemount-tests"

        # Copy auth tokens to the expected location if not already there
        if [[ "$auth_tokens_file" != "$HOME/.onemount-tests/.auth_tokens.json" ]]; then
            # Ensure target is a file, not a directory
            if [[ -d "$HOME/.onemount-tests/.auth_tokens.json" ]]; then
                print_warning "Removing incorrect directory at auth tokens location"
                rm -rf "$HOME/.onemount-tests/.auth_tokens.json"
            fi

            cp "$auth_tokens_file" "$HOME/.onemount-tests/.auth_tokens.json"
            chmod 600 "$HOME/.onemount-tests/.auth_tokens.json"
            print_info "Copied auth tokens to expected location"
        fi

        # Verify the tokens file is valid JSON
        if command -v jq >/dev/null 2>&1; then
            if jq empty "$HOME/.onemount-tests/.auth_tokens.json" 2>/dev/null; then
                print_success "Auth tokens are valid JSON"

                # Check token expiration if jq is available
                EXPIRES_AT=$(jq -r '.expires_at // 0' "$HOME/.onemount-tests/.auth_tokens.json" 2>/dev/null || echo "0")
                CURRENT_TIME=$(date +%s)

                if [[ "$EXPIRES_AT" != "0" ]] && [[ "$EXPIRES_AT" -le "$CURRENT_TIME" ]]; then
                    print_warning "Auth tokens appear to be expired"
                    print_warning "You may need to refresh your authentication"
                else
                    print_success "Auth tokens are valid"
                fi
            else
                print_error "Invalid auth tokens format"
                return 1
            fi
        else
            print_success "Auth tokens found (JSON validation skipped - jq not available)"
        fi
    elif [[ -n "$ONEMOUNT_AUTH_TOKENS" ]]; then
        print_info "Setting up auth tokens from environment variable..."

        # Ensure the target directory exists
        mkdir -p "$HOME/.onemount-tests"

        # Write auth tokens from environment variable
        echo "$ONEMOUNT_AUTH_TOKENS" > "$HOME/.onemount-tests/.auth_tokens.json"
        chmod 600 "$HOME/.onemount-tests/.auth_tokens.json"

        print_success "Auth tokens configured from environment"
    else
        print_info "No auth tokens found - system tests will be skipped"
        print_info "To enable system tests:"
        print_info "  - Copy auth tokens to workspace root as 'auth_tokens.json', or"
        print_info "  - Copy auth tokens to test-artifacts/auth_tokens.json, or"
        print_info "  - Set ONEMOUNT_AUTH_TOKENS environment variable"
        print_info ""
        print_info "SECURITY NOTE: Use dedicated test OneDrive account, NOT production!"
        print_info "Production tokens should NEVER be used for testing."
    fi
}

# Build function
build_onemount() {
    print_info "Building OneMount binaries..."
    
    # Use the project's build script for CGO compatibility
    if [[ -f "scripts/cgo-helper.sh" ]]; then
        bash scripts/cgo-helper.sh
    fi
    
    # Build main binary
    mkdir -p build
    CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
        -o build/onemount \
        -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
        ./cmd/onemount
    
    # Build launcher (if GUI dependencies are available)
    if pkg-config --exists webkit2gtk-4.1; then
        CGO_CFLAGS=-Wno-deprecated-declarations go build -v \
            -o build/onemount-launcher \
            -ldflags="-X github.com/auriora/onemount/cmd/common.commit=$(git rev-parse HEAD 2>/dev/null || echo 'unknown')" \
            ./cmd/onemount-launcher
        print_success "Built onemount and onemount-launcher"
    else
        print_warning "GUI dependencies not available, skipping launcher build"
        print_success "Built onemount"
    fi
}

# Test functions
run_unit_tests() {
    print_info "Running unit tests..."
    
    local cmd="go test -v ./... -short"
    
    if [[ "$SEQUENTIAL" == "true" ]]; then
        cmd="$cmd -p 1 -parallel 1"
    fi
    
    if [[ "$VERBOSE" == "true" ]]; then
        cmd="$cmd -v"
    fi
    
    cmd="$cmd -timeout $TIMEOUT"
    
    print_info "Executing: $cmd"
    if eval "$cmd"; then
        print_success "Unit tests passed"
        return 0
    else
        print_error "Unit tests failed"
        return 1
    fi
}

run_integration_tests() {
    print_info "Running integration tests..."
    
    local cmd="go test -v ./pkg/testutil/integration_test_env_test.go -timeout $TIMEOUT"
    
    print_info "Executing: $cmd"
    if eval "$cmd"; then
        print_success "Integration tests passed"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

run_system_tests() {
    print_info "Running system tests..."
    
    # Check for auth tokens
    if [[ ! -f "$HOME/.onemount-tests/.auth_tokens.json" ]]; then
        print_error "OneDrive auth tokens not found"
        print_error "Mount auth tokens to $HOME/.onemount-tests/.auth_tokens.json"
        print_error "Or set ONEMOUNT_AUTH_TOKENS environment variable"
        return 1
    fi
    
    local cmd="go test -v -timeout $TIMEOUT ./tests/system -run 'TestSystemST_.*'"
    
    print_info "Executing: $cmd"
    if eval "$cmd"; then
        print_success "System tests passed"
        return 0
    else
        print_error "System tests failed"
        return 1
    fi
}

run_all_tests() {
    print_info "Running all tests..."
    
    local failed_tests=()
    
    # Run unit tests
    if ! run_unit_tests; then
        failed_tests+=("unit")
    fi
    
    # Run integration tests
    if ! run_integration_tests; then
        failed_tests+=("integration")
    fi
    
    # Run system tests (if auth tokens available)
    if [[ -f "$HOME/.onemount-tests/.auth_tokens.json" ]]; then
        if ! run_system_tests; then
            failed_tests+=("system")
        fi
    else
        print_warning "Skipping system tests - no auth tokens found"
    fi
    
    # Report results
    if [[ ${#failed_tests[@]} -eq 0 ]]; then
        print_success "All tests passed!"
        return 0
    else
        print_error "Failed test categories: ${failed_tests[*]}"
        return 1
    fi
}

run_coverage() {
    print_info "Running tests with coverage analysis..."
    
    mkdir -p coverage
    
    local cmd="go test -v -coverprofile=coverage/coverage.out ./..."
    
    if [[ "$SEQUENTIAL" == "true" ]]; then
        cmd="$cmd -p 1 -parallel 1"
    fi
    
    cmd="$cmd -timeout $TIMEOUT"
    
    print_info "Executing: $cmd"
    if eval "$cmd"; then
        # Generate HTML coverage report
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        go tool cover -func=coverage/coverage.out
        print_success "Coverage analysis complete"
        print_info "Coverage report saved to coverage/coverage.html"
        return 0
    else
        print_error "Coverage analysis failed"
        return 1
    fi
}

# Parse command line arguments
COMMAND="${1:-help}"
shift || true

# Default values
TIMEOUT="${ONEMOUNT_TEST_TIMEOUT:-5m}"
VERBOSE="${ONEMOUNT_TEST_VERBOSE:-false}"
SEQUENTIAL="false"

# Parse options
while [[ $# -gt 0 ]]; do
    case $1 in
        --verbose)
            VERBOSE="true"
            shift
            ;;
        --timeout)
            TIMEOUT="$2"
            shift 2
            ;;
        --sequential)
            SEQUENTIAL="true"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
case "$COMMAND" in
    help|--help|-h)
        show_help
        ;;
    build)
        setup_environment
        build_onemount
        ;;
    unit)
        setup_environment
        build_onemount
        run_unit_tests
        ;;
    integration)
        setup_environment
        build_onemount
        run_integration_tests
        ;;
    system)
        setup_environment
        build_onemount
        run_system_tests
        ;;
    all)
        setup_environment
        build_onemount
        run_all_tests
        ;;
    coverage)
        setup_environment
        build_onemount
        run_coverage
        ;;
    shell)
        setup_environment
        print_info "Starting interactive shell..."
        exec /bin/bash
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_help
        exit 1
        ;;
esac
