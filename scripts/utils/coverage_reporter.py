"""
Native Python implementation for coverage reporting.
Replaces coverage-report.sh with native Python operations.
"""

import json
import re
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from jinja2 import Template
from rich.console import Console
from rich.progress import Progress, SpinnerColumn, TextColumn
from rich.table import Table

from .paths import get_project_paths
from .shell import run_command, CommandError

console = Console()


class CoverageError(Exception):
    """Exception raised when coverage operations fail."""
    pass


class CoverageReporter:
    """Native Python coverage reporter for OneMount."""
    
    def __init__(self, verbose: bool = False, ci_mode: bool = False):
        self.verbose = verbose
        self.ci_mode = ci_mode
        self.paths = get_project_paths()
        self.coverage_dir = self.paths["project_root"] / "coverage"
        
        # Default thresholds
        self.threshold_line = 80
        self.threshold_func = 90
        self.threshold_branch = 70
        
    def _log_info(self, message: str):
        """Log info message."""
        if not self.ci_mode:
            console.print(f"[blue][INFO][/blue] {message}")
        else:
            print(f"[INFO] {message}")
    
    def _log_success(self, message: str):
        """Log success message."""
        if not self.ci_mode:
            console.print(f"[green][SUCCESS][/green] {message}")
        else:
            print(f"[SUCCESS] {message}")
    
    def _log_warning(self, message: str):
        """Log warning message."""
        if not self.ci_mode:
            console.print(f"[yellow][WARNING][/yellow] {message}")
        else:
            print(f"[WARNING] {message}")
    
    def _log_error(self, message: str):
        """Log error message."""
        if not self.ci_mode:
            console.print(f"[red][ERROR][/red] {message}")
        else:
            print(f"[ERROR] {message}")
    
    def set_thresholds(self, line: int = None, func: int = None, branch: int = None):
        """Set coverage thresholds."""
        if line is not None:
            self.threshold_line = line
        if func is not None:
            self.threshold_func = func
        if branch is not None:
            self.threshold_branch = branch
    
    def check_coverage_file(self) -> bool:
        """Check if coverage file exists."""
        coverage_file = self.coverage_dir / "coverage.out"
        
        if not coverage_file.exists():
            self._log_error(f"Coverage file not found: {coverage_file}")
            self._log_info("Run 'make coverage' first to generate coverage data")
            return False
        
        return True
    
    def ensure_coverage_directory(self):
        """Create coverage directory if it doesn't exist."""
        self.coverage_dir.mkdir(parents=True, exist_ok=True)
    
    def generate_html_report(self) -> bool:
        """Generate HTML coverage report."""
        try:
            coverage_file = self.coverage_dir / "coverage.out"
            html_report = self.coverage_dir / "coverage.html"
            
            self._log_info("Generating HTML coverage report...")
            
            result = run_command(
                ["go", "tool", "cover", f"-html={coverage_file}", f"-o={html_report}"],
                capture_output=True,
                check=True,
                verbose=self.verbose,
                timeout=30
            )
            
            self._log_success(f"HTML report generated: {html_report}")
            return True
            
        except CommandError as e:
            self._log_error(f"Failed to generate HTML report: {e}")
            return False
    
    def generate_function_coverage(self) -> Tuple[float, str]:
        """Generate function coverage analysis and return total coverage."""
        try:
            coverage_file = self.coverage_dir / "coverage.out"
            func_report = self.coverage_dir / "coverage-func.txt"
            
            self._log_info("Generating function coverage analysis...")
            
            result = run_command(
                ["go", "tool", "cover", f"-func={coverage_file}"],
                capture_output=True,
                check=True,
                verbose=self.verbose,
                timeout=30
            )
            
            # Save function coverage to file
            with open(func_report, 'w') as f:
                f.write(result.stdout)
            
            # Extract total coverage
            total_coverage = 0.0
            for line in result.stdout.split('\n'):
                if 'total:' in line:
                    # Extract percentage from line like "total:                  (statements)    85.2%"
                    match = re.search(r'(\d+\.?\d*)%', line)
                    if match:
                        total_coverage = float(match.group(1))
                    break
            
            self._log_info(f"Total coverage: {total_coverage}%")
            return total_coverage, result.stdout
            
        except CommandError as e:
            self._log_error(f"Failed to generate function coverage: {e}")
            return 0.0, ""
    
    def generate_package_analysis(self, func_output: str) -> Dict[str, float]:
        """Generate package-by-package coverage analysis."""
        try:
            self._log_info("Generating package analysis...")
            
            package_coverage = {}
            
            for line in func_output.split('\n'):
                if line.strip() and not line.startswith('total:') and '.go:' in line:
                    # Parse line like "github.com/auriora/onemount/internal/fs/file.go:123:    FunctionName    85.7%"
                    parts = line.split()
                    if len(parts) >= 3:
                        file_path = parts[0]
                        coverage_str = parts[-1]
                        
                        # Extract package from file path
                        if '/' in file_path:
                            package = '/'.join(file_path.split('/')[:-1])  # Remove filename
                        else:
                            package = "root"
                        
                        # Extract coverage percentage
                        match = re.search(r'(\d+\.?\d*)%', coverage_str)
                        if match:
                            coverage = float(match.group(1))
                            
                            if package not in package_coverage:
                                package_coverage[package] = []
                            package_coverage[package].append(coverage)
            
            # Calculate average coverage per package
            package_averages = {}
            for package, coverages in package_coverage.items():
                if coverages:
                    package_averages[package] = sum(coverages) / len(coverages)
            
            # Save package analysis
            analysis_file = self.coverage_dir / "package-analysis.txt"
            with open(analysis_file, 'w') as f:
                f.write("OneMount Coverage Analysis Report\n")
                f.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
                f.write(f"Total Coverage: {self.get_total_coverage()}%\n\n")
                f.write("Package Coverage Details:\n")
                f.write("========================\n")
                
                for package, avg_coverage in sorted(package_averages.items()):
                    f.write(f"{package:<50} {avg_coverage:6.1f}%\n")
            
            return package_averages
            
        except Exception as e:
            self._log_error(f"Failed to generate package analysis: {e}")
            return {}
    
    def get_total_coverage(self) -> float:
        """Get total coverage from function analysis."""
        try:
            _, func_output = self.generate_function_coverage()
            for line in func_output.split('\n'):
                if 'total:' in line:
                    match = re.search(r'(\d+\.?\d*)%', line)
                    if match:
                        return float(match.group(1))
            return 0.0
        except:
            return 0.0
    
    def generate_json_report(self, total_coverage: float, package_coverage: Dict[str, float]) -> bool:
        """Generate JSON coverage report for programmatic access."""
        try:
            self._log_info("Generating JSON coverage report...")
            
            json_report = self.coverage_dir / "coverage.json"
            
            report_data = {
                "timestamp": datetime.now().isoformat(),
                "total_coverage": total_coverage,
                "thresholds": {
                    "line": self.threshold_line,
                    "function": self.threshold_func,
                    "branch": self.threshold_branch
                },
                "packages": [
                    {"name": pkg, "coverage": cov}
                    for pkg, cov in package_coverage.items()
                ],
                "files": []  # Could be expanded later
            }
            
            with open(json_report, 'w') as f:
                json.dump(report_data, f, indent=2)
            
            return True
            
        except Exception as e:
            self._log_error(f"Failed to generate JSON report: {e}")
            return False
    
    def update_coverage_history(self, total_coverage: float) -> bool:
        """Update coverage history."""
        try:
            self._log_info("Updating coverage history...")
            
            history_file = self.coverage_dir / "coverage_history.json"
            
            # Load existing history
            if history_file.exists():
                with open(history_file, 'r') as f:
                    history = json.load(f)
            else:
                history = []
            
            # Add new entry
            new_entry = {
                'timestamp': int(datetime.now().timestamp()),
                'total_coverage': total_coverage,
                'date': datetime.now().isoformat()
            }
            
            history.append(new_entry)
            
            # Keep only last 100 entries
            history = history[-100:]
            
            # Save updated history
            with open(history_file, 'w') as f:
                json.dump(history, f, indent=2)
            
            return True
            
        except Exception as e:
            self._log_warning(f"Could not update coverage history: {e}")
            return False

    def check_thresholds(self, total_coverage: float) -> bool:
        """Check if coverage meets thresholds."""
        self._log_info("Checking coverage thresholds...")

        threshold_passed = True

        if total_coverage < self.threshold_line:
            self._log_error(f"Line coverage {total_coverage}% is below threshold of {self.threshold_line}%")
            threshold_passed = False
        else:
            self._log_success(f"Line coverage {total_coverage}% meets threshold of {self.threshold_line}%")

        return threshold_passed

    def generate_coverage_gaps(self, func_output: str) -> bool:
        """Generate coverage gaps report."""
        try:
            self._log_info("Analyzing coverage gaps...")

            gaps_file = self.coverage_dir / "coverage-gaps.txt"

            with open(gaps_file, 'w') as f:
                f.write("Coverage Gaps Analysis\n")
                f.write("=====================\n")
                f.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n")
                f.write(f"Files with coverage below {self.threshold_line}%:\n")

                for line in func_output.split('\n'):
                    if line.strip() and not line.startswith('total:') and '.go:' in line:
                        parts = line.split()
                        if len(parts) >= 3:
                            file_path = parts[0]
                            coverage_str = parts[-1]

                            match = re.search(r'(\d+\.?\d*)%', coverage_str)
                            if match:
                                coverage = float(match.group(1))
                                if coverage < self.threshold_line:
                                    f.write(f"{file_path:<60} {coverage_str}\n")

            return True

        except Exception as e:
            self._log_error(f"Failed to generate coverage gaps: {e}")
            return False

    def generate_summary_report(self, total_coverage: float, threshold_passed: bool) -> bool:
        """Generate summary report."""
        try:
            self._log_info("Generating summary report...")

            summary_file = self.coverage_dir / "summary.txt"

            with open(summary_file, 'w') as f:
                f.write("OneMount Coverage Summary\n")
                f.write("========================\n")
                f.write(f"Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")
                f.write(f"Total Coverage: {total_coverage}%\n\n")

                f.write("Thresholds:\n")
                status = "✅ PASS" if threshold_passed else "❌ FAIL"
                f.write(f"- Line Coverage: {self.threshold_line}% {status}\n")
                f.write(f"- Function Coverage: {self.threshold_func}% (target)\n")
                f.write(f"- Branch Coverage: {self.threshold_branch}% (target)\n\n")

                f.write("Reports Generated:\n")
                f.write(f"- HTML Report: {self.coverage_dir}/coverage.html\n")
                f.write(f"- Function Analysis: {self.coverage_dir}/coverage-func.txt\n")
                f.write(f"- Package Analysis: {self.coverage_dir}/package-analysis.txt\n")
                f.write(f"- Coverage Gaps: {self.coverage_dir}/coverage-gaps.txt\n")
                f.write(f"- JSON Report: {self.coverage_dir}/coverage.json\n")
                f.write(f"- Coverage History: {self.coverage_dir}/coverage_history.json\n")

            return True

        except Exception as e:
            self._log_error(f"Failed to generate summary report: {e}")
            return False

    def display_summary(self, total_coverage: float, threshold_passed: bool):
        """Display coverage summary to console."""
        if not self.ci_mode:
            console.print("\n[bold cyan]📊 Coverage Report Summary[/bold cyan]")

            # Create summary table
            summary_table = Table()
            summary_table.add_column("Metric", style="cyan")
            summary_table.add_column("Value", style="green")
            summary_table.add_column("Status", style="yellow")

            status = "✅ PASS" if threshold_passed else "❌ FAIL"
            summary_table.add_row("Total Coverage", f"{total_coverage}%", status)
            summary_table.add_row("Line Threshold", f"{self.threshold_line}%", "Target")
            summary_table.add_row("Function Threshold", f"{self.threshold_func}%", "Target")
            summary_table.add_row("Branch Threshold", f"{self.threshold_branch}%", "Target")

            console.print(summary_table)

            console.print(f"\n[bold cyan]📁 Reports Location:[/bold cyan] {self.coverage_dir}")
        else:
            # CI mode - simple output
            print(f"Total Coverage: {total_coverage}%")
            print(f"Threshold: {self.threshold_line}%")
            print(f"Status: {'PASS' if threshold_passed else 'FAIL'}")

    def generate_comprehensive_report(self) -> bool:
        """Generate comprehensive coverage report."""
        try:
            # Check prerequisites
            if not self.check_coverage_file():
                return False

            # Ensure coverage directory
            self.ensure_coverage_directory()

            # Generate HTML report
            if not self.generate_html_report():
                return False

            # Generate function coverage and get total
            total_coverage, func_output = self.generate_function_coverage()
            if total_coverage == 0.0:
                return False

            # Generate package analysis
            package_coverage = self.generate_package_analysis(func_output)

            # Generate JSON report
            self.generate_json_report(total_coverage, package_coverage)

            # Update coverage history
            self.update_coverage_history(total_coverage)

            # Check thresholds
            threshold_passed = self.check_thresholds(total_coverage)

            # Generate coverage gaps
            self.generate_coverage_gaps(func_output)

            # Generate summary
            self.generate_summary_report(total_coverage, threshold_passed)

            # Display summary
            self.display_summary(total_coverage, threshold_passed)

            self._log_success("Coverage analysis complete!")

            return threshold_passed

        except Exception as e:
            self._log_error(f"Failed to generate coverage report: {e}")
            return False


def generate_coverage_report(
    verbose: bool = False,
    ci_mode: bool = False,
    threshold_line: int = 80,
    threshold_func: int = 90,
    threshold_branch: int = 70
) -> bool:
    """
    Convenience function to generate coverage report.

    Args:
        verbose: Enable verbose output
        ci_mode: Enable CI mode (machine-readable output)
        threshold_line: Line coverage threshold
        threshold_func: Function coverage threshold
        threshold_branch: Branch coverage threshold

    Returns:
        True if coverage meets thresholds, False otherwise
    """
    reporter = CoverageReporter(verbose=verbose, ci_mode=ci_mode)
    reporter.set_thresholds(line=threshold_line, func=threshold_func, branch=threshold_branch)
    return reporter.generate_comprehensive_report()
