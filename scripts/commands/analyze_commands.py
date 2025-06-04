"""
Code and project analysis commands for OneMount development CLI.
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
from utils.shell import run_command, run_command_with_progress

console = Console()

# Create the analyze app
analyze_app = typer.Typer(
    name="analyze",
    help="Code and project analysis tools",
    no_args_is_help=True,
)


@analyze_app.command()
def test_suite(
    ctx: typer.Context,
    mode: str = typer.Option("analyze", help="Analysis mode (analyze/resolve)"),
    output_dir: str = typer.Option("tmp", help="Output directory for reports"),
):
    """
    🧪 Analyze test suite for duplicates and issues.
    
    Analyzes the test suite to find duplicate test IDs, naming issues,
    and other potential problems.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    valid_modes = ["analyze", "resolve"]
    if mode not in valid_modes:
        console.print(f"[red]Invalid mode: {mode}. Must be one of: {', '.join(valid_modes)}[/red]")
        raise typer.Exit(1)
    
    console.print(f"[blue]Analyzing test suite in {mode} mode...[/blue]")
    
    paths = get_project_paths()
    script_path = paths["scripts_dir"] / "test_suite_tool.py"
    
    if not script_path.exists():
        console.print(f"[red]Test suite analysis script not found: {script_path}[/red]")
        console.print("[yellow]This functionality may have been moved or removed.[/yellow]")
        raise typer.Exit(1)
    
    cmd = [sys.executable, str(script_path), f"--{mode}", output_dir]
    
    try:
        run_command_with_progress(
            cmd,
            f"Analyzing test suite ({mode})",
            verbose=verbose,
            timeout=300,  # 5 minutes
        )
        
        console.print(f"[green]✅ Test suite analysis ({mode}) completed![/green]")
        console.print(f"[dim]Check {output_dir} for detailed reports[/dim]")
    
    except Exception as e:
        console.print(f"[red]Test suite analysis failed: {e}[/red]")
        raise typer.Exit(1)


@analyze_app.command()
def coverage_trends(
    ctx: typer.Context,
    input_file: str = typer.Option(..., "--input", help="Input coverage history JSON file"),
    output_file: str = typer.Option(..., "--output", help="Output HTML report file"),
    plot: bool = typer.Option(False, "--plot", help="Generate trend plot"),
):
    """
    📊 Analyze coverage trends over time.
    
    Analyzes coverage trends from historical data and generates reports.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Analyzing coverage trends...[/blue]")
    
    # Check if input file exists
    input_path = Path(input_file)
    if not input_path.exists():
        console.print(f"[red]Input file not found: {input_path}[/red]")
        raise typer.Exit(1)
    
    paths = get_project_paths()
    script_path = paths["scripts_dir"] / "coverage-trend-analysis.py"
    
    if not script_path.exists():
        console.print(f"[red]Coverage trend analysis script not found: {script_path}[/red]")
        console.print("[yellow]This functionality may have been moved or removed.[/yellow]")
        raise typer.Exit(1)
    
    cmd = [
        sys.executable, str(script_path),
        "--input", str(input_path),
        "--output", output_file
    ]
    
    if plot:
        cmd.append("--plot")
    
    try:
        run_command_with_progress(
            cmd,
            "Analyzing coverage trends",
            verbose=verbose,
            timeout=180,  # 3 minutes
        )
        
        console.print("[green]✅ Coverage trends analysis completed![/green]")
        console.print(f"[dim]Report saved to: {output_file}[/dim]")
    
    except Exception as e:
        console.print(f"[red]Coverage trends analysis failed: {e}[/red]")
        raise typer.Exit(1)


@analyze_app.command()
def code_quality(ctx: typer.Context):
    """
    🔍 Analyze code quality metrics.
    
    Runs various code quality checks including linting, formatting,
    and static analysis.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Analyzing code quality...[/blue]")
    
    # Check if Go is available
    import shutil
    if not shutil.which("go"):
        console.print("[red]Go is not available.[/red]")
        raise typer.Exit(1)
    
    console.print("\n[bold cyan]🔍 Running Code Quality Checks[/bold cyan]")
    
    checks = []
    
    # Go vet
    console.print("[dim]Running go vet...[/dim]")
    try:
        result = run_command(
            ["go", "vet", "./..."],
            capture_output=True,
            verbose=verbose,
        )
        checks.append(("go vet", "✅ Passed", "No issues found"))
    except Exception as e:
        checks.append(("go vet", "❌ Failed", str(e)))
    
    # Go fmt check
    console.print("[dim]Checking go fmt...[/dim]")
    try:
        result = run_command(
            ["go", "fmt", "./..."],
            capture_output=True,
            verbose=verbose,
        )
        if result.stdout.strip():
            checks.append(("go fmt", "⚠️  Issues", "Files need formatting"))
        else:
            checks.append(("go fmt", "✅ Passed", "All files formatted"))
    except Exception as e:
        checks.append(("go fmt", "❌ Failed", str(e)))
    
    # Go mod tidy check
    console.print("[dim]Checking go mod tidy...[/dim]")
    try:
        # Run go mod tidy and check if anything changed
        result = run_command(
            ["go", "mod", "tidy"],
            capture_output=True,
            verbose=verbose,
        )
        
        # Check if go.mod or go.sum changed
        from utils.git import get_modified_files
        modified = get_modified_files()
        mod_files = [f for f in modified if f in ["go.mod", "go.sum"]]
        
        if mod_files:
            checks.append(("go mod tidy", "⚠️  Issues", f"Modified: {', '.join(mod_files)}"))
        else:
            checks.append(("go mod tidy", "✅ Passed", "Dependencies are tidy"))
    except Exception as e:
        checks.append(("go mod tidy", "❌ Failed", str(e)))
    
    # Check for common issues
    console.print("[dim]Checking for common issues...[/dim]")
    try:
        # Look for TODO/FIXME comments
        result = run_command(
            ["grep", "-r", "--include=*.go", "-n", "-i", "TODO\\|FIXME", "."],
            capture_output=True,
            check=False,
            verbose=verbose,
        )
        
        if result.stdout.strip():
            todo_count = len(result.stdout.strip().split('\n'))
            checks.append(("TODO/FIXME", "⚠️  Found", f"{todo_count} items need attention"))
        else:
            checks.append(("TODO/FIXME", "✅ Clean", "No pending items"))
    except Exception:
        checks.append(("TODO/FIXME", "❓ Unknown", "Could not check"))
    
    # Display results
    quality_table = Table()
    quality_table.add_column("Check", style="cyan")
    quality_table.add_column("Status", style="green")
    quality_table.add_column("Details", style="dim")
    
    for check, status, details in checks:
        quality_table.add_row(check, status, details)
    
    console.print(quality_table)
    
    # Summary
    passed = sum(1 for _, status, _ in checks if "✅" in status)
    total = len(checks)
    
    if passed == total:
        console.print(f"\n[green]✅ All {total} code quality checks passed![/green]")
    else:
        failed = total - passed
        console.print(f"\n[yellow]⚠️  {failed} of {total} checks need attention[/yellow]")


@analyze_app.command()
def dependencies(ctx: typer.Context):
    """
    📦 Analyze project dependencies.
    
    Analyzes Go module dependencies, versions, and potential issues.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    console.print("[blue]Analyzing project dependencies...[/blue]")
    
    # Check if Go is available
    import shutil
    if not shutil.which("go"):
        console.print("[red]Go is not available.[/red]")
        raise typer.Exit(1)
    
    console.print("\n[bold cyan]📦 Go Module Dependencies[/bold cyan]")
    
    try:
        # Get direct dependencies
        result = run_command(
            ["go", "list", "-m", "-f", "{{.Path}} {{.Version}}", "all"],
            capture_output=True,
            verbose=verbose,
        )
        
        if result.stdout.strip():
            deps = []
            lines = result.stdout.strip().split('\n')
            
            for line in lines[1:]:  # Skip the first line (current module)
                if ' ' in line:
                    path, version = line.split(' ', 1)
                    deps.append((path, version))
            
            # Show top-level dependencies
            deps_table = Table()
            deps_table.add_column("Module", style="cyan")
            deps_table.add_column("Version", style="green")
            deps_table.add_column("Type", style="yellow")
            
            # Get direct dependencies
            direct_result = run_command(
                ["go", "list", "-m", "-f", "{{.Path}} {{.Version}}", "-mod=readonly"],
                capture_output=True,
                check=False,
                verbose=verbose,
            )
            
            direct_deps = set()
            if direct_result.stdout:
                for line in direct_result.stdout.strip().split('\n')[1:]:
                    if ' ' in line:
                        path, _ = line.split(' ', 1)
                        direct_deps.add(path)
            
            # Display dependencies (limit to first 20)
            for path, version in deps[:20]:
                dep_type = "Direct" if path in direct_deps else "Indirect"
                deps_table.add_row(path, version, dep_type)
            
            console.print(deps_table)
            
            if len(deps) > 20:
                console.print(f"[dim]... and {len(deps) - 20} more dependencies[/dim]")
            
            console.print(f"\n[dim]Total dependencies: {len(deps)}[/dim]")
        else:
            console.print("[yellow]No dependencies found.[/yellow]")
    
    except Exception as e:
        console.print(f"[red]Failed to analyze dependencies: {e}[/red]")
        raise typer.Exit(1)
    
    # Check for outdated dependencies
    console.print("\n[bold cyan]🔄 Dependency Updates[/bold cyan]")
    
    try:
        result = run_command(
            ["go", "list", "-u", "-m", "all"],
            capture_output=True,
            check=False,
            verbose=verbose,
        )
        
        if result.stdout.strip():
            outdated = []
            lines = result.stdout.strip().split('\n')
            
            for line in lines:
                if '[' in line and ']' in line:  # Has update available
                    parts = line.split()
                    if len(parts) >= 2:
                        module = parts[0]
                        current = parts[1]
                        if '[' in line:
                            available = line.split('[')[1].split(']')[0]
                            outdated.append((module, current, available))
            
            if outdated:
                update_table = Table()
                update_table.add_column("Module", style="cyan")
                update_table.add_column("Current", style="yellow")
                update_table.add_column("Available", style="green")
                
                for module, current, available in outdated[:10]:
                    update_table.add_row(module, current, available)
                
                console.print(update_table)
                
                if len(outdated) > 10:
                    console.print(f"[dim]... and {len(outdated) - 10} more updates available[/dim]")
                
                console.print(f"\n[yellow]⚠️  {len(outdated)} dependencies have updates available[/yellow]")
                console.print("[dim]Run 'go get -u ./...' to update all dependencies[/dim]")
            else:
                console.print("[green]✅ All dependencies are up to date[/green]")
        else:
            console.print("[yellow]Could not check for updates.[/yellow]")
    
    except Exception as e:
        console.print(f"[yellow]Warning: Could not check for updates: {e}[/yellow]")


@analyze_app.command()
def project_stats(ctx: typer.Context):
    """
    📈 Show project statistics.
    
    Display various statistics about the project including lines of code,
    file counts, and other metrics.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print(Panel.fit(
        "[bold blue]OneMount Project Statistics[/bold blue]",
        border_style="blue"
    ))
    
    paths = get_project_paths()
    project_root = paths["project_root"]
    
    # File statistics
    console.print("\n[bold cyan]📁 File Statistics[/bold cyan]")
    
    file_stats = {}
    total_files = 0
    total_lines = 0
    
    # Count files by extension
    for file_path in project_root.rglob("*"):
        if file_path.is_file() and not any(part.startswith('.') for part in file_path.parts):
            # Skip hidden files and directories
            if file_path.name.startswith('.'):
                continue
            
            # Skip build and vendor directories
            if any(part in ['build', 'vendor', 'node_modules', '__pycache__'] for part in file_path.parts):
                continue
            
            ext = file_path.suffix.lower()
            if not ext:
                ext = 'no extension'
            
            if ext not in file_stats:
                file_stats[ext] = {'count': 0, 'lines': 0}
            
            file_stats[ext]['count'] += 1
            total_files += 1
            
            # Count lines for text files
            if ext in ['.go', '.py', '.sh', '.md', '.txt', '.json', '.yaml', '.yml']:
                try:
                    with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                        lines = len(f.readlines())
                        file_stats[ext]['lines'] += lines
                        total_lines += lines
                except Exception:
                    pass
    
    # Display file statistics
    files_table = Table()
    files_table.add_column("File Type", style="cyan")
    files_table.add_column("Count", style="green")
    files_table.add_column("Lines", style="yellow")
    
    # Sort by count
    sorted_stats = sorted(file_stats.items(), key=lambda x: x[1]['count'], reverse=True)
    
    for ext, stats in sorted_stats[:10]:  # Top 10
        files_table.add_row(
            ext,
            str(stats['count']),
            str(stats['lines']) if stats['lines'] > 0 else "-"
        )
    
    console.print(files_table)
    console.print(f"[dim]Total: {total_files} files, {total_lines:,} lines of code[/dim]")
    
    # Directory structure
    console.print("\n[bold cyan]📂 Directory Structure[/bold cyan]")
    
    important_dirs = [
        ("cmd", "Command-line applications"),
        ("internal", "Internal implementation"),
        ("pkg", "Reusable packages"),
        ("tests", "Test files"),
        ("scripts", "Development scripts"),
        ("docs", "Documentation"),
        ("assets", "Application assets"),
    ]
    
    dirs_table = Table()
    dirs_table.add_column("Directory", style="cyan")
    dirs_table.add_column("Files", style="green")
    dirs_table.add_column("Purpose", style="dim")
    
    for dir_name, purpose in important_dirs:
        dir_path = project_root / dir_name
        if dir_path.exists():
            file_count = len([f for f in dir_path.rglob("*") if f.is_file()])
            dirs_table.add_row(dir_name, str(file_count), purpose)
        else:
            dirs_table.add_row(dir_name, "0", f"{purpose} (missing)")
    
    console.print(dirs_table)
