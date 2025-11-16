# Webhook & Delta Loop Stabilization

**Date**: 2025-11-16  
**Type**: Feature  
**Component**: Filesystem / Delta Sync / Configuration  
**Status**: Complete

## Summary

- Default delta polling now honors the 5-minute requirement and logs any deviation, while webhook notifications immediately wake the delta loop.
- Enabling webhooks without an HTTPS `publicUrl` now fails fast, ensuring we only attempt compliant subscription endpoints.
- Subscription manager renewals recreate webhooks on failure so the filesystem automatically falls back to polling and resumes push notifications without a restart.
- Documented the behavioral changes (interval logging, webhook validation/renewal) for future Task 27 verification work.

## Changes

1. `cmd/common/config.go`, `cmd/onemount/main.go`: set the default delta interval to 300 seconds, tighten validation to require HTTPS webhook URLs, and update CLI help/logging.
2. `internal/fs/delta.go`, `internal/fs/filesystem_types.go`: add interval tracking/logging, switch the no-subscription path to the requirement-compliant default, and surface webhook-triggered delta runs in logs.
3. `internal/fs/subscription.go`: allow renewal checks to run on a configurable ticker, recreate subscriptions when renewal fails or the ID disappears, and log the lifecycle transitions; added helper API for tests.
4. Tests (`cmd/common/config_test.go`, `internal/graph/subscriptions_test.go`, `internal/fs/subscription_test.go`): cover the new defaults/validation and ensure subscription manager recovery is exercised by the Graph mock.
5. `docs/updates/index.md`: link to this implementation note so Task 27 progress has traceability.

## Testing

- `go test ./internal/fs -run Subscription -count=1`
- `go test ./internal/graph -run Subscription`
- `go test ./cmd/common` *(fails because gotk3 requires GLib ≥ 2.68 in the container; the new config tests compile but the package still depends on the GTK bindings).* 

## Rules Consulted

- AGENT-GUIDE-Coding-Standards (priority 100) – Go code keeps idiomatic comments; Python-specific docstring/type-hint mandates are not applicable here.
- AGENT-GUIDE-General-Preferences (priority 50)
- AGENT-GUIDE-Operational-Best-Practices (priority 40)
- AGENT-RULE-Documentation-Conventions (priority 20)
