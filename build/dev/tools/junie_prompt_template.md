# Implement GitHub Issue

## Instructions
I need your help implementing a GitHub issue for the OneMount project. I'll provide the issue number, and I'd like you to:

1. Extract the issue details from the GitHub repository
2. Open the corresponding JetBrains task (creating a branch)
3. Gather relevant documentation
4. Analyze the issue and documentation to create an implementation plan
5. Implement the changes according to the plan
6. Update the issue comment with implementation details
7. Close the issue if resolved
8. Commit the changes to git

## Issue Number
[ISSUE_NUMBER]

## Project Context
OneMount is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing. It's written in Go and uses FUSE to implement the filesystem.

## Implementation Process
Please follow this process:

1. **Issue Details**
   - Analyze the issue description, labels, and any referenced documentation

2. **Gather Documentation**
   - Identify relevant documentation based on the issue type and labels
   - Review the documentation to understand the context and requirements

3. **Create Implementation Plan**
   - Based on the issue and documentation, create a detailed implementation plan
   - Break down the implementation into manageable steps

4. **Implement Changes**
   - Follow the implementation plan to make the necessary changes
   - Write or update tests as needed
   - Run tests and ensure they pass

5. **Update Issue Comment**
   - Prepare a summary of the implementation
   - Update the issue comment with the implementation details

6. **Close Issue**
   - If the implementation resolves the issue, close it using the 'gh' command

7. **Commit Changes**
   - Commit the changes to git with a descriptive message

## Additional Requirements
- Follow the project's coding standards and best practices
- Ensure backward compatibility unless explicitly stated otherwise
- Handle edge cases appropriately
- Add appropriate error handling and logging
- Document your changes according to the project's documentation guidelines

## Output Format
Please provide your response in the following format:

1. **Issue Analysis**: A brief analysis of the issue and its requirements
2. **Implementation Plan**: A detailed plan for implementing the changes
3. **Implementation**: The actual implementation of the changes
4. **Testing**: How you tested the changes
5. **Documentation**: Any documentation updates
6. **Issue Comment**: A summary of the implementation for the issue comment
7. **Next Steps**: Any recommended follow-up actions

## Command Examples

### Create a git branch
Create a new branch with the same name as the issue number using the 'git' command.

### Build and Test
```
# Build the project
make

# Run specific tests, e.g.
go test ./internal/fs/...
```

### Commit Changes
```
# Example git commands for committing changes
git add .
git commit -m "Fix #${ISSUE_NUMBER}: Brief description of changes"
git push origin issue-${ISSUE_NUMBER}
```

### Update Issue Comment
```
# Example GitHub CLI command to add a comment and close an issue
gh issue comment ${ISSUE_NUMBER} --body "Implementation completed:
- [Description of changes made]
- [Tests added/modified]
- [Documentation updated]
```

## Project-Specific Notes

### Project Structure
OneMount follows this structure:

- **assets/** - Project assets
  - **examples/** - Example files
  - **icons/** - Icon files
- **build/** - Build artifacts and configuration
  - **cli/** - Command-line interface build files
  - **dev/** - Development build files
  - **package/** - Packaging files
- **cmd/** - Command-line applications
  - **common/** - Shared code between applications
  - **onemount/** - Main filesystem application
  - **onemount-launcher/** - GUI launcher application
- **configs/** - Configuration files and resources
  - **resources/** - Resource files for the application
- **data/** - Data files and resources for the project
- **deployments/** - Deployment configurations
  - **desktop/** - Desktop environment integration
  - **systemd/** - Systemd service files
- **docs/** - Documentation
- **internal/** - Internal implementation code
  - **fs/** - Filesystem implementation
  - **ui/** - GUI implementation
    - **systemd/** - Systemd integration for the UI
  - **nemo/** - Nemo file manager integration
- **pkg/** - Shared packages
  - **errors/** - Error handling utilities
  - **graph/** - Microsoft Graph API client
  - **logging/** - Logging utilities
  - **quickxorhash/** - QuickXorHash implementation
  - **testutil/** - Testing utilities
  - **util/** - General utilities

### Tech Stack
- **Go** - Primary programming language
- **FUSE (go-fuse/v2)** - Filesystem implementation
- **GTK3 (gotk3)** - GUI components
- **bbolt** - Embedded database for caching
- **testify** - Testing framework

### Best Practices
1. **Code Organization**
   - Group related functionality into separate files
   - Use interfaces to decouple components
   - Follow Go's standard project layout

2. **Error Handling**
   - Return errors to callers instead of handling them internally
   - Use structured logging with zerolog
   - Avoid using `log.Fatal()` in library code

3. **Testing**
   - Write both unit and integration tests
   - Use 'testify' for assertions
   - Test edge cases, especially around network connectivity

4. **Performance**
   - Cache filesystem metadata and file contents
   - Minimize network requests
   - Use concurrent operations where appropriate

5. **Documentation**
   - Document public APIs with godoc-compatible comments
   - Add comments explaining complex logic
   - Update existing documentation rather than adding new documentation
   - Add links to new documentation in relevant existing documentation

6. **Method Logging**
   - Use the method logging framework for all public methods
   - Log method entry and exit, including parameters and return values

7. **D-Bus Integration**
   - Use the D-Bus interface for file status updates
   - Ensure backward compatibility with extended attributes

8. **Microsoft Graph API Integration**
   - Use direct API endpoints when available for better performance
   - Cache API responses appropriately to reduce network traffic
