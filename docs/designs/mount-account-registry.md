# Mount-Account Registry Design

**Date**: 2026-01-27  
**Status**: Proposed  
**Priority**: CRITICAL

## Problem Statement

The current account-based token storage has a fundamental chicken-and-egg problem:
- To find tokens by account, you need the account name
- To get the account name, you need to read the token file  
- But you don't know which token file without the account name

This makes it impossible to:
1. Display account names in the launcher before authentication
2. Determine if a mount point needs authentication
3. Support multiple mounts for the same account
4. Properly prompt for authentication when needed

## Solution: Mount-Account Registry

### Architecture

```
~/.config/onemount/mounts.json  ← Mount point → Account mapping
~/.cache/onemount/accounts/{hash}/auth_tokens.json  ← Tokens by account
```

### Registry File Format

```json
{
  "mounts": {
    "/home/user/OneDrive": {
      "account": "user@example.com",
      "created": "2026-01-27T12:00:00Z",
      "updated": "2026-01-27T13:00:00Z",
      "label": "Personal OneDrive"
    },
    "/home/user/Work": {
      "account": "user@company.com",
      "created": "2026-01-27T14:00:00Z",
      "label": "Work Account"
    }
  }
}
```

### Lookup Flow

```
1. User creates mount at /home/user/OneDrive
2. Check registry: Does this mount point exist?
   
   NO → Prompt for authentication
        ↓
        Get account from OAuth response
        ↓
        Save to registry: mount → account
        ↓
        Save tokens: accounts/{hash}/auth_tokens.json
   
   YES → Get account from registry
         ↓
         Look for tokens: accounts/{hash}/auth_tokens.json
         ↓
         Tokens exist? → Use them (refresh if needed)
         Tokens missing? → Prompt for authentication
```

### Benefits

1. **No Chicken-and-Egg Problem**: Registry provides account name independently
2. **Clear Authentication State**: Know if mount needs auth before trying
3. **Multiple Mounts Per Account**: Same account can have multiple mount points
4. **User-Friendly Labels**: Optional labels for better UX
5. **Audit Trail**: Created/updated timestamps for troubleshooting

### API

```go
// Create registry
registry, err := config.NewMountsRegistry(configDir)

// Register new mount
err = registry.SetMount("/home/user/OneDrive", "user@example.com", "Personal")

// Look up account for mount
account, exists := registry.GetAccount("/home/user/OneDrive")

// Get all mounts for an account
mounts := registry.GetMountsByAccount("user@example.com")

// Remove mount
err = registry.RemoveMount("/home/user/OneDrive")
```

### Integration Points

#### 1. Main Binary (cmd/onemount/main.go)

```go
// Load registry
registry, err := config.NewMountsRegistry(config.ConfigDir)

// Check if mount is registered
account, exists := registry.GetAccount(mountPoint)
if !exists {
    // New mount - authenticate and register
    auth, err := graph.Authenticate(...)
    registry.SetMount(mountPoint, auth.Account, "")
} else {
    // Existing mount - find tokens by account
    tokenPath := graph.GetAuthTokensPathByAccount(cacheDir, account)
    if _, err := os.Stat(tokenPath); err != nil {
        // Tokens missing - re-authenticate
        auth, err := graph.Authenticate(...)
    }
}
```

#### 2. Launcher (cmd/onemount-launcher/main.go)

```go
// Load registry
registry, err := config.NewMountsRegistry(config.ConfigDir)

// Get account for display
account, exists := registry.GetAccount(mountPoint)
if !exists {
    // Show "Not configured" or prompt to set up
    label.SetText("Click to configure")
} else {
    // Show account name
    label.SetText(account)
}
```

#### 3. GetKnownMounts (internal/ui/onemount.go)

```go
func GetKnownMounts(configDir string) []string {
    registry, err := config.NewMountsRegistry(configDir)
    if err != nil {
        return []string{}
    }
    return registry.ListMounts()
}
```

### Migration Strategy

1. **Phase 1**: Implement registry (this design)
2. **Phase 2**: Update main binary to use registry
3. **Phase 3**: Update launcher to use registry
4. **Phase 4**: Migrate existing mounts
   - Scan old token locations
   - Extract account from tokens
   - Populate registry
5. **Phase 5**: Remove old path-based lookup code

### Migration Code

```go
func MigrateToRegistry(cacheDir, configDir string) error {
    registry, err := config.NewMountsRegistry(configDir)
    if err != nil {
        return err
    }

    // Scan cache directory for old instance-based tokens
    dirents, err := os.ReadDir(cacheDir)
    if err != nil {
        return err
    }

    for _, dirent := range dirents {
        if !dirent.IsDir() || dirent.Name() == "accounts" {
            continue
        }

        // Try to load tokens from this directory
        tokenPath := filepath.Join(cacheDir, dirent.Name(), "auth_tokens.json")
        auth, err := graph.LoadAuthTokens(tokenPath)
        if err != nil {
            continue
        }

        // Unescape the mount point name
        mountPoint := unit.UnitNamePathUnescape(dirent.Name())
        
        // Register in new system
        registry.SetMount(mountPoint, auth.Account, "")
        
        // Migrate tokens to account-based location
        newPath := graph.GetAuthTokensPathByAccount(cacheDir, auth.Account)
        graph.MigrateTokens(tokenPath, newPath)
    }

    return nil
}
```

### Error Handling

- **Registry file corrupt**: Start with empty registry, log warning
- **Mount point not in registry**: Treat as new mount, prompt for auth
- **Account in registry but no tokens**: Prompt for re-authentication
- **Multiple mounts for same account**: Supported, share tokens

### Security Considerations

- Registry file permissions: 0600 (owner read/write only)
- Registry location: `~/.config/onemount/` (user-specific)
- No sensitive data in registry (account email is not secret)
- Tokens remain in separate cache directory with proper permissions

### Testing Requirements

1. Unit tests for registry operations (DONE)
2. Integration test: Create mount → Authenticate → Verify registry
3. Integration test: Restart → Load registry → Find tokens
4. Integration test: Multiple mounts, same account
5. Migration test: Old tokens → New registry + account-based tokens

### Documentation Updates

- User guide: How mounts are configured
- Developer guide: Registry API usage
- Troubleshooting: Registry file location and format
- Migration guide: Moving from old to new system

## Implementation Checklist

- [x] Create `internal/config/mounts.go` with registry implementation
- [x] Create unit tests for registry
- [ ] Update `cmd/onemount/main.go` to use registry
- [ ] Update `cmd/onemount-launcher/main.go` to use registry
- [ ] Update `internal/ui/onemount.go` GetKnownMounts()
- [ ] Implement migration function
- [ ] Add integration tests
- [ ] Update documentation
- [ ] Test with real OneDrive accounts

## Related Issues

- `docs/issues/account-storage-incomplete-migration.md` - Root cause analysis
- Account-based storage was implemented but lookup was incomplete
- This design completes the architecture properly
