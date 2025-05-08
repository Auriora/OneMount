#!/usr/bin/env python3
"""
Script to implement a GitHub issue with Junie AI assistance.

This script:
1. Extracts an issue from the GitHub repo based on issue number
2. Opens the corresponding JetBrains task (creating a branch)
3. Gathers relevant documentation
4. Implements the changes with Junie AI assistance
5. Updates the issue comment with implementation details
6. Closes the issue if resolved
7. Commits the changes to git

Usage:
    python3 implement_github_issue.py <issue_number>
"""

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path

# Constants
GITHUB_ISSUES_FILE = 'data/github_issues.json'
REPO_ROOT = Path(__file__).resolve().parent.parent.parent

def parse_arguments():
    """Parse command line arguments."""
    parser = argparse.ArgumentParser(description='Implement a GitHub issue with Junie AI assistance.')
    parser.add_argument('issue_number', type=int, help='The GitHub issue number to implement')
    return parser.parse_args()

def load_github_issues():
    """Load GitHub issues from the JSON file."""
    issues_file = REPO_ROOT / GITHUB_ISSUES_FILE
    if not issues_file.exists():
        print(f"Error: GitHub issues file not found at {issues_file}")
        sys.exit(1)

    with open(issues_file, 'r') as f:
        return json.load(f)

def find_issue_by_number(issues, issue_number):
    """Find an issue by its number."""
    for issue in issues:
        if issue.get('number') == issue_number:
            return issue
    return None

def open_jetbrains_task(issue_number):
    """Open the JetBrains task with the same ID as the issue number."""
    print(f"Opening JetBrains task for issue #{issue_number}...")

    # This command will vary depending on your JetBrains IDE and setup
    # For GoLand, you might use something like:
    try:
        # TODO Fix this
        # subprocess.run(['goland', f'--task={issue_number}'], check=True)
        print(f"JetBrains task for issue #{issue_number} opened successfully")
        return True
    except subprocess.CalledProcessError as e:
        print(f"Error opening JetBrains task: {e}")
        return False

def gather_relevant_documentation(issue):
    """Gather relevant documentation based on the issue."""
    print("Gathering relevant documentation...")

    # Extract documentation references from the issue body
    body = issue.get('body', '')
    doc_references = []

    # Look for links to documentation in the issue body
    for line in body.split('\n'):
        if '(' in line and ')' in line and '[' in line and ']' in line:
            # This is a markdown link
            link_text = line[line.find('[')+1:line.find(']')]
            link_url = line[line.find('(')+1:line.find(')')]

            # Only include internal documentation links
            if not link_url.startswith('http'):
                doc_references.append((link_text, link_url))

    # Add standard documentation based on issue labels
    labels = [label.get('name') for label in issue.get('labels', [])]

    if 'documentation' in labels:
        doc_references.append(('Documentation Guidelines', 'docs/guides/documentation-guidelines.md'))

    if 'testing' in labels:
        doc_references.append(('Test Guidelines', 'docs/guides/testing/test-guidelines.md'))

    if 'architecture' in labels:
        doc_references.append(('Architecture Design', 'docs/2-architecture-and-design/architecture-design.md'))

    if 'framework' in labels:
        doc_references.append(('Framework Design', 'docs/2-architecture-and-design/framework-design.md'))

    # Always include development guidelines
    doc_references.append(('Development Guidelines', 'docs/DEVELOPMENT.md'))

    # Print the gathered documentation
    print("Relevant documentation:")
    for doc_name, doc_path in doc_references:
        print(f"- {doc_name}: {doc_path}")
    return doc_references

def create_junie_prompt(issue, doc_references):
    """Create a Junie prompt for implementing the issue."""
    print("Creating Junie prompt...")

    # Extract issue details
    issue_number = issue.get('number')
    issue_title = issue.get('title')
    issue_body = issue.get('body')
    issue_labels = [label.get('name') for label in issue.get('labels', [])]

    # Read the template file
    template_path = REPO_ROOT / 'tmp' / 'junie_prompt_template.md'
    try:
        with open(template_path, 'r') as f:
            template = f.read()
    except FileNotFoundError:
        print(f"Error: Template file not found at {template_path}")
        sys.exit(1)

    # Replace placeholders in the template
    prompt = template.replace('[ISSUE_NUMBER]', str(issue_number))

    # Add issue-specific information at the end
#     prompt += f"""
# # Implement GitHub Issue
#
# ## Instructions
# I need your help implementing a GitHub issue for the OneMount project. I'll provide the issue number, and I'd like you to:
#
# 1. Extract the issue details from the GitHub repository
# 2. Open the corresponding JetBrains task (creating a branch)
# 3. Gather relevant documentation
# 4. Analyze the issue and documentation to create an implementation plan
# 5. Implement the changes according to the plan
# 6. Update the issue comment with implementation details
# 7. Close the issue if resolved
# 8. Commit the changes to git
#
# ## Issue Number
# {issue_number}
#
# ## Project Context
# OneMount is a native Linux filesystem for Microsoft OneDrive that performs on-demand file downloads rather than syncing. It's written in Go and uses FUSE to implement the filesystem.
#
# The project follows a structured organization:
# - **cmd/** - Command-line applications
# - **configs/** - Configuration files and resources
# - **data/** - Data files and resources
# - **docs/** - Documentation
# - **internal/** - Internal implementation code
# - **scripts/** - Utility scripts
#
# ## Implementation Process
# Please follow this process:
#
# 1. **Extract Issue Details**
#    - Use the GitHub API or local JSON file to get the issue details
#    - Analyze the issue description, labels, and any referenced documentation
#
# 2. **Open JetBrains Task**
#    - Open the JetBrains task with the same ID as the issue number
#    - This should create a branch with the same name
#
# 3. **Gather Documentation**
#    - Identify relevant documentation based on the issue type and labels
#    - Review the documentation to understand the context and requirements
#
# 4. **Create Implementation Plan**
#    - Based on the issue and documentation, create a detailed implementation plan
#    - Break down the implementation into manageable steps
#
# 5. **Implement Changes**
#    - Follow the implementation plan to make the necessary changes
#    - Write or update tests as needed
#    - Ensure all tests pass
#
# 6. **Update Issue Comment**
#    - Prepare a summary of the implementation
#    - Update the issue comment with the implementation details
#
# 7. **Close Issue**
#    - If the implementation resolves the issue, close it
#
# 8. **Commit Changes**
#    - Commit the changes to git with a descriptive message
#
# ## Additional Requirements
# - Follow the project's coding standards and best practices
# - Ensure backward compatibility unless explicitly stated otherwise
# - Handle edge cases appropriately
# - Add appropriate error handling and logging
# - Document your changes according to the project's documentation guidelines
#
# ## Output Format
# Please provide your response in the following format:
#
# 1. **Issue Analysis**: A brief analysis of the issue and its requirements
# 2. **Implementation Plan**: A detailed plan for implementing the changes
# 3. **Implementation**: The actual implementation of the changes
# 4. **Testing**: How you tested the changes
# 5. **Documentation**: Any documentation updates
# 6. **Issue Comment**: A summary of the implementation for the issue comment
# 7. **Next Steps**: Any recommended follow-up actions
# """

    # Add additional considerations based on issue labels and content
    if any(label in issue_labels for label in ['complex', 'large']):
        prompt += """
## Additional Considerations

### For Complex Issues
```
This issue seems complex and might benefit from breaking it down:

1. Can you help me divide this issue into smaller, manageable tasks?
2. What dependencies exist between these tasks?
3. Is there a specific order in which these tasks should be implemented?
```
"""
    elif 'testing' in issue_labels:
        prompt += """
## Additional Considerations

### For Testing-Related Changes
```
This issue involves testing improvements:

1. What testing frameworks and tools are used in the project?
2. How should I structure the tests to ensure good coverage?
3. What edge cases should I consider in my tests?
4. How can I ensure my tests are reliable and not flaky?
```
"""

    if any(label in issue_labels for label in ['performance']):
        prompt += """
### For Performance-Critical Changes
```
This change might affect performance:

1. How can I benchmark the current performance before making changes?
2. What metrics should I monitor to ensure my changes don't degrade performance?
3. Are there any specific performance requirements mentioned in the documentation?
4. How can I minimize network requests and optimize caching?
```
"""

    if any(label in issue_labels for label in ['ui']):
        prompt += """
### For UI Changes
```
This issue involves UI changes:

1. Are there any design mockups or specifications I should follow?
2. How can I ensure the UI changes are accessible and follow the project's UI guidelines?
3. What testing should I perform to verify the UI changes work correctly across different environments?
4. How do I integrate with GTK3 using gotk3?
```
"""

    if any(label in issue_labels for label in ['graph-api']) or 'graph api' in issue_body.lower():
        prompt += """
### For Microsoft Graph API Integration
```
This issue involves Microsoft Graph API integration:

1. Which API endpoints should I use for optimal performance?
2. How should I implement caching for API responses?
3. How should I handle authentication and token refresh?
4. What error handling and retry logic should I implement for network operations?
```
"""

    if any(label in issue_labels for label in ['filesystem']) or 'filesystem' in issue_body.lower() or 'fuse' in issue_body.lower():
        prompt += """
### For Filesystem Operations
```
This issue involves filesystem operations:

1. How do I ensure proper FUSE integration?
2. How should I handle offline mode functionality?
3. What error recovery mechanisms should I implement for interrupted operations?
4. How do I ensure proper cleanup of resources?
```
"""

    # Add command examples
    prompt += """
## Command Examples

### Extract Issue Information
Issues can be extracted using the 'gh' command-line tool.


### Find Relevant Documentation
```
search_project "documentation" docs/
```

### Create a git branch
Create a new branch with the same name as the issue number using the 'git' command.

### Build and Test
```
# Build the project
make

# Run all tests
make test

# Run specific tests
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

Closes #${ISSUE_NUMBER}"
```
"""

    # Add project-specific notes
    prompt += """
## Project-Specific Notes

### Project Structure
OneMount follows this structure:
- **cmd/** - Command-line applications
- **configs/** - Configuration files and resources
- **data/** - Data files and resources
- **docs/** - Documentation
- **internal/** - Internal implementation code
  - **fs/** - Filesystem implementation
  - **ui/** - GUI implementation
  - **testutil/** - Testing utilities
- **scripts/** - Utility scripts

### Tech Stack
- **Go** - Primary programming language
- **FUSE (go-fuse/v2)** - Filesystem implementation
- **GTK3 (gotk3)** - GUI components
- **bbolt** - Embedded database for caching
- **zerolog** - Structured logging
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
"""

    # Add issue-specific information
    prompt += f"""
## Issue Details

### Issue #{issue_number}: {issue_title}

#### Description
{issue_body}

#### Labels
{', '.join(issue_labels)}

#### Relevant Documentation
"""

    # Add documentation references
    for doc_name, doc_path in doc_references:
        prompt += f"- [{doc_name}]({doc_path})\n"

    return prompt

def update_issue_comment(issue_number, implementation_summary):
    """Update the issue comment with implementation details."""
    print(f"Updating comment for issue #{issue_number}...")

    # In a real implementation, this would use the GitHub API
    # For this example, we'll just print the comment
    print(f"Comment for issue #{issue_number}:")
    print(implementation_summary)

    return True

def close_issue(issue_number):
    """Close the issue if resolved."""
    print(f"Closing issue #{issue_number}...")

    # In a real implementation, this would use the GitHub API
    # For this example, we'll just print a message
    print(f"Issue #{issue_number} closed")

    return True

def commit_changes(issue_number, issue_title):
    """Commit the changes to git."""
    print("Committing changes to git...")

    # Create a commit message
    commit_message = f"Implement #{issue_number}: {issue_title}"

    # In a real implementation, this would use git commands
    # For this example, we'll just print the commit message
    print(f"Commit message: {commit_message}")

    return True

def main():
    """Main function."""
    args = parse_arguments()
    issue_number = args.issue_number

    print(f"Implementing GitHub issue #{issue_number}...")

    # Load GitHub issues
    issues = load_github_issues()

    # Find the issue
    issue = find_issue_by_number(issues, issue_number)
    if not issue:
        print(f"Error: Issue #{issue_number} not found")
        sys.exit(1)

    # Print issue details
    print(f"Issue #{issue_number}: {issue.get('title')}")
    print(f"State: {issue.get('state')}")

    # Check if the issue is open
    if issue.get('state') != 'OPEN':
        print(f"Warning: Issue #{issue_number} is not open (state: {issue.get('state')})")
        response = input("Do you want to continue? (y/n): ")
        if response.lower() != 'y':
            sys.exit(0)

    # Open JetBrains task
    if not open_jetbrains_task(issue_number):
        print("Warning: Failed to open JetBrains task")
        response = input("Do you want to continue? (y/n): ")
        if response.lower() != 'y':
            sys.exit(0)

    # Gather relevant documentation
    doc_references = gather_relevant_documentation(issue)

    # Create Junie prompt
    prompt = create_junie_prompt(issue, doc_references)

    # Save the prompt to a file
    prompt_file = REPO_ROOT / 'tmp' / f'junie_prompt_issue_{issue_number}.md'
    os.makedirs(prompt_file.parent, exist_ok=True)
    with open(prompt_file, 'w') as f:
        f.write(prompt)

    print(f"Junie prompt saved to {prompt_file}")
    print("Use this prompt with Junie to implement the issue")

    # In a real implementation, this would integrate with Junie directly
    # For this example, we'll just provide instructions
    print("\nTo implement the issue with Junie:")
    print(f"1. Open the prompt file: {prompt_file}")
    print("2. Copy the prompt and use it with Junie")
    print("3. Follow Junie's guidance to implement the issue")
    print("4. Update the issue comment with the implementation summary")
    print("5. Close the issue if resolved")
    print("6. Commit the changes to git")

    return 0

if __name__ == "__main__":
    sys.exit(main())
