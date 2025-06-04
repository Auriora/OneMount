"""
Testing and quality assurance commands for OneMount development CLI.
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
from utils.paths import ensure_coverage_directories, get_project_paths
from utils.shell import run_command, run_command_with_progress, run_script, ensure_executable
from utils.coverage_reporter import generate_coverage_report
from utils.system_test_runner import run_system_tests
from utils.docker_test_runner import run_docker_tests, build_docker_image, clean_docker_resources, show_docker_auth_help

console = Console()

# Create the test app
test_app = typer.Typer(
    name="test",
    help="Testing and quality assurance operations",
    no_args_is_help=True,
)


@test_app.command()
def coverage(
    ctx: typer.Context,
    threshold_line: int = typer.Option(80, "--threshold-line", help="Line coverage threshold percentage"),
    threshold_func: int = typer.Option(90, "--threshold-func", help="Function coverage threshold percentage"),
    ci: bool = typer.Option(False, "--ci", help="Enable CI mode (machine-readable output)"),
):
    """
    📊 Generate comprehensive coverage reports.
    
    Generates HTML, JSON, and text coverage reports with threshold checking.
    Includes package-by-package analysis and coverage history tracking.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    ensure_coverage_directories()

    console.print("[blue]Generating coverage reports...[/blue]")

    # Use native Python coverage reporter
    success = generate_coverage_report(
        verbose=verbose,
        ci_mode=ci,
        threshold_line=threshold_line,
        threshold_func=threshold_func,
        threshold_branch=70  # Default branch threshold
    )

    if not success:
        console.print("[red]Coverage generation failed or thresholds not met[/red]")
        raise typer.Exit(1)

    console.print("[green]✅ Coverage reports generated successfully![/green]")

    # Show coverage summary
    paths = get_project_paths()
    coverage_dir = paths["project_root"] / "coverage"
    if coverage_dir.exists():
        console.print(f"\n[cyan]📊 Reports available in {coverage_dir}:[/cyan]")

        reports = [
            ("HTML Report", "coverage.html"),
            ("Function Analysis", "coverage-func.txt"),
            ("Package Analysis", "package-analysis.txt"),
            ("Coverage Gaps", "coverage-gaps.txt"),
            ("JSON Report", "coverage.json"),
            ("Summary", "summary.txt"),
        ]

        for name, filename in reports:
            report_path = coverage_dir / filename
            if report_path.exists():
                console.print(f"  • {name}: {report_path}")


@test_app.command()
def system(
    ctx: typer.Context,
    category: str = typer.Option(
        "comprehensive",
        help="Test category to run (comprehensive/performance/reliability/integration/stress/all)"
    ),
    timeout: str = typer.Option("30m", help="Test timeout duration"),
):
    """
    🧪 Run system tests with real OneDrive integration.

    Runs comprehensive system tests using a real OneDrive account.
    Requires authentication tokens to be set up.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    valid_categories = ["comprehensive", "performance", "reliability", "integration", "stress", "all"]
    if category not in valid_categories:
        console.print(f"[red]Invalid category: {category}. Must be one of: {', '.join(valid_categories)}[/red]")
        raise typer.Exit(1)

    console.print(f"[blue]Running {category} system tests...[/blue]")

    # Use native Python system test runner
    success = run_system_tests(
        category=category,
        timeout=timeout,
        verbose=verbose
    )

    if not success:
        console.print("[red]System tests failed[/red]")
        raise typer.Exit(1)

    console.print("[green]✅ System tests completed successfully![/green]")


# Docker test commands
docker_app = typer.Typer(help="🐳 Docker test orchestration commands")
test_app.add_typer(docker_app, name="docker")


@docker_app.command()
def unit(
    ctx: typer.Context,
    rebuild: bool = typer.Option(False, "--rebuild", help="Force rebuild of Docker image"),
    timeout: Optional[str] = typer.Option(None, "--timeout", help="Test timeout duration"),
    sequential: bool = typer.Option(False, "--sequential", help="Run tests sequentially"),
    clean: bool = typer.Option(False, "--clean", help="Clean up Docker resources after tests"),
):
    """
    🧪 Run unit tests in Docker container.

    Runs unit tests in an isolated Docker environment with all dependencies.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running unit tests in Docker...[/blue]")

    success = run_docker_tests(
        test_type="unit",
        rebuild=rebuild,
        timeout=timeout,
        verbose=verbose,
        sequential=sequential,
        clean=clean
    )

    if not success:
        raise typer.Exit(1)


@docker_app.command()
def integration(
    ctx: typer.Context,
    rebuild: bool = typer.Option(False, "--rebuild", help="Force rebuild of Docker image"),
    timeout: Optional[str] = typer.Option(None, "--timeout", help="Test timeout duration"),
    sequential: bool = typer.Option(False, "--sequential", help="Run tests sequentially"),
    clean: bool = typer.Option(False, "--clean", help="Clean up Docker resources after tests"),
):
    """
    🧪 Run integration tests in Docker container.

    Runs integration tests in an isolated Docker environment.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running integration tests in Docker...[/blue]")

    success = run_docker_tests(
        test_type="integration",
        rebuild=rebuild,
        timeout=timeout,
        verbose=verbose,
        sequential=sequential,
        clean=clean
    )

    if not success:
        raise typer.Exit(1)


@docker_app.command()
def system(
    ctx: typer.Context,
    rebuild: bool = typer.Option(False, "--rebuild", help="Force rebuild of Docker image"),
    timeout: Optional[str] = typer.Option(None, "--timeout", help="Test timeout duration"),
    sequential: bool = typer.Option(False, "--sequential", help="Run tests sequentially"),
    clean: bool = typer.Option(False, "--clean", help="Clean up Docker resources after tests"),
):
    """
    🧪 Run system tests in Docker container.

    Runs system tests with real OneDrive integration in Docker.
    Requires authentication tokens to be set up.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running system tests in Docker...[/blue]")

    success = run_docker_tests(
        test_type="system",
        rebuild=rebuild,
        timeout=timeout,
        verbose=verbose,
        sequential=sequential,
        clean=clean
    )

    if not success:
        raise typer.Exit(1)


@docker_app.command()
def all(
    ctx: typer.Context,
    rebuild: bool = typer.Option(False, "--rebuild", help="Force rebuild of Docker image"),
    timeout: Optional[str] = typer.Option(None, "--timeout", help="Test timeout duration"),
    sequential: bool = typer.Option(False, "--sequential", help="Run tests sequentially"),
    clean: bool = typer.Option(False, "--clean", help="Clean up Docker resources after tests"),
):
    """
    🧪 Run all tests in Docker container.

    Runs unit, integration, and system tests in Docker.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running all tests in Docker...[/blue]")

    success = run_docker_tests(
        test_type="all",
        rebuild=rebuild,
        timeout=timeout,
        verbose=verbose,
        sequential=sequential,
        clean=clean
    )

    if not success:
        raise typer.Exit(1)


@docker_app.command()
def coverage(
    ctx: typer.Context,
    rebuild: bool = typer.Option(False, "--rebuild", help="Force rebuild of Docker image"),
    timeout: Optional[str] = typer.Option(None, "--timeout", help="Test timeout duration"),
    sequential: bool = typer.Option(False, "--sequential", help="Run tests sequentially"),
    clean: bool = typer.Option(False, "--clean", help="Clean up Docker resources after tests"),
):
    """
    📊 Run tests with coverage analysis in Docker.

    Runs tests with coverage analysis in Docker and generates reports.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running coverage analysis in Docker...[/blue]")

    success = run_docker_tests(
        test_type="coverage",
        rebuild=rebuild,
        timeout=timeout,
        verbose=verbose,
        sequential=sequential,
        clean=clean
    )

    if not success:
        raise typer.Exit(1)


@docker_app.command()
def build(
    ctx: typer.Context,
    no_cache: bool = typer.Option(False, "--no-cache", help="Force rebuild without cache"),
):
    """
    🔨 Build Docker test image.

    Builds the OneMount test Docker image.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Building Docker test image...[/blue]")

    success = build_docker_image(no_cache=no_cache, verbose=verbose)

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ Docker image built successfully![/green]")


@docker_app.command()
def clean(
    ctx: typer.Context,
):
    """
    🧹 Clean up Docker test resources.

    Removes Docker containers, images, and test artifacts.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Cleaning up Docker test resources...[/blue]")

    success = clean_docker_resources(verbose=verbose)

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ Docker cleanup completed![/green]")


@docker_app.command()
def setup_auth():
    """
    🔐 Show authentication setup help for Docker system tests.

    Displays instructions for setting up OneDrive authentication tokens.
    """
    show_docker_auth_help()



@test_app.command()
def unit(
    ctx: typer.Context,
    package: Optional[str] = typer.Option(None, help="Specific package to test (e.g., ./internal/fs/...)"),
    verbose_go: bool = typer.Option(False, "--verbose-go", help="Enable verbose Go test output"),
    race: bool = typer.Option(False, "--race", help="Enable race condition detection"),
    timeout: str = typer.Option("5m", help="Test timeout duration"),
):
    """
    🔬 Run unit tests.
    
    Run Go unit tests with optional race detection and verbose output.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Running unit tests...[/blue]")
    
    # Build Go test command
    cmd = ["go", "test"]
    
    if verbose_go:
        cmd.append("-v")
    
    if race:
        cmd.append("-race")
    
    cmd.extend(["-timeout", timeout])
    
    if package:
        cmd.append(package)
    else:
        cmd.append("./...")
    
    try:
        run_command_with_progress(
            cmd,
            "Running unit tests",
            verbose=verbose,
            timeout=None,  # Use Go's own timeout
        )
        
        console.print("[green]✅ Unit tests passed![/green]")
    
    except Exception as e:
        console.print(f"[red]Unit tests failed: {e}[/red]")
        raise typer.Exit(1)


@test_app.command()
def integration(
    ctx: typer.Context,
    timeout: str = typer.Option("10m", help="Test timeout duration"),
):
    """
    🔗 Run integration tests.
    
    Run integration tests that test component interactions.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Running integration tests...[/blue]")
    
    # Run integration tests
    cmd = [
        "go", "test",
        "-v",
        "-timeout", timeout,
        "-tags", "integration",
        "./tests/integration/...",
    ]
    
    try:
        run_command_with_progress(
            cmd,
            "Running integration tests",
            verbose=verbose,
            timeout=None,  # Use Go's own timeout
        )
        
        console.print("[green]✅ Integration tests passed![/green]")
    
    except Exception as e:
        console.print(f"[red]Integration tests failed: {e}[/red]")
        raise typer.Exit(1)


@test_app.command()
def all(
    ctx: typer.Context,
    timeout: str = typer.Option("15m", help="Test timeout duration"),
    race: bool = typer.Option(False, "--race", help="Enable race condition detection"),
):
    """
    🎯 Run all tests (unit + integration).
    
    Run the complete test suite including unit and integration tests.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Running all tests...[/blue]")
    
    # Build Go test command
    cmd = ["go", "test", "-v", "-timeout", timeout]
    
    if race:
        cmd.append("-race")
    
    cmd.append("./...")
    
    try:
        run_command_with_progress(
            cmd,
            "Running all tests",
            verbose=verbose,
            timeout=None,  # Use Go's own timeout
        )
        
        console.print("[green]✅ All tests passed![/green]")
    
    except Exception as e:
        console.print(f"[red]Tests failed: {e}[/red]")
        raise typer.Exit(1)


@test_app.command()
def status(ctx: typer.Context):
    """
    📋 Show testing status and information.
    
    Display information about test coverage, recent test runs,
    and testing environment setup.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print(Panel.fit(
        "[bold blue]OneMount Testing Status[/bold blue]",
        border_style="blue"
    ))
    
    paths = get_project_paths()
    
    # Coverage status
    console.print("\n[bold cyan]📊 Coverage Status[/bold cyan]")
    coverage_table = Table()
    coverage_table.add_column("Report", style="cyan")
    coverage_table.add_column("Status", style="green")
    coverage_table.add_column("Location", style="dim")
    
    coverage_files = [
        ("Coverage Data", "coverage.out"),
        ("HTML Report", "coverage.html"),
        ("JSON Report", "coverage.json"),
        ("Coverage History", "coverage_history.json"),
    ]
    
    coverage_dir = paths["coverage_dir"]
    for name, filename in coverage_files:
        file_path = coverage_dir / filename
        if file_path.exists():
            import datetime
            modified = datetime.datetime.fromtimestamp(file_path.stat().st_mtime)
            status = f"✅ {modified.strftime('%Y-%m-%d %H:%M')}"
        else:
            status = "❌ Not found"
        
        coverage_table.add_row(name, status, str(file_path))
    
    console.print(coverage_table)
    
    # Test environment
    console.print("\n[bold cyan]🧪 Test Environment[/bold cyan]")
    env_table = Table()
    env_table.add_column("Component", style="cyan")
    env_table.add_column("Status", style="green")
    env_table.add_column("Details", style="dim")
    
    # Check for auth tokens
    auth_tokens_path = Path.home() / ".onemount-tests" / ".auth_tokens.json"
    if auth_tokens_path.exists():
        auth_status = "✅ Available"
        auth_details = str(auth_tokens_path)
    else:
        auth_status = "❌ Missing"
        auth_details = "Required for system tests"
    
    env_table.add_row("Auth Tokens", auth_status, auth_details)
    
    # Check for Docker
    import shutil
    if shutil.which("docker"):
        docker_status = "✅ Available"
        docker_details = "Docker tests enabled"
    else:
        docker_status = "❌ Missing"
        docker_details = "Docker tests disabled"
    
    env_table.add_row("Docker", docker_status, docker_details)
    
    # Check test directories
    test_dirs = [
        ("Unit Tests", paths["project_root"]),
        ("Integration Tests", paths["tests_dir"] / "integration"),
        ("System Tests", paths["tests_dir"] / "system"),
    ]
    
    for name, test_dir in test_dirs:
        if test_dir.exists():
            # Count test files
            test_files = list(test_dir.rglob("*_test.go"))
            status = "✅ Available"
            details = f"{len(test_files)} test files"
        else:
            status = "❌ Missing"
            details = "Directory not found"
        
        env_table.add_row(name, status, details)
    
    console.print(env_table)
