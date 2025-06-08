#!/bin/bash

# Lightweight Elastic Runner Monitor
# Monitors GitHub Actions queue and manages elastic runners

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MIN_RUNNERS="${MIN_RUNNERS:-2}"
MAX_RUNNERS="${MAX_RUNNERS:-5}"
SCALE_UP_THRESHOLD="${SCALE_UP_THRESHOLD:-2}"
SCALE_DOWN_THRESHOLD="${SCALE_DOWN_THRESHOLD:-0}"
CHECK_INTERVAL="${CHECK_INTERVAL:-30}"
COOLDOWN_PERIOD="${COOLDOWN_PERIOD:-300}"

# Logging functions
log_info() { echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')] INFO:${NC} $1"; }
log_success() { echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS:${NC} $1"; }
log_warning() { echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"; }
log_error() { echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"; }

# Load environment variables
load_env() {
    if [[ -f "/workspace/.env" ]]; then
        set -a
        source "/workspace/.env"
        set +a
        log_info "Loaded environment from /workspace/.env"
    elif [[ -f ".env" ]]; then
        set -a
        source ".env"
        set +a
        log_info "Loaded environment from .env"
    else
        log_error "No .env file found"
        exit 1
    fi
}

# GitHub API helper
github_api() {
    local endpoint="$1"
    curl -s -H "Authorization: token $GITHUB_TOKEN" \
         -H "Accept: application/vnd.github.v3+json" \
         "https://api.github.com/repos/$GITHUB_REPOSITORY/$endpoint"
}

# Get current queue length
get_queue_length() {
    local queued_runs
    if command -v jq >/dev/null 2>&1; then
        queued_runs=$(github_api "actions/runs?status=queued&per_page=100" | jq '.workflow_runs | length' 2>/dev/null || echo "0")
    else
        # Fallback without jq - count "queued" occurrences
        queued_runs=$(github_api "actions/runs?status=queued&per_page=100" | grep -o '"status":"queued"' | wc -l)
    fi
    echo "${queued_runs:-0}"
}

# Get current elastic runner count and status
get_elastic_runners() {
    if command -v jq >/dev/null 2>&1; then
        github_api "actions/runners" | jq -r '.runners[] | select(.name | startswith("onemount-elastic")) | "\(.name):\(.status):\(.busy)"' 2>/dev/null || echo ""
    else
        # Fallback without jq - parse JSON manually
        github_api "actions/runners" | grep -A 20 '"name":"onemount-elastic' | \
        awk '/"name":"onemount-elastic[^"]*"/ { name=$0; getline; getline; getline; getline; getline; getline; getline; getline; getline; status=$0; getline; busy=$0; gsub(/.*"name":"/, "", name); gsub(/".*/, "", name); gsub(/.*"status":"/, "", status); gsub(/".*/, "", status); gsub(/.*"busy":/, "", busy); gsub(/[,}].*/, "", busy); print name":"status":"busy }'
    fi
}

# Count online elastic runners
count_online_runners() {
    get_elastic_runners | grep ":online:" | wc -l
}

# Count busy elastic runners
count_busy_runners() {
    get_elastic_runners | grep ":online:true" | wc -l
}

# Start a new elastic runner
start_runner() {
    local runner_id="$1"
    local runner_name="onemount-elastic-${runner_id}"
    
    log_info "Starting elastic runner: $runner_name"
    
    # Use our manage-elastic-runners.sh script
    if [[ -f "/workspace/scripts/manage-elastic-runners.sh" ]]; then
        /workspace/scripts/manage-elastic-runners.sh start "$runner_id"
    else
        # Fallback to direct Docker command
        docker run -d \
            --name "$runner_name" \
            --env-file /tmp/onemount.env \
            -e "RUNNER_NAME=$runner_name" \
            -e "RUNNER_ALLOW_RUNASROOT=1" \
            --tmpfs /tmp \
            --tmpfs /opt/actions-runner/_update \
            --label "com.docker.compose.project=onemount-runners" \
            --restart unless-stopped \
            onemount-github-runner:latest run
    fi
    
    log_success "Started elastic runner: $runner_name"
}

# Stop an idle elastic runner
stop_idle_runner() {
    local runner_name="$1"
    
    log_info "Stopping idle runner: $runner_name"
    
    # Use our manage-elastic-runners.sh script
    if [[ -f "/workspace/scripts/manage-elastic-runners.sh" ]]; then
        /workspace/scripts/manage-elastic-runners.sh stop "$runner_name"
    else
        # Fallback to direct Docker command
        docker stop "$runner_name" 2>/dev/null || true
        docker rm "$runner_name" 2>/dev/null || true
    fi
    
    log_success "Stopped idle runner: $runner_name"
}

# Auto-scaling logic
auto_scale() {
    local queue_length
    local online_runners
    local busy_runners
    local idle_runners
    
    queue_length=$(get_queue_length)
    online_runners=$(count_online_runners)
    busy_runners=$(count_busy_runners)
    idle_runners=$((online_runners - busy_runners))
    
    log_info "Queue: $queue_length, Online: $online_runners, Busy: $busy_runners, Idle: $idle_runners"
    
    # Scale up logic
    if [[ $queue_length -ge $SCALE_UP_THRESHOLD && $online_runners -lt $MAX_RUNNERS ]]; then
        local needed_runners=$((queue_length - online_runners))
        local max_new_runners=$((MAX_RUNNERS - online_runners))
        local new_runners=$((needed_runners > max_new_runners ? max_new_runners : needed_runners))
        
        if [[ $new_runners -gt 0 ]]; then
            log_info "Scaling up: adding $new_runners runners"
            
            for ((i=1; i<=new_runners; i++)); do
                # Find next available ID
                local next_id=1
                while get_elastic_runners | grep -q "onemount-elastic-${next_id}:"; do
                    ((next_id++))
                done
                
                start_runner "$next_id"
                sleep 5  # Give time for registration
            done
        fi
    fi
    
    # Scale down logic
    if [[ $queue_length -le $SCALE_DOWN_THRESHOLD && $online_runners -gt $MIN_RUNNERS && $idle_runners -gt 0 ]]; then
        local excess_runners=$((online_runners - MIN_RUNNERS))
        local to_remove=$((idle_runners > excess_runners ? excess_runners : idle_runners))
        
        if [[ $to_remove -gt 0 ]]; then
            log_info "Scaling down: removing $to_remove idle runners"
            
            # Get list of idle runners
            local idle_runner_names=()
            while IFS= read -r line; do
                if [[ -n "$line" && "$line" == *":online:false" ]]; then
                    local name=$(echo "$line" | cut -d: -f1)
                    idle_runner_names+=("$name")
                fi
            done <<< "$(get_elastic_runners)"
            
            # Remove excess idle runners
            local removed=0
            for runner_name in "${idle_runner_names[@]}"; do
                if [[ $removed -ge $to_remove ]]; then
                    break
                fi
                
                stop_idle_runner "$runner_name"
                ((removed++))
                sleep 2
            done
        fi
    fi
}

# Monitor and auto-scale
monitor() {
    log_info "Starting elastic runner monitoring..."
    log_info "Min runners: $MIN_RUNNERS, Max runners: $MAX_RUNNERS"
    log_info "Scale up threshold: $SCALE_UP_THRESHOLD, Scale down threshold: $SCALE_DOWN_THRESHOLD"
    log_info "Check interval: ${CHECK_INTERVAL}s, Cooldown: ${COOLDOWN_PERIOD}s"
    
    local last_scale_time=0
    
    while true; do
        local current_time=$(date +%s)
        
        # Only scale if cooldown period has passed
        if [[ $((current_time - last_scale_time)) -ge $COOLDOWN_PERIOD ]]; then
            auto_scale
            last_scale_time=$current_time
        else
            local remaining=$((COOLDOWN_PERIOD - (current_time - last_scale_time)))
            log_info "Cooldown active, $remaining seconds remaining"
        fi
        
        sleep "$CHECK_INTERVAL"
    done
}

# Show status
status() {
    local queue_length
    local online_runners
    local busy_runners
    
    queue_length=$(get_queue_length)
    online_runners=$(count_online_runners)
    busy_runners=$(count_busy_runners)
    
    echo "=== Elastic Runner Status ==="
    echo "Queue length: $queue_length"
    echo "Online runners: $online_runners"
    echo "Busy runners: $busy_runners"
    echo "Idle runners: $((online_runners - busy_runners))"
    echo "Min/Max runners: $MIN_RUNNERS/$MAX_RUNNERS"
    echo ""
    echo "=== Runner Details ==="
    get_elastic_runners | while IFS= read -r line; do
        if [[ -n "$line" ]]; then
            local name=$(echo "$line" | cut -d: -f1)
            local status=$(echo "$line" | cut -d: -f2)
            local busy=$(echo "$line" | cut -d: -f3)
            echo "$name: $status (busy: $busy)"
        fi
    done
}

# Main execution
case "${1:-help}" in
    monitor)
        load_env
        monitor
        ;;
    status)
        load_env
        status
        ;;
    help|--help|-h)
        echo "Usage: $0 {monitor|status|help}"
        echo ""
        echo "Commands:"
        echo "  monitor  - Start continuous monitoring and auto-scaling"
        echo "  status   - Show current runner status"
        echo "  help     - Show this help"
        ;;
    *)
        echo "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac
