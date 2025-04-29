# Changelog

All notable changes to the onedriver project will be documented in this file.

## [Unreleased]

## Release highlights

* D-Bus Interface: Added D-Bus interface for file status updates and improved integration with file managers
* Background Download Manager: Added background download manager for improved performance with large files
* File Status Tracking: Added file status tracking and Nemo integration for better user experience

### Added
- Added drive rename option for better customization of OneDrive mounts
- Added D-Bus interface for file status updates and improved integration with file managers
- Added background download manager for improved performance with large files
- Added configurable delta query interval for more responsive file synchronization
- Added file status tracking and Nemo integration for better user experience
- Added content cache cleanup with expiration support for better resource management
- Added validation for mistyped single-letter mountpoints to prevent user errors
- Added filesystem statistics functionality with `--stats` flag for better diagnostics
- Added in-memory extended attributes (xattrs) support for improved compatibility
- Added thumbnail support to filesystem operations
- Added daemon mode to support running in background
- Added configurable log output option for better troubleshooting
- Added enhanced method logging framework for improved debugging and workflow analysis
- Added developer tools including workflow analyzer and code complexity analyzer

### Changed
- Switched to using fuse3 instead of fuse2 for improved performance and compatibility
- Refactored path handling to prevent potential deadlocks
- Optimized inode path handling to prevent redundant locking
- Replaced deprecated ioutil with modern os and io package functions
- Enhanced performance for directory operations and file handling
- Improved authentication flows with better error handling and context support
- Enhanced unmount reliability and improved error handling for better user experience
- Refactored analyzers to reusable functions for improved code maintainability
- Improved error handling and thread safety for token refresh and socketio connections
- Updated dependencies to latest compatible versions for better security and performance

### Fixed
- Locked opened file inodes to prevent a use-after-free race condition
- Handled Inode serialization to prevent race conditions
- Created config directory if not exists to prevent startup failures
- Fixed root path handling in Inode.Path() method
- Fixed regex in parseAuthCode function to correctly capture the entire auth code
- Fixed reauth error handling in OAuth2 flow
- Handled nil entries and out-of-range offsets in directory listing
- Handled negative size cases in OneDrive directory items
- Handled filesystem limitation errors for extended attributes
- Handled unavailable D-Bus service more gracefully
- Fixed various synchronization issues and improved error recovery

### Testing
- Added Go test run configurations for all tests to improve developer experience
- Refactored tests to use table-driven approach for better organization and coverage
- Improved test reliability by replacing fixed sleeps with dynamic waiting
- Enhanced test setup with improved mount handling and cleanup
- Added mock authentication option for test environments
- Standardized test patterns using testify's require/assert packages consistently
- Added cleanup logic to tests to ensure consistent state
- Fixed D-Bus conflicts by generating unique service names for test instances
- Fixed database access issues by adding retry logic and stale lock file cleanup
- Fixed race conditions in socketio-go integration with proper synchronization
- Added comprehensive unit tests for graph package functionality
- Improved error handling in test setup files for better diagnostics
- Refactored mount directory handling in tests for better reliability
- Consolidated test constants for better code organization

### Documentation
- Added NixOS installation instructions
- Updated openSUSE Leap installation instructions
- Updated Gentoo overlay URL in README
- Added mention of the manual page and "onedriver --help" in the README
- Removed outdated documentation and added threading design documentation
- Added comprehensive SRS (Software Requirements Specification) documentation with detailed requirements
- Added guidelines for method logging and D-Bus integration
- Added comprehensive developer guidelines for new contributors
- Added detailed architecture documentation and enhanced code comments
- Added comprehensive code review checklist
- Added workflow analysis documentation
- Added detailed logging implementation analysis documentation

## [v0.14.1] - 2023-10-18

## Release highlights

* Compatibility Improvements: Fixed compatibility with Ubuntu 20.04 and Debian 11
* Bug Fix: Fixed file redownloads issue

### Fixed
- Fixed file redownloads issue
- Improved compatibility with Ubuntu 20.04 and Debian 11
- Fixed cgo-helper.sh script

### Documentation
- Added Gentoo installation instructions

## [v0.14.0] - 2023-07-14

## Release highlights

* Performance Improvement: Now using local filesystem as a file content cache instead of boltdb
* Compatibility Enhancement: Disallowed restricted characters in filenames
* Security: Added CodeQL for security analysis

### Added
- Disallowed restricted characters in filenames for better compatibility
- Implemented quickxorhash for all account types

### Changed
- Now using local filesystem as a file content cache instead of boltdb for improved performance

### Fixed
- Fixed EL8 build issues

### Security
- Added CodeQL for security analysis

## [v0.13.0] - 2022-11-01

## Release highlights

* UI Improvements: Added configuration UI and rewrote GUI in Go
* Performance Enhancement: Implemented multipart downloads for better performance with large files
* Configuration: Added config file support

### Added
- Added UI for configuration
- Implemented multipart downloads for better performance with large files
- Added config file support

### Changed
- Rewrote GUI in Go
- Using testify for more succinct tests
- Cleaned up dependencies

### Fixed
- Improved error handling when selecting an invalid mountpoint
- Enhanced logging to terminal when mountpoint is not valid

## [v0.12.0] - 2022-01-15

## Release highlights

* Shared Folders: Added support for shared folders
* File Uploads: Implemented file upload resumability
* Thumbnails: Added support for file thumbnails
* Offline Mode: Enhanced offline mode functionality

### Added
- Added support for shared folders
- Implemented file upload resumability
- Added support for file thumbnails

### Changed
- Improved authentication flow
- Enhanced offline mode functionality

### Fixed
- Fixed various synchronization issues
- Improved error handling for network failures

## [v0.11.2] - 2021-11-28

## Release highlights

* File Handling: Fixed issues with file uploads and improved handling of special characters
* Security: Enhanced token storage security
* Performance: Fixed memory leaks in file handling

### Fixed
- Fixed issues with file uploads
- Improved handling of special characters in filenames
- Fixed memory leaks in file handling

### Security
- Enhanced token storage security

## [v0.11.1] - 2021-09-05

## Release highlights

* Stability: Fixed crash on startup with certain configurations
* Authentication: Improved error handling for authentication failures

### Fixed
- Fixed crash on startup with certain configurations
- Improved error handling for authentication failures

## [v0.11.0] - 2021-08-22

## Release highlights

* Integration: Added D-Bus interface for file status updates
* Authentication: Implemented background authentication refresh
* Performance: Improved caching mechanism for better performance

### Added
- Added D-Bus interface for file status updates
- Implemented background authentication refresh

### Changed
- Improved caching mechanism for better performance
- Enhanced logging for easier troubleshooting

### Fixed
- Fixed issues with large file handling
- Improved error recovery for network interruptions

## [v0.10.1] - 2021-06-13

## Release highlights

* Bug Fix: Fixed regression in file upload handling
* Compatibility: Improved compatibility with newer Linux distributions

### Fixed
- Fixed regression in file upload handling
- Improved compatibility with newer Linux distributions

## [v0.10.0] - 2021-06-06

## Release highlights

* File Locking: Added support for file locking
* Conflict Resolution: Implemented improved conflict resolution
* Performance: Enhanced performance for directory operations and improved memory usage

### Added
- Added support for file locking
- Implemented improved conflict resolution

### Changed
- Enhanced performance for directory operations
- Improved memory usage for large directories

### Fixed
- Fixed issues with concurrent file operations
- Improved error handling for API rate limits

## [v0.9.2] - 2021-03-28

### Fixed
- Fixed issues with file metadata caching
- Improved handling of network timeouts
- Enhanced error reporting

## [v0.9.1] - 2021-03-14

### Fixed
- Fixed regression in offline mode
- Improved handling of special characters in paths

## [v0.9.0] - 2021-03-07

### Added
- Added improved offline support
- Implemented better caching of file metadata

### Changed
- Enhanced performance for file operations
- Improved memory usage

## [Earlier Versions]

For changes in earlier versions, please refer to the commit history or release notes.
