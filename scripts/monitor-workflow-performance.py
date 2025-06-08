#!/usr/bin/env python3
"""
OneMount Workflow Performance Monitor

This script monitors GitHub Actions workflow performance and provides
optimization recommendations.
"""

import json
import sys
import time
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional

import requests
import yaml


class WorkflowMonitor:
    """Monitor and analyze GitHub Actions workflow performance."""
    
    def __init__(self, repo: str, token: str):
        self.repo = repo
        self.token = token
        self.headers = {
            'Authorization': f'token {token}',
            'Accept': 'application/vnd.github.v3+json'
        }
        self.base_url = f'https://api.github.com/repos/{repo}'
        
        # Load optimization config
        config_path = Path('.github/workflow-optimization.yml')
        if config_path.exists():
            with open(config_path) as f:
                self.config = yaml.safe_load(f)
        else:
            self.config = {}
    
    def get_workflow_runs(self, days: int = 7) -> List[Dict]:
        """Get workflow runs from the last N days."""
        since = (datetime.now() - timedelta(days=days)).isoformat()
        
        url = f'{self.base_url}/actions/runs'
        params = {
            'created': f'>{since}',
            'per_page': 100
        }
        
        response = requests.get(url, headers=self.headers, params=params)
        response.raise_for_status()
        
        return response.json()['workflow_runs']
    
    def analyze_performance(self, runs: List[Dict]) -> Dict:
        """Analyze workflow performance metrics."""
        analysis = {
            'total_runs': len(runs),
            'workflows': {},
            'performance_issues': [],
            'recommendations': []
        }
        
        # Group runs by workflow
        for run in runs:
            workflow_name = run['name']
            if workflow_name not in analysis['workflows']:
                analysis['workflows'][workflow_name] = {
                    'runs': [],
                    'avg_duration': 0,
                    'success_rate': 0,
                    'failures': 0
                }
            
            # Calculate duration in minutes
            if run['conclusion'] and run['created_at'] and run['updated_at']:
                start = datetime.fromisoformat(run['created_at'].replace('Z', '+00:00'))
                end = datetime.fromisoformat(run['updated_at'].replace('Z', '+00:00'))
                duration = (end - start).total_seconds() / 60
                
                analysis['workflows'][workflow_name]['runs'].append({
                    'id': run['id'],
                    'duration': duration,
                    'conclusion': run['conclusion'],
                    'created_at': run['created_at']
                })
        
        # Calculate metrics for each workflow
        for workflow_name, data in analysis['workflows'].items():
            runs_data = data['runs']
            if not runs_data:
                continue
                
            # Average duration
            durations = [r['duration'] for r in runs_data if r['duration'] > 0]
            data['avg_duration'] = sum(durations) / len(durations) if durations else 0
            
            # Success rate
            successful = len([r for r in runs_data if r['conclusion'] == 'success'])
            data['success_rate'] = (successful / len(runs_data)) * 100
            data['failures'] = len(runs_data) - successful
            
            # Check against performance targets
            targets = self.config.get('performance_targets', {})
            target_key = workflow_name.lower().replace(' ', '_').replace('(', '').replace(')', '')
            
            if target_key in targets:
                target_range = targets[target_key]
                if isinstance(target_range, str) and '-' in target_range:
                    max_target = float(target_range.split('-')[1])
                    if data['avg_duration'] > max_target:
                        analysis['performance_issues'].append({
                            'workflow': workflow_name,
                            'issue': 'duration_exceeded',
                            'current': f"{data['avg_duration']:.1f}m",
                            'target': target_range,
                            'severity': 'high' if data['avg_duration'] > max_target * 1.5 else 'medium'
                        })
        
        # Generate recommendations
        analysis['recommendations'] = self._generate_recommendations(analysis)
        
        return analysis
    
    def _generate_recommendations(self, analysis: Dict) -> List[Dict]:
        """Generate optimization recommendations based on analysis."""
        recommendations = []
        
        for workflow_name, data in analysis['workflows'].items():
            # High duration recommendation
            if data['avg_duration'] > 15:
                recommendations.append({
                    'workflow': workflow_name,
                    'type': 'performance',
                    'priority': 'high',
                    'title': 'Consider self-hosted runners',
                    'description': f'Workflow takes {data["avg_duration"]:.1f}m on average. '
                                 'Self-hosted runners could reduce this by 60-80%.',
                    'action': 'Setup self-hosted runners using scripts/setup-optimized-runners.sh'
                })
            
            # Low success rate recommendation
            if data['success_rate'] < 90:
                recommendations.append({
                    'workflow': workflow_name,
                    'type': 'reliability',
                    'priority': 'high',
                    'title': 'Improve workflow reliability',
                    'description': f'Success rate is {data["success_rate"]:.1f}%. '
                                 'Consider adding retry logic and better error handling.',
                    'action': 'Review failed runs and add appropriate retry mechanisms'
                })
            
            # Caching recommendation
            if 'ci' in workflow_name.lower() or 'coverage' in workflow_name.lower():
                recommendations.append({
                    'workflow': workflow_name,
                    'type': 'caching',
                    'priority': 'medium',
                    'title': 'Optimize caching strategy',
                    'description': 'Enhanced Go module and build caching can reduce build times.',
                    'action': 'Implement multi-level caching for Go modules and build artifacts'
                })
        
        return recommendations
    
    def print_report(self, analysis: Dict):
        """Print a formatted performance report."""
        print("🚀 OneMount Workflow Performance Report")
        print("=" * 50)
        print(f"📊 Analysis Period: Last 7 days")
        print(f"📈 Total Runs: {analysis['total_runs']}")
        print()
        
        # Workflow performance table
        print("📋 Workflow Performance:")
        print("-" * 80)
        print(f"{'Workflow':<30} {'Avg Duration':<15} {'Success Rate':<15} {'Failures':<10}")
        print("-" * 80)
        
        for workflow_name, data in analysis['workflows'].items():
            duration_str = f"{data['avg_duration']:.1f}m"
            success_str = f"{data['success_rate']:.1f}%"
            failures_str = str(data['failures'])
            
            print(f"{workflow_name:<30} {duration_str:<15} {success_str:<15} {failures_str:<10}")
        
        print()
        
        # Performance issues
        if analysis['performance_issues']:
            print("⚠️  Performance Issues:")
            print("-" * 50)
            for issue in analysis['performance_issues']:
                severity_icon = "🔴" if issue['severity'] == 'high' else "🟡"
                print(f"{severity_icon} {issue['workflow']}: {issue['issue']}")
                print(f"   Current: {issue['current']}, Target: {issue['target']}")
            print()
        
        # Recommendations
        if analysis['recommendations']:
            print("💡 Optimization Recommendations:")
            print("-" * 50)
            for i, rec in enumerate(analysis['recommendations'], 1):
                priority_icon = "🔴" if rec['priority'] == 'high' else "🟡" if rec['priority'] == 'medium' else "🟢"
                print(f"{i}. {priority_icon} {rec['title']} ({rec['workflow']})")
                print(f"   {rec['description']}")
                print(f"   Action: {rec['action']}")
                print()
    
    def save_report(self, analysis: Dict, filename: str = 'workflow-performance-report.json'):
        """Save the analysis report to a JSON file."""
        with open(filename, 'w') as f:
            json.dump(analysis, f, indent=2, default=str)
        print(f"📄 Report saved to {filename}")


def main():
    """Main function."""
    if len(sys.argv) < 3:
        print("Usage: python3 monitor-workflow-performance.py <repo> <github_token>")
        print("Example: python3 monitor-workflow-performance.py Auriora/OneMount ghp_xxx")
        sys.exit(1)
    
    repo = sys.argv[1]
    token = sys.argv[2]
    
    try:
        monitor = WorkflowMonitor(repo, token)
        print("🔍 Fetching workflow runs...")
        runs = monitor.get_workflow_runs(days=7)
        
        print("📊 Analyzing performance...")
        analysis = monitor.analyze_performance(runs)
        
        monitor.print_report(analysis)
        monitor.save_report(analysis)
        
        # Exit with error code if there are high-priority issues
        high_priority_issues = [
            r for r in analysis['recommendations'] 
            if r['priority'] == 'high'
        ]
        
        if high_priority_issues:
            print(f"\n⚠️  Found {len(high_priority_issues)} high-priority issues!")
            sys.exit(1)
        else:
            print("\n✅ No critical performance issues found!")
            sys.exit(0)
            
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)


if __name__ == '__main__':
    main()
