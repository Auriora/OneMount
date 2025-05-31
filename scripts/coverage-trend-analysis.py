#!/usr/bin/env python3
"""
Coverage Trend Analysis Tool for OneMount

This script analyzes coverage trends over time and generates visualizations
and reports to help track coverage progress and identify regressions.

Usage:
    python3 coverage-trend-analysis.py --input coverage_history.json --output trends.html
"""

import json
import argparse
import sys
from datetime import datetime
from pathlib import Path

try:
    import matplotlib.pyplot as plt
    import matplotlib.dates as mdates
    import pandas as pd
    import numpy as np
    HAS_PLOTTING = True
except ImportError:
    plt = None
    mdates = None
    pd = None
    np = None
    HAS_PLOTTING = False
    print("Warning: matplotlib/pandas not available. Only text reports will be generated.")

def load_coverage_history(file_path):
    """Load coverage history from JSON file."""
    try:
        with open(file_path, 'r') as f:
            data = json.load(f)
        return data
    except FileNotFoundError:
        print(f"Error: Coverage history file not found: {file_path}")
        return []
    except json.JSONDecodeError:
        print(f"Error: Invalid JSON in coverage history file: {file_path}")
        return []

def analyze_trends(history_data):
    """Analyze coverage trends and detect patterns."""
    if len(history_data) < 2:
        return {
            'trend': 'insufficient_data',
            'change': 0,
            'regression_count': 0,
            'improvement_count': 0,
            'average_coverage': 0,
            'latest_coverage': 0
        }
    
    # Convert to pandas DataFrame for easier analysis
    df_data = []
    for entry in history_data:
        df_data.append({
            'timestamp': datetime.fromtimestamp(entry['timestamp']),
            'coverage': entry['total_coverage']
        })
    
    if HAS_PLOTTING:
        df = pd.DataFrame(df_data)
        df = df.sort_values('timestamp')
        
        # Calculate trend
        coverage_values = df['coverage'].values
        timestamps = np.arange(len(coverage_values))
        
        # Linear regression for trend
        if len(coverage_values) > 1:
            slope, intercept = np.polyfit(timestamps, coverage_values, 1)
            trend = 'improving' if slope > 0.1 else 'declining' if slope < -0.1 else 'stable'
        else:
            slope = 0
            trend = 'stable'
        
        # Count regressions and improvements
        regression_count = 0
        improvement_count = 0
        
        for i in range(1, len(coverage_values)):
            diff = coverage_values[i] - coverage_values[i-1]
            if diff < -1:  # Regression threshold: 1%
                regression_count += 1
            elif diff > 1:  # Improvement threshold: 1%
                improvement_count += 1
        
        return {
            'trend': trend,
            'change': slope,
            'regression_count': regression_count,
            'improvement_count': improvement_count,
            'average_coverage': np.mean(coverage_values),
            'latest_coverage': coverage_values[-1],
            'min_coverage': np.min(coverage_values),
            'max_coverage': np.max(coverage_values),
            'std_coverage': np.std(coverage_values)
        }
    else:
        # Simple analysis without pandas/numpy
        coverages = [entry['total_coverage'] for entry in history_data]
        latest = coverages[-1]
        previous = coverages[-2]
        
        return {
            'trend': 'improving' if latest > previous else 'declining' if latest < previous else 'stable',
            'change': latest - previous,
            'regression_count': 0,
            'improvement_count': 0,
            'average_coverage': sum(coverages) / len(coverages),
            'latest_coverage': latest,
            'min_coverage': min(coverages),
            'max_coverage': max(coverages)
        }

def generate_plot(history_data, output_path):
    """Generate coverage trend plot."""
    if not HAS_PLOTTING or len(history_data) < 2:
        return False
    
    # Prepare data
    timestamps = [datetime.fromtimestamp(entry['timestamp']) for entry in history_data]
    coverages = [entry['total_coverage'] for entry in history_data]
    
    # Create plot
    plt.figure(figsize=(12, 8))
    
    # Main coverage plot
    plt.subplot(2, 1, 1)
    plt.plot(timestamps, coverages, 'b-', linewidth=2, marker='o', markersize=4)
    plt.axhline(y=80, color='r', linestyle='--', alpha=0.7, label='Target (80%)')
    plt.title('OneMount Coverage Trend Over Time', fontsize=16, fontweight='bold')
    plt.ylabel('Coverage (%)', fontsize=12)
    plt.grid(True, alpha=0.3)
    plt.legend()
    
    # Format x-axis
    plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))
    plt.gca().xaxis.set_major_locator(mdates.DayLocator(interval=max(1, len(timestamps)//10)))
    plt.xticks(rotation=45)
    
    # Coverage change plot
    plt.subplot(2, 1, 2)
    if len(coverages) > 1:
        changes = [coverages[i] - coverages[i-1] for i in range(1, len(coverages))]
        change_timestamps = timestamps[1:]
        
        colors = ['green' if change >= 0 else 'red' for change in changes]
        plt.bar(change_timestamps, changes, color=colors, alpha=0.7, width=0.8)
        plt.axhline(y=0, color='black', linestyle='-', alpha=0.5)
        plt.title('Coverage Changes', fontsize=14)
        plt.ylabel('Change (%)', fontsize=12)
        plt.grid(True, alpha=0.3)
        
        # Format x-axis
        plt.gca().xaxis.set_major_formatter(mdates.DateFormatter('%Y-%m-%d'))
        plt.gca().xaxis.set_major_locator(mdates.DayLocator(interval=max(1, len(change_timestamps)//10)))
        plt.xticks(rotation=45)
    
    plt.tight_layout()
    
    # Save plot
    plot_path = output_path.replace('.html', '_plot.png')
    plt.savefig(plot_path, dpi=300, bbox_inches='tight')
    plt.close()
    
    return plot_path

def generate_html_report(history_data, analysis, output_path, plot_path=None):
    """Generate HTML coverage trend report."""
    
    # Determine trend emoji and color
    trend_info = {
        'improving': ('📈', 'green', 'Coverage is improving over time'),
        'declining': ('📉', 'red', 'Coverage is declining over time'),
        'stable': ('📊', 'blue', 'Coverage is stable'),
        'insufficient_data': ('❓', 'gray', 'Insufficient data for trend analysis')
    }
    
    emoji, color, description = trend_info.get(analysis['trend'], ('❓', 'gray', 'Unknown trend'))
    
    # Generate recent history table
    recent_history = ""
    if len(history_data) > 0:
        recent_entries = history_data[-10:]  # Last 10 entries
        recent_history = "<table class='history-table'>\n"
        recent_history += "<tr><th>Date</th><th>Coverage</th><th>Change</th></tr>\n"
        
        for i, entry in enumerate(recent_entries):
            date = datetime.fromtimestamp(entry['timestamp']).strftime('%Y-%m-%d %H:%M')
            coverage = entry['total_coverage']
            
            if i > 0:
                change = coverage - recent_entries[i-1]['total_coverage']
                change_str = f"{change:+.1f}%"
                change_class = "positive" if change >= 0 else "negative"
            else:
                change_str = "-"
                change_class = "neutral"
            
            recent_history += f"<tr><td>{date}</td><td>{coverage:.1f}%</td><td class='{change_class}'>{change_str}</td></tr>\n"
        
        recent_history += "</table>\n"
    
    # Generate HTML content
    html_content = f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OneMount Coverage Trend Analysis</title>
    <style>
        body {{
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }}
        .container {{
            background: white;
            border-radius: 8px;
            padding: 30px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }}
        h1 {{
            color: #2c3e50;
            border-bottom: 3px solid #3498db;
            padding-bottom: 10px;
        }}
        .summary {{
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }}
        .metric {{
            background: #f8f9fa;
            padding: 20px;
            border-radius: 6px;
            border-left: 4px solid #3498db;
        }}
        .metric h3 {{
            margin: 0 0 10px 0;
            color: #2c3e50;
        }}
        .metric .value {{
            font-size: 2em;
            font-weight: bold;
            color: {color};
        }}
        .trend {{
            font-size: 1.5em;
            margin: 20px 0;
            padding: 15px;
            background: #e8f4fd;
            border-radius: 6px;
            border-left: 4px solid {color};
        }}
        .history-table {{
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }}
        .history-table th, .history-table td {{
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }}
        .history-table th {{
            background-color: #f2f2f2;
            font-weight: bold;
        }}
        .positive {{ color: #27ae60; }}
        .negative {{ color: #e74c3c; }}
        .neutral {{ color: #7f8c8d; }}
        .plot {{
            text-align: center;
            margin: 30px 0;
        }}
        .plot img {{
            max-width: 100%;
            height: auto;
            border-radius: 6px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }}
        .footer {{
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            color: #7f8c8d;
            text-align: center;
        }}
    </style>
</head>
<body>
    <div class="container">
        <h1>📊 OneMount Coverage Trend Analysis</h1>
        
        <div class="trend">
            {emoji} <strong>Trend:</strong> {description}
        </div>
        
        <div class="summary">
            <div class="metric">
                <h3>Latest Coverage</h3>
                <div class="value">{analysis['latest_coverage']:.1f}%</div>
            </div>
            <div class="metric">
                <h3>Average Coverage</h3>
                <div class="value">{analysis['average_coverage']:.1f}%</div>
            </div>
            <div class="metric">
                <h3>Trend Change</h3>
                <div class="value">{analysis['change']:+.2f}%</div>
            </div>
            <div class="metric">
                <h3>Data Points</h3>
                <div class="value">{len(history_data)}</div>
            </div>
        </div>
        
        {f'<div class="plot"><img src="{Path(plot_path).name}" alt="Coverage Trend Plot"></div>' if plot_path else ''}
        
        <h2>📈 Recent Coverage History</h2>
        {recent_history}
        
        <div class="footer">
            <p>Generated on {datetime.now().strftime('%Y-%m-%d %H:%M:%S')} | OneMount Coverage Analysis Tool</p>
        </div>
    </div>
</body>
</html>
"""
    
    # Write HTML file
    with open(output_path, 'w') as f:
        f.write(html_content)
    
    return True

def main():
    parser = argparse.ArgumentParser(description='Analyze OneMount coverage trends')
    parser.add_argument('--input', required=True, help='Input coverage history JSON file')
    parser.add_argument('--output', required=True, help='Output HTML report file')
    parser.add_argument('--plot', action='store_true', help='Generate trend plot (requires matplotlib)')
    
    args = parser.parse_args()
    
    # Load coverage history
    history_data = load_coverage_history(args.input)
    
    if not history_data:
        print("No coverage history data available.")
        sys.exit(1)
    
    # Analyze trends
    analysis = analyze_trends(history_data)
    
    # Generate plot if requested and possible
    plot_path = None
    if args.plot and HAS_PLOTTING:
        plot_path = generate_plot(history_data, args.output)
        if plot_path:
            print(f"Coverage trend plot generated: {plot_path}")
    
    # Generate HTML report
    if generate_html_report(history_data, analysis, args.output, plot_path):
        print(f"Coverage trend report generated: {args.output}")
        
        # Print summary to console
        print(f"\nCoverage Trend Summary:")
        print(f"  Latest Coverage: {analysis['latest_coverage']:.1f}%")
        print(f"  Trend: {analysis['trend']}")
        print(f"  Change Rate: {analysis['change']:+.2f}%")
        print(f"  Data Points: {len(history_data)}")
    else:
        print("Failed to generate HTML report")
        sys.exit(1)

if __name__ == '__main__':
    main()
