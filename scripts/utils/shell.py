"""
Shell command execution utilities for OneMount development CLI.
"""

import os
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Optional, Union

import typer
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn

from .paths import get_project_paths

console = Console()


class CommandError(Exception):
    """Exception raised when a command fails."""
    
    def __init__(self, command: List[str], returncode: int, stderr: str = ""):
        self.command = command
        self.returncode = returncode
        self.stderr = stderr
        super().__init__(f"Command failed: {' '.join(command)} (exit code: {returncode})")


def run_command(
    command: Union[str, List[str]],
    cwd: Optional[Path] = None,
    env: Optional[Dict[str, str]] = None,
    capture_output: bool = False,
    check: bool = True,
    verbose: bool = False,
    timeout: Optional[int] = None,
    input_text: Optional[str] = None,
) -> subprocess.CompletedProcess:
    """
    Run a shell command with enhanced error handling and logging.
    
    Args:
        command: Command to run (string or list of arguments)
        cwd: Working directory (defaults to project root)
        env: Environment variables (merged with current environment)
        capture_output: Whether to capture stdout/stderr
        check: Whether to raise exception on non-zero exit code
        verbose: Whether to show command being executed
        timeout: Command timeout in seconds
        input_text: Text to send to stdin
    
    Returns:
        CompletedProcess object
    
    Raises:
        CommandError: If command fails and check=True
    """
    if cwd is None:
        cwd = get_project_paths()["project_root"]
    
    # Convert string command to list
    if isinstance(command, str):
        command = command.split()
    
    # Merge environment variables
    if env:
        full_env = os.environ.copy()
        full_env.update(env)
    else:
        full_env = None
    
    if verbose:
        console.print(f"[dim]Running: {' '.join(command)}[/dim]")
        if cwd != get_project_paths()["project_root"]:
            console.print(f"[dim]Working directory: {cwd}[/dim]")
    
    try:
        result = subprocess.run(
            command,
            cwd=cwd,
            env=full_env,
            capture_output=capture_output,
            text=True,
            check=False,  # We handle checking manually
            timeout=timeout,
            input=input_text,
        )
        
        if check and result.returncode != 0:
            error_msg = result.stderr if result.stderr else "No error message"
            if verbose:
                console.print(f"[red]Command failed with exit code {result.returncode}[/red]")
                console.print(f"[red]Error: {error_msg}[/red]")
            raise CommandError(command, result.returncode, error_msg)
        
        return result
        
    except subprocess.TimeoutExpired as e:
        if verbose:
            console.print(f"[red]Command timed out after {timeout} seconds[/red]")
        raise CommandError(command, -1, f"Command timed out after {timeout} seconds")
    
    except FileNotFoundError as e:
        if verbose:
            console.print(f"[red]Command not found: {command[0]}[/red]")
        raise CommandError(command, -1, f"Command not found: {command[0]}")


def run_command_with_progress(
    command: Union[str, List[str]],
    description: str,
    cwd: Optional[Path] = None,
    env: Optional[Dict[str, str]] = None,
    verbose: bool = False,
    timeout: Optional[int] = None,
) -> subprocess.CompletedProcess:
    """
    Run a command with a progress spinner.
    
    Args:
        command: Command to run
        description: Description to show in progress spinner
        cwd: Working directory
        env: Environment variables
        verbose: Whether to show verbose output
        timeout: Command timeout in seconds
    
    Returns:
        CompletedProcess object
    """
    with Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        console=console,
        transient=True,
    ) as progress:
        task = progress.add_task(description, total=None)
        
        try:
            result = run_command(
                command=command,
                cwd=cwd,
                env=env,
                capture_output=True,
                check=True,
                verbose=verbose,
                timeout=timeout,
            )
            progress.update(task, description=f"✅ {description}")
            return result
            
        except CommandError as e:
            progress.update(task, description=f"❌ {description}")
            raise


def check_command_available(command: str) -> bool:
    """Check if a command is available in PATH."""
    import shutil
    return shutil.which(command) is not None


def get_command_version(command: str, version_flag: str = "--version") -> Optional[str]:
    """Get version of a command."""
    try:
        result = run_command(
            [command, version_flag],
            capture_output=True,
            check=True,
            verbose=False,
            timeout=5,
        )
        return result.stdout.strip().split('\n')[0]
    except (CommandError, subprocess.TimeoutExpired):
        return None


def run_make_target(
    target: str,
    verbose: bool = False,
    env: Optional[Dict[str, str]] = None,
) -> subprocess.CompletedProcess:
    """
    Run a Make target.
    
    Args:
        target: Make target to run
        verbose: Whether to show verbose output
        env: Additional environment variables
    
    Returns:
        CompletedProcess object
    """
    return run_command(
        ["make", target],
        verbose=verbose,
        env=env,
        capture_output=not verbose,
    )


def run_go_command(
    args: List[str],
    verbose: bool = False,
    env: Optional[Dict[str, str]] = None,
    cwd: Optional[Path] = None,
) -> subprocess.CompletedProcess:
    """
    Run a Go command.
    
    Args:
        args: Go command arguments (without 'go')
        verbose: Whether to show verbose output
        env: Additional environment variables
        cwd: Working directory
    
    Returns:
        CompletedProcess object
    """
    return run_command(
        ["go"] + args,
        cwd=cwd,
        verbose=verbose,
        env=env,
        capture_output=not verbose,
    )


def run_docker_command(
    args: List[str],
    verbose: bool = False,
    env: Optional[Dict[str, str]] = None,
) -> subprocess.CompletedProcess:
    """
    Run a Docker command.
    
    Args:
        args: Docker command arguments (without 'docker')
        verbose: Whether to show verbose output
        env: Additional environment variables
    
    Returns:
        CompletedProcess object
    """
    return run_command(
        ["docker"] + args,
        verbose=verbose,
        env=env,
        capture_output=not verbose,
    )


def run_script(
    script_path: Path,
    args: Optional[List[str]] = None,
    verbose: bool = False,
    env: Optional[Dict[str, str]] = None,
) -> subprocess.CompletedProcess:
    """
    Run a shell script.
    
    Args:
        script_path: Path to the script
        args: Script arguments
        verbose: Whether to show verbose output
        env: Additional environment variables
    
    Returns:
        CompletedProcess object
    """
    if not script_path.exists():
        raise CommandError([str(script_path)], -1, f"Script not found: {script_path}")
    
    command = [str(script_path)]
    if args:
        command.extend(args)
    
    return run_command(
        command,
        verbose=verbose,
        env=env,
        capture_output=not verbose,
    )


def ensure_executable(script_path: Path):
    """Ensure a script is executable."""
    if script_path.exists():
        script_path.chmod(script_path.stat().st_mode | 0o755)
