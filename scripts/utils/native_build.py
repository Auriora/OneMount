"""
Native Python implementation for native Debian package building.
Replaces build-deb-native.sh with native Python operations.
"""

import os
import shutil
import subprocess
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn

from .paths import get_project_paths
from .shell import run_command, CommandError

console = Console()


class NativeBuildError(Exception):
    """Exception raised when native build operations fail."""
    pass


class NativePackageBuilder:
    """Native Python Debian package builder for OneMount."""
    
    def __init__(self, verbose: bool = False):
        self.verbose = verbose
        self.paths = get_project_paths()
        
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
    
    def check_prerequisites(self) -> bool:
        """Check if all prerequisites for native building are met."""
        # Check if we're in the right directory
        makefile = self.paths["project_root"] / "Makefile"
        deb_packaging = self.paths["project_root"] / "packaging" / "deb"
        
        if not makefile.exists():
            self._log_error("This script must be run from the OneMount project root directory")
            self._log_error(f"Missing: {makefile}")
            return False
        
        if not deb_packaging.exists():
            self._log_error("This script must be run from the OneMount project root directory")
            self._log_error(f"Missing: {deb_packaging}")
            return False
        
        return True
    
    def check_build_tools(self) -> bool:
        """Check for required build tools."""
        self._log_info("Checking for required build tools...")
        
        required_tools = ["dpkg-buildpackage", "debuild", "go", "git", "rsync"]
        missing_tools = []
        
        for tool in required_tools:
            if not shutil.which(tool):
                missing_tools.append(tool)
        
        if missing_tools:
            self._log_error(f"Missing required tools: {', '.join(missing_tools)}")
            self._log_info("Please install them with:")
            console.print("sudo apt install build-essential debhelper devscripts dpkg-dev golang git rsync")
            return False
        
        return True
    
    def get_version_info(self) -> Tuple[str, str]:
        """Extract version information from spec file."""
        spec_file = self.paths["project_root"] / "packaging" / "rpm" / "onemount.spec"
        
        if not spec_file.exists():
            raise NativeBuildError(f"Spec file not found: {spec_file}")
        
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
            raise NativeBuildError("Could not extract version/release from spec file")
        
        return version, release
    
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
        
        # Clean version-specific directories
        version, _ = self.get_version_info()
        version_dir = project_root / f"onemount-{version}"
        if version_dir.exists():
            shutil.rmtree(version_dir)
        
        # Clean vendor directory
        vendor_dir = project_root / "vendor"
        if vendor_dir.exists():
            shutil.rmtree(vendor_dir)
    
    def create_source_tarball(self, version: str) -> bool:
        """Create source tarball for package building."""
        try:
            self._log_info("Creating source tarball...")
            
            # Create version-specific directory
            version_dir = self.paths["temp_dir"] / f"onemount-{version}"
            version_dir.mkdir(parents=True, exist_ok=True)
            
            # Get list of files from git
            result = run_command(
                ["git", "ls-files"],
                capture_output=True,
                check=True,
                verbose=self.verbose,
                timeout=30
            )

            filelist_path = self.paths["temp_dir"] / "filelist.txt"
            with open(filelist_path, 'w') as f:
                f.write(result.stdout)

            # Get current commit
            result = run_command(
                ["git", "rev-parse", "HEAD"],
                capture_output=True,
                check=True,
                verbose=self.verbose,
                timeout=10
            )

            # Create commit file in project root first
            commit_path = self.paths["project_root"] / ".commit"
            with open(commit_path, 'w') as f:
                f.write(result.stdout.strip())

            # Add commit file to filelist
            with open(filelist_path, 'a') as f:
                f.write(".commit\n")

            # Copy source files using rsync
            run_command(
                ["rsync", "-a", f"--files-from={filelist_path}", ".", str(version_dir) + "/"],
                check=True,
                verbose=self.verbose,
                timeout=120
            )
            
            # Move debian packaging
            deb_source = version_dir / "packaging" / "deb"
            deb_dest = version_dir / "debian"
            if deb_source.exists():
                shutil.move(str(deb_source), str(deb_dest))
            
            return True
            
        except (CommandError, Exception) as e:
            self._log_error(f"Failed to create source tarball: {e}")
            return False
    
    def create_vendor_directory(self, version: str) -> bool:
        """Create Go vendor directory."""
        try:
            self._log_info("Creating Go vendor directory...")
            
            # Create vendor directory
            run_command(
                ["go", "mod", "vendor"],
                check=True,
                verbose=self.verbose,
                timeout=300  # 5 minutes for vendor creation
            )
            
            # Copy vendor directory to version directory
            vendor_source = self.paths["project_root"] / "vendor"
            version_dir = self.paths["temp_dir"] / f"onemount-{version}"
            vendor_dest = version_dir / "vendor"
            
            if vendor_source.exists():
                shutil.copytree(vendor_source, vendor_dest)
            
            return True
            
        except (CommandError, Exception) as e:
            self._log_error(f"Failed to create vendor directory: {e}")
            return False

    def create_source_package_tarball(self, version: str) -> bool:
        """Create source package tarball."""
        try:
            self._log_info("Creating source tarballs...")

            # Create tarball
            version_dir = self.paths["temp_dir"] / f"onemount-{version}"

            # For dpkg-buildpackage, the tarball needs to be in the parent directory of the source
            # So we put it in temp_dir, and also copy to deb_dir for final storage
            temp_tarball_path = self.paths["temp_dir"] / f"onemount_{version}.orig.tar.gz"
            final_tarball_path = self.paths["deb_dir"] / f"onemount_{version}.orig.tar.gz"

            # Verify the version directory exists
            if not version_dir.exists():
                self._log_error(f"Version directory does not exist: {version_dir}")
                return False

            # Create tarball using absolute paths (avoid directory change issues)
            run_command(
                ["tar", "-czf", str(temp_tarball_path), "-C", str(self.paths["temp_dir"]), f"onemount-{version}"],
                check=True,
                verbose=self.verbose,
                timeout=120
            )

            # Copy to final location for storage
            shutil.copy2(temp_tarball_path, final_tarball_path)

            self._log_success("Source tarball created")
            return True

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to create source tarball: {e}")
            return False

    def build_source_package(self, version: str) -> bool:
        """Build Debian source package."""
        try:
            self._log_info("Building source package...")

            version_dir = self.paths["temp_dir"] / f"onemount-{version}"

            # Build source package in the version directory
            run_command(
                ["dpkg-buildpackage", "-S", "-sa", "-d", "-us", "-uc"],
                check=True,
                verbose=self.verbose,
                timeout=300,  # 5 minutes
                cwd=str(version_dir)
            )

            # Move source package files to deb directory
            temp_dir = self.paths["temp_dir"]
            deb_dir = self.paths["deb_dir"]

            for pattern in [f"onemount_{version}*.dsc", f"onemount_{version}*_source.*"]:
                for file_path in temp_dir.glob(pattern):
                    dest_path = deb_dir / file_path.name
                    shutil.move(str(file_path), str(dest_path))

            self._log_success("Source package built")
            return True

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to build source package: {e}")
            return False

    def build_binary_package(self, version: str) -> bool:
        """Build Debian binary package."""
        try:
            self._log_info("Building binary package...")

            version_dir = self.paths["temp_dir"] / f"onemount-{version}"

            # Build binary package in the version directory
            run_command(
                ["dpkg-buildpackage", "-b", "-d", "-us", "-uc"],
                check=True,
                verbose=self.verbose,
                timeout=600,  # 10 minutes for binary build
                cwd=str(version_dir)
            )

            # Move binary package files to deb directory
            temp_dir = self.paths["temp_dir"]
            deb_dir = self.paths["deb_dir"]

            # Move .deb files
            for pattern in [f"onemount_{version}*.deb", f"onemount*_{version}*.deb", f"onemount_{version}*_amd64.*"]:
                for file_path in temp_dir.glob(pattern):
                    dest_path = deb_dir / file_path.name
                    try:
                        shutil.move(str(file_path), str(dest_path))
                    except FileNotFoundError:
                        # Some patterns might not match, that's okay
                        pass

            self._log_success("Binary package built")
            return True

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to build binary package: {e}")
            return False

    def cleanup_build_artifacts(self):
        """Clean up build artifacts but keep packages."""
        try:
            self._log_info("Cleaning up build artifacts...")

            # Clean temp directory
            if self.paths["temp_dir"].exists():
                shutil.rmtree(self.paths["temp_dir"])
                self.paths["temp_dir"].mkdir(parents=True)

            # Clean vendor directory
            vendor_dir = self.paths["project_root"] / "vendor"
            if vendor_dir.exists():
                shutil.rmtree(vendor_dir)

        except Exception as e:
            self._log_warning(f"Could not clean up all build artifacts: {e}")

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

    def build_debian_package(self, clean: bool = False) -> bool:
        """
        Main method to build Debian packages natively.

        Args:
            clean: Whether to clean before building

        Returns:
            True if build succeeded, False otherwise
        """
        try:
            # Check prerequisites
            if not self.check_prerequisites():
                return False

            if not self.check_build_tools():
                return False

            # Get version information
            version, release = self.get_version_info()
            self._log_info(f"Building OneMount v{version}-{release} Debian package natively...")

            # Prepare build environment
            if clean:
                self._log_info("Cleaning build artifacts...")
            self.prepare_build_directories()

            # Create source tarball
            if not self.create_source_tarball(version):
                return False

            # Create vendor directory
            if not self.create_vendor_directory(version):
                return False

            # Create source package tarball
            if not self.create_source_package_tarball(version):
                return False

            # Build source package
            if not self.build_source_package(version):
                return False

            # Build binary package
            if not self.build_binary_package(version):
                return False

            # Clean up build artifacts
            self.cleanup_build_artifacts()

            # Show results
            self.show_build_results()

            self._log_success("Native Debian package build completed!")
            return True

        except NativeBuildError as e:
            self._log_error(str(e))
            return False
        except Exception as e:
            self._log_error(f"Unexpected error during build: {e}")
            return False


def build_debian_package_native(verbose: bool = False, clean: bool = False) -> bool:
    """
    Convenience function to build Debian packages natively.

    Args:
        verbose: Enable verbose output
        clean: Clean before building

    Returns:
        True if build succeeded, False otherwise
    """
    builder = NativePackageBuilder(verbose=verbose)
    return builder.build_debian_package(clean=clean)
