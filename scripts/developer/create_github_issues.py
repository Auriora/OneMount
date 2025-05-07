#!/usr/bin/env python3
"""
GitHub Issue Creator

This script creates GitHub issues using the format from the 'data/github_issues_7MAY25.json' file,
but excludes the ID field. It uses the GitHub API to create issues and includes proper error
handling and authentication.

Requirements:
    - Python 3.6+
    - Required Python packages:
      - requests

    You can install the required packages using pip:
    ```bash
    pip install requests
    ```

Usage:
    python create_github_issues.py --repo owner/repo --token YOUR_TOKEN [--file path/to/issues.json] [--dry-run]

Arguments:
    --repo: The GitHub repository in the format 'owner/repo'
    --token: Your GitHub personal access token
    --file: (Optional) Path to a JSON file containing issues in the same format as 'data/github_issues_7MAY25.json'
    --dry-run: (Optional) Print the issues that would be created without actually creating them

Examples:
    # Dry run to see what would be created without actually creating issues
    python create_github_issues.py --repo username/repo --token ghp_abc123 --dry-run

    # Create issues from a sample file
    python create_github_issues.py --repo username/repo --token ghp_abc123 --file scripts/developer/sample_issues.json

    # Create issues from the default file
    python create_github_issues.py --repo username/repo --token ghp_abc123

GitHub Personal Access Token:
    To create a GitHub personal access token:

    1. Go to GitHub Settings > Developer settings > Personal access tokens (https://github.com/settings/tokens)
    2. Click "Generate new token"
    3. Give your token a descriptive name
    4. Select the "repo" scope (this allows the token to create issues)
    5. Click "Generate token"
    6. Copy the token and use it with the `--token` argument

    Note: Keep your token secure and do not commit it to version control.

Input File Format:
    The input file should be a JSON array of issue objects. Each issue object should have the following structure:

    ```json
    {
      "title": "Issue title",
      "body": "Issue description with markdown formatting",
      "labels": [
        {
          "name": "label-name",
          "description": "Label description",
          "color": "color-code"
        }
      ],
      "assignees": ["username1", "username2"],
      "state": "OPEN",
      "closed": false,
      "comments": []
    }
    ```

    The script will extract the necessary fields (title, body, labels, assignees) and exclude the ID field when creating the issues.

Sample Data:
    A sample data file is provided at `scripts/developer/sample_issues.json` for testing purposes.
    This file contains sample issues with realistic content that follows the template format.

Notes:
    - The script includes rate limiting consideration with a 1-second delay between issue creation requests
      to avoid hitting GitHub API rate limits.
    - Error handling is included for file loading and API requests.
    - The script will not use the actual 'data/github_issues_7MAY25.json' file for testing since those issues already exist.
"""

import argparse
import json
import requests
import sys
import time
from typing import Dict, List, Any, Optional

def load_issues(file_path: str) -> List[Dict[str, Any]]:
    """
    Load issues from a JSON file.

    Args:
        file_path: Path to the JSON file containing issues

    Returns:
        List of issue dictionaries
    """
    try:
        with open(file_path, 'r') as f:
            issues = json.load(f)
        return issues
    except FileNotFoundError:
        print(f"Error: File '{file_path}' not found.")
        sys.exit(1)
    except json.JSONDecodeError:
        print(f"Error: File '{file_path}' is not valid JSON.")
        sys.exit(1)

def prepare_issue_for_creation(issue: Dict[str, Any]) -> Dict[str, Any]:
    """
    Prepare an issue for creation by excluding the ID field and formatting it for the GitHub API.

    Args:
        issue: The issue dictionary from the JSON file

    Returns:
        A dictionary formatted for the GitHub API
    """
    # Create a new dictionary with only the fields needed for issue creation
    prepared_issue = {
        "title": issue.get("title", ""),
        "body": issue.get("body", ""),
    }

    # Add labels if they exist
    if "labels" in issue and issue["labels"]:
        # GitHub API expects just the label names, not the full label objects
        prepared_issue["labels"] = [label["name"] for label in issue["labels"] if "name" in label]

    # Add assignees if they exist
    if "assignees" in issue and issue["assignees"]:
        prepared_issue["assignees"] = issue["assignees"]

    # Note: We're excluding the 'number' field as requested
    # Also excluding 'closed', 'comments', 'state', and 'url' as they're not needed for creation

    return prepared_issue

def create_github_issue(repo: str, token: str, issue_data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
    """
    Create a GitHub issue using the GitHub API.

    Args:
        repo: The GitHub repository in the format 'owner/repo'
        token: GitHub personal access token
        issue_data: Issue data formatted for the GitHub API

    Returns:
        The created issue data if successful, None otherwise
    """
    url = f"https://api.github.com/repos/{repo}/issues"
    headers = {
        "Authorization": f"token {token}",
        "Accept": "application/vnd.github.v3+json"
    }

    try:
        response = requests.post(url, json=issue_data, headers=headers)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error creating issue '{issue_data.get('title')}': {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response: {e.response.text}")
        return None

def main():
    """Main function to parse arguments and create GitHub issues."""
    parser = argparse.ArgumentParser(description="Create GitHub issues from a JSON file.")
    parser.add_argument("--repo", required=True, help="GitHub repository in the format 'owner/repo'")
    parser.add_argument("--token", required=True, help="GitHub personal access token")
    parser.add_argument("--file", default="data/github_issues_7MAY25.json", 
                        help="Path to JSON file containing issues (default: data/github_issues_7MAY25.json)")
    parser.add_argument("--dry-run", action="store_true", 
                        help="Print issues without creating them")

    args = parser.parse_args()

    # Load issues from the JSON file
    issues = load_issues(args.file)
    print(f"Loaded {len(issues)} issues from {args.file}")

    # Create a sample issue for testing if no issues are found
    if not issues:
        print("No issues found in the file. Creating a sample issue for testing.")
        issues = [{
            "title": "Sample Issue for Testing",
            "body": "This is a sample issue created for testing the script.",
            "labels": [{"name": "test"}],
            "assignees": []
        }]

    # Process each issue
    created_count = 0
    for i, issue in enumerate(issues):
        # Prepare the issue for creation
        prepared_issue = prepare_issue_for_creation(issue)

        if args.dry_run:
            print(f"\nIssue {i+1}:")
            print(json.dumps(prepared_issue, indent=2))
        else:
            print(f"\nCreating issue {i+1}: {prepared_issue['title']}")
            created_issue = create_github_issue(args.repo, args.token, prepared_issue)

            if created_issue:
                print(f"Successfully created issue #{created_issue['number']}: {created_issue['title']}")
                created_count += 1

                # Sleep to avoid hitting rate limits
                if i < len(issues) - 1:
                    time.sleep(1)

    if args.dry_run:
        print(f"\nDry run completed. {len(issues)} issues would be created.")
    else:
        print(f"\nCreated {created_count} out of {len(issues)} issues.")

if __name__ == "__main__":
    main()
