#!/bin/bash
# OneMount GitHub Actions Self-Hosted Runner Management Script
# Provides easy setup and management of the containerized runner

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
OneMount GitHub Actions Self-Hosted Runner Management

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  setup             Interactive setup of runner configuration
  build             Build the runner Docker image
  start             Start the runner container
  stop              Stop the runner container
  restart           Restart the runner container
  logs              Show runner logs
  shell             Start interactive shell in runner container
  test              Test the runner environment
  clean             Clean up runner containers and volumes
  status            Show runner status

Options:
  --dev             Use development configuration
  --verbose         Enable verbose output
  --follow          Follow logs (for logs command)

Environment Setup:
  Create a .env file with:
    GITHUB_TOKEN=ghp_your_token_here
    GITHUB_REPOSITORY=owner/repo
    AUTH_TOKENS_B64=base64_encoded_tokens

Examples:
  # Interactive setup
  $0 setup

  # Build and start runner
  $0 build
  $0 start

  # View logs
  $0 logs --follow

  # Development shell
  $0 shell --dev

  # Clean up everything
  $0 clean

EOF
}

# Function for interactive setup
interactive_setup() {
    print_info "OneMount GitHub Actions Runner Setup"
    echo ""
    
    # Check if .env exists
    if [[ -f .env ]]; then
        print_warning ".env file already exists"
        read -p "Do you want to overwrite it? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Using existing .env file"
            return 0
        fi
    fi
    
    # Get GitHub token
    echo "Step 1: GitHub Personal Access Token"
    echo "You need a GitHub personal access token with 'repo' scope."
    echo "Create one at: https://github.com/settings/tokens"
    echo ""
    read -p "Enter your GitHub token: " -s GITHUB_TOKEN
    echo ""
    
    if [[ -z "$GITHUB_TOKEN" ]]; then
        print_error "GitHub token is required"
        return 1
    fi
    
    # Get repository
    echo ""
    echo "Step 2: Repository"
    read -p "Enter repository (owner/repo): " GITHUB_REPOSITORY
    
    if [[ -z "$GITHUB_REPOSITORY" ]]; then
        print_error "Repository is required"
        return 1
    fi
    
    # Get runner name
    echo ""
    echo "Step 3: Runner Configuration"
    read -p "Enter runner name (default: onemount-docker-runner): " RUNNER_NAME
    RUNNER_NAME=${RUNNER_NAME:-onemount-docker-runner}
    
    # Check for auth tokens
    echo ""
    echo "Step 4: OneDrive Authentication (Optional)"
    AUTH_TOKENS_B64=""
    
    if [[ -f ~/.cache/onemount/auth_tokens.json ]]; then
        print_success "Found existing OneMount auth tokens"
        read -p "Use existing auth tokens? (Y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            AUTH_TOKENS_B64=$(base64 -w 0 ~/.cache/onemount/auth_tokens.json)
            print_success "Auth tokens encoded"
        fi
    else
        print_warning "No existing auth tokens found"
        echo "You can set up authentication later or provide a path to auth_tokens.json"
        read -p "Path to auth_tokens.json (optional): " AUTH_PATH
        
        if [[ -n "$AUTH_PATH" ]] && [[ -f "$AUTH_PATH" ]]; then
            AUTH_TOKENS_B64=$(base64 -w 0 "$AUTH_PATH")
            print_success "Auth tokens encoded from $AUTH_PATH"
        fi
    fi
    
    # Create .env file
    cat > .env << EOF
# OneMount GitHub Actions Runner Configuration
GITHUB_TOKEN=$GITHUB_TOKEN
GITHUB_REPOSITORY=$GITHUB_REPOSITORY
RUNNER_NAME=$RUNNER_NAME
RUNNER_LABELS=self-hosted,linux,onemount-testing
RUNNER_GROUP=Default

# OneDrive Authentication (base64 encoded)
AUTH_TOKENS_B64=$AUTH_TOKENS_B64

EOF
    
    chmod 600 .env
    print_success ".env file created successfully"
    
    echo ""
    print_info "Setup completed! Next steps:"
    echo "1. Run: $0 build"
    echo "2. Run: $0 start"
    echo "3. Check logs: $0 logs --follow"
}

# Function to build the image
build_image() {
    print_info "Building OneMount GitHub Actions runner image..."
    
    if docker build -f packaging/docker/Dockerfile.github-runner -t onemount-github-runner .; then
        print_success "Runner image built successfully"
    else
        print_error "Failed to build runner image"
        return 1
    fi
}

# Function to start the runner
start_runner() {
    local compose_file="docker/compose/docker-compose.runner.yml"
    local service="github-runner"

    if [[ "$1" == "--dev" ]]; then
        service="runner-dev"
        print_info "Starting development runner..."
    else
        print_info "Starting GitHub Actions runner..."
    fi

    if [[ ! -f .env ]]; then
        print_error ".env file not found. Run '$0 setup' first."
        return 1
    fi

    docker-compose -f "$compose_file" up -d "$service"
    print_success "Runner started successfully"

    if [[ "$service" == "github-runner" ]]; then
        print_info "Runner is now waiting for jobs from GitHub Actions"
        print_info "View logs with: $0 logs --follow"
    fi
}

# Function to stop the runner
stop_runner() {
    print_info "Stopping GitHub Actions runner..."
    docker-compose -f docker/compose/docker-compose.runner.yml down
    print_success "Runner stopped"
}

# Function to show logs
show_logs() {
    local follow_flag=""
    if [[ "$1" == "--follow" ]]; then
        follow_flag="-f"
    fi

    docker-compose -f docker/compose/docker-compose.runner.yml logs $follow_flag github-runner
}

# Function to start shell
start_shell() {
    local service="github-runner"
    if [[ "$1" == "--dev" ]]; then
        service="runner-dev"
    fi

    print_info "Starting interactive shell..."
    docker-compose -f docker/compose/docker-compose.runner.yml run --rm "$service" shell
}

# Function to test environment
test_environment() {
    print_info "Testing runner environment..."
    docker-compose -f docker/compose/docker-compose.runner.yml run --rm github-runner test
}

# Function to show status
show_status() {
    print_info "Runner container status:"
    docker-compose -f docker/compose/docker-compose.runner.yml ps

    echo ""
    print_info "Runner logs (last 20 lines):"
    docker-compose -f docker/compose/docker-compose.runner.yml logs --tail=20 github-runner
}

# Function to clean up
clean_up() {
    print_warning "This will remove all runner containers and volumes"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Cleaning up runner resources..."
        docker-compose -f docker/compose/docker-compose.runner.yml down -v
        docker image rm onemount-github-runner 2>/dev/null || true
        print_success "Cleanup completed"
    else
        print_info "Cleanup cancelled"
    fi
}

# Parse arguments
DEV_MODE=false
VERBOSE=false
FOLLOW=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --dev)
            DEV_MODE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --follow)
            FOLLOW=true
            shift
            ;;
        *)
            COMMAND="$1"
            shift
            ;;
    esac
done

# Main execution
case "${COMMAND:-}" in
    setup)
        interactive_setup
        ;;
    build)
        build_image
        ;;
    start)
        if [[ "$DEV_MODE" == "true" ]]; then
            start_runner --dev
        else
            start_runner
        fi
        ;;
    stop)
        stop_runner
        ;;
    restart)
        stop_runner
        if [[ "$DEV_MODE" == "true" ]]; then
            start_runner --dev
        else
            start_runner
        fi
        ;;
    logs)
        if [[ "$FOLLOW" == "true" ]]; then
            show_logs --follow
        else
            show_logs
        fi
        ;;
    shell)
        if [[ "$DEV_MODE" == "true" ]]; then
            start_shell --dev
        else
            start_shell
        fi
        ;;
    test)
        test_environment
        ;;
    status)
        show_status
        ;;
    clean)
        clean_up
        ;;
    --help|-h|help)
        show_usage
        ;;
    *)
        if [[ -n "$COMMAND" ]]; then
            print_error "Unknown command: $COMMAND"
        fi
        show_usage
        exit 1
        ;;
esac
