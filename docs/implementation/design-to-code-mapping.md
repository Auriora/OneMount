# Design to Code Mapping

This document provides a mapping between the design elements in the PlantUML diagrams and the actual code artifacts in the onedriver codebase.

## Core Engine Class Diagram

| Design Element | Code Artifact | Description |
|----------------|---------------|-------------|
| Filesystem | fs/fs.go | The main filesystem implementation |
| Inode | fs/inode.go | Represents files and directories in the filesystem |
| DriveItem | fs/graph/drive_item.go | Represents items in OneDrive |
| UploadManager | fs/upload_manager.go | Manages file uploads to OneDrive |
| DownloadManager | fs/download_manager.go | Manages file downloads from OneDrive |
| Auth | fs/graph/oauth2.go | Manages authentication with Microsoft Graph API |
| LoopbackCache | fs/content_cache.go | Handles caching of file content |
| ThumbnailCache | fs/thumbnail_cache.go | Handles caching of thumbnails |
| FileStatus | fs/file_status.go | Tracks the status of files |
| UploadSession | fs/upload_session.go | Manages individual upload operations |
| DownloadSession | fs/download_manager.go | Manages individual download operations |
| Cache | fs/cache.go | General caching functionality for the filesystem |
| DBus | fs/dbus.go | D-Bus interface for file status updates |
| DirOperations | fs/dir_operations.go | Directory-specific filesystem operations |
| FileOperations | fs/file_operations.go | File-specific filesystem operations |
| ThumbnailOperations | fs/file_operations_thumbnails.go, fs/fuse_thumbnail_handlers.go | Thumbnail-specific operations |
| MetadataOperations | fs/metadata_operations.go | Metadata-specific filesystem operations |
| MethodDecorators | fs/method_decorators.go | Decorators for method logging and performance tracking |
| OfflineMode | fs/offline.go | Handles filesystem operations in offline mode |
| Profiler | fs/profiler.go | Performance profiling functionality |
| SignalHandlers | fs/signal_handlers.go | Handles system signals |
| Stats | fs/stats.go | Collects and reports filesystem statistics |
| Sync | fs/sync.go | Synchronization functionality |
| XattrOperations | fs/xattr_operations.go | Extended attribute operations |

## Graph API Class Diagram

| Design Element | Code Artifact | Description |
|----------------|---------------|-------------|
| AuthConfig | fs/graph/oauth2.go | Configures the authentication flow |
| Auth | fs/graph/oauth2.go | Represents authentication tokens |
| AuthError | fs/graph/oauth2.go | Represents authentication errors |
| AuthGTK | fs/graph/oauth2_gtk.go, fs/graph/oauth2_gtk.c, fs/graph/oauth2_gtk.h | GUI-based authentication components |
| AuthHeadless | fs/graph/oauth2_headless.go | Terminal-based authentication components |
| DriveItemParent | fs/graph/drive_item.go | Describes a DriveItem's parent |
| Folder | fs/graph/drive_item.go | Represents a folder in OneDrive |
| File | fs/graph/drive_item.go | Represents a file in OneDrive |
| Deleted | fs/graph/drive_item.go | Represents a deleted item in OneDrive |
| Hashes | fs/graph/hashes.go | Contains file hashes for integrity checking |
| HashFunctions | fs/graph/quickxorhash | Implements QuickXORHash algorithm for file integrity |
| DriveItem | fs/graph/drive_item.go | Represents items in OneDrive |
| User | fs/graph/graph.go | Represents user information |
| Drive | fs/graph/graph.go | Contains information about the user's OneDrive |
| DriveQuota | fs/graph/graph.go | Contains quota information for the drive |
| Header | fs/graph/graph.go | Represents HTTP headers for Graph API requests |
| ResponseCache | fs/graph/response_cache.go | Caches API responses |
| Thumbnails | fs/graph/thumbnails.go | Handles thumbnail retrieval from OneDrive |
| MockGraph | fs/graph/mock_graph.go | Mocking functionality for testing |

## UI Class Diagram

| Design Element | Code Artifact | Description |
|----------------|---------------|-------------|
| LauncherApplication | cmd/onedriver-launcher/main.go | The main launcher application |
| MountRow | cmd/onedriver-launcher/main.go:newMountRow | Creates a row in the list box for a mountpoint |
| SettingsDialog | cmd/onedriver-launcher/main.go:newSettingsDialog | Creates the settings dialog |
| UIUtilities | ui/onedriver.go, ui/widgets.go | Utility functions for the UI |
| UIWidgets | ui/widgets.go | Custom GTK widgets for the UI |
| SystemdIntegration | ui/systemd/systemd.go | Functions for interacting with systemd |
| OnedriverCLI | cmd/onedriver/main.go | The main command-line interface |
| CommonConfig | cmd/common/config.go | Configuration shared between the launcher and CLI |

## Authentication Workflow Sequence Diagram

| Design Element | Code Artifact | Description |
|----------------|---------------|-------------|
| Initial Authentication | fs/graph/oauth2.go:Authenticate | Authenticates to OneDrive |
| GUI Authentication | fs/graph/oauth2_gtk.go:getAuthCodeGTK | Authenticates using a GTK-based browser window |
| Headless Authentication | fs/graph/oauth2_headless.go:getAuthCodeHeadless | Authenticates in terminal mode |
| Token Refresh | fs/graph/oauth2.go:refreshToken | Refreshes the authentication token when expired |
| API Request with Authentication | fs/graph/graph.go:Request | Makes authenticated requests to the Graph API |

## File Access Workflow Sequence Diagram

| Design Element | Code Artifact | Description |
|----------------|---------------|-------------|
| File Access (Cached) | fs/file_operations.go:Read | Reads file content from cache |
| File Access (Not Cached) | fs/file_operations.go:Read, fs/download_manager.go | Downloads file content from OneDrive |
| File Access (Large File) | fs/graph/drive_item.go:GetItemContentStream | Downloads large files in chunks |
| File Access (Offline Mode) | fs/file_operations.go:Read, fs/offline.go | Handles file access in offline mode |
| File Access (Thumbnails) | fs/file_operations_thumbnails.go, fs/fuse_thumbnail_handlers.go | Handles thumbnail access |
| File Access (Metadata) | fs/metadata_operations.go | Handles metadata access |
| File Access (Extended Attributes) | fs/xattr_operations.go | Handles extended attribute access |

## File Modification Workflow Sequence Diagram

| Design Element | Code Artifact | Description |
|----------------|---------------|-------------|
| File Modification (Small File) | fs/file_operations.go:Write, fs/upload_manager.go | Writes file content and uploads to OneDrive |
| File Modification (Large File) | fs/file_operations.go:Write, fs/upload_session.go | Uploads large files in chunks |
| File Modification (Offline Mode) | fs/file_operations.go:Write, fs/offline.go | Handles file modification in offline mode |
| File Modification (Conflict Resolution) | fs/upload_session.go:Upload | Resolves conflicts between local and server versions |
| File Modification (Metadata) | fs/metadata_operations.go | Handles metadata modification |
| File Modification (Extended Attributes) | fs/xattr_operations.go | Handles extended attribute modification |
| Directory Modification | fs/dir_operations.go | Handles directory creation, deletion, and modification |

## Delta Synchronization Workflow Sequence Diagram

| Design Element | Code Artifact | Description |
|----------------|---------------|-------------|
| Initial Delta Synchronization | fs/delta.go:DeltaLoop | Synchronizes changes from OneDrive |
| Periodic Delta Synchronization | fs/delta.go:DeltaLoop | Periodically checks for changes |
| Subscription-based Change Notification | fs/subscription.go | Handles change notifications from OneDrive |
| Sync Management | fs/sync.go | Manages synchronization between local and remote states |
| Cache Management | fs/cache.go | Manages caching of synchronized items |
| Offline Synchronization | fs/offline.go | Handles synchronization in offline mode |
