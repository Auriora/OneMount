# Changelog

All notable changes to the OneMount project will be documented in this file.

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
- Added mention of the manual page and "OneMount --help" in the README
- Removed outdated documentation and added threading design documentation
- Added comprehensive SRS (Software Requirements Specification) documentation with detailed requirements
- Added guidelines for method logging and D-Bus integration
- Added comprehensive developer guidelines for new contributors
- Added detailed architecture documentation and enhanced code comments
- Added comprehensive code review checklist
- Added workflow analysis documentation
- Added detailed logging implementation analysis documentation
