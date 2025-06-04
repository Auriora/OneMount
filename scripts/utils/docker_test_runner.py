"""
Native Python implementation for OneMount Docker test orchestration.
Replaces run-tests-docker.sh with native Python operations.
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


class DockerTestError(Exception):
    """Exception raised when Docker test operations fail."""
    pass


class DockerTestRunner:
    """Native Python Docker test runner for OneMount."""
    
    def __init__(self, verbose: bool = False):
        self.verbose = verbose
        self.paths = get_project_paths()
        
        # Configuration
        self.compose_file = "docker/compose/docker-compose.test.yml"
        self.test_image = "onemount-test-runner:latest"
        self.auth_tokens_path = Path.home() / ".onemount-tests" / ".auth_tokens.json"
        
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
    
    def check_docker(self) -> bool:
        """Check if Docker is available and running."""
        self._log_info("Checking Docker availability...")
        
        # Check if Docker is installed
        if not shutil.which("docker"):
            self._log_error("Docker is not installed or not in PATH")
            return False
        
        # Check if Docker daemon is running
        try:
            run_command(
                ["docker", "info"],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=10
            )
        except (CommandError, Exception):
            self._log_error("Docker daemon is not running or user lacks permissions")
            self._log_info("Try: sudo systemctl start docker")
            self._log_info("Or add user to docker group: sudo usermod -aG docker $USER")
            return False
        
        return True
    
    def check_docker_compose(self) -> bool:
        """Check if Docker Compose is available."""
        # Check for docker-compose command
        if shutil.which("docker-compose"):
            return True
        
        # Check for docker compose plugin
        try:
            run_command(
                ["docker", "compose", "version"],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=10
            )
            return True
        except (CommandError, Exception):
            self._log_error("Docker Compose is not available")
            self._log_info("Install docker-compose or use Docker with compose plugin")
            return False
    
    def get_compose_command(self) -> List[str]:
        """Get the appropriate Docker Compose command."""
        if shutil.which("docker-compose"):
            return ["docker-compose"]
        else:
            return ["docker", "compose"]
    
    def build_image(self, no_cache: bool = False) -> bool:
        """Build the Docker test image."""
        try:
            self._log_info("Building OneMount test Docker image...")
            
            cmd = ["docker", "build"]
            
            if no_cache:
                cmd.append("--no-cache")
            
            cmd.extend([
                "-f", "packaging/docker/Dockerfile.test-runner",
                "-t", self.test_image,
                "."
            ])
            
            # Enable BuildKit for better build performance
            env = os.environ.copy()
            env["DOCKER_BUILDKIT"] = "1"
            
            run_command(
                cmd,
                check=True,
                verbose=self.verbose,
                timeout=600,  # 10 minutes for image build
                cwd=str(self.paths["project_root"]),
                env=env
            )
            
            self._log_success("Docker image built successfully")
            return True
            
        except (CommandError, Exception) as e:
            self._log_error(f"Failed to build Docker image: {e}")
            return False
    
    def image_exists(self) -> bool:
        """Check if the test image exists."""
        try:
            run_command(
                ["docker", "image", "inspect", self.test_image],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=10
            )
            return True
        except (CommandError, Exception):
            return False
    
    def run_tests(
        self,
        service: str,
        rebuild: bool = False,
        timeout: Optional[str] = None,
        verbose: bool = False,
        sequential: bool = False
    ) -> bool:
        """Run tests using Docker Compose."""
        try:
            if not self.check_docker_compose():
                return False
            
            compose_cmd = self.get_compose_command()
            compose_file_path = self.paths["project_root"] / self.compose_file
            
            if not compose_file_path.exists():
                self._log_error(f"Docker Compose file not found: {compose_file_path}")
                return False
            
            # Check if auth tokens exist for system tests
            if service == "system-tests" and not self.auth_tokens_path.exists():
                self._log_error("OneDrive auth tokens not found for system tests")
                self._log_info("Run 'dev.py test docker setup-auth' for setup instructions")
                return False
            
            self._log_info(f"Running {service} with Docker Compose...")
            
            # Build image if it doesn't exist or if rebuild requested
            if rebuild or not self.image_exists():
                if not self.build_image():
                    return False
            
            # Set user ID and group ID for permission compatibility
            env = os.environ.copy()
            env["USER_ID"] = str(os.getuid())
            env["GROUP_ID"] = str(os.getgid())
            
            # Set test environment variables
            if timeout:
                env["ONEMOUNT_TEST_TIMEOUT"] = timeout
            if verbose:
                env["ONEMOUNT_TEST_VERBOSE"] = "true"
            if sequential:
                env["ONEMOUNT_TEST_SEQUENTIAL"] = "true"
            
            # Run the Docker Compose service
            cmd = compose_cmd + ["-f", str(compose_file_path), "run", "--rm", service]
            
            run_command(
                cmd,
                check=True,
                verbose=self.verbose,
                timeout=None,  # Use service's own timeout
                cwd=str(self.paths["project_root"]),
                env=env
            )
            
            self._log_success(f"{service} completed successfully")
            return True
            
        except (CommandError, Exception) as e:
            self._log_error(f"{service} failed: {e}")
            return False
    
    def clean_docker(self) -> bool:
        """Clean up Docker resources."""
        try:
            self._log_info("Cleaning up OneMount Docker test resources...")
            
            compose_cmd = self.get_compose_command()
            compose_file_path = self.paths["project_root"] / self.compose_file
            
            # Stop and remove containers
            if compose_file_path.exists():
                try:
                    run_command(
                        compose_cmd + ["-f", str(compose_file_path), "down", "--remove-orphans"],
                        capture_output=True,
                        check=False,  # Don't fail if nothing to clean
                        verbose=False,
                        timeout=30
                    )
                except Exception:
                    pass  # Ignore cleanup errors
            
            # Remove test containers
            try:
                result = run_command(
                    ["docker", "ps", "-a", "--filter", "name=onemount-", "--format", "{{.Names}}"],
                    capture_output=True,
                    check=False,
                    verbose=False,
                    timeout=30
                )
                
                if result.stdout.strip():
                    container_names = result.stdout.strip().split('\n')
                    for name in container_names:
                        if name.strip():
                            try:
                                run_command(
                                    ["docker", "rm", "-f", name.strip()],
                                    capture_output=True,
                                    check=False,
                                    verbose=False,
                                    timeout=30
                                )
                            except Exception:
                                pass  # Ignore individual container cleanup errors
            except Exception:
                pass  # Ignore container listing errors
            
            # Remove test image
            try:
                run_command(
                    ["docker", "rmi", self.test_image],
                    capture_output=True,
                    check=False,
                    verbose=False,
                    timeout=30
                )
            except Exception:
                pass  # Ignore image removal errors
            
            # Clean up test artifacts
            test_artifacts_dir = self.paths["project_root"] / "test-artifacts"
            if test_artifacts_dir.exists():
                self._log_info("Cleaning up test artifacts...")
                shutil.rmtree(test_artifacts_dir)
            
            self._log_success("Docker cleanup complete")
            return True
            
        except Exception as e:
            self._log_warning(f"Some cleanup operations failed: {e}")
            return True  # Don't fail the overall operation for cleanup issues
    
    def show_auth_setup_help(self):
        """Show authentication setup help."""
        console.print()
        console.print("[blue]Setting up OneDrive Authentication for System Tests[/blue]")
        console.print()
        console.print("System tests require valid OneDrive authentication tokens. Follow these steps:")
        console.print()
        console.print("1. Build OneMount:")
        console.print("   [yellow]make onemount[/yellow]")
        console.print()
        console.print("2. Authenticate with your test OneDrive account:")
        console.print("   [yellow]./build/onemount --auth-only[/yellow]")
        console.print()
        console.print("3. Create test directory and copy tokens:")
        console.print("   [yellow]mkdir -p ~/.onemount-tests[/yellow]")
        console.print("   [yellow]cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json[/yellow]")
        console.print()
        console.print("4. Verify the tokens file exists:")
        console.print("   [yellow]ls -la ~/.onemount-tests/.auth_tokens.json[/yellow]")
        console.print()
        console.print("5. Now you can run system tests:")
        console.print("   [yellow]dev.py test docker system[/yellow]")
        console.print()
        console.print("[yellow]Important Notes:[/yellow]")
        console.print("- Use a dedicated test OneDrive account, not your production account")
        console.print("- The auth tokens file will be mounted into the Docker container")
        console.print("- System tests create and delete files in /onemount_system_tests/ on OneDrive")
        console.print()

    def run_docker_tests(
        self,
        test_type: str = "unit",
        rebuild: bool = False,
        timeout: Optional[str] = None,
        verbose: bool = False,
        sequential: bool = False,
        clean: bool = False
    ) -> bool:
        """
        Main method to run Docker tests.

        Args:
            test_type: Type of tests to run (unit/integration/system/all/coverage)
            rebuild: Force rebuild of Docker image
            timeout: Test timeout duration
            verbose: Enable verbose output
            sequential: Run tests sequentially
            clean: Clean up Docker resources after tests

        Returns:
            True if tests succeeded, False otherwise
        """
        try:
            self._log_info("OneMount Docker Test Runner")
            self._log_info(f"Test type: {test_type}")
            if timeout:
                self._log_info(f"Timeout: {timeout}")
            console.print()

            # Check prerequisites
            if not self.check_docker():
                return False

            # Map test types to Docker Compose services
            service_map = {
                "unit": "unit-tests",
                "integration": "integration-tests",
                "system": "system-tests",
                "all": "test-runner",  # Will run all tests
                "coverage": "coverage",
                "shell": "shell"
            }

            if test_type not in service_map:
                self._log_error(f"Unknown test type: {test_type}")
                self._log_info(f"Valid types: {', '.join(service_map.keys())}")
                return False

            service = service_map[test_type]

            # Special handling for auth setup help
            if test_type == "system" and not self.auth_tokens_path.exists():
                self.show_auth_setup_help()
                return False

            # Run the tests
            success = self.run_tests(
                service=service,
                rebuild=rebuild,
                timeout=timeout,
                verbose=verbose,
                sequential=sequential
            )

            # Clean up if requested
            if clean:
                self.clean_docker()

            console.print()
            if success:
                self._log_success("Docker tests completed successfully!")
            else:
                self._log_error("Docker tests failed!")

            return success

        except DockerTestError as e:
            self._log_error(str(e))
            return False
        except Exception as e:
            self._log_error(f"Unexpected error during Docker tests: {e}")
            return False


def run_docker_tests(
    test_type: str = "unit",
    rebuild: bool = False,
    timeout: Optional[str] = None,
    verbose: bool = False,
    sequential: bool = False,
    clean: bool = False
) -> bool:
    """
    Convenience function to run Docker tests.

    Args:
        test_type: Type of tests to run (unit/integration/system/all/coverage)
        rebuild: Force rebuild of Docker image
        timeout: Test timeout duration
        verbose: Enable verbose output
        sequential: Run tests sequentially
        clean: Clean up Docker resources after tests

    Returns:
        True if tests succeeded, False otherwise
    """
    runner = DockerTestRunner(verbose=verbose)
    return runner.run_docker_tests(
        test_type=test_type,
        rebuild=rebuild,
        timeout=timeout,
        verbose=verbose,
        sequential=sequential,
        clean=clean
    )


def build_docker_image(no_cache: bool = False, verbose: bool = False) -> bool:
    """
    Convenience function to build Docker test image.

    Args:
        no_cache: Force rebuild without cache
        verbose: Enable verbose output

    Returns:
        True if build succeeded, False otherwise
    """
    runner = DockerTestRunner(verbose=verbose)
    return runner.build_image(no_cache=no_cache)


def clean_docker_resources(verbose: bool = False) -> bool:
    """
    Convenience function to clean Docker test resources.

    Args:
        verbose: Enable verbose output

    Returns:
        True if cleanup succeeded, False otherwise
    """
    runner = DockerTestRunner(verbose=verbose)
    return runner.clean_docker()


def show_docker_auth_help():
    """Show Docker authentication setup help."""
    runner = DockerTestRunner()
    runner.show_auth_setup_help()
