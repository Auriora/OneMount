# Manual Nemo Extension Testing Guide

## Automation Status

**Last Updated**: 2026-01-26

### Automated Tests (80% Coverage) ✅

**Most tests have been automated** as part of Task 45.3. The following tests are now automated:

#### Go Integration Tests (60% - D-Bus Protocol)

**File**: `internal/fs/nemo_extension_test.go`

| Test | Status | Test Function | Coverage |
|------|--------|---------------|----------|
| Service Discovery | ✅ AUTOMATED | `TestIT_FS_NemoExtension_ServiceDiscovery` | Full |
| GetFileStatus Method | ✅ AUTOMATED | `TestIT_FS_NemoExtension_GetFileStatus` | Full |
| Signal Subscription | ✅ AUTOMATED | `TestIT_FS_NemoExtension_SignalSubscription` | Full |
| Signal Reception | ✅ AUTOMATED | `TestIT_FS_NemoExtension_SignalReception` | Full |
| Error Handling | ✅ AUTOMATED | `TestIT_FS_NemoExtension_ErrorHandling` | Full |
| Performance | ✅ AUTOMATED | `TestIT_FS_NemoExtension_Performance` | Full |

**Run Go integration tests**:
```bash
docker compose -f docker/compose/docker-compose.test.yml \
  -f docker/compose/docker-compose.auth.yml run --rm \
  test-runner go test -v -run "TestIT_FS_NemoExtension" ./internal/fs
```

#### Python Unit Tests (20% - Extension Logic)

**File**: `internal/nemo/tests/test_nemo_extension.py`

| Test Category | Status | Coverage |
|---------------|--------|----------|
| Extension Initialization | ✅ AUTOMATED | Full |
| D-Bus Connection | ✅ AUTOMATED | Full |
| Mount Point Detection | ✅ AUTOMATED | Full |
| File Status Retrieval | ✅ AUTOMATED | Full |
| Status-to-Emblem Mapping | ✅ AUTOMATED | Full |
| Mount Filtering | ✅ AUTOMATED | Full |
| Signal Handling | ✅ AUTOMATED | Full |
| Error Handling | ✅ AUTOMATED | Full |

**Run Python unit tests**:
```bash
docker compose -f docker/compose/docker-compose.test.yml run --rm \
  test-runner python3 -m pytest internal/nemo/tests/test_nemo_extension.py -v
```

**Related Documentation**:
- Go Tests: `docs/testing/nemo-extension-automation-complete.md`
- Python Tests: `docs/testing/nemo-extension-python-tests-complete.md`
- Analysis: `docs/testing/manual-tests-automation-analysis.md`

### Manual Tests (20% Coverage) ⚠️

The following aspects still require manual verification:
- **Visual icon appearance** - Verifying emblems actually display in Nemo
- **Icon clarity and visibility** - Checking icon rendering quality
- **View mode compatibility** - Testing list, grid, and compact views
- **Desktop environment integration** - Real-world usage testing

**Note**: Manual testing is only needed for visual verification. All D-Bus protocol correctness and extension logic is verified by automated tests.

---

## Overview

This guide provides step-by-step instructions for manually testing the OneMount Nemo file manager extension. The extension adds visual status indicators (emblems) to files and folders in Nemo, showing their synchronization state with OneDrive.

**Test Scope**: Nemo extension installation, emblem display, D-Bus integration, status updates

**Requirements Validated**: Requirement 10.3 - Nemo extension provides file status information

**⚠️ IMPORTANT**: Most tests are now automated (80% coverage). Manual testing is only needed for visual icon verification. See the "Automation Status" section above for automated test coverage.

---

## Prerequisites

### System Requirements

- Linux system with Nemo file manager installed (Cinnamon desktop or standalone)
- OneMount installed and configured
- Valid OneDrive authentication tokens
- Python 3 with required dependencies

### Required Software Installation

```bash
# Install Nemo and Python dependencies (Ubuntu/Debian)
sudo apt-get install nemo python3-nemo python3-gi

# Install Nemo and Python dependencies (Fedora/RHEL)
sudo dnf install nemo python3-nemo python3-gobject

# Verify Nemo is installed
nemo --version
```

### Verify Python Dependencies

```bash
# Check Python 3 is available
python3 --version

# Verify GObject introspection
python3 -c "import gi; gi.require_version('Nemo', '3.0'); from gi.repository import Nemo; print('Nemo bindings OK')"
```

---

## Extension Installation

### Step 1: Locate Extension File

The Nemo extension is located at:
```
internal/nemo/src/nemo-onemount.py
```

### Step 2: Install Extension

```bash
# Create Nemo extensions directory if it doesn't exist
mkdir -p ~/.local/share/nemo-python/extensions/

# Copy the extension file
cp internal/nemo/src/nemo-onemount.py ~/.local/share/nemo-python/extensions/

# Make it executable
chmod +x ~/.local/share/nemo-python/extensions/nemo-onemount.py
```

### Step 3: Restart Nemo

```bash
# Quit all Nemo instances
nemo -q

# Wait a moment
sleep 2

# Start Nemo
nemo &
```

### Step 4: Verify Extension is Loaded

```bash
# Start Nemo with debug output
nemo --quit
nemo --debug 2>&1 | grep -i onemount
```

**Expected Output**: Should see messages about loading the OneMount extension


---

## Test Environment Setup

### 1. Mount OneMount

```bash
# Create mount point
mkdir -p ~/test-onedrive-mount

# Mount OneMount
onemount ~/test-onedrive-mount

# Verify mount
mount | grep test-onedrive-mount
```

### 2. Verify D-Bus Service is Running

```bash
# Check D-Bus service name file
cat /tmp/onemount-dbus-service-name

# Verify service is registered
SERVICE_NAME=$(cat /tmp/onemount-dbus-service-name)
dbus-send --session --print-reply \
  --dest=$SERVICE_NAME \
  /org/onemount/FileStatus \
  org.freedesktop.DBus.Introspectable.Introspect
```

### 3. Prepare Test Files

Create various test files to verify different status states:

```bash
# Create a new file (will be "Modified" or "Uploading")
echo "Test content" > ~/test-onedrive-mount/new-test-file.txt

# Wait for upload to complete
sleep 5

# Access an existing file (will be "Cached" after download)
cat ~/test-onedrive-mount/existing-file.txt > /dev/null
```

---

## Test Procedures

**Note**: Many D-Bus communication tests have been automated. See `internal/fs/nemo_extension_test.go` for automated test coverage. The tests below focus on visual verification that requires manual inspection.

### Automated Tests (No Manual Testing Required)

The following tests are now automated and run as part of the CI/CD pipeline:

#### Go Integration Tests (D-Bus Protocol)

- ✅ **Service Discovery** - `TestIT_FS_NemoExtension_ServiceDiscovery`
  - Verifies Nemo extension can discover OneMount D-Bus service
  - Tests service name resolution and connection establishment

- ✅ **GetFileStatus Method Calls** - `TestIT_FS_NemoExtension_GetFileStatus`
  - Verifies GetFileStatus D-Bus method returns correct status for all file states
  - Tests with Local, Downloading, Syncing, Modified, Error, and Conflict statuses
  - Verifies Unknown status for non-existent files

- ✅ **Signal Subscription** - `TestIT_FS_NemoExtension_SignalSubscription`
  - Verifies Nemo extension can subscribe to FileStatusChanged signals
  - Tests D-Bus signal handler setup

- ✅ **Signal Reception** - `TestIT_FS_NemoExtension_SignalReception`
  - Verifies Nemo extension receives FileStatusChanged signals
  - Tests signal data correctness (path and status)
  - Tests multiple rapid signal updates

- ✅ **Error Handling** - `TestIT_FS_NemoExtension_ErrorHandling`
  - Verifies graceful handling when D-Bus service unavailable
  - Tests error messages and fallback behavior

- ✅ **Performance** - `TestIT_FS_NemoExtension_Performance`
  - Verifies GetFileStatus queries complete in < 10ms per file
  - Tests with 50+ files to ensure performance requirements met

#### Python Unit Tests (Extension Logic)

Run with: `docker compose -f docker/compose/docker-compose.test.yml run --rm test-runner python3 -m pytest internal/nemo/tests/test_nemo_extension.py -v`

- ✅ **Extension Initialization** - `test_extension_initialization_success`, `test_extension_initialization_no_dbus`
  - Verifies extension initializes correctly with and without D-Bus
  - Tests attribute setup (bus, proxy, cache, mounts)

- ✅ **D-Bus Connection** - `test_dbus_connection_success`, `test_dbus_connection_failure`
  - Verifies D-Bus connection establishment and failure handling
  - Tests service discovery and proxy creation

- ✅ **Mount Point Detection** - `test_get_onemount_mounts_success`, `test_get_onemount_mounts_no_mounts`, `test_get_onemount_mounts_file_error`
  - Verifies OneMount mount point detection from /proc/mounts
  - Tests with various mount configurations and error conditions

- ✅ **File Status Retrieval** - `test_get_file_status_dbus_success`, `test_get_file_status_cached`, `test_get_file_status_dbus_fallback_to_xattr`
  - Verifies file status retrieval via D-Bus and xattr fallback
  - Tests caching behavior and error handling

- ✅ **Status-to-Emblem Mapping** - `test_emblem_assignment_all_statuses`, `test_emblem_assignment_unrecognized_status`
  - Verifies correct emblem assignment for all file statuses
  - Tests: Cloud→emblem-synchronizing-offline, Local→emblem-default, Modified→emblem-synchronizing-locally-modified, etc.
  - Tests unrecognized status handling (emblem-question)

- ✅ **Mount Filtering** - `test_no_emblem_for_non_onemount_files`, `test_update_file_info_no_path`
  - Verifies emblems only appear on files within OneMount mounts
  - Tests edge cases (no path, outside mount)

- ✅ **Signal Handling** - `test_file_status_changed_signal`, `test_file_status_changed_signal_error`
  - Verifies FileStatusChanged signal reception and processing
  - Tests cache updates and emblem refresh triggers
  - Tests error handling during signal processing

- ✅ **Error Handling** - `test_dbus_reconnection_on_error`, `test_module_init_function`
  - Verifies D-Bus reconnection on communication errors
  - Tests graceful degradation and fallback behavior
  - Tests module initialization function

**Automation Coverage**: 80% of Nemo extension testing is now automated (60% D-Bus protocol + 20% extension logic), focusing on D-Bus protocol correctness and extension logic verification.

### Manual Tests (Visual Verification Required)

### Test 1: Extension Installation Verification

**Objective**: Confirm the extension is properly installed and loaded

**Steps**:

1. **Check extension file exists**:
   ```bash
   ls -l ~/.local/share/nemo-python/extensions/nemo-onemount.py
   ```
   
   **Expected**: File exists and is executable

2. **Start Nemo with debug output**:
   ```bash
   nemo --quit
   nemo --debug 2>&1 | tee /tmp/nemo-debug.log
   ```

3. **Search for extension loading messages**:
   ```bash
   grep -i "onemount\|extension" /tmp/nemo-debug.log
   ```
   
   **Expected**: Messages indicating extension was loaded

4. **Check for Python errors**:
   ```bash
   grep -i "error\|exception\|traceback" /tmp/nemo-debug.log | grep -i onemount
   ```
   
   **Expected**: No errors related to OneMount extension

**Pass Criteria**:
- ✅ Extension file is present and executable
- ✅ Nemo loads the extension without errors
- ✅ No Python exceptions in debug output
- ✅ Extension initialization messages appear in logs

---

### Test 2: Status Icons Appear on Files

**Objective**: Verify status emblems are displayed on files in OneMount mount

**Steps**:

1. **Open Nemo and navigate to OneMount mount**:
   ```bash
   nemo ~/test-onedrive-mount &
   ```

2. **Observe file icons**:
   - Look for emblem overlays on file icons
   - Check both files and folders
   - Verify emblems are visible and distinct

3. **Compare with non-OneMount directory**:
   ```bash
   # Open a regular directory
   nemo ~/Documents &
   ```
   
   **Expected**: Files in regular directory have no OneMount emblems

4. **Take screenshots** (for documentation):
   - Screenshot of OneMount directory with emblems
   - Screenshot of regular directory without emblems

**Expected Emblems**:
- Files should have status emblems overlaid on their icons
- Different file states should show different emblems
- Emblems should be clearly visible

**Pass Criteria**:
- ✅ Emblems appear on files in OneMount mount
- ✅ No emblems on files outside OneMount mount
- ✅ Emblems are visually distinct and recognizable
- ✅ Both files and folders show appropriate emblems

---

### Test 3: Icon States for Different File Statuses

**Objective**: Verify correct emblems are shown for each file status

**Setup**: Create files in different states

```bash
# 1. Ghost/Cloud file (not cached)
# Clear cache and list a file without accessing it
fusermount3 -uz ~/test-onedrive-mount
rm -rf ~/.cache/onemount/
onemount ~/test-onedrive-mount
# Don't access the file yet

# 2. Cached file
cat ~/test-onedrive-mount/test-file.txt > /dev/null

# 3. Modified file
echo "Modified content" >> ~/test-onedrive-mount/test-file.txt

# 4. New file (uploading)
echo "New file" > ~/test-onedrive-mount/new-file.txt
```

**Steps**:

1. **Open Nemo to OneMount mount**:
   ```bash
   nemo ~/test-onedrive-mount &
   ```

2. **Verify each file status emblem**:

   | File State | Expected Emblem | Icon Description |
   |------------|----------------|------------------|
   | Cloud (Ghost) | emblem-synchronizing-offline | Cloud icon |
   | Cached (Local) | emblem-default | Checkmark or default |
   | Modified | emblem-synchronizing-locally-modified | Pencil or edit icon |
   | Syncing | emblem-synchronizing | Sync arrows |
   | Downloading | emblem-downloads | Download arrow |
   | Out of Sync | emblem-important | Exclamation mark |
   | Error | emblem-error | Red X or error icon |
   | Conflict | emblem-warning | Warning triangle |
   | Unknown | emblem-question | Question mark |

3. **Document observed emblems**:
   - Take screenshot of each file state
   - Note which emblem appears for each state
   - Verify emblem matches expected icon

**Pass Criteria**:
- ✅ Each file status shows a distinct emblem
- ✅ Emblems match the documented mapping
- ✅ Emblems are appropriate for the file state
- ✅ No missing or incorrect emblems

---


### Test 4: Icon Updates During File Operations

**Objective**: Verify emblems update in real-time as file status changes

**Steps**:

1. **Open Nemo to OneMount mount** (keep window visible):
   ```bash
   nemo ~/test-onedrive-mount &
   ```

2. **Test download operation**:
   ```bash
   # In terminal: Clear cache and access a file
   fusermount3 -uz ~/test-onedrive-mount
   rm -rf ~/.cache/onemount/
   onemount ~/test-onedrive-mount
   
   # Access a large file to see download progress
   cat ~/test-onedrive-mount/large-file.pdf > /dev/null &
   ```
   
   **Observe in Nemo**:
   - Initial state: Cloud emblem (emblem-synchronizing-offline)
   - During download: Downloading emblem (emblem-downloads)
   - After download: Cached emblem (emblem-default)

3. **Test file modification**:
   ```bash
   # Modify a file
   echo "Modified" >> ~/test-onedrive-mount/test-file.txt
   ```
   
   **Observe in Nemo**:
   - Should change to Modified emblem (emblem-synchronizing-locally-modified)
   - May briefly show Syncing emblem during upload
   - Should return to Cached emblem after upload completes

4. **Test file creation**:
   ```bash
   # Create new file
   echo "New file" > ~/test-onedrive-mount/new-test-file.txt
   ```
   
   **Observe in Nemo**:
   - New file should appear with Modified or Syncing emblem
   - Should transition to Cached after upload

5. **Test file deletion**:
   ```bash
   # Delete file
   rm ~/test-onedrive-mount/new-test-file.txt
   ```
   
   **Observe in Nemo**:
   - File should disappear from listing
   - No errors or stale entries

**Pass Criteria**:
- ✅ Emblems update automatically during operations
- ✅ State transitions are visible (Cloud → Downloading → Cached)
- ✅ Updates occur within 1-2 seconds of state change
- ✅ No stale or incorrect emblems after operations

---

### Test 5: Context Menu Integration

**Objective**: Verify OneMount context menu items work correctly

**Steps**:

1. **Open Nemo to OneMount mount**:
   ```bash
   nemo ~/test-onedrive-mount &
   ```

2. **Test file context menu**:
   - Right-click on a file in OneMount mount
   - Look for "OneMount: Refresh status emblems" menu item
   - Click the menu item
   - Verify emblem refreshes

3. **Test folder background context menu**:
   - Right-click on empty space in OneMount folder
   - Look for "OneMount: Refresh folder emblems" menu item
   - Click the menu item
   - Verify all emblems in folder refresh

4. **Test outside OneMount mount**:
   - Navigate to a non-OneMount directory (e.g., ~/Documents)
   - Right-click on a file
   - Verify OneMount menu items do NOT appear

**Expected Context Menu Items**:
- "OneMount: Refresh status emblems" (on files/folders)
- "OneMount: Refresh folder emblems" (on folder background)
- Items should only appear within OneMount mounts

**Pass Criteria**:
- ✅ Context menu items appear in OneMount mounts
- ✅ Context menu items do NOT appear outside OneMount mounts
- ✅ Refresh actions work and update emblems
- ✅ No errors when clicking menu items

---

### Test 6: Multiple Windows and Tabs

**Objective**: Verify extension works correctly with multiple Nemo windows/tabs

**Steps**:

1. **Open multiple Nemo windows**:
   ```bash
   nemo ~/test-onedrive-mount &
   nemo ~/test-onedrive-mount &
   ```

2. **Perform file operation in terminal**:
   ```bash
   echo "Test" > ~/test-onedrive-mount/multi-window-test.txt
   ```

3. **Observe both windows**:
   - Both windows should show the new file
   - Both windows should show correct emblem
   - Emblems should update in both windows

4. **Test with tabs**:
   - Open Nemo with one window
   - Open new tab (Ctrl+T)
   - Navigate both tabs to OneMount mount
   - Perform file operation
   - Verify both tabs update

**Pass Criteria**:
- ✅ Emblems appear correctly in all windows
- ✅ Updates propagate to all open windows/tabs
- ✅ No conflicts or race conditions
- ✅ Performance remains acceptable with multiple windows

---

### Test 7: Performance with Large Directories

**Objective**: Verify extension performance with directories containing many files

**Steps**:

1. **Create test directory with many files**:
   ```bash
   mkdir -p ~/test-onedrive-mount/large-dir
   for i in {1..100}; do
     echo "File $i" > ~/test-onedrive-mount/large-dir/file-$i.txt
   done
   ```

2. **Open directory in Nemo**:
   ```bash
   nemo ~/test-onedrive-mount/large-dir &
   ```

3. **Measure performance**:
   - Time how long it takes for all emblems to appear
   - Check CPU usage (use `top` or `htop`)
   - Verify Nemo remains responsive

4. **Test scrolling**:
   - Scroll through the file list
   - Verify emblems load smoothly
   - Check for lag or freezing

5. **Test refresh**:
   - Press F5 to refresh
   - Verify emblems reload correctly
   - Check refresh time

**Performance Expectations**:
- Initial emblem load: < 5 seconds for 100 files
- Refresh time: < 3 seconds
- CPU usage: < 25% during emblem updates
- No UI freezing or lag

**Pass Criteria**:
- ✅ All emblems appear within acceptable time
- ✅ Nemo remains responsive during emblem updates
- ✅ No excessive CPU or memory usage
- ✅ Scrolling and navigation remain smooth

---

### Test 8: D-Bus Communication Verification

**Objective**: Verify extension communicates correctly with OneMount via D-Bus

**Steps**:

1. **Monitor D-Bus traffic**:
   ```bash
   # Terminal 1: Monitor D-Bus
   dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'" &
   ```

2. **Open Nemo and perform operations**:
   ```bash
   # Terminal 2: Open Nemo
   nemo ~/test-onedrive-mount &
   
   # Perform file operation
   echo "Test" > ~/test-onedrive-mount/dbus-test.txt
   ```

3. **Observe D-Bus monitor**:
   - Should see `FileStatusChanged` signals
   - Signals should contain file path and status
   - Extension should receive and process signals

4. **Test GetFileStatus method call**:
   ```bash
   # Check if extension calls GetFileStatus
   # This happens when emblems are displayed
   ```

5. **Verify fallback to xattrs**:
   ```bash
   # Stop D-Bus temporarily
   SERVICE_NAME=$(cat /tmp/onemount-dbus-service-name)
   # Kill OneMount to stop D-Bus service
   fusermount3 -uz ~/test-onedrive-mount
   
   # Remount without D-Bus
   unset DBUS_SESSION_BUS_ADDRESS
   onemount ~/test-onedrive-mount
   
   # Open Nemo - should still show emblems via xattrs
   nemo ~/test-onedrive-mount &
   ```

**Pass Criteria**:
- ✅ Extension receives D-Bus signals
- ✅ Extension calls GetFileStatus method
- ✅ Extension falls back to xattrs when D-Bus unavailable
- ✅ No errors in D-Bus communication

---


### Test 9: Extension Error Handling

**Objective**: Verify extension handles errors gracefully

**Steps**:

1. **Test with OneMount not running**:
   ```bash
   # Unmount OneMount
   fusermount3 -uz ~/test-onedrive-mount
   
   # Open Nemo to the (now empty) mount point
   nemo ~/test-onedrive-mount &
   ```
   
   **Expected**: No errors, no emblems (normal behavior)

2. **Test with invalid mount point**:
   ```bash
   # Navigate to non-existent directory
   nemo ~/non-existent-mount &
   ```
   
   **Expected**: No crashes, no error dialogs

3. **Test with permission errors**:
   ```bash
   # Create file with no read permissions
   touch ~/test-onedrive-mount/no-perms.txt
   chmod 000 ~/test-onedrive-mount/no-perms.txt
   
   # Open in Nemo
   nemo ~/test-onedrive-mount &
   ```
   
   **Expected**: File shows with "Unknown" emblem or no emblem

4. **Test with D-Bus service restart**:
   ```bash
   # While Nemo is open, restart OneMount
   fusermount3 -uz ~/test-onedrive-mount
   onemount ~/test-onedrive-mount
   
   # Refresh Nemo (F5)
   ```
   
   **Expected**: Extension reconnects to D-Bus, emblems reappear

**Pass Criteria**:
- ✅ No crashes or error dialogs
- ✅ Graceful degradation when OneMount unavailable
- ✅ Automatic reconnection when service restarts
- ✅ Appropriate fallback behavior (Unknown emblem or no emblem)

---

## Icon Reference

### Complete Emblem Mapping

| Status | Emblem Name | Visual Description | When It Appears |
|--------|-------------|-------------------|-----------------|
| Cloud | emblem-synchronizing-offline | Cloud icon | File exists in OneDrive but not cached locally |
| Local (Cached) | emblem-default | Checkmark | File is cached locally and synced |
| LocalModified | emblem-synchronizing-locally-modified | Pencil/edit icon | File modified locally, pending upload |
| Syncing | emblem-synchronizing | Circular arrows | File currently being synchronized |
| Downloading | emblem-downloads | Download arrow | File being downloaded from OneDrive |
| OutOfSync | emblem-important | Exclamation mark | File needs to be updated from OneDrive |
| Error | emblem-error | Red X or error symbol | Error occurred during sync |
| Conflict | emblem-warning | Warning triangle | Conflict between local and remote versions |
| Unknown | emblem-question | Question mark | Status cannot be determined |

### Emblem Visual Examples

**Note**: Actual emblem appearance depends on your icon theme. The descriptions above are typical representations.

To see your system's emblems:
```bash
# View available emblems
ls /usr/share/icons/*/emblems/
```

---

## Troubleshooting

### Issue: No emblems appear on files

**Possible Causes**:
1. Extension not installed correctly
2. Nemo not restarted after installation
3. OneMount not running
4. File not in OneMount mount

**Solutions**:

```bash
# 1. Verify extension is installed
ls -l ~/.local/share/nemo-python/extensions/nemo-onemount.py

# 2. Restart Nemo
nemo -q
sleep 2
nemo &

# 3. Verify OneMount is mounted
mount | grep onemount

# 4. Check Nemo debug output
nemo --quit
nemo --debug 2>&1 | grep -i onemount
```

---

### Issue: Extension not loading

**Possible Causes**:
1. Missing Python dependencies
2. Python syntax errors in extension
3. Nemo Python support not installed

**Solutions**:

```bash
# 1. Check Python dependencies
python3 -c "import gi; gi.require_version('Nemo', '3.0'); from gi.repository import Nemo"

# 2. Test extension syntax
python3 -m py_compile ~/.local/share/nemo-python/extensions/nemo-onemount.py

# 3. Install Nemo Python support
sudo apt-get install python3-nemo  # Ubuntu/Debian
sudo dnf install python3-nemo      # Fedora/RHEL

# 4. Check Nemo extensions directory
ls -la ~/.local/share/nemo-python/extensions/
```

---

### Issue: Emblems not updating

**Possible Causes**:
1. D-Bus signals not being received
2. Extension cache not being cleared
3. Nemo not refreshing file info

**Solutions**:

```bash
# 1. Check D-Bus connection
dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'"

# 2. Use context menu to refresh
# Right-click in Nemo → "OneMount: Refresh folder emblems"

# 3. Force Nemo refresh
# Press F5 in Nemo window

# 4. Restart Nemo
nemo -q
nemo &
```

---

### Issue: Wrong emblems displayed

**Possible Causes**:
1. Status cache out of sync
2. D-Bus communication issues
3. Extended attributes not updated

**Solutions**:

```bash
# 1. Check actual file status
getfattr -n user.onemount.status ~/test-onedrive-mount/file.txt

# 2. Query via D-Bus
SERVICE_NAME=$(cat /tmp/onemount-dbus-service-name)
dbus-send --session --print-reply \
  --dest=$SERVICE_NAME \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/path/to/file.txt"

# 3. Clear extension cache and refresh
# Use context menu: "OneMount: Refresh status emblems"
```

---

### Issue: High CPU usage

**Possible Causes**:
1. Too many files being processed
2. Excessive D-Bus traffic
3. Extension polling too frequently

**Solutions**:

```bash
# 1. Monitor CPU usage
top -p $(pgrep nemo)

# 2. Check D-Bus traffic
dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'" | wc -l

# 3. Reduce directory size or close Nemo windows
```

---

### Issue: Extension crashes Nemo

**Possible Causes**:
1. Python exception in extension code
2. Incompatible Nemo version
3. Missing dependencies

**Solutions**:

```bash
# 1. Check Nemo crash logs
journalctl --user -u nemo -n 100

# 2. Run Nemo with debug output
nemo --quit
nemo --debug 2>&1 | tee /tmp/nemo-crash.log

# 3. Temporarily disable extension
mv ~/.local/share/nemo-python/extensions/nemo-onemount.py \
   ~/.local/share/nemo-python/extensions/nemo-onemount.py.disabled
nemo -q
nemo &
```

---

## Debugging Commands

### Check Extension Status

```bash
# Verify extension file
ls -l ~/.local/share/nemo-python/extensions/nemo-onemount.py

# Check Python syntax
python3 -m py_compile ~/.local/share/nemo-python/extensions/nemo-onemount.py

# Test Python imports
python3 -c "import gi; gi.require_version('Nemo', '3.0'); from gi.repository import Nemo, GObject, Gio, GLib; import dbus"
```

### Monitor D-Bus Communication

```bash
# Monitor all OneMount D-Bus signals
dbus-monitor --session "type='signal',interface='org.onemount.FileStatus'"

# Monitor method calls
dbus-monitor --session "type='method_call',interface='org.onemount.FileStatus'"

# Check service registration
dbus-send --session --print-reply \
  --dest=org.freedesktop.DBus \
  /org/freedesktop/DBus \
  org.freedesktop.DBus.ListNames | grep onemount
```

### Check File Status

```bash
# Via extended attributes
getfattr -n user.onemount.status ~/test-onedrive-mount/file.txt

# Via D-Bus
SERVICE_NAME=$(cat /tmp/onemount-dbus-service-name)
dbus-send --session --print-reply \
  --dest=$SERVICE_NAME \
  /org/onemount/FileStatus \
  org.onemount.FileStatus.GetFileStatus \
  string:"/path/to/file.txt"
```

### Nemo Debug Output

```bash
# Start Nemo with full debug output
nemo --quit
NEMO_DEBUG=1 nemo --debug 2>&1 | tee /tmp/nemo-full-debug.log

# Filter for OneMount messages
grep -i onemount /tmp/nemo-full-debug.log
```

---


## Results Documentation Template

### Test Execution Record

**Date**: _______________  
**Tester**: _______________  
**OneMount Version**: _______________  
**Nemo Version**: _______________  
**System**: _______________  
**Desktop Environment**: _______________  

### Test Results

| Test # | Test Name | Status | Notes |
|--------|-----------|--------|-------|
| 1 | Extension Installation | ☐ Pass ☐ Fail | |
| 2 | Status Icons Appear | ☐ Pass ☐ Fail | |
| 3 | Icon States | ☐ Pass ☐ Fail | |
| 4 | Icon Updates | ☐ Pass ☐ Fail | |
| 5 | Context Menu | ☐ Pass ☐ Fail | |
| 6 | Multiple Windows | ☐ Pass ☐ Fail | |
| 7 | Large Directories | ☐ Pass ☐ Fail | |
| 8 | D-Bus Communication | ☐ Pass ☐ Fail | |
| 9 | Error Handling | ☐ Pass ☐ Fail | |

### Emblem Verification

| File Status | Expected Emblem | Observed Emblem | Correct? |
|-------------|----------------|-----------------|----------|
| Cloud | emblem-synchronizing-offline | | ☐ Yes ☐ No |
| Cached | emblem-default | | ☐ Yes ☐ No |
| Modified | emblem-synchronizing-locally-modified | | ☐ Yes ☐ No |
| Syncing | emblem-synchronizing | | ☐ Yes ☐ No |
| Downloading | emblem-downloads | | ☐ Yes ☐ No |
| Out of Sync | emblem-important | | ☐ Yes ☐ No |
| Error | emblem-error | | ☐ Yes ☐ No |
| Conflict | emblem-warning | | ☐ Yes ☐ No |
| Unknown | emblem-question | | ☐ Yes ☐ No |

### Performance Metrics

| Metric | Measurement | Acceptable? |
|--------|-------------|-------------|
| Initial emblem load (100 files) | _____ seconds | ☐ Yes ☐ No |
| Emblem update latency | _____ seconds | ☐ Yes ☐ No |
| CPU usage during updates | _____ % | ☐ Yes ☐ No |
| Memory usage | _____ MB | ☐ Yes ☐ No |
| Nemo responsiveness | ☐ Good ☐ Acceptable ☐ Poor | |

### D-Bus Integration

| Feature | Working? | Notes |
|---------|----------|-------|
| Signal reception | ☐ Yes ☐ No | |
| Method calls | ☐ Yes ☐ No | |
| Service discovery | ☐ Yes ☐ No | |
| Fallback to xattrs | ☐ Yes ☐ No | |

### Issues Found

| Issue # | Description | Severity | Reproducible | Steps to Reproduce |
|---------|-------------|----------|--------------|-------------------|
| | | | | |

### Screenshots

☐ Emblems on files in OneMount mount  
☐ Each emblem state (Cloud, Cached, Modified, etc.)  
☐ Context menu items  
☐ Multiple windows with emblems  
☐ Large directory performance  

### Overall Assessment

☐ **PASS** - All tests passed, extension working correctly  
☐ **FAIL** - One or more critical tests failed  
☐ **PARTIAL** - Minor issues found but core functionality works  

**Comments**: _______________________________________________

---

## Pass/Fail Criteria

### Overall Pass Criteria

The Nemo extension test suite **PASSES** if:

1. ✅ Extension installs and loads without errors
2. ✅ Status emblems appear on files in OneMount mounts
3. ✅ Correct emblems are shown for each file status
4. ✅ Emblems update in real-time during file operations
5. ✅ Context menu items work correctly
6. ✅ Extension works with multiple windows/tabs
7. ✅ Performance is acceptable with large directories
8. ✅ D-Bus communication works correctly
9. ✅ Extension handles errors gracefully
10. ✅ Fallback to xattrs works when D-Bus unavailable

### Critical Failures

The test suite **FAILS** if:

1. ❌ Extension fails to load or crashes Nemo
2. ❌ No emblems appear on any files
3. ❌ Emblems are consistently wrong or misleading
4. ❌ Extension causes Nemo to freeze or become unresponsive
5. ❌ D-Bus communication completely broken
6. ❌ Extension shows emblems on non-OneMount files
7. ❌ Context menu items cause errors or crashes

### Performance Requirements

- Emblem load time: < 5 seconds for 100 files
- Update latency: < 2 seconds after status change
- CPU usage: < 25% during emblem updates
- Memory usage: < 50 MB additional for extension
- Nemo remains responsive at all times

### Acceptable Limitations

The following are acceptable:

- ✅ Slight delay (< 2s) in emblem updates
- ✅ Emblems may not update if Nemo window not focused
- ✅ Performance degradation with > 1000 files in directory
- ✅ Fallback to xattrs when D-Bus unavailable

### Unacceptable Limitations

The following are NOT acceptable:

- ❌ Emblems never appear
- ❌ Nemo crashes or freezes
- ❌ Emblems on non-OneMount files
- ❌ No fallback when D-Bus unavailable
- ❌ Excessive CPU or memory usage

---

## Known Limitations and Edge Cases

### Limitation 1: Icon Theme Dependency

**Description**: Emblem appearance depends on the system icon theme.

**Impact**: Emblems may look different on different systems.

**Workaround**: None - this is expected behavior.

**Acceptable**: Yes - emblems are theme-dependent by design.

---

### Limitation 2: Refresh Delay

**Description**: Emblems may take 1-2 seconds to update after status change.

**Impact**: Brief period where emblem doesn't match actual status.

**Workaround**: Use context menu "Refresh" option for immediate update.

**Acceptable**: Yes - small delay is acceptable for performance.

---

### Limitation 3: Large Directory Performance

**Description**: Performance may degrade with > 1000 files in a directory.

**Impact**: Slower emblem loading and updates.

**Workaround**: Avoid opening very large directories in Nemo.

**Acceptable**: Yes - performance trade-off for large directories.

---

### Limitation 4: Background Window Updates

**Description**: Emblems may not update in background Nemo windows.

**Impact**: Need to focus window or refresh to see updates.

**Workaround**: Focus window or press F5 to refresh.

**Acceptable**: Yes - optimization to reduce resource usage.

---

## Additional Resources

### Extension Source Code

- Location: `internal/nemo/src/nemo-onemount.py`
- Documentation: `internal/nemo/README.nemo-extension.md`
- Tests: `internal/nemo/tests/`

### Related Documentation

- [D-Bus Integration Guide](./manual-dbus-integration-guide.md)
- [D-Bus Fallback Guide](./manual-dbus-fallback-guide.md)
- [File Status System](../2-architecture/file-status.md)
- [Extended Attributes](../2-architecture/extended-attributes.md)

### Nemo Extension Development

- [Nemo Python Extensions](https://github.com/linuxmint/nemo-extensions)
- [GObject Introspection](https://gi.readthedocs.io/)
- [D-Bus Python Tutorial](https://dbus.freedesktop.org/doc/dbus-python/)

### Related Requirements

- Requirement 10.1: Extended attribute updates
- Requirement 10.2: D-Bus signal emission
- Requirement 10.3: Nemo extension integration
- Requirement 10.4: D-Bus fallback behavior
- Requirement 10.5: Download progress updates

---

## Conclusion

This guide provides comprehensive testing procedures for verifying the OneMount Nemo extension. The extension must display accurate status emblems on files, update in real-time, and integrate seamlessly with the Nemo file manager.

**Key Takeaways**:
- Extension must work with both D-Bus and xattr fallback
- Emblems must be accurate and update in real-time
- Performance must remain acceptable even with many files
- Extension must handle errors gracefully without crashing Nemo
- Context menu integration provides manual refresh capability

**Success Criteria**: All emblems display correctly, update in real-time, and the extension integrates seamlessly with Nemo without performance issues or crashes.
