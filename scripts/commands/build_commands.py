"""
Build and packaging commands for OneMount development CLI.
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
from utils.paths import ensure_build_directories, get_project_paths
from utils.shell import run_command, run_command_with_progress, run_script, ensure_executable
from utils.docker_build import build_debian_package_docker

console = Console()

# Create the build app
build_app = typer.Typer(
    name="build",
    help="Build and packaging operations",
    no_args_is_help=True,
)


@build_app.command()
def deb(
    ctx: typer.Context,
    docker: bool = typer.Option(False, "--docker", help="Use Docker for building"),
    native: bool = typer.Option(False, "--native", help="Use native tools for building"),
    clean: bool = typer.Option(False, "--clean", help="Clean before building"),
):
    """
    🔨 Build Debian packages.
    
    Choose between Docker-based building (recommended) or native building.
    Docker building provides a consistent environment and doesn't require
    local Debian packaging tools.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    if not docker and not native:
        console.print("[yellow]Please specify either --docker or --native[/yellow]")
        console.print("Use --docker for containerized building (recommended)")
        console.print("Use --native for local building (requires packaging tools)")
        raise typer.Exit(1)
    
    if clean:
        console.print("[yellow]Cleaning build artifacts...[/yellow]")
        paths = get_project_paths()
        if paths["build_dir"].exists():
            import shutil
            shutil.rmtree(paths["build_dir"])
        ensure_build_directories()
    
    paths = get_project_paths()
    
    if docker:
        console.print("[blue]Building Debian package with Docker...[/blue]")

        # Use native Python Docker implementation
        success = build_debian_package_docker(
            verbose=verbose,
            clean=clean,
            force_rebuild_image=False
        )

        if not success:
            console.print("[red]Docker build failed[/red]")
            raise typer.Exit(1)
        
    elif native:
        console.print("[blue]Building Debian package natively...[/blue]")
        
        # Check for required tools
        required_tools = ["dpkg-buildpackage", "debuild", "go"]
        missing_tools = []
        
        import shutil
        for tool in required_tools:
            if not shutil.which(tool):
                missing_tools.append(tool)
        
        if missing_tools:
            console.print(f"[red]Missing required tools: {', '.join(missing_tools)}[/red]")
            console.print("Install with: sudo apt-get install build-essential devscripts golang-go")
            raise typer.Exit(1)
        
        script_path = paths["legacy_scripts"]["build_deb_native"]
        
        if not script_path.exists():
            console.print(f"[red]Script not found: {script_path}[/red]")
            raise typer.Exit(1)
        
        ensure_executable(script_path)
        run_command_with_progress(
            [str(script_path)],
            "Building Debian package natively",
            verbose=verbose,
            timeout=1800,  # 30 minutes
        )
    
    # Show build results
    console.print("\n[green]✅ Build completed successfully![/green]")
    
    # List generated packages
    deb_dir = paths["deb_dir"]
    if deb_dir.exists():
        deb_files = list(deb_dir.glob("*.deb"))
        if deb_files:
            console.print(f"\n[cyan]📦 Generated packages in {deb_dir}:[/cyan]")
            for deb_file in deb_files:
                file_size = deb_file.stat().st_size / (1024 * 1024)  # MB
                console.print(f"  • {deb_file.name} ({file_size:.1f} MB)")


@build_app.command()
def manifest(
    ctx: typer.Context,
    target: str = typer.Option(..., help="Target packaging system (makefile/rpm/debian)"),
    type: Optional[str] = typer.Option(None, help="Installation type (user/system) - required for makefile"),
    action: str = typer.Option(..., help="Action to perform (install/uninstall/validate/files)"),
    dry_run: bool = typer.Option(False, "--dry-run", help="Show what would be done without executing"),
):
    """
    📋 Generate installation commands from manifest.
    
    This command uses the installation manifest to generate commands for
    different packaging systems and installation types.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    # Validate arguments
    valid_targets = ["makefile", "rpm", "debian"]
    valid_actions = ["install", "uninstall", "validate", "files"]
    
    if target not in valid_targets:
        console.print(f"[red]Invalid target: {target}. Must be one of: {', '.join(valid_targets)}[/red]")
        raise typer.Exit(1)
    
    if action not in valid_actions:
        console.print(f"[red]Invalid action: {action}. Must be one of: {', '.join(valid_actions)}[/red]")
        raise typer.Exit(1)
    
    if target == "makefile" and action in ["install", "uninstall"] and not type:
        console.print("[red]Error: --type is required for makefile install/uninstall actions[/red]")
        console.print("Use --type user for user installation or --type system for system-wide installation")
        raise typer.Exit(1)
    
    if type and type not in ["user", "system"]:
        console.print(f"[red]Invalid type: {type}. Must be 'user' or 'system'[/red]")
        raise typer.Exit(1)
    
    # Find manifest file
    paths = get_project_paths()
    manifest_path = paths["packaging_dir"] / "install-manifest.json"
    
    if not manifest_path.exists():
        console.print(f"[red]Error: Manifest file not found at {manifest_path}[/red]")
        console.print("Make sure you're in the OneMount project root directory")
        raise typer.Exit(1)
    
    # Import and use the manifest parser
    try:
        # Import the existing manifest parser
        manifest_parser_path = paths["legacy_scripts"]["manifest_parser"]
        if not manifest_parser_path.exists():
            console.print(f"[red]Manifest parser not found: {manifest_parser_path}[/red]")
            raise typer.Exit(1)
        
        # Add scripts directory to Python path and import
        import sys
        scripts_dir = str(paths["scripts_dir"])
        if scripts_dir not in sys.path:
            sys.path.insert(0, scripts_dir)
        
        from manifest_parser import InstallManifestParser
        
        parser_obj = InstallManifestParser(manifest_path)
        
        if verbose:
            console.print(f"[dim]Using manifest: {manifest_path}[/dim]")
            console.print(f"[dim]Target: {target}, Type: {type}, Action: {action}[/dim]")
        
        # Generate commands based on target and action
        commands = []
        
        if target == "makefile":
            if action == "install":
                commands = parser_obj.generate_makefile_install(type, dry_run)
            elif action == "uninstall":
                commands = parser_obj.generate_makefile_uninstall(type, dry_run)
            elif action == "validate":
                commands = parser_obj.generate_validation()
            elif action == "files":
                files = parser_obj.get_all_files(type)
                console.print(f"[cyan]Files for {type} installation:[/cyan]")
                for file_info in files:
                    console.print(f"  {file_info['source']} → {file_info['dest']}")
                return
        
        elif target == "rpm":
            if action == "install":
                commands = parser_obj.generate_rpm_install()
            elif action == "files":
                commands = parser_obj.generate_rpm_files()
        
        elif target == "debian":
            if action == "install":
                commands = parser_obj.generate_debian_install()
        
        # Output the commands
        if dry_run and action in ["install", "uninstall"]:
            console.print(f"[yellow]Dry run - showing what would be done for {action}:[/yellow]")
        
        for command in commands:
            console.print(command)
    
    except Exception as e:
        console.print(f"[red]Error generating commands: {e}[/red]")
        if verbose:
            import traceback
            console.print(f"[dim]{traceback.format_exc()}[/dim]")
        raise typer.Exit(1)


@build_app.command()
def binaries(
    ctx: typer.Context,
    target: Optional[str] = typer.Option(None, help="Specific binary to build (onemount/onemount-launcher/all)"),
    clean: bool = typer.Option(False, "--clean", help="Clean before building"),
):
    """
    🔧 Build OneMount binaries.
    
    Build the main OneMount binaries using the Makefile targets.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    if not ensure_environment():
        raise typer.Exit(1)
    
    ensure_build_directories()
    
    if clean:
        console.print("[yellow]Cleaning build artifacts...[/yellow]")
        from utils.shell import run_make_target
        run_make_target("clean", verbose=verbose)
    
    # Determine what to build
    if target is None or target == "all":
        targets = ["onemount", "onemount-launcher"]
    elif target in ["onemount", "onemount-launcher"]:
        targets = [target]
    else:
        console.print(f"[red]Invalid target: {target}. Must be 'onemount', 'onemount-launcher', or 'all'[/red]")
        raise typer.Exit(1)
    
    # Build each target
    from utils.shell import run_make_target
    
    for make_target in targets:
        console.print(f"[blue]Building {make_target}...[/blue]")
        run_command_with_progress(
            ["make", make_target],
            f"Building {make_target}",
            verbose=verbose,
            timeout=600,  # 10 minutes
        )
    
    # Show build results
    console.print("\n[green]✅ Build completed successfully![/green]")
    
    # List generated binaries
    from utils.paths import get_binary_paths
    binary_paths = get_binary_paths()
    
    console.print(f"\n[cyan]🔧 Generated binaries:[/cyan]")
    for name, path in binary_paths.items():
        if path.exists():
            file_size = path.stat().st_size / (1024 * 1024)  # MB
            console.print(f"  • {name}: {path} ({file_size:.1f} MB)")
        else:
            console.print(f"  • {name}: [dim]not built[/dim]")


@build_app.command()
def status(ctx: typer.Context):
    """
    📊 Show build status and information.
    
    Display information about the current build state, available binaries,
    and build directory structure.
    """
    verbose = ctx.obj.get("verbose", False) if ctx.obj else False
    
    console.print(Panel.fit(
        "[bold blue]OneMount Build Status[/bold blue]",
        border_style="blue"
    ))
    
    # Build directory status
    paths = get_project_paths()
    
    console.print("\n[bold cyan]📁 Build Directory Structure[/bold cyan]")
    structure_table = Table()
    structure_table.add_column("Directory", style="cyan")
    structure_table.add_column("Status", style="green")
    structure_table.add_column("Contents", style="dim")
    
    build_dirs = [
        ("Build Root", paths["build_dir"]),
        ("Binaries", paths["binaries_dir"]),
        ("Packages", paths["packages_dir"]),
        ("Debian Packages", paths["deb_dir"]),
        ("RPM Packages", paths["rpm_dir"]),
        ("Source Packages", paths["source_dir"]),
        ("Docker Build", paths["docker_dir"]),
        ("Temporary", paths["temp_dir"]),
    ]
    
    for name, path in build_dirs:
        if path.exists():
            contents = list(path.iterdir()) if path.is_dir() else []
            content_count = len(contents)
            status = "✅ Exists"
            content_info = f"{content_count} items" if content_count > 0 else "empty"
        else:
            status = "❌ Missing"
            content_info = "not created"
        
        structure_table.add_row(name, status, content_info)
    
    console.print(structure_table)
    
    # Binary status
    console.print("\n[bold cyan]🔧 Binary Status[/bold cyan]")
    from utils.paths import get_binary_paths
    binary_paths = get_binary_paths()
    
    binary_table = Table()
    binary_table.add_column("Binary", style="cyan")
    binary_table.add_column("Status", style="green")
    binary_table.add_column("Size", style="dim")
    binary_table.add_column("Modified", style="dim")
    
    for name, path in binary_paths.items():
        if path.exists():
            file_size = path.stat().st_size / (1024 * 1024)  # MB
            import datetime
            modified = datetime.datetime.fromtimestamp(path.stat().st_mtime)
            status = "✅ Built"
            size_info = f"{file_size:.1f} MB"
            modified_info = modified.strftime("%Y-%m-%d %H:%M")
        else:
            status = "❌ Not built"
            size_info = "-"
            modified_info = "-"
        
        binary_table.add_row(name, status, size_info, modified_info)
    
    console.print(binary_table)
    
    # Package status
    console.print("\n[bold cyan]📦 Package Status[/bold cyan]")
    from utils.paths import get_package_paths
    package_paths = get_package_paths()
    
    package_table = Table()
    package_table.add_column("Package Type", style="cyan")
    package_table.add_column("Directory", style="dim")
    package_table.add_column("Packages", style="green")
    
    for pkg_type, pkg_dir in package_paths.items():
        if pkg_dir.exists():
            if pkg_type == "deb_dir":
                packages = list(pkg_dir.glob("*.deb"))
                pkg_type_name = "Debian"
            elif pkg_type == "rpm_dir":
                packages = list(pkg_dir.glob("*.rpm"))
                pkg_type_name = "RPM"
            elif pkg_type == "source_dir":
                packages = list(pkg_dir.glob("*.tar.*"))
                pkg_type_name = "Source"
            else:
                packages = []
                pkg_type_name = pkg_type
            
            package_count = f"{len(packages)} packages"
        else:
            package_count = "Directory not found"
            pkg_type_name = pkg_type.replace("_dir", "").upper()
        
        package_table.add_row(pkg_type_name, str(pkg_dir), package_count)
    
    console.print(package_table)
