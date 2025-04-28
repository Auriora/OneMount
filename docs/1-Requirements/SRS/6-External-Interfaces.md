# 6. External Interfaces

This section describes the interfaces between the onedriver system and external entities, including user interfaces and APIs.

## 6.1 User Interfaces

Onedriver provides two main user interfaces:

### 6.1.1 Command-Line Interface (CLI)
The onedriver command-line interface allows users to:
- Mount OneDrive as a filesystem
- Specify mount options and parameters
- View debug and status information

Example usage:
```bash
onedriver [OPTIONS] MOUNTPOINT
```

### 6.1.2 Graphical User Interface (GUI)
The onedriver-launcher application provides a graphical interface built with GTK3 that allows users to:
- Select a mount point directory
- View account information
- Mount and unmount OneDrive
- Configure settings
- View status and error messages

The GUI includes the following components:
- Directory chooser dialog for selecting mount points
- Message dialogs for notifications and confirmations
- System tray integration for status indication

## 6.2 APIs and External Systems

### 6.2.1 Microsoft Graph API
Onedriver interfaces with the Microsoft Graph API to access OneDrive resources. The integration includes:

- **Authentication**: OAuth 2.0 authentication flow for Microsoft accounts
- **Resource Access**: HTTP requests to access files, folders, and metadata
- **Methods**: GET, POST, PATCH, PUT, DELETE operations for CRUD functionality
- **Endpoints**:
  - User information: `/me`
  - Drive information: `/me/drive`
  - Files and folders: `/me/drive/items/{id}`
  - Children items: `/me/drive/items/{id}/children`

### 6.2.2 FUSE (Filesystem in Userspace)
Onedriver uses the FUSE interface to implement a filesystem that can be mounted in Linux. This interface:
- Translates filesystem operations (read, write, create, etc.) to appropriate Graph API calls
- Handles file and directory metadata
- Manages file content caching and retrieval

### 6.2.3 Systemd Integration
Onedriver integrates with systemd for service management:
- Service units for automatic mounting
- Status monitoring
- Start/stop/enable/disable functionality
