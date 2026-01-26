"""
Native Python implementation for Docker-based package building.
Replaces build-deb-docker.sh with native Python Docker operations.
"""

import os
import shutil
import tarfile
import tempfile
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import docker
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn

from .paths import get_project_paths
from .shell import run_command, CommandError
from .git import get_git_info

console = Console()


class DockerBuildError(Exception):
    """Exception raised when Docker build operations fail."""
    pass


class DockerPackageBuilder:
    """Native Python Docker package builder for OneMount."""
    
    def __init__(self, verbose: bool = False):
        self.verbose = verbose
        self.docker_client = None
        self.paths = get_project_paths()
        
    def __enter__(self):
        """Context manager entry."""
        try:
            self.docker_client = docker.from_env()
            # Test Docker connection
            self.docker_client.ping()
            return self
        except Exception as e:
            raise DockerBuildError(f"Failed to connect to Docker: {e}")
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        if self.docker_client:
            self.docker_client.close()
    
    def _log_info(self, message: str):
        """Log info message."""
        console.print(f"[blue][INFO][/blue] {message}")
    
    def _log_success(self, message: str):
        """Log success message."""
        console.print(f"[green][SUCCESS][/green] {message}")
    
    def _log_warning(self, message: str):
        """Log warning message."""
        console.print(f"[yellow][WARNING][/yellow] {message}")
    
    def _log_error(self, message: str):
        """Log error message."""
        console.print(f"[red][ERROR][/red] {message}")
    
    def check_docker_available(self) -> bool:
        """Check if Docker is available and running."""
        try:
            if not shutil.which("docker"):
                self._log_error("Docker is not installed or not in PATH")
                return False
            
            # Test Docker daemon
            result = run_command(
                ["docker", "info"],
                capture_output=True,
                check=False,
                verbose=False,
                timeout=10
            )
            
            if result.returncode != 0:
                self._log_error("Docker daemon is not running")
                return False
            
            return True
            
        except Exception as e:
            self._log_error(f"Failed to check Docker availability: {e}")
            return False
    
    def get_version_info(self) -> Tuple[str, str]:
        """Extract version information from spec file."""
        spec_file = self.paths["project_root"] / "packaging" / "rpm" / "onemount.spec"
        
        if not spec_file.exists():
            raise DockerBuildError(f"Spec file not found: {spec_file}")
        
        version = None
        release = None
        
        with open(spec_file, 'r') as f:
            for line in f:
                line = line.strip()
                if line.startswith("Version:"):
                    version = line.split(":", 1)[1].strip()
                elif line.startswith("Release:"):
                    # Extract just the number part
                    release_part = line.split(":", 1)[1].strip()
                    release = release_part.split()[0]  # Get first part before any spaces
        
        if not version or not release:
            raise DockerBuildError("Could not extract version/release from spec file")
        
        return version, release
    
    def ensure_docker_image(self, force_rebuild: bool = False) -> str:
        """Ensure Docker build image exists and is up to date."""
        image_name = "onemount-deb-builder"
        dockerfile_path = self.paths["project_root"] / "docker" / "images" / "deb-builder" / "Dockerfile"
        
        if not dockerfile_path.exists():
            raise DockerBuildError(f"Dockerfile not found: {dockerfile_path}")
        
        try:
            # Check if image exists
            image_exists = False
            try:
                image = self.docker_client.images.get(image_name)
                image_exists = True
                
                if not force_rebuild:
                    # Check if Dockerfile is newer than image
                    dockerfile_time = dockerfile_path.stat().st_mtime
                    image_created = image.attrs['Created']
                    
                    # Parse Docker timestamp
                    from datetime import datetime
                    import dateutil.parser
                    image_time = dateutil.parser.parse(image_created).timestamp()
                    
                    if dockerfile_time <= image_time:
                        self._log_info(f"Docker image '{image_name}' is up to date")
                        return image_name
                
            except docker.errors.ImageNotFound:
                image_exists = False
            
            # Build or rebuild image
            if not image_exists:
                self._log_info(f"Docker image '{image_name}' not found, building it...")
            else:
                self._log_info("Dockerfile is newer than image, rebuilding...")
            
            # Build the image
            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
                transient=True,
            ) as progress:
                task = progress.add_task("Building Docker image...", total=None)
                
                try:
                    image, build_logs = self.docker_client.images.build(
                        path=str(self.paths["project_root"]),
                        dockerfile=str(dockerfile_path.relative_to(self.paths["project_root"])),
                        tag=image_name,
                        rm=True,
                        forcerm=True,
                    )
                    
                    if self.verbose:
                        for log in build_logs:
                            if 'stream' in log:
                                console.print(f"[dim]{log['stream'].strip()}[/dim]")
                    
                    progress.update(task, description="✅ Docker image built successfully")
                    
                except docker.errors.BuildError as e:
                    progress.update(task, description="❌ Docker image build failed")
                    raise DockerBuildError(f"Failed to build Docker image: {e}")
            
            self._log_success("Docker image built successfully")
            return image_name
            
        except Exception as e:
            raise DockerBuildError(f"Failed to ensure Docker image: {e}")
    
    def prepare_build_directories(self):
        """Create and clean build directory structure."""
        self._log_info("Creating build directory structure...")
        
        # Create directories
        build_dirs = [
            self.paths["build_dir"],
            self.paths["packages_dir"],
            self.paths["deb_dir"],
            self.paths["temp_dir"],
        ]
        
        for dir_path in build_dirs:
            dir_path.mkdir(parents=True, exist_ok=True)
        
        # Clean up previous builds
        self._log_info("Cleaning up previous builds...")
        
        # Clean temp directory
        if self.paths["temp_dir"].exists():
            shutil.rmtree(self.paths["temp_dir"])
            self.paths["temp_dir"].mkdir(parents=True)
        
        # Clean deb packages
        for deb_file in self.paths["deb_dir"].glob("*.deb"):
            deb_file.unlink()
        for dsc_file in self.paths["deb_dir"].glob("*.dsc"):
            dsc_file.unlink()
        for changes_file in self.paths["deb_dir"].glob("*.changes"):
            changes_file.unlink()
        for tar_file in self.paths["deb_dir"].glob("*.tar.*"):
            tar_file.unlink()
        
        # Clean project root artifacts
        project_root = self.paths["project_root"]
        for pattern in ["*.deb", "*.dsc", "*.changes", "*.tar.*", "*.build*", "*.upload"]:
            for file_path in project_root.glob(pattern):
                if file_path.is_file():
                    file_path.unlink()
        
        # Clean specific files
        for filename in ["filelist.txt", ".commit"]:
            file_path = project_root / filename
            if file_path.exists():
                file_path.unlink()

    def create_build_script(self, version: str, release: str) -> str:
        """Create the Docker build script content."""
        script_content = f'''#!/bin/bash
set -e

# Colors for output
RED='\\033[0;31m'
GREEN='\\033[0;32m'
YELLOW='\\033[1;33m'
BLUE='\\033[0;34m'
NC='\\033[0m' # No Color

print_status() {{
    echo -e "${{BLUE}}[INFO]${{NC}} $1"
}}

print_success() {{
    echo -e "${{GREEN}}[SUCCESS]${{NC}} $1"
}}

print_error() {{
    echo -e "${{RED}}[ERROR]${{NC}} $1"
}}

cd /build

# Set up environment for build user
export HOME=/tmp
export GOPATH=/tmp/go
export GOCACHE=/tmp/go-cache
export GOMODCACHE=/tmp/go/pkg/mod
mkdir -p "$GOPATH" "$GOCACHE" "$GOMODCACHE"

VERSION="{version}"
RELEASE="{release}"

print_status "Inside Docker: Building OneMount v${{VERSION}}-${{RELEASE}}..."

# Create source tarball
print_status "Creating source tarball..."
mkdir -p "build/temp/onemount-${{VERSION}}"

# Copy source files
git ls-files > build/temp/filelist.txt
git rev-parse HEAD > build/temp/.commit
rsync -a --files-from=build/temp/filelist.txt . "build/temp/onemount-${{VERSION}}/"
# Copy the commit file separately since it's generated in build/temp
cp build/temp/.commit "build/temp/onemount-${{VERSION}}/"

# Move Ubuntu packaging (compatible with Debian)
mv "build/temp/onemount-${{VERSION}}/packaging/ubuntu" "build/temp/onemount-${{VERSION}}/debian"

# Create vendor directory
print_status "Creating Go vendor directory..."
go mod vendor
cp -R vendor/ "build/temp/onemount-${{VERSION}}/"

# Create tarballs
print_status "Creating source tarballs..."
cd build/temp && tar -czf "onemount_${{VERSION}}.orig.tar.gz" "onemount-${{VERSION}}"
cd /build

print_success "Source tarball created"

# Build source package
print_status "Building source package..."
cd "build/temp/onemount-${{VERSION}}"
dpkg-buildpackage -S -sa -d -us -uc
cd /build

# Move source package files to deb directory
mv build/temp/onemount_${{VERSION}}*.dsc build/packages/deb/
mv build/temp/onemount_${{VERSION}}*_source.* build/packages/deb/

print_success "Source package built"

# Build binary package
print_status "Building binary package..."
cd "build/temp/onemount-${{VERSION}}"
dpkg-buildpackage -b -d -us -uc
cd /build

# Move binary package files to deb directory
mv build/temp/onemount_${{VERSION}}*.deb build/packages/deb/
mv build/temp/onemount*_${{VERSION}}*.deb build/packages/deb/ 2>/dev/null || true
mv build/temp/onemount_${{VERSION}}*_amd64.* build/packages/deb/ 2>/dev/null || true
# Move source tarball to packages directory
mv build/temp/onemount_${{VERSION}}.orig.tar.gz build/packages/deb/ 2>/dev/null || true

print_success "Binary package built"

# Clean up build artifacts but keep packages
print_status "Cleaning up build artifacts..."
rm -rf build/temp/* vendor/

print_success "Docker build completed!"
print_status "Built packages:"
ls -la build/packages/deb/ 2>/dev/null || echo "No package files found"
'''
        return script_content

    def run_docker_build(self, image_name: str, version: str, release: str) -> bool:
        """Run the Docker build process."""
        self._log_info(f"Building OneMount v{version}-{release} Ubuntu package using Docker...")

        try:
            # Create build script
            script_content = self.create_build_script(version, release)
            script_path = self.paths["project_root"] / "docker-build-script.sh"

            with open(script_path, 'w') as f:
                f.write(script_content)
            script_path.chmod(0o755)

            # Get host user info for permission handling
            host_uid = os.getuid()
            host_gid = os.getgid()

            # Run Docker container
            self._log_info("Starting Docker build container...")

            with Progress(
                SpinnerColumn(),
                TextColumn("[progress.description]{task.description}"),
                console=console,
                transient=True,
            ) as progress:
                task = progress.add_task("Running Docker build...", total=None)

                try:
                    container = self.docker_client.containers.run(
                        image_name,
                        command="./docker-build-script.sh",
                        volumes={
                            str(self.paths["project_root"]): {'bind': '/build', 'mode': 'rw'}
                        },
                        working_dir="/build",
                        user=f"{host_uid}:{host_gid}",
                        environment={
                            'HOME': '/tmp',
                            'GOPATH': '/tmp/go',
                            'GOCACHE': '/tmp/go-cache',
                            'GOMODCACHE': '/tmp/go/pkg/mod',
                        },
                        remove=True,
                        detach=False,
                        stdout=True,
                        stderr=True,
                    )

                    # Get container output
                    output = container.decode('utf-8')

                    if self.verbose:
                        console.print("[dim]Docker build output:[/dim]")
                        console.print(output)

                    progress.update(task, description="✅ Docker build completed")

                except docker.errors.ContainerError as e:
                    progress.update(task, description="❌ Docker build failed")
                    self._log_error(f"Docker build failed: {e}")
                    if self.verbose and e.stderr:
                        console.print(f"[red]Error output:[/red]\n{e.stderr.decode('utf-8')}")
                    return False

                except Exception as e:
                    progress.update(task, description="❌ Docker build failed")
                    self._log_error(f"Unexpected error during Docker build: {e}")
                    return False

            # Clean up build script
            if script_path.exists():
                script_path.unlink()

            self._log_success("Docker-based Ubuntu package build completed!")
            return True

        except Exception as e:
            self._log_error(f"Failed to run Docker build: {e}")
            return False

    def show_build_results(self):
        """Show the results of the build process."""
        deb_dir = self.paths["deb_dir"]

        if not deb_dir.exists():
            self._log_warning("No package directory found")
            return

        deb_files = list(deb_dir.glob("*.deb"))

        if deb_files:
            self._log_info("Built packages:")
            for deb_file in deb_files:
                file_size = deb_file.stat().st_size / (1024 * 1024)  # MB
                console.print(f"  • {deb_file.name} ({file_size:.1f} MB)")
        else:
            self._log_warning("No package files found in build/packages/deb/")

    def build_debian_package(self, clean: bool = False, force_rebuild_image: bool = False) -> bool:
        """
        Main method to build Debian packages using Docker.

        Args:
            clean: Whether to clean before building
            force_rebuild_image: Whether to force rebuild of Docker image

        Returns:
            True if build succeeded, False otherwise
        """
        try:
            # Check Docker availability
            if not self.check_docker_available():
                return False

            # Get version information
            version, release = self.get_version_info()
            self._log_info(f"Building OneMount v{version}-{release}")

            # Prepare build environment
            if clean:
                self._log_info("Cleaning build artifacts...")
            self.prepare_build_directories()

            # Ensure Docker image
            image_name = self.ensure_docker_image(force_rebuild_image)

            # Run the build
            success = self.run_docker_build(image_name, version, release)

            if success:
                self.show_build_results()

            return success

        except DockerBuildError as e:
            self._log_error(str(e))
            return False
        except Exception as e:
            self._log_error(f"Unexpected error during build: {e}")
            return False


def build_debian_package_docker(verbose: bool = False, clean: bool = False, force_rebuild_image: bool = False) -> bool:
    """
    Convenience function to build Debian packages using Docker.

    Args:
        verbose: Enable verbose output
        clean: Clean before building
        force_rebuild_image: Force rebuild of Docker image

    Returns:
        True if build succeeded, False otherwise
    """
    with DockerPackageBuilder(verbose=verbose) as builder:
        return builder.build_debian_package(clean=clean, force_rebuild_image=force_rebuild_image)
