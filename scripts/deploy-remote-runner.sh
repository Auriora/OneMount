#!/bin/bash
# Deploy OneMount GitHub Actions Runner to Remote Docker Host
# Deploys and manages the runner on a remote Docker host

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
REMOTE_HOST="172.16.1.104"
REMOTE_USER="ubuntu"  # Change this to your username
REMOTE_PORT="22"
PROJECT_NAME="onemount"
REMOTE_PROJECT_DIR="/opt/onemount-runner"

# Function to show usage
show_usage() {
    cat << EOF
OneMount Remote GitHub Actions Runner Deployment

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  setup             Setup the remote runner (interactive)
  deploy            Deploy runner to remote host
  start             Start the remote runner
  stop              Stop the remote runner
  status            Check runner status
  logs              View runner logs
  shell             Connect to runner shell
  update            Update runner code and restart
  clean             Clean up remote runner

Options:
  --host HOST       Remote host (default: $REMOTE_HOST)
  --user USER       Remote user (default: $REMOTE_USER)
  --port PORT       SSH port (default: $REMOTE_PORT)
  --project-dir DIR Remote project directory (default: $REMOTE_PROJECT_DIR)

Environment Variables:
  GITHUB_TOKEN      GitHub personal access token
  GITHUB_REPOSITORY Repository in format 'owner/repo'
  AUTH_TOKENS_B64   Base64-encoded OneDrive auth tokens

Examples:
  # Interactive setup
  $0 setup

  # Deploy with custom host
  $0 deploy --host 172.16.1.104 --user myuser

  # Check status
  $0 status

  # View logs
  $0 logs

  # Update and restart
  $0 update

EOF
}

# Function to check SSH connectivity
check_ssh() {
    print_info "Testing SSH connection to $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT..."
    
    if ssh -p "$REMOTE_PORT" -o ConnectTimeout=10 -o BatchMode=yes "$REMOTE_USER@$REMOTE_HOST" "echo 'SSH connection successful'" 2>/dev/null; then
        print_success "SSH connection established"
        return 0
    else
        print_error "SSH connection failed"
        print_info "Please ensure:"
        print_info "1. SSH key is configured for $REMOTE_USER@$REMOTE_HOST"
        print_info "2. Host is reachable and SSH service is running"
        print_info "3. User has appropriate permissions"
        return 1
    fi
}

# Function to check remote Docker
check_remote_docker() {
    print_info "Checking Docker on remote host..."
    
    if ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "docker --version && docker-compose --version" 2>/dev/null; then
        print_success "Docker and Docker Compose are available"
        return 0
    else
        print_error "Docker or Docker Compose not found on remote host"
        print_info "Please install Docker and Docker Compose on the remote host"
        return 1
    fi
}

# Function for interactive setup
interactive_setup() {
    print_info "OneMount Remote GitHub Actions Runner Setup"
    echo ""
    
    # Get remote host details
    read -p "Remote host IP/hostname (default: $REMOTE_HOST): " input_host
    REMOTE_HOST=${input_host:-$REMOTE_HOST}
    
    read -p "Remote username (default: $REMOTE_USER): " input_user
    REMOTE_USER=${input_user:-$REMOTE_USER}
    
    read -p "SSH port (default: $REMOTE_PORT): " input_port
    REMOTE_PORT=${input_port:-$REMOTE_PORT}
    
    read -p "Remote project directory (default: $REMOTE_PROJECT_DIR): " input_dir
    REMOTE_PROJECT_DIR=${input_dir:-$REMOTE_PROJECT_DIR}
    
    # Test connectivity
    if ! check_ssh; then
        return 1
    fi
    
    if ! check_remote_docker; then
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
    
    # Create remote environment file
    print_info "Creating remote configuration..."
    
    cat > /tmp/remote-runner.env << EOF
# OneMount GitHub Actions Runner Configuration
GITHUB_TOKEN=$GITHUB_TOKEN
GITHUB_REPOSITORY=$GITHUB_REPOSITORY
RUNNER_NAME=onemount-runner-$REMOTE_HOST
RUNNER_LABELS=self-hosted,linux,onemount-testing,docker-remote
RUNNER_GROUP=Default

# OneDrive Authentication (base64 encoded)
AUTH_TOKENS_B64=$AUTH_TOKENS_B64

EOF
    
    print_success "Configuration created"
    
    # Save configuration for future use
    cat > .remote-runner-config << EOF
REMOTE_HOST=$REMOTE_HOST
REMOTE_USER=$REMOTE_USER
REMOTE_PORT=$REMOTE_PORT
REMOTE_PROJECT_DIR=$REMOTE_PROJECT_DIR
EOF
    
    print_success "Setup completed! Run '$0 deploy' to deploy the runner."
}

# Function to load saved configuration
load_config() {
    if [[ -f .remote-runner-config ]]; then
        source .remote-runner-config
        print_info "Loaded configuration: $REMOTE_USER@$REMOTE_HOST:$REMOTE_PORT"
    fi
}

# Function to deploy runner to remote host
deploy_runner() {
    print_info "Deploying OneMount GitHub Actions runner to $REMOTE_USER@$REMOTE_HOST..."
    
    # Check connectivity
    if ! check_ssh; then
        return 1
    fi
    
    if ! check_remote_docker; then
        return 1
    fi
    
    # Create remote directory
    print_info "Creating remote project directory..."
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "sudo mkdir -p $REMOTE_PROJECT_DIR && sudo chown $REMOTE_USER:$REMOTE_USER $REMOTE_PROJECT_DIR"
    
    # Copy project files
    print_info "Copying project files..."
    
    # Create a temporary archive with necessary files
    tar -czf /tmp/onemount-runner.tar.gz \
        --exclude='.git' \
        --exclude='build' \
        --exclude='test-artifacts' \
        --exclude='coverage' \
        --exclude='*.deb' \
        --exclude='*.rpm' \
        --exclude='vendor' \
        .
    
    # Copy and extract on remote
    scp -P "$REMOTE_PORT" /tmp/onemount-runner.tar.gz "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PROJECT_DIR/"
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_PROJECT_DIR && tar -xzf onemount-runner.tar.gz && rm onemount-runner.tar.gz"
    
    # Copy environment file if it exists
    if [[ -f /tmp/remote-runner.env ]]; then
        scp -P "$REMOTE_PORT" /tmp/remote-runner.env "$REMOTE_USER@$REMOTE_HOST:$REMOTE_PROJECT_DIR/.env"
        rm /tmp/remote-runner.env
    fi
    
    # Build the runner image on remote host
    print_info "Building runner image on remote host..."
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_PROJECT_DIR && docker build -f packaging/docker/Dockerfile.github-runner -t onemount-github-runner ."
    
    print_success "Deployment completed successfully"
    
    # Clean up local temp file
    rm -f /tmp/onemount-runner.tar.gz
}

# Function to start remote runner
start_runner() {
    print_info "Starting remote GitHub Actions runner..."
    
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_PROJECT_DIR && docker-compose -f docker-compose.runner.yml up -d github-runner"
    
    print_success "Remote runner started"
    print_info "View logs with: $0 logs"
}

# Function to stop remote runner
stop_runner() {
    print_info "Stopping remote GitHub Actions runner..."
    
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_PROJECT_DIR && docker-compose -f docker-compose.runner.yml down"
    
    print_success "Remote runner stopped"
}

# Function to check runner status
check_status() {
    print_info "Checking remote runner status..."
    
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_PROJECT_DIR && docker-compose -f docker-compose.runner.yml ps"
}

# Function to view logs
view_logs() {
    print_info "Viewing remote runner logs..."
    
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_PROJECT_DIR && docker-compose -f docker-compose.runner.yml logs -f github-runner"
}

# Function to connect to shell
connect_shell() {
    print_info "Connecting to remote runner shell..."
    
    ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_PROJECT_DIR && docker-compose -f docker-compose.runner.yml exec github-runner /bin/bash"
}

# Function to update runner
update_runner() {
    print_info "Updating remote runner..."
    
    # Stop current runner
    stop_runner
    
    # Deploy updated code
    deploy_runner
    
    # Start runner
    start_runner
    
    print_success "Runner updated successfully"
}

# Function to clean up
clean_runner() {
    print_warning "This will remove the remote runner and all its data"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Cleaning up remote runner..."
        
        ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "cd $REMOTE_PROJECT_DIR && docker-compose -f docker-compose.runner.yml down -v && docker image rm onemount-github-runner 2>/dev/null || true"
        
        read -p "Remove project directory $REMOTE_PROJECT_DIR? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            ssh -p "$REMOTE_PORT" "$REMOTE_USER@$REMOTE_HOST" "sudo rm -rf $REMOTE_PROJECT_DIR"
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
            REMOTE_HOST="$2"
            shift 2
            ;;
        --user)
            REMOTE_USER="$2"
            shift 2
            ;;
        --port)
            REMOTE_PORT="$2"
            shift 2
            ;;
        --project-dir)
            REMOTE_PROJECT_DIR="$2"
            shift 2
            ;;
        *)
            COMMAND="$1"
            shift
            ;;
    esac
done

# Load saved configuration
load_config

# Main execution
case "${COMMAND:-}" in
    setup)
        interactive_setup
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
