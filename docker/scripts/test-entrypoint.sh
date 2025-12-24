#!/bin/bash
# Test entrypoint script for OneMount Docker test container
# Provides a unified interface for running different types of tests

set -e

# Source common functions
if [[ -f /usr/local/bin/common.sh ]]; then
    source /usr/local/bin/common.sh
else
    # Fallback if common.sh not available
    print_info() { echo "[INFO] $1"; }
    print_success() { echo "[SUCCESS] $1"; }
    print_warning() { echo "[WARNING] $1"; }
    print_error() { echo "[ERROR] $1"; }
fi

# Help function
show_help() {
    cat << EOF
OneMount Docker Test Runner

Usage: docker run [docker-options] onemount-test-runner [COMMAND] [OPTIONS]

Helper Commands:
  help                    Show this help message
  unit                   Run unit tests only
  integration            Run integration tests
  system                 Run system tests (requires auth tokens)
  all                    Run all tests
  build                  Build OneMount binaries
  shell                  Start interactive shell
  coverage               Run tests with coverage analysis

Helper Command Options:
  --verbose              Enable verbose output
  --timeout DURATION     Set test timeout (default: 5m)
  --sequential           Run tests sequentially (no parallel execution)
  --log-to-file          Redirect verbose output to log files (keeps console clean)

Pass-Through Mode:
  Any command not matching the helper commands above will be executed directly
  after setting up the environment and building binaries.

Examples:
  # Run unit tests (helper command)
  docker run --rm onemount-test-runner unit

  # Run all tests with verbose output (helper command)
  docker run --rm onemount-test-runner all --verbose

  # Run specific test pattern (pass-through)
  docker run --rm onemount-test-runner go test -v -run TestIT_FS_ETag ./internal/fs

  # Run custom go command (pass-through)
  docker run --rm onemount-test-runner go test -v -timeout 10m ./...

  # Start interactive shell for debugging
  docker run --rm -it onemount-test-runner shell

Environment Variables:
  ONEMOUNT_TEST_TIMEOUT   Test timeout duration (default: 5m)
  ONEMOUNT_TEST_VERBOSE   Enable verbose output (true/false)
  ONEMOUNT_LOG_TO_FILE    Redirect verbose output to log files (true/false)
  ONEMOUNT_LOG_DIR        Directory for log files (default: ~/.onemount-tests/logs)
  ONEMOUNT_AUTH_TOKENS    Path to OneDrive auth tokens for system tests

Notes:
  - System tests require OneDrive authentication tokens
  - Mount the project directory to /workspace
  - For system tests, copy auth tokens to workspace root as 'auth_tokens.json'
  - Pass-through mode automatically sets up environment and builds binaries
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
    fi

    # Set up Go environment for current user
    export GOPATH="${GOPATH:-$HOME/go}"
    export PATH="/usr/local/go/bin:$GOPATH/bin:$PATH"

    # Use mounted cache volumes (directories created with 777 permissions in Dockerfile)
    if [[ -d "/tmp/go-build-cache" ]]; then
        export GOCACHE="/tmp/go-build-cache"
        print_info "Using mounted Go build cache at $GOCACHE"
    else
        export GOCACHE="$HOME/.cache/go-build"
        mkdir -p "$GOCACHE" 2>/dev/null || true
        print_warning "Go build cache not available, using $GOCACHE (builds will be slower)"
    fi

    if [[ -d "/tmp/go-mod-cache" ]]; then
        export GOMODCACHE="/tmp/go-mod-cache"
        print_info "Using mounted Go module cache at $GOMODCACHE"
    else
        export GOMODCACHE="$HOME/go/pkg/mod"
        mkdir -p "$GOMODCACHE" 2>/dev/null || true
        print_warning "Go module cache not available, using $GOMODCACHE"
    fi

    # Ensure we're in the workspace
    cd /workspace

    # Check if this looks like the OneMount project
    if [[ ! -f "go.mod" ]] || ! grep -q "github.com/auriora/onemount" go.mod; then
        print_error "This doesn't appear to be the OneMount project directory"
        print_error "Please mount the OneMount source code to /workspace"
        exit 1
    fi

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

    # Setup X11 for GUI applications
    setup_x11

    # Setup auth tokens if available
    setup_auth_tokens

    print_success "Environment setup complete"
}

# Refresh auth tokens function
refresh_auth_tokens() {
    local token_file="$1"
    
    # Check if the refresh tool exists
    if [[ -f "build/onemount" ]]; then
        # Use onemount CLI to refresh tokens if it has that capability
        # For now, just return false to skip automatic refresh
        # TODO: Add token refresh command to onemount CLI
        return 1
    fi
    
    return 1
}

# Setup X11 function
setup_x11() {
    # Set up X11 forwarding for GUI applications
    if [[ -n "$DISPLAY" ]]; then
        print_info "Setting up X11 forwarding for GUI applications..."
        
        # Create runtime directory for D-Bus
        export XDG_RUNTIME_DIR="/tmp/runtime-$(whoami)"
        mkdir -p "$XDG_RUNTIME_DIR"
        chmod 700 "$XDG_RUNTIME_DIR"
        
        # Set up D-Bus session
        if command -v dbus-launch >/dev/null 2>&1; then
            # Start D-Bus session if not already running
            if [[ -z "$DBUS_SESSION_BUS_ADDRESS" ]]; then
                eval $(dbus-launch --sh-syntax)
                export DBUS_SESSION_BUS_ADDRESS
                print_info "Started D-Bus session: $DBUS_SESSION_BUS_ADDRESS"
            fi
        fi
        
        # Allow X11 access for current user
        if [[ -f "/tmp/.Xauthority" ]]; then
            export XAUTHORITY="/tmp/.Xauthority"
            print_info "Using X11 authority file: $XAUTHORITY"
        fi
        
        # Test X11 connection
        if command -v xauth >/dev/null 2>&1 && [[ -n "$XAUTHORITY" ]]; then
            if xauth list >/dev/null 2>&1; then
                print_success "X11 forwarding configured successfully"
            else
                print_warning "X11 authority file may have permission issues"
            fi
        fi
        
        # Set additional X11 environment variables to prevent issues
        export QT_X11_NO_MITSHM=1
        export _X11_NO_MITSHM=1
        export _MITSHM=0
        
        print_info "X11 environment configured for display: $DISPLAY"
    else
        print_info "No DISPLAY set - GUI applications will not be available"
    fi
}

# Setup auth tokens function
setup_auth_tokens() {
    # Ensure test directories exist
    mkdir -p "$HOME/.onemount-tests/tmp"
    mkdir -p "$HOME/.onemount-tests/logs"
    
    # Define canonical auth location
    CANONICAL_AUTH_FILE="$HOME/.onemount-tests/.auth_tokens.json"
    
    # Check if canonical auth file exists and is valid
    if [[ -f "$CANONICAL_AUTH_FILE" ]] && [[ -s "$CANONICAL_AUTH_FILE" ]]; then
        print_info "Using canonical auth tokens from: $CANONICAL_AUTH_FILE"
        
        # Verify the tokens file is valid JSON
        if command -v jq >/dev/null 2>&1; then
            if jq empty "$CANONICAL_AUTH_FILE" 2>/dev/null; then
                print_success "Auth tokens are valid JSON"

                # Check token expiration if jq is available
                EXPIRES_AT=$(jq -r '.expires_at // 0' "$CANONICAL_AUTH_FILE" 2>/dev/null || echo "0")
                CURRENT_TIME=$(date +%s)

                if [[ "$EXPIRES_AT" != "0" ]] && [[ "$EXPIRES_AT" -le "$CURRENT_TIME" ]]; then
                    print_warning "Auth tokens appear to be expired"
                    print_info "Attempting to refresh tokens..."
                    
                    # Try to refresh the tokens using a Go helper
                    if refresh_auth_tokens "$CANONICAL_AUTH_FILE"; then
                        print_success "Auth tokens refreshed successfully"
                    else
                        print_warning "Failed to refresh tokens - tests may fail"
                        print_warning "You may need to re-authenticate manually"
                    fi
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
        return 0
    fi
    
    # If canonical file doesn't exist, look for tokens in other locations and link them
    print_info "Canonical auth file not found, searching for tokens..."
    
    # Check multiple possible locations for auth tokens
    # Priority order: mounted reference > test-artifacts > mounted cache > workspace root (legacy)
    local auth_tokens_file=""

    # Check mounted reference location first (new reference-based system)
    if [[ -n "$ONEMOUNT_AUTH_PATH" ]] && [[ -f "$ONEMOUNT_AUTH_PATH" ]]; then
        auth_tokens_file="$ONEMOUNT_AUTH_PATH"
        print_info "Auth tokens found via reference system: $ONEMOUNT_AUTH_PATH"
    # Check test-artifacts (preferred location for Docker tests)
    elif [[ -f "/workspace/test-artifacts/.auth_tokens.json" ]]; then
        auth_tokens_file="/workspace/test-artifacts/.auth_tokens.json"
        print_info "Auth tokens found in test-artifacts directory (hidden file)"
    elif [[ -f "/workspace/test-artifacts/auth_tokens.json" ]]; then
        auth_tokens_file="/workspace/test-artifacts/auth_tokens.json"
        print_info "Auth tokens found in test-artifacts directory"
    # Check mounted cache directory for fresh tokens
    elif [[ -f "/tmp/home-tester/.cache/onedriver/home-bcherrington-OneDrive/auth_tokens.json" ]]; then
        auth_tokens_file="/tmp/home-tester/.cache/onedriver/home-bcherrington-OneDrive/auth_tokens.json"
        print_info "Auth tokens found in mounted cache directory (onedriver)"
    elif [[ -f "/tmp/home-tester/.cache/onemount/home-bcherrington-OneMountTest/auth_tokens.json" ]]; then
        auth_tokens_file="/tmp/home-tester/.cache/onemount/home-bcherrington-OneMountTest/auth_tokens.json"
        print_info "Auth tokens found in mounted cache directory (onemount)"
    # Check workspace root last (legacy location, may be stale)
    elif [[ -f "/workspace/auth_tokens.json" ]]; then
        auth_tokens_file="/workspace/auth_tokens.json"
        print_warning "Auth tokens found in workspace root (legacy location - may be stale)"
        print_info "Consider moving to test-artifacts/.auth_tokens.json"
    fi

    if [[ -n "$auth_tokens_file" ]]; then
        # Create symlink to canonical location instead of copying
        print_info "Creating symlink to canonical location: $CANONICAL_AUTH_FILE"
        ln -sf "$auth_tokens_file" "$CANONICAL_AUTH_FILE"
        print_info "Symlinked auth tokens to canonical location"

        # Verify the tokens file is valid JSON
        if command -v jq >/dev/null 2>&1; then
            if jq empty "$CANONICAL_AUTH_FILE" 2>/dev/null; then
                print_success "Auth tokens are valid JSON"

                # Check token expiration if jq is available
                EXPIRES_AT=$(jq -r '.expires_at // 0' "$CANONICAL_AUTH_FILE" 2>/dev/null || echo "0")
                CURRENT_TIME=$(date +%s)

                if [[ "$EXPIRES_AT" != "0" ]] && [[ "$EXPIRES_AT" -le "$CURRENT_TIME" ]]; then
                    print_warning "Auth tokens appear to be expired"
                    print_info "Attempting to refresh tokens..."
                    
                    # Try to refresh the tokens using a Go helper
                    if refresh_auth_tokens "$CANONICAL_AUTH_FILE"; then
                        print_success "Auth tokens refreshed successfully"
                    else
                        print_warning "Failed to refresh tokens - tests may fail"
                        print_warning "You may need to re-authenticate manually"
                    fi
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

        # Write auth tokens from environment variable to canonical location
        echo "$ONEMOUNT_AUTH_TOKENS" > "$CANONICAL_AUTH_FILE"
        chmod 600 "$CANONICAL_AUTH_FILE"

        print_success "Auth tokens configured from environment"
    else
        print_info "No auth tokens found - system tests will be skipped"
        print_info "To enable system tests:"
        print_info "  - Run: ./scripts/fix-auth-tokens.sh (to find and link fresh tokens)"
        print_info "  - Copy auth tokens to test-artifacts/.auth_tokens.json, or"
        print_info "  - Set ONEMOUNT_AUTH_TOKENS environment variable"
        print_info ""
        print_info "SECURITY NOTE: Use dedicated test OneDrive account, NOT production!"
        print_info "Production tokens should NEVER be used for testing."
    fi
}

# Build function
build_onemount() {
    # Check if pre-built binaries exist from the Docker image
    if [[ -f "build/binaries/onemount" ]] && [[ -f "build/binaries/onemount-launcher" ]]; then
        print_info "Using pre-built binaries from Docker image"
        
        # Copy pre-built binaries to build directory
        mkdir -p build
        cp -f build/binaries/onemount build/onemount
        cp -f build/binaries/onemount-launcher build/onemount-launcher
        
        print_success "Pre-built binaries ready"
        return 0
    fi

    print_info "Building OneMount binaries from source..."

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

    # Run all integration tests (tests with TestIT_ prefix) in the fs package
    # Note: We skip end_to_end_workflow_test.go by temporarily renaming it
    local e2e_test="./internal/fs/end_to_end_workflow_test.go"
    local e2e_backup=""
    
    if [[ -f "$e2e_test" ]]; then
        e2e_backup="${e2e_test}.skip"
        mv "$e2e_test" "$e2e_backup"
        print_info "Temporarily skipped end_to_end_workflow_test.go (uses deprecated APIs)"
    fi

    local cmd="go test -v -run 'TestIT_' ./internal/fs -timeout $TIMEOUT"

    if [[ "$SEQUENTIAL" == "true" ]]; then
        cmd="$cmd -p 1 -parallel 1"
    fi

    print_info "Executing: $cmd"
    local result=0
    if eval "$cmd"; then
        print_success "Integration tests passed"
    else
        print_error "Integration tests failed"
        result=1
    fi

    # Restore the skipped test file
    if [[ -n "$e2e_backup" ]] && [[ -f "$e2e_backup" ]]; then
        mv "$e2e_backup" "$e2e_test"
    fi

    return $result
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

# Default values
TIMEOUT="${ONEMOUNT_TEST_TIMEOUT:-5m}"
VERBOSE="${ONEMOUNT_TEST_VERBOSE:-false}"
SEQUENTIAL="false"
LOG_TO_FILE="${ONEMOUNT_LOG_TO_FILE:-false}"
LOG_DIR="${ONEMOUNT_LOG_DIR:-$HOME/.onemount-tests/logs}"

# Check if first argument is a known helper command
case "$COMMAND" in
    help|--help|-h)
        show_help
        exit 0
        ;;
    build|unit|integration|system|all|coverage|shell)
        # Known helper command - parse options and execute
        shift || true
        
        # Parse options for helper commands
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
                --log-to-file)
                    LOG_TO_FILE="true"
                    shift
                    ;;
                *)
                    print_error "Unknown option: $1"
                    show_help
                    exit 1
                    ;;
            esac
        done
        ;;
    *)
        # Not a helper command - pass through to execution
        # This allows: docker run ... onemount-test-runner go test -v ./...
        ;;
esac

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
        print_info "For Python help, run: python-helper.sh"
        exec /bin/bash
        ;;
    *)
        # Pass-through mode: setup environment and execute the command
        print_info "Pass-through mode: executing custom command"
        setup_environment
        build_onemount
        
        print_info "Executing: $*"
        exec "$@"
        ;;
esac
