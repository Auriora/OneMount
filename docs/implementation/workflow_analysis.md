# OneDriver Workflow Analysis

This document presents the results of a reverse engineering effort to analyze the primary workflows in OneDriver (file upload, download, conflict resolution) and illustrate the sequence of invoked functions.

## Approach

To analyze the primary workflows in OneDriver, I took the following approach:

1. **Code Analysis**: Examined the codebase to understand the architecture and identify the key components involved in the primary workflows.
2. **Logging Instrumentation**: Created a script that configures OneDriver to log function calls at DEBUG level.
3. **Workflow Execution**: Executed the primary workflows (file upload, download, conflict resolution) to generate logs.
4. **Log Analysis**: Analyzed the logs to extract the sequence of function calls.
5. **Sequence Diagram Generation**: Created PlantUML sequence diagrams to visualize the workflows.

## Key Components

The analysis identified the following key components involved in the primary workflows:

- **Filesystem (FS)**: The core component that implements the FUSE filesystem interface.
- **Upload Manager (UM)**: Manages file uploads to OneDrive.
- **Download Manager (DM)**: Manages file downloads from OneDrive.
- **Delta Loop (Delta)**: Synchronizes changes between the local filesystem and OneDrive.
- **Content Cache (Cache)**: Caches file content locally.
- **Graph API Integration (Graph)**: Communicates with the Microsoft Graph API.

## Primary Workflows

### File Upload Workflow

When a file is created or modified in the OneDriver filesystem, the following sequence of function calls occurs:

```plantuml
@startuml
title OneDriver File Upload Workflow

actor User
participant "Filesystem" as FS
participant "UploadManager" as UM
participant "GraphAPI" as Graph

User -> FS: Write to file
FS -> FS: Inode.Write()
FS -> FS: Inode.SetContent()
FS -> FS: Inode.SetHasChanges(true)
FS -> UM: QueueUpload(inode)
UM -> UM: QueueUploadWithPriority(inode, priority)
UM -> UM: NewUploadSession(inode, data)
UM -> UM: Add to upload queue
UM --> FS: Return UploadSession
FS --> User: Return success

note over UM
  In background:
  uploadLoop processes the queue
end note

UM -> Graph: CreateUploadSession(item)
Graph --> UM: Return upload URL
UM -> Graph: UploadBytes(url, content)
Graph --> UM: Return DriveItem
UM -> FS: Update inode with new DriveItem
FS -> FS: Inode.SetHasChanges(false)

@enduml
```

Key observations:
- File uploads are queued and processed asynchronously by the upload manager.
- The upload manager uses the Microsoft Graph API to create an upload session and upload the file content.
- After a successful upload, the inode is updated with the new DriveItem information.

### File Download Workflow

When a file is accessed in the OneDriver filesystem and its content is not in the cache, the following sequence of function calls occurs:

```plantuml
@startuml
title OneDriver File Download Workflow

actor User
participant "Filesystem" as FS
participant "DownloadManager" as DM
participant "ContentCache" as Cache
participant "GraphAPI" as Graph

User -> FS: Open/Read file
FS -> Cache: Check if content is cached
Cache --> FS: Content not in cache
FS -> DM: QueueDownload(id)
DM -> DM: Create DownloadSession
DM -> DM: Add to download queue
DM --> FS: Return DownloadSession
FS -> DM: WaitForDownload(id)

note over DM
  In background:
  worker processes the queue
end note

DM -> DM: processDownload(id)
DM -> FS: Get inode path
DM -> Graph: DownloadContent(id)
Graph --> DM: Return content stream
DM -> Cache: Store content in cache
DM -> DM: Mark download as completed
DM --> FS: Download completed
FS -> Cache: Get content from cache
Cache --> FS: Return content
FS --> User: Return file content

@enduml
```

Key observations:
- File downloads are queued and processed by the download manager.
- The download manager uses the Microsoft Graph API to download the file content.
- Downloaded content is stored in the cache for future access.
- The filesystem waits for the download to complete before returning the content to the user.

### Conflict Resolution Workflow

When a conflict is detected between local and remote changes, the following sequence of function calls occurs:

```plantuml
@startuml
title OneDriver Conflict Resolution Workflow

participant "DeltaLoop" as Delta
participant "Filesystem" as FS
participant "GraphAPI" as Graph

Delta -> Delta: processDelta(delta)
Delta -> FS: Check for local changes
FS -> FS: Check for offline changes
FS -> FS: Check for pending uploads
FS --> Delta: Local changes exist
Delta -> FS: MarkFileConflict(id, message)
FS -> FS: Create FileStatus with StatusConflict
Delta -> Delta: Create conflict copy name
Delta -> FS: InsertChild(parentID, conflictInode)
FS -> FS: Add conflict copy to filesystem
Delta -> FS: Keep local version as is

note over FS
  User can now see both versions
  and resolve the conflict manually
end note

@enduml
```

Key observations:
- Conflicts are detected by the delta loop when processing changes from OneDrive.
- When a conflict is detected, the file is marked as having a conflict.
- A conflict copy of the remote file is created with a timestamp in the name.
- Both the local version and the conflict copy are kept, allowing the user to resolve the conflict manually.

## Conclusion

The analysis of the primary workflows in OneDriver reveals a well-designed system with clear separation of concerns:

1. **Asynchronous Processing**: File uploads and downloads are processed asynchronously, improving performance and user experience.
2. **Caching**: File content is cached locally, reducing the need for network requests.
3. **Conflict Handling**: Conflicts are detected and resolved by creating conflict copies, allowing users to decide how to resolve them.
4. **Component Separation**: The system is divided into well-defined components (filesystem, upload manager, download manager, etc.) with clear responsibilities.

The sequence diagrams provided in this document illustrate the flow of function calls in each of the primary workflows, providing a clear understanding of how OneDriver operates.