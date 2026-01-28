# Tasks: Authentication and Account Management

## Overview

Implementation tasks for authentication and account management functionality.

## Phase 1: Investigation & Prototyping (1-2 days)

- [x] 4.9.1 Analyze current token storage usage
  - Analyze current token storage usage across codebase
  - Prototype account-based storage functions in `internal/graph/oauth2.go`
  - Test hash generation and collision resistance
  - Verify Docker environment behavior with new approach
  - Document findings and edge cases
  - _Requirements: 1.6_

## Phase 2: Core Implementation (2-3 days)

- [x] 4.9.2 Implement account-based token storage
  - Add `GetAuthTokensPathByAccount()` function to `internal/graph/oauth2.go`
  - Add `hashAccount()` helper function
  - Implement `FindAuthTokens()` with fallback logic and auto-migration
  - Add automatic token migration from old locations
  - Update `SaveAuthTokens()` to use account-based path
  - Update `LoadAuthTokens()` to search multiple locations
  - Add unit tests for new functions
  - Add migration logic tests
  - _Requirements: 1.6_

## Phase 3: Integration & Testing (2-3 days)

- [x] 4.9.3 Integration and testing
  - Update `cmd/onemount/main.go` to use account-based paths
  - Update test fixtures in `internal/testutil/helpers/` to use account-based paths
  - Update Docker auth setup scripts
  - Add integration tests for token migration
  - Test Docker environment token access with new paths
  - Test multiple account scenarios
  - Verify backward compatibility with old token locations
  - _Requirements: 1.6, 13.2, 13.4_

## Phase 4: Documentation & Cleanup (1 day)

- [x] 4.9.4 Documentation and cleanup
  - Update `docs/2-architecture/authentication.md` with new token storage architecture
  - Add migration guide in `docs/guides/user/`
  - Update test documentation in `docs/4-testing/`
  - Add deprecation warnings for old token paths in logs
  - Update CHANGELOG with migration notes
  - _Requirements: 1.6, 12.1_

## Phase 5: Authentication Flow Testing

- [x] 4.1 Review OAuth2 code structure
  - Read and analyze `internal/graph/oauth2.go`, `oauth2_gtk.go`, `oauth2_headless.go`
  - Review `internal/graph/authenticator.go` interface and implementations
  - Compare implementation against design document
  - Document any deviations from architecture
  - _Requirements: 1.1, 1.5_

- [x] 4.2 Test interactive authentication flow
  - Use Docker shell for interactive testing
  - Launch OneMount with GUI authentication (if GTK available)
  - Complete Microsoft OAuth2 flow
  - Verify tokens are stored correctly
  - Check file permissions on token storage
  - _Requirements: 1.1, 1.2_

- [x] 4.3 Test token refresh mechanism
  - Manually expire access token
  - Trigger operation requiring authentication
  - Verify automatic token refresh occurs
  - Check that new tokens are persisted
  - _Requirements: 1.3_

- [x] 4.4 Test authentication failure scenarios
  - Test with invalid credentials
  - Test with network disconnection during auth
  - Test with expired refresh token
  - Verify error messages are clear and actionable
  - _Requirements: 1.4_

- [x] 4.5 Test headless authentication
  - Run OneMount in headless mode
  - Verify device code flow is used
  - Complete authentication via browser
  - Verify tokens are stored correctly
  - _Requirements: 1.5_

## Phase 6: Integration Tests

- [x] 4.6 Create authentication integration tests
  - Write test for complete OAuth2 flow with mock server
  - Write test for token refresh with mock responses
  - Write test for authentication failure scenarios
  - Run tests in Docker
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

## Phase 7: Property-Based Tests

- [x] 4.7.1 Property 1: OAuth2 Token Storage Security
  - Generate random valid OAuth2 completions
  - Verify tokens stored with proper security attributes
  - Run 100+ iterations per property test
  - _Requirements: 1.2_

- [x] 4.7.2 Property 2: Automatic Token Refresh
  - Generate random expired tokens with valid refresh tokens
  - Verify automatic refresh occurs without user intervention
  - Test with various expiration scenarios
  - _Requirements: 1.3_

- [x] 4.7.3 Property 3: Re-authentication on Refresh Failure
  - Generate random token refresh failure scenarios
  - Verify user re-authentication prompt occurs
  - Test with various failure types
  - _Requirements: 1.4_

- [x] 4.7.4 Property 4: Headless Authentication Method
  - Generate random headless system configurations
  - Verify device code flow is used
  - Test with various headless scenarios
  - _Requirements: 1.5_

## Phase 8: Multi-Account Testing

- [ ] Test multiple account mounting
  - Mount personal OneDrive account
  - Mount work OneDrive account
  - Mount shared drive
  - Verify separate tokens for each account
  - Verify separate caches for each account
  - Verify separate delta sync loops
  - _Requirements: 14.1-14.8_

## Phase 9: Security Testing

- [ ] Test token encryption
  - Verify AES-256 encryption is used
  - Test file permissions (0600)
  - Verify TLS 1.2+ for API calls
  - Test rate limiting on auth failures
  - _Requirements: 22.1-22.7_

## Estimated Total Effort

- Phase 1: 1-2 days
- Phase 2: 2-3 days
- Phase 3: 2-3 days
- Phase 4: 1 day
- Phase 5-7: Already completed
- Phase 8-9: 2-3 days

**Total**: 8-12 days

## Success Criteria

- [ ] All authentication tests pass in Docker
- [ ] Token migration works automatically
- [ ] Multiple accounts can be mounted simultaneously
- [ ] Tokens are stored securely with proper encryption
- [ ] Backward compatibility maintained
- [ ] Documentation updated
