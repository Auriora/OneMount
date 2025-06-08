#!/bin/bash

# OneMount Elastic Environment Setup Script
# Creates the main .env file for elastic runner system

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if .env already exists
check_existing_env() {
    if [[ -f "$PROJECT_ROOT/.env" ]]; then
        print_warning "Main .env file already exists"
        echo "Current content:"
        echo "----------------------------------------"
        cat "$PROJECT_ROOT/.env"
        echo "----------------------------------------"
        
        read -p "Do you want to recreate it? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Keeping existing .env file"
            return 1
        fi
    fi
    return 0
}

# Try to extract configuration from existing runner configs
extract_from_runners() {
    local github_token=""
    local github_repository=""
    local auth_tokens_b64=""
    
    # Check if we can find existing configuration
    for runner_dir in "$PROJECT_ROOT/.runners"/*; do
        if [[ -d "$runner_dir" && -f "$runner_dir/.env" ]]; then
            print_info "Found existing runner config in $(basename "$runner_dir")"
            
            # Try to extract values (safely)
            if [[ -r "$runner_dir/.env" ]]; then
                source "$runner_dir/.env" 2>/dev/null || true
                if [[ -n "${GITHUB_TOKEN:-}" ]]; then
                    github_token="$GITHUB_TOKEN"
                fi
                if [[ -n "${GITHUB_REPOSITORY:-}" ]]; then
                    github_repository="$GITHUB_REPOSITORY"
                fi
                if [[ -n "${AUTH_TOKENS_B64:-}" ]]; then
                    auth_tokens_b64="$AUTH_TOKENS_B64"
                fi
                break
            fi
        fi
    done
    
    echo "$github_token|$github_repository|$auth_tokens_b64"
}

# Interactive setup
interactive_setup() {
    print_info "Setting up elastic runner environment..."
    
    # Try to extract from existing configs
    local extracted_config
    extracted_config=$(extract_from_runners)
    IFS='|' read -r existing_token existing_repo existing_auth <<< "$extracted_config"
    
    # GitHub Token
    local github_token="$existing_token"
    if [[ -z "$github_token" ]]; then
        echo
        print_info "GitHub Personal Access Token is required"
        echo "The token needs 'repo' scope for private repositories"
        echo "Create one at: https://github.com/settings/tokens"
        echo
        read -p "Enter GitHub Token: " -s github_token
        echo
    else
        print_success "Using existing GitHub token (${github_token:0:8}...)"
    fi
    
    # GitHub Repository
    local github_repository="${existing_repo:-Auriora/OneMount}"
    echo
    print_info "GitHub Repository (default: $github_repository)"
    read -p "Repository [${github_repository}]: " repo_input
    if [[ -n "$repo_input" ]]; then
        github_repository="$repo_input"
    fi
    
    # OneDrive Auth Tokens (optional)
    local auth_tokens_b64="$existing_auth"
    if [[ -z "$auth_tokens_b64" ]]; then
        echo
        print_info "OneDrive authentication tokens (optional)"
        echo "If you have auth_tokens.json, we can encode it for you"
        read -p "Path to auth_tokens.json (or press Enter to skip): " auth_file
        
        if [[ -n "$auth_file" && -f "$auth_file" ]]; then
            auth_tokens_b64=$(base64 -w 0 "$auth_file")
            print_success "Auth tokens encoded successfully"
        fi
    else
        print_success "Using existing auth tokens"
    fi
    
    # Docker Host
    local docker_host="${DOCKER_HOST:-172.16.1.104:2376}"
    echo
    print_info "Docker Host (default: $docker_host)"
    read -p "Docker Host [${docker_host}]: " host_input
    if [[ -n "$host_input" ]]; then
        docker_host="$host_input"
    fi
    
    # Scaling Configuration
    echo
    print_info "Scaling Configuration"
    read -p "Minimum runners [1]: " min_runners
    min_runners="${min_runners:-1}"
    
    read -p "Maximum runners [5]: " max_runners
    max_runners="${max_runners:-5}"
    
    read -p "Scale up threshold (queued jobs) [2]: " scale_up_threshold
    scale_up_threshold="${scale_up_threshold:-2}"
    
    # Create .env file
    cat > "$PROJECT_ROOT/.env" << EOF
# OneMount Elastic GitHub Actions Runner Configuration
# Generated on $(date)

# GitHub Configuration
GITHUB_TOKEN=$github_token
GITHUB_REPOSITORY=$github_repository

# Docker Configuration
DOCKER_HOST=tcp://$docker_host

# Scaling Configuration
MIN_RUNNERS=$min_runners
MAX_RUNNERS=$max_runners
SCALE_UP_THRESHOLD=$scale_up_threshold
SCALE_DOWN_THRESHOLD=0
CHECK_INTERVAL=30
COOLDOWN_PERIOD=300

# OneDrive Authentication (optional)
AUTH_TOKENS_B64=$auth_tokens_b64

# Runner Configuration
RUNNER_LABELS=self-hosted,Linux,onemount-testing,optimized,elastic
RUNNER_GROUP=Default

EOF
    
    chmod 600 "$PROJECT_ROOT/.env"
    print_success "Environment configuration created: $PROJECT_ROOT/.env"
}

# Validate configuration
validate_config() {
    print_info "Validating configuration..."
    
    if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
        print_error "No .env file found"
        return 1
    fi
    
    source "$PROJECT_ROOT/.env"
    
    # Check required variables
    local errors=0
    
    if [[ -z "${GITHUB_TOKEN:-}" ]]; then
        print_error "GITHUB_TOKEN is required"
        ((errors++))
    fi
    
    if [[ -z "${GITHUB_REPOSITORY:-}" ]]; then
        print_error "GITHUB_REPOSITORY is required"
        ((errors++))
    fi
    
    if [[ -z "${DOCKER_HOST:-}" ]]; then
        print_error "DOCKER_HOST is required"
        ((errors++))
    fi
    
    if [[ $errors -gt 0 ]]; then
        print_error "Configuration validation failed"
        return 1
    fi
    
    print_success "Configuration validation passed"
    
    # Show summary
    echo
    print_info "Configuration Summary:"
    echo "Repository: $GITHUB_REPOSITORY"
    echo "Docker Host: $DOCKER_HOST"
    echo "Min/Max Runners: $MIN_RUNNERS/$MAX_RUNNERS"
    echo "Scale Up Threshold: $SCALE_UP_THRESHOLD queued jobs"
    echo "Auth Tokens: $(if [[ -n "${AUTH_TOKENS_B64:-}" ]]; then echo "Configured"; else echo "Not configured"; fi)"
}

# Show usage
show_usage() {
    cat << EOF
OneMount Elastic Environment Setup

Usage: $0 [COMMAND]

Commands:
  setup             Interactive environment setup
  validate          Validate existing configuration
  show              Show current configuration
  help              Show this help

Examples:
  $0 setup                          # Interactive setup
  $0 validate                       # Check configuration

EOF
}

# Show current configuration
show_config() {
    if [[ ! -f "$PROJECT_ROOT/.env" ]]; then
        print_error "No .env file found. Run '$0 setup' first."
        return 1
    fi
    
    print_info "Current configuration:"
    echo "----------------------------------------"
    # Show config but mask sensitive data
    sed 's/GITHUB_TOKEN=.*/GITHUB_TOKEN=***MASKED***/' "$PROJECT_ROOT/.env" | \
    sed 's/AUTH_TOKENS_B64=.*/AUTH_TOKENS_B64=***MASKED***/'
    echo "----------------------------------------"
}

# Main execution
main() {
    local command="${1:-help}"
    
    case "$command" in
        setup)
            if check_existing_env; then
                interactive_setup
                validate_config
            fi
            ;;
        validate)
            validate_config
            ;;
        show)
            show_config
            ;;
        help|--help|-h)
            show_usage
            ;;
        *)
            print_error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

main "$@"
