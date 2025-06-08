"""
Testing and quality assurance commands for OneMount development CLI.
"""

import sys
import subprocess
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
    json_output: Optional[str] = typer.Option(None, "--json-output", help="Path to save JSON test output for CI reporting"),
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
        verbose=verbose,
        json_output=json_output
    )

    if not success:
        console.print("[red]System tests failed[/red]")
        raise typer.Exit(1)

    console.print("[green]✅ System tests completed successfully![/green]")


# Docker test commands
docker_app = typer.Typer(help="🐳 Docker test orchestration commands")
test_app.add_typer(docker_app, name="docker")

# Nemo extension test commands
nemo_app = typer.Typer(help="🗂️ Nemo extension test commands")
test_app.add_typer(nemo_app, name="nemo")


@docker_app.command()
def unit(
    ctx: typer.Context,
    rebuild_image: bool = typer.Option(False, "--rebuild-image", help="Force rebuild of Docker image"),
    recreate_container: bool = typer.Option(False, "--recreate-container", help="Force recreation of container"),
    no_reuse: bool = typer.Option(False, "--no-reuse", help="Don't reuse existing containers"),
    timeout: Optional[str] = typer.Option(None, "--timeout", help="Test timeout duration"),
    sequential: bool = typer.Option(False, "--sequential", help="Run tests sequentially"),
    clean: bool = typer.Option(False, "--clean", help="Clean up Docker resources after tests"),
    development: bool = typer.Option(False, "--dev", help="Use development mode with persistent containers"),
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
        rebuild_image=rebuild_image,
        recreate_container=recreate_container,
        reuse_container=not no_reuse,
        timeout=timeout,
        verbose=verbose,
        sequential=sequential,
        clean=clean,
        development=development
    )

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ Unit tests completed successfully![/green]")


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
        rebuild_image=rebuild,
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
        rebuild_image=rebuild,
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
        rebuild_image=rebuild,
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
        rebuild_image=rebuild,
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
    tag: Optional[str] = typer.Option(None, "--tag", help="Custom tag for the image"),
    development: bool = typer.Option(False, "--dev", help="Build development image"),
    use_compose: bool = typer.Option(False, "--compose", help="Use Docker Compose for building (experimental)"),
):
    """
    🔨 Build Docker test image.

    Builds the OneMount test Docker image using Docker Compose by default.
    The image is tagged and ready for use without rebuilding on each test run.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    image_type = "development" if development else "production"
    console.print(f"[blue]Building Docker test image ({image_type})...[/blue]")

    success = build_docker_image(
        no_cache=no_cache,
        tag=tag,
        development=development,
        use_compose=use_compose,
        verbose=verbose
    )

    if not success:
        raise typer.Exit(1)

    final_tag = tag or ("onemount-test-runner:dev" if development else "onemount-test-runner:latest")
    console.print(f"[green]✅ Docker image built successfully: {final_tag}[/green]")

    if development:
        console.print("[yellow]💡 Use this image for fast development with container reuse[/yellow]")
    else:
        console.print("[yellow]💡 Image is ready for testing. Use 'make docker-test-unit' for fast execution[/yellow]")


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


# Nemo extension test commands implementation
def _run_nemo_tests(
    test_type: str = "all",
    verbose: bool = False,
    coverage: bool = False,
    test_file: Optional[str] = None,
    test_function: Optional[str] = None,
    pytest_args: Optional[str] = None,
) -> bool:
    """
    Run Nemo extension tests using the Python test runner.

    Args:
        test_type: Type of tests to run (unit/integration/dbus/mock/all)
        verbose: Enable verbose output
        coverage: Generate coverage reports
        test_file: Specific test file to run
        test_function: Specific test function to run
        pytest_args: Additional pytest arguments

    Returns:
        bool: True if tests passed, False otherwise
    """
    paths = get_project_paths()
    nemo_dir = paths["project_root"] / "internal" / "nemo"

    if not nemo_dir.exists():
        console.print(f"[red]Nemo extension directory not found: {nemo_dir}[/red]")
        return False

    # Check if test runner exists
    test_runner = nemo_dir / "run_tests.py"
    if not test_runner.exists():
        console.print(f"[red]Nemo test runner not found: {test_runner}[/red]")
        console.print("[yellow]Run 'scripts/dev.py test nemo setup' to initialize the test suite[/yellow]")
        return False

    # Build command
    cmd = ["python3", str(test_runner)]

    # Add test type filter
    if test_type == "unit":
        cmd.append("--unit-only")
    elif test_type == "integration":
        cmd.append("--integration-only")
    elif test_type == "dbus":
        cmd.append("--dbus-only")
    elif test_type == "mock":
        cmd.append("--mock-only")
    # "all" runs everything by default

    # Add options
    if verbose:
        cmd.append("--verbose")

    if coverage:
        cmd.append("--coverage")

    if test_file:
        cmd.extend(["--test-file", test_file])

    if test_function:
        cmd.extend(["--test-function", test_function])

    if pytest_args:
        cmd.extend(["--pytest-args", pytest_args])

    try:
        # Change to nemo directory for execution
        result = subprocess.run(
            cmd,
            cwd=nemo_dir,
            capture_output=False,
            text=True,
            check=True
        )
        return True

    except subprocess.CalledProcessError as e:
        console.print(f"[red]Nemo tests failed with exit code {e.returncode}[/red]")
        return False
    except Exception as e:
        console.print(f"[red]Error running Nemo tests: {e}[/red]")
        return False


def _check_nemo_dependencies() -> bool:
    """Check if Nemo test dependencies are available."""
    paths = get_project_paths()
    nemo_dir = paths["project_root"] / "internal" / "nemo"

    if not nemo_dir.exists():
        return False

    test_runner = nemo_dir / "run_tests.py"
    if not test_runner.exists():
        return False

    # Check dependencies using the test runner
    try:
        result = subprocess.run(
            ["python3", str(test_runner), "--check-deps"],
            cwd=nemo_dir,
            capture_output=True,
            text=True,
            check=True
        )
        return True
    except subprocess.CalledProcessError:
        return False
    except Exception:
        return False


@nemo_app.command()
def unit(
    ctx: typer.Context,
    verbose_pytest: bool = typer.Option(False, "--verbose-pytest", help="Enable verbose pytest output"),
    coverage: bool = typer.Option(False, "--coverage", help="Generate coverage reports"),
    test_file: Optional[str] = typer.Option(None, "--test-file", help="Specific test file to run"),
    test_function: Optional[str] = typer.Option(None, "--test-function", help="Specific test function to run"),
):
    """
    🔬 Run Nemo extension unit tests.

    Run unit tests for the OneMount Nemo file manager extension.
    Tests individual components in isolation with mocked dependencies.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running Nemo extension unit tests...[/blue]")

    success = _run_nemo_tests(
        test_type="unit",
        verbose=verbose or verbose_pytest,
        coverage=coverage,
        test_file=test_file,
        test_function=test_function
    )

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ Nemo unit tests passed![/green]")


@nemo_app.command()
def integration(
    ctx: typer.Context,
    verbose_pytest: bool = typer.Option(False, "--verbose-pytest", help="Enable verbose pytest output"),
    coverage: bool = typer.Option(False, "--coverage", help="Generate coverage reports"),
):
    """
    🔗 Run Nemo extension integration tests.

    Run integration tests for D-Bus communication between the Go service
    and the Python Nemo extension.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running Nemo extension integration tests...[/blue]")

    success = _run_nemo_tests(
        test_type="integration",
        verbose=verbose or verbose_pytest,
        coverage=coverage
    )

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ Nemo integration tests passed![/green]")


@nemo_app.command()
def dbus(
    ctx: typer.Context,
    verbose_pytest: bool = typer.Option(False, "--verbose-pytest", help="Enable verbose pytest output"),
    coverage: bool = typer.Option(False, "--coverage", help="Generate coverage reports"),
):
    """
    🔌 Run Nemo extension D-Bus tests.

    Run tests specifically focused on D-Bus functionality including
    service connection, method calls, and signal handling.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running Nemo extension D-Bus tests...[/blue]")

    success = _run_nemo_tests(
        test_type="dbus",
        verbose=verbose or verbose_pytest,
        coverage=coverage
    )

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ Nemo D-Bus tests passed![/green]")


@nemo_app.command()
def mock(
    ctx: typer.Context,
    verbose_pytest: bool = typer.Option(False, "--verbose-pytest", help="Enable verbose pytest output"),
    coverage: bool = typer.Option(False, "--coverage", help="Generate coverage reports"),
):
    """
    🎭 Run Nemo extension mock tests.

    Run tests with mocked dependencies to verify behavior in offline
    scenarios, service unavailability, and error conditions.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running Nemo extension mock tests...[/blue]")

    success = _run_nemo_tests(
        test_type="mock",
        verbose=verbose or verbose_pytest,
        coverage=coverage
    )

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ Nemo mock tests passed![/green]")


@nemo_app.command()
def all(
    ctx: typer.Context,
    verbose_pytest: bool = typer.Option(False, "--verbose-pytest", help="Enable verbose pytest output"),
    coverage: bool = typer.Option(False, "--coverage", help="Generate coverage reports"),
):
    """
    🎯 Run all Nemo extension tests.

    Run the complete Nemo extension test suite including unit,
    integration, D-Bus, and mock tests.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running all Nemo extension tests...[/blue]")

    success = _run_nemo_tests(
        test_type="all",
        verbose=verbose or verbose_pytest,
        coverage=coverage
    )

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ All Nemo tests passed![/green]")


@nemo_app.command()
def status(ctx: typer.Context):
    """
    📋 Show Nemo extension test status.

    Display information about the Nemo extension test suite including
    test coverage, dependencies, and recent test runs.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    console.print(Panel.fit(
        "[bold blue]Nemo Extension Test Status[/bold blue]",
        border_style="blue"
    ))

    paths = get_project_paths()
    nemo_dir = paths["project_root"] / "internal" / "nemo"

    # Test suite status
    console.print("\n[bold cyan]🗂️ Test Suite Status[/bold cyan]")
    status_table = Table()
    status_table.add_column("Component", style="cyan")
    status_table.add_column("Status", style="green")
    status_table.add_column("Details", style="dim")

    # Check if Nemo directory exists
    if nemo_dir.exists():
        nemo_status = "✅ Available"
        nemo_details = str(nemo_dir)
    else:
        nemo_status = "❌ Missing"
        nemo_details = "Nemo extension directory not found"

    status_table.add_row("Nemo Extension", nemo_status, nemo_details)

    # Check test runner
    test_runner = nemo_dir / "run_tests.py"
    if test_runner.exists():
        runner_status = "✅ Available"
        runner_details = "Python test runner ready"
    else:
        runner_status = "❌ Missing"
        runner_details = "Test runner not found"

    status_table.add_row("Test Runner", runner_status, runner_details)

    # Check dependencies
    if _check_nemo_dependencies():
        deps_status = "✅ Available"
        deps_details = "All dependencies installed"
    else:
        deps_status = "❌ Missing"
        deps_details = "Run 'scripts/dev.py test nemo setup' to install"

    status_table.add_row("Dependencies", deps_status, deps_details)

    # Check test files
    if nemo_dir.exists():
        tests_dir = nemo_dir / "tests"
        if tests_dir.exists():
            test_files = list(tests_dir.glob("test_*.py"))
            tests_status = "✅ Available"
            tests_details = f"{len(test_files)} test files found"
        else:
            tests_status = "❌ Missing"
            tests_details = "Tests directory not found"
    else:
        tests_status = "❌ Missing"
        tests_details = "Nemo directory not found"

    status_table.add_row("Test Files", tests_status, tests_details)

    console.print(status_table)

    # Coverage status
    if nemo_dir.exists():
        console.print("\n[bold cyan]📊 Coverage Status[/bold cyan]")
        coverage_table = Table()
        coverage_table.add_column("Report", style="cyan")
        coverage_table.add_column("Status", style="green")
        coverage_table.add_column("Location", style="dim")

        coverage_files = [
            ("HTML Report", "htmlcov/index.html"),
            ("XML Report", "coverage.xml"),
            ("Coverage Data", ".coverage"),
        ]

        for name, filename in coverage_files:
            file_path = nemo_dir / filename
            if file_path.exists():
                import datetime
                modified = datetime.datetime.fromtimestamp(file_path.stat().st_mtime)
                status = f"✅ {modified.strftime('%Y-%m-%d %H:%M')}"
            else:
                status = "❌ Not found"

            coverage_table.add_row(name, status, str(file_path))

        console.print(coverage_table)


@nemo_app.command()
def setup(ctx: typer.Context):
    """
    🔧 Set up Nemo extension test environment.

    Initialize the Nemo extension test suite and install dependencies.
    This command checks for required dependencies and provides setup instructions.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    console.print("[blue]Setting up Nemo extension test environment...[/blue]")

    paths = get_project_paths()
    nemo_dir = paths["project_root"] / "internal" / "nemo"

    if not nemo_dir.exists():
        console.print(f"[red]Nemo extension directory not found: {nemo_dir}[/red]")
        console.print("[yellow]The Nemo extension may not be implemented yet.[/yellow]")
        raise typer.Exit(1)

    # Check if test runner exists
    test_runner = nemo_dir / "run_tests.py"
    if not test_runner.exists():
        console.print(f"[red]Test runner not found: {test_runner}[/red]")
        console.print("[yellow]The Nemo test suite may not be implemented yet.[/yellow]")
        raise typer.Exit(1)

    # Check dependencies
    console.print("\n[cyan]Checking dependencies...[/cyan]")

    try:
        result = subprocess.run(
            ["python3", str(test_runner), "--check-deps"],
            cwd=nemo_dir,
            capture_output=True,
            text=True,
            check=True
        )
        console.print("[green]✅ All dependencies are available![/green]")
        console.print(result.stdout)

    except subprocess.CalledProcessError as e:
        console.print("[yellow]⚠️ Some dependencies are missing:[/yellow]")
        console.print(e.stdout)
        console.print(e.stderr)

        console.print("\n[cyan]To install missing dependencies:[/cyan]")
        console.print(f"  cd {nemo_dir}")
        console.print("  pip install -r requirements.txt")

        # Try to install automatically
        install = typer.confirm("Would you like to install missing dependencies now?")
        if install:
            try:
                console.print("[blue]Installing dependencies...[/blue]")
                subprocess.run(
                    ["pip", "install", "-r", "requirements.txt"],
                    cwd=nemo_dir,
                    check=True
                )
                console.print("[green]✅ Dependencies installed successfully![/green]")
            except subprocess.CalledProcessError as e:
                console.print(f"[red]Failed to install dependencies: {e}[/red]")
                raise typer.Exit(1)
        else:
            console.print("[yellow]Please install dependencies manually before running tests.[/yellow]")
            raise typer.Exit(1)

    except Exception as e:
        console.print(f"[red]Error checking dependencies: {e}[/red]")
        raise typer.Exit(1)

    # Show usage information
    console.print("\n[bold cyan]🎉 Nemo Extension Test Suite Ready![/bold cyan]")
    console.print("\n[cyan]Available commands:[/cyan]")
    console.print("  scripts/dev.py test nemo unit        # Run unit tests")
    console.print("  scripts/dev.py test nemo integration # Run integration tests")
    console.print("  scripts/dev.py test nemo dbus        # Run D-Bus tests")
    console.print("  scripts/dev.py test nemo mock        # Run mock tests")
    console.print("  scripts/dev.py test nemo all         # Run all tests")
    console.print("  scripts/dev.py test nemo status      # Show test status")

    console.print("\n[cyan]With coverage:[/cyan]")
    console.print("  scripts/dev.py test nemo all --coverage")

    console.print("\n[cyan]Specific tests:[/cyan]")
    console.print("  scripts/dev.py test nemo unit --test-file test_simple.py")
    console.print("  scripts/dev.py test nemo unit --test-file test_simple.py --test-function test_mount_point_parsing")


@nemo_app.command()
def coverage(
    ctx: typer.Context,
    html: bool = typer.Option(True, "--html/--no-html", help="Generate HTML coverage report"),
    xml: bool = typer.Option(False, "--xml", help="Generate XML coverage report"),
    term: bool = typer.Option(True, "--term/--no-term", help="Show terminal coverage report"),
):
    """
    📊 Generate Nemo extension coverage reports.

    Run all Nemo extension tests with coverage analysis and generate
    comprehensive coverage reports.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Generating Nemo extension coverage reports...[/blue]")

    # Build coverage arguments
    coverage_args = []
    if html:
        coverage_args.append("--cov-report=html")
    if xml:
        coverage_args.append("--cov-report=xml")
    if term:
        coverage_args.append("--cov-report=term-missing")

    pytest_args = " ".join(coverage_args) if coverage_args else None

    success = _run_nemo_tests(
        test_type="all",
        verbose=verbose,
        coverage=True,
        pytest_args=pytest_args
    )

    if not success:
        raise typer.Exit(1)

    console.print("[green]✅ Nemo coverage reports generated![/green]")

    # Show coverage report locations
    paths = get_project_paths()
    nemo_dir = paths["project_root"] / "internal" / "nemo"

    console.print(f"\n[cyan]📊 Coverage reports available:[/cyan]")

    if html:
        html_report = nemo_dir / "htmlcov" / "index.html"
        if html_report.exists():
            console.print(f"  • HTML Report: {html_report}")

    if xml:
        xml_report = nemo_dir / "coverage.xml"
        if xml_report.exists():
            console.print(f"  • XML Report: {xml_report}")

    coverage_data = nemo_dir / ".coverage"
    if coverage_data.exists():
        console.print(f"  • Coverage Data: {coverage_data}")


@nemo_app.command()
def go_dbus(
    ctx: typer.Context,
    verbose_go: bool = typer.Option(False, "--verbose-go", help="Enable verbose Go test output"),
    timeout: str = typer.Option("5m", help="Test timeout duration"),
):
    """
    🔌 Run Go D-Bus server tests.

    Run the Go D-Bus server tests that complement the Python Nemo extension tests.
    These tests verify the D-Bus service functionality from the Go side.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running Go D-Bus server tests...[/blue]")

    paths = get_project_paths()
    fs_dir = paths["project_root"] / "internal" / "fs"

    if not fs_dir.exists():
        console.print(f"[red]Filesystem module directory not found: {fs_dir}[/red]")
        raise typer.Exit(1)

    # Build Go test command for D-Bus tests
    cmd = ["go", "test"]

    if verbose or verbose_go:
        cmd.append("-v")

    cmd.extend(["-timeout", timeout])
    cmd.extend(["-run", "DBus"])  # Run only D-Bus related tests
    cmd.append("./...")

    try:
        result = subprocess.run(
            cmd,
            cwd=fs_dir,
            capture_output=False,
            text=True,
            check=True
        )

        console.print("[green]✅ Go D-Bus tests passed![/green]")

    except subprocess.CalledProcessError as e:
        console.print(f"[red]Go D-Bus tests failed with exit code {e.returncode}[/red]")
        raise typer.Exit(1)
    except Exception as e:
        console.print(f"[red]Error running Go D-Bus tests: {e}[/red]")
        raise typer.Exit(1)


@nemo_app.command()
def full(
    ctx: typer.Context,
    verbose_pytest: bool = typer.Option(False, "--verbose-pytest", help="Enable verbose pytest output"),
    verbose_go: bool = typer.Option(False, "--verbose-go", help="Enable verbose Go test output"),
    coverage: bool = typer.Option(False, "--coverage", help="Generate coverage reports"),
    timeout: str = typer.Option("10m", help="Test timeout duration"),
):
    """
    🎯 Run complete Nemo extension test suite.

    Run both Python Nemo extension tests and Go D-Bus server tests
    for comprehensive coverage of the entire Nemo integration.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False

    if not ensure_environment():
        raise typer.Exit(1)

    console.print("[blue]Running complete Nemo extension test suite...[/blue]")

    # Run Python tests first
    console.print("\n[cyan]1. Running Python Nemo extension tests...[/cyan]")
    python_success = _run_nemo_tests(
        test_type="all",
        verbose=verbose or verbose_pytest,
        coverage=coverage
    )

    if not python_success:
        console.print("[red]Python Nemo tests failed[/red]")
        raise typer.Exit(1)

    console.print("[green]✅ Python Nemo tests passed![/green]")

    # Run Go D-Bus tests
    console.print("\n[cyan]2. Running Go D-Bus server tests...[/cyan]")

    paths = get_project_paths()
    fs_dir = paths["project_root"] / "internal" / "fs"

    if fs_dir.exists():
        cmd = ["go", "test"]

        if verbose or verbose_go:
            cmd.append("-v")

        cmd.extend(["-timeout", timeout])
        cmd.extend(["-run", "DBus"])
        cmd.append("./...")

        try:
            subprocess.run(
                cmd,
                cwd=fs_dir,
                capture_output=False,
                text=True,
                check=True
            )
            console.print("[green]✅ Go D-Bus tests passed![/green]")

        except subprocess.CalledProcessError as e:
            console.print(f"[red]Go D-Bus tests failed with exit code {e.returncode}[/red]")
            raise typer.Exit(1)
        except Exception as e:
            console.print(f"[red]Error running Go D-Bus tests: {e}[/red]")
            raise typer.Exit(1)
    else:
        console.print("[yellow]⚠️ Go filesystem module not found, skipping Go D-Bus tests[/yellow]")

    console.print("\n[green]🎉 Complete Nemo extension test suite passed![/green]")



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

    # Check Nemo extension tests
    nemo_dir = paths["project_root"] / "internal" / "nemo"
    if nemo_dir.exists():
        tests_dir = nemo_dir / "tests"
        test_runner = nemo_dir / "run_tests.py"

        if tests_dir.exists() and test_runner.exists():
            test_files = list(tests_dir.glob("test_*.py"))
            nemo_status = "✅ Available"
            nemo_details = f"{len(test_files)} Python test files"
        else:
            nemo_status = "⚠️ Partial"
            nemo_details = "Extension found but tests incomplete"
    else:
        nemo_status = "❌ Missing"
        nemo_details = "Nemo extension not found"

    env_table.add_row("Nemo Extension", nemo_status, nemo_details)

    console.print(env_table)
