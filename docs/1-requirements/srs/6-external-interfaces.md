# 6. External Interfaces

This section describes the interfaces between the OneMount system and external entities, including user interfaces and APIs.

## 6.1 User Interfaces

OneMount provides two main user interfaces:

### 6.1.1 Command-Line Interface (CLI)
The OneMount command-line interface allows users to:
- Mount OneDrive as a filesystem
- Specify mount options and parameters
- View debug and status information

Example usage:
```bash
onemount [OPTIONS] MOUNTPOINT
```

### 6.1.2 Graphical User Interface (GUI)
The OneMount-launcher application provides a graphical interface built with GTK3 that allows users to:
- Select a mount point directory
- View account information
- Mount and unmount OneDrive
- Configure settings
- View status and error messages

The GUI includes the following components:
- Directory chooser dialog for selecting mount points
- Message dialogs for notifications and confirmations
- System tray integration for status indication

### 6.1.3 Nemo File Manager Integration
The OneMount Nemo extension integrates with the Nemo file manager to provide:
- OneDrive mount displayed in the sidebar as a network or cloud mount
- File status icons (cloud, local, syncing, etc.) for each file
- Tooltips with detailed status information
- Context menu options for OneMount operations

The integration uses:
- Python extension for Nemo
- D-Bus interface for real-time status updates
- Extended attributes as a fallback mechanism

## 6.2 APIs and External Systems

### 6.2.1 Microsoft Graph API
OneMount interfaces with the Microsoft Graph API to access OneDrive resources. The integration includes:

- **Authentication**: OAuth 2.0 authentication flow for Microsoft accounts
- **Resource Access**: HTTP requests to access files, folders, and metadata
- **Methods**: GET, POST, PATCH, PUT, DELETE operations for CRUD functionality
- **Endpoints**:
  - User information: `/me`
  - Drive information: `/me/drive`
  - Files and folders: `/me/drive/items/{id}`
  - Children items: `/me/drive/items/{id}/children`

### 6.2.2 FUSE (Filesystem in Userspace)
OneMount uses the FUSE interface to implement a filesystem that can be mounted in Linux. This interface:
- Translates filesystem operations (read, write, create, etc.) to appropriate Graph API calls
- Handles file and directory metadata
- Manages file content caching and retrieval

### 6.2.3 Systemd Integration
OneMount integrates with systemd for service management:
- Service units for automatic mounting
- Status monitoring
- Start/stop/enable/disable functionality

### 6.2.4 D-Bus Interface
OneMount provides a D-Bus interface for file status updates:

- **Service Name**: `org.OneMount.FileStatus`
- **Object Path**: `/org/onemount/FileStatus`
- **Interface**: `org.onemount.FileStatus`

The interface includes:

**Methods**:
- `GetFileStatus(path string) (status string)`: Returns the status of a file at the given path
- `GetFileStatusBatch(paths []string) (statuses map[string]string)`: Returns the status of multiple files in a single call

**Signals**:
- `FileStatusChanged(path string, status string)`: Emitted when a file's status changes
- `BatchStatusChanged(statuses map[string]string)`: Emitted when multiple files' statuses change simultaneously

This interface enables:
- Real-time updates without polling
- Reduced overhead compared to reading extended attributes
- Better integration with other applications
