"""
Native Python implementation for release management.
Replaces release.sh with native Python Git operations.
"""

import re
import subprocess
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import git
from packaging import version
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn

from .paths import get_project_paths
from .shell import run_command, CommandError

console = Console()


class ReleaseError(Exception):
    """Exception raised when release operations fail."""
    pass


class ReleaseManager:
    """Native Python release manager for OneMount."""
    
    def __init__(self, verbose: bool = False):
        self.verbose = verbose
        self.paths = get_project_paths()
        self.repo = None
        
    def __enter__(self):
        """Context manager entry."""
        try:
            self.repo = git.Repo(self.paths["project_root"])
            return self
        except git.InvalidGitRepositoryError:
            raise ReleaseError("Not a valid Git repository")
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        pass
    
    def _log_info(self, message: str):
        """Log info message."""
        console.print(f"[blue][INFO][/blue] {message}")
    
    def _log_success(self, message: str):
        """Log success message."""
        console.print(f"[green][SUCCESS][/green] {message}")
    
    def _log_warning(self, message: str):
        """Log warning message."""
        console.print(f"[yellow][WARNING][/yellow] {message}")
    
    def _log_error(self, message: str):
        """Log error message."""
        console.print(f"[red][ERROR][/red] {message}")
    
    def check_prerequisites(self) -> bool:
        """Check if all prerequisites for release are met."""
        # Check if we're in the right directory
        bumpversion_cfg = self.paths["project_root"] / ".bumpversion.cfg"
        common_go = self.paths["project_root"] / "cmd" / "common" / "common.go"
        
        if not bumpversion_cfg.exists():
            self._log_error("This script must be run from the OneMount project root directory")
            self._log_error(f"Missing: {bumpversion_cfg}")
            return False
        
        if not common_go.exists():
            self._log_error("This script must be run from the OneMount project root directory")
            self._log_error(f"Missing: {common_go}")
            return False
        
        # Check if bumpversion is available
        try:
            result = run_command(
                ["bumpversion", "--help"],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=5
            )
        except (CommandError, FileNotFoundError):
            self._log_error("bumpversion is not installed or not in PATH")
            self._log_error("Install with: pip install bump2version")
            return False
        
        return True
    
    def get_current_version(self) -> Optional[str]:
        """Get the current version from bumpversion config."""
        bumpversion_cfg = self.paths["project_root"] / ".bumpversion.cfg"
        
        try:
            with open(bumpversion_cfg, 'r') as f:
                content = f.read()
            
            # Look for current_version line
            match = re.search(r'^current_version\s*=\s*(.+)$', content, re.MULTILINE)
            if match:
                return match.group(1).strip()
            
            return None
            
        except Exception as e:
            self._log_error(f"Failed to read current version: {e}")
            return None
    
    def show_current_version(self):
        """Display the current version."""
        current_version = self.get_current_version()
        if current_version:
            self._log_info(f"Current version: {current_version}")
        else:
            self._log_error("Could not determine current version")
    
    def validate_bump_type(self, bump_type: str) -> bool:
        """Validate the bump type."""
        valid_types = ["num", "release", "patch", "minor", "major"]
        if bump_type not in valid_types:
            self._log_error(f"Invalid bump type: {bump_type}")
            self._log_error(f"Valid types: {', '.join(valid_types)}")
            return False
        return True
    
    def check_working_directory_clean(self) -> bool:
        """Check if the working directory is clean."""
        try:
            if self.repo.is_dirty():
                self._log_error("Working directory is not clean")
                self._log_error("Please commit or stash your changes before releasing")
                return False
            
            # Check for untracked files
            untracked = self.repo.untracked_files
            if untracked:
                self._log_warning(f"Untracked files found: {', '.join(untracked)}")
                self._log_warning("Consider adding them to .gitignore or committing them")
            
            return True
            
        except Exception as e:
            self._log_error(f"Failed to check working directory status: {e}")
            return False
    
    def bump_version(self, bump_type: str, dry_run: bool = False) -> bool:
        """Bump the version using bumpversion."""
        try:
            # Show current version
            self.show_current_version()
            
            # Build bumpversion command
            cmd = ["bumpversion"]
            
            if dry_run:
                cmd.append("--dry-run")
                cmd.append("--verbose")
            
            cmd.append(bump_type)
            
            # Run bumpversion
            if dry_run:
                self._log_info("Dry run - showing what would be done:")
            else:
                self._log_info(f"Bumping version ({bump_type})...")
            
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
                transient=True,
            ) as progress:
                task = progress.add_task("Running bumpversion...", total=None)
                
                try:
                    result = run_command(
                        cmd,
                        capture_output=True,
                        check=True,
                        verbose=self.verbose,
                        timeout=30
                    )
                    
                    if self.verbose or dry_run:
                        console.print(result.stdout)
                    
                    if dry_run:
                        progress.update(task, description="✅ Dry run completed")
                    else:
                        progress.update(task, description="✅ Version bumped successfully")
                    
                except CommandError as e:
                    progress.update(task, description="❌ Version bump failed")
                    self._log_error(f"bumpversion failed: {e}")
                    if e.stderr:
                        console.print(f"[red]Error output:[/red]\n{e.stderr}")
                    return False
            
            if not dry_run:
                self._log_success("Version bumped successfully!")
                self.show_current_version()
            
            return True
            
        except Exception as e:
            self._log_error(f"Failed to bump version: {e}")
            return False
    
    def push_tags(self, dry_run: bool = False) -> bool:
        """Push tags to GitHub to trigger package building."""
        try:
            if dry_run:
                self._log_info("Dry run - would push tags to GitHub")
                return True
            
            self._log_info("Pushing tags to GitHub to trigger package building...")
            
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
                transient=True,
            ) as progress:
                task = progress.add_task("Pushing tags...", total=None)
                
                try:
                    # Push tags using Git command (more reliable than GitPython for this)
                    result = run_command(
                        ["git", "push", "origin", "--tags"],
                        capture_output=True,
                        check=True,
                        verbose=self.verbose,
                        timeout=60
                    )
                    
                    if self.verbose:
                        console.print(result.stdout)
                    
                    progress.update(task, description="✅ Tags pushed successfully")
                    
                except CommandError as e:
                    progress.update(task, description="❌ Failed to push tags")
                    self._log_error(f"Failed to push tags: {e}")
                    if e.stderr:
                        console.print(f"[red]Error output:[/red]\n{e.stderr}")
                    self._log_info("You can push manually later with: git push origin --tags")
                    return False
            
            self._log_success("Tags pushed successfully!")
            self._log_info("GitHub Actions will now build packages and create a release")
            self._log_info("Check the progress at: https://github.com/Auriora/OneMount/actions")
            
            return True
            
        except Exception as e:
            self._log_error(f"Failed to push tags: {e}")
            return False

    def create_release(self, bump_type: str, dry_run: bool = False, no_push: bool = False) -> bool:
        """
        Create a release by bumping version and optionally pushing tags.

        Args:
            bump_type: Type of version bump (num/release/patch/minor/major)
            dry_run: Show what would be done without executing
            no_push: Don't push tags to GitHub

        Returns:
            True if release creation succeeded, False otherwise
        """
        try:
            # Check prerequisites
            if not self.check_prerequisites():
                return False

            # Validate bump type
            if not self.validate_bump_type(bump_type):
                return False

            # Check working directory (only for non-dry-run)
            if not dry_run and not self.check_working_directory_clean():
                return False

            # Bump version
            if not self.bump_version(bump_type, dry_run):
                return False

            # Push tags if requested
            if no_push:
                if not dry_run:
                    self._log_warning("Skipping tag push (--no-push specified)")
                    self._log_info("To trigger package building later, run: git push origin --tags")
            else:
                if not self.push_tags(dry_run):
                    return False

            return True

        except Exception as e:
            self._log_error(f"Failed to create release: {e}")
            return False


def create_release(bump_type: str, verbose: bool = False, dry_run: bool = False, no_push: bool = False) -> bool:
    """
    Convenience function to create a release.

    Args:
        bump_type: Type of version bump (num/release/patch/minor/major)
        verbose: Enable verbose output
        dry_run: Show what would be done without executing
        no_push: Don't push tags to GitHub

    Returns:
        True if release creation succeeded, False otherwise
    """
    with ReleaseManager(verbose=verbose) as manager:
        return manager.create_release(bump_type, dry_run=dry_run, no_push=no_push)


def get_current_version() -> Optional[str]:
    """Get the current version from bumpversion config."""
    with ReleaseManager() as manager:
        return manager.get_current_version()


def show_release_usage():
    """Show usage information for release commands."""
    console.print("[bold]Usage:[/bold] release bump <bump_type> [options]")
    console.print("")
    console.print("[bold]Bump types:[/bold]")
    console.print("  [cyan]num[/cyan]       - Bump release candidate number (0.1.0rc1 → 0.1.0rc2)")
    console.print("  [cyan]release[/cyan]   - Release current RC (0.1.0rc1 → 0.1.0)")
    console.print("  [cyan]patch[/cyan]     - Bump patch version (0.1.0 → 0.1.1)")
    console.print("  [cyan]minor[/cyan]     - Bump minor version (0.1.0 → 0.2.0)")
    console.print("  [cyan]major[/cyan]     - Bump major version (0.1.0 → 1.0.0)")
    console.print("")
    console.print("[bold]Options:[/bold]")
    console.print("  [cyan]--dry-run[/cyan]  - Show what would be done without making changes")
    console.print("  [cyan]--no-push[/cyan]  - Don't push tags to GitHub (skip package building)")
    console.print("")
    console.print("[bold]Examples:[/bold]")
    console.print("  [dim]./scripts/dev release bump num[/dim]                    # Bump RC number and trigger package build")
    console.print("  [dim]./scripts/dev release bump release --dry-run[/dim]      # Preview release without changes")
    console.print("  [dim]./scripts/dev release bump patch --no-push[/dim]       # Bump patch but don't trigger build")
