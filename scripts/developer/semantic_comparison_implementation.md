# Semantic Issue Comparison Implementation

## Overview

This document describes a new approach for comparing GitHub issues semantically using a Large Language Model (LLM) like GPT-4. The approach is designed to understand the intent and principles behind each issue, rather than just comparing text similarity.

## Limitations of the Previous Approach

The previous approach used a simple text-based comparison method:
1. It compared issues by title to identify potential duplicates
2. For issues with matching titles, it compared the body content to determine if it's an exact duplicate
3. This approach only identified exact duplicates with identical text

This method has several limitations:
- It can't identify semantic duplicates with different wording
- It doesn't understand the intent or principles behind the issues
- It misses duplicates that describe the same problem in different ways
- It can't identify issues that address the same core functionality but with different implementation details

## New Approach: Semantic Comparison with LLM

The new approach uses an LLM (like GPT-4) to understand the semantic meaning of issues:

1. **Extract Issue Content**: Extract the title and body from each issue
2. **Semantic Analysis**: Use an LLM to analyze the intent, principles, and core functionality described in each issue
3. **Similarity Scoring**: Assign a similarity score (0-100) based on how semantically similar the issues are
4. **Categorization**: Categorize issues as duplicates or similar based on the similarity score
5. **Detailed Explanation**: Provide a detailed explanation of why issues are or are not duplicates

### Benefits of the LLM Approach

- **Understanding Intent**: The LLM can understand the intent behind an issue, not just the text
- **Contextual Understanding**: The LLM can understand the context of the issue within the software development domain
- **Flexible Matching**: The LLM can identify duplicates even when they use different terminology
- **Detailed Explanations**: The LLM provides detailed explanations of why issues are or are not duplicates
- **Confidence Scoring**: The LLM provides a similarity score to indicate the confidence of the match

## Implementation Details

A Python script (`semantic_issue_comparison.py`) has been created to implement this approach. The script:

1. Loads issues from JSON files
2. Extracts the title and body from each issue
3. Uses the OpenAI API to compare issues semantically
4. Categorizes issues as duplicates or similar based on a similarity score
5. Generates a detailed report of the findings
6. Saves the results to JSON files for further analysis

### Requirements

- Python 3.6+
- OpenAI API key
- Required Python packages:
  - `openai`
  - `tqdm`

### Installation

```bash
# Install required packages
pip install openai tqdm
```

### Usage

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-api-key"

# Run the script with default settings
python scripts/developer/semantic_issue_comparison.py

# Run with custom settings
python scripts/developer/semantic_issue_comparison.py \
  --new-issues path/to/new_issues.json \
  --existing-issues path/to/existing_issues.json \
  --threshold 75
```

### Output Files

The script generates the following output files:

1. `data/semantic_duplicates.json`: Contains issues identified as semantic duplicates
2. `data/similar_issues.json`: Contains issues identified as similar but not duplicates
3. `data/semantic_issue_comparison_report.md`: A detailed report of the findings
4. `semantic_comparison.log`: A log file with detailed information about the comparison process

## Example Prompt for Semantic Comparison

The script uses the following prompt to compare issues semantically:

```
You are an expert software developer tasked with identifying duplicate issues in a GitHub repository. You need to determine if two issues are semantically similar or duplicates of each other, even if they have different wording or formatting.

Please compare these two issues and determine if they are duplicates or semantically similar:

ISSUE 1:
[Title and body of the first issue]

ISSUE 2:
[Title and body of the second issue]

Analyze the intent, principles, and core functionality described in both issues.
Return your response in the following JSON format:
{
    "is_duplicate": true/false,
    "similarity_score": 0-100,
    "explanation": "Your detailed explanation of why these issues are or are not duplicates"
}

A similarity score of:
- 0-30: Not similar
- 31-70: Somewhat similar but addressing different aspects
- 71-100: Very similar or duplicate issues
```

## Conclusion

This semantic comparison approach provides a more sophisticated way to identify duplicate issues by understanding the intent and principles behind each issue. It leverages the power of LLMs to perform semantic analysis that goes beyond simple text comparison.

By implementing this approach, we can more effectively identify duplicate issues and reduce the manual effort required to review and triage issues.