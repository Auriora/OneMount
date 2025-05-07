# Software Design Specification Traceability Matrix

This document provides a mapping between the numbered requirements in the Software Requirements Specification (SRS), the architectural elements described in the Software Architecture Specification (SAS), and their implementation in the Software Design Specification (SDS).

## Functional Requirements

### Filesystem Operations

| Requirement ID | Requirement Description | Architecture Specification Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| FR-FS-001 | The system shall mount OneDrive as a native Linux filesystem using FUSE. | SAS 3.1.2 (External Entities - Linux Filesystem), SAS 3.1.3 (System Interfaces - FUSE Interface), SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 5.1 (Key Architectural Decisions - Use of FUSE) | SDS 2.1 (Class Diagram - Filesystem class), SDS 2.2 (Component Diagram - Filesystem package), SDS 6.1 (Dependencies - go-fuse/v2) |
| FR-FS-002 | The system shall support standard file operations (read, write, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode), SAS 3.4.3 (Sequence Diagrams - File Access Workflow) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 3.1 (Sequence Diagram - File Access Workflow), SDS 4.2 (API Endpoints/Methods - GetItem, GetItemContent, Remove, Mkdir) |
| FR-FS-003 | The system shall support standard directory operations (list, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 4.2.3 (API Endpoints/Methods - GetItemChildren), SDS 4.2.4 (API Endpoints/Methods - Mkdir), SDS 4.2.5 (API Endpoints/Methods - Remove) |
| FR-FS-004 | The system shall download files on-demand when accessed rather than syncing all files. | SAS 1.3 (System Overview), SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.4.3 (Sequence Diagrams - File Access Workflow), SAS 4.2.2 (Scalability), SAS 5.1 (Key Architectural Decisions - On-demand file download) | SDS 2.1 (Class Diagram - DownloadManager class), SDS 3.1 (Sequence Diagram - File Access Not Cached), SDS 5.2.1 (Entity Definitions - Filesystem - downloads attribute) |
| FR-FS-005 | The system shall cache file metadata to improve performance. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy), SAS 5.1 (Key Architectural Decisions - Local caching) | SDS 2.1 (Class Diagram - LoopbackCache and ThumbnailCache classes), SDS 5.2.1 (Entity Definitions - Filesystem - metadata attribute), SDS 5.3 (Database Schema - Inodes table), SDS 6.2 (Performance Considerations - Caching) |
| FR-FS-006 | The system shall handle file conflicts between local and remote changes. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - FileStatus enum with StatusConflict), SDS 5.2.1 (Entity Definitions - Inode - hasChanges attribute), SDS 5.4 (Data Validation Rules) |
| FR-FS-007 | The system shall cache thumbnails for quick file previews. | SAS 3.5.1 (Deployment Diagram - thumbnails/), SAS 4.2.3 (Caching Strategy - Thumbnail cache for quick previews) | SDS 2.1 (Class Diagram - ThumbnailCache class), SDS 5.2.1 (Entity Definitions - Filesystem - thumbnails attribute), SDS 5.3 (Database Schema - Thumbnails table) |
| FR-FS-008 | The system shall use Microsoft Graph API's direct thumbnail endpoint to retrieve thumbnails without downloading the original file. | SAS 3.1.2 (External Entities - Microsoft OneDrive / Graph API), SAS 3.1.3 (System Interfaces - Microsoft Graph API Interface), SAS 4.2.1 (Performance Requirements), SAS 4.2.3 (Caching Strategy - Thumbnail cache for quick previews) | SDS 4.2.6 (API Endpoints/Methods - GetThumbnail), SDS 6.2 (Performance Considerations - Direct API Endpoints) |
| FR-FS-009 | The system shall use subscription-based change notifications to receive real-time updates from OneDrive. | SAS 3.4.1 (Runtime Processes), SAS 4.2.2 (Scalability), SAS 4.2.2.1 (Subscription-based Change Notification) | SDS 2.1 (Class Diagram - Subscription class), SDS 3.1 (Sequence Diagram - Change Notification Workflow), SDS 4.2.7 (API Endpoints/Methods - CreateSubscription) |

### Authentication

| Requirement ID | Requirement Description | Architecture Specification Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| FR-AUTH-001 | The system shall authenticate with Microsoft accounts using OAuth 2.0. | SAS 3.1.2 (External Entities - Microsoft OneDrive / Graph API), SAS 3.1.3 (System Interfaces - Microsoft Graph API Interface), SAS 3.2.3 (Key Abstractions - Auth), SAS 4.1.2 (Authentication and Authorization) | SDS 2.1 (Class Diagram - Auth class), SDS 3.1 (Sequence Diagram - Authentication Workflow), SDS 4.3 (Authentication and Authorization - OAuth2 for authentication) |
| FR-AUTH-002 | The system shall securely store authentication tokens. | SAS 3.5.1 (Deployment Diagram - auth_tokens.json), SAS 4.1.1 (Security Requirements), SAS 4.1.2 (Authentication and Authorization), SAS 4.1.3 (Data Protection) | SDS 2.1 (Class Diagram - Auth class with token attributes), SDS 4.3 (Authentication and Authorization - Secure storage of refresh tokens), SDS 6.3 (Security Considerations - Token Storage) |
| FR-AUTH-003 | The system shall automatically refresh authentication tokens when they expire. | SAS 3.2.3 (Key Abstractions - Auth), SAS 4.1.2 (Authentication and Authorization), SAS 5.3 (Quality Attribute Scenarios - Security Scenario) | SDS 2.1 (Class Diagram - Auth class with Refresh method), SDS 3.1 (Sequence Diagram - API Request with Authentication), SDS 4.3 (Authentication and Authorization - Automatic token refresh) |
| FR-AUTH-004 | The system shall support re-authentication when refresh tokens are invalid. | SAS 4.1.2 (Authentication and Authorization), SAS 5.3 (Quality Attribute Scenarios - Security Scenario) | SDS 3.1 (Sequence Diagram - Authentication Workflow with refresh failed path), SDS 4.3 (Authentication and Authorization) |

### Offline Functionality

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| FR-OFF-001 | The system shall provide access to previously accessed files when offline. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.3 (Other Crosscutting Concerns - Availability), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - Filesystem with offline attribute), SDS 3.1 (Sequence Diagram - File Access Offline Mode), SDS 5.2.1 (Entity Definitions - Filesystem - offline attribute) |
| FR-OFF-002 | The system shall cache file content for offline access. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy), SAS 5.1 (Key Architectural Decisions - Local caching) | SDS 2.1 (Class Diagram - LoopbackCache class), SDS 3.1 (Sequence Diagram - File Access Cached), SDS 5.2.1 (Entity Definitions - Filesystem - content attribute) |
| FR-OFF-003 | The system shall automatically detect network connectivity changes. | SAS 4.3 (Other Crosscutting Concerns - Availability - Automatic reconnection), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - Filesystem with offline attribute), SDS 6.4 (Error Handling - Retry Logic) |
| FR-OFF-004 | The system shall synchronize changes made offline when connectivity is restored. | SAS 4.3 (Other Crosscutting Concerns - Availability - Automatic reconnection), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - UploadManager class), SDS 5.2.1 (Entity Definitions - Inode - hasChanges attribute), SDS 6.4 (Error Handling - Retry Logic) |

### User Interface

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| FR-UI-001 | The system shall provide a command-line interface for mounting and configuration. | SAS 3.1.1 (System Context Diagram - Command Line Interface), SAS 3.2.1 (Logical Components - Command Line Interface) | SDS 2.2 (Component Diagram - Command Line package), SDS 6.1 (Dependencies - pflag for command-line argument parsing) |
| FR-UI-002 | The system shall provide a graphical user interface for mounting and configuration. | SAS 3.1.1 (System Context Diagram - UI Components), SAS 3.2.1 (Logical Components - User Interface) | SDS 2.2 (Component Diagram - User Interface package), SDS 6.1 (Dependencies - gotk3 for GUI components) |
| FR-UI-003 | The system shall display file status and synchronization information. | SAS 3.2.1 (Logical Components - User Interface), SAS 4.3 (Other Crosscutting Concerns - Usability) | SDS 2.1 (Class Diagram - FileStatus enum and FileStatusInfo class), SDS 5.2.1 (Entity Definitions - Filesystem - statuses attribute) |
| FR-UI-004 | The system shall provide system tray integration for status indication. | SAS 3.1.2 (External Entities - Desktop Environment), SAS 3.2.1 (Logical Components - User Interface) | SDS 2.2 (Component Diagram - User Interface package), SDS 6.1 (Dependencies - gotk3 for GUI components) |

### Statistics and Analysis

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| FR-STAT-001 | The system shall provide a statistics command to analyze OneDrive content metadata. | SAS 3.2.1 (Logical Components - Command Line Interface) | *Not yet implemented* |
| FR-STAT-002 | The system shall analyze file type distribution in the statistics command. | SAS 3.2.1 (Logical Components - Command Line Interface) | *Not yet implemented* |
| FR-STAT-003 | The system shall analyze directory depth statistics in the statistics command. | SAS 3.2.1 (Logical Components - Command Line Interface) | *Not yet implemented* |
| FR-STAT-004 | The system shall analyze file size distribution in the statistics command. | SAS 3.2.1 (Logical Components - Command Line Interface) | *Not yet implemented* |
| FR-STAT-005 | The system shall analyze file age information in the statistics command. | SAS 3.2.1 (Logical Components - Command Line Interface) | *Not yet implemented* |

### Integration with External Systems

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| FR-INT-001 | The system shall provide a D-Bus interface for file status updates. | SAS 3.1.3 (System Interfaces - D-Bus Interface), SAS 3.4.2 (Process Communication - D-Bus) | SDS 2.1 (Class Diagram - Filesystem with dbusServer attribute), SDS 6.1 (Dependencies - dbus for D-Bus integration) |
| FR-INT-002 | The system shall expose methods for getting file status through the D-Bus interface. | SAS 3.1.3 (System Interfaces - D-Bus Interface), SAS 3.4.2 (Process Communication - D-Bus) | SDS 2.1 (Class Diagram - FileStatusInfo class), SDS 5.2.1 (Entity Definitions - Filesystem - statuses attribute) |
| FR-INT-003 | The system shall emit signals for file status changes through the D-Bus interface. | SAS 3.1.3 (System Interfaces - D-Bus Interface), SAS 3.4.2 (Process Communication - D-Bus) | SDS 2.1 (Class Diagram - FileStatusInfo class), SDS 5.2.1 (Entity Definitions - Filesystem - statuses attribute) |
| FR-INT-004 | The system shall integrate with the Nemo file manager to display OneDrive in the sidebar. | SAS 3.1.2 (External Entities - Desktop Environment), SAS 4.3 (Other Crosscutting Concerns - Usability - Seamless integration with desktop environment) | SDS 2.2 (Component Diagram - Nemo Extension), SDS 6.1 (Dependencies - Nemo file manager) |
| FR-INT-005 | The system shall display file status icons in the Nemo file manager. | SAS 3.1.2 (External Entities - Desktop Environment), SAS 4.3 (Other Crosscutting Concerns - Usability - Seamless integration with desktop environment) | SDS 2.1 (Class Diagram - FileStatus enum), SDS 2.2 (Component Diagram - Nemo Extension) |
| FR-INT-006 | The system shall fall back to extended attributes if D-Bus is not available. | SAS 3.4.2 (Process Communication) | SDS 2.1 (Class Diagram - Inode with xattrs attribute), SDS 5.2.2 (Entity Definitions - Inode - xattrs attribute) |

### Developer Tools

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| FR-DEV-001 | The system shall provide a method logging framework for debugging. | SAS 3.3.3 (Development Environment - Error Handling - Use structured logging with zerolog), SAS 4.2.4 (Performance Monitoring - Structured logging of operation times) | SDS 2.1 (Class Diagram - MethodDecorator), SDS 6.1 (Dependencies - zerolog), SDS 6.4 (Error Handling - Structured Logging) |
| FR-DEV-002 | The system shall log method entry and exit with parameters and return values. | SAS 3.3.3 (Development Environment - Error Handling - Use structured logging with zerolog), SAS 4.2.4 (Performance Monitoring - Structured logging of operation times) | SDS 2.1 (Class Diagram - MethodDecorator), SDS 6.4 (Error Handling - Parameter Logging) |
| FR-DEV-003 | The system shall include execution duration in method logs. | SAS 3.3.3 (Development Environment - Error Handling - Use structured logging with zerolog), SAS 4.2.4 (Performance Monitoring - Structured logging of operation times) | SDS 2.1 (Class Diagram - MethodDecorator), SDS 6.4 (Error Handling - Execution Timing) |

## Non-Functional Requirements

### Performance

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| NFR-PERF-001 | The system shall minimize network requests to improve performance. | SAS 3.2.1 (Logical Components - Cache Management), SAS 4.2.1 (Performance Requirements), SAS 4.2.3 (Caching Strategy) | SDS 2.1 (Class Diagram - ResponseCache class), SDS 6.2 (Performance Considerations - Caching), SDS 6.2 (Performance Considerations - Response Caching) |
| NFR-PERF-002 | The system shall use concurrent operations where appropriate. | SAS 3.3.3 (Development Environment - Performance - Use concurrent operations where appropriate), SAS 3.4.1 (Runtime Processes), SAS 4.2.2 (Scalability - Concurrent file transfers) | SDS 2.1 (Class Diagram - UploadManager and DownloadManager with worker goroutines), SDS 6.2 (Performance Considerations - Concurrent Operations) |
| NFR-PERF-003 | The system shall implement efficient caching to reduce API calls. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy) | SDS 2.1 (Class Diagram - LoopbackCache, ThumbnailCache, and ResponseCache classes), SDS 5.3 (Database Schema), SDS 6.2 (Performance Considerations - Caching) |
| NFR-PERF-004 | The system shall support chunked downloads for large files. | SAS 3.2.3 (Key Abstractions - UploadManager/DownloadManager - Handle chunking for large files) | SDS 3.1 (Sequence Diagram - File Access Large File), SDS 4.2.2 (API Endpoints/Methods - GetItemContentStream), SDS 6.2 (Performance Considerations - Chunked Transfers) |

### Security

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| NFR-SEC-001 | The system shall store authentication tokens with appropriate file permissions. | SAS 3.5.1 (Deployment Diagram - auth_tokens.json), SAS 4.1.1 (Security Requirements), SAS 4.1.3 (Data Protection - Local file permissions for cached content) | SDS 4.3 (Authentication and Authorization - Secure storage of refresh tokens), SDS 6.3 (Security Considerations - Token Storage) |
| NFR-SEC-002 | The system shall use HTTPS for all API communications. | SAS 3.5.1 (Deployment Diagram - HTTPS), SAS 4.1.3 (Data Protection - HTTPS for all communication with Microsoft Graph API) | SDS 4.3 (Authentication and Authorization), SDS 6.3 (Security Considerations - HTTPS) |
| NFR-SEC-003 | The system shall not expose authentication tokens to non-privileged users. | SAS 4.1.1 (Security Requirements), SAS 4.1.3 (Data Protection), SAS 4.1.4 (Security Patterns - Principle of least privilege) | SDS 4.3 (Authentication and Authorization), SDS 6.3 (Security Considerations - Minimal Permissions) |

### Reliability

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference |
|----------------|-------------------------|--------------------------------|--------------------------------|
| NFR-REL-001 | The system shall handle network errors and retry operations. | SAS 3.2.3 (Key Abstractions - UploadManager/DownloadManager - Implement retry logic and error handling), SAS 4.3 (Other Crosscutting Concerns - Reliability - Retry logic for network operations) | SDS 2.1 (Class Diagram - UploadManager and DownloadManager classes), SDS 6.4 (Error Handling - Retry Logic) |
| NFR-REL-002 | The system shall recover gracefully from crashes. | SAS 4.3 (Other Crosscutting Concerns - Availability - Crash recovery architecture) | SDS 2.1 (Class Diagram - Filesystem with db attribute), SDS 5.3 (Database Schema - Persistent storage), SDS 6.4 (Error Handling - Crash Recovery) |
| NFR-REL-003 | The system shall maintain data integrity during synchronization. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes) | SDS 2.1 (Class Diagram - FileStatus enum with StatusConflict), SDS 5.4 (Data Validation Rules), SDS 6.4 (Error Handling - Graceful Degradation) |
| NFR-REL-004 | The system shall handle API rate limiting gracefully. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Graceful handling of API rate limits) | SDS 4.4 (Rate Limiting and Quotas), SDS 6.4 (Error Handling - Retry Logic with exponential backoff) |
| NFR-REL-005 | The system shall use QuickXORHash for file integrity verification. | SAS 3.3.1 (Module Organization - fs/graph/quickxorhash), SAS 3.2.3 (Key Abstractions - UploadManager/DownloadManager), SAS 4.3 (Other Crosscutting Concerns - Reliability - Data integrity verification) | SDS 2.1 (Class Diagram - UploadManager with hash verification), SDS 4.2.8 (API Endpoints/Methods - VerifyHash), SDS 6.3 (Security Considerations - Data Integrity) |

## Architectural Decisions

| Architecture Decision ID | Description | Design Specification Reference |
|--------------------------|-------------|--------------------------------|
| AD-001 | Use of FUSE for filesystem implementation | SDS 2.1 (Class Diagram - Filesystem class), SDS 6.1 (Dependencies - go-fuse/v2) |
| AD-002 | On-demand file download instead of full sync | SDS 2.1 (Class Diagram - DownloadManager class), SDS 3.1 (Sequence Diagram - File Access Not Cached) |
| AD-003 | Local caching of metadata and content | SDS 2.1 (Class Diagram - LoopbackCache and ThumbnailCache classes), SDS 5.3 (Database Schema), SDS 6.2 (Performance Considerations - Caching) |
| AD-004 | Use of BBolt for embedded database | SDS 2.1 (Class Diagram - Filesystem with db attribute), SDS 5.3 (Database Schema), SDS 6.1 (Dependencies - bbolt) |
| AD-005 | Use of GTK3 for GUI components | SDS 2.2 (Component Diagram - User Interface package), SDS 6.1 (Dependencies - gotk3) |
| AD-006 | Structured logging with zerolog | SDS 6.1 (Dependencies - zerolog), SDS 6.4 (Error Handling - Structured Logging) |
| AD-007 | D-Bus interface for file status updates | SDS 2.1 (Class Diagram - Filesystem with dbusServer attribute), SDS 6.1 (Dependencies - dbus) |
| AD-008 | Concurrent operations for uploads and downloads | SDS 2.1 (Class Diagram - UploadManager and DownloadManager with worker goroutines), SDS 6.2 (Performance Considerations - Concurrent Operations) |
