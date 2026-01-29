# Design: Error Handling Recovery

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### 12. Error Handling Component

**Location**: `internal/errors/`, `internal/logging/` throughout codebase

**Verification Steps**:
1. Review error types and handling
2. Test network error scenarios in Docker
3. Test API rate limiting
4. Test crash recovery
5. Test error logging

**Expected Interfaces**:
- Custom error types in `internal/errors`
- Structured logging with zerolog in `internal/logging`
- Error context propagation and monitoring
- Error wrapping with context

**Verification Criteria**:
- Errors are logged with context
- Network errors trigger retries
- Rate limits trigger backoff
- Crashes don't corrupt state
- Error messages are user-friendly
- All tests run in Docker containers

### 16. Network Error Pattern Recognition Component

**Location**: `internal/graph/network_feedback.go`, `internal/fs/offline.go`

**Verification Steps**:
1. Review error pattern matching code
2. Test each recognized error pattern
3. Test offline state transition on pattern match
4. Test logging of detected patterns

**Recognized Error Patterns**:
- "no such host"
- "network is unreachable"
- "connection refused"
- "connection timed out"
- "dial tcp"
- "context deadline exceeded"
- "no route to host"
- "network is down"
- "temporary failure in name resolution"
- "operation timed out"

**Verification Criteria**:
- All error patterns are recognized correctly
- Offline state is triggered on pattern match
- Error patterns are logged with context
- False positives are minimized
- Pattern matching is case-insensitive where appropriate

### Error Handling Properties

**Property 35: Network Error Logging**
*For any* network error occurrence, the system should log the error with appropriate context information
**Validates: Requirements 11.1**

**Property 36: Rate Limit Backoff**
*For any* API rate limit encounter, the system should implement exponential backoff
**Validates: Requirements 11.2**
