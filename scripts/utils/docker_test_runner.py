"""
Native Python implementation for OneMount Docker test orchestration.
Replaces run-tests-docker.sh with native Python operations.
"""

import os
import shutil
import subprocess
import hashlib
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
        self.build_compose_file = "docker/compose/docker-compose.build.yml"
        self.base_image_name = "onemount-test-runner"
        self.auth_tokens_path = Path.home() / ".onemount-tests" / ".auth_tokens.json"

        # Image and container naming
        self.image_tag = self._get_image_tag()
        self.test_image = f"{self.base_image_name}:{self.image_tag}"
        self.dev_image = f"{self.base_image_name}:dev"

    def _get_image_tag(self) -> str:
        """Generate image tag based on git commit and dockerfile hash."""
        try:
            # Get git commit hash
            result = run_command(
                ["git", "rev-parse", "--short", "HEAD"],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=10,
                cwd=str(self.paths["project_root"])
            )
            git_hash = result.stdout.strip()

            # Get dockerfile hash for cache invalidation
            dockerfile_path = self.paths["project_root"] / "packaging/docker/Dockerfile.test-runner"
            if dockerfile_path.exists():
                with open(dockerfile_path, 'rb') as f:
                    dockerfile_hash = hashlib.md5(f.read()).hexdigest()[:8]
            else:
                dockerfile_hash = "unknown"

            return f"{git_hash}-{dockerfile_hash}"

        except Exception:
            # Fallback to timestamp-based tag
            import time
            return f"dev-{int(time.time())}"
        
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
    
    def build_image(self, no_cache: bool = False, tag: str = None, development: bool = False, use_compose: bool = False) -> bool:
        """Build the Docker test image with improved caching and tagging."""
        try:
            target_tag = tag or (self.dev_image if development else self.test_image)
            self._log_info(f"Building OneMount test Docker image: {target_tag}")

            if use_compose and self.check_docker_compose():
                return self._build_with_compose(no_cache, target_tag, development)
            else:
                return self._build_with_docker(no_cache, target_tag, development)

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to build Docker image: {e}")
            return False

    def _build_with_compose(self, no_cache: bool, target_tag: str, development: bool) -> bool:
        """Build image using Docker Compose."""
        try:
            compose_cmd = self.get_compose_command()
            build_compose_file = self.paths["project_root"] / self.build_compose_file

            if not build_compose_file.exists():
                self._log_warning("Build compose file not found, falling back to direct Docker build")
                return self._build_with_docker(no_cache, target_tag, development)

            # Set environment variables for the build
            env = os.environ.copy()
            env["DOCKER_BUILDKIT"] = "1"
            env["ONEMOUNT_TEST_IMAGE"] = target_tag

            # Determine which profile to use
            if no_cache:
                profile = "build-no-cache"
                service = "test-runner-no-cache-build"
            elif development:
                profile = "build-dev"
                service = "test-runner-dev-build"
            else:
                profile = "build"
                service = "test-runner-build"

            # Build using compose
            cmd = compose_cmd + [
                "-f", str(build_compose_file),
                "--profile", profile,
                "build", service
            ]

            run_command(
                cmd,
                check=True,
                verbose=self.verbose,
                timeout=900,  # 15 minutes for image build
                cwd=str(self.paths["project_root"]),
                env=env
            )

            # Add additional tags if needed
            self._add_additional_tags(target_tag, development)

            self._log_success(f"Docker image built successfully with Compose: {target_tag}")
            return True

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to build with Compose: {e}")
            return False

    def _build_with_docker(self, no_cache: bool, target_tag: str, development: bool) -> bool:
        """Build image using direct Docker commands."""
        try:
            cmd = ["docker", "build"]

            if no_cache:
                cmd.append("--no-cache")

            # Add primary tag
            cmd.extend([
                "-f", "packaging/docker/Dockerfile.test-runner",
                "-t", target_tag
            ])

            cmd.append(".")

            # Enable BuildKit for better build performance and caching
            env = os.environ.copy()
            env["DOCKER_BUILDKIT"] = "1"

            run_command(
                cmd,
                check=True,
                verbose=self.verbose,
                timeout=900,  # 15 minutes for image build
                cwd=str(self.paths["project_root"]),
                env=env
            )

            # Add additional tags
            self._add_additional_tags(target_tag, development)

            self._log_success(f"Docker image built successfully: {target_tag}")
            return True

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to build with Docker: {e}")
            return False

    def _add_additional_tags(self, target_tag: str, development: bool):
        """Add additional tags to the built image."""
        try:
            additional_tags = []

            # Add latest tag for non-development builds
            if not development and target_tag != f"{self.base_image_name}:latest":
                additional_tags.append(f"{self.base_image_name}:latest")

            # Add git-based tag if available
            try:
                result = run_command(
                    ["git", "rev-parse", "--short", "HEAD"],
                    capture_output=True,
                    check=True,
                    verbose=False,
                    timeout=10,
                    cwd=str(self.paths["project_root"])
                )
                git_tag = f"{self.base_image_name}:git-{result.stdout.strip()}"
                if git_tag != target_tag:
                    additional_tags.append(git_tag)
            except Exception:
                pass  # Git tag is optional

            # Apply additional tags
            for tag in additional_tags:
                try:
                    run_command(
                        ["docker", "tag", target_tag, tag],
                        capture_output=True,
                        check=True,
                        verbose=False,
                        timeout=30
                    )
                    self._log_info(f"Added tag: {tag}")
                except Exception:
                    self._log_warning(f"Failed to add tag: {tag}")

        except Exception:
            pass  # Additional tagging is optional
    
    def image_exists(self, tag: str = None) -> bool:
        """Check if the test image exists."""
        image_name = tag or self.test_image
        try:
            run_command(
                ["docker", "image", "inspect", image_name],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=10
            )
            return True
        except (CommandError, Exception):
            return False

    def container_exists(self, name: str) -> bool:
        """Check if a container exists."""
        try:
            run_command(
                ["docker", "container", "inspect", name],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=10
            )
            return True
        except (CommandError, Exception):
            return False

    def container_running(self, name: str) -> bool:
        """Check if a container is running."""
        try:
            result = run_command(
                ["docker", "container", "inspect", name, "--format", "{{.State.Running}}"],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=10
            )
            return result.stdout.strip().lower() == "true"
        except (CommandError, Exception):
            return False

    def get_container_name(self, test_type: str, development: bool = False) -> str:
        """Generate container name based on test type and mode."""
        suffix = "dev" if development else "test"
        return f"onemount-{test_type}-{suffix}"
    
    def run_tests(
        self,
        service: str,
        test_type: str = "unit",
        rebuild_image: bool = False,
        recreate_container: bool = False,
        reuse_container: bool = True,
        timeout: Optional[str] = None,
        verbose: bool = False,
        sequential: bool = False,
        development: bool = False
    ) -> bool:
        """Run tests using Docker with enhanced container management."""
        try:
            if not self.check_docker():
                return False

            # Check if auth tokens exist for system tests
            if service == "system-tests" and not self.auth_tokens_path.exists():
                self._log_error("OneDrive auth tokens not found for system tests")
                self._log_info("Run 'dev.py test docker setup-auth' for setup instructions")
                return False

            # Determine image and container strategy
            target_image = self.dev_image if development else self.test_image
            container_name = self.get_container_name(test_type, development)

            self._log_info(f"Running {test_type} tests...")
            self._log_info(f"Image: {target_image}")
            self._log_info(f"Container: {container_name}")

            # Build image if needed
            if rebuild_image or not self.image_exists(target_image):
                if not self.build_image(development=development):
                    return False

            # Handle container lifecycle
            if reuse_container and self.container_exists(container_name):
                if recreate_container:
                    self._log_info(f"Recreating container: {container_name}")
                    self._remove_container(container_name)
                elif self.container_running(container_name):
                    self._log_info(f"Reusing running container: {container_name}")
                    return self._exec_in_container(container_name, test_type, verbose, timeout, sequential)
                else:
                    self._log_info(f"Starting existing container: {container_name}")
                    return self._start_and_exec_container(container_name, test_type, verbose, timeout, sequential)

            # Create and run new container
            return self._run_new_container(
                target_image, container_name, test_type,
                verbose, timeout, sequential, reuse_container, development
            )
        except (CommandError, Exception) as e:
            self._log_error(f"Test execution failed: {e}")
            return False

    def _remove_container(self, name: str):
        """Remove a container."""
        try:
            run_command(
                ["docker", "rm", "-f", name],
                capture_output=True,
                check=False,
                verbose=False,
                timeout=30
            )
        except Exception:
            pass  # Ignore removal errors

    def _exec_in_container(self, name: str, test_type: str, verbose: bool, timeout: Optional[str], sequential: bool) -> bool:
        """Execute tests in an existing running container."""
        try:
            cmd = ["docker", "exec", name, "/usr/local/bin/test-entrypoint.sh", test_type]

            if verbose:
                cmd.append("--verbose")
            if timeout:
                cmd.extend(["--timeout", timeout])
            if sequential:
                cmd.append("--sequential")

            run_command(
                cmd,
                check=True,
                verbose=self.verbose,
                timeout=None,
                cwd=str(self.paths["project_root"])
            )

            self._log_success(f"Tests completed successfully in container: {name}")
            return True

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to execute tests in container {name}: {e}")
            return False

    def _start_and_exec_container(self, name: str, test_type: str, verbose: bool, timeout: Optional[str], sequential: bool) -> bool:
        """Start an existing container and execute tests."""
        try:
            # Start the container
            run_command(
                ["docker", "start", name],
                capture_output=True,
                check=True,
                verbose=False,
                timeout=30
            )

            # Execute tests
            return self._exec_in_container(name, test_type, verbose, timeout, sequential)

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to start container {name}: {e}")
            return False

    def _run_new_container(
        self,
        image: str,
        name: str,
        test_type: str,
        verbose: bool,
        timeout: Optional[str],
        sequential: bool,
        reuse: bool,
        development: bool
    ) -> bool:
        """Run tests in a new container."""
        try:
            cmd = ["docker", "run"]

            # Container lifecycle management
            if reuse:
                cmd.extend(["--name", name])
            else:
                cmd.append("--rm")

            # Volume mounts
            cmd.extend([
                "-v", f"{self.paths['project_root']}:/workspace:rw",
                "-v", f"{self.paths['project_root']}/test-artifacts:/home/tester/.onemount-tests:rw"
            ])

            # Copy auth tokens to test-artifacts if available
            if self.auth_tokens_path.exists():
                test_artifacts_dir = self.paths['project_root'] / "test-artifacts"
                test_artifacts_dir.mkdir(exist_ok=True)
                auth_tokens_dest = test_artifacts_dir / ".auth_tokens.json"
                if not auth_tokens_dest.exists():
                    import shutil
                    shutil.copy2(self.auth_tokens_path, auth_tokens_dest)
                    self._log_info(f"Copied auth tokens to {auth_tokens_dest}")

            # FUSE support for filesystem testing
            cmd.extend([
                "--device", "/dev/fuse:/dev/fuse",
                "--cap-add", "SYS_ADMIN",
                "--security-opt", "apparmor:unconfined"
            ])

            # User permissions
            cmd.extend(["--user", f"{os.getuid()}:{os.getgid()}"])

            # Environment variables
            env_vars = []
            if timeout:
                env_vars.append(f"ONEMOUNT_TEST_TIMEOUT={timeout}")
            if verbose:
                env_vars.append("ONEMOUNT_TEST_VERBOSE=true")
            if sequential:
                env_vars.append("ONEMOUNT_TEST_SEQUENTIAL=true")

            for env_var in env_vars:
                cmd.extend(["-e", env_var])

            # Interactive mode for development
            if development:
                cmd.extend(["-it"])

            # Image and command
            cmd.append(image)
            cmd.append(test_type)

            # Add test options
            if verbose:
                cmd.append("--verbose")
            if timeout:
                cmd.extend(["--timeout", timeout])
            if sequential:
                cmd.append("--sequential")

            run_command(
                cmd,
                check=True,
                verbose=self.verbose,
                timeout=None,
                cwd=str(self.paths["project_root"])
            )

            self._log_success(f"Tests completed successfully in new container: {name}")
            return True

        except (CommandError, Exception) as e:
            self._log_error(f"Failed to run new container {name}: {e}")
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
        rebuild_image: bool = False,
        recreate_container: bool = False,
        reuse_container: bool = True,
        timeout: Optional[str] = None,
        verbose: bool = False,
        sequential: bool = False,
        clean: bool = False,
        development: bool = False
    ) -> bool:
        """
        Main method to run Docker tests with enhanced container management.

        Args:
            test_type: Type of tests to run (unit/integration/system/all/coverage)
            rebuild_image: Force rebuild of Docker image
            recreate_container: Force recreation of container
            reuse_container: Reuse existing containers (default: True)
            timeout: Test timeout duration
            verbose: Enable verbose output
            sequential: Run tests sequentially
            clean: Clean up Docker resources after tests
            development: Use development mode with persistent containers

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
                test_type=test_type,
                rebuild_image=rebuild_image,
                recreate_container=recreate_container,
                reuse_container=reuse_container,
                timeout=timeout,
                verbose=verbose,
                sequential=sequential,
                development=development
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
    rebuild_image: bool = False,
    recreate_container: bool = False,
    reuse_container: bool = True,
    timeout: Optional[str] = None,
    verbose: bool = False,
    sequential: bool = False,
    clean: bool = False,
    development: bool = False
) -> bool:
    """
    Convenience function to run Docker tests with enhanced container management.

    Args:
        test_type: Type of tests to run (unit/integration/system/all/coverage)
        rebuild_image: Force rebuild of Docker image
        recreate_container: Force recreation of container
        reuse_container: Reuse existing containers (default: True)
        timeout: Test timeout duration
        verbose: Enable verbose output
        sequential: Run tests sequentially
        clean: Clean up Docker resources after tests
        development: Use development mode with persistent containers

    Returns:
        True if tests succeeded, False otherwise
    """
    runner = DockerTestRunner(verbose=verbose)
    return runner.run_docker_tests(
        test_type=test_type,
        rebuild_image=rebuild_image,
        recreate_container=recreate_container,
        reuse_container=reuse_container,
        timeout=timeout,
        verbose=verbose,
        sequential=sequential,
        clean=clean,
        development=development
    )


def build_docker_image(no_cache: bool = False, tag: str = None, development: bool = False, use_compose: bool = False, verbose: bool = False) -> bool:
    """
    Convenience function to build Docker test image with enhanced tagging.

    Args:
        no_cache: Force rebuild without cache
        tag: Custom tag for the image
        development: Build development image
        use_compose: Use Docker Compose for building (default: True)
        verbose: Enable verbose output

    Returns:
        True if build succeeded, False otherwise
    """
    runner = DockerTestRunner(verbose=verbose)
    return runner.build_image(no_cache=no_cache, tag=tag, development=development, use_compose=use_compose)


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
