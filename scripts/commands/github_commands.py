"""
GitHub integration and issue management commands for OneMount development CLI.
"""

import sys
from pathlib import Path
from typing import Optional

import typer
from rich.console import Console

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from utils.environment import ensure_environment
from utils.paths import get_project_paths
from utils.shell import run_command, run_command_with_progress

console = Console()

# Create the github app
github_app = typer.Typer(
    name="github",
    help="GitHub integration and issue management",
    no_args_is_help=True,
)


@github_app.command()
def create_issues(
    ctx: typer.Context,
    repo: Optional[str] = typer.Option(None, help="GitHub repository (owner/repo)"),
    token: Optional[str] = typer.Option(None, help="GitHub personal access token"),
    file: str = typer.Option("data/github_issues_7MAY25.json", help="Path to JSON file containing issues"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be created without creating"),
):
    """
    🐙 Create GitHub issues from JSON file.
    
    Creates GitHub issues from a structured JSON file. Useful for bulk
    issue creation and project setup.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Creating GitHub issues...[/blue]")
    
    # Check if file exists
    file_path = Path(file)
    if not file_path.exists():
        console.print(f"[red]Issues file not found: {file_path}[/red]")
        raise typer.Exit(1)
    
    # Build command
    paths = get_project_paths()
    script_path = paths["scripts_dir"] / "create_github_issues.py"
    
    if not script_path.exists():
        console.print(f"[red]GitHub issues script not found: {script_path}[/red]")
        console.print("[yellow]This functionality may have been moved or removed.[/yellow]")
        raise typer.Exit(1)
    
    cmd = [sys.executable, str(script_path), "--file", str(file_path)]
    
    if repo:
        cmd.extend(["--repo", repo])
    if token:
        cmd.extend(["--token", token])
    if dry_run:
        cmd.append("--dry-run")
    
    try:
        run_command_with_progress(
            cmd,
            "Creating GitHub issues",
            verbose=verbose,
            timeout=300,  # 5 minutes
        )
        
        if dry_run:
            console.print("[yellow]Dry run completed - no issues created[/yellow]")
        else:
            console.print("[green]✅ GitHub issues created successfully![/green]")
    
    except Exception as e:
        console.print(f"[red]Failed to create GitHub issues: {e}[/red]")
        raise typer.Exit(1)


@github_app.command()
def implement(
    ctx: typer.Context,
    issue_number: int = typer.Argument(help="GitHub issue number to implement"),
):
    """
    🤖 Implement a GitHub issue with AI assistance.
    
    Uses AI assistance to implement the specified GitHub issue.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print(f"[blue]Implementing GitHub issue #{issue_number}...[/blue]")
    
    paths = get_project_paths()
    script_path = paths["scripts_dir"] / "implement_github_issue.py"
    
    if not script_path.exists():
        console.print(f"[red]GitHub implementation script not found: {script_path}[/red]")
        console.print("[yellow]This functionality may have been moved or removed.[/yellow]")
        raise typer.Exit(1)
    
    cmd = [sys.executable, str(script_path), str(issue_number)]
    
    try:
        run_command_with_progress(
            cmd,
            f"Implementing issue #{issue_number}",
            verbose=verbose,
            timeout=1800,  # 30 minutes
        )
        
        console.print(f"[green]✅ Issue #{issue_number} implementation completed![/green]")
    
    except Exception as e:
        console.print(f"[red]Failed to implement issue #{issue_number}: {e}[/red]")
        raise typer.Exit(1)


@github_app.command()
def analyze_issues(
    ctx: typer.Context,
    file: str = typer.Option("data/github_issues_7MAY25.json", help="Path to GitHub issues JSON file"),
):
    """
    📊 Analyze GitHub issues structure and content.
    
    Analyzes the structure and content of GitHub issues from a JSON file.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Analyzing GitHub issues...[/blue]")
    
    paths = get_project_paths()
    script_path = paths["scripts_dir"] / "analyze_issues.py"
    
    if not script_path.exists():
        console.print(f"[red]GitHub analysis script not found: {script_path}[/red]")
        console.print("[yellow]This functionality may have been moved or removed.[/yellow]")
        raise typer.Exit(1)
    
    cmd = [sys.executable, str(script_path)]
    
    try:
        run_command_with_progress(
            cmd,
            "Analyzing GitHub issues",
            verbose=verbose,
            timeout=120,  # 2 minutes
        )
        
        console.print("[green]✅ GitHub issues analysis completed![/green]")
    
    except Exception as e:
        console.print(f"[red]Failed to analyze GitHub issues: {e}[/red]")
        raise typer.Exit(1)


@github_app.command()
def status(ctx: typer.Context):
    """
    📋 Show GitHub integration status.
    
    Display information about GitHub repository, authentication, and available tools.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    from rich.panel import Panel
    from rich.table import Table
    from utils.git import get_git_info
    
    console.print(Panel.fit(
        "[bold blue]GitHub Integration Status[/bold blue]",
        border_style="blue"
    ))
    
    # Repository information
    console.print("\n[bold cyan]📁 Repository Information[/bold cyan]")
    repo_table = Table()
    repo_table.add_column("Property", style="cyan")
    repo_table.add_column("Value", style="green")
    
    git_info = get_git_info()
    
    repo_url = git_info.get("Repository", "Unknown")
    if "github.com" in repo_url:
        repo_status = "✅ GitHub repository"
    else:
        repo_status = "⚠️  Not a GitHub repository"
    
    repo_table.add_row("Repository URL", repo_url)
    repo_table.add_row("Repository Type", repo_status)
    repo_table.add_row("Current Branch", git_info.get("Current Branch", "Unknown"))
    repo_table.add_row("Latest Commit", git_info.get("Latest Commit (Short)", "Unknown"))
    
    console.print(repo_table)
    
    # GitHub tools
    console.print("\n[bold cyan]🛠️  GitHub Tools[/bold cyan]")
    tools_table = Table()
    tools_table.add_column("Tool", style="cyan")
    tools_table.add_column("Status", style="green")
    tools_table.add_column("Purpose", style="dim")
    
    import shutil
    from utils.shell import get_command_version
    
    tools = [
        ("gh", "GitHub CLI for repository management"),
        ("git", "Version control"),
        ("curl", "API requests"),
    ]
    
    for tool, purpose in tools:
        if shutil.which(tool):
            version = get_command_version(tool) or "Available"
            status = "✅ Available"
        else:
            version = "Not installed"
            status = "❌ Missing"
        
        tools_table.add_row(tool, f"{status} ({version})", purpose)
    
    console.print(tools_table)
    
    # Authentication status
    console.print("\n[bold cyan]🔐 Authentication[/bold cyan]")
    auth_table = Table()
    auth_table.add_column("Method", style="cyan")
    auth_table.add_column("Status", style="green")
    auth_table.add_column("Details", style="dim")
    
    # Check GitHub CLI authentication
    if shutil.which("gh"):
        try:
            result = run_command(
                ["gh", "auth", "status"],
                capture_output=True,
                check=False,
                verbose=False,
            )
            if result.returncode == 0:
                gh_auth_status = "✅ Authenticated"
                gh_auth_details = "GitHub CLI is authenticated"
            else:
                gh_auth_status = "❌ Not authenticated"
                gh_auth_details = "Run 'gh auth login'"
        except Exception:
            gh_auth_status = "❓ Unknown"
            gh_auth_details = "Could not check status"
    else:
        gh_auth_status = "⚠️  CLI not available"
        gh_auth_details = "Install GitHub CLI"
    
    auth_table.add_row("GitHub CLI", gh_auth_status, gh_auth_details)
    
    # Check for environment variables
    import os
    if os.getenv("GITHUB_TOKEN"):
        token_status = "✅ Token available"
        token_details = "GITHUB_TOKEN environment variable set"
    else:
        token_status = "⚠️  No token"
        token_details = "Set GITHUB_TOKEN for API access"
    
    auth_table.add_row("API Token", token_status, token_details)
    
    console.print(auth_table)


@github_app.command()
def workflows(ctx: typer.Context):
    """
    ⚙️  Show GitHub Actions workflow status.
    
    Display information about GitHub Actions workflows and their status.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    from rich.table import Table
    
    console.print("[blue]Checking GitHub Actions workflows...[/blue]")
    
    # Check if GitHub CLI is available
    import shutil
    if not shutil.which("gh"):
        console.print("[red]GitHub CLI (gh) is not available.[/red]")
        console.print("Install it to view workflow status: https://cli.github.com/")
        raise typer.Exit(1)
    
    try:
        # Get workflow runs
        result = run_command(
            ["gh", "run", "list", "--limit", "10"],
            capture_output=True,
            verbose=verbose,
        )
        
        if result.returncode != 0:
            console.print("[red]Failed to get workflow information.[/red]")
            console.print("Make sure you're authenticated with 'gh auth login'")
            raise typer.Exit(1)
        
        console.print("\n[bold cyan]🔄 Recent Workflow Runs[/bold cyan]")
        
        if result.stdout.strip():
            # Parse and display workflow runs
            lines = result.stdout.strip().split('\n')
            if len(lines) > 1:  # Skip header
                workflow_table = Table()
                workflow_table.add_column("Status", style="green")
                workflow_table.add_column("Workflow", style="cyan")
                workflow_table.add_column("Branch", style="yellow")
                workflow_table.add_column("Event", style="dim")
                workflow_table.add_column("Date", style="dim")
                
                for line in lines[1:]:  # Skip header
                    parts = line.split('\t')
                    if len(parts) >= 5:
                        status = parts[0]
                        workflow = parts[1]
                        branch = parts[2]
                        event = parts[3]
                        date = parts[4]
                        
                        # Format status with emoji
                        if "completed" in status.lower():
                            status_display = "✅ " + status
                        elif "in_progress" in status.lower():
                            status_display = "🔄 " + status
                        elif "failed" in status.lower():
                            status_display = "❌ " + status
                        else:
                            status_display = status
                        
                        workflow_table.add_row(
                            status_display,
                            workflow,
                            branch,
                            event,
                            date
                        )
                
                console.print(workflow_table)
            else:
                console.print("[yellow]No workflow runs found.[/yellow]")
        else:
            console.print("[yellow]No workflow runs found.[/yellow]")
    
    except Exception as e:
        console.print(f"[red]Failed to get workflow status: {e}[/red]")
        raise typer.Exit(1)


@github_app.command()
def get_test_results(
    ctx: typer.Context,
    output_dir: str = typer.Option("./test-results-download", "--output", "-o", help="Output directory for downloaded results"),
    workflow: Optional[str] = typer.Option(None, "--workflow", help="Filter by workflow name"),
    run_id: Optional[int] = typer.Option(None, "--run-id", help="Download artifacts from specific run ID"),
    token: Optional[str] = typer.Option(None, "--token", help="GitHub token (or set GITHUB_TOKEN env var)"),
):
    """
    📊 Download test results and reports from GitHub Actions.

    Downloads test results, coverage reports, and other artifacts from GitHub Actions workflows.
    Supports downloading from latest runs or specific run IDs.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    try:
        # Import here to avoid dependency issues if requests is not available
        from utils.github_test_retriever import GitHubTestRetriever

        console.print("[blue]Retrieving test results from GitHub Actions...[/blue]")

        retriever = GitHubTestRetriever(token=token)
        output_path = Path(output_dir)

        if run_id:
            # Download artifacts from specific run
            console.print(f"Downloading artifacts from run {run_id}")
            artifacts = retriever.get_run_artifacts(run_id)

            if not artifacts:
                console.print(f"[yellow]No artifacts found for run {run_id}[/yellow]")
                return

            downloaded = {}
            for artifact in artifacts:
                console.print(f"Downloading: {artifact['name']}")
                try:
                    extract_dir = retriever.download_artifact(artifact['id'], output_path / artifact['name'])
                    downloaded[artifact['name']] = extract_dir
                    console.print(f"✅ Extracted to: {extract_dir}")
                except Exception as e:
                    console.print(f"❌ Failed to download {artifact['name']}: {e}")

            if downloaded:
                retriever.show_test_summary(output_path)
        else:
            # Download latest test results
            downloaded = retriever.get_latest_test_results(output_path)

            if downloaded:
                retriever.show_test_summary(output_path)
                console.print(f"\n[bold green]✅ Test results downloaded to: {output_path}[/bold green]")

                # Show usage examples
                console.print("\n[bold cyan]📋 Usage Examples:[/bold cyan]")
                console.print("  • View JUnit XML reports in your IDE or CI tools")
                console.print("  • Open HTML coverage reports in your browser")
                console.print("  • Parse JSON reports programmatically")
                console.print("  • Analyze test trends over time")
            else:
                console.print("[yellow]No test artifacts found in recent workflow runs[/yellow]")
                console.print("Try running workflows first or specify a specific --run-id")

    except ImportError:
        console.print("[red]Missing dependencies for GitHub test retrieval.[/red]")
        console.print("Install required packages: pip install requests")
        raise typer.Exit(1)
    except Exception as e:
        console.print(f"[red]Error retrieving test results: {e}[/red]")
        if verbose:
            import traceback
            console.print(f"[dim]{traceback.format_exc()}[/dim]")
        raise typer.Exit(1)
