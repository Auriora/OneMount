#!/usr/bin/env python3
"""
Test script for implement_github_issue.py.

This script tests that the implement_github_issue.py script generates
a prompt that follows the new format.

Usage:
    python3 test_implement_github_issue.py
"""

import os
import sys
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock

# Add the parent directory to the path so we can import implement_github_issue
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
import implement_github_issue


class TestImplementGitHubIssue(unittest.TestCase):
    """Test cases for implement_github_issue.py."""

    def setUp(self):
        """Set up test fixtures."""
        # Create a mock issue
        self.issue = {
            "number": 117,
            "title": "Implement Testing Recommendations for OneMount",
            "body": "## Description\nImplement testing recommendations to improve the overall quality and reliability of the OneMount test suite.\n\n## Rationale\nThe current test suite has some limitations that affect test reliability and coverage. Implementing these recommendations will improve test quality, reliability, and coverage.\n\n## Impact\nThis implementation will affect the test suite and will improve test quality, reliability, and coverage.\n\n## Relevant Documentation\n- [Test Implementation Execution Plan](../docs/0-project-management/test-implementation-execution-plan.md)\n\n## Implementation Notes\n- Target ≥80% coverage by adding table-driven unit tests\n- Focus on filesystem operations, error conditions, and concurrency scenarios\n- Replace raw goroutines with `context.Context` management and `sync.WaitGroup`\n- Handle cancellations and orderly shutdowns properly\n- Adopt a uniform error-wrapping strategy across modules\n- Leverage Go's `errors` package or a chosen wrapper for clarity and consistency",
            "state": "OPEN",
            "labels": [
                {"name": "enhancement"},
                {"name": "testing"},
            ],
        }

        # Create mock documentation references
        self.doc_references = [
            ("Test Guidelines", "docs/guides/testing/test-guidelines.md"),
            ("Development Guidelines", "docs/DEVELOPMENT.md"),
        ]

    def test_create_junie_prompt(self):
        """Test that create_junie_prompt generates a prompt that follows the new format."""
        # Call the function
        prompt = implement_github_issue.create_junie_prompt(self.issue, self.doc_references)

        # Check that the prompt contains the expected sections
        self.assertIn("# OneMount GitHub Issue Implementation Prompt", prompt)
        self.assertIn("## Overview", prompt)
        self.assertIn("## Workflow Steps", prompt)
        self.assertIn("### 1. Extract and Understand the Issue", prompt)
        self.assertIn("### 2. Open JetBrains Task and Create Branch", prompt)
        self.assertIn("### 3. Review Relevant Documentation", prompt)
        self.assertIn("### 4. Plan Implementation", prompt)
        self.assertIn("### 5. Implementation", prompt)
        self.assertIn("### 6. Testing", prompt)
        self.assertIn("### 7. Update Issue and Commit Changes", prompt)
        self.assertIn("### 8. Final Checklist", prompt)
        self.assertIn("## Command Examples", prompt)
        self.assertIn("## Project-Specific Notes", prompt)
        self.assertIn("## Issue Details", prompt)

        # Check that the prompt contains the issue details
        self.assertIn(f"### Issue #{self.issue['number']}: {self.issue['title']}", prompt)
        self.assertIn("#### Description", prompt)
        self.assertIn("#### Labels", prompt)
        self.assertIn("enhancement, testing", prompt)
        self.assertIn("#### Relevant Documentation", prompt)
        self.assertIn("- [Test Guidelines](docs/guides/testing/test-guidelines.md)", prompt)
        self.assertIn("- [Development Guidelines](docs/DEVELOPMENT.md)", prompt)

        # Since this is a testing-related issue, check that it includes the testing-related additional considerations
        self.assertIn("### For Testing-Related Changes", prompt)
        self.assertIn("This issue involves testing improvements:", prompt)
        self.assertIn("1. What testing frameworks and tools are used in the project?", prompt)

        # Print the prompt for manual inspection
        print(prompt)


if __name__ == "__main__":
    unittest.main()
