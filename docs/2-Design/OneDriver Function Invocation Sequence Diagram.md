```plantuml
@startuml
title OneDriver Function Invocation Sequence

actor User
participant "Main" as Main
participant "Filesystem" as FS
participant "DeltaLoop" as Delta
participant "ContentCache" as Cache
participant "DownloadManager" as DM
participant "UploadManager" as UM
participant "GraphAPI" as Graph

== Initialization ==
User -> Main: Start Application
Main -> Main: setupFlags()
Main -> Main: setupLogging()
Main -> FS: initializeFilesystem()
FS -> FS: Initialize shared HTTP client
FS -> FS: Start D-Bus server
FS -> Delta: Start delta goroutine
FS -> Cache: Initialize content cache
FS -> Cache: Start content cache cleanup routine
FS -> FS: Serving filesystem

== Normal Operation ==
Delta -> Graph: Fetch deltas
Graph --> Delta: Return deltas
Delta -> FS: Process deltas
FS -> FS: Found content in cache
FS -> FS: Open file

== File Synchronization ==
Delta -> Graph: Fetch deltas
Graph --> Delta: Return deltas (with changes)
Delta -> FS: Process deltas
FS -> FS: Overwriting local item
FS -> DM: Queue file for download
DM -> Graph: Download file
Graph --> DM: File content
DM -> FS: File download completed

== Error Handling ==
Delta -> Graph: Fetch deltas
Graph --> Delta: Error (Auth empty)
Delta -> FS: Error fetching children page
FS -> FS: Switch to offline mode

== Creating Files from Deltas ==
Delta -> FS: Creating inode from delta
FS -> FS: Create new file/folder

== Shutdown ==
User -> Main: Send interrupt signal
Main -> FS: Signal received, cleaning up
FS -> Cache: Stopping cache cleanup routine
Cache --> FS: Cache cleanup routine stopped
FS -> Delta: Stopping delta loop
Delta --> FS: Delta loop stopped successfully
FS -> DM: Stopping download manager
DM --> FS: Download manager stopped successfully
FS -> UM: Stopping upload manager
UM --> FS: Upload manager stopped successfully
FS -> FS: Wait for resources to be released
FS -> Main: Unmount filesystem

@enduml
```

# OneDriver Function Invocation Sequence Diagram

Based on the analysis of the onedriver.log file, I've created a sequence diagram that illustrates the main function invocation flows in the OneDriver application.

## Key Components

The diagram shows interactions between these main components:

1. **Main** - The application entry point
2. **Filesystem** - Core filesystem implementation
3. **DeltaLoop** - Handles synchronization with OneDrive
4. **ContentCache** - Manages local file caching
5. **DownloadManager** - Handles file downloads
6. **UploadManager** - Handles file uploads
7. **GraphAPI** - Microsoft Graph API interface

## Main Flows

The sequence diagram captures these key operational flows:

### Initialization
- Application startup and configuration
- Filesystem initialization
- Starting background services (D-Bus, delta sync, cache cleanup)

### Normal Operation
- Delta synchronization with OneDrive
- File access from cache

### File Synchronization
- Detecting changes from OneDrive
- Downloading modified files
- Updating local cache

### Error Handling
- Authentication errors
- Switching to offline mode

### File Creation
- Creating new inodes from delta information

### Shutdown
- Signal handling
- Graceful shutdown of all components
- Resource cleanup

## Implementation Details

The code analysis reveals that OneDriver uses:

1. A delta-based synchronization mechanism to efficiently track changes
2. A local content cache to minimize network requests
3. Separate managers for uploads and downloads
4. Background goroutines for continuous synchronization
5. Graceful error handling with offline mode support

This sequence diagram provides a high-level overview of how the different components interact during the application lifecycle, from startup to shutdown.