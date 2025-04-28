# Test Cases Traceability Matrix

This document provides a mapping between the numbered requirements in the Software Requirements Specification (SRS), the architectural elements described in the Software Architecture Specification (SAS), the design elements in the Software Design Specification (SDS), and the test cases that verify them.

## Functional Requirements

### Filesystem Operations

| Requirement ID | Requirement Description | Architecture Specification Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| FR-FS-001 | The system shall mount OneDrive as a native Linux filesystem using FUSE. | SAS 3.1.2 (External Entities - Linux Filesystem), SAS 3.1.3 (System Interfaces - FUSE Interface), SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 5.1 (Key Architectural Decisions - Use of FUSE) | SDS 2.1 (Class Diagram - Filesystem class), SDS 2.2 (Component Diagram - Filesystem package), SDS 6.1 (Dependencies - go-fuse/v2) | TC-01, TC-16, TC-17 |
| FR-FS-002 | The system shall support standard file operations (read, write, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode), SAS 3.4.3 (Sequence Diagrams - File Access Workflow) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 3.1 (Sequence Diagram - File Access Workflow), SDS 4.2 (API Endpoints/Methods - GetItem, GetItemContent, Remove, Mkdir) | TC-02, TC-05, TC-21, TC-22, TC-23, TC-25 |
| FR-FS-003 | The system shall support standard directory operations (list, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 4.2.3 (API Endpoints/Methods - GetItemChildren), SDS 4.2.4 (API Endpoints/Methods - Mkdir), SDS 4.2.5 (API Endpoints/Methods - Remove) | TC-16, TC-17, TC-20, TC-22, TC-27 |
| FR-FS-004 | The system shall download files on-demand when accessed rather than syncing all files. | SAS 1.3 (System Overview), SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.4.3 (Sequence Diagrams - File Access Workflow), SAS 4.2.2 (Scalability), SAS 5.1 (Key Architectural Decisions - On-demand file download) | SDS 2.1 (Class Diagram - DownloadManager class), SDS 3.1 (Sequence Diagram - File Access Not Cached), SDS 5.2.1 (Entity Definitions - Filesystem - downloads attribute) | TC-05, TC-23 |
| FR-FS-005 | The system shall cache file metadata to improve performance. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy), SAS 5.1 (Key Architectural Decisions - Local caching) | SDS 2.1 (Class Diagram - LoopbackCache and ThumbnailCache classes), SDS 5.2.1 (Entity Definitions - Filesystem - metadata attribute), SDS 5.3 (Database Schema - Inodes table), SDS 6.2 (Performance Considerations - Caching) | TC-01, TC-06 |
| FR-FS-006 | The system shall handle file conflicts between local and remote changes. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - FileStatus enum with StatusConflict), SDS 5.2.1 (Entity Definitions - Inode - hasChanges attribute), SDS 5.4 (Data Validation Rules) | TC-03, TC-10 |

### Authentication

| Requirement ID | Requirement Description | Architecture Specification Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| FR-AUTH-001 | The system shall authenticate with Microsoft accounts using OAuth 2.0. | SAS 3.1.2 (External Entities - Microsoft OneDrive / Graph API), SAS 3.1.3 (System Interfaces - Microsoft Graph API Interface), SAS 3.2.3 (Key Abstractions - Auth), SAS 4.1.2 (Authentication and Authorization) | SDS 2.1 (Class Diagram - Auth class), SDS 3.1 (Sequence Diagram - Authentication Workflow), SDS 4.3 (Authentication and Authorization - OAuth2 for authentication) | TC-11, TC-13, TC-15 |
| FR-AUTH-002 | The system shall securely store authentication tokens. | SAS 3.5.1 (Deployment Diagram - auth_tokens.json), SAS 4.1.1 (Security Requirements), SAS 4.1.2 (Authentication and Authorization), SAS 4.1.3 (Data Protection) | SDS 2.1 (Class Diagram - Auth class with token attributes), SDS 4.3 (Authentication and Authorization - Secure storage of refresh tokens), SDS 6.3 (Security Considerations - Token Storage) | TC-14, TC-15 |
| FR-AUTH-003 | The system shall automatically refresh authentication tokens when they expire. | SAS 3.2.3 (Key Abstractions - Auth), SAS 4.1.2 (Authentication and Authorization), SAS 5.3 (Quality Attribute Scenarios - Security Scenario) | SDS 2.1 (Class Diagram - Auth class with Refresh method), SDS 3.1 (Sequence Diagram - API Request with Authentication), SDS 4.3 (Authentication and Authorization - Automatic token refresh) | TC-12 |
| FR-AUTH-004 | The system shall support re-authentication when refresh tokens are invalid. | SAS 4.1.2 (Authentication and Authorization), SAS 5.3 (Quality Attribute Scenarios - Security Scenario) | SDS 3.1 (Sequence Diagram - Authentication Workflow with refresh failed path), SDS 4.3 (Authentication and Authorization) | TC-11, TC-15 |

### Offline Functionality

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| FR-OFF-001 | The system shall provide access to previously accessed files when offline. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.3 (Other Crosscutting Concerns - Availability), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - Filesystem with offline attribute), SDS 3.1 (Sequence Diagram - File Access Offline Mode), SDS 5.2.1 (Entity Definitions - Filesystem - offline attribute) | TC-06 |
| FR-OFF-002 | The system shall cache file content for offline access. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy), SAS 5.1 (Key Architectural Decisions - Local caching) | SDS 2.1 (Class Diagram - LoopbackCache class), SDS 3.1 (Sequence Diagram - File Access Cached), SDS 5.2.1 (Entity Definitions - Filesystem - content attribute) | TC-06, TC-31 |
| FR-OFF-003 | The system shall automatically detect network connectivity changes. | SAS 4.3 (Other Crosscutting Concerns - Availability - Automatic reconnection), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - Filesystem with offline attribute), SDS 6.4 (Error Handling - Retry Logic) | TC-10 |
| FR-OFF-004 | The system shall synchronize changes made offline when connectivity is restored. | SAS 4.3 (Other Crosscutting Concerns - Availability - Automatic reconnection), SAS 5.3 (Quality Attribute Scenarios - Availability Scenario) | SDS 2.1 (Class Diagram - UploadManager class), SDS 5.2.1 (Entity Definitions - Inode - hasChanges attribute), SDS 6.4 (Error Handling - Retry Logic) | TC-10, TC-31 |

### File System Operations

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| FR-FS-002 | The system shall support standard file operations (read, write, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode), SAS 3.4.3 (Sequence Diagrams - File Access Workflow) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 3.1 (Sequence Diagram - File Access Workflow), SDS 4.2 (API Endpoints/Methods - GetItem, GetItemContent, Remove, Mkdir) | TC-18, TC-19, TC-21, TC-22, TC-23, TC-24, TC-25, TC-26 |
| FR-FS-003 | The system shall support standard directory operations (list, create, delete, rename). | SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.2.3 (Key Abstractions - Filesystem, Inode) | SDS 2.1 (Class Diagram - Filesystem and Inode classes), SDS 4.2.3 (API Endpoints/Methods - GetItemChildren), SDS 4.2.4 (API Endpoints/Methods - Mkdir), SDS 4.2.5 (API Endpoints/Methods - Remove) | TC-20, TC-27 |
| FR-FS-004 | The system shall download files on-demand when accessed rather than syncing all files. | SAS 1.3 (System Overview), SAS 3.2.1 (Logical Components - Filesystem Implementation), SAS 3.4.3 (Sequence Diagrams - File Access Workflow), SAS 4.2.2 (Scalability), SAS 5.1 (Key Architectural Decisions - On-demand file download) | SDS 2.1 (Class Diagram - DownloadManager class), SDS 3.1 (Sequence Diagram - File Access Not Cached), SDS 5.2.1 (Entity Definitions - Filesystem - downloads attribute) | TC-05, TC-23 |

## Non-Functional Requirements

### Performance

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| NFR-PERF-001 | The system shall minimize network requests to improve performance. | SAS 3.2.1 (Logical Components - Cache Management), SAS 4.2.1 (Performance Requirements), SAS 4.2.3 (Caching Strategy) | SDS 2.1 (Class Diagram - ResponseCache class), SDS 6.2 (Performance Considerations - Caching), SDS 6.2 (Performance Considerations - Response Caching) | TC-01, TC-05, TC-06 |
| NFR-PERF-002 | The system shall use concurrent operations where appropriate. | SAS 3.3.3 (Development Environment - Performance - Use concurrent operations where appropriate), SAS 3.4.1 (Runtime Processes), SAS 4.2.2 (Scalability - Concurrent file transfers) | SDS 2.1 (Class Diagram - UploadManager and DownloadManager with worker goroutines), SDS 6.2 (Performance Considerations - Concurrent Operations) | TC-02, TC-03, TC-04, TC-05 |
| NFR-PERF-003 | The system shall implement efficient caching to reduce API calls. | SAS 3.2.1 (Logical Components - Cache Management), SAS 3.2.3 (Key Abstractions - Cache), SAS 4.2.3 (Caching Strategy) | SDS 2.1 (Class Diagram - LoopbackCache, ThumbnailCache, and ResponseCache classes), SDS 5.3 (Database Schema), SDS 6.2 (Performance Considerations - Caching) | TC-01, TC-05, TC-06 |
| NFR-PERF-004 | The system shall support chunked downloads for large files. | SAS 3.2.3 (Key Abstractions - UploadManager/DownloadManager - Handle chunking for large files) | SDS 3.1 (Sequence Diagram - File Access Large File), SDS 4.2.2 (API Endpoints/Methods - GetItemContentStream), SDS 6.2 (Performance Considerations - Chunked Transfers) | TC-04 |

### Security

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| NFR-SEC-001 | The system shall store authentication tokens with appropriate file permissions. | SAS 3.5.1 (Deployment Diagram - auth_tokens.json), SAS 4.1.1 (Security Requirements), SAS 4.1.3 (Data Protection - Local file permissions for cached content) | SDS 4.3 (Authentication and Authorization - Secure storage of refresh tokens), SDS 6.3 (Security Considerations - Token Storage) | TC-14 |
| NFR-SEC-002 | The system shall use HTTPS for all API communications. | SAS 3.5.1 (Deployment Diagram - HTTPS), SAS 4.1.3 (Data Protection - HTTPS for all communication with Microsoft Graph API) | SDS 4.3 (Authentication and Authorization), SDS 6.3 (Security Considerations - HTTPS) | TC-11, TC-12, TC-13 |
| NFR-SEC-003 | The system shall not expose authentication tokens to non-privileged users. | SAS 4.1.1 (Security Requirements), SAS 4.1.3 (Data Protection), SAS 4.1.4 (Security Patterns - Principle of least privilege) | SDS 4.3 (Authentication and Authorization), SDS 6.3 (Security Considerations - Minimal Permissions) | TC-14 |

### Reliability

| Requirement ID | Requirement Description | Architecture Document Reference | Design Specification Reference | Test Case IDs |
|----------------|-------------------------|--------------------------------|--------------------------------|--------------|
| NFR-REL-001 | The system shall handle network errors and retry operations. | SAS 3.2.3 (Key Abstractions - UploadManager/DownloadManager - Implement retry logic and error handling), SAS 4.3 (Other Crosscutting Concerns - Reliability - Retry logic for network operations) | SDS 2.1 (Class Diagram - UploadManager and DownloadManager classes), SDS 6.4 (Error Handling - Retry Logic) | TC-06, TC-07, TC-08, TC-09, TC-10 |
| NFR-REL-003 | The system shall maintain data integrity during synchronization. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes) | SDS 2.1 (Class Diagram - FileStatus enum with StatusConflict), SDS 5.4 (Data Validation Rules), SDS 6.4 (Error Handling - Graceful Degradation) | TC-03, TC-10 |
| NFR-REL-004 | The system shall handle API rate limiting gracefully. | SAS 4.3 (Other Crosscutting Concerns - Reliability - Graceful handling of API rate limits) | SDS 4.4 (Rate Limiting and Quotas), SDS 6.4 (Error Handling - Retry Logic with exponential backoff) | TC-11 |

## Architectural Decisions

| Architecture Decision ID | Description | Design Specification Reference | Test Case IDs |
|--------------------------|-------------|--------------------------------|--------------|
| AD-001 | Use of FUSE for filesystem implementation | SDS 2.1 (Class Diagram - Filesystem class), SDS 6.1 (Dependencies - go-fuse/v2) | TC-01, TC-16, TC-17 |
| AD-002 | On-demand file download instead of full sync | SDS 2.1 (Class Diagram - DownloadManager class), SDS 3.1 (Sequence Diagram - File Access Not Cached) | TC-05, TC-23 |
| AD-003 | Local caching of metadata and content | SDS 2.1 (Class Diagram - LoopbackCache and ThumbnailCache classes), SDS 5.3 (Database Schema), SDS 6.2 (Performance Considerations - Caching) | TC-01, TC-05, TC-06, TC-31 |
| AD-004 | Use of BBolt for embedded database | SDS 2.1 (Class Diagram - Filesystem with db attribute), SDS 5.3 (Database Schema), SDS 6.1 (Dependencies - bbolt) | TC-01, TC-06, TC-31 |
| AD-007 | D-Bus interface for file status updates | SDS 2.1 (Class Diagram - Filesystem with dbusServer attribute), SDS 6.1 (Dependencies - dbus) | TC-26 |
| AD-008 | Concurrent operations for uploads and downloads | SDS 2.1 (Class Diagram - UploadManager and DownloadManager with worker goroutines), SDS 6.2 (Performance Considerations - Concurrent Operations) | TC-02, TC-03, TC-04, TC-05 |

## Use Cases

| Use Case ID | Use Case Name | Related Requirements | Test Case IDs |
|-------------|---------------|----------------------|--------------|
| UC-FS-001 | Mount OneDrive Filesystem | FR-FS-001, FR-AUTH-001, FR-AUTH-002, FR-UI-001, FR-UI-002, NFR-SEC-001, NFR-USE-003 | TC-01, TC-11, TC-14 |
| UC-FS-002 | Access and Modify Files | FR-FS-002, FR-FS-004, FR-FS-005, NFR-PERF-001, NFR-PERF-003, NFR-PERF-004, NFR-REL-001, NFR-REL-003 | TC-02, TC-05, TC-21, TC-22, TC-23, TC-25 |
| UC-OFF-001 | Work with Files Offline | FR-OFF-001, FR-OFF-002, FR-OFF-003, FR-OFF-004, NFR-REL-001, NFR-REL-003 | TC-06, TC-07, TC-08, TC-09, TC-10, TC-31 |
| UC-FS-003 | Handle File Conflicts | FR-FS-006, FR-OFF-004, NFR-REL-003, NFR-USE-001 | TC-03, TC-10 |
| UC-INT-001 | View File Status in Nemo File Manager | FR-INT-001, FR-INT-002, FR-INT-003, FR-INT-004, FR-INT-005, FR-INT-006, NFR-USE-002 | TC-26 |

## Test Case Coverage Summary

| Test Case ID | Test Case Title | Requirements Covered | Architecture Elements Covered | Design Elements Covered |
|--------------|----------------|----------------------|-------------------------------|------------------------|
| TC-01 | Directory Tree Synchronization | FR-FS-001, FR-FS-005, NFR-PERF-001, NFR-PERF-003 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 5.1 | SDS 2.1, SDS 2.2, SDS 6.1 |
| TC-02 | File Upload Synchronization | FR-FS-002, NFR-PERF-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| TC-03 | Repeated File Upload | FR-FS-006, NFR-PERF-002, NFR-REL-003 | SAS 4.3, SAS 5.3 | SDS 2.1, SDS 5.2.1, SDS 5.4 |
| TC-04 | Large File Upload | NFR-PERF-002, NFR-PERF-004 | SAS 3.2.3, SAS 3.4.1, SAS 4.2.2 | SDS 3.1, SDS 4.2.2, SDS 6.2 |
| TC-05 | File Download Synchronization | FR-FS-002, FR-FS-004, NFR-PERF-001, NFR-PERF-002, NFR-PERF-003 | SAS 1.3, SAS 3.2.1, SAS 3.4.3, SAS 4.2.2, SAS 5.1 | SDS 2.1, SDS 3.1, SDS 5.2.1 |
| TC-06 | Offline File Access | FR-OFF-001, FR-OFF-002, NFR-PERF-001, NFR-PERF-003, NFR-REL-001 | SAS 3.2.1, SAS 3.2.3, SAS 4.3, SAS 5.3 | SDS 2.1, SDS 3.1, SDS 5.2.1 |
| TC-07 | Offline File Creation | NFR-REL-001 | SAS 3.2.3, SAS 4.3 | SDS 2.1, SDS 6.4 |
| TC-08 | Offline File Modification | NFR-REL-001 | SAS 3.2.3, SAS 4.3 | SDS 2.1, SDS 6.4 |
| TC-09 | Offline File Deletion | NFR-REL-001 | SAS 3.2.3, SAS 4.3 | SDS 2.1, SDS 6.4 |
| TC-10 | Reconnection Synchronization | FR-FS-006, FR-OFF-003, FR-OFF-004, NFR-REL-001, NFR-REL-003 | SAS 4.3, SAS 5.3 | SDS 2.1, SDS 5.2.1, SDS 6.4 |
| TC-11 | Unauthenticated Request Handling | FR-AUTH-001, FR-AUTH-004, NFR-SEC-002 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.3, SAS 4.1.2 | SDS 2.1, SDS 3.1, SDS 4.3 |
| TC-12 | Authentication Token Refresh | FR-AUTH-003, NFR-SEC-002 | SAS 3.2.3, SAS 4.1.2, SAS 5.3 | SDS 2.1, SDS 3.1, SDS 4.3 |
| TC-13 | Invalid Authentication Code Format | FR-AUTH-001, NFR-SEC-002 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.3, SAS 4.1.2 | SDS 2.1, SDS 3.1, SDS 4.3 |
| TC-14 | Authentication Persistence | FR-AUTH-002, NFR-SEC-001, NFR-SEC-003 | SAS 3.5.1, SAS 4.1.1, SAS 4.1.2, SAS 4.1.3 | SDS 2.1, SDS 4.3, SDS 6.3 |
| TC-15 | Authentication Failure with Network Available | FR-AUTH-001, FR-AUTH-004 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.3, SAS 4.1.2 | SDS 2.1, SDS 3.1, SDS 4.3 |
| TC-16 | Directory Reading | FR-FS-001, FR-FS-003 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 5.1 | SDS 2.1, SDS 2.2, SDS 6.1 |
| TC-17 | Directory Listing with Shell Commands | FR-FS-001, FR-FS-003 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 5.1 | SDS 2.1, SDS 2.2, SDS 6.1 |
| TC-18 | File Creation and Modification Time | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| TC-19 | File Permissions | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| TC-20 | Directory Creation and Removal | FR-FS-003 | SAS 3.2.1, SAS 3.2.3 | SDS 2.1, SDS 4.2.3, SDS 4.2.4, SDS 4.2.5 |
| TC-21 | File Writing with Offset | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| TC-22 | File Movement Operations | FR-FS-002, FR-FS-003 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| TC-23 | Positional File Operations | FR-FS-002, FR-FS-004 | SAS 1.3, SAS 3.2.1, SAS 3.2.3, SAS 3.4.3, SAS 4.2.2, SAS 5.1 | SDS 2.1, SDS 3.1, SDS 4.2, SDS 5.2.1 |
| TC-24 | Case Sensitivity Handling | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| TC-25 | Special Characters in Filenames | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| TC-26 | Trash Functionality | FR-FS-002, FR-INT-001, FR-INT-002, FR-INT-003, FR-INT-004, FR-INT-005, FR-INT-006 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 3.4.2, SAS 4.3 | SDS 2.1, SDS 6.1 |
| TC-27 | Directory Paging | FR-FS-003 | SAS 3.2.1, SAS 3.2.3 | SDS 2.1, SDS 4.2.3, SDS 4.2.4, SDS 4.2.5 |
| TC-28 | Application-Specific Save Patterns | FR-FS-002 | SAS 3.2.1, SAS 3.2.3, SAS 3.4.3 | SDS 2.1, SDS 3.1, SDS 4.2 |
| TC-29 | Authentication Configuration Merging | FR-AUTH-001 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.3, SAS 4.1.2 | SDS 2.1, SDS 3.1, SDS 4.3 |
| TC-30 | Resource Path Handling | FR-FS-001, FR-FS-002 | SAS 3.1.2, SAS 3.1.3, SAS 3.2.1, SAS 3.2.3, SAS 3.4.3, SAS 5.1 | SDS 2.1, SDS 2.2, SDS 3.1, SDS 4.2, SDS 6.1 |
| TC-31 | Offline Changes Caching | FR-OFF-002, FR-OFF-004 | SAS 3.2.1, SAS 3.2.3, SAS 4.2.3, SAS 4.3, SAS 5.1, SAS 5.3 | SDS 2.1, SDS 3.1, SDS 5.2.1, SDS 6.4 |