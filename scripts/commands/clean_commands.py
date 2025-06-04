"""
Cleanup commands for OneMount development CLI.
"""

import shutil
import sys
from pathlib import Path
from typing import List

import typer
from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.prompt import Confirm

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from utils.environment import ensure_environment
from utils.paths import get_cleanable_paths, get_project_paths

console = Console()

# Create the clean app
clean_app = typer.Typer(
    name="clean",
    help="Cleanup operations",
    no_args_is_help=True,
)


def format_size(size_bytes: int) -> str:
    """Format file size in human-readable format."""
    for unit in ['B', 'KB', 'MB', 'GB']:
        if size_bytes < 1024.0:
            return f"{size_bytes:.1f} {unit}"
        size_bytes /= 1024.0
    return f"{size_bytes:.1f} TB"


def get_directory_size(path: Path) -> int:
    """Get total size of directory and all its contents."""
    total_size = 0
    try:
        if path.is_file():
            return path.stat().st_size
        elif path.is_dir():
            for item in path.rglob('*'):
                if item.is_file():
                    try:
                        total_size += item.stat().st_size
                    except (OSError, FileNotFoundError):
                        # Skip files that can't be accessed
                        pass
    except (OSError, FileNotFoundError):
        pass
    return total_size


def clean_paths(paths: List[Path], description: str, dry_run: bool = False) -> tuple[int, int]:
    """
    Clean a list of paths.
    
    Returns:
        Tuple of (files_removed, total_size_freed)
    """
    files_removed = 0
    total_size_freed = 0
    
    for path in paths:
        if not path.exists():
            continue
        
        try:
            size = get_directory_size(path)
            
            if dry_run:
                console.print(f"  Would remove: {path} ({format_size(size)})")
                total_size_freed += size
                files_removed += 1 if path.is_file() else len(list(path.rglob('*')))
            else:
                if path.is_file():
                    path.unlink()
                    files_removed += 1
                elif path.is_dir():
                    file_count = len(list(path.rglob('*')))
                    shutil.rmtree(path)
                    files_removed += file_count
                
                total_size_freed += size
                console.print(f"  Removed: {path} ({format_size(size)})")
        
        except (OSError, PermissionError) as e:
            console.print(f"  [yellow]Warning: Could not remove {path}: {e}[/yellow]")
    
    return files_removed, total_size_freed


@clean_app.command()
def list(ctx: typer.Context):
    """
    📋 List all cleanable artifacts and their sizes.
    
    Shows what files and directories can be cleaned up, organized by category,
    along with their sizes.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print(Panel.fit(
        "[bold blue]Cleanable Artifacts[/bold blue]",
        border_style="blue"
    ))
    
    cleanable = get_cleanable_paths()
    
    total_size = 0
    total_items = 0
    
    for category, paths in cleanable.items():
        if not paths:
            continue
        
        # Filter to existing paths
        existing_paths = [p for p in paths if p.exists()]
        if not existing_paths:
            continue
        
        console.print(f"\n[bold cyan]🗂️  {category.replace('_', ' ').title()}[/bold cyan]")
        
        category_table = Table()
        category_table.add_column("Path", style="dim")
        category_table.add_column("Type", style="cyan")
        category_table.add_column("Size", style="green")
        
        category_size = 0
        category_items = 0
        
        for path in existing_paths:
            size = get_directory_size(path)
            path_type = "Directory" if path.is_dir() else "File"
            
            category_table.add_row(
                str(path.relative_to(get_project_paths()["project_root"])),
                path_type,
                format_size(size)
            )
            
            category_size += size
            category_items += 1
        
        console.print(category_table)
        console.print(f"[dim]Category total: {category_items} items, {format_size(category_size)}[/dim]")
        
        total_size += category_size
        total_items += category_items
    
    console.print(f"\n[bold green]📊 Total: {total_items} items, {format_size(total_size)}[/bold green]")


@clean_app.command()
def build(
    ctx: typer.Context,
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be cleaned without doing it"),
    force: bool = typer.Option(False, "--force", help="Skip confirmation prompt"),
):
    """
    🔨 Clean build artifacts.
    
    Removes all build artifacts including binaries, packages, and temporary build files.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print("[blue]Cleaning build artifacts...[/blue]")
    
    cleanable = get_cleanable_paths()
    build_paths = cleanable.get("build_artifacts", [])
    
    if not build_paths:
        console.print("[green]No build artifacts to clean.[/green]")
        return
    
    # Filter to existing paths
    existing_paths = [p for p in build_paths if p.exists()]
    if not existing_paths:
        console.print("[green]No build artifacts found.[/green]")
        return
    
    # Calculate total size
    total_size = sum(get_directory_size(p) for p in existing_paths)
    
    if not dry_run and not force:
        console.print(f"This will remove {len(existing_paths)} build artifacts ({format_size(total_size)})")
        if not Confirm.ask("Continue?"):
            console.print("[yellow]Cancelled.[/yellow]")
            return
    
    files_removed, size_freed = clean_paths(existing_paths, "build artifacts", dry_run)
    
    if dry_run:
        console.print(f"[yellow]Dry run: Would remove {files_removed} items ({format_size(size_freed)})[/yellow]")
    else:
        console.print(f"[green]✅ Cleaned {files_removed} items, freed {format_size(size_freed)}[/green]")


@clean_app.command()
def coverage(
    ctx: typer.Context,
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be cleaned without doing it"),
    force: bool = typer.Option(False, "--force", help="Skip confirmation prompt"),
):
    """
    📊 Clean coverage files.
    
    Removes coverage reports and data files.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print("[blue]Cleaning coverage files...[/blue]")
    
    cleanable = get_cleanable_paths()
    coverage_paths = cleanable.get("coverage_files", [])
    
    if not coverage_paths:
        console.print("[green]No coverage files to clean.[/green]")
        return
    
    # Filter to existing paths
    existing_paths = [p for p in coverage_paths if p.exists()]
    if not existing_paths:
        console.print("[green]No coverage files found.[/green]")
        return
    
    # Calculate total size
    total_size = sum(get_directory_size(p) for p in existing_paths)
    
    if not dry_run and not force:
        console.print(f"This will remove {len(existing_paths)} coverage files ({format_size(total_size)})")
        if not Confirm.ask("Continue?"):
            console.print("[yellow]Cancelled.[/yellow]")
            return
    
    files_removed, size_freed = clean_paths(existing_paths, "coverage files", dry_run)
    
    if dry_run:
        console.print(f"[yellow]Dry run: Would remove {files_removed} items ({format_size(size_freed)})[/yellow]")
    else:
        console.print(f"[green]✅ Cleaned {files_removed} items, freed {format_size(size_freed)}[/green]")


@clean_app.command()
def temp(
    ctx: typer.Context,
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be cleaned without doing it"),
    force: bool = typer.Option(False, "--force", help="Skip confirmation prompt"),
):
    """
    🗑️  Clean temporary files.
    
    Removes temporary files, swap files, and other transient artifacts.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print("[blue]Cleaning temporary files...[/blue]")
    
    cleanable = get_cleanable_paths()
    temp_paths = (
        cleanable.get("temp_files", []) +
        cleanable.get("log_files", []) +
        cleanable.get("go_cache", [])
    )
    
    if not temp_paths:
        console.print("[green]No temporary files to clean.[/green]")
        return
    
    # Filter to existing paths
    existing_paths = [p for p in temp_paths if p.exists()]
    if not existing_paths:
        console.print("[green]No temporary files found.[/green]")
        return
    
    # Calculate total size
    total_size = sum(get_directory_size(p) for p in existing_paths)
    
    if not dry_run and not force:
        console.print(f"This will remove {len(existing_paths)} temporary files ({format_size(total_size)})")
        if not Confirm.ask("Continue?"):
            console.print("[yellow]Cancelled.[/yellow]")
            return
    
    files_removed, size_freed = clean_paths(existing_paths, "temporary files", dry_run)
    
    if dry_run:
        console.print(f"[yellow]Dry run: Would remove {files_removed} items ({format_size(size_freed)})[/yellow]")
    else:
        console.print(f"[green]✅ Cleaned {files_removed} items, freed {format_size(size_freed)}[/green]")


@clean_app.command()
def packages(
    ctx: typer.Context,
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be cleaned without doing it"),
    force: bool = typer.Option(False, "--force", help="Skip confirmation prompt"),
):
    """
    📦 Clean package files.
    
    Removes built packages (.deb, .rpm, etc.) from the project root.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print("[blue]Cleaning package files...[/blue]")
    
    cleanable = get_cleanable_paths()
    package_paths = cleanable.get("package_files", [])
    
    if not package_paths:
        console.print("[green]No package files to clean.[/green]")
        return
    
    # Filter to existing paths
    existing_paths = [p for p in package_paths if p.exists()]
    if not existing_paths:
        console.print("[green]No package files found.[/green]")
        return
    
    # Calculate total size
    total_size = sum(get_directory_size(p) for p in existing_paths)
    
    if not dry_run and not force:
        console.print(f"This will remove {len(existing_paths)} package files ({format_size(total_size)})")
        if not Confirm.ask("Continue?"):
            console.print("[yellow]Cancelled.[/yellow]")
            return
    
    files_removed, size_freed = clean_paths(existing_paths, "package files", dry_run)
    
    if dry_run:
        console.print(f"[yellow]Dry run: Would remove {files_removed} items ({format_size(size_freed)})[/yellow]")
    else:
        console.print(f"[green]✅ Cleaned {files_removed} items, freed {format_size(size_freed)}[/green]")


@clean_app.command()
def python(
    ctx: typer.Context,
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be cleaned without doing it"),
    force: bool = typer.Option(False, "--force", help="Skip confirmation prompt"),
):
    """
    🐍 Clean Python cache files.
    
    Removes Python cache files and bytecode from the scripts directory.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print("[blue]Cleaning Python cache files...[/blue]")
    
    cleanable = get_cleanable_paths()
    python_paths = cleanable.get("python_cache", [])
    
    if not python_paths:
        console.print("[green]No Python cache files to clean.[/green]")
        return
    
    # Filter to existing paths
    existing_paths = [p for p in python_paths if p.exists()]
    if not existing_paths:
        console.print("[green]No Python cache files found.[/green]")
        return
    
    # Calculate total size
    total_size = sum(get_directory_size(p) for p in existing_paths)
    
    if not dry_run and not force:
        console.print(f"This will remove {len(existing_paths)} Python cache files ({format_size(total_size)})")
        if not Confirm.ask("Continue?"):
            console.print("[yellow]Cancelled.[/yellow]")
            return
    
    files_removed, size_freed = clean_paths(existing_paths, "Python cache files", dry_run)
    
    if dry_run:
        console.print(f"[yellow]Dry run: Would remove {files_removed} items ({format_size(size_freed)})[/yellow]")
    else:
        console.print(f"[green]✅ Cleaned {files_removed} items, freed {format_size(size_freed)}[/green]")


@clean_app.command()
def all(
    ctx: typer.Context,
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be cleaned without doing it"),
    force: bool = typer.Option(False, "--force", help="Skip confirmation prompt"),
):
    """
    🧹 Clean all artifacts.
    
    Removes all cleanable artifacts including build files, coverage reports,
    temporary files, packages, and cache files.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print("[blue]Cleaning all artifacts...[/blue]")
    
    cleanable = get_cleanable_paths()
    
    # Collect all paths
    all_paths = []
    for category_paths in cleanable.values():
        all_paths.extend(category_paths)
    
    if not all_paths:
        console.print("[green]No artifacts to clean.[/green]")
        return
    
    # Filter to existing paths
    existing_paths = [p for p in all_paths if p.exists()]
    if not existing_paths:
        console.print("[green]No artifacts found.[/green]")
        return
    
    # Calculate total size
    total_size = sum(get_directory_size(p) for p in existing_paths)
    
    if not dry_run and not force:
        console.print(f"This will remove {len(existing_paths)} artifacts ({format_size(total_size)})")
        console.print("[yellow]This includes build artifacts, coverage files, temporary files, packages, and cache files.[/yellow]")
        if not Confirm.ask("Continue?"):
            console.print("[yellow]Cancelled.[/yellow]")
            return
    
    files_removed, size_freed = clean_paths(existing_paths, "all artifacts", dry_run)
    
    if dry_run:
        console.print(f"[yellow]Dry run: Would remove {files_removed} items ({format_size(size_freed)})[/yellow]")
    else:
        console.print(f"[green]✅ Cleaned {files_removed} items, freed {format_size(size_freed)}[/green]")
