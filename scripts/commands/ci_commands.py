"""
CI commands for OneMount development CLI.
Handles CI setup and management operations.
"""

import sys
from pathlib import Path
from typing import Optional

import typer
from rich.console import Console

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from utils.environment import ensure_environment
from utils.ci_setup import check_auth, generate_secret, verify_setup, run_full_setup, show_status

console = Console()

# Create the CI app
ci_app = typer.Typer(help="🔄 CI setup and management commands")


@ci_app.command()
def check_auth_cmd(
    ctx: typer.Context,
):
    """
    🔐 Check if OneMount authentication is available.
    
    Verifies that valid OneDrive authentication tokens exist
    and are not expired.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Checking OneMount authentication...[/blue]")
    
    success = check_auth(verbose=verbose)
    
    if not success:
        console.print("[red]Authentication check failed[/red]")
        raise typer.Exit(1)
    
    console.print("[green]✅ Authentication check passed![/green]")


@ci_app.command()
def generate_secret_cmd(
    ctx: typer.Context,
):
    """
    🔑 Generate GitHub secret value for CI.
    
    Creates a base64-encoded secret value from your OneDrive
    authentication tokens for use in GitHub Actions.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Generating GitHub secret value...[/blue]")
    
    success = generate_secret(verbose=verbose)
    
    if not success:
        console.print("[red]Secret generation failed[/red]")
        raise typer.Exit(1)
    
    console.print("[green]✅ Secret generated successfully![/green]")


@ci_app.command()
def verify_setup_cmd(
    ctx: typer.Context,
):
    """
    ✅ Verify the complete CI setup.
    
    Checks that all components are in place for CI system tests:
    - Authentication tokens
    - Workflow files
    - OneDrive access
    - Test runner availability
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Verifying CI setup...[/blue]")
    
    success = verify_setup(verbose=verbose)
    
    if not success:
        console.print("[red]CI setup verification failed[/red]")
        raise typer.Exit(1)
    
    console.print("[green]✅ CI setup verification passed![/green]")


@ci_app.command()
def setup(
    ctx: typer.Context,
):
    """
    🚀 Run the complete CI setup process.
    
    Performs all steps needed to set up CI system tests:
    1. Check authentication
    2. Generate GitHub secret
    3. Verify complete setup
    
    This is the main command for setting up CI from scratch.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Running complete CI setup...[/blue]")
    
    success = run_full_setup(verbose=verbose)
    
    if not success:
        console.print("[red]CI setup failed[/red]")
        raise typer.Exit(1)
    
    console.print("[green]✅ CI setup completed successfully![/green]")


@ci_app.command()
def status(
    ctx: typer.Context,
):
    """
    📊 Show CI setup status.
    
    Displays the current status of all CI components:
    - Authentication status
    - Workflow file presence
    - Test runner availability
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    show_status(verbose=verbose)


# Alias commands for better UX
@ci_app.command(name="check-auth")
def check_auth_alias(ctx: typer.Context):
    """🔐 Check if OneMount authentication is available (alias)."""
    check_auth_cmd(ctx)


@ci_app.command(name="generate-secret")
def generate_secret_alias(ctx: typer.Context):
    """🔑 Generate GitHub secret value for CI (alias)."""
    generate_secret_cmd(ctx)


@ci_app.command(name="verify-setup")
def verify_setup_alias(ctx: typer.Context):
    """✅ Verify the complete CI setup (alias)."""
    verify_setup_cmd(ctx)


if __name__ == "__main__":
    ci_app()
