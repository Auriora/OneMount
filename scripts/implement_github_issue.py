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
REPO_ROOT = Path(__file__).resolve().parent.parent.parent.parent

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

    # Add issue-specific information at the end of the template
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
