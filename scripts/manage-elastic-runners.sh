#!/bin/bash

# Elastic GitHub Actions Runner Management Script
# Manages elastic runners with persistent naming and tokens

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker/compose/docker-compose.elastic.yml"
ENV_FILE=".env"
STACK_NAME="onemount-runners"
RUNNER_PREFIX="onemount-elastic"
MAX_RUNNERS=5
MIN_RUNNERS=2

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if running on remote server
is_remote() {
    [[ "${HOSTNAME}" == *"172.16.1.104"* ]] || [[ "${SSH_CLIENT}" != "" ]]
}

# Get Docker command prefix
get_docker_cmd() {
    if is_remote; then
        echo "docker"
    else
        echo "ssh 172.16.1.104 docker"
    fi
}

# Get Docker Compose command prefix
get_compose_cmd() {
    if is_remote; then
        echo "docker compose -f /tmp/docker-compose.elastic.yml --env-file /tmp/onemount.env"
    else
        echo "ssh 172.16.1.104 'cd /tmp && docker compose -f docker-compose.elastic.yml --env-file onemount.env'"
    fi
}

# Copy files to remote if needed
sync_files() {
    if ! is_remote; then
        log_info "Syncing files to remote server..."
        scp "$COMPOSE_FILE" 172.16.1.104:/tmp/docker-compose.elastic.yml
        scp "$ENV_FILE" 172.16.1.104:/tmp/onemount.env
        log_success "Files synced to remote server"
    fi
}

# List current runners
list_runners() {
    log_info "Current elastic runners:"
    
    DOCKER_CMD=$(get_docker_cmd)
    
    # List containers with elastic runner prefix
    if is_remote; then
        docker ps -a --format "table {{.Names}}\t{{.Status}}\t{{.Image}}" \
            --filter "name=${RUNNER_PREFIX}" | head -20
    else
        ssh 172.16.1.104 "docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Image}}' \
            --filter 'name=${RUNNER_PREFIX}'" | head -20
    fi
}

# Start a new elastic runner
start_runner() {
    local runner_id="$1"
    local runner_name="${RUNNER_PREFIX}-${runner_id}"
    
    log_info "Starting elastic runner: $runner_name"
    
    sync_files
    
    # Use the working simple runner approach with persistence
    DOCKER_CMD=$(get_docker_cmd)
    
    # Create persistent volumes for this runner
    $DOCKER_CMD volume create "${STACK_NAME}_${runner_name}-work" 2>/dev/null || true
    $DOCKER_CMD volume create "${STACK_NAME}_${runner_name}-data" 2>/dev/null || true
    $DOCKER_CMD volume create "${STACK_NAME}_${runner_name}-credentials" 2>/dev/null || true
    
    # Start the runner with persistence
    if is_remote; then
        docker run -d \
            --name "$runner_name" \
            --env-file /tmp/onemount.env \
            -e "RUNNER_NAME=$runner_name" \
            -e "RUNNER_ALLOW_RUNASROOT=1" \
            -v "${STACK_NAME}_${runner_name}-work:/opt/actions-runner/_work" \
            -v "${STACK_NAME}_${runner_name}-data:/opt/actions-runner/.runner" \
            -v "${STACK_NAME}_${runner_name}-credentials:/opt/actions-runner/.credentials" \
            --tmpfs /tmp \
            --tmpfs /opt/actions-runner/_update \
            --label "com.docker.compose.project=${STACK_NAME}" \
            --label "com.docker.compose.service=elastic-runner" \
            --restart unless-stopped \
            onemount-github-runner:latest run
    else
        ssh 172.16.1.104 "docker run -d \
            --name '$runner_name' \
            --env-file /tmp/onemount.env \
            -e 'RUNNER_NAME=$runner_name' \
            -e 'RUNNER_ALLOW_RUNASROOT=1' \
            -v '${STACK_NAME}_${runner_name}-work:/opt/actions-runner/_work' \
            -v '${STACK_NAME}_${runner_name}-data:/opt/actions-runner/.runner' \
            -v '${STACK_NAME}_${runner_name}-credentials:/opt/actions-runner/.credentials' \
            --tmpfs /tmp \
            --tmpfs /opt/actions-runner/_update \
            --label 'com.docker.compose.project=${STACK_NAME}' \
            --label 'com.docker.compose.service=elastic-runner' \
            --restart unless-stopped \
            onemount-github-runner:latest run"
    fi
    
    log_success "Started elastic runner: $runner_name"
}

# Stop and remove a runner
stop_runner() {
    local runner_name="$1"
    
    log_info "Stopping elastic runner: $runner_name"
    
    DOCKER_CMD=$(get_docker_cmd)
    
    # Stop and remove container
    $DOCKER_CMD stop "$runner_name" 2>/dev/null || true
    $DOCKER_CMD rm "$runner_name" 2>/dev/null || true
    
    log_success "Stopped elastic runner: $runner_name"
}

# Scale runners to target count
scale_runners() {
    local target_count="$1"
    
    log_info "Scaling elastic runners to $target_count instances"
    
    DOCKER_CMD=$(get_docker_cmd)
    
    # Get current running elastic runners
    current_runners=($($DOCKER_CMD ps --format "{{.Names}}" --filter "name=${RUNNER_PREFIX}" | sort))
    current_count=${#current_runners[@]}
    
    log_info "Current runners: $current_count, Target: $target_count"
    
    if [ "$current_count" -lt "$target_count" ]; then
        # Scale up
        local needed=$((target_count - current_count))
        log_info "Scaling up: adding $needed runners"
        
        for i in $(seq 1 $needed); do
            # Find next available ID
            local next_id=1
            while $DOCKER_CMD ps -a --format "{{.Names}}" | grep -q "${RUNNER_PREFIX}-${next_id}"; do
                ((next_id++))
            done
            
            start_runner "$next_id"
            sleep 5  # Give time for registration
        done
        
    elif [ "$current_count" -gt "$target_count" ]; then
        # Scale down
        local excess=$((current_count - target_count))
        log_info "Scaling down: removing $excess runners"
        
        # Remove excess runners (last ones first)
        for ((i=current_count; i>target_count; i--)); do
            if [ ${#current_runners[@]} -ge $i ]; then
                stop_runner "${current_runners[$((i-1))]}"
            fi
        done
    else
        log_success "Already at target scale: $target_count runners"
    fi
}

# Show usage
usage() {
    echo "Usage: $0 {start|stop|list|scale|restart|status}"
    echo ""
    echo "Commands:"
    echo "  start <id>     - Start a new elastic runner with given ID"
    echo "  stop <name>    - Stop and remove a specific runner"
    echo "  list           - List all elastic runners"
    echo "  scale <count>  - Scale to specific number of runners"
    echo "  restart        - Restart all elastic runners"
    echo "  status         - Show detailed status"
    echo ""
    echo "Examples:"
    echo "  $0 start 1                    # Start onemount-elastic-1"
    echo "  $0 stop onemount-elastic-1    # Stop specific runner"
    echo "  $0 scale 3                    # Scale to 3 runners"
    echo "  $0 list                       # List all runners"
}

# Main command handling
case "${1:-}" in
    start)
        if [ -z "$2" ]; then
            log_error "Runner ID required"
            usage
            exit 1
        fi
        start_runner "$2"
        ;;
    stop)
        if [ -z "$2" ]; then
            log_error "Runner name required"
            usage
            exit 1
        fi
        stop_runner "$2"
        ;;
    list)
        list_runners
        ;;
    scale)
        if [ -z "$2" ]; then
            log_error "Target count required"
            usage
            exit 1
        fi
        scale_runners "$2"
        ;;
    restart)
        log_info "Restarting all elastic runners..."
        scale_runners 0
        sleep 5
        scale_runners "$MIN_RUNNERS"
        ;;
    status)
        list_runners
        echo ""
        log_info "Checking GitHub registration status..."
        # This would need GitHub API integration
        ;;
    *)
        usage
        exit 1
        ;;
esac
