#!/bin/bash
# GitHub Actions Self-Hosted Runner Entrypoint for OneMount
# Handles runner registration, credential setup, and execution

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Function to show usage
show_usage() {
    cat << EOF
OneMount GitHub Actions Self-Hosted Runner

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  register          Register the runner with GitHub
  start             Start the runner (after registration)
  run               Register and start the runner
  setup-auth        Setup OneDrive authentication
  refresh-tokens    Refresh OneDrive authentication tokens
  token-status      Show authentication token status
  test              Test the runner environment
  init-workspace    Initialize workspace from source code
  sync-workspace    Sync workspace with latest source code
  shell             Start interactive shell
  exec              Execute a script or command

Environment Variables (required for registration):
  GITHUB_TOKEN      GitHub personal access token with repo scope
  GITHUB_REPOSITORY Repository in format 'owner/repo'
  RUNNER_NAME       Name for this runner (default: onemount-docker-runner)
  RUNNER_LABELS     Comma-separated labels (default: self-hosted,linux,onemount-testing)

Environment Variables (optional):
  AUTH_TOKENS_B64                Base64-encoded OneDrive auth tokens
  RUNNER_GROUP                   Runner group (default: Default)
  ONEMOUNT_SYNC_WORKSPACE        Set to 'true' to sync workspace on startup
  ONEMOUNT_AUTO_REFRESH_TOKENS   Set to 'false' to disable auto token refresh (default: true)
  ONEMOUNT_TOKEN_REFRESH_INTERVAL Seconds between token refresh attempts (default: 3600)

Examples:
  # Register and start runner
  docker run -e GITHUB_TOKEN=ghp_xxx -e GITHUB_REPOSITORY=owner/repo onemount-runner run

  # Setup with auth tokens
  docker run -e AUTH_TOKENS_B64=\$(base64 -w 0 ~/.cache/onemount/auth_tokens.json) \\
             -e GITHUB_TOKEN=ghp_xxx -e GITHUB_REPOSITORY=owner/repo \\
             onemount-runner run

  # Interactive shell for debugging
  docker run -it onemount-runner shell

  # Execute a script
  docker run onemount-runner exec /workspace/scripts/manage-runners.sh status

EOF
}

# Function to setup OneDrive authentication
setup_auth() {
    print_info "Setting up OneDrive authentication..."

    # Use the token manager to ensure fresh tokens
    if /usr/local/bin/token-manager.sh ensure; then
        print_success "Authentication tokens are ready"

        # Show token status
        /usr/local/bin/token-manager.sh status

        # Set up periodic token refresh if enabled
        if [[ "${ONEMOUNT_AUTO_REFRESH_TOKENS:-true}" == "true" ]]; then
            setup_token_refresh_daemon
        fi
    else
        print_error "Failed to setup authentication tokens"
        print_info "Available options:"
        print_info "1. Provide AUTH_TOKENS_B64 environment variable"
        print_info "2. Mount existing tokens to /opt/onemount-ci/auth_tokens.json"
        print_info "3. Run manual authentication in the container"
        return 1
    fi
}

# Function to setup periodic token refresh daemon
setup_token_refresh_daemon() {
    local refresh_interval="${ONEMOUNT_TOKEN_REFRESH_INTERVAL:-3600}"  # Default: 1 hour

    print_info "Setting up token refresh daemon (interval: ${refresh_interval}s)"

    # Create a background process to refresh tokens periodically
    (
        while true; do
            sleep "$refresh_interval"
            print_info "Performing scheduled token refresh..."
            if /usr/local/bin/token-manager.sh ensure; then
                print_success "Scheduled token refresh completed"
            else
                print_warning "Scheduled token refresh failed"
            fi
        done
    ) &

    # Store the PID for potential cleanup
    echo $! > /tmp/token-refresh-daemon.pid
    print_success "Token refresh daemon started (PID: $!)"
}

# Function to register the runner
register_runner() {
    print_info "Registering GitHub Actions runner..."
    
    # Check required environment variables
    if [[ -z "$GITHUB_TOKEN" ]]; then
        print_error "GITHUB_TOKEN environment variable is required"
        return 1
    fi
    
    if [[ -z "$GITHUB_REPOSITORY" ]]; then
        print_error "GITHUB_REPOSITORY environment variable is required (format: owner/repo)"
        return 1
    fi
    
    # Set defaults
    RUNNER_NAME=${RUNNER_NAME:-"onemount-docker-runner"}
    RUNNER_LABELS=${RUNNER_LABELS:-"self-hosted,linux,onemount-testing"}
    RUNNER_GROUP=${RUNNER_GROUP:-"Default"}
    
    print_info "Repository: $GITHUB_REPOSITORY"
    print_info "Runner name: $RUNNER_NAME"
    print_info "Labels: $RUNNER_LABELS"
    
    # Get registration token
    print_info "Getting registration token..."
    REGISTRATION_TOKEN=$(curl -s -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/$GITHUB_REPOSITORY/actions/runners/registration-token" | \
        jq -r '.token')
    
    if [[ "$REGISTRATION_TOKEN" == "null" ]] || [[ -z "$REGISTRATION_TOKEN" ]]; then
        print_error "Failed to get registration token"
        print_error "Check your GITHUB_TOKEN and GITHUB_REPOSITORY"
        return 1
    fi
    
    # Register the runner
    cd /opt/actions-runner
    ./config.sh \
        --url "https://github.com/$GITHUB_REPOSITORY" \
        --token "$REGISTRATION_TOKEN" \
        --name "$RUNNER_NAME" \
        --labels "$RUNNER_LABELS" \
        --runnergroup "$RUNNER_GROUP" \
        --work "/workspace" \
        --replace \
        --unattended
    
    print_success "Runner registered successfully"
}

# Function to start the runner
start_runner() {
    print_info "Starting GitHub Actions runner..."
    cd /opt/actions-runner
    
    # Setup auth if provided
    setup_auth
    
    print_success "Runner is starting..."
    print_info "Runner will process jobs for repository: $GITHUB_REPOSITORY"
    print_info "Press Ctrl+C to stop the runner"
    
    # Start the runner
    ./run.sh
}

# Function to fix permissions and switch to runner user
fix_permissions_and_switch_user() {
    if [[ "$(id -u)" == "0" ]]; then
        print_info "Fixing permissions and switching to runner user..."

        # Initialize workspace if using Docker volumes
        /usr/local/bin/init-workspace.sh init

        # Fix ownership of the entire actions-runner directory
        chown -R runner:runner /opt/actions-runner

        # Create and fix ownership of auth directory
        mkdir -p /opt/onemount-ci
        chown -R runner:runner /opt/onemount-ci

        # Switch to runner user and re-execute the script with preserved environment
        print_info "Switching to runner user..."
        exec su runner -c "cd /opt/actions-runner && env \
            GITHUB_TOKEN='$GITHUB_TOKEN' \
            GITHUB_REPOSITORY='$GITHUB_REPOSITORY' \
            RUNNER_NAME='$RUNNER_NAME' \
            RUNNER_LABELS='$RUNNER_LABELS' \
            RUNNER_GROUP='$RUNNER_GROUP' \
            AUTH_TOKENS_B64='$AUTH_TOKENS_B64' \
            RUNNER_ALLOW_RUNASROOT='$RUNNER_ALLOW_RUNASROOT' \
            ONEMOUNT_SYNC_WORKSPACE='$ONEMOUNT_SYNC_WORKSPACE' \
            ONEMOUNT_AUTO_REFRESH_TOKENS='$ONEMOUNT_AUTO_REFRESH_TOKENS' \
            ONEMOUNT_TOKEN_REFRESH_INTERVAL='$ONEMOUNT_TOKEN_REFRESH_INTERVAL' \
            $0 $*"
    else
        print_info "Already running as runner user"
    fi
}

# Function to test the environment
test_environment() {
    print_info "Testing runner environment..."
    
    # Test Go
    if go version; then
        print_success "Go is working"
    else
        print_error "Go is not working"
    fi
    
    # Test FUSE
    if [[ -c /dev/fuse ]]; then
        print_success "FUSE device is available"
    else
        print_error "FUSE device is not available"
    fi
    
    # Test auth tokens using token manager
    print_info "Testing authentication tokens..."
    if /usr/local/bin/token-manager.sh validate; then
        print_success "Auth tokens are valid"
        /usr/local/bin/token-manager.sh status
    else
        print_error "Auth tokens are invalid or missing"
    fi
    
    print_success "Environment test completed"
}

# Main execution
# Fix permissions and switch to runner user for runner commands
case "${1:-}" in
    register|start|run)
        fix_permissions_and_switch_user "$@"
        ;;
esac

# Execute the actual command
case "${1:-}" in
    register)
        register_runner
        ;;
    start)
        start_runner
        ;;
    run)
        register_runner && start_runner
        ;;
    setup-auth)
        setup_auth
        ;;
    refresh-tokens)
        /usr/local/bin/token-manager.sh refresh
        ;;
    token-status)
        /usr/local/bin/token-manager.sh status
        ;;
    test)
        test_environment
        ;;
    init-workspace)
        /usr/local/bin/init-workspace.sh init
        ;;
    sync-workspace)
        /usr/local/bin/init-workspace.sh sync
        ;;
    shell)
        print_info "Starting interactive shell..."
        exec /bin/bash
        ;;
    exec)
        if [[ -z "$2" ]]; then
            print_error "exec command requires a script or command to execute"
            show_usage
            exit 1
        fi
        shift  # Remove 'exec' from arguments
        print_info "Executing: $*"
        exec "$@"
        ;;
    --help|-h|help)
        show_usage
        ;;
    *)
        if [[ -n "$1" ]]; then
            print_error "Unknown command: $1"
        fi
        show_usage
        exit 1
        ;;
esac
