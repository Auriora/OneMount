# Test Cases Traceability Matrix

This document provides a mapping between the numbered requirements in the Software Requirements Specification (SRS), the architectural elements described in the Software Architecture Specification (SAS), the design elements in the Software Design Specification (SDS), and the test cases that verify them.

## Functional Requirements

### Filesystem Operations

| Requirement ID | Requirement Description | Architecture Specification Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| FR-FS-001 | The system shall mount OneDrive as a native Linux filesystem using FUSE. | SAS 3.1.2 (External Entities - Linux Filesystem), SAS 3.1.3 (System Interfaces - FUSE Interface), SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 5.1 (Key Architectural Decisions - Use of FUSE) | SDS 2.1 (Class Diagram - Filesystem class), SDS 2.2 (Component Diagram - Filesystem package), SDS 6.1 (Dependencies - go-fuse/v2) | UT-01, ST-01, ST-02 |
| FR-FS-002 | The system shall support standard file operations (read, write, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode), SAS 3.4.3 (Sequence Diagrams - File Access Workflow) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 3.1 (Sequence Diagram - File Access Workflow), SDS 4.2 (API Endpoints/Methods - GetItem, GetItemContent, Remove, Mkdir) | UT-02, UT-05, ST-06, ST-07, ST-08, ST-10 |
| FR-FS-003 | The system shall support standard directory operations (list, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 4.2.3 (API Endpoints/Methods - GetItemChildren), SDS 4.2.4 (API Endpoints/Methods - Mkdir), SDS 4.2.5 (API Endpoints/Methods - Remove) | ST-01, ST-02, ST-05, ST-07, ST-12 |
| FR-FS-004 | The system shall download files on-demand when accessed rather than syncing all files. | SAS 1.3 (System Overview), SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.4.3 (Sequence Diagrams - File Access Workflow), SAS 4.2.2 (Scalability), SAS 5.1 (Key Architectural Decisions - On-demand file download) | SDS 2.1 (Class Diagram - DownloadManager class), SDS 3.1 (Sequence Diagram - File Access Not Cached), SDS 5.2.1 (Entity Definitions - Filesystem - downloads attribute) | UT-05, ST-08 |
| FR-FS-005 | The system shall cache file metadata to improve performance. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy), SAS 5.1 (Key Architectural Decisions - Local caching) | SDS 2.1 (Class Diagram - LoopbackCache and ThumbnailCache classes), SDS 5.2.1 (Entity Definitions - Filesystem - metadata attribute), SDS 5.3 (Database Schema - Inodes table), SDS 6.2 (Performance Considerations - Caching) | UT-01, IT-01 |
| FR-FS-006 | The system shall handle file conflicts between local and remote changes. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - FileStatus enum with StatusConflict), SDS 5.2.1 (Entity Definitions - Inode - hasChanges attribute), SDS 5.4 (Data Validation Rules) | UT-03, IT-05 |

### Authentication

| Requirement ID | Requirement Description | Architecture Specification Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| FR-AUTH-001 | The system shall authenticate with Microsoft accounts using OAuth 2.0. | SAS 3.1.2 (External Entities - Microsoft OneDrive / Graph API), SAS 3.1.3 (System Interfaces - Microsoft Graph API Interface), SAS 3.2.3 (Key Abstractions - Auth), SAS 4.1.2 (Authentication and Authorization) | SDS 2.1 (Class Diagram - Auth class), SDS 3.1 (Sequence Diagram - Authentication Workflow), SDS 4.3 (Authentication and Authorization - OAuth2 for authentication) | SEC-01, SEC-03, SEC-05 |
| FR-AUTH-002 | The system shall securely store authentication tokens. | SAS 3.5.1 (Deployment Diagram - auth_tokens.json), SAS 4.1.1 (Security Requirements), SAS 4.1.2 (Authentication and Authorization), SAS 4.1.3 (Data Protection) | SDS 2.1 (Class Diagram - Auth class with token attributes), SDS 4.3 (Authentication and Authorization - Secure storage of refresh tokens), SDS 6.3 (Security Considerations - Token Storage) | SEC-04, SEC-05 |
| FR-AUTH-003 | The system shall automatically refresh authentication tokens when they expire. | SAS 3.2.3 (Key Abstractions - Auth), SAS 4.1.2 (Authentication and Authorization), SAS 5.3 (Quality Attribute Scenarios - Security Scenario) | SDS 2.1 (Class Diagram - Auth class with Refresh method), SDS 3.1 (Sequence Diagram - API Request with Authentication), SDS 4.3 (Authentication and Authorization - Automatic token refresh) | SEC-02 |
| FR-AUTH-004 | The system shall support re-authentication when refresh tokens are invalid. | SAS 4.1.2 (Authentication and Authorization), SAS 5.3 (Quality Attribute Scenarios - Security Scenario) | SDS 3.1 (Sequence Diagram - Authentication Workflow with refresh failed path), SDS 4.3 (Authentication and Authorization) | SEC-01, SEC-05 |

### Offline Functionality

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| FR-OFF-001 | The system shall provide access to previously accessed files when offline. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.3 (Other Crosscutting Concerns - Availability), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - Filesystem with offline attribute), SDS 3.1 (Sequence Diagram - File Access Offline Mode), SDS 5.2.1 (Entity Definitions - Filesystem - offline attribute) | IT-01 |
| FR-OFF-002 | The system shall cache file content for offline access. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy), SAS 5.1 (Key Architectural Decisions - Local caching) | SDS 2.1 (Class Diagram - LoopbackCache class), SDS 3.1 (Sequence Diagram - File Access Cached), SDS 5.2.1 (Entity Definitions - Filesystem - content attribute) | IT-01, IT-06 |
| FR-OFF-003 | The system shall automatically detect network connectivity changes. | SAS 4.3 (Other Crosscutting Concerns - Availability - Automatic reconnection), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - Filesystem with offline attribute), SDS 6.4 (Error Handling - Retry Logic) | IT-05 |
| FR-OFF-004 | The system shall synchronize changes made offline when connectivity is restored. | SAS 4.3 (Other Crosscutting Concerns - Availability - Automatic reconnection), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - UploadManager class), SDS 5.2.1 (Entity Definitions - Inode - hasChanges attribute), SDS 6.4 (Error Handling - Retry Logic) | IT-05, IT-06 |

### File System Operations

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| FR-FS-002 | The system shall support standard file operations (read, write, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode), SAS 3.4.3 (Sequence Diagrams - File Access Workflow) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 3.1 (Sequence Diagram - File Access Workflow), SDS 4.2 (API Endpoints/Methods - GetItem, GetItemContent, Remove, Mkdir) | ST-03, ST-04, ST-06, ST-07, ST-08, ST-09, ST-10, ST-11 |
| FR-FS-003 | The system shall support standard directory operations (list, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 4.2.3 (API Endpoints/Methods - GetItemChildren), SDS 4.2.4 (API Endpoints/Methods - Mkdir), SDS 4.2.5 (API Endpoints/Methods - Remove) | ST-05, ST-12 |
| FR-FS-004 | The system shall download files on-demand when accessed rather than syncing all files. | SAS 1.3 (System Overview), SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.4.3 (Sequence Diagrams - File Access Workflow), SAS 4.2.2 (Scalability), SAS 5.1 (Key Architectural Decisions - On-demand file download) | SDS 2.1 (Class Diagram - DownloadManager class), SDS 3.1 (Sequence Diagram - File Access Not Cached), SDS 5.2.1 (Entity Definitions - Filesystem - downloads attribute) | UT-05, ST-08 |

## Non-Functional Requirements

### Performance

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| NFR-PERF-001 | The system shall minimize network requests to improve performance. | SAS 3.2.1 (Logical Components - Cache Management), SAS 4.2.1 (Performance Requirements), SAS 4.2.3 (Caching Strategy) | SDS 2.1 (Class Diagram - ResponseCache class), SDS 6.2 (Performance Considerations - Caching), SDS 6.2 (Performance Considerations - Response Caching) | UT-01, UT-05, IT-01 |
| NFR-PERF-002 | The system shall use concurrent operations where appropriate. | SAS 3.3.3 (Development Environment - Performance - Use concurrent operations where appropriate), SAS 3.4.1 (Runtime Processes), SAS 4.2.2 (Scalability - Concurrent file transfers) | SDS 2.1 (Class Diagram - UploadManager and DownloadManager with worker goroutines), SDS 6.2 (Performance Considerations - Concurrent Operations) | UT-02, UT-03, UT-04, UT-05 |
| NFR-PERF-003 | The system shall implement efficient caching to reduce API calls. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy) | SDS 2.1 (Class Diagram - LoopbackCache, ThumbnailCache, and ResponseCache classes), SDS 5.3 (Database Schema), SDS 6.2 (Performance Considerations - Caching) | UT-01, UT-05, IT-01 |
| NFR-PERF-004 | The system shall support chunked downloads for large files. | SAS 3.2.3 (Key Abstractions - UploadManager/DownloadManager - Handle chunking for large files) | SDS 3.1 (Sequence Diagram - File Access Large File), SDS 4.2.2 (API Endpoints/Methods - GetItemContentStream), SDS 6.2 (Performance Considerations - Chunked Transfers) | UT-04 |

### Security

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| NFR-SEC-001 | The system shall store authentication tokens with appropriate file permissions. | SAS 3.5.1 (Deployment Diagram - auth_tokens.json), SAS 4.1.1 (Security Requirements), SAS 4.1.3 (Data Protection - Local file permissions for cached content) | SDS 4.3 (Authentication and Authorization - Secure storage of refresh tokens), SDS 6.3 (Security Considerations - Token Storage) | SEC-04 |
| NFR-SEC-002 | The system shall use HTTPS for all API communications. | SAS 3.5.1 (Deployment Diagram - HTTPS), SAS 4.1.3 (Data Protection - HTTPS for all communication with Microsoft Graph API) | SDS 4.3 (Authentication and Authorization), SDS 6.3 (Security Considerations - HTTPS) | SEC-01, SEC-02, SEC-03 |
| NFR-SEC-003 | The system shall not expose authentication tokens to non-privileged users. | SAS 4.1.1 (Security Requirements), SAS 4.1.3 (Data Protection), SAS 4.1.4 (Security Patterns - Principle of least privilege) | SDS 4.3 (Authentication and Authorization), SDS 6.3 (Security Considerations - Minimal Permissions) | SEC-04 |

### Reliability

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| NFR-REL-001 | The system shall handle network errors and retry operations. | SAS 3.2.3 (Key Abstractions - UploadManager/DownloadManager - Implement retry logic and error handling), SAS 4.3 (Other Crosscutting Concerns - Reliability - Retry logic for network operations) | SDS 2.1 (Class Diagram - UploadManager and DownloadManager classes), SDS 6.4 (Error Handling - Retry Logic) | IT-01, IT-02, IT-03, IT-04, IT-05 |
| NFR-REL-003 | The system shall maintain data integrity during synchronization. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes) | SDS 2.1 (Class Diagram - FileStatus enum with StatusConflict), SDS 5.4 (Data Validation Rules), SDS 6.4 (Error Handling - Graceful Degradation) | UT-03, IT-05 |
| NFR-REL-004 | The system shall handle API rate limiting gracefully. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Graceful handling of API rate limits) | SDS 4.4 (Rate Limiting and Quotas), SDS 6.4 (Error Handling - Retry Logic with exponential backoff) | SEC-01 |

## Architectural Decisions

| Architecture Decision ID | Description | Design Specification Reference | Test Case IDs |
|--------------------------|-------------|--------------------------------|--------------|
| AD-001 | Use of FUSE for filesystem implementation | SDS 2.1 (Class Diagram - Filesystem class), SDS 6.1 (Dependencies - go-fuse/v2) | UT-01, ST-01, ST-02 |
| AD-002 | On-demand file download instead of full sync | SDS 2.1 (Class Diagram - DownloadManager class), SDS 3.1 (Sequence Diagram - File Access Not Cached) | UT-05, ST-08 |
| AD-003 | Local caching of metadata and content | SDS 2.1 (Class Diagram - LoopbackCache and ThumbnailCache classes), SDS 5.3 (Database Schema), SDS 6.2 (Performance Considerations - Caching) | UT-01, UT-05, IT-01, IT-06 |
| AD-004 | Use of BBolt for embedded database | SDS 2.1 (Class Diagram - Filesystem with db attribute), SDS 5.3 (Database Schema), SDS 6.1 (Dependencies - bbolt) | UT-01, IT-01, IT-06 |
| AD-007 | D-Bus interface for file status updates | SDS 2.1 (Class Diagram - Filesystem with dbusServer attribute), SDS 6.1 (Dependencies - dbus) | ST-11 |
| AD-008 | Concurrent operations for uploads and downloads | SDS 2.1 (Class Diagram - UploadManager and DownloadManager with worker goroutines), SDS 6.2 (Performance Considerations - Concurrent Operations) | UT-02, UT-03, UT-04, UT-05 |

## Use Cases

| Use Case ID | Use Case Name | Related Requirements | Test Case IDs |
|-------------|---------------|----------------------|--------------|
| UC-FS-001 | Mount OneDrive Filesystem | FR-FS-001, FR-AUTH-001, FR-AUTH-002, FR-UI-001, FR-UI-002, NFR-SEC-001, NFR-USE-003 | UT-01, SEC-01, SEC-04 |
| UC-FS-002 | Access and Modify Files | FR-FS-002, FR-FS-004, FR-FS-005, NFR-PERF-001, NFR-PERF-003, NFR-PERF-004, NFR-REL-001, NFR-REL-003 | UT-02, UT-05, ST-06, ST-07, ST-08, ST-10 |
| UC-OFF-001 | Work with Files Offline | FR-OFF-001, FR-OFF-002, FR-OFF-003, FR-OFF-004, NFR-REL-001, NFR-REL-003 | IT-01, IT-02, IT-03, IT-04, IT-05, IT-06 |
| UC-FS-003 | Handle File Conflicts | FR-FS-006, FR-OFF-004, NFR-REL-003, NFR-USE-001 | UT-03, IT-05 |
| UC-INT-001 | View File Status in Nemo File Manager | FR-INT-001, FR-INT-002, FR-INT-003, FR-INT-004, FR-INT-005, FR-INT-006, NFR-USE-002 | ST-11 |

## Test Case Coverage Summary

| Test Case ID | Test Case Title | Requirements Covered | Architecture Elements Covered | Design Elements Covered |
|--------------|----------------|----------------------|-------------------------------|------------------------|
| UT-01 | Directory Tree Synchronization | FR-FS-001, FR-FS-005, NFR-PERF-001, NFR-PERF-003 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 5.1 | SDS 2.1, SDS 2.2, SDS 6.1 |
| UT-02 | File Upload Synchronization | FR-FS-002, NFR-PERF-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| UT-03 | Repeated File Upload | FR-FS-006, NFR-PERF-002, NFR-REL-003 | SAS 4.3, SAS 5.3 | SDS 2.1, SDS 5.2.1, SDS 5.4 |
| UT-04 | Large File Upload | NFR-PERF-002, NFR-PERF-004 | SAS 3.2.3, SAS 3.4.1, SAS 4.2.2 | SDS 3.1, SDS 4.2.2, SDS 6.2 |
| UT-05 | File Download Synchronization | FR-FS-002, FR-FS-004, NFR-PERF-001, NFR-PERF-002, NFR-PERF-003 | SAS 1.3, SAS 3.2.1, SAS 3.4.3, SAS 4.2.2, SAS 5.1 | SDS 2.1, SDS 3.1, SDS 5.2.1 |
| IT-01 | Offline File Access | FR-OFF-001, FR-OFF-002, NFR-PERF-001, NFR-PERF-003, NFR-REL-001 | SAS 3.2.1, SAS 3.2.3, SAS 4.3, SAS 5.3 | SDS 2.1, SDS 3.1, SDS 5.2.1 |
| IT-02 | Offline File Creation | NFR-REL-001 | SAS 3.2.3, SAS 4.3 | SDS 2.1, SDS 6.4 |
| IT-03 | Offline File Modification | NFR-REL-001 | SAS 3.2.3, SAS 4.3 | SDS 2.1, SDS 6.4 |
| IT-04 | Offline File Deletion | NFR-REL-001 | SAS 3.2.3, SAS 4.3 | SDS 2.1, SDS 6.4 |
| IT-05 | Reconnection Synchronization | FR-FS-006, FR-OFF-003, FR-OFF-004, NFR-REL-001, NFR-REL-003 | SAS 4.3, SAS 5.3 | SDS 2.1, SDS 5.2.1, SDS 6.4 |
| SEC-01 | Unauthenticated Request Handling | FR-AUTH-001, FR-AUTH-004, NFR-SEC-002 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.3, SAS 4.1.2 | SDS 2.1, SDS 3.1, SDS 4.3 |
| SEC-02 | Authentication Token Refresh | FR-AUTH-003, NFR-SEC-002 | SAS 3.2.3, SAS 4.1.2, SAS 5.3 | SDS 2.1, SDS 3.1, SDS 4.3 |
| SEC-03 | Invalid Authentication Code Format | FR-AUTH-001, NFR-SEC-002 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.3, SAS 4.1.2 | SDS 2.1, SDS 3.1, SDS 4.3 |
| SEC-04 | Authentication Persistence | FR-AUTH-002, NFR-SEC-001, NFR-SEC-003 | SAS 3.5.1, SAS 4.1.1, SAS 4.1.2, SAS 4.1.3 | SDS 2.1, SDS 4.3, SDS 6.3 |
| SEC-05 | Authentication Failure with Network Available | FR-AUTH-001, FR-AUTH-004 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.3, SAS 4.1.2 | SDS 2.1, SDS 3.1, SDS 4.3 |
| ST-01 | Directory Reading | FR-FS-001, FR-FS-003 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 5.1 | SDS 2.1, SDS 2.2, SDS 6.1 |
| ST-02 | Directory Listing with Shell Commands | FR-FS-001, FR-FS-003 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 5.1 | SDS 2.1, SDS 2.2, SDS 6.1 |
| ST-03 | File Creation and Modification Time | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| ST-04 | File Permissions | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| ST-05 | Directory Creation and Removal | FR-FS-003 | SAS 3.2.1, SAS 3.2.3 | SDS 2.1, SDS 4.2.3, SDS 4.2.4, SDS 4.2.5 |
| ST-06 | File Writing with Offset | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| ST-07 | File Movement Operations | FR-FS-002, FR-FS-003 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| ST-08 | Positional File Operations | FR-FS-002, FR-FS-004 | SAS 1.3, SAS 3.2.1, SAS 3.2.3, SAS 3.4.3, SAS 4.2.2, SAS 5.1 | SDS 2.1, SDS 3.1, SDS 4.2, SDS 5.2.1 |
| ST-09 | Case Sensitivity Handling | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| ST-10 | Special Characters in Filenames | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| ST-11 | Trash Functionality | FR-FS-002, FR-INT-001, FR-INT-002, FR-INT-003, FR-INT-004, FR-INT-005, FR-INT-006 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 3.4.2, SAS 4.3 | SDS 2.1, SDS 6.1 |
| ST-12 | Directory Paging | FR-FS-003 | SAS 3.2.1, SAS 3.2.3 | SDS 2.1, SDS 4.2.3, SDS 4.2.4, SDS 4.2.5 |
| ST-13 | Application-Specific Save Patterns | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| SEC-06 | Authentication Configuration Merging | FR-AUTH-001 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.3, SAS 4.1.2 | SDS 2.1, SDS 3.1, SDS 4.3 |
| SEC-07 | Resource Path Handling | FR-FS-001, FR-FS-002 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 3.2.3, SAS 3.4.3, SAS 5.1 | SDS 2.1, SDS 2.2, SDS 3.1, SDS 4.2, SDS 6.1 |
| IT-06 | Offline Changes Caching | FR-OFF-002, FR-OFF-004 | SAS 3.2.1, SAS 3.2.3, SAS 4.2.3, SAS 4.3, SAS 5.1, SAS 5.3 | SDS 2.1, SDS 3.1, SDS 5.2.1, SDS 6.4 |
