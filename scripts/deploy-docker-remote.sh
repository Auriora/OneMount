#!/bin/bash
# Deploy OneMount GitHub Actions Runner to Remote Docker Host via Docker API
# Connects directly to Docker daemon over TCP port 2375

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

# Default configuration
DOCKER_HOST="172.16.1.104:2375"
RUNNER_NAME="onemount-runner-remote"
CONTAINER_NAME="onemount-github-runner"
IMAGE_NAME="onemount-github-runner"

# Function to show usage
show_usage() {
    cat << EOF
OneMount Remote Docker GitHub Actions Runner

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  setup             Interactive setup and configuration
  build             Build the runner image on remote Docker
  deploy            Deploy and start the runner container
  start             Start the runner container
  stop              Stop the runner container
  restart           Restart the runner container
  status            Check runner and container status
  logs              View runner logs
  shell             Connect to runner shell
  update            Update runner image and restart
  clean             Clean up runner container and image

Options:
  --host HOST       Docker host (default: $DOCKER_HOST)
  --name NAME       Runner name (default: $RUNNER_NAME)

Environment Variables:
  GITHUB_TOKEN      GitHub personal access token (required)
  GITHUB_REPOSITORY Repository in format 'owner/repo' (required)
  AUTH_TOKENS_B64   Base64-encoded OneDrive auth tokens (optional)

Examples:
  # Interactive setup
  $0 setup

  # Build and deploy
  $0 build
  $0 deploy

  # Check status and logs
  $0 status
  $0 logs

  # Update runner
  $0 update

EOF
}

# Function to test Docker connection
test_docker_connection() {
    print_info "Testing connection to Docker host $DOCKER_HOST..."
    
    if DOCKER_HOST="tcp://$DOCKER_HOST" docker version >/dev/null 2>&1; then
        print_success "Docker connection established"
        
        # Show Docker info
        print_info "Remote Docker info:"
        DOCKER_HOST="tcp://$DOCKER_HOST" docker version --format "{{.Server.Version}}" | sed 's/^/  Docker version: /'
        DOCKER_HOST="tcp://$DOCKER_HOST" docker info --format "{{.OperatingSystem}}" | sed 's/^/  OS: /'
        return 0
    else
        print_error "Failed to connect to Docker host $DOCKER_HOST"
        print_info "Please ensure:"
        print_info "1. Docker daemon is running on $DOCKER_HOST"
        print_info "2. Port 2375 is accessible"
        print_info "3. Docker daemon is configured to accept TCP connections"
        return 1
    fi
}

# Function for interactive setup
interactive_setup() {
    print_info "OneMount Remote Docker GitHub Actions Runner Setup"
    echo ""
    
    # Get Docker host
    read -p "Docker host (default: $DOCKER_HOST): " input_host
    DOCKER_HOST=${input_host:-$DOCKER_HOST}
    
    # Test Docker connection
    if ! test_docker_connection; then
        return 1
    fi
    
    # Get GitHub configuration
    echo ""
    print_info "GitHub Configuration"
    
    if [[ -z "$GITHUB_TOKEN" ]]; then
        echo "You need a GitHub personal access token with 'repo' scope."
        echo "Create one at: https://github.com/settings/tokens"
        echo ""
        read -p "Enter your GitHub token: " -s GITHUB_TOKEN
        echo ""
    fi
    
    if [[ -z "$GITHUB_REPOSITORY" ]]; then
        read -p "Enter repository (owner/repo): " GITHUB_REPOSITORY
    fi
    
    # Get runner name
    read -p "Runner name (default: $RUNNER_NAME): " input_name
    RUNNER_NAME=${input_name:-$RUNNER_NAME}
    
    # Check for auth tokens
    echo ""
    print_info "OneDrive Authentication (Optional)"
    
    if [[ -z "$AUTH_TOKENS_B64" ]]; then
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
            read -p "Path to auth_tokens.json (optional): " AUTH_PATH
            
            if [[ -n "$AUTH_PATH" ]] && [[ -f "$AUTH_PATH" ]]; then
                AUTH_TOKENS_B64=$(base64 -w 0 "$AUTH_PATH")
                print_success "Auth tokens encoded from $AUTH_PATH"
            fi
        fi
    fi
    
    # Save configuration
    cat > .docker-remote-config << EOF
DOCKER_HOST=$DOCKER_HOST
RUNNER_NAME=$RUNNER_NAME
GITHUB_TOKEN=$GITHUB_TOKEN
GITHUB_REPOSITORY=$GITHUB_REPOSITORY
AUTH_TOKENS_B64=$AUTH_TOKENS_B64
EOF
    
    chmod 600 .docker-remote-config
    print_success "Configuration saved to .docker-remote-config"
    
    echo ""
    print_success "Setup completed! Next steps:"
    echo "1. Run: $0 build"
    echo "2. Run: $0 deploy"
    echo "3. Check: $0 status"
}

# Function to load configuration
load_config() {
    if [[ -f .docker-remote-config ]]; then
        source .docker-remote-config
        print_info "Loaded configuration for Docker host: $DOCKER_HOST"
    fi
}

# Function to build the image on remote Docker
build_image() {
    print_info "Building OneMount GitHub Actions runner image on $DOCKER_HOST..."
    
    if ! test_docker_connection; then
        return 1
    fi
    
    # Create build context
    print_info "Creating build context..."
    tar -czf /tmp/onemount-build-context.tar.gz \
        --exclude='.git' \
        --exclude='build' \
        --exclude='test-artifacts' \
        --exclude='coverage' \
        --exclude='*.deb' \
        --exclude='*.rpm' \
        --exclude='vendor' \
        .
    
    # Build image on remote Docker
    print_info "Building image on remote Docker host..."
    DOCKER_HOST="tcp://$DOCKER_HOST" docker build \
        -f packaging/docker/Dockerfile.github-runner \
        -t "$IMAGE_NAME" \
        - < /tmp/onemount-build-context.tar.gz
    
    # Clean up
    rm /tmp/onemount-build-context.tar.gz
    
    print_success "Image built successfully on remote Docker host"
}

# Function to deploy the runner
deploy_runner() {
    print_info "Deploying GitHub Actions runner to $DOCKER_HOST..."
    
    if ! test_docker_connection; then
        return 1
    fi
    
    # Check if configuration is loaded
    if [[ -z "$GITHUB_TOKEN" ]] || [[ -z "$GITHUB_REPOSITORY" ]]; then
        print_error "GitHub configuration not found. Run '$0 setup' first."
        return 1
    fi
    
    # Stop existing container if running
    print_info "Stopping existing container if running..."
    DOCKER_HOST="tcp://$DOCKER_HOST" docker stop "$CONTAINER_NAME" 2>/dev/null || true
    DOCKER_HOST="tcp://$DOCKER_HOST" docker rm "$CONTAINER_NAME" 2>/dev/null || true
    
    # Create and start new container
    print_info "Creating and starting runner container..."
    DOCKER_HOST="tcp://$DOCKER_HOST" docker run -d \
        --name "$CONTAINER_NAME" \
        --restart unless-stopped \
        --device /dev/fuse \
        --cap-add SYS_ADMIN \
        --security-opt apparmor:unconfined \
        --dns 8.8.8.8 \
        --dns 8.8.4.4 \
        -e "GITHUB_TOKEN=$GITHUB_TOKEN" \
        -e "GITHUB_REPOSITORY=$GITHUB_REPOSITORY" \
        -e "RUNNER_NAME=$RUNNER_NAME" \
        -e "RUNNER_LABELS=self-hosted,linux,onemount-testing,docker-remote" \
        -e "AUTH_TOKENS_B64=$AUTH_TOKENS_B64" \
        -v onemount-runner-workspace:/workspace \
        -v onemount-runner-work:/opt/actions-runner/_work \
        "$IMAGE_NAME" run
    
    print_success "Runner deployed and started successfully"
    print_info "Container name: $CONTAINER_NAME"
    print_info "Runner name: $RUNNER_NAME"
    print_info "View logs with: $0 logs"
}

# Function to start the runner
start_runner() {
    print_info "Starting runner container..."
    
    if DOCKER_HOST="tcp://$DOCKER_HOST" docker start "$CONTAINER_NAME"; then
        print_success "Runner started successfully"
    else
        print_error "Failed to start runner container"
        return 1
    fi
}

# Function to stop the runner
stop_runner() {
    print_info "Stopping runner container..."
    
    if DOCKER_HOST="tcp://$DOCKER_HOST" docker stop "$CONTAINER_NAME"; then
        print_success "Runner stopped successfully"
    else
        print_error "Failed to stop runner container"
        return 1
    fi
}

# Function to restart the runner
restart_runner() {
    print_info "Restarting runner container..."
    
    if DOCKER_HOST="tcp://$DOCKER_HOST" docker restart "$CONTAINER_NAME"; then
        print_success "Runner restarted successfully"
    else
        print_error "Failed to restart runner container"
        return 1
    fi
}

# Function to check status
check_status() {
    print_info "Checking runner status on $DOCKER_HOST..."
    
    if ! test_docker_connection; then
        return 1
    fi
    
    echo ""
    print_info "Container status:"
    DOCKER_HOST="tcp://$DOCKER_HOST" docker ps -a --filter "name=$CONTAINER_NAME" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
    
    echo ""
    print_info "Container logs (last 10 lines):"
    DOCKER_HOST="tcp://$DOCKER_HOST" docker logs --tail 10 "$CONTAINER_NAME" 2>/dev/null || print_warning "Container not found or not running"
    
    echo ""
    print_info "Docker host resources:"
    DOCKER_HOST="tcp://$DOCKER_HOST" docker system df
}

# Function to view logs
view_logs() {
    print_info "Viewing runner logs..."
    
    DOCKER_HOST="tcp://$DOCKER_HOST" docker logs -f "$CONTAINER_NAME"
}

# Function to connect to shell
connect_shell() {
    print_info "Connecting to runner shell..."
    
    DOCKER_HOST="tcp://$DOCKER_HOST" docker exec -it "$CONTAINER_NAME" /bin/bash
}

# Function to update runner
update_runner() {
    print_info "Updating runner..."
    
    # Build new image
    build_image
    
    # Redeploy with new image
    deploy_runner
    
    print_success "Runner updated successfully"
}

# Function to clean up
clean_runner() {
    print_warning "This will remove the runner container and image"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Cleaning up runner..."
        
        # Stop and remove container
        DOCKER_HOST="tcp://$DOCKER_HOST" docker stop "$CONTAINER_NAME" 2>/dev/null || true
        DOCKER_HOST="tcp://$DOCKER_HOST" docker rm "$CONTAINER_NAME" 2>/dev/null || true
        
        # Remove image
        DOCKER_HOST="tcp://$DOCKER_HOST" docker rmi "$IMAGE_NAME" 2>/dev/null || true
        
        # Remove volumes
        read -p "Remove persistent volumes? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            DOCKER_HOST="tcp://$DOCKER_HOST" docker volume rm onemount-runner-workspace onemount-runner-work 2>/dev/null || true
        fi
        
        print_success "Cleanup completed"
    else
        print_info "Cleanup cancelled"
    fi
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
            DOCKER_HOST="$2"
            shift 2
            ;;
        --name)
            RUNNER_NAME="$2"
            shift 2
            ;;
        *)
            COMMAND="$1"
            shift
            ;;
    esac
done

# Load configuration
load_config

# Main execution
case "${COMMAND:-}" in
    setup)
        interactive_setup
        ;;
    build)
        build_image
        ;;
    deploy)
        deploy_runner
        ;;
    start)
        start_runner
        ;;
    stop)
        stop_runner
        ;;
    restart)
        restart_runner
        ;;
    status)
        check_status
        ;;
    logs)
        view_logs
        ;;
    shell)
        connect_shell
        ;;
    update)
        update_runner
        ;;
    clean)
        clean_runner
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
