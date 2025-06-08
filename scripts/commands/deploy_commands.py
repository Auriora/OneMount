"""
Deployment and CI/CD commands for OneMount development CLI.
"""

import os
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
from utils.shell import run_command, run_command_with_progress, ensure_executable

console = Console()

# Create the deploy app
deploy_app = typer.Typer(
    name="deploy",
    help="Deployment and CI/CD operations",
    no_args_is_help=True,
)


@deploy_app.command()
def docker_remote(
    ctx: typer.Context,
    host: str = typer.Option(..., help="Remote Docker host"),
    port: str = typer.Option("2375", help="Docker API port"),
):
    """
    🐳 Deploy to remote Docker host.
    
    Deploy OneMount to a remote Docker host using the Docker API.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print(f"[blue]Deploying to remote Docker host {host}:{port}...[/blue]")
    
    paths = get_project_paths()
    script_path = paths["legacy_scripts"]["deploy_docker_remote"]
    
    if not script_path.exists():
        console.print(f"[red]Deploy script not found: {script_path}[/red]")
        raise typer.Exit(1)
    
    # Set environment variables for the deployment script
    env = os.environ.copy()
    env['DOCKER_HOST'] = f"tcp://{host}:{port}"
    
    ensure_executable(script_path)
    
    try:
        run_command_with_progress(
            [str(script_path)],
            f"Deploying to {host}:{port}",
            env=env,
            verbose=verbose,
            timeout=600,  # 10 minutes
        )
        
        console.print("[green]✅ Deployment completed successfully![/green]")
    
    except Exception as e:
        console.print(f"[red]Deployment failed: {e}[/red]")
        raise typer.Exit(1)


@deploy_app.command()
def setup_ci(ctx: typer.Context):
    """
    ⚙️  Setup personal CI environment.
    
    Configure the personal CI environment for OneMount development.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Setting up personal CI environment...[/blue]")
    
    paths = get_project_paths()
    script_path = paths["legacy_scripts"]["setup_personal_ci"]
    
    if not script_path.exists():
        console.print(f"[red]CI setup script not found: {script_path}[/red]")
        raise typer.Exit(1)
    
    ensure_executable(script_path)
    
    try:
        run_command_with_progress(
            [str(script_path)],
            "Setting up CI environment",
            verbose=verbose,
            timeout=300,  # 5 minutes
        )
        
        console.print("[green]✅ CI environment setup completed![/green]")
    
    except Exception as e:
        console.print(f"[red]CI setup failed: {e}[/red]")
        raise typer.Exit(1)


@deploy_app.command()
def status(ctx: typer.Context):
    """
    📋 Show deployment status and configuration.
    
    Display information about deployment configuration and available tools.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print(Panel.fit(
        "[bold blue]Deployment Status[/bold blue]",
        border_style="blue"
    ))
    
    # Docker configuration
    console.print("\n[bold cyan]🐳 Docker Configuration[/bold cyan]")
    docker_table = Table()
    docker_table.add_column("Property", style="cyan")
    docker_table.add_column("Value", style="green")
    docker_table.add_column("Status", style="yellow")
    
    # Check Docker availability
    import shutil
    if shutil.which("docker"):
        docker_status = "✅ Available"
        
        # Get Docker version
        try:
            result = run_command(
                ["docker", "--version"],
                capture_output=True,
                verbose=False,
            )
            docker_version = result.stdout.strip() if result.stdout else "Unknown"
        except Exception:
            docker_version = "Error getting version"
        
        # Check Docker daemon
        try:
            result = run_command(
                ["docker", "info"],
                capture_output=True,
                check=False,
                verbose=False,
            )
            if result.returncode == 0:
                daemon_status = "✅ Running"
            else:
                daemon_status = "❌ Not running"
        except Exception:
            daemon_status = "❓ Unknown"
    else:
        docker_status = "❌ Not available"
        docker_version = "Not installed"
        daemon_status = "N/A"
    
    docker_table.add_row("Docker CLI", docker_version, docker_status)
    docker_table.add_row("Docker Daemon", "Local daemon", daemon_status)
    
    # Check for remote Docker host
    docker_host = os.getenv("DOCKER_HOST")
    if docker_host:
        docker_table.add_row("Remote Host", docker_host, "✅ Configured")
    else:
        docker_table.add_row("Remote Host", "Not configured", "⚠️  Using local")
    
    console.print(docker_table)
    
    # CI/CD Tools
    console.print("\n[bold cyan]⚙️  CI/CD Tools[/bold cyan]")
    tools_table = Table()
    tools_table.add_column("Tool", style="cyan")
    tools_table.add_column("Status", style="green")
    tools_table.add_column("Purpose", style="dim")
    
    from utils.shell import get_command_version
    
    tools = [
        ("docker", "Container deployment"),
        ("docker-compose", "Multi-container applications"),
        ("kubectl", "Kubernetes deployment"),
        ("helm", "Kubernetes package management"),
        ("gh", "GitHub Actions integration"),
    ]
    
    for tool, purpose in tools:
        if shutil.which(tool):
            version = get_command_version(tool) or "Available"
            status = f"✅ {version}"
        else:
            status = "❌ Not installed"
        
        tools_table.add_row(tool, status, purpose)
    
    console.print(tools_table)
    
    # Environment Variables
    console.print("\n[bold cyan]🔧 Environment Configuration[/bold cyan]")
    env_table = Table()
    env_table.add_column("Variable", style="cyan")
    env_table.add_column("Status", style="green")
    env_table.add_column("Purpose", style="dim")
    
    env_vars = [
        ("DOCKER_HOST", "Remote Docker host"),
        ("DOCKER_TLS_VERIFY", "Docker TLS verification"),
        ("DOCKER_CERT_PATH", "Docker certificate path"),
        ("GITHUB_TOKEN", "GitHub API access"),
        ("CI", "CI environment indicator"),
    ]
    
    for var, purpose in env_vars:
        value = os.getenv(var)
        if value:
            # Mask sensitive values
            if "TOKEN" in var or "PASSWORD" in var:
                display_value = f"✅ Set ({value[:8]}...)"
            else:
                display_value = f"✅ {value}"
        else:
            display_value = "⚠️  Not set"
        
        env_table.add_row(var, display_value, purpose)
    
    console.print(env_table)
    
    # Deployment Scripts
    console.print("\n[bold cyan]📜 Deployment Scripts[/bold cyan]")
    scripts_table = Table()
    scripts_table.add_column("Script", style="cyan")
    scripts_table.add_column("Status", style="green")
    scripts_table.add_column("Purpose", style="dim")
    
    paths = get_project_paths()
    scripts = [
        ("deploy-docker-remote.sh", "Remote Docker deployment"),
        ("setup-personal-ci.sh", "CI environment setup"),
        ("manage-runner.sh", "Local runner management"),
        ("manage-runners.sh", "Simple 2-runner management"),
        ("deploy-remote-runner.sh", "Remote runner deployment"),
    ]
    
    for script_name, purpose in scripts:
        script_path = paths["scripts_dir"] / script_name
        if script_path.exists():
            status = "✅ Available"
        else:
            status = "❌ Missing"
        
        scripts_table.add_row(script_name, status, purpose)
    
    console.print(scripts_table)


@deploy_app.command()
def test_connection(
    ctx: typer.Context,
    host: str = typer.Option(..., help="Remote Docker host to test"),
    port: str = typer.Option("2375", help="Docker API port"),
):
    """
    🔍 Test connection to remote Docker host.
    
    Test connectivity and authentication to a remote Docker host.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print(f"[blue]Testing connection to {host}:{port}...[/blue]")
    
    # Check if Docker is available
    import shutil
    if not shutil.which("docker"):
        console.print("[red]Docker CLI is not available.[/red]")
        raise typer.Exit(1)
    
    # Set Docker host environment
    env = os.environ.copy()
    env['DOCKER_HOST'] = f"tcp://{host}:{port}"
    
    try:
        # Test basic connectivity
        console.print("[dim]Testing Docker daemon connectivity...[/dim]")
        result = run_command(
            ["docker", "version"],
            env=env,
            capture_output=True,
            verbose=verbose,
            timeout=10,
        )
        
        console.print("[green]✅ Successfully connected to Docker daemon[/green]")
        
        # Test Docker info
        console.print("[dim]Getting Docker daemon information...[/dim]")
        result = run_command(
            ["docker", "info", "--format", "{{.ServerVersion}}"],
            env=env,
            capture_output=True,
            verbose=verbose,
            timeout=10,
        )
        
        server_version = result.stdout.strip() if result.stdout else "Unknown"
        console.print(f"[green]✅ Docker daemon version: {server_version}[/green]")
        
        # Test image listing
        console.print("[dim]Testing image access...[/dim]")
        result = run_command(
            ["docker", "images", "--format", "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"],
            env=env,
            capture_output=True,
            verbose=verbose,
            timeout=15,
        )
        
        if result.stdout.strip():
            image_count = len(result.stdout.strip().split('\n')) - 1  # Subtract header
            console.print(f"[green]✅ Found {image_count} images on remote host[/green]")
        else:
            console.print("[yellow]⚠️  No images found on remote host[/yellow]")
        
        console.print(f"\n[green]🎉 Connection to {host}:{port} is working correctly![/green]")
    
    except Exception as e:
        console.print(f"[red]❌ Connection failed: {e}[/red]")
        console.print(f"\n[yellow]Troubleshooting tips:[/yellow]")
        console.print(f"1. Ensure Docker daemon is running on {host}")
        console.print(f"2. Check if port {port} is accessible")
        console.print(f"3. Verify firewall settings")
        console.print(f"4. For TLS, set DOCKER_TLS_VERIFY and DOCKER_CERT_PATH")
        raise typer.Exit(1)
