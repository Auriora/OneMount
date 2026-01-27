# Mount-Account Registry Integration

**Date**: 2026-01-27  
**Status**: Complete  
**Related Design**: `docs/designs/mount-account-registry.md`  
**Related Issue**: `docs/issues/account-storage-incomplete-migration.md`

## Problem

The account-based storage system had a fundamental chicken-and-egg problem:
- Need account name to find tokens
- Need tokens to get account name

This resulted in the launcher showing errors about missing tokens and not being able to determine the account for a mount point.

## Solution

Implemented a mount-account registry system that provides an independent mapping from mount points to OneDrive accounts.

### Registry Implementation

Created `internal/config/mounts.go` with:
- Thread-safe registry using `sync.RWMutex`
- Persistent JSON storage at `~/.config/onemount/mounts.json`
- CRUD operations for mount-account mappings
- Comprehensive unit tests (all passing)

### Integration Points

1. **Launcher (`cmd/onemount-launcher/main.go`)**:
   - Loads registry to display account names for mounts
   - Shows account email even when filesystem is not mounted
   - Falls back to drive name or path if account not available

2. **Main Binary (`cmd/onemount/main.go`)**:
   - Registers mount after successful authentication
   - Works in all modes: normal mount, auth-only, stats
   - Logs registration success/failure

3. **UI Package (`internal/ui/onemount.go`)**:
   - `GetKnownMounts()` now uses registry instead of scanning cache directories
   - Returns list of registered mount points

## Files Modified

- `cmd/onemount-launcher/main.go` - Added registry integration with import alias
- `cmd/onemount/main.go` - Added registry registration after authentication
- `internal/ui/onemount.go` - Updated to use registry for known mounts
- `internal/config/mounts.go` - Registry implementation (already existed)
- `internal/config/mounts_test.go` - Unit tests (already existed)

## Import Alias Pattern

Both binaries use an import alias to avoid shadowing:
```go
import mountconfig "github.com/auriora/onemount/internal/config"
```

This is necessary because both have function parameters named `config` of type `*common.Config`.

## Correct Flow

1. **First Mount (No Registry Entry)**:
   - User creates mount point
   - System authenticates → gets account
   - System registers mount → account mapping
   - System saves tokens by account hash
   - Launcher can now display account name

2. **Subsequent Mounts (Registry Entry Exists)**:
   - Launcher loads registry
   - Displays account name from registry
   - On mount: system looks for tokens by account hash
   - If tokens missing: prompts for authentication

## Testing

- All unit tests pass: `go test ./internal/config`
- Both binaries build successfully
- Registry operations are thread-safe and persistent

## Next Steps

1. Test full flow: create mount → authenticate → verify registry → verify launcher display
2. Implement migration function to populate registry from existing tokens
3. Add cleanup: remove registry entries when mounts are deleted

## Rules Applied

- **coding-standards.md** (Priority 100): DRY principle, error handling, thread safety
- **general-preferences.md** (Priority 50): Direct implementation, SOLID principles
- **operational-best-practices.md** (Priority 40): Tool-driven exploration, minimal edits
