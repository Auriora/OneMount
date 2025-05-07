#!/usr/bin/env python3
"""
Semantic Issue Comparison Script

This script compares GitHub issues semantically using a Large Language Model (LLM) to understand
the intent and principles behind each issue, rather than just comparing text similarity.

For detailed documentation, see:
- scripts/developer/semantic_comparison_implementation.md
- scripts/developer/semantic_comparison_summary.md

Key Features:
- Semantic Understanding: Uses an LLM to understand the intent behind issues
- Similarity Scoring: Assigns a score (0-100) based on semantic similarity
- Detailed Explanations: Provides explanations of why issues are or are not duplicates
- Flexible Matching: Identifies duplicates even with different terminology
- Optimization: Includes optimizations to reduce API calls

Usage:
    export OPENAI_API_KEY="your-api-key"
    ./scripts/developer/semantic_issue_comparison.py

Output Files:
    - data/semantic_duplicates.json: Issues identified as semantic duplicates
    - data/similar_issues.json: Issues identified as similar but not duplicates
    - data/semantic_issue_comparison_report.md: Detailed report of findings
    - semantic_comparison.log: Log file with comparison process details

Requirements:
    - Python 3.6+
    - OpenAI API key
    - Required packages: openai, tqdm
"""
import json
import os
import argparse
import time
from datetime import datetime
import logging
import sys

# Check for required packages
required_packages = ['openai', 'tqdm']
missing_packages = []

for package in required_packages:
    try:
        __import__(package)
    except ImportError:
        missing_packages.append(package)

if missing_packages:
    print(f"Error: Missing required packages: {', '.join(missing_packages)}")
    print("Please install them using:")
    print(f"pip install {' '.join(missing_packages)}")
    sys.exit(1)

import openai
from tqdm import tqdm

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.FileHandler("semantic_comparison.log"),
        logging.StreamHandler()
    ]
)
logger = logging.getLogger("semantic_issue_comparison")

def load_json_file(file_path):
    """Load a JSON file and return its contents."""
    with open(file_path, 'r') as f:
        return json.load(f)

def save_json_file(data, file_path):
    """Save data to a JSON file with pretty formatting."""
    with open(file_path, 'w') as f:
        json.dump(data, f, indent=2)

def extract_issue_content(issue):
    """
    Extract the relevant content from an issue for semantic comparison.

    This is the first step in the semantic comparison process, where we extract
    the title and body from each issue to prepare for LLM analysis.

    Returns:
        str: A formatted string containing the issue's title and body
    """
    title = issue.get('title', '')
    body = issue.get('body', '')

    return f"Title: {title}\n\nBody:\n{body}"

def compare_issues_semantically(new_issue_content, existing_issue_content, api_key):
    """
    Compare two issues semantically using OpenAI's API.

    This is the core function that implements the semantic analysis using an LLM.
    It sends the issue content to the OpenAI API and asks the model to analyze
    the intent, principles, and core functionality described in both issues.

    The LLM provides:
    - A determination of whether the issues are duplicates
    - A similarity score (0-100)
    - A detailed explanation of why the issues are or are not duplicates

    Args:
        new_issue_content (str): Content of the new issue
        existing_issue_content (str): Content of the existing issue
        api_key (str): OpenAI API key

    Returns:
        dict: Comparison result with similarity score and explanation
    """
    openai.api_key = api_key

    try:
        response = openai.ChatCompletion.create(
            model="gpt-4",
            messages=[
                {"role": "system", "content": "You are an expert software developer tasked with identifying duplicate issues in a GitHub repository. You need to determine if two issues are semantically similar or duplicates of each other, even if they have different wording or formatting."},
                {"role": "user", "content": f"""
                Please compare these two issues and determine if they are duplicates or semantically similar:

                ISSUE 1:
                {new_issue_content}

                ISSUE 2:
                {existing_issue_content}

                Analyze the intent, principles, and core functionality described in both issues.
                Return your response in the following JSON format:
                {{
                    "is_duplicate": true/false,
                    "similarity_score": 0-100,
                    "explanation": "Your detailed explanation of why these issues are or are not duplicates"
                }}

                A similarity score of:
                - 0-30: Not similar
                - 31-70: Somewhat similar but addressing different aspects
                - 71-100: Very similar or duplicate issues

                Only return the JSON object, nothing else.
                """}
            ],
            temperature=0.3,
            max_tokens=1000
        )

        result = json.loads(response.choices[0].message.content)
        return result

    except Exception as e:
        logger.error(f"Error calling OpenAI API: {str(e)}")
        # Return a default result in case of error
        return {
            "is_duplicate": False,
            "similarity_score": 0,
            "explanation": f"Error during comparison: {str(e)}"
        }

def compare_all_issues(new_issues, existing_issues, api_key, similarity_threshold=70):
    """
    Compare all new issues with existing issues to identify duplicates.

    This function orchestrates the comparison process by:
    1. Iterating through all new issues
    2. Comparing each new issue with existing issues
    3. Categorizing issues as duplicates or similar based on similarity scores
    4. Optimizing performance by skipping comparisons when titles are completely different

    The function implements the categorization step of the semantic comparison approach,
    using the similarity threshold to determine if issues are duplicates or just similar.

    Args:
        new_issues (list): List of new issues
        existing_issues (list): List of existing issues
        api_key (str): OpenAI API key
        similarity_threshold (int): Threshold for considering issues as duplicates (default: 70)

    Returns:
        tuple: (duplicates, similar_issues) - Lists of duplicate and similar issues with comparison results
    """
    duplicates = []
    similar_issues = []

    for new_issue in tqdm(new_issues, desc="Comparing issues"):
        new_issue_content = extract_issue_content(new_issue)

        # Check against all existing issues
        for existing_issue in existing_issues:
            existing_issue_content = extract_issue_content(existing_issue)

            # Skip if titles are completely different (optional optimization)
            if not any(word in existing_issue.get('title', '').lower() 
                      for word in new_issue.get('title', '').lower().split() 
                      if len(word) > 3):
                continue

            # Perform semantic comparison
            result = compare_issues_semantically(new_issue_content, existing_issue_content, api_key)

            # Log the comparison result
            logger.info(f"Compared '{new_issue.get('title', '')}' with '{existing_issue.get('title', '')}': "
                       f"Score: {result.get('similarity_score', 0)}, "
                       f"Is duplicate: {result.get('is_duplicate', False)}")

            # Add to appropriate list based on similarity score
            if result.get('is_duplicate', False) or result.get('similarity_score', 0) >= similarity_threshold:
                duplicates.append({
                    'new_issue': new_issue,
                    'existing_issue': existing_issue,
                    'comparison_result': result
                })
                break  # Stop comparing this new issue once a duplicate is found
            elif result.get('similarity_score', 0) >= 50:  # Threshold for similar but not duplicate
                similar_issues.append({
                    'new_issue': new_issue,
                    'existing_issue': existing_issue,
                    'comparison_result': result
                })

        # Add a small delay to avoid rate limiting
        time.sleep(1)

    return duplicates, similar_issues

def create_markdown_report(new_issues, duplicates, similar_issues):
    """
    Create a markdown report of findings and actions.

    This function generates a detailed report that summarizes the results of the
    semantic comparison process. The report includes:
    - A summary of the total issues analyzed and results
    - A list of actions taken
    - Details of each duplicate issue with similarity scores and explanations
    - Details of each similar issue with similarity scores and explanations

    This is the final step in the semantic comparison process, providing
    human-readable documentation of the results.

    Args:
        new_issues (list): List of all new issues analyzed
        duplicates (list): List of duplicate issues with comparison results
        similar_issues (list): List of similar issues with comparison results

    Returns:
        str: Markdown-formatted report of findings and actions
    """
    now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")

    report = f"""# Semantic Issue Comparison Report
Generated: {now}

## Summary
- Total new issues analyzed: {len(new_issues)}
- Semantic duplicates found: {len(duplicates)}
- Similar issues (not duplicates): {len(similar_issues)}

## Actions Taken
1. Moved {len(duplicates)} semantic duplicate issues to `data/semantic_duplicates.json`
2. Moved {len(similar_issues)} similar issues to `data/similar_issues.json` for review

## Semantic Duplicates
The following issues were identified as semantic duplicates and moved to `data/semantic_duplicates.json`:

"""

    for item in duplicates:
        report += f"### {item['new_issue'].get('title', 'No Title')}\n"
        report += f"**Matched with:** {item['existing_issue'].get('title', 'No Title')}\n"
        report += f"**Similarity Score:** {item['comparison_result'].get('similarity_score', 0)}\n"
        report += f"**Explanation:** {item['comparison_result'].get('explanation', 'No explanation provided')}\n\n"

    report += "\n## Similar Issues\n"
    report += "The following issues were identified as similar (but not duplicates) and moved to `data/similar_issues.json` for review:\n\n"

    for item in similar_issues:
        report += f"### {item['new_issue'].get('title', 'No Title')}\n"
        report += f"**Similar to:** {item['existing_issue'].get('title', 'No Title')}\n"
        report += f"**Similarity Score:** {item['comparison_result'].get('similarity_score', 0)}\n"
        report += f"**Explanation:** {item['comparison_result'].get('explanation', 'No explanation provided')}\n\n"

    return report

def main():
    """
    Main function to run the semantic issue comparison.

    This function orchestrates the entire semantic comparison process:
    1. Parses command-line arguments for customization
    2. Loads issues from JSON files
    3. Compares issues semantically using the LLM-based approach
    4. Saves results to output files
    5. Generates a detailed markdown report

    The process implements the semantic comparison approach described in:
    - scripts/developer/semantic_comparison_implementation.md
    - scripts/developer/semantic_comparison_summary.md

    Returns:
        int: Exit code (0 for success, 1 for error)
    """
    parser = argparse.ArgumentParser(description="Compare issues semantically using OpenAI's API")
    parser.add_argument("--new-issues", default="data/test_implementation_details_issue.json",
                        help="Path to JSON file containing new issues")
    parser.add_argument("--existing-issues", default="data/github_issues_7MAY25.json",
                        help="Path to JSON file containing existing issues")
    parser.add_argument("--api-key", help="OpenAI API key (or set OPENAI_API_KEY environment variable)")
    parser.add_argument("--threshold", type=int, default=70,
                        help="Similarity threshold for considering issues as duplicates (0-100)")
    args = parser.parse_args()

    # Get API key from args or environment variable
    api_key = args.api_key or os.environ.get("OPENAI_API_KEY")
    if not api_key:
        logger.error("OpenAI API key not provided. Use --api-key or set OPENAI_API_KEY environment variable.")
        return 1

    # File paths
    new_issues_path = args.new_issues
    existing_issues_path = args.existing_issues
    duplicates_path = "data/semantic_duplicates.json"
    similar_issues_path = "data/similar_issues.json"
    report_path = "data/semantic_issue_comparison_report.md"

    # Load the JSON files
    logger.info(f"Loading new issues from {new_issues_path}...")
    new_issues = load_json_file(new_issues_path)

    logger.info(f"Loading existing issues from {existing_issues_path}...")
    existing_issues = load_json_file(existing_issues_path)

    logger.info(f"Comparing {len(new_issues)} new issues with {len(existing_issues)} existing issues...")
    duplicates, similar_issues = compare_all_issues(
        new_issues, existing_issues, api_key, args.threshold
    )

    # Save duplicates to semantic_duplicates.json
    logger.info(f"Saving {len(duplicates)} semantic duplicates to {duplicates_path}...")
    save_json_file(duplicates, duplicates_path)

    # Save similar issues to similar_issues.json
    logger.info(f"Saving {len(similar_issues)} similar issues to {similar_issues_path}...")
    save_json_file(similar_issues, similar_issues_path)

    # Create and save the markdown report
    logger.info(f"Creating markdown report at {report_path}...")
    report = create_markdown_report(new_issues, duplicates, similar_issues)
    with open(report_path, 'w') as f:
        f.write(report)

    logger.info("Semantic issue comparison completed successfully!")
    return 0

if __name__ == "__main__":
    exit(main())
