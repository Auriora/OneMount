

# OneDriver Use Cases with UML Diagrams

## UML Use Case Diagram

```plantuml
@startuml OneDriver Use Case Diagram

' Define actors
actor User
actor "Microsoft OneDrive API" as OneDriveAPI

' Define use cases
rectangle "OneDriver System" {
  usecase "Upload File" as Upload
  usecase "Download File" as Download
  usecase "Resolve Conflicts" as ConflictRes
}

' Define relationships
User --> Upload
User --> Download
User --> ConflictRes

Upload --> OneDriveAPI
Download --> OneDriveAPI
ConflictRes --> OneDriveAPI

@enduml
```

This PlantUML diagram represents the same use case diagram that was previously shown in Mermaid format. It shows:

1. Two actors: User and Microsoft OneDrive API
2. Three use cases within the OneDriver System: Upload File, Download File, and Resolve Conflicts
3. The relationships between actors and use cases, showing which actor interacts with which use case

The diagram maintains the same structure and relationships as the original Mermaid diagram while using PlantUML's specific syntax for use case diagrams.

## Use Case 1: Upload File

| Field | Description |
|-------|-------------|
| **Use Case ID** | UC-01 |
| **Name** | Upload File to OneDrive |
| **Actors** | User, Microsoft OneDrive API |
| **Preconditions** | 1. User is authenticated with Microsoft account<br>2. OneDriver is mounted and running<br>3. User has write permissions for the target location<br>4. Network connection is available |
| **Postconditions** | 1. File is successfully uploaded to OneDrive<br>2. Local file system reflects the upload status<br>3. File metadata is updated in the local cache |
| **Main Flow** | 1. User creates or modifies a file in the OneDriver mounted directory<br>2. System detects the file change event<br>3. System queues the file for upload<br>4. System uploads the file to OneDrive using Microsoft Graph API<br>5. System updates the local cache with new file metadata<br>6. System notifies the user of successful upload |
| **Alternative Flows** | **A1: Network Unavailable**<br>1. System detects network is unavailable<br>2. System marks file for deferred upload<br>3. System stores file changes in local cache<br>4. System periodically attempts to reconnect<br>5. When connection is restored, system resumes upload<br><br>**A2: Upload Quota Exceeded**<br>1. OneDrive API returns quota exceeded error<br>2. System notifies user of quota limitation<br>3. System marks file for retry when space is available<br><br>**A3: Authentication Failure**<br>1. System detects authentication token is expired<br>2. System attempts to refresh authentication<br>3. If refresh fails, system prompts user to re-authenticate<br>4. Upload process resumes after successful authentication |

## Use Case 2: Download File

| Field | Description |
|-------|-------------|
| **Use Case ID** | UC-02 |
| **Name** | Download File from OneDrive |
| **Actors** | User, Microsoft OneDrive API |
| **Preconditions** | 1. User is authenticated with Microsoft account<br>2. OneDriver is mounted and running<br>3. File exists in OneDrive<br>4. Network connection is available |
| **Postconditions** | 1. File is successfully downloaded to local filesystem<br>2. File metadata is updated in local cache<br>3. File is accessible to the user |
| **Main Flow** | 1. User attempts to access a file in the OneDriver mounted directory<br>2. System checks if file content is available locally<br>3. If not available, system requests file content from OneDrive API<br>4. System downloads the file content<br>5. System stores the file in local cache<br>6. System presents the file to the user<br>7. System updates access metadata |
| **Alternative Flows** | **A1: Network Unavailable**<br>1. System detects network is unavailable<br>2. System checks if file is available in offline cache<br>3. If available in cache, system serves cached version<br>4. If not in cache, system notifies user that file is unavailable offline<br><br>**A2: File No Longer Exists Remotely**<br>1. OneDrive API returns file not found error<br>2. System removes file reference from local directory<br>3. System notifies user that file no longer exists<br><br>**A3: Insufficient Local Storage**<br>1. System detects insufficient storage for download<br>2. System notifies user of storage limitation<br>3. System suggests clearing cache or freeing disk space<br>4. Download is paused until sufficient space is available |

## Use Case 3: Resolve File Conflicts

| Field | Description |
|-------|-------------|
| **Use Case ID** | UC-03 |
| **Name** | Resolve File Conflicts |
| **Actors** | User, Microsoft OneDrive API |
| **Preconditions** | 1. User is authenticated with Microsoft account<br>2. OneDriver is mounted and running<br>3. Same file has been modified both locally and remotely<br>4. Network connection is available |
| **Postconditions** | 1. Conflict is resolved according to user's decision or automatic policy<br>2. File system reflects the resolved state<br>3. OneDrive is synchronized with the resolution |
| **Main Flow** | 1. System detects conflicting versions of the same file<br>2. System compares local and remote file metadata (modification times, sizes)<br>3. System applies automatic conflict resolution policy if configured<br>4. If manual resolution is required, system notifies user of conflict<br>5. User selects preferred version or requests to keep both<br>6. System implements the resolution decision<br>7. System synchronizes the resolved state with OneDrive<br>8. System updates local cache with resolution metadata |
| **Alternative Flows** | **A1: Keep Both Versions**<br>1. User chooses to keep both versions<br>2. System renames the conflicting file with suffix indicating local/remote origin<br>3. System uploads/retains both versions<br>4. System updates metadata for both files<br><br>**A2: Automatic Resolution Based on Policy**<br>1. System applies pre-configured conflict resolution policy<br>2. System resolves conflict without user intervention<br>3. System logs the automatic resolution action<br>4. System synchronizes the resolved state<br><br>**A3: Merge Changes**<br>1. For supported file types, system offers to merge changes<br>2. User reviews and confirms merge<br>3. System creates merged version<br>4. System uploads merged version to OneDrive<br>5. System updates local cache with merged file |

These use cases provide a comprehensive overview of the core functionality in OneDriver, focusing on file operations and conflict management. The structured format follows industry standards for use case documentation, making it easy to understand the system's behavior from a user perspective.