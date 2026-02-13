# Scripts Directory Cleanup Analysis

**Date**: 2026-02-13  
**Task**: Review scripts/ for duplicates, redundancies, and one-off scripts

## Executive Summary

The `scripts/` directory contains 33 shell scripts and a Python CLI framework. Analysis reveals:
- **7 one-off/obsolete scripts** that can be removed
- **6 duplicate/redundant scripts** that can be consolidated
- **4 test-specific scripts** that should be moved to tests/
- **16 active scripts** that should be retained

## Detailed Analysis

### 1. One-Off / Obsolete Scripts (REMOVE)

These scripts were created for specific tasks and are no longer needed:

1. **`cleanup-old-auth-scripts.sh`** - One-time cleanup script
   - Purpose: Remove old auth scripts (already completed)
   - Status: Task completed, script no longer needed
   - Action: DELETE

2. **`cleanup-old-runners.sh`** - One-time cleanup script
   - Purpose: Remove specific offline GitHub runners (IDs 39, 40)
   - Status: Task completed, hardcoded runner IDs
   - Action: DELETE

3. **`label-unlabeled-tests.sh`** - One-time migration script
   - Purpose: Add test prefixes (Task 46.1.2 Part 1)
   - Status: Task completed
   - Action: DELETE

4. **`label-remaining-tests.sh`** - One-time migration script
   - Purpose: Add test prefixes (Task 46.1.2 Part 2)
   - Status: Task completed
   - Action: DELETE

5. **`label-final-tests.sh`** - One-time migration script
   - Purpose: Add test prefixes (Task 46.1.2 Part 3)
   - Status: Task completed
   - Action: DELETE

6. **`update-auth-token-paths.sh`** - One-time migration script
   - Purpose: Update test files to use centralized auth path
   - Status: Migration completed
   - Action: DELETE

7. **`host-sync-codex-config.sh`** - Development-specific script
   - Purpose: Sync Codex configuration (development tool)
   - Status: Tool-specific, not project-related
   - Action: DELETE or move to personal scripts

### 2. Duplicate / Redundant Scripts (CONSOLIDATE)

These scripts have overlapping functionality:

#### A. Mount Timeout Scripts (3 scripts → 1 script)

1. **`debug-mount-timeout.sh`** - Diagnostic tool
2. **`fix-mount-timeout.sh`** - Fix implementation
3. **`test-mount-timeout-fix.sh`** - Validation test

**Recommendation**: Consolidate into single `mount-timeout-tools.sh` with subcommands:
```bash
./scripts/mount-timeout-tools.sh diagnose
./scripts/mount-timeout-tools.sh fix
./scripts/mount-timeout-tools.sh test
```

#### B. Remote Docker Deployment Scripts (2 scripts → 1 script)

1. **`deploy-docker-remote.sh`** - Full-featured deployment
2. **`deploy-optimized-remote.sh`** - Optimized version

**Recommendation**: Merge into `deploy-docker-remote.sh` with `--optimized` flag

#### C. Runner Management Scripts (2 scripts → 1 script)

1. **`manage-runner.sh`** - Single runner management
2. **`manage-runners.sh`** - Multiple runners management

**Recommendation**: Merge into `manage-runners.sh` (handles both single and multiple)

### 3. Test-Specific Scripts (MOVE to tests/)

These scripts are test utilities and should live with tests:

1. **`test-task-5.4-filesystem-operations.sh`** → `tests/system/task-5.4-filesystem-operations.sh`
2. **`test-task-5.5-unmounting-cleanup.sh`** → `tests/system/task-5.5-unmounting-cleanup.sh`
3. **`test-task-5.6-signal-handling.sh`** → `tests/system/task-5.6-signal-handling.sh`
4. **`test-cache-management.sh`** → `tests/system/cache-management.sh`

### 4. Scripts to Retain (16 active scripts)

#### Core Development Tools
- **`dev`** / **`dev.py`** - Main CLI entry point (KEEP)
- **`install-completion.sh`** - Shell completion installer (KEEP)
- **`install-codex-cli.sh`** - Codex CLI installer (KEEP)

#### Testing Infrastructure
- **`timeout-test-wrapper.sh`** - Critical for hanging test prevention (KEEP)
- **`debug-hanging-tests.sh`** - Enhanced debugging for FUSE tests (KEEP)
- **`test-with-progress-monitor.sh`** - Progress monitoring (KEEP)
- **`audit-integration-tests.sh`** - Test audit tool (KEEP)

#### Authentication
- **`setup-auth-reference.sh`** - Reference-based auth setup (KEEP)
- **`manual-auth.sh`** - Manual authentication (KEEP)
- **`interactive-auth.sh`** - Interactive authentication (KEEP)
- **`verify-docker-auth.sh`** - Auth verification (KEEP)

#### Docker & Deployment
- **`test-docker-fixes.sh`** - Docker validation (KEEP)
- **`fix-remote-docker.sh`** - Remote Docker diagnostics (KEEP)
- **`test-x11-forwarding.sh`** - X11 testing (KEEP)

#### CI/CD
- **`setup-github-secrets.sh`** - GitHub secrets setup (KEEP)
- **`validate-workflows.sh`** - Workflow validation (KEEP)

#### Utilities
- **`manage-workspace.sh`** - Workspace management (KEEP)
- **`cgo-helper.sh`** - CGO build helper (KEEP)
- **`curl-graph.sh`** - Graph API testing (KEEP)
- **`detect-*.sh`** (3 scripts) - Build tag detection (KEEP)

### 5. Python CLI Framework (KEEP)

The modern Python CLI framework is well-organized:
```
scripts/
├── dev.py                    # Main entry point
├── commands/                 # Command modules (8 files)
├── utils/                    # Shared utilities (10 files)
└── requirements-dev-cli.txt  # Dependencies
```

**Status**: Well-structured, actively maintained, KEEP ALL

## Recommended Actions

### Phase 1: Remove Obsolete Scripts (Immediate)
```bash
# Remove one-off scripts
rm scripts/cleanup-old-auth-scripts.sh
rm scripts/cleanup-old-runners.sh
rm scripts/label-unlabeled-tests.sh
rm scripts/label-remaining-tests.sh
rm scripts/label-final-tests.sh
rm scripts/update-auth-token-paths.sh
rm scripts/host-sync-codex-config.sh
```

### Phase 2: Consolidate Duplicates (Next Sprint)

1. **Mount Timeout Tools**
   ```bash
   # Create consolidated script
   cat > scripts/mount-timeout-tools.sh << 'EOF'
   #!/bin/bash
   # Consolidated mount timeout diagnostic and fix tool
   case "$1" in
     diagnose) # debug-mount-timeout.sh logic
     fix)      # fix-mount-timeout.sh logic
     test)     # test-mount-timeout-fix.sh logic
   esac
   EOF
   
   # Remove old scripts
   rm scripts/debug-mount-timeout.sh
   rm scripts/fix-mount-timeout.sh
   rm scripts/test-mount-timeout-fix.sh
   ```

2. **Remote Docker Deployment**
   ```bash
   # Merge deploy-optimized-remote.sh into deploy-docker-remote.sh
   # Add --optimized flag to deploy-docker-remote.sh
   rm scripts/deploy-optimized-remote.sh
   ```

3. **Runner Management**
   ```bash
   # Merge manage-runner.sh into manage-runners.sh
   rm scripts/manage-runner.sh
   ```

### Phase 3: Reorganize Test Scripts (Next Sprint)

```bash
# Move test scripts to tests/system/
mkdir -p tests/system
mv scripts/test-task-5.4-filesystem-operations.sh tests/system/
mv scripts/test-task-5.5-unmounting-cleanup.sh tests/system/
mv scripts/test-task-5.6-signal-handling.sh tests/system/
mv scripts/test-cache-management.sh tests/system/

# Update references in documentation
```

## Impact Analysis

### Before Cleanup
- Total scripts: 33 shell scripts + Python CLI
- Maintenance burden: HIGH
- Duplication: 6 scripts
- Obsolete: 7 scripts

### After Cleanup
- Total scripts: 16 shell scripts + Python CLI
- Maintenance burden: MEDIUM
- Duplication: 0 scripts
- Obsolete: 0 scripts

### Benefits
1. **Reduced Complexity**: 51% reduction in shell scripts (33 → 16)
2. **Improved Maintainability**: No duplicate functionality
3. **Better Organization**: Test scripts with tests
4. **Clearer Purpose**: Each script has a clear, unique role

## Migration Checklist

- [ ] Phase 1: Remove 7 obsolete scripts
- [ ] Update documentation references
- [ ] Phase 2: Consolidate mount timeout scripts
- [ ] Phase 2: Merge remote deployment scripts
- [ ] Phase 2: Merge runner management scripts
- [ ] Phase 3: Move test scripts to tests/
- [ ] Update CI/CD workflows if needed
- [ ] Update README.md and docs/
- [ ] Test all consolidated scripts
- [ ] Create git commit with detailed changelog

## Notes

- All changes should be made in a feature branch
- Test each consolidation before removing old scripts
- Update any CI/CD workflows that reference moved scripts
- Document new consolidated script interfaces
- Consider adding deprecation warnings before removal

## References

- Testing conventions: `.kiro/steering/testing-conventions.md`
- Script documentation: `scripts/README.md`
- Development CLI: `scripts/dev.py`
