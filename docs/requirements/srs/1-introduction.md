# 1. Introduction

This section provides an introduction to the OneMount software system and the SRS document.

## 1.1 Purpose
The purpose of this Software Requirements Specification (SRS) document is to define the requirements for the OneMount system, a native Linux filesystem for Microsoft OneDrive. This document serves as a reference for developers, testers, and stakeholders to understand the system's functionality, constraints, and design considerations.

## 1.2 Scope
OneMount is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing. The system:

**Will do:**
- Provide a native Linux filesystem interface to Microsoft OneDrive
- Perform on-demand file downloads instead of full synchronization
- Support offline mode functionality
- Provide a GUI launcher application
- Cache filesystem metadata and file contents
- Support authentication with Microsoft accounts

**Will not do:**
- Support other cloud storage providers (e.g., Google Drive, Dropbox)
- Provide Windows or macOS compatibility
- Implement full local synchronization of all files

## 1.3 Definitions and Acronyms
- **FUSE**: Filesystem in Userspace - a software interface for Unix and Unix-like computer operating systems that lets non-privileged users create their own file systems without editing kernel code
- **OneDrive**: Microsoft's cloud storage service
- **Go/Golang**: The primary programming language used for OneMount
- **GTK3**: GIMP Toolkit version 3, used for GUI components
- **bbolt**: An embedded key/value database for Go, used for caching
- **zerolog**: A structured logging library for Go
- **testify**: A testing toolkit for Go
- **Graph API**: Microsoft's RESTful web API interface for accessing Microsoft Cloud service resources

## 1.4 Stakeholders
The following table identifies the key stakeholders for the OneMount system, their roles, and their primary goals:

| Stakeholder | Role | Goal |
|-------------|------|------|
| Linux Users | End Users | Access OneDrive files on Linux without syncing entire account |
| Windows/Mac Users Migrating to Linux | End Users | Easily transition files from Windows/Mac to Linux via OneDrive |
| Mobile Device Users | End Users | Access photos and files uploaded from mobile devices on Linux |
| Users with Limited Storage | End Users | Access large OneDrive accounts without using equivalent local storage |
| Users with Poor Internet | End Users | Work with OneDrive files even with unreliable internet connection |
| Developers | Contributors | Extend and improve the OneMount codebase |
| File Manager Developers | Integration Partners | Integrate file managers (like Nemo) with OneMount for better user experience |
| Package Maintainers | Distributors | Package OneMount for different Linux distributions (Fedora, Ubuntu, Arch, etc.) |
| Microsoft | Service Provider | Enable cross-platform access to OneDrive service |
| System Administrators | IT Support | Deploy OneMount in organizational environments |
| Office Suite Users | End Users | Seamlessly work with Office documents stored in OneDrive |
| Photographers/Media Creators | End Users | Access and edit media files stored in OneDrive |
| Students/Academics | End Users | Access educational materials and research stored in OneDrive |
| Business Users | End Users | Access work documents and collaborate through OneDrive |
