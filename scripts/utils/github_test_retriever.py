#!/usr/bin/env python3
"""
GitHub Test Results Retriever for OneMount

Retrieves test results, coverage reports, and artifacts from GitHub Actions workflows.
"""

import json
import os
import zipfile
from pathlib import Path
from typing import Dict, List, Optional, Any
import requests
from rich.console import Console
from rich.table import Table
from rich.progress import Progress, DownloadColumn, BarColumn, TextColumn

console = Console()


class GitHubTestRetriever:
    """Retrieve test results and artifacts from GitHub Actions."""
    
    def __init__(self, token: Optional[str] = None):
        """Initialize with GitHub token."""
        self.token = token or os.getenv('GITHUB_TOKEN')
        if not self.token:
            raise ValueError("GitHub token required. Set GITHUB_TOKEN environment variable or pass token parameter.")
        
        self.headers = {
            'Authorization': f'token {self.token}',
            'Accept': 'application/vnd.github.v3+json'
        }
        self.base_url = 'https://api.github.com'
        self.repo = 'Auriora/OneMount'
    
    def get_workflow_runs(self, workflow_name: Optional[str] = None, limit: int = 10) -> List[Dict[str, Any]]:
        """Get recent workflow runs."""
        url = f"{self.base_url}/repos/{self.repo}/actions/runs"
        params = {'per_page': limit}
        
        if workflow_name:
            # First get workflow ID
            workflows_url = f"{self.base_url}/repos/{self.repo}/actions/workflows"
            response = requests.get(workflows_url, headers=self.headers)
            response.raise_for_status()
            
            workflows = response.json()['workflows']
            workflow_id = None
            for workflow in workflows:
                if workflow_name.lower() in workflow['name'].lower():
                    workflow_id = workflow['id']
                    break
            
            if workflow_id:
                params['workflow_id'] = workflow_id
        
        response = requests.get(url, headers=self.headers, params=params)
        response.raise_for_status()
        return response.json()['workflow_runs']
    
    def get_run_artifacts(self, run_id: int) -> List[Dict[str, Any]]:
        """Get artifacts for a specific workflow run."""
        url = f"{self.base_url}/repos/{self.repo}/actions/runs/{run_id}/artifacts"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()['artifacts']
    
    def download_artifact(self, artifact_id: int, output_dir: Path) -> Path:
        """Download and extract an artifact."""
        url = f"{self.base_url}/repos/{self.repo}/actions/artifacts/{artifact_id}/zip"
        
        # Create output directory
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Download artifact
        response = requests.get(url, headers=self.headers, stream=True)
        response.raise_for_status()
        
        zip_path = output_dir / f"artifact_{artifact_id}.zip"
        
        # Download with progress bar
        total_size = int(response.headers.get('content-length', 0))
        with Progress(
            TextColumn("[bold blue]Downloading artifact..."),
            BarColumn(),
            DownloadColumn(),
            console=console
        ) as progress:
            task = progress.add_task("download", total=total_size)
            
            with open(zip_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    if chunk:
                        f.write(chunk)
                        progress.update(task, advance=len(chunk))
        
        # Extract zip file
        extract_dir = output_dir / f"artifact_{artifact_id}"
        with zipfile.ZipFile(zip_path, 'r') as zip_ref:
            zip_ref.extractall(extract_dir)
        
        # Remove zip file
        zip_path.unlink()
        
        return extract_dir
    
    def get_check_runs(self, commit_sha: str) -> List[Dict[str, Any]]:
        """Get check runs for a specific commit."""
        url = f"{self.base_url}/repos/{self.repo}/commits/{commit_sha}/check-runs"
        response = requests.get(url, headers=self.headers)
        response.raise_for_status()
        return response.json()['check_runs']
    
    def get_latest_test_results(self, output_dir: Path = Path("./test-results-download")) -> Dict[str, Path]:
        """Download latest test results from all workflows."""
        console.print("[bold blue]Retrieving latest test results from GitHub...[/bold blue]")
        
        # Get recent workflow runs
        runs = self.get_workflow_runs(limit=20)
        
        downloaded_artifacts = {}
        
        for run in runs:
            if run['conclusion'] not in ['success', 'failure']:
                continue  # Skip cancelled/pending runs
            
            console.print(f"Checking run: {run['name']} - {run['conclusion']}")
            
            # Get artifacts for this run
            artifacts = self.get_run_artifacts(run['id'])
            
            for artifact in artifacts:
                artifact_name = artifact['name']
                
                # Skip if we already have this type of artifact
                if artifact_name in downloaded_artifacts:
                    continue
                
                console.print(f"Downloading artifact: {artifact_name}")
                
                try:
                    extract_dir = self.download_artifact(artifact['id'], output_dir / artifact_name)
                    downloaded_artifacts[artifact_name] = extract_dir
                    console.print(f"✅ Downloaded: {artifact_name}")
                except Exception as e:
                    console.print(f"❌ Failed to download {artifact_name}: {e}")
        
        return downloaded_artifacts
    
    def show_test_summary(self, results_dir: Path):
        """Display a summary of downloaded test results."""
        console.print("\n[bold green]Test Results Summary[/bold green]")
        
        table = Table(title="Downloaded Test Artifacts")
        table.add_column("Artifact Type", style="cyan")
        table.add_column("Files", style="magenta")
        table.add_column("Location", style="green")
        
        for artifact_dir in results_dir.iterdir():
            if artifact_dir.is_dir():
                files = list(artifact_dir.rglob("*"))
                file_count = len([f for f in files if f.is_file()])
                table.add_row(
                    artifact_dir.name,
                    str(file_count),
                    str(artifact_dir)
                )
        
        console.print(table)
        
        # Show specific test result files
        junit_files = list(results_dir.rglob("junit.xml"))
        json_files = list(results_dir.rglob("*.json"))
        coverage_files = list(results_dir.rglob("coverage.*"))
        
        if junit_files:
            console.print(f"\n[bold yellow]JUnit XML Reports:[/bold yellow]")
            for file in junit_files:
                console.print(f"  📄 {file}")
        
        if json_files:
            console.print(f"\n[bold yellow]JSON Reports:[/bold yellow]")
            for file in json_files:
                console.print(f"  📄 {file}")
        
        if coverage_files:
            console.print(f"\n[bold yellow]Coverage Reports:[/bold yellow]")
            for file in coverage_files:
                console.print(f"  📄 {file}")


def main():
    """Main function for CLI usage."""
    import argparse
    
    parser = argparse.ArgumentParser(description="Retrieve test results from GitHub Actions")
    parser.add_argument("--token", help="GitHub token (or set GITHUB_TOKEN env var)")
    parser.add_argument("--output", "-o", default="./test-results-download", 
                       help="Output directory for downloaded results")
    parser.add_argument("--workflow", help="Filter by workflow name")
    parser.add_argument("--run-id", type=int, help="Download artifacts from specific run ID")
    
    args = parser.parse_args()
    
    try:
        retriever = GitHubTestRetriever(token=args.token)
        output_dir = Path(args.output)
        
        if args.run_id:
            # Download artifacts from specific run
            console.print(f"Downloading artifacts from run {args.run_id}")
            artifacts = retriever.get_run_artifacts(args.run_id)
            
            for artifact in artifacts:
                console.print(f"Downloading: {artifact['name']}")
                extract_dir = retriever.download_artifact(artifact['id'], output_dir / artifact['name'])
                console.print(f"✅ Extracted to: {extract_dir}")
        else:
            # Download latest test results
            downloaded = retriever.get_latest_test_results(output_dir)
            
            if downloaded:
                retriever.show_test_summary(output_dir)
                console.print(f"\n[bold green]✅ Test results downloaded to: {output_dir}[/bold green]")
            else:
                console.print("[yellow]No test artifacts found in recent workflow runs[/yellow]")
    
    except Exception as e:
        console.print(f"[red]Error: {e}[/red]")
        return 1
    
    return 0


if __name__ == "__main__":
    exit(main())
