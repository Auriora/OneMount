"""
Native Python implementation for OneMount CI setup.
Replaces setup-personal-ci.sh with native Python operations.
"""

import base64
import json
import os
import shutil
import time
from pathlib import Path
from typing import Dict, Optional, Tuple

import requests
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from .paths import get_project_paths
from .shell import run_command, CommandError

console = Console()


class CISetupError(Exception):
    """Exception raised when CI setup operations fail."""
    pass


class CISetup:
    """Native Python CI setup for OneMount."""
    
    def __init__(self, verbose: bool = False):
        self.verbose = verbose
        self.paths = get_project_paths()
        self.selected_auth_file = None
        
        # Auth file locations
        self.auth_file = Path.home() / ".cache" / "onemount" / "auth_tokens.json"
        self.test_auth_file = Path.home() / ".onemount-tests" / ".auth_tokens.json"
        
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
    
    def check_auth(self) -> bool:
        """Check if OneMount authentication exists and is valid."""
        self._log_info("Checking OneMount authentication...")
        
        # Check both locations
        if self.auth_file.exists():
            self.selected_auth_file = self.auth_file
            self._log_info(f"Found authentication at: {self.auth_file}")
        elif self.test_auth_file.exists():
            self.selected_auth_file = self.test_auth_file
            self._log_info(f"Found authentication at: {self.test_auth_file}")
        else:
            self._log_error("OneMount authentication not found")
            self._log_error("Checked locations:")
            self._log_error(f"  - {self.auth_file}")
            self._log_error(f"  - {self.test_auth_file}")
            self._log_error("")
            self._log_error("Please authenticate with OneMount first:")
            self._log_error("  make onemount")
            self._log_error("  ./build/onemount --auth-only")
            self._log_error("")
            self._log_error("Follow the authentication prompts to sign in to your OneDrive account.")
            return False
        
        # Check if the file is valid JSON
        try:
            with open(self.selected_auth_file, 'r') as f:
                auth_data = json.load(f)
        except (json.JSONDecodeError, Exception) as e:
            self._log_error("Authentication file exists but is not valid JSON")
            self._log_error("Please re-authenticate with OneMount")
            return False
        
        # Check token expiration
        expires_at = auth_data.get('expires_at', 0)
        current_time = int(time.time())
        
        if expires_at <= current_time:
            self._log_warning("Authentication tokens appear to be expired")
            self._log_warning("Please re-authenticate with OneMount:")
            self._log_warning("  ./build/onemount --auth-only")
            return False
        
        # Get account info if available
        account = auth_data.get('account', 'Unknown')
        time_left = expires_at - current_time
        hours_left = time_left // 3600
        
        self._log_success("OneMount authentication found and valid")
        self._log_info(f"Account: {account}")
        self._log_info(f"Token expires in: {hours_left} hours")
        
        return True
    
    def generate_secret(self) -> bool:
        """Generate the GitHub secret value."""
        self._log_info("Generating GitHub secret value...")
        
        if not self.check_auth():
            return False
        
        # Generate base64-encoded secret using the selected auth file
        try:
            with open(self.selected_auth_file, 'rb') as f:
                secret_value = base64.b64encode(f.read()).decode('utf-8')
        except Exception as e:
            self._log_error(f"Failed to read auth file: {e}")
            return False
        
        self._log_success("GitHub secret value generated successfully!")
        console.print()
        
        # Display instructions in a nice panel
        instructions = f"""[bold cyan]Secret Name:[/bold cyan] ONEDRIVE_PERSONAL_TOKENS

[bold cyan]Secret Value:[/bold cyan]
{secret_value}

[bold yellow]To add this secret to your GitHub repository:[/bold yellow]
1. Go to your repository on GitHub
2. Click Settings → Secrets and variables → Actions
3. Click 'New repository secret'
4. Name: ONEDRIVE_PERSONAL_TOKENS
5. Value: Paste the value above
6. Click 'Add secret'

After adding the secret, the CI workflow will automatically run system tests!"""
        
        console.print(Panel(instructions, title="GitHub Secret Configuration", border_style="green"))
        console.print()
        
        return True
    
    def test_onedrive_access(self) -> bool:
        """Test OneDrive access using the auth tokens."""
        self._log_info("Testing OneDrive access...")
        
        if not self.selected_auth_file:
            self._log_error("No auth file selected")
            return False
        
        try:
            with open(self.selected_auth_file, 'r') as f:
                auth_data = json.load(f)
        except Exception as e:
            self._log_error(f"Failed to read auth file: {e}")
            return False
        
        access_token = auth_data.get('access_token')
        if not access_token:
            self._log_error("Could not extract access token from authentication file")
            return False
        
        # Test OneDrive API access
        try:
            headers = {'Authorization': f'Bearer {access_token}'}
            response = requests.get(
                'https://graph.microsoft.com/v1.0/me/drive/root',
                headers=headers,
                timeout=10
            )
            
            if response.status_code == 200:
                drive_data = response.json()
                drive_name = drive_data.get('name', 'Unknown')
                self._log_success("OneDrive access verified successfully")
                self._log_info(f"Drive Name: {drive_name}")
                return True
            else:
                self._log_error("Failed to access OneDrive")
                self._log_error("Please check your internet connection and re-authenticate")
                return False
                
        except requests.RequestException as e:
            self._log_error(f"Failed to access OneDrive: {e}")
            self._log_error("Please check your internet connection and re-authenticate")
            return False
    
    def verify_setup(self) -> bool:
        """Verify the complete CI setup."""
        self._log_info("Verifying CI setup...")
        
        # Check if workflow file exists
        workflow_file = self.paths["project_root"] / ".github" / "workflows" / "system-tests-personal.yml"
        if not workflow_file.exists():
            self._log_error(f"CI workflow file not found: {workflow_file}")
            self._log_error("Please ensure the workflow file is committed to your repository")
            return False
        
        self._log_success(f"CI workflow file found: {workflow_file}")
        
        # Check authentication
        if not self.check_auth():
            return False
        
        # Test OneDrive access
        if not self.test_onedrive_access():
            return False
        
        # Check if system test runner exists
        self._log_info("Checking if system tests can run locally...")
        
        # Check for the new Python system test runner
        system_test_runner = self.paths["project_root"] / "scripts" / "utils" / "system_test_runner.py"
        if system_test_runner.exists():
            self._log_success("Python system test runner found")
        else:
            # Fallback to legacy script
            legacy_script = self.paths["project_root"] / "scripts" / "run-system-tests.sh"
            if not legacy_script.exists():
                self._log_error("System test script not found")
                return False
            self._log_info("Legacy system test script found")
        
        # Copy auth tokens to test location (if not already there)
        test_dir = Path.home() / ".onemount-tests"
        test_dir.mkdir(exist_ok=True)
        
        copied_auth_file = False
        if self.selected_auth_file != self.test_auth_file:
            try:
                shutil.copy2(self.selected_auth_file, self.test_auth_file)
                self.test_auth_file.chmod(0o600)
                copied_auth_file = True
            except Exception as e:
                self._log_error(f"Failed to copy auth file: {e}")
                return False
        else:
            self._log_info("Auth tokens already in test location")
        
        # Test the system test runner
        try:
            # Try the new Python implementation first
            from .system_test_runner import SystemTestRunner
            runner = SystemTestRunner(verbose=False)
            if runner.check_prerequisites():
                self._log_success("System test runner is ready")
            else:
                self._log_error("System test runner has issues")
                return False
        except ImportError:
            # Fallback to legacy script
            try:
                run_command(
                    ["./scripts/run-system-tests.sh", "--help"],
                    capture_output=True,
                    check=True,
                    verbose=False,
                    timeout=10,
                    cwd=str(self.paths["project_root"])
                )
                self._log_success("Legacy system test script is executable and ready")
            except (CommandError, Exception):
                self._log_error("System test script has issues")
                return False
        
        # Clean up test auth file only if we copied it
        if copied_auth_file:
            try:
                self.test_auth_file.unlink()
                self._log_info("Cleaned up temporary auth file")
            except Exception:
                pass  # Ignore cleanup errors
        
        self._log_success("✅ CI setup verification completed successfully!")
        console.print()
        
        # Display completion instructions
        completion_info = """[bold green]Your setup is ready![/bold green] To complete the CI configuration:

1. Run: [yellow]dev.py ci generate-secret[/yellow]
2. Add the generated secret to GitHub
3. Push your code to trigger the CI tests

You can also manually trigger tests in GitHub Actions:
- Go to Actions tab → System Tests (Personal OneDrive) → Run workflow"""
        
        console.print(Panel(completion_info, title="Setup Complete", border_style="green"))
        console.print()
        
        return True

    def run_full_setup(self) -> bool:
        """Run the complete CI setup process."""
        console.print(Panel(
            "[bold blue]OneMount Personal OneDrive CI Setup[/bold blue]",
            subtitle="Setting up CI system tests with your personal OneDrive",
            border_style="blue"
        ))
        console.print()

        self._log_info("This script will help you set up CI system tests with your personal OneDrive.")
        console.print()

        # Step 1: Check authentication
        self._log_info("Step 1: Check authentication")
        if not self.check_auth():
            return False

        console.print()

        # Step 2: Generate GitHub secret
        self._log_info("Step 2: Generate GitHub secret")
        if not self.generate_secret():
            return False

        # Step 3: Verify complete setup
        self._log_info("Step 3: Verify complete setup")
        if not self.verify_setup():
            return False

        return True

    def show_status(self):
        """Show CI setup status."""
        console.print(Panel(
            "[bold blue]OneMount CI Setup Status[/bold blue]",
            border_style="blue"
        ))
        console.print()

        # Create status table
        table = Table()
        table.add_column("Component", style="cyan")
        table.add_column("Status", style="green")
        table.add_column("Details", style="dim")

        # Check auth status
        if self.check_auth():
            auth_status = "✅ Valid"
            auth_details = f"Account: {self.selected_auth_file}"
        else:
            auth_status = "❌ Missing/Invalid"
            auth_details = "Authentication required"

        table.add_row("Authentication", auth_status, auth_details)

        # Check workflow file
        workflow_file = self.paths["project_root"] / ".github" / "workflows" / "system-tests-personal.yml"
        if workflow_file.exists():
            workflow_status = "✅ Found"
            workflow_details = str(workflow_file)
        else:
            workflow_status = "❌ Missing"
            workflow_details = "Workflow file not found"

        table.add_row("CI Workflow", workflow_status, workflow_details)

        # Check system test runner
        system_test_runner = self.paths["project_root"] / "scripts" / "utils" / "system_test_runner.py"
        if system_test_runner.exists():
            runner_status = "✅ Python Runner"
            runner_details = "Native Python implementation"
        else:
            legacy_script = self.paths["project_root"] / "scripts" / "run-system-tests.sh"
            if legacy_script.exists():
                runner_status = "⚠️ Legacy Script"
                runner_details = "Shell script (consider migrating)"
            else:
                runner_status = "❌ Missing"
                runner_details = "No test runner found"

        table.add_row("Test Runner", runner_status, runner_details)

        console.print(table)
        console.print()


def check_auth(verbose: bool = False) -> bool:
    """
    Convenience function to check authentication.

    Args:
        verbose: Enable verbose output

    Returns:
        True if authentication is valid, False otherwise
    """
    setup = CISetup(verbose=verbose)
    return setup.check_auth()


def generate_secret(verbose: bool = False) -> bool:
    """
    Convenience function to generate GitHub secret.

    Args:
        verbose: Enable verbose output

    Returns:
        True if secret generation succeeded, False otherwise
    """
    setup = CISetup(verbose=verbose)
    return setup.generate_secret()


def verify_setup(verbose: bool = False) -> bool:
    """
    Convenience function to verify CI setup.

    Args:
        verbose: Enable verbose output

    Returns:
        True if setup verification succeeded, False otherwise
    """
    setup = CISetup(verbose=verbose)
    return setup.verify_setup()


def run_full_setup(verbose: bool = False) -> bool:
    """
    Convenience function to run full CI setup.

    Args:
        verbose: Enable verbose output

    Returns:
        True if full setup succeeded, False otherwise
    """
    setup = CISetup(verbose=verbose)
    return setup.run_full_setup()


def show_status(verbose: bool = False):
    """
    Convenience function to show CI setup status.

    Args:
        verbose: Enable verbose output
    """
    setup = CISetup(verbose=verbose)
    setup.show_status()
