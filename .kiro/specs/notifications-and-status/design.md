# Design: Notifications And Status

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/design.md`. Sections below are verbatim.

### 10. User Notification and Feedback System

**Location**: `internal/graph/network_feedback.go`, `internal/fs/dbus.go`

**Notification Mechanisms**:
1. **D-Bus Signals**: Primary notification mechanism when D-Bus is available
2. **Logging**: All notifications are logged regardless of feedback level
3. **Callbacks**: Application-level callbacks for programmatic handling
4. **Extended Attributes**: File status information via xattr

**Feedback Levels**:
- **None**: Suppress user notifications, logging only
- **Basic**: Simple connectivity status messages (default)
- **Detailed**: Comprehensive network and sync information

**Notification Types**:
1. **Network State Changes**:
   - Network Connected: Emitted when connectivity is restored
   - Network Disconnected: Emitted when connectivity is lost
   
2. **Synchronization Events**:
   - Sync Started: Emitted when offline-to-online sync begins
   - Sync Completed: Emitted when all changes are synchronized
   - Conflicts Detected: Emitted when conflicts require attention
   - Sync Failed: Emitted when synchronization errors occur

3. **File Status Updates**:
   - Download Progress: File download status changes
   - Upload Progress: File upload status changes
   - Error Status: File operation errors

**D-Bus Signal Format**:
```
Interface: com.github.jstaf.onedriver.FileStatus
Signals:
  - NetworkStateChanged(connected: bool)
  - SyncStatusChanged(status: string, details: string)
  - FileStatusChanged(path: string, status: string)
```

**Configuration Options**:
- `--feedback-level`: Set notification verbosity (none, basic, detailed)
- `--offline-mode`: Enable manual offline mode
- `--query-offline-status`: Query current network state
- `--query-cache-status`: Query cached files for offline planning

**User Experience**:
- Users receive timely notifications about network state changes
- Synchronization progress is visible and understandable
- Conflicts are clearly communicated with resolution options
- Offline mode can be manually controlled for planned disconnections
- Cache status helps users plan for offline work

**Verification Criteria**:
- D-Bus notifications are emitted correctly when available
- Feedback level configuration is respected
- Network state changes trigger appropriate notifications
- Synchronization events are communicated to users
- Manual offline mode can be activated and deactivated
- Cache status queries provide accurate information
- Notifications work correctly even when D-Bus is unavailable

### 11. File Status and D-Bus Component

**Location**: `internal/fs/file_status.go`, `internal/fs/dbus.go`

**Verification Steps**:
1. Review file status tracking
2. Test D-Bus server initialization
3. Test status signal emission
4. Test extended attribute fallback
5. Test Nemo extension integration

**Expected Interfaces**:
- `FileStatusInfo` struct
- `FileStatusDBusServer` for D-Bus communication
- Extended attributes (user.onemount.status)
- Nemo extension Python script

**Verification Criteria**:
- File status updates correctly
- D-Bus signals are sent when available
- Extended attributes work as fallback
- Nemo extension displays status icons
- Status persists across filesystem restarts
