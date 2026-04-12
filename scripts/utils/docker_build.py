"""
Docker-based Debian package building for OneMount.

Delegates to the canonical shell script (scripts/build-deb-package.sh)
for actual build logic. This module provides the Python CLI interface.
"""

import subprocess

from rich.console import Console

from .paths import get_project_paths

console = Console()


class DockerBuildError(Exception):
    """Exception raised when Docker build operations fail."""
    pass


def build_debian_package_docker(verbose: bool = False, clean: bool = False, force_rebuild_image: bool = False) -> bool:
    """
    Build Debian packages using Docker via the canonical shell script.

    Args:
        verbose: Enable verbose output
        clean: Clean before building
        force_rebuild_image: Force rebuild of Docker image (not used with shell script)

    Returns:
        True if build succeeded, False otherwise
    """
    paths = get_project_paths()
    script_path = paths["scripts_dir"] / "build-deb-package.sh"

    if not script_path.exists():
        console.print(f"[red]Build script not found: {script_path}[/red]")
        return False

    # Make sure script is executable
    script_path.chmod(0o755)

    # Handle clean flag
    if clean:
        console.print("[yellow]Cleaning build artifacts...[/yellow]")
        import shutil
        if paths["build_dir"].exists():
            shutil.rmtree(paths["build_dir"])
        paths["build_dir"].mkdir(parents=True, exist_ok=True)

    # Run the canonical build script
    try:
        console.print(f"[blue]Running build script: {script_path}[/blue]")
        result = subprocess.run(
            [str(script_path)],
            cwd=str(paths["project_root"]),
            check=False,
            capture_output=not verbose,
            text=True
        )

        if result.returncode != 0:
            console.print("[red]Build script failed[/red]")
            if not verbose and result.stderr:
                console.print(f"[red]Error output:[/red]\n{result.stderr}")
            return False

        if not verbose and result.stdout:
            # Show summary even in non-verbose mode
            lines = result.stdout.strip().split('\n')
            for line in lines[-10:]:  # Show last 10 lines
                console.print(line)

        return True

    except Exception as e:
        console.print(f"[red]Failed to run build script: {e}[/red]")
        return False
