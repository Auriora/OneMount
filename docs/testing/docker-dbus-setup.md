# D-Bus Session Bus Setup in Docker Test Environment

## Overview

This document describes how D-Bus session bus is configured in the OneMount Docker test environment to enable D-Bus integration tests.

## Problem

17 D-Bus integration tests were failing with the error:
```
dial unix /tmp/runtime-tester/bus: connect: no such file or directory
```

This occurred because while the `DBUS_SESSION_BUS_ADDRESS` environment variable was set in the Docker compose file, no D-Bus daemon was actually running in the container.

## Solution

We implemented a D-Bus session bus that starts automatically in the test container entrypoint script. The solution uses `dbus-daemon --session` to create a session bus that persists for the lifetime of the container.

### Key Components

1. **D-Bus Daemon Installation**: The test runner Dockerfile already includes `dbus` and `dbus-user-session` packages
2. **Runtime Directory**: Created at `/tmp/runtime-tester` with proper permissions (700)
3. **D-Bus Daemon Startup**: Launched in the entrypoint script before tests run
4. **Environment Variables**: `DBUS_SESSION_BUS_ADDRESS` set to point to the session bus socket

### Implementation Details

#### 1. Dockerfile (docker/images/test-runner/Dockerfile)

The Dockerfile already includes D-Bus packages:
```dockerfile
RUN apt-get update && apt-get install -y \
    dbus \
    dbus-user-session \
    ...
```

#### 2. Entrypoint Script (docker/scripts/test-entrypoint.sh)

The entrypoint script now includes a `setup_dbus()` function that:
- Creates the XDG runtime directory
- Starts a D-Bus session daemon if not already running
- Exports the `DBUS_SESSION_BUS_ADDRESS` environment variable
- Verifies the D-Bus connection

```bash
setup_dbus() {
    # Create runtime directory for D-Bus
    export XDG_RUNTIME_DIR="/tmp/runtime-tester"
    mkdir -p "$XDG_RUNTIME_DIR"
    chmod 700 "$XDG_RUNTIME_DIR"
    
    # Start D-Bus session if not already running
    if [[ -z "$DBUS_SESSION_BUS_ADDRESS" ]] || ! dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames >/dev/null 2>&1; then
        print_info "Starting D-Bus session daemon..."
        
        # Start dbus-daemon in session mode
        dbus-daemon --session --fork --address="unix:path=$XDG_RUNTIME_DIR/bus"
        
        # Set the session bus address
        export DBUS_SESSION_BUS_ADDRESS="unix:path=$XDG_RUNTIME_DIR/bus"
        
        # Wait for D-Bus to be ready
        for i in {1..10}; do
            if dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames >/dev/null 2>&1; then
                print_success "D-Bus session daemon started successfully"
                return 0
            fi
            sleep 0.5
        done
        
        print_error "Failed to start D-Bus session daemon"
        return 1
    else
        print_info "D-Bus session already running: $DBUS_SESSION_BUS_ADDRESS"
    fi
}
```

#### 3. Docker Compose (docker/compose/docker-compose.test.yml)

The compose file sets the environment variable (already present):
```yaml
environment:
  - XDG_RUNTIME_DIR=/tmp/runtime-tester
  - DBUS_SESSION_BUS_ADDRESS=unix:path=/tmp/runtime-tester/bus
```

## Testing D-Bus Setup

### Manual Verification

To verify D-Bus is working in the container:

```bash
# Start a shell in the test container
docker compose -f docker/compose/docker-compose.test.yml run --rm shell

# Inside the container, test D-Bus connection
dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames

# Expected output: List of service names on the session bus
```

### Running D-Bus Tests

To run all D-Bus integration tests:

```bash
# Run all D-Bus tests
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "TestIT_FS_DBus" ./internal/fs

# Run specific D-Bus test
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "TestIT_FS_DBus_GetFileStatus" ./internal/fs
```

## Architecture

### D-Bus Session Bus Lifecycle

```
Container Start
    ↓
Entrypoint Script Runs
    ↓
setup_environment() called
    ↓
setup_dbus() called
    ↓
Create XDG_RUNTIME_DIR (/tmp/runtime-tester)
    ↓
Start dbus-daemon --session --fork
    ↓
Export DBUS_SESSION_BUS_ADDRESS
    ↓
Verify D-Bus connection
    ↓
Tests Run (D-Bus available)
    ↓
Container Exit (D-Bus daemon terminates)
```

### D-Bus in Test Context

When tests create a D-Bus server:

1. Test creates `DBusServer` instance
2. Server connects to session bus at `$DBUS_SESSION_BUS_ADDRESS`
3. Server registers service name (e.g., `com.github.auriora.onemount.test123`)
4. Test client connects to same session bus
5. Client calls methods on the server
6. Server responds via D-Bus
7. Test verifies behavior

## Troubleshooting

### D-Bus Connection Errors

If you see errors like:
```
dial unix /tmp/runtime-tester/bus: connect: no such file or directory
```

Check:
1. Is `XDG_RUNTIME_DIR` set? `echo $XDG_RUNTIME_DIR`
2. Does the runtime directory exist? `ls -la /tmp/runtime-tester`
3. Is D-Bus daemon running? `ps aux | grep dbus-daemon`
4. Is `DBUS_SESSION_BUS_ADDRESS` set? `echo $DBUS_SESSION_BUS_ADDRESS`
5. Can you connect to D-Bus? `dbus-send --session --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames`

### Permission Errors

If you see permission errors:
```
Failed to create directory '/tmp/runtime-tester': Permission denied
```

Check:
1. Is the runtime directory owned by the test user? `ls -la /tmp/runtime-tester`
2. Does the directory have correct permissions (700)? `stat /tmp/runtime-tester`

### D-Bus Daemon Won't Start

If D-Bus daemon fails to start:
1. Check if another D-Bus daemon is already running: `ps aux | grep dbus-daemon`
2. Check if the socket file already exists: `ls -la /tmp/runtime-tester/bus`
3. Try removing stale socket: `rm -f /tmp/runtime-tester/bus`
4. Check D-Bus daemon logs: `journalctl -xe` (if systemd is available)

## Alternatives Considered

### Option 1: dbus-run-session Wrapper

We could wrap test execution with `dbus-run-session`:
```bash
dbus-run-session -- go test -v ./...
```

**Pros:**
- Simple, one-line solution
- Automatically cleans up session

**Cons:**
- Creates new session for each test run
- Slower startup time
- Harder to debug session issues

### Option 2: Mock D-Bus

We could mock D-Bus for integration tests:
```go
type MockDBusConn struct {
    // Mock implementation
}
```

**Pros:**
- No external dependencies
- Faster test execution
- More control over test scenarios

**Cons:**
- Not testing real D-Bus behavior
- Misses integration issues
- More test code to maintain

### Option 3: Host D-Bus Socket

We could mount the host's D-Bus socket:
```yaml
volumes:
  - /run/user/1000/bus:/tmp/runtime-tester/bus:ro
```

**Pros:**
- Uses real system D-Bus
- No daemon management needed

**Cons:**
- Requires host D-Bus to be running
- Security concerns (container accessing host D-Bus)
- Not portable across different host configurations
- Tests could interfere with host D-Bus services

### Selected Solution: Session Daemon in Container

We chose to start a D-Bus session daemon in the container because:
- ✅ Tests real D-Bus behavior
- ✅ Isolated from host system
- ✅ Portable across different environments
- ✅ Easy to debug and troubleshoot
- ✅ Matches production behavior

## References

- [D-Bus Specification](https://dbus.freedesktop.org/doc/dbus-specification.html)
- [D-Bus Tutorial](https://dbus.freedesktop.org/doc/dbus-tutorial.html)
- [godbus Documentation](https://pkg.go.dev/github.com/godbus/dbus/v5)
- [OneMount D-Bus Integration](../../docs/2-architecture/dbus-integration.md)

## Related Files

- `docker/images/test-runner/Dockerfile` - D-Bus package installation
- `docker/scripts/test-entrypoint.sh` - D-Bus daemon startup
- `docker/compose/docker-compose.test.yml` - Environment configuration
- `internal/fs/dbus.go` - D-Bus server implementation
- `internal/fs/dbus_test.go` - D-Bus integration tests
