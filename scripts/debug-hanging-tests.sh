#!/bin/bash

# Enhanced debugging script for hanging FUSE filesystem tests
# Usage: ./scripts/debug-hanging-tests.sh <test-pattern> <timeout-seconds>

set -e

TEST_PATTERN="${1:-TestIT_FS_ETag_01_CacheValidationWithIfNoneMatch_Fixed}"
TIMEOUT_SECONDS="${2:-120}"
DEBUG_DIR="test-artifacts/debug"
LOG_FILE="$DEBUG_DIR/test-execution.log"
GOROUTINE_DUMP="$DEBUG_DIR/goroutines.txt"
PROCESS_INFO="$DEBUG_DIR/process-info.txt"
FUSE_DEBUG="$DEBUG_DIR/fuse-debug.log"

echo "🔍 Enhanced Debugging for Hanging FUSE Tests"
echo "=============================================="
echo "Test pattern: $TEST_PATTERN"
echo "Timeout: ${TIMEOUT_SECONDS}s"
echo "Debug directory: $DEBUG_DIR"
echo ""

# Setup debug directory
mkdir -p "$DEBUG_DIR"
rm -f "$DEBUG_DIR"/*

# Function to capture system state
capture_system_state() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] Capturing system state..." | tee -a "$PROCESS_INFO"
    
    # Process information
    {
        echo "=== PROCESS TREE ==="
        pstree -p $$ 2>/dev/null || echo "pstree not available"
        echo ""
        
        echo "=== DOCKER PROCESSES ==="
        docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "Docker not available"
        echo ""
        
        echo "=== MOUNT POINTS ==="
        mount | grep -E "(fuse|onemount)" || echo "No FUSE mounts found"
        echo ""
        
        echo "=== MEMORY USAGE ==="
        free -h 2>/dev/null || echo "Memory info not available"
        echo ""
        
        echo "=== DISK USAGE ==="
        df -h /tmp 2>/dev/null || echo "Disk info not available"
        echo ""
        
    } >> "$PROCESS_INFO"
}

# Function to capture goroutine dump
capture_goroutine_dump() {
    local test_pid=$1
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo "[$timestamp] Attempting to capture goroutine dump..." | tee -a "$GOROUTINE_DUMP"
    
    # Try to send SIGQUIT to get goroutine dump
    if kill -QUIT $test_pid 2>/dev/null; then
        echo "Sent SIGQUIT to process $test_pid" >> "$GOROUTINE_DUMP"
        sleep 2
    else
        echo "Failed to send SIGQUIT to process $test_pid" >> "$GOROUTINE_DUMP"
    fi
    
    # Try to use pstack if available
    if command -v pstack >/dev/null 2>&1; then
        echo "=== PSTACK OUTPUT ===" >> "$GOROUTINE_DUMP"
        pstack $test_pid 2>/dev/null >> "$GOROUTINE_DUMP" || echo "pstack failed" >> "$GOROUTINE_DUMP"
    fi
}

# Function to monitor FUSE operations
monitor_fuse_operations() {
    local container_name=$1
    
    echo "🔍 Monitoring FUSE operations in container: $container_name"
    
    # Try to enable FUSE debugging in the container
    docker exec "$container_name" bash -c '
        echo "=== FUSE DEBUG INFO ===" 
        echo "FUSE device status:"
        ls -la /dev/fuse 2>/dev/null || echo "FUSE device not found"
        echo ""
        
        echo "FUSE mount points:"
        mount | grep fuse || echo "No FUSE mounts"
        echo ""
        
        echo "FUSE processes:"
        ps aux | grep -E "(fuse|onemount)" | grep -v grep || echo "No FUSE processes"
        echo ""
        
        echo "Open file descriptors:"
        lsof | grep -E "(fuse|onemount)" | head -10 || echo "No FUSE file descriptors"
        echo ""
    ' >> "$FUSE_DEBUG" 2>&1 &
}

# Enhanced container monitoring with proper timeout
monitor_container_with_timeout() {
    local container_id=$1
    local container_name=$2
    local start_time=$(date +%s)
    local last_output_time=$start_time
    local heartbeat_interval=10
    local debug_interval=30
    local last_debug_time=$start_time
    
    echo "📊 Enhanced container monitor started"
    echo "🐳 Monitoring container: $container_name ($container_id)"
    echo ""
    
    # Initial system state capture
    capture_system_state
    
    while docker ps --quiet --filter "id=$container_id" | grep -q "$container_id"; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))
        local since_output=$((current_time - last_output_time))
        local since_debug=$((current_time - last_debug_time))
        
        # Check for new output
        if [[ -f "$LOG_FILE" ]]; then
            local log_size=$(wc -l < "$LOG_FILE" 2>/dev/null || echo "0")
            if [[ $log_size -gt ${last_log_size:-0} ]]; then
                last_output_time=$current_time
                last_log_size=$log_size
                echo "📝 New output detected (${log_size} lines total)"
                
                # Show recent output
                echo "   Recent output:"
                tail -3 "$LOG_FILE" 2>/dev/null | sed 's/^/   | /' || echo "   (no recent output)"
            fi
        fi
        
        # Regular heartbeat
        if [[ $((elapsed % heartbeat_interval)) -eq 0 ]]; then
            echo "⏱️  HEARTBEAT: ${elapsed}s elapsed | No output for ${since_output}s"
            
            # Show container status
            local container_status=$(docker inspect "$container_id" --format='{{.State.Status}}' 2>/dev/null || echo "unknown")
            echo "   Container status: $container_status"
        fi
        
        # Detailed debugging at intervals
        if [[ $since_debug -gt $debug_interval ]]; then
            echo "🔬 DETAILED DEBUG (${elapsed}s elapsed):"
            
            # Capture system state
            capture_system_state
            
            # Monitor FUSE operations
            monitor_fuse_operations "$container_name"
            
            # Check for specific hang indicators
            if [[ $since_output -gt 45 ]]; then
                echo "⚠️  HANG DETECTED: No output for ${since_output}s"
                echo "   This suggests the test is stuck in a blocking operation"
                
                # Try to identify what's blocking
                if [[ -f "$LOG_FILE" ]]; then
                    echo "   Last few log lines:"
                    tail -5 "$LOG_FILE" 2>/dev/null | sed 's/^/   | /' || echo "   (no log content)"
                fi
                
                # Try to get container processes
                echo "   Container processes:"
                docker exec "$container_id" ps aux 2>/dev/null | head -10 || echo "   (cannot get processes)"
            fi
            
            last_debug_time=$current_time
            echo ""
        fi
        
        # Check timeout - this is the key fix
        if [[ $elapsed -gt $TIMEOUT_SECONDS ]]; then
            echo "❌ TIMEOUT: Container exceeded ${TIMEOUT_SECONDS}s limit"
            echo "🔪 Attempting graceful shutdown..."
            
            # Final debug capture
            capture_system_state
            
            # Try to get final container state
            echo "   Final container status:"
            docker inspect "$container_id" --format='{{.State.Status}} ({{.State.ExitCode}})' 2>/dev/null || echo "   (cannot get status)"
            
            # Try graceful termination first
            echo "   Sending SIGTERM to container..."
            docker kill --signal=TERM "$container_id" 2>/dev/null || true
            sleep 5
            
            # Force kill if still running
            if docker ps --quiet --filter "id=$container_id" | grep -q "$container_id"; then
                echo "   Force killing container..."
                docker kill "$container_id" 2>/dev/null || true
            fi
            
            return 1
        fi
        
        sleep 2
    done
    
    local final_time=$(date +%s)
    local total_elapsed=$((final_time - start_time))
    echo "✅ Container completed after ${total_elapsed}s"
    
    # Final system state capture
    capture_system_state
    
    return 0
}

# Function to run test with enhanced debugging
run_test_with_debug() {
    echo "🧪 Starting test execution with enhanced debugging..."
    
    # Clear log files
    > "$LOG_FILE"
    > "$PROCESS_INFO"
    > "$GOROUTINE_DUMP"
    > "$FUSE_DEBUG"
    
    # Start container and get its name
    echo "🐳 Starting Docker container..."
    local container_id=$(docker compose -f docker/compose/docker-compose.test.yml run -d test-runner \
        bash -c "
            echo 'Enhanced debug test started at: \$(date)'
            echo 'Running: go test -v -run $TEST_PATTERN ./internal/fs -timeout ${TIMEOUT_SECONDS}s'
            echo 'Debug mode: Detailed logging enabled'
            echo ''
            
            # Enable Go race detector and verbose output
            export GODEBUG=gctrace=1
            
            # Run test with maximum verbosity
            go test -v -race -run '$TEST_PATTERN' ./internal/fs -timeout ${TIMEOUT_SECONDS}s 2>&1
            
            echo ''
            echo 'Test execution completed at: \$(date)'
        ")
    
    if [[ -z "$container_id" ]]; then
        echo "❌ Failed to start container"
        return 1
    fi
    
    echo "🐳 Container started: $container_id"
    
    # Get container name
    local container_name=$(docker ps --format "{{.Names}}" --filter "id=$container_id")
    echo "🐳 Container name: $container_name"
    
    # Stream logs to file in background
    docker logs -f "$container_id" > "$LOG_FILE" 2>&1 &
    local log_pid=$!
    
    # Monitor the container with proper timeout
    monitor_container_with_timeout "$container_id" "$container_name" &
    local monitor_pid=$!
    
    # Wait for container to complete OR timeout
    local exit_code=0
    local wait_result
    
    # Use timeout command to enforce hard limit
    if timeout $((TIMEOUT_SECONDS + 10)) docker wait "$container_id" >/dev/null 2>&1; then
        # Container completed normally
        wait_result="completed"
    else
        # Container timed out or failed
        wait_result="timeout"
        exit_code=124
        
        echo "❌ HARD TIMEOUT: Killing container after $((TIMEOUT_SECONDS + 10))s"
        docker kill "$container_id" >/dev/null 2>&1 || true
    fi
    
    # Stop monitoring
    kill $monitor_pid 2>/dev/null || true
    kill $log_pid 2>/dev/null || true
    wait $monitor_pid 2>/dev/null || true
    wait $log_pid 2>/dev/null || true
    
    # Get final container exit code if it completed normally
    if [[ "$wait_result" == "completed" ]]; then
        exit_code=$(docker inspect "$container_id" --format='{{.State.ExitCode}}' 2>/dev/null || echo "1")
    fi
    
    # Clean up container
    docker rm "$container_id" >/dev/null 2>&1 || true
    
    echo ""
    echo "📋 FINAL RESULTS:"
    echo "Exit code: $exit_code"
    echo "Wait result: $wait_result"
    echo "Debug files created in: $DEBUG_DIR"
    
    if [[ $exit_code -eq 0 ]]; then
        echo "✅ Test completed successfully"
    else
        echo "❌ Test failed or timed out"
        echo ""
        echo "📄 Last 20 lines of output:"
        tail -20 "$LOG_FILE" 2>/dev/null || echo "(no log file)"
        echo ""
        echo "🔍 Debug information available in:"
        echo "   - Test output: $LOG_FILE"
        echo "   - Process info: $PROCESS_INFO"
        echo "   - Goroutine dump: $GOROUTINE_DUMP"
        echo "   - FUSE debug: $FUSE_DEBUG"
    fi
    
    return $exit_code
}

# Main execution
echo "🔧 Setting up enhanced debugging environment..."

# Check prerequisites
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running"
    exit 1
fi

if ! command -v pstree >/dev/null 2>&1; then
    echo "⚠️  pstree not available - process tree monitoring disabled"
fi

echo "✅ Environment ready"
echo ""

# Run the test with enhanced debugging
if run_test_with_debug; then
    echo "🎉 SUCCESS: Test completed without hanging"
    exit 0
else
    echo "💥 FAILURE: Test failed or hung - debug information captured"
    exit 1
fi