"""
Git operations and information utilities for OneMount development CLI.
"""

import subprocess
from pathlib import Path
from typing import Dict, List, Optional

from .paths import get_project_paths


def run_git_command(args: List[str], cwd: Optional[Path] = None) -> Optional[str]:
    """
    Run a git command and return the output.
    
    Args:
        args: Git command arguments (without 'git')
        cwd: Working directory (defaults to project root)
    
    Returns:
        Command output as string, or None if command failed
    """
    if cwd is None:
        cwd = get_project_paths()["project_root"]
    
    try:
        result = subprocess.run(
            ["git"] + args,
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            return result.stdout.strip()
        return None
    except (subprocess.TimeoutExpired, subprocess.CalledProcessError, FileNotFoundError):
        return None


def get_git_info() -> Dict[str, str]:
    """Get comprehensive Git repository information."""
    info = {}
    
    # Basic repository info
    info["Repository"] = run_git_command(["remote", "get-url", "origin"]) or "Unknown"
    info["Current Branch"] = run_git_command(["branch", "--show-current"]) or "Unknown"
    info["Latest Commit"] = run_git_command(["rev-parse", "HEAD"]) or "Unknown"
    info["Latest Commit (Short)"] = run_git_command(["rev-parse", "--short", "HEAD"]) or "Unknown"
    
    # Working directory status
    status_output = run_git_command(["status", "--porcelain"])
    if status_output is not None:
        if status_output:
            info["Working Directory"] = "Modified files present"
            info["Modified Files"] = str(len(status_output.split('\n')))
        else:
            info["Working Directory"] = "Clean"
            info["Modified Files"] = "0"
    else:
        info["Working Directory"] = "Unknown"
        info["Modified Files"] = "Unknown"
    
    # Branch tracking info
    tracking_info = run_git_command(["status", "-b", "--porcelain"])
    if tracking_info:
        first_line = tracking_info.split('\n')[0]
        if "ahead" in first_line or "behind" in first_line:
            info["Branch Status"] = first_line.replace("## ", "")
        else:
            info["Branch Status"] = "Up to date"
    else:
        info["Branch Status"] = "Unknown"
    
    # Last commit info
    last_commit_msg = run_git_command(["log", "-1", "--pretty=format:%s"])
    if last_commit_msg:
        info["Last Commit Message"] = last_commit_msg[:60] + ("..." if len(last_commit_msg) > 60 else "")
    
    last_commit_author = run_git_command(["log", "-1", "--pretty=format:%an"])
    if last_commit_author:
        info["Last Commit Author"] = last_commit_author
    
    last_commit_date = run_git_command(["log", "-1", "--pretty=format:%cr"])
    if last_commit_date:
        info["Last Commit Date"] = last_commit_date
    
    # Tag info
    latest_tag = run_git_command(["describe", "--tags", "--abbrev=0"])
    if latest_tag:
        info["Latest Tag"] = latest_tag
        
        # Distance from latest tag
        tag_distance = run_git_command(["rev-list", f"{latest_tag}..HEAD", "--count"])
        if tag_distance:
            info["Commits Since Tag"] = tag_distance
    
    return info


def is_git_repository() -> bool:
    """Check if current directory is a Git repository."""
    return run_git_command(["rev-parse", "--git-dir"]) is not None


def get_modified_files() -> List[str]:
    """Get list of modified files in the working directory."""
    status_output = run_git_command(["status", "--porcelain"])
    if not status_output:
        return []
    
    files = []
    for line in status_output.split('\n'):
        if line.strip():
            # Extract filename from git status output
            filename = line[3:].strip()
            files.append(filename)
    
    return files


def get_untracked_files() -> List[str]:
    """Get list of untracked files."""
    output = run_git_command(["ls-files", "--others", "--exclude-standard"])
    if not output:
        return []
    
    return output.split('\n')


def get_current_branch() -> Optional[str]:
    """Get the current Git branch name."""
    return run_git_command(["branch", "--show-current"])


def get_remote_url() -> Optional[str]:
    """Get the remote origin URL."""
    return run_git_command(["remote", "get-url", "origin"])


def get_commit_hash(short: bool = False) -> Optional[str]:
    """Get the current commit hash."""
    if short:
        return run_git_command(["rev-parse", "--short", "HEAD"])
    else:
        return run_git_command(["rev-parse", "HEAD"])


def is_working_directory_clean() -> bool:
    """Check if the working directory is clean (no uncommitted changes)."""
    status_output = run_git_command(["status", "--porcelain"])
    return status_output == ""


def get_tags() -> List[str]:
    """Get list of all tags."""
    output = run_git_command(["tag", "--list"])
    if not output:
        return []
    
    return output.split('\n')


def get_latest_tag() -> Optional[str]:
    """Get the latest tag."""
    return run_git_command(["describe", "--tags", "--abbrev=0"])


def create_tag(tag_name: str, message: Optional[str] = None) -> bool:
    """
    Create a new Git tag.
    
    Args:
        tag_name: Name of the tag to create
        message: Optional tag message
    
    Returns:
        True if tag was created successfully
    """
    args = ["tag"]
    if message:
        args.extend(["-a", tag_name, "-m", message])
    else:
        args.append(tag_name)
    
    result = run_git_command(args)
    return result is not None


def push_tags() -> bool:
    """Push all tags to remote."""
    result = run_git_command(["push", "--tags"])
    return result is not None


def get_branch_commits_ahead_behind(branch: str = "origin/main") -> tuple[int, int]:
    """
    Get number of commits ahead and behind compared to a branch.
    
    Returns:
        Tuple of (commits_ahead, commits_behind)
    """
    # Get commits ahead
    ahead_output = run_git_command(["rev-list", "--count", f"{branch}..HEAD"])
    commits_ahead = int(ahead_output) if ahead_output and ahead_output.isdigit() else 0
    
    # Get commits behind
    behind_output = run_git_command(["rev-list", "--count", f"HEAD..{branch}"])
    commits_behind = int(behind_output) if behind_output and behind_output.isdigit() else 0
    
    return commits_ahead, commits_behind
