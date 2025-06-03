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
  test              Test the runner environment
  shell             Start interactive shell

Environment Variables (required for registration):
  GITHUB_TOKEN      GitHub personal access token with repo scope
  GITHUB_REPOSITORY Repository in format 'owner/repo'
  RUNNER_NAME       Name for this runner (default: onemount-docker-runner)
  RUNNER_LABELS     Comma-separated labels (default: self-hosted,linux,onemount-testing)

Environment Variables (optional):
  AUTH_TOKENS_B64   Base64-encoded OneDrive auth tokens
  RUNNER_GROUP      Runner group (default: Default)

Examples:
  # Register and start runner
  docker run -e GITHUB_TOKEN=ghp_xxx -e GITHUB_REPOSITORY=owner/repo onemount-runner run

  # Setup with auth tokens
  docker run -e AUTH_TOKENS_B64=\$(base64 -w 0 ~/.cache/onemount/auth_tokens.json) \\
             -e GITHUB_TOKEN=ghp_xxx -e GITHUB_REPOSITORY=owner/repo \\
             onemount-runner run

  # Interactive shell for debugging
  docker run -it onemount-runner shell

EOF
}

# Function to setup OneDrive authentication
setup_auth() {
    print_info "Setting up OneDrive authentication..."
    
    if [[ -n "$AUTH_TOKENS_B64" ]]; then
        print_info "Using provided auth tokens..."
        echo "$AUTH_TOKENS_B64" | base64 -d > /opt/onemount-ci/.auth_tokens.json
        chmod 600 /opt/onemount-ci/.auth_tokens.json
        
        # Verify the tokens file is valid JSON
        if jq empty /opt/onemount-ci/.auth_tokens.json 2>/dev/null; then
            print_success "Auth tokens configured successfully"
            
            # Check token expiration
            EXPIRES_AT=$(jq -r '.expires_at // 0' /opt/onemount-ci/.auth_tokens.json)
            CURRENT_TIME=$(date +%s)
            
            if [[ "$EXPIRES_AT" -le "$CURRENT_TIME" ]]; then
                print_warning "Auth tokens appear to be expired"
                print_warning "You may need to refresh your authentication"
            else
                print_success "Auth tokens are valid (expires in $((EXPIRES_AT - CURRENT_TIME)) seconds)"
            fi
        else
            print_error "Invalid auth tokens format"
            return 1
        fi
    else
        print_warning "No auth tokens provided via AUTH_TOKENS_B64"
        print_info "You can provide tokens by setting AUTH_TOKENS_B64 environment variable"
        print_info "Generate with: base64 -w 0 ~/.cache/onemount/auth_tokens.json"
    fi
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
    
    # Test auth tokens
    if [[ -f /opt/onemount-ci/.auth_tokens.json ]]; then
        if jq empty /opt/onemount-ci/.auth_tokens.json 2>/dev/null; then
            print_success "Auth tokens are valid JSON"
        else
            print_error "Auth tokens are invalid JSON"
        fi
    else
        print_warning "No auth tokens found at /opt/onemount-ci/.auth_tokens.json"
    fi
    
    print_success "Environment test completed"
}

# Main execution
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
    test)
        test_environment
        ;;
    shell)
        print_info "Starting interactive shell..."
        exec /bin/bash
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
