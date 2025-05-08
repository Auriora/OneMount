#!/usr/bin/env python3
import subprocess
import re
import os

def get_github_repo_from_git():
    """
    Get the GitHub repository from the git remote origin URL of the current directory.
    
    Returns:
        str: The GitHub repository in the format 'owner/repo' if successful, None otherwise
    """
    try:
        # Check if the current directory is within a git repository
        git_check = subprocess.run(['git', 'rev-parse', '--is-inside-work-tree'], 
                                  capture_output=True, text=True)
        if git_check.returncode != 0 or git_check.stdout.strip() != 'true':
            print("Current directory is not within a git repository")
            return None
        
        # Get the remote origin URL
        git_remote = subprocess.run(['git', 'config', '--get', 'remote.origin.url'], 
                                   capture_output=True, text=True)
        if git_remote.returncode != 0:
            print("Failed to get remote origin URL")
            return None
        
        remote_url = git_remote.stdout.strip()
        print(f"Remote origin URL: {remote_url}")
        
        # Parse the URL to get the GitHub repository
        # Handle different URL formats:
        # - HTTPS: https://github.com/owner/repo.git
        # - SSH: git@github.com:owner/repo.git
        
        # HTTPS format
        https_match = re.match(r'https://github\.com/([^/]+)/([^/.]+)(?:\.git)?', remote_url)
        if https_match:
            owner, repo = https_match.groups()
            return f"{owner}/{repo}"
        
        # SSH format
        ssh_match = re.match(r'git@github\.com:([^/]+)/([^/.]+)(?:\.git)?', remote_url)
        if ssh_match:
            owner, repo = ssh_match.groups()
            return f"{owner}/{repo}"
        
        print(f"Could not parse GitHub repository from URL: {remote_url}")
        return None
    
    except Exception as e:
        print(f"Error getting GitHub repository from git: {e}")
        return None

if __name__ == "__main__":
    print("Testing GitHub repository detection from git...")
    repo = get_github_repo_from_git()
    if repo:
        print(f"Successfully detected GitHub repository: {repo}")
    else:
        print("Failed to detect GitHub repository")