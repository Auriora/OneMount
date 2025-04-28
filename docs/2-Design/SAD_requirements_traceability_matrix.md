# Requirements Traceability Matrix

This document provides a mapping between the numbered requirements in the Software Requirements Specification (SRS) and the architectural elements described in the Software Architecture Document (SAD).

## Functional Requirements

### Filesystem Operations

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| FR-FS-001 | The system shall mount OneDrive as a native Linux filesystem using FUSE. | SAD 3.1.2 (External Entities - Linux Filesystem), SAD 3.1.3 (System Interfaces - FUSE Interface), SAD 3.2.1 (Logical Components - Filesystem Implementation), SAD 5.1 (Key Architectural Decisions - Use of FUSE) |
| FR-FS-002 | The system shall support standard file operations (read, write, create, delete, rename). | SAD 3.2.1 (Logical Components - Filesystem Implementation), SAD 3.2.3 (Key Abstractions - Filesystem, Inode), SAD 3.4.3 (Sequence Diagrams - File Access Workflow) |
| FR-FS-003 | The system shall support standard directory operations (list, create, delete, rename). | SAD 3.2.1 (Logical Components - Filesystem Implementation), SAD 3.2.3 (Key Abstractions - Filesystem, Inode) |
| FR-FS-004 | The system shall download files on-demand when accessed rather than syncing all files. | SAD 1.3 (System Overview), SAD 3.2.1 (Logical Components - Filesystem Implementation), SAD 3.4.3 (Sequence Diagrams - File Access Workflow), SAD 4.2.2 (Scalability), SAD 5.1 (Key Architectural Decisions - On-demand file download) |
| FR-FS-005 | The system shall cache file metadata to improve performance. | SAD 3.2.1 (Logical Components - Cache Management), SAD 3.2.3 (Key Abstractions - Cache), SAD 4.2.3 (Caching Strategy), SAD 5.1 (Key Architectural Decisions - Local caching) |
| FR-FS-006 | The system shall handle file conflicts between local and remote changes. | SAD 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes), SAD 5.3 (Quality Attribute Scenarios - Availability Scenario) |

### Authentication

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| FR-AUTH-001 | The system shall authenticate with Microsoft accounts using OAuth 2.0. | SAD 3.1.2 (External Entities - Microsoft OneDrive / Graph API), SAD 3.1.3 (System Interfaces - Microsoft Graph API Interface), SAD 3.2.3 (Key Abstractions - Auth), SAD 4.1.2 (Authentication and Authorization) |
| FR-AUTH-002 | The system shall securely store authentication tokens. | SAD 3.5.1 (Deployment Diagram - auth_tokens.json), SAD 4.1.1 (Security Requirements), SAD 4.1.2 (Authentication and Authorization), SAD 4.1.3 (Data Protection) |
| FR-AUTH-003 | The system shall automatically refresh authentication tokens when they expire. | SAD 3.2.3 (Key Abstractions - Auth), SAD 4.1.2 (Authentication and Authorization), SAD 5.3 (Quality Attribute Scenarios - Security Scenario) |
| FR-AUTH-004 | The system shall support re-authentication when refresh tokens are invalid. | SAD 4.1.2 (Authentication and Authorization), SAD 5.3 (Quality Attribute Scenarios - Security Scenario) |

### Offline Functionality

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| FR-OFF-001 | The system shall provide access to previously accessed files when offline. | SAD 3.2.1 (Logical Components - Cache Management), SAD 3.2.3 (Key Abstractions - Cache), SAD 4.3 (Other Crosscutting Concerns - Availability), SAD 5.3 (Quality Attribute Scenarios - Availability Scenario) |
| FR-OFF-002 | The system shall cache file content for offline access. | SAD 3.2.1 (Logical Components - Cache Management), SAD 3.2.3 (Key Abstractions - Cache), SAD 4.2.3 (Caching Strategy), SAD 5.1 (Key Architectural Decisions - Local caching) |
| FR-OFF-003 | The system shall automatically detect network connectivity changes. | SAD 4.3 (Other Crosscutting Concerns - Availability - Automatic reconnection), SAD 5.3 (Quality Attribute Scenarios - Availability Scenario) |
| FR-OFF-004 | The system shall synchronize changes made offline when connectivity is restored. | SAD 4.3 (Other Crosscutting Concerns - Availability - Automatic reconnection), SAD 5.3 (Quality Attribute Scenarios - Availability Scenario) |

### User Interface

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| FR-UI-001 | The system shall provide a command-line interface for mounting and configuration. | SAD 3.1.1 (System Context Diagram - Command Line Interface), SAD 3.2.1 (Logical Components - Command Line Interface) |
| FR-UI-002 | The system shall provide a graphical user interface for mounting and configuration. | SAD 3.1.1 (System Context Diagram - UI Components), SAD 3.2.1 (Logical Components - User Interface) |
| FR-UI-003 | The system shall display file status and synchronization information. | SAD 3.2.1 (Logical Components - User Interface), SAD 4.3 (Other Crosscutting Concerns - Usability) |
| FR-UI-004 | The system shall provide system tray integration for status indication. | SAD 3.1.2 (External Entities - Desktop Environment), SAD 3.2.1 (Logical Components - User Interface) |

### Statistics and Analysis

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| FR-STAT-001 | The system shall provide a statistics command to analyze OneDrive content metadata. | SAD 3.2.1 (Logical Components - Command Line Interface) |
| FR-STAT-002 | The system shall analyze file type distribution in the statistics command. | SAD 3.2.1 (Logical Components - Command Line Interface) |
| FR-STAT-003 | The system shall analyze directory depth statistics in the statistics command. | SAD 3.2.1 (Logical Components - Command Line Interface) |
| FR-STAT-004 | The system shall analyze file size distribution in the statistics command. | SAD 3.2.1 (Logical Components - Command Line Interface) |
| FR-STAT-005 | The system shall analyze file age information in the statistics command. | SAD 3.2.1 (Logical Components - Command Line Interface) |

### Integration with External Systems

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| FR-INT-001 | The system shall provide a D-Bus interface for file status updates. | SAD 3.1.3 (System Interfaces - D-Bus Interface), SAD 3.4.2 (Process Communication - D-Bus) |
| FR-INT-002 | The system shall expose methods for getting file status through the D-Bus interface. | SAD 3.1.3 (System Interfaces - D-Bus Interface), SAD 3.4.2 (Process Communication - D-Bus) |
| FR-INT-003 | The system shall emit signals for file status changes through the D-Bus interface. | SAD 3.1.3 (System Interfaces - D-Bus Interface), SAD 3.4.2 (Process Communication - D-Bus) |
| FR-INT-004 | The system shall integrate with the Nemo file manager to display OneDrive in the sidebar. | SAD 3.1.2 (External Entities - Desktop Environment), SAD 4.3 (Other Crosscutting Concerns - Usability - Seamless integration with desktop environment) |
| FR-INT-005 | The system shall display file status icons in the Nemo file manager. | SAD 3.1.2 (External Entities - Desktop Environment), SAD 4.3 (Other Crosscutting Concerns - Usability - Seamless integration with desktop environment) |
| FR-INT-006 | The system shall fall back to extended attributes if D-Bus is not available. | SAD 3.4.2 (Process Communication) |

### Developer Tools

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| FR-DEV-001 | The system shall provide a method logging framework for debugging. | SAD 3.3.3 (Development Environment - Error Handling - Use structured logging with zerolog), SAD 4.2.4 (Performance Monitoring - Structured logging of operation times) |
| FR-DEV-002 | The system shall log method entry and exit with parameters and return values. | SAD 3.3.3 (Development Environment - Error Handling - Use structured logging with zerolog), SAD 4.2.4 (Performance Monitoring - Structured logging of operation times) |
| FR-DEV-003 | The system shall include execution duration in method logs. | SAD 3.3.3 (Development Environment - Error Handling - Use structured logging with zerolog), SAD 4.2.4 (Performance Monitoring - Structured logging of operation times) |
| FR-DEV-004 | The system shall provide a workflow analyzer tool for developers. | SAD 3.3.3 (Development Environment - Tools) |
| FR-DEV-005 | The system shall provide a code complexity analyzer tool. | SAD 3.3.3 (Development Environment - Tools) |

## Non-Functional Requirements

### Performance

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| NFR-PERF-001 | The system shall minimize network requests to improve performance. | SAD 3.2.1 (Logical Components - Cache Management), SAD 4.2.1 (Performance Requirements), SAD 4.2.3 (Caching Strategy) |
| NFR-PERF-002 | The system shall use concurrent operations where appropriate. | SAD 3.3.3 (Development Environment - Performance - Use concurrent operations where appropriate), SAD 3.4.1 (Runtime Processes), SAD 4.2.2 (Scalability - Concurrent file transfers) |
| NFR-PERF-003 | The system shall implement efficient caching to reduce API calls. | SAD 3.2.1 (Logical Components - Cache Management), SAD 3.2.3 (Key Abstractions - Cache), SAD 4.2.3 (Caching Strategy) |
| NFR-PERF-004 | The system shall support chunked downloads for large files. | SAD 3.2.3 (Key Abstractions - UploadManager/DownloadManager - Handle chunking for large files) |

### Security

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| NFR-SEC-001 | The system shall store authentication tokens with appropriate file permissions. | SAD 3.5.1 (Deployment Diagram - auth_tokens.json), SAD 4.1.1 (Security Requirements), SAD 4.1.3 (Data Protection - Local file permissions for cached content) |
| NFR-SEC-002 | The system shall use HTTPS for all API communications. | SAD 3.5.1 (Deployment Diagram - HTTPS), SAD 4.1.3 (Data Protection - HTTPS for all communication with Microsoft Graph API) |
| NFR-SEC-003 | The system shall not expose authentication tokens to non-privileged users. | SAD 4.1.1 (Security Requirements), SAD 4.1.3 (Data Protection), SAD 4.1.4 (Security Patterns - Principle of least privilege) |

### Usability

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| NFR-USE-001 | The system shall provide clear error messages. | SAD 4.3 (Other Crosscutting Concerns - Usability - Helpful error messages) |
| NFR-USE-002 | The system shall integrate with the Linux desktop environment. | SAD 3.1.2 (External Entities - Desktop Environment), SAD 3.1.3 (System Interfaces - GTK3 Interface, D-Bus Interface), SAD 4.3 (Other Crosscutting Concerns - Usability - Seamless integration with desktop environment) |
| NFR-USE-003 | The system shall provide documentation for installation and usage. | SAD 4.3 (Other Crosscutting Concerns - Usability), SAD 3.3.3 (Development Environment - Documentation) |

### Reliability

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| NFR-REL-001 | The system shall handle network errors and retry operations. | SAD 3.2.3 (Key Abstractions - UploadManager/DownloadManager - Implement retry logic and error handling), SAD 4.3 (Other Crosscutting Concerns - Reliability - Retry logic for network operations) |
| NFR-REL-002 | The system shall recover gracefully from crashes. | SAD 4.3 (Other Crosscutting Concerns - Availability - Crash recovery mechanisms) |
| NFR-REL-003 | The system shall maintain data integrity during synchronization. | SAD 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes) |
| NFR-REL-004 | The system shall handle API rate limiting gracefully. | SAD 4.3 (Other Crosscutting Concerns - Reliability - Graceful handling of API rate limits) |

### Maintainability

| Requirement ID | Requirement Description | Architecture Document Reference |
|----------------|-------------------------|--------------------------------|
| NFR-MNT-001 | The system shall follow Go's standard project layout. | SAD 3.3.2 (Code Structure), SAD 3.3.3 (Development Environment - Code Organization - Follow Go's standard project layout) |
| NFR-MNT-002 | The system shall include comprehensive test coverage. | SAD 3.3.3 (Development Environment - Testing), SAD 4.3 (Other Crosscutting Concerns - Maintainability - Comprehensive test suite) |
| NFR-MNT-003 | The system shall use structured logging for debugging. | SAD 3.3.3 (Development Environment - Error Handling - Use structured logging with zerolog), SAD 4.2.4 (Performance Monitoring - Structured logging of operation times) |
| NFR-MNT-004 | The system shall document public APIs with godoc-compatible comments. | SAD 3.3.3 (Development Environment - Documentation - Document public APIs with godoc-compatible comments) |

## Use Cases

| Use Case ID | Use Case Name | Related Requirements | Architecture Document Reference |
|-------------|---------------|----------------------|--------------------------------|
| UC-FS-001 | Mount OneDrive Filesystem | FR-FS-001, FR-AUTH-001, FR-AUTH-002, FR-UI-001, FR-UI-002, NFR-SEC-001, NFR-USE-003 | SAD 3.1.1 (System Context Diagram), SAD 3.1.2 (External Entities), SAD 3.1.3 (System Interfaces), SAD 3.2.1 (Logical Components) |
| UC-FS-002 | Access and Modify Files | FR-FS-002, FR-FS-004, FR-FS-005, NFR-PERF-001, NFR-PERF-003, NFR-PERF-004, NFR-REL-001, NFR-REL-003 | SAD 3.2.1 (Logical Components), SAD 3.2.3 (Key Abstractions), SAD 3.4.3 (Sequence Diagrams - File Access Workflow), SAD 4.2.3 (Caching Strategy) |
| UC-OFF-001 | Work with Files Offline | FR-OFF-001, FR-OFF-002, FR-OFF-003, FR-OFF-004, NFR-REL-001, NFR-REL-003 | SAD 3.2.1 (Logical Components - Cache Management), SAD 4.3 (Other Crosscutting Concerns - Availability), SAD 5.3 (Quality Attribute Scenarios - Availability Scenario) |
| UC-FS-003 | Handle File Conflicts | FR-FS-006, FR-OFF-004, NFR-REL-003, NFR-USE-001 | SAD 4.3 (Other Crosscutting Concerns - Reliability - Conflict resolution for concurrent changes) |
| UC-STAT-001 | Analyze OneDrive Content with Statistics | FR-STAT-001, FR-STAT-002, FR-STAT-003, FR-STAT-004, FR-STAT-005, NFR-PERF-001, NFR-USE-003 | SAD 3.2.1 (Logical Components - Command Line Interface) |
| UC-INT-001 | View File Status in Nemo File Manager | FR-INT-001, FR-INT-002, FR-INT-003, FR-INT-004, FR-INT-005, FR-INT-006, NFR-USE-002 | SAD 3.1.2 (External Entities - Desktop Environment), SAD 3.1.3 (System Interfaces - D-Bus Interface), SAD 3.4.2 (Process Communication - D-Bus), SAD 4.3 (Other Crosscutting Concerns - Usability - Seamless integration with desktop environment) |