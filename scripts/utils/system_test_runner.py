"""
Native Python implementation for OneMount system test runner.
Replaces run-system-tests.sh with native Python operations.
"""

import json
import os
import shutil
import subprocess
import time
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn

from .paths import get_project_paths
from .shell import run_command, CommandError

console = Console()


class SystemTestError(Exception):
    """Exception raised when system test operations fail."""
    pass


class SystemTestRunner:
    """Native Python system test runner for OneMount."""
    
    def __init__(self, verbose: bool = False):
        self.verbose = verbose
        self.paths = get_project_paths()
        
        # Configuration
        self.auth_tokens_path = Path.home() / ".onemount-tests" / ".auth_tokens.json"
        self.test_log_path = Path.home() / ".onemount-tests" / "logs" / "system_tests.log"
        self.timeout = "30m"
        
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
        """Check if all prerequisites for system testing are met."""
        self._log_info("Checking prerequisites...")
        
        # Check if auth tokens exist
        if not self.auth_tokens_path.exists():
            self._log_error(f"Authentication tokens not found at {self.auth_tokens_path}")
            
            # Provide different instructions based on environment
            if os.getenv("CI") or os.getenv("GITHUB_ACTIONS"):
                self._log_error("In CI environment, ensure secrets are properly configured:")
                self._log_error("  - For service principal: AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID")
                self._log_error("  - For test account: ONEDRIVE_TEST_TOKENS")
                self._log_error("See docs/testing/ci-system-tests-setup.md for details")
            else:
                self._log_error("Please run OneMount authentication first:")
                self._log_error("  make onemount")
                self._log_error("  ./build/onemount --auth-only")
                self._log_error("  mkdir -p ~/.onemount-tests")
                self._log_error("  cp ~/.cache/onemount/auth_tokens.json ~/.onemount-tests/.auth_tokens.json")
            return False
        
        # Check if Go is installed
        if not shutil.which("go"):
            self._log_error("Go is not installed or not in PATH")
            return False
        
        # Check if we're in the OneMount project directory
        go_mod = self.paths["project_root"] / "go.mod"
        if not go_mod.exists():
            self._log_error("This script must be run from the OneMount project root directory")
            return False
        
        # Check go.mod content
        try:
            with open(go_mod, 'r') as f:
                content = f.read()
                if "github.com/auriora/onemount" not in content:
                    self._log_error("This script must be run from the OneMount project root directory")
                    return False
        except Exception as e:
            self._log_error(f"Could not read go.mod: {e}")
            return False
        
        # Create log directory if it doesn't exist
        self.test_log_path.parent.mkdir(parents=True, exist_ok=True)
        
        # Check for CI environment and provide additional info
        if os.getenv("CI") or os.getenv("GITHUB_ACTIONS"):
            self._log_info("Running in CI environment")
            
            # Validate token file format
            try:
                with open(self.auth_tokens_path, 'r') as f:
                    json.load(f)
            except json.JSONDecodeError:
                self._log_error("Auth tokens file is not valid JSON")
                return False
            except Exception as e:
                self._log_error(f"Could not read auth tokens file: {e}")
                return False
            
            # Check token expiration
            try:
                with open(self.auth_tokens_path, 'r') as f:
                    tokens = json.load(f)
                    expires_at = tokens.get('expires_at', 0)
                    current_time = int(time.time())
                    
                    if expires_at <= current_time:
                        self._log_warning(f"Auth tokens appear to be expired (expires_at: {expires_at}, current: {current_time})")
                        self._log_warning("Tests may fail due to expired tokens")
                    else:
                        self._log_info(f"Auth tokens are valid (expires in {expires_at - current_time} seconds)")
            except Exception as e:
                self._log_warning(f"Could not check token expiration: {e}")
        
        self._log_success("Prerequisites check passed")
        return True
    
    def run_test_category(self, category: str, test_pattern: Optional[str] = None) -> bool:
        """Run specific test category."""
        try:
            self._log_info(f"Running {category} tests...")
            
            # Build the go test command
            cmd = ["go", "test", "-v", "-timeout", self.timeout]
            
            if test_pattern:
                cmd.extend(["-run", test_pattern])
            
            cmd.append("./tests/system")
            
            # Run the tests
            run_command(
                cmd,
                check=True,
                verbose=self.verbose,
                timeout=None,  # Use Go's own timeout
                cwd=str(self.paths["project_root"])
            )
            
            self._log_success(f"{category} tests completed successfully")
            return True
            
        except (CommandError, Exception) as e:
            self._log_error(f"{category} tests failed: {e}")
            return False
    
    def run_all_tests(self) -> bool:
        """Run all system test categories."""
        failed_tests = []
        
        self._log_info("Running all system test categories...")
        
        # Define test categories and their patterns
        test_categories = [
            ("Comprehensive", "TestSystemST_COMPREHENSIVE_01_AllOperations"),
            ("Performance", "TestSystemST_PERFORMANCE_01_UploadDownloadSpeed"),
            ("Reliability", "TestSystemST_RELIABILITY_01_ErrorRecovery"),
            ("Integration", "TestSystemST_INTEGRATION_01_MountUnmount"),
            ("Stress", "TestSystemST_STRESS_01_HighLoad"),
        ]
        
        for category, pattern in test_categories:
            if not self.run_test_category(category, pattern):
                failed_tests.append(category)
        
        # Report results
        if not failed_tests:
            self._log_success("All system tests completed successfully!")
            return True
        else:
            self._log_error(f"The following test categories failed: {', '.join(failed_tests)}")
            return False
    
    def set_timeout(self, timeout: str):
        """Set test timeout."""
        self.timeout = timeout
    
    def run_system_tests(
        self,
        category: str = "comprehensive",
        timeout: str = "30m",
        verbose: bool = False
    ) -> bool:
        """
        Main method to run system tests.
        
        Args:
            category: Test category to run
            timeout: Test timeout duration
            verbose: Enable verbose output
            
        Returns:
            True if tests succeeded, False otherwise
        """
        try:
            # Update configuration
            self.timeout = timeout
            self.verbose = verbose
            
            self._log_info("OneMount System Test Runner")
            self._log_info(f"Test category: {category}")
            self._log_info(f"Timeout: {timeout}")
            self._log_info(f"Log file: {self.test_log_path}")
            console.print()
            
            # Check prerequisites
            if not self.check_prerequisites():
                return False
            
            # Define test patterns for each category
            test_patterns = {
                "comprehensive": "TestSystemST_COMPREHENSIVE_01_AllOperations",
                "performance": "TestSystemST_PERFORMANCE_01_UploadDownloadSpeed",
                "reliability": "TestSystemST_RELIABILITY_01_ErrorRecovery",
                "integration": "TestSystemST_INTEGRATION_01_MountUnmount",
                "stress": "TestSystemST_STRESS_01_HighLoad",
            }
            
            # Run tests based on category
            if category == "all":
                success = self.run_all_tests()
            elif category in test_patterns:
                success = self.run_test_category(category.title(), test_patterns[category])
            else:
                self._log_error(f"Unknown test category: {category}")
                return False
            
            console.print()
            if success:
                self._log_success("System tests completed successfully!")
                self._log_info(f"Check the log file for detailed output: {self.test_log_path}")
            else:
                self._log_error("System tests failed!")
                self._log_info(f"Check the log file for error details: {self.test_log_path}")
            
            return success
            
        except SystemTestError as e:
            self._log_error(str(e))
            return False
        except Exception as e:
            self._log_error(f"Unexpected error during system tests: {e}")
            return False


def run_system_tests(
    category: str = "comprehensive",
    timeout: str = "30m",
    verbose: bool = False
) -> bool:
    """
    Convenience function to run system tests.
    
    Args:
        category: Test category to run
        timeout: Test timeout duration
        verbose: Enable verbose output
        
    Returns:
        True if tests succeeded, False otherwise
    """
    runner = SystemTestRunner(verbose=verbose)
    return runner.run_system_tests(category=category, timeout=timeout, verbose=verbose)
