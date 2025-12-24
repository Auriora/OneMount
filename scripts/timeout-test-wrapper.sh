#!/bin/bash

# Simple timeout wrapper for hanging tests
# Usage: ./scripts/timeout-test-wrapper.sh <test-pattern> <timeout-seconds>

set -e

TEST_PATTERN="${1:-TestIT_FS_ETag_01_CacheValidationWithIfNoneMatch_Fixed}"
TIMEOUT_SECONDS="${2:-60}"
DEBUG_DIR="test-artifacts/debug"
LOG_FILE="$DEBUG_DIR/timeout-test.log"

echo "⏰ Timeout Test Wrapper"
echo "======================"
echo "Test pattern: $TEST_PATTERN"
echo "Timeout: ${TIMEOUT_SECONDS}s"
echo "Log file: $LOG_FILE"
echo ""

# Setup debug directory
mkdir -p "$DEBUG_DIR"
rm -f "$LOG_FILE"

# Function to kill container by pattern
kill_test_containers() {
    echo "🔪 Killing any running test containers..."
    
    # Find and kill containers with test runner pattern
    local containers=$(docker ps -q --filter "name=test-runner" 2>/dev/null || true)
    if [[ -n "$containers" ]]; then
        echo "   Found containers: $containers"
        docker kill $containers 2>/dev/null || true
        docker rm $containers 2>/dev/null || true
    fi
    
    # Also kill by image pattern
    containers=$(docker ps -q --filter "ancestor=onemount-test-runner" 2>/dev/null || true)
    if [[ -n "$containers" ]]; then
        echo "   Found containers by image: $containers"
        docker kill $containers 2>/dev/null || true
        docker rm $containers 2>/dev/null || true
    fi
}

# Cleanup function
cleanup() {
    echo ""
    echo "🧹 Cleaning up..."
    kill_test_containers
    
    # Kill any background processes
    jobs -p | xargs -r kill 2>/dev/null || true
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

echo "🧪 Starting test with hard timeout..."

# Start the test in background
docker compose -f docker/compose/docker-compose.test.yml \
    $(test -f docker/compose/docker-compose.auth.yml && echo "-f docker/compose/docker-compose.auth.yml") \
    run --rm test-runner \
    bash -c "
        echo 'Test started at: \$(date)'
        echo 'Running: go test -v -run $TEST_PATTERN ./internal/fs -timeout $((TIMEOUT_SECONDS - 10))s'
        echo ''
        
        # Run test with slightly shorter timeout than wrapper
        go test -v -run '$TEST_PATTERN' ./internal/fs -timeout $((TIMEOUT_SECONDS - 10))s 2>&1
        
        echo ''
        echo 'Test completed at: \$(date)'
    " 2>&1 | tee "$LOG_FILE" &

# Get the background job PID
TEST_JOB_PID=$!

echo "🎯 Test job started (PID: $TEST_JOB_PID)"
echo "⏰ Will timeout after ${TIMEOUT_SECONDS}s"
echo ""

# Monitor the test with heartbeat
start_time=$(date +%s)
last_heartbeat=0
last_lines=0

while kill -0 $TEST_JOB_PID 2>/dev/null; do
    current_time=$(date +%s)
    elapsed=$((current_time - start_time))
    
    # Check if we've exceeded our timeout
    if [[ $elapsed -ge $TIMEOUT_SECONDS ]]; then
        echo "⏰ TIMEOUT: Killing test after ${TIMEOUT_SECONDS}s"
        kill $TEST_JOB_PID 2>/dev/null || true
        sleep 2
        kill -9 $TEST_JOB_PID 2>/dev/null || true
        break
    fi
    
    # Show heartbeat every 10 seconds
    if [[ $((elapsed - last_heartbeat)) -ge 10 ]]; then
        remaining=$((TIMEOUT_SECONDS - elapsed))
        echo "⏱️  HEARTBEAT: Test running for ${elapsed}s (timeout in ${remaining}s)"
        last_heartbeat=$elapsed
        
        # Show recent output
        if [[ -f "$LOG_FILE" ]]; then
            lines=$(wc -l < "$LOG_FILE" 2>/dev/null || echo "0")
            echo "   Log file has $lines lines"
            
            # Show last few lines if there's new content
            if [[ $lines -gt $last_lines ]]; then
                echo "   Recent output:"
                tail -2 "$LOG_FILE" 2>/dev/null | sed 's/^/   | /' || echo "   (no recent output)"
                last_lines=$lines
            else
                # Check if test has actually failed/panicked
                if grep -q "panic:\|FAIL\|exit status" "$LOG_FILE" 2>/dev/null; then
                    echo "   ⚠️  Test appears to have failed - checking if process is still alive..."
                    # Give it a moment to clean up
                    sleep 3
                    if ! kill -0 $TEST_JOB_PID 2>/dev/null; then
                        echo "   ✅ Process has terminated"
                        break
                    fi
                fi
            fi
        fi
        echo ""
    fi
    
    sleep 2
done

# Wait for the job to complete and get exit code
wait $TEST_JOB_PID 2>/dev/null
exit_code=$?

current_time=$(date +%s)
total_elapsed=$((current_time - start_time))

echo ""
echo "📋 FINAL RESULTS:"
echo "Total time: ${total_elapsed}s"
echo "Exit code: $exit_code"

if [[ $exit_code -eq 0 ]]; then
    echo "✅ Test completed successfully"
elif [[ $total_elapsed -ge $TIMEOUT_SECONDS ]]; then
    echo "❌ Test timed out after ${TIMEOUT_SECONDS}s"
    echo ""
    echo "🔍 This confirms the test is hanging - it exceeded the timeout"
    echo "📄 Last 10 lines of output:"
    tail -10 "$LOG_FILE" 2>/dev/null || echo "(no log file)"
elif [[ $exit_code -eq 130 ]]; then
    echo "⚠️  Test was interrupted (Ctrl+C)"
else
    echo "❌ Test failed with exit code: $exit_code"
    echo ""
    echo "📄 Last 10 lines of output:"
    tail -10 "$LOG_FILE" 2>/dev/null || echo "(no log file)"
fi

echo ""
echo "📁 Full log available at: $LOG_FILE"

exit $exit_code