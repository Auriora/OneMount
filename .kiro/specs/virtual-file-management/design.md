# Design: Virtual File Management

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### 2B. Virtual File Management Component

**Location**: `cmd/common/xdg.go`, `internal/fs/virtual_files.go`

**Verification Steps**:
1. Review virtual file creation and serving
2. Test `.xdg-volume-info` immediate availability
3. Test virtual file persistence with `local-*` identifiers
4. Test overlay policy resolution
5. Test virtual file exclusion from sync operations

**Expected Interfaces**:
- `CreateVirtualFile()` method for virtual file creation
- `ServeVirtualFile()` method for immediate serving
- `SetOverlayPolicy()` for precedence management
- Virtual file metadata with `is_virtual=TRUE` flag

**Verification Criteria**:
- `.xdg-volume-info` available immediately on mount
- Virtual files bypass Graph API lookups
- Virtual files persist with `local-*` identifiers
- Overlay policies resolve conflicts correctly
- Virtual files excluded from upload/delete operations

### `.xdg-volume-info` Handling

`.xdg-volume-info` is a local-only virtual file. Creating or refreshing it:

- Assigns a `local-` ID and registers it in the metadata cache.
- Updates only that entry when the account name changes; the root’s other children remain untouched.
- Never clears the entire root cache, preventing unnecessary re-fetches after maintenance.

### Virtual Items and Overlay Policies

- Virtual items live in the same BBolt buckets as remote-backed entries and are flagged with `virtual=true`, `remote_id=NULL`, and `overlay_policy` (`LOCAL_WINS`, `REMOTE_WINS`, or `MERGED`).
- `LOCAL_WINS` (default for `.xdg-volume-info`, policy folders, or pinned views) hides any remote item that collides by name so FUSE can enumerate a single authoritative child list without merging logic.
- `REMOTE_WINS` is useful for placeholder entries that should disappear when remote content exists (e.g., onboarding tips).
- `MERGED` allows special folders to overlay remote children (e.g., “Pinned” folder containing aliases to real items). The metadata entry stores references to the underlying `item_id`s so state transitions continue to work.
- Because these flags live inside the metadata DB, FUSE handles `readdir` with a single query. The sync engine simply ignores virtual entries when generating upload/delete workqueues.
