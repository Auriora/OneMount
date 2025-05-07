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

    If an issue in the JSON file already has an "id" field, it will be skipped as it's assumed
    the issue has already been created on GitHub.

    After successfully creating an issue on GitHub, the script will update the JSON file with the
    issue's ID to prevent duplicate creation in future runs.

Sample Data:
    A sample data file is provided at `scripts/developer/sample_issues.json` for testing purposes.
    This file contains sample issues with realistic content that follows the template format.

Label Management:
    The script downloads the list of labels from the GitHub repository and stores them in a JSON file
    in the 'data/' folder. It then checks if the labels used in the issues exist in the repository.
    If a label doesn't exist, it will be created before creating the issues.

Notes:
    - The script includes rate limiting consideration with a 1-second delay between issue creation requests
      to avoid hitting GitHub API rate limits.
    - Error handling is included for file loading and API requests.
    - The script will not use the actual 'data/github_issues_7MAY25.json' file for testing since those issues already exist.
    - Issues with an existing ID field in the JSON file will be skipped.
    - The JSON file will be updated with the IDs of newly created issues.
"""

import argparse
import json
import os
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

def update_json_file(file_path: str, issues: List[Dict[str, Any]]):
    """
    Update the JSON file with the updated issues.

    Args:
        file_path: Path to the JSON file
        issues: List of updated issue dictionaries
    """
    try:
        with open(file_path, 'w') as f:
            json.dump(issues, f, indent=2)
        print(f"Updated {file_path} with issue IDs")
    except Exception as e:
        print(f"Error updating {file_path}: {e}")

def get_github_labels(repo: str, token: str) -> List[Dict[str, Any]]:
    """
    Get the list of labels from a GitHub repository.

    Args:
        repo: The GitHub repository in the format 'owner/repo'
        token: GitHub personal access token

    Returns:
        List of label dictionaries
    """
    url = f"https://api.github.com/repos/{repo}/labels"
    headers = {
        "Authorization": f"token {token}",
        "Accept": "application/vnd.github.v3+json"
    }

    labels = []
    page = 1
    per_page = 100

    try:
        while True:
            response = requests.get(f"{url}?page={page}&per_page={per_page}", headers=headers)
            response.raise_for_status()

            page_labels = response.json()
            if not page_labels:
                break

            labels.extend(page_labels)
            page += 1

            # Sleep to avoid hitting rate limits
            time.sleep(0.5)
    except requests.exceptions.RequestException as e:
        print(f"Error getting labels: {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response: {e.response.text}")
        return []

    return labels

def save_labels_to_file(labels: List[Dict[str, Any]], file_path: str = "data/github_labels.json"):
    """
    Save the list of labels to a JSON file.

    Args:
        labels: List of label dictionaries
        file_path: Path to the JSON file (default: data/github_labels.json)
    """
    # Create the directory if it doesn't exist
    os.makedirs(os.path.dirname(file_path), exist_ok=True)

    try:
        with open(file_path, 'w') as f:
            json.dump(labels, f, indent=2)
        print(f"Saved {len(labels)} labels to {file_path}")
    except Exception as e:
        print(f"Error saving labels to {file_path}: {e}")

def create_github_label(repo: str, token: str, label_data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
    """
    Create a GitHub label using the GitHub API.

    Args:
        repo: The GitHub repository in the format 'owner/repo'
        token: GitHub personal access token
        label_data: Label data formatted for the GitHub API

    Returns:
        The created label data if successful, None otherwise
    """
    url = f"https://api.github.com/repos/{repo}/labels"
    headers = {
        "Authorization": f"token {token}",
        "Accept": "application/vnd.github.v3+json"
    }

    try:
        response = requests.post(url, json=label_data, headers=headers)
        response.raise_for_status()
        return response.json()
    except requests.exceptions.RequestException as e:
        print(f"Error creating label '{label_data.get('name')}': {e}")
        if hasattr(e, 'response') and e.response is not None:
            print(f"Response: {e.response.text}")
        return None

def ensure_labels_exist(repo: str, token: str, required_labels: List[Dict[str, Any]], existing_labels: List[Dict[str, Any]]) -> bool:
    """
    Ensure that all required labels exist in the repository.

    Args:
        repo: The GitHub repository in the format 'owner/repo'
        token: GitHub personal access token
        required_labels: List of required label dictionaries
        existing_labels: List of existing label dictionaries

    Returns:
        True if all required labels exist or were created, False otherwise
    """
    existing_label_names = {label["name"] for label in existing_labels}
    all_labels_exist = True

    for label in required_labels:
        if label["name"] not in existing_label_names:
            print(f"Label '{label['name']}' does not exist. Creating...")

            # Prepare label data for creation
            label_data = {
                "name": label["name"],
                "color": label.get("color", "ededed"),  # Default color if not provided
                "description": label.get("description", "")
            }

            created_label = create_github_label(repo, token, label_data)
            if created_label:
                print(f"Successfully created label '{label['name']}'")
                existing_labels.append(created_label)
                existing_label_names.add(label["name"])
            else:
                print(f"Failed to create label '{label['name']}'")
                all_labels_exist = False

            # Sleep to avoid hitting rate limits
            time.sleep(0.5)

    return all_labels_exist

def main():
    """Main function to parse arguments and create GitHub issues."""
    parser = argparse.ArgumentParser(description="Create GitHub issues from a JSON file.")
    parser.add_argument("--repo", required=True, help="GitHub repository in the format 'owner/repo'")
    parser.add_argument("--token", required=True, help="GitHub personal access token")
    parser.add_argument("--file", default="data/github_issues_7MAY25.json", 
                        help="Path to JSON file containing issues (default: data/github_issues_7MAY25.json)")
    parser.add_argument("--dry-run", action="store_true", 
                        help="Print issues without creating them")
    parser.add_argument("--labels-file", default="data/github_labels.json",
                        help="Path to JSON file to store GitHub labels (default: data/github_labels.json)")

    args = parser.parse_args()

    # Download labels from GitHub and save them to a file
    print(f"Downloading labels from GitHub repository {args.repo}...")
    labels = get_github_labels(args.repo, args.token)
    if labels:
        save_labels_to_file(labels, args.labels_file)
    else:
        print("Warning: Failed to download labels from GitHub. Label creation may fail.")
        labels = []

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

    # Extract all unique labels from the issues
    required_labels = []
    for issue in issues:
        if "labels" in issue and issue["labels"]:
            for label in issue["labels"]:
                if label not in required_labels and "name" in label:
                    required_labels.append(label)

    # Ensure all required labels exist
    if required_labels and not args.dry_run:
        print(f"Ensuring {len(required_labels)} labels exist in the repository...")
        ensure_labels_exist(args.repo, args.token, required_labels, labels)

    # Process each issue
    created_count = 0
    skipped_count = 0
    file_updated = False

    for i, issue in enumerate(issues):
        # Skip issues that already have an ID
        if "id" in issue:
            print(f"\nSkipping issue {i+1}: {issue['title']} (already has ID: {issue['id']})")
            skipped_count += 1
            continue

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

                # Update the issue in the original list with the ID from GitHub
                issues[i]["id"] = created_issue["number"]
                file_updated = True

                created_count += 1

                # Sleep to avoid hitting rate limits
                if i < len(issues) - 1:
                    time.sleep(1)

    # Update the JSON file with the new IDs if any issues were created
    if file_updated and not args.dry_run:
        update_json_file(args.file, issues)

    if args.dry_run:
        print(f"\nDry run completed. {len(issues) - skipped_count} issues would be created, {skipped_count} would be skipped.")
    else:
        print(f"\nCreated {created_count} issues, skipped {skipped_count} issues with existing IDs.")

if __name__ == "__main__":
    main()
