# onedriver Architecture Documentation

## Overview

onedriver is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing the entire OneDrive content. This document explains the architecture of onedriver, including the relationships between its major components.

## High-Level Architecture

onedriver is built on several key technologies:

1. **FUSE (Filesystem in Userspace)**: Allows implementing a filesystem in user space rather than kernel space
2. **Microsoft Graph API**: Provides access to OneDrive data and operations
3. **BBolt**: An embedded key-value database for caching filesystem metadata
4. **Go**: The implementation language, providing concurrency features via goroutines and channels

The application is structured into several logical components:

```
┌─────────────────────────────────────────────────────────────────┐
│                        onedriver                                │
│                                                                 │
│  ┌───────────┐    ┌───────────┐    ┌───────────────────────┐    │
│  │    UI     │    │  Command  │    │     Filesystem        │    │
│  │           │◄───┤   Line    │◄───┤                       │    │
│  │ (GTK3)    │    │ Interface │    │  (FUSE Implementation)│    │
│  └───────────┘    └───────────┘    └───────────────────────┘    │
│                                              ▲                  │
│                                              │                  │
│                                     ┌────────┴─────────┐        │
│                                     │                  │        │
│                                     ▼                  ▼        │
│                              ┌──────────────┐  ┌─────────────┐  │
│                              │  Graph API   │  │   Cache     │  │
│                              │  Integration │  │  Management │  │
│                              └──────────────┘  └─────────────┘  │
│                                     ▲                ▲          │
│                                     │                │          │
│                                     ▼                ▼          │
│                              ┌──────────────┐  ┌─────────────┐  │
│                              │   Network    │  │   Local     │  │
│                              │    Layer     │  │  Storage    │  │
│                              └──────────────┘  └─────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Component Descriptions

### Filesystem (fs package)

The core of onedriver is the Filesystem implementation, which provides a FUSE-compatible interface to OneDrive. Key components include:

- **Inode**: Represents files and directories in the filesystem
- **Filesystem**: The main FUSE implementation that handles file operations
- **Cache**: Manages local caching of file content and metadata
- **Upload/Download Managers**: Handle file transfers to and from OneDrive

### Graph API Integration (fs/graph package)

This component handles communication with Microsoft's Graph API:

- **Auth**: Manages OAuth2 authentication with Microsoft
- **DriveItem**: Represents files and folders in OneDrive
- **API Client**: Handles HTTP requests to the Graph API endpoints

### Cache Management

onedriver uses a sophisticated caching system to minimize network requests:

- **Metadata Cache**: Stores file and directory metadata in memory and in a BBolt database
- **Content Cache**: Stores file contents on the local filesystem
- **Delta Synchronization**: Uses the Graph API's delta query to efficiently sync changes

### Command Line Interface (cmd package)

Provides the user interface for mounting, unmounting, and configuring onedriver:

- **Argument Parsing**: Handles command-line flags and arguments
- **Configuration**: Manages user configuration settings
- **Signal Handling**: Manages graceful shutdown on system signals

### UI (ui package)

Provides a graphical interface for onedriver:

- **GTK3 Interface**: Shows status and allows basic operations
- **Systemd Integration**: Manages the onedriver service

## Key Interactions

### Filesystem Operations

1. When a file is accessed:
   - The filesystem checks if the file is in the local cache
   - If not, it requests the file from OneDrive via the Graph API
   - The file is cached locally for future access

2. When a file is modified:
   - Changes are made to the local cache
   - On flush/close, changes are uploaded to OneDrive
   - If offline, changes are tracked and synchronized when online

### Authentication Flow

1. User initiates authentication
2. OAuth2 flow opens a browser window for Microsoft login
3. After successful login, an access token is obtained
4. The token is refreshed automatically when needed

### Delta Synchronization

1. Periodically, onedriver requests changes from OneDrive using a delta link
2. Changes are applied to the local cache
3. Conflicts are resolved based on modification times and other heuristics

## Offline Mode

onedriver supports working offline:

1. When network connectivity is lost, onedriver switches to offline mode
2. Users can continue to access cached files and make changes
3. Changes are tracked and synchronized when connectivity is restored

## Performance Considerations

- **On-demand downloading**: Files are only downloaded when accessed, saving bandwidth and storage
- **Caching**: Frequently accessed files remain in the cache to improve performance
- **Concurrent operations**: Upload and download operations run concurrently for better throughput

## Security

- **Authentication**: Uses OAuth2 for secure authentication with Microsoft
- **Token storage**: Access tokens are stored securely on the local filesystem
- **Permissions**: File permissions are mapped between OneDrive and the local filesystem

## Extensibility

The architecture is designed to be extensible:

- **Modular design**: Components are loosely coupled for easier maintenance and extension
- **Interface-based**: Key components implement interfaces that can be replaced or mocked for testing
- **Configuration options**: Many behaviors can be customized through configuration