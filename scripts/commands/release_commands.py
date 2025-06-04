"""
Release management and version control commands for OneMount development CLI.
"""

import sys
from pathlib import Path
from typing import Optional

import typer
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from utils.environment import ensure_environment
from utils.paths import get_project_paths
from utils.shell import run_command, run_command_with_progress, run_script, ensure_executable
from utils.git import get_git_info, get_latest_tag, is_working_directory_clean
from utils.release_manager import create_release, get_current_version, show_release_usage

console = Console()

# Create the release app
release_app = typer.Typer(
    name="release",
    help="Release management and version control",
    no_args_is_help=True,
)


def get_current_version() -> Optional[str]:
    """Get the current version from .bumpversion.cfg."""
    paths = get_project_paths()
    bumpversion_cfg = paths["bumpversion_cfg"]
    
    if not bumpversion_cfg.exists():
        return None
    
    try:
        content = bumpversion_cfg.read_text()
        for line in content.split('\n'):
            if line.startswith('current_version'):
                return line.split('=')[1].strip()
    except Exception:
        return None
    
    return None


@release_app.command()
def bump(
    ctx: typer.Context,
    bump_type: str = typer.Argument(help="Version bump type (num/release/patch/minor/major)"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be done without executing"),
    no_push: bool = typer.Option(False, "--no-push", help="Don't push tags to GitHub"),
):
    """
    🚀 Bump version and trigger release.
    
    Bumps the version using bumpversion and optionally creates and pushes Git tags.
    
    Bump types:
    - num: Increment release candidate number (0.1.0rc1 -> 0.1.0rc2)
    - release: Release current RC to stable (0.1.0rc1 -> 0.1.0)
    - patch: Increment patch version (0.1.0 -> 0.1.1)
    - minor: Increment minor version (0.1.0 -> 0.2.0)
    - major: Increment major version (0.1.0 -> 1.0.0)
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    # Use native Python release manager
    success = create_release(
        bump_type=bump_type,
        verbose=verbose,
        dry_run=dry_run,
        no_push=no_push
    )

    if not success:
        console.print("[red]Release creation failed[/red]")
        raise typer.Exit(1)

    if dry_run:
        console.print("[yellow]Dry run completed - no changes made[/yellow]")
    else:
        console.print("[green]✅ Release created successfully![/green]")
        if not no_push:
            console.print("[blue]Tags pushed to GitHub - this will trigger package builds[/blue]")


@release_app.command()
def status(ctx: typer.Context):
    """
    📋 Show release status and version information.
    
    Display current version, Git status, and release readiness.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print(Panel.fit(
        "[bold blue]OneMount Release Status[/bold blue]",
        border_style="blue"
    ))
    
    # Version information
    console.print("\n[bold cyan]📦 Version Information[/bold cyan]")
    version_table = Table()
    version_table.add_column("Property", style="cyan")
    version_table.add_column("Value", style="green")
    
    current_version = get_current_version()
    version_table.add_row("Current Version", current_version or "Unknown")
    
    latest_tag = get_latest_tag()
    version_table.add_row("Latest Git Tag", latest_tag or "None")
    
    # Check if version matches tag
    if current_version and latest_tag:
        if current_version == latest_tag or current_version == latest_tag.lstrip('v'):
            version_status = "✅ In sync"
        else:
            version_status = "⚠️  Out of sync"
    else:
        version_status = "❓ Unknown"
    
    version_table.add_row("Version/Tag Status", version_status)
    
    console.print(version_table)
    
    # Git status
    console.print("\n[bold cyan]📝 Git Status[/bold cyan]")
    git_table = Table()
    git_table.add_column("Property", style="cyan")
    git_table.add_column("Value", style="green")
    
    git_info = get_git_info()
    
    # Key git information for releases
    git_items = [
        ("Current Branch", git_info.get("Current Branch", "Unknown")),
        ("Working Directory", git_info.get("Working Directory", "Unknown")),
        ("Branch Status", git_info.get("Branch Status", "Unknown")),
        ("Latest Commit", git_info.get("Latest Commit (Short)", "Unknown")),
        ("Last Commit Date", git_info.get("Last Commit Date", "Unknown")),
    ]
    
    for key, value in git_items:
        git_table.add_row(key, value)
    
    console.print(git_table)
    
    # Release readiness
    console.print("\n[bold cyan]🚀 Release Readiness[/bold cyan]")
    readiness_table = Table()
    readiness_table.add_column("Check", style="cyan")
    readiness_table.add_column("Status", style="green")
    readiness_table.add_column("Details", style="dim")
    
    # Check working directory
    is_clean = is_working_directory_clean()
    clean_status = "✅ Clean" if is_clean else "❌ Modified files"
    clean_details = "Ready for release" if is_clean else "Commit changes first"
    readiness_table.add_row("Working Directory", clean_status, clean_details)
    
    # Check if on main branch
    current_branch = git_info.get("Current Branch", "")
    is_main = current_branch in ["main", "master"]
    branch_status = "✅ On main" if is_main else f"⚠️  On {current_branch}"
    branch_details = "Ready for release" if is_main else "Consider switching to main"
    readiness_table.add_row("Branch", branch_status, branch_details)
    
    # Check bumpversion config
    paths = get_project_paths()
    has_bumpversion = paths["bumpversion_cfg"].exists()
    bump_status = "✅ Available" if has_bumpversion else "❌ Missing"
    bump_details = "Version bumping enabled" if has_bumpversion else "Install bumpversion"
    readiness_table.add_row("Bumpversion Config", bump_status, bump_details)
    
    console.print(readiness_table)
    
    # Overall readiness
    all_ready = is_clean and is_main and has_bumpversion
    if all_ready:
        console.print("\n[green]✅ Ready for release![/green]")
        console.print("[dim]Use './dev.py release bump <type>' to create a release[/dim]")
    else:
        console.print("\n[yellow]⚠️  Not ready for release[/yellow]")
        console.print("[dim]Address the issues above before releasing[/dim]")


@release_app.command()
def history(
    ctx: typer.Context,
    limit: int = typer.Option(10, "--limit", "-n", help="Number of releases to show"),
):
    """
    📚 Show release history.
    
    Display recent Git tags and release information.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print("[blue]Fetching release history...[/blue]")
    
    try:
        # Get all tags sorted by version
        result = run_command(
            ["git", "tag", "--sort=-version:refname"],
            capture_output=True,
            verbose=verbose,
        )
        
        if not result.stdout.strip():
            console.print("[yellow]No releases found.[/yellow]")
            return
        
        tags = result.stdout.strip().split('\n')[:limit]
        
        console.print(f"\n[bold cyan]📚 Recent Releases (last {len(tags)})[/bold cyan]")
        
        history_table = Table()
        history_table.add_column("Tag", style="cyan")
        history_table.add_column("Date", style="green")
        history_table.add_column("Commit", style="dim")
        history_table.add_column("Message", style="white")
        
        for tag in tags:
            # Get tag information
            try:
                # Get tag date
                date_result = run_command(
                    ["git", "log", "-1", "--format=%ci", tag],
                    capture_output=True,
                    verbose=False,
                )
                tag_date = date_result.stdout.strip()[:10] if date_result.stdout else "Unknown"
                
                # Get commit hash
                commit_result = run_command(
                    ["git", "rev-list", "-n", "1", tag],
                    capture_output=True,
                    verbose=False,
                )
                commit_hash = commit_result.stdout.strip()[:8] if commit_result.stdout else "Unknown"
                
                # Get commit message
                msg_result = run_command(
                    ["git", "log", "-1", "--format=%s", tag],
                    capture_output=True,
                    verbose=False,
                )
                commit_msg = msg_result.stdout.strip() if msg_result.stdout else "No message"
                
                # Truncate long messages
                if len(commit_msg) > 50:
                    commit_msg = commit_msg[:47] + "..."
                
                history_table.add_row(tag, tag_date, commit_hash, commit_msg)
            
            except Exception:
                history_table.add_row(tag, "Unknown", "Unknown", "Error getting info")
        
        console.print(history_table)
    
    except Exception as e:
        console.print(f"[red]Failed to get release history: {e}[/red]")
        raise typer.Exit(1)


@release_app.command()
def check(ctx: typer.Context):
    """
    🔍 Check release configuration and dependencies.
    
    Verify that all tools and configurations needed for releases are available.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print("[blue]Checking release configuration...[/blue]")
    
    console.print("\n[bold cyan]🔧 Release Tools[/bold cyan]")
    tools_table = Table()
    tools_table.add_column("Tool", style="cyan")
    tools_table.add_column("Status", style="green")
    tools_table.add_column("Version", style="dim")
    tools_table.add_column("Required For", style="yellow")
    
    # Check required tools
    tools = [
        ("git", "Version control"),
        ("bump2version", "Version bumping"),
        ("gh", "GitHub CLI (optional)"),
    ]
    
    import shutil
    from utils.shell import get_command_version
    
    for tool, purpose in tools:
        if shutil.which(tool):
            version = get_command_version(tool) or "Unknown"
            status = "✅ Available"
        else:
            version = "Not installed"
            status = "❌ Missing"
        
        tools_table.add_row(tool, status, version, purpose)
    
    console.print(tools_table)
    
    # Check configuration files
    console.print("\n[bold cyan]📄 Configuration Files[/bold cyan]")
    config_table = Table()
    config_table.add_column("File", style="cyan")
    config_table.add_column("Status", style="green")
    config_table.add_column("Purpose", style="yellow")
    
    paths = get_project_paths()
    configs = [
        (paths["bumpversion_cfg"], ".bumpversion.cfg", "Version configuration"),
        (paths["go_mod"], "go.mod", "Go module definition"),
        (paths["makefile"], "Makefile", "Build automation"),
    ]
    
    for path, name, purpose in configs:
        if path.exists():
            status = "✅ Found"
        else:
            status = "❌ Missing"
        
        config_table.add_row(name, status, purpose)
    
    console.print(config_table)
    
    # Check GitHub integration
    console.print("\n[bold cyan]🐙 GitHub Integration[/bold cyan]")
    github_table = Table()
    github_table.add_column("Component", style="cyan")
    github_table.add_column("Status", style="green")
    github_table.add_column("Details", style="dim")
    
    # Check remote URL
    git_info = get_git_info()
    remote_url = git_info.get("Repository", "")
    if "github.com" in remote_url:
        github_status = "✅ GitHub remote"
        github_details = remote_url
    else:
        github_status = "⚠️  Non-GitHub remote"
        github_details = remote_url or "No remote configured"
    
    github_table.add_row("Remote Repository", github_status, github_details)
    
    # Check GitHub CLI
    if shutil.which("gh"):
        gh_status = "✅ Available"
        gh_details = "Can create releases automatically"
    else:
        gh_status = "⚠️  Not available"
        gh_details = "Manual release creation required"
    
    github_table.add_row("GitHub CLI", gh_status, gh_details)
    
    console.print(github_table)
