# 1. Introduction

This section provides an introduction to the onedriver software system and the SRS document.

## 1.1 Purpose
The purpose of this Software Requirements Specification (SRS) document is to define the requirements for the onedriver system, a native Linux filesystem for Microsoft OneDrive. This document serves as a reference for developers, testers, and stakeholders to understand the system's functionality, constraints, and design considerations.

## 1.2 Scope
Onedriver is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing. The system:

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
- **Go/Golang**: The primary programming language used for onedriver
- **GTK3**: GIMP Toolkit version 3, used for GUI components
- **bbolt**: An embedded key/value database for Go, used for caching
- **zerolog**: A structured logging library for Go
- **testify**: A testing toolkit for Go
- **Graph API**: Microsoft's RESTful web API interface for accessing Microsoft Cloud service resources
