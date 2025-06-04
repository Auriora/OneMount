"""
Environment validation and information utilities for OneMount development CLI.
"""

import os
import shutil
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import typer
from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.text import Text

from .paths import get_project_paths
from .git import get_git_info

console = Console()


def check_tool_version(tool: str, version_flag: str = "--version") -> Tuple[bool, str]:
    """
    Check if a tool is available and get its version.
    
    Returns:
        Tuple of (is_available, version_string)
    """
    if not shutil.which(tool):
        return False, "Not installed"
    
    try:
        result = subprocess.run(
            [tool, version_flag],
            capture_output=True,
            text=True,
            timeout=5
        )
        if result.returncode == 0:
            # Extract first line and clean it up
            version = result.stdout.strip().split('\n')[0]
            return True, version
        else:
            return True, "Unknown version"
    except (subprocess.TimeoutExpired, subprocess.CalledProcessError, FileNotFoundError):
        return True, "Version check failed"


def check_python_environment() -> Dict[str, str]:
    """Check Python environment details."""
    env_info = {
        "Python Version": f"{sys.version_info.major}.{sys.version_info.minor}.{sys.version_info.micro}",
        "Python Executable": sys.executable,
        "Virtual Environment": "None",
        "Platform": sys.platform,
    }
    
    # Check for virtual environment
    if hasattr(sys, 'real_prefix') or (hasattr(sys, 'base_prefix') and sys.base_prefix != sys.prefix):
        venv_path = os.environ.get('VIRTUAL_ENV', sys.prefix)
        env_info["Virtual Environment"] = venv_path
    
    return env_info


def check_required_tools() -> Dict[str, Tuple[bool, str]]:
    """Check availability and versions of required development tools."""
    tools = {
        "go": ("--version", True),
        "git": ("--version", True),
        "make": ("--version", True),
        "docker": ("--version", False),
        "python3": ("--version", True),
        "pip": ("--version", False),
    }
    
    results = {}
    for tool, (flag, required) in tools.items():
        available, version = check_tool_version(tool, flag)
        results[tool] = (available, version, required)
    
    return results


def check_go_environment() -> Dict[str, str]:
    """Check Go environment details."""
    go_info = {}
    
    try:
        # Get Go version
        result = subprocess.run(["go", "version"], capture_output=True, text=True, timeout=5)
        if result.returncode == 0:
            go_info["Version"] = result.stdout.strip()
        
        # Get GOPATH
        result = subprocess.run(["go", "env", "GOPATH"], capture_output=True, text=True, timeout=5)
        if result.returncode == 0:
            go_info["GOPATH"] = result.stdout.strip()
        
        # Get GOROOT
        result = subprocess.run(["go", "env", "GOROOT"], capture_output=True, text=True, timeout=5)
        if result.returncode == 0:
            go_info["GOROOT"] = result.stdout.strip()
        
        # Get module info if in a Go module
        paths = get_project_paths()
        if paths["go_mod"].exists():
            result = subprocess.run(
                ["go", "list", "-m"],
                cwd=paths["project_root"],
                capture_output=True,
                text=True,
                timeout=5
            )
            if result.returncode == 0:
                go_info["Current Module"] = result.stdout.strip()
    
    except (subprocess.TimeoutExpired, subprocess.CalledProcessError, FileNotFoundError):
        go_info["Error"] = "Failed to get Go environment info"
    
    return go_info


def validate_project_structure() -> Dict[str, bool]:
    """Validate that we're in a proper OneMount project directory."""
    paths = get_project_paths()
    
    validations = {
        "Project Root": paths["project_root"].exists(),
        "Go Module": paths["go_mod"].exists(),
        "Makefile": paths["makefile"].exists(),
        "Scripts Directory": paths["scripts_dir"].exists(),
        "Internal Directory": paths["internal_dir"].exists(),
        "Cmd Directory": paths["cmd_dir"].exists(),
        "Pkg Directory": paths["pkg_dir"].exists(),
    }
    
    # Check if go.mod contains the correct module
    if validations["Go Module"]:
        try:
            go_mod_content = paths["go_mod"].read_text()
            validations["Correct Module"] = "github.com/auriora/onemount" in go_mod_content
        except Exception:
            validations["Correct Module"] = False
    else:
        validations["Correct Module"] = False
    
    return validations


def show_environment_info():
    """Display comprehensive development environment information."""
    console.print(Panel.fit(
        "[bold blue]OneMount Development Environment Information[/bold blue]",
        border_style="blue"
    ))
    
    # Python Environment
    console.print("\n[bold cyan]🐍 Python Environment[/bold cyan]")
    python_table = Table(show_header=False, box=None, padding=(0, 2))
    python_table.add_column("Property", style="dim")
    python_table.add_column("Value", style="green")
    
    python_info = check_python_environment()
    for key, value in python_info.items():
        python_table.add_row(key, value)
    
    console.print(python_table)
    
    # Go Environment
    console.print("\n[bold cyan]🔧 Go Environment[/bold cyan]")
    go_table = Table(show_header=False, box=None, padding=(0, 2))
    go_table.add_column("Property", style="dim")
    go_table.add_column("Value", style="green")
    
    go_info = check_go_environment()
    for key, value in go_info.items():
        go_table.add_row(key, value)
    
    console.print(go_table)
    
    # Development Tools
    console.print("\n[bold cyan]🛠️  Development Tools[/bold cyan]")
    tools_table = Table()
    tools_table.add_column("Tool", style="cyan")
    tools_table.add_column("Status", style="green")
    tools_table.add_column("Version", style="dim")
    tools_table.add_column("Required", style="yellow")
    
    tools_info = check_required_tools()
    for tool, (available, version, required) in tools_info.items():
        status = "✅ Available" if available else "❌ Missing"
        required_text = "Yes" if required else "Optional"
        
        if not available and required:
            status = "[red]❌ Missing[/red]"
            required_text = "[red]Required[/red]"
        
        tools_table.add_row(tool, status, version, required_text)
    
    console.print(tools_table)
    
    # Project Structure
    console.print("\n[bold cyan]📁 Project Structure[/bold cyan]")
    structure_table = Table()
    structure_table.add_column("Component", style="cyan")
    structure_table.add_column("Status", style="green")
    structure_table.add_column("Path", style="dim")
    
    paths = get_project_paths()
    validations = validate_project_structure()
    
    for component, is_valid in validations.items():
        status = "✅ Found" if is_valid else "❌ Missing"
        
        # Get the corresponding path
        path_key = component.lower().replace(" ", "_")
        path = paths.get(path_key, "")
        
        structure_table.add_row(component, status, str(path))
    
    console.print(structure_table)
    
    # Git Information
    console.print("\n[bold cyan]📝 Git Repository[/bold cyan]")
    git_table = Table(show_header=False, box=None, padding=(0, 2))
    git_table.add_column("Property", style="dim")
    git_table.add_column("Value", style="green")
    
    git_info = get_git_info()
    for key, value in git_info.items():
        git_table.add_row(key, str(value))
    
    console.print(git_table)
    
    # Summary
    console.print("\n[bold cyan]📊 Summary[/bold cyan]")
    
    # Check for critical issues
    critical_issues = []
    
    # Check required tools
    for tool, (available, _, required) in tools_info.items():
        if required and not available:
            critical_issues.append(f"Missing required tool: {tool}")
    
    # Check project structure
    if not validations.get("Correct Module", False):
        critical_issues.append("Not in OneMount project directory")
    
    if critical_issues:
        console.print("[red]❌ Critical Issues Found:[/red]")
        for issue in critical_issues:
            console.print(f"  • {issue}")
        console.print("\n[yellow]Please resolve these issues before proceeding.[/yellow]")
    else:
        console.print("[green]✅ Environment looks good! Ready for development.[/green]")


def ensure_environment() -> bool:
    """
    Ensure the development environment is properly set up.
    
    Returns:
        True if environment is ready, False otherwise
    """
    # Check required tools
    tools_info = check_required_tools()
    missing_required = [
        tool for tool, (available, _, required) in tools_info.items()
        if required and not available
    ]
    
    if missing_required:
        console.print(f"[red]Missing required tools: {', '.join(missing_required)}[/red]")
        return False
    
    # Check project structure
    validations = validate_project_structure()
    if not validations.get("Correct Module", False):
        console.print("[red]Not in OneMount project directory[/red]")
        return False
    
    return True
