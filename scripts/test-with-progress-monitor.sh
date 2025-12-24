#!/bin/bash

# Test runner with external progress monitoring
# Usage: ./scripts/test-with-progress-monitor.sh <test-pattern> <timeout-seconds>

set -e

TEST_PATTERN="${1:-TestIT_FS_ETag}"
TIMEOUT_SECONDS="${2:-60}"
LOG_FILE="/tmp/test-progress-monitor.log"

echo "🚀 Starting test with progress monitoring..."
echo "Test pattern: $TEST_PATTERN"
echo "Timeout: ${TIMEOUT_SECONDS}s"
echo "Log file: $LOG_FILE"
echo ""

# Function to monitor test progress
monitor_progress() {
    local test_pid=$1
    local start_time=$(date +%s)
    local last_output_time=$start_time
    local heartbeat_interval=5
    
    echo "📊 Progress monitor started (PID: $$)"
    echo "🎯 Monitoring test process (PID: $test_pid)"
    echo ""
    
    while kill -0 $test_pid 2>/dev/null; do
        local current_time=$(date +%s)
        local elapsed=$((current_time - start_time))
        local since_output=$((current_time - last_output_time))
        
        # Check if we have new output
        if [[ -f "$LOG_FILE" ]]; then
            local log_size=$(wc -l < "$LOG_FILE" 2>/dev/null || echo "0")
            if [[ $log_size -gt ${last_log_size:-0} ]]; then
                last_output_time=$current_time
                last_log_size=$log_size
                echo "📝 New output detected (${log_size} lines total)"
            fi
        fi
        
        # Show heartbeat
        if [[ $((elapsed % heartbeat_interval)) -eq 0 ]]; then
            echo "⏱️  HEARTBEAT: Test running for ${elapsed}s | No output for ${since_output}s"
            
            # Show what the test process is doing
            if command -v pstree >/dev/null 2>&1; then
                echo "🔍 Process tree:"
                pstree -p $test_pid 2>/dev/null | head -5 || echo "   (pstree failed)"
            fi
            
            # Check if test is stuck (no output for too long)
            if [[ $since_output -gt 30 ]]; then
                echo "⚠️  WARNING: No output for ${since_output}s - test may be hung"
            fi
            
            echo ""
        fi
        
        # Check timeout
        if [[ $elapsed -gt $TIMEOUT_SECONDS ]]; then
            echo "❌ TIMEOUT: Test exceeded ${TIMEOUT_SECONDS}s limit"
            echo "🔪 Killing test process..."
            kill -TERM $test_pid 2>/dev/null || true
            sleep 2
            kill -KILL $test_pid 2>/dev/null || true
            return 1
        fi
        
        sleep 1
    done
    
    local final_time=$(date +%s)
    local total_elapsed=$((final_time - start_time))
    echo "✅ Test process completed after ${total_elapsed}s"
    return 0
}

# Function to run the test
run_test() {
    echo "🧪 Starting test execution..."
    
    # Clear log file
    > "$LOG_FILE"
    
    # Run test in Docker with output to log file
    docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner \
        bash -c "
            echo 'Test started at: \$(date)'
            echo 'Running: go test -v -run $TEST_PATTERN ./internal/fs -timeout ${TIMEOUT_SECONDS}s'
            echo ''
            go test -v -run '$TEST_PATTERN' ./internal/fs -timeout ${TIMEOUT_SECONDS}s 2>&1
        " | tee "$LOG_FILE" &
    
    local test_pid=$!
    echo "🎯 Test process started (PID: $test_pid)"
    echo ""
    
    # Start progress monitor in background
    monitor_progress $test_pid &
    local monitor_pid=$!
    
    # Wait for test to complete
    local exit_code=0
    wait $test_pid || exit_code=$?
    
    # Stop monitor
    kill $monitor_pid 2>/dev/null || true
    wait $monitor_pid 2>/dev/null || true
    
    echo ""
    echo "📋 FINAL RESULTS:"
    echo "Exit code: $exit_code"
    
    if [[ $exit_code -eq 0 ]]; then
        echo "✅ Test completed successfully"
    else
        echo "❌ Test failed or timed out"
        echo ""
        echo "📄 Last 20 lines of output:"
        tail -20 "$LOG_FILE" 2>/dev/null || echo "(no log file)"
    fi
    
    return $exit_code
}

# Main execution
echo "🔧 Setting up test environment..."

# Ensure log directory exists
mkdir -p "$(dirname "$LOG_FILE")"

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running"
    exit 1
fi

echo "✅ Environment ready"
echo ""

# Run the test with monitoring
if run_test; then
    echo "🎉 SUCCESS: Test completed without hanging"
    exit 0
else
    echo "💥 FAILURE: Test failed or hung"
    exit 1
fi