#!/usr/bin/env python3
"""
OneMount Development CLI Tool

A unified command-line interface for OneMount development, build, and testing operations.
Built with Typer and Rich for a modern CLI experience.

Usage:
    python scripts/dev.py [COMMAND] [OPTIONS]
    # or make it executable:
    ./scripts/dev.py [COMMAND] [OPTIONS]

Commands:
    info        - Show development environment information
    build       - Build and packaging operations
    test        - Testing and quality assurance
    release     - Release management and version control
    github      - GitHub integration and issue management
    deploy      - Deployment and CI/CD operations
    analyze     - Code and project analysis tools
    clean       - Cleanup operations

Examples:
    ./scripts/dev.py info
    ./scripts/dev.py build deb --docker
    ./scripts/dev.py test coverage --threshold 80
    ./scripts/dev.py release bump minor
    ./scripts/dev.py clean all
"""

import sys
from pathlib import Path

# Add the scripts directory to Python path for imports
SCRIPTS_DIR = Path(__file__).parent
sys.path.insert(0, str(SCRIPTS_DIR))

import typer
from rich.console import Console
from rich.traceback import install

# Install rich traceback handler for better error display
install(show_locals=True)

# Initialize Rich console
console = Console()

# Import command modules
from commands.build_commands import build_app
from commands.test_commands import test_app
from commands.release_commands import release_app
from commands.github_commands import github_app
from commands.deploy_commands import deploy_app
from commands.analyze_commands import analyze_app
from commands.clean_commands import clean_app
from commands.ci_commands import ci_app

# Import utilities
from utils.environment import show_environment_info

# Create the main Typer app
app = typer.Typer(
    name="dev",
    help="OneMount Development CLI Tool",
    epilog="For more information, visit: https://github.com/Auriora/OneMount",
    no_args_is_help=True,
    rich_markup_mode="rich",
    context_settings={"help_option_names": ["-h", "--help"]},
)

# Add command groups
app.add_typer(build_app, name="build", help="🔨 Build and packaging operations")
app.add_typer(test_app, name="test", help="🧪 Testing and quality assurance")
app.add_typer(release_app, name="release", help="🚀 Release management and version control")
app.add_typer(github_app, name="github", help="🐙 GitHub integration and issue management")
app.add_typer(deploy_app, name="deploy", help="🚢 Deployment and CI/CD operations")
app.add_typer(analyze_app, name="analyze", help="📊 Code and project analysis tools")
app.add_typer(clean_app, name="clean", help="🧹 Cleanup operations")
app.add_typer(ci_app, name="ci", help="🔄 CI setup and management")


@app.command()
def info():
    """
    📋 Show comprehensive development environment information.

    Displays information about:
    - Python environment and dependencies
    - Go toolchain and version
    - Git repository status
    - Project structure and paths
    - Available development tools
    """
    show_environment_info()


@app.command()
def completion(
    shell: str = typer.Argument(help="Shell type (bash/zsh/fish/powershell)")
):
    """
    🔧 Generate shell completion scripts.

    Generate shell completion scripts for the OneMount development CLI.

    Examples:
        # Install bash completion
        ./scripts/dev completion bash > ~/.local/share/bash-completion/completions/dev

        # Install zsh completion
        ./scripts/dev completion zsh > ~/.local/share/zsh/site-functions/_dev

        # Install fish completion
        ./scripts/dev completion fish > ~/.config/fish/completions/dev.fish
    """
    import subprocess
    import sys

    valid_shells = ["bash", "zsh", "fish", "powershell"]
    if shell not in valid_shells:
        console.print(f"[red]Invalid shell: {shell}. Must be one of: {', '.join(valid_shells)}[/red]")
        raise typer.Exit(1)

    try:
        # Generate completion using typer's built-in completion
        result = subprocess.run(
            [sys.executable, __file__, "--show-completion", shell],
            capture_output=True,
            text=True,
            check=True
        )

        console.print(result.stdout, end="")

    except subprocess.CalledProcessError as e:
        console.print(f"[red]Failed to generate completion for {shell}: {e}[/red]")
        raise typer.Exit(1)


@app.callback()
def main(
    ctx: typer.Context,
    verbose: bool = typer.Option(
        False, 
        "--verbose", 
        "-v", 
        help="Enable verbose output for debugging"
    ),
):
    """
    OneMount Development CLI Tool
    
    A unified interface for all OneMount development tasks including building,
    testing, releasing, and project management.
    """
    # Store verbose flag in context for use by subcommands
    ctx.ensure_object(dict)
    ctx.obj["verbose"] = verbose
    
    if verbose:
        console.print("[dim]Verbose mode enabled[/dim]")


if __name__ == "__main__":
    app()
