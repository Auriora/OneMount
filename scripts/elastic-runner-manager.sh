#!/bin/bash

# OneMount Elastic GitHub Actions Runner Manager
# Provides auto-scaling capabilities for Docker-based self-hosted runners

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
DOCKER_HOST="${DOCKER_HOST:-172.16.1.104:2376}"
RUNNER_IMAGE="${RUNNER_IMAGE:-onemount-github-runner:latest}"
MIN_RUNNERS="${MIN_RUNNERS:-1}"
MAX_RUNNERS="${MAX_RUNNERS:-5}"
SCALE_UP_THRESHOLD="${SCALE_UP_THRESHOLD:-2}"
SCALE_DOWN_THRESHOLD="${SCALE_DOWN_THRESHOLD:-0}"
CHECK_INTERVAL="${CHECK_INTERVAL:-30}"
COOLDOWN_PERIOD="${COOLDOWN_PERIOD:-300}"

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

# Load environment variables
load_env() {
    local env_file="$PROJECT_ROOT/.env"
    if [[ -f "$env_file" ]]; then
        set -a
        source "$env_file"
        set +a
    else
        print_error "Environment file not found: $env_file"
        print_info "Run './scripts/manage-runner.sh setup' first"
        exit 1
    fi
}

# Get GitHub API data
github_api() {
    local endpoint="$1"
    curl -s -H "Authorization: token $GITHUB_TOKEN" \
         -H "Accept: application/vnd.github.v3+json" \
         "https://api.github.com/repos/$GITHUB_REPOSITORY/$endpoint"
}

# Get current queue length
get_queue_length() {
    local queued_runs
    queued_runs=$(github_api "actions/runs?status=queued&per_page=100" | jq '.workflow_runs | length')
    echo "${queued_runs:-0}"
}

# Get current runner count
get_runner_count() {
    local running_containers
    running_containers=$(DOCKER_HOST="tcp://$DOCKER_HOST" docker ps -q --filter "name=onemount-runner-elastic-*" | wc -l)
    echo "$running_containers"
}

# Get runner status
get_runner_status() {
    local runners
    runners=$(github_api "actions/runners" | jq -r '.runners[] | select(.name | startswith("onemount-runner-elastic")) | "\(.name):\(.status):\(.busy)"')
    echo "$runners"
}

# Scale up runners
scale_up() {
    local current_count="$1"
    local target_count="$2"
    
    print_info "Scaling up from $current_count to $target_count runners"
    
    for ((i=current_count+1; i<=target_count; i++)); do
        local runner_name="onemount-runner-elastic-$i"
        local container_name="$runner_name"
        
        print_info "Starting runner: $runner_name"
        
        # Start new runner container
        DOCKER_HOST="tcp://$DOCKER_HOST" docker run -d \
            --name "$container_name" \
            --restart unless-stopped \
            --device /dev/fuse \
            --cap-add SYS_ADMIN \
            --security-opt apparmor:unconfined \
            --dns 8.8.8.8 \
            --dns 8.8.4.4 \
            -e "GITHUB_TOKEN=$GITHUB_TOKEN" \
            -e "GITHUB_REPOSITORY=$GITHUB_REPOSITORY" \
            -e "RUNNER_NAME=$runner_name" \
            -e "RUNNER_LABELS=self-hosted,Linux,onemount-testing,optimized,elastic" \
            -e "AUTH_TOKENS_B64=${AUTH_TOKENS_B64:-}" \
            -v "onemount-runner-elastic-$i-workspace:/workspace" \
            -v "onemount-runner-elastic-$i-work:/opt/actions-runner/_work" \
            "$RUNNER_IMAGE" run
        
        print_success "Started runner: $runner_name"
        sleep 5  # Brief delay between starts
    done
}

# Scale down runners
scale_down() {
    local current_count="$1"
    local target_count="$2"
    
    print_info "Scaling down from $current_count to $target_count runners"
    
    # Get list of idle runners
    local idle_runners=()
    while IFS= read -r line; do
        if [[ -n "$line" ]]; then
            local name=$(echo "$line" | cut -d: -f1)
            local status=$(echo "$line" | cut -d: -f2)
            local busy=$(echo "$line" | cut -d: -f3)
            
            if [[ "$status" == "online" && "$busy" == "false" ]]; then
                idle_runners+=("$name")
            fi
        fi
    done <<< "$(get_runner_status)"
    
    # Remove excess idle runners
    local to_remove=$((current_count - target_count))
    local removed=0
    
    for runner_name in "${idle_runners[@]}"; do
        if [[ $removed -ge $to_remove ]]; then
            break
        fi
        
        print_info "Stopping runner: $runner_name"
        
        # Stop and remove container
        DOCKER_HOST="tcp://$DOCKER_HOST" docker stop "$runner_name" || true
        DOCKER_HOST="tcp://$DOCKER_HOST" docker rm "$runner_name" || true
        
        print_success "Stopped runner: $runner_name"
        ((removed++))
    done
}

# Auto-scaling logic
auto_scale() {
    local queue_length
    local current_runners
    local target_runners
    
    queue_length=$(get_queue_length)
    current_runners=$(get_runner_count)
    target_runners=$current_runners
    
    print_info "Queue: $queue_length, Current runners: $current_runners"
    
    # Scale up logic
    if [[ $queue_length -ge $SCALE_UP_THRESHOLD && $current_runners -lt $MAX_RUNNERS ]]; then
        target_runners=$((queue_length > MAX_RUNNERS ? MAX_RUNNERS : queue_length))
        if [[ $target_runners -gt $current_runners ]]; then
            scale_up "$current_runners" "$target_runners"
        fi
    fi
    
    # Scale down logic
    if [[ $queue_length -le $SCALE_DOWN_THRESHOLD && $current_runners -gt $MIN_RUNNERS ]]; then
        target_runners=$MIN_RUNNERS
        if [[ $target_runners -lt $current_runners ]]; then
            scale_down "$current_runners" "$target_runners"
        fi
    fi
}

# Monitor and auto-scale
monitor() {
    print_info "Starting elastic runner monitoring..."
    print_info "Min runners: $MIN_RUNNERS, Max runners: $MAX_RUNNERS"
    print_info "Scale up threshold: $SCALE_UP_THRESHOLD, Scale down threshold: $SCALE_DOWN_THRESHOLD"
    print_info "Check interval: ${CHECK_INTERVAL}s, Cooldown: ${COOLDOWN_PERIOD}s"
    
    local last_scale_time=0
    
    while true; do
        local current_time=$(date +%s)
        
        # Only scale if cooldown period has passed
        if [[ $((current_time - last_scale_time)) -ge $COOLDOWN_PERIOD ]]; then
            auto_scale
            last_scale_time=$current_time
        else
            local remaining=$((COOLDOWN_PERIOD - (current_time - last_scale_time)))
            print_info "Cooldown active, $remaining seconds remaining"
        fi
        
        sleep "$CHECK_INTERVAL"
    done
}

# Show status
show_status() {
    local queue_length
    local current_runners
    
    queue_length=$(get_queue_length)
    current_runners=$(get_runner_count)
    
    echo "=== Elastic Runner Status ==="
    echo "Queue length: $queue_length"
    echo "Current runners: $current_runners"
    echo "Min/Max runners: $MIN_RUNNERS/$MAX_RUNNERS"
    echo ""
    echo "=== Runner Details ==="
    get_runner_status | while IFS= read -r line; do
        if [[ -n "$line" ]]; then
            local name=$(echo "$line" | cut -d: -f1)
            local status=$(echo "$line" | cut -d: -f2)
            local busy=$(echo "$line" | cut -d: -f3)
            echo "$name: $status (busy: $busy)"
        fi
    done
}

# Clean up all elastic runners
cleanup() {
    print_info "Cleaning up all elastic runners..."
    
    # Stop and remove all elastic runner containers
    DOCKER_HOST="tcp://$DOCKER_HOST" docker ps -a -q --filter "name=onemount-runner-elastic-*" | \
        xargs -r DOCKER_HOST="tcp://$DOCKER_HOST" docker rm -f
    
    # Remove volumes (optional)
    DOCKER_HOST="tcp://$DOCKER_HOST" docker volume ls -q --filter "name=onemount-runner-elastic-*" | \
        xargs -r DOCKER_HOST="tcp://$DOCKER_HOST" docker volume rm
    
    print_success "Cleanup completed"
}

# Show usage
show_usage() {
    cat << EOF
OneMount Elastic Runner Manager

Usage: $0 [COMMAND] [OPTIONS]

Commands:
  monitor           Start auto-scaling monitoring (runs continuously)
  status            Show current runner status
  scale-up [N]      Manually scale up to N runners
  scale-down [N]    Manually scale down to N runners
  cleanup           Remove all elastic runners
  help              Show this help

Environment Variables:
  DOCKER_HOST              Remote Docker host (default: 172.16.1.104:2376)
  MIN_RUNNERS              Minimum runners (default: 1)
  MAX_RUNNERS              Maximum runners (default: 5)
  SCALE_UP_THRESHOLD       Queue length to trigger scale up (default: 2)
  SCALE_DOWN_THRESHOLD     Queue length to trigger scale down (default: 0)
  CHECK_INTERVAL           Monitoring interval in seconds (default: 30)
  COOLDOWN_PERIOD          Cooldown between scaling actions (default: 300)

Examples:
  $0 monitor                    # Start auto-scaling
  $0 status                     # Check current status
  MIN_RUNNERS=2 MAX_RUNNERS=10 $0 monitor  # Custom scaling limits

EOF
}

# Main execution
main() {
    local command="${1:-help}"
    
    case "$command" in
        monitor)
            load_env
            monitor
            ;;
        status)
            load_env
            show_status
            ;;
        scale-up)
            load_env
            local target="${2:-$(($(get_runner_count) + 1))}"
            scale_up "$(get_runner_count)" "$target"
            ;;
        scale-down)
            load_env
            local target="${2:-$(($(get_runner_count) - 1))}"
            scale_down "$(get_runner_count)" "$target"
            ;;
        cleanup)
            cleanup
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
