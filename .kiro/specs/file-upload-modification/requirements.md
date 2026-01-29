# Requirements Document: File Upload Modification

## Source
Moved from `.kiro/specs/archive/system-verification-and-fix/requirements.md`. Sections below are verbatim.

## Requirements

### Requirement 4: File Modification and Upload Verification

**User Story:** As a user, I want to edit files locally and have changes automatically uploaded to OneDrive so that my work is synchronized.

#### Acceptance Criteria

1. WHEN the user modifies a file, THE OneMount System SHALL mark the file as having local changes
2. WHEN the user saves a modified file, THE OneMount System SHALL queue the file for upload
3. WHEN uploading a file smaller than 250 MB, THE OneMount System SHALL use PUT `/items/{id}/content` with the file content
4. WHEN uploading a file larger than 250 MB, THE OneMount System SHALL create an upload session using POST `/createUploadSession`
5. WHEN using an upload session, THE OneMount System SHALL upload the file in chunks to the session URL
6. IF an upload fails due to network issues, THEN THE OneMount System SHALL retry with exponential backoff
7. WHEN an upload completes successfully, THE OneMount System SHALL update the file's ETag from the response
8. WHEN an upload completes successfully, THE OneMount System SHALL clear the modified flag
9. WHEN the user creates a directory, THE OneMount System SHALL create the directory on the server and assign it a unique ID
10. WHEN the user deletes an empty directory using Rmdir, THE OneMount System SHALL remove the directory from the server
11. IF the user attempts to delete a non-empty directory, THEN THE OneMount System SHALL return ENOTEMPTY error
12. WHEN a directory is deleted, THE OneMount System SHALL remove the directory from the parent's children list
13. WHEN a directory is deleted, THE OneMount System SHALL remove the directory inode from the filesystem's internal tracking
