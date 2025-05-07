import json
import os
from collections import Counter

# Load the GitHub issues data
with open('data/github_issues_7MAY25.json', 'r') as f:
    issues = json.load(f)

print(f"Total number of issues: {len(issues)}")

# Print the structure of the first issue to understand the available fields
if issues:
    print("\nStructure of an issue:")
    for key in issues[0].keys():
        print(f"- {key}")

# Analyze titles
print("\nAnalyzing titles...")
title_words = Counter()
title_lengths = []
for issue in issues:
    if 'title' in issue:
        title = issue['title']
        title_lengths.append(len(title))
        words = title.lower().split()
        title_words.update(words)

if title_lengths:
    avg_title_length = sum(title_lengths) / len(title_lengths)
    print(f"Average title length: {avg_title_length:.2f} characters")
    print(f"Most common words in titles: {title_words.most_common(10)}")

# Analyze bodies
print("\nAnalyzing bodies...")
body_lengths = []
has_sections = 0
has_code_blocks = 0
has_lists = 0
has_images = 0

for issue in issues:
    if 'body' in issue and issue['body']:
        body = issue['body']
        body_lengths.append(len(body))
        
        # Check for common sections
        if '## ' in body or '# ' in body:
            has_sections += 1
        
        # Check for code blocks
        if '```' in body:
            has_code_blocks += 1
        
        # Check for lists
        if '- ' in body or '* ' in body or any(line.strip().startswith('1.') for line in body.split('\n')):
            has_lists += 1
        
        # Check for images
        if '![' in body:
            has_images += 1

if body_lengths:
    avg_body_length = sum(body_lengths) / len(body_lengths)
    print(f"Average body length: {avg_body_length:.2f} characters")
    print(f"Issues with sections: {has_sections} ({has_sections/len(issues)*100:.2f}%)")
    print(f"Issues with code blocks: {has_code_blocks} ({has_code_blocks/len(issues)*100:.2f}%)")
    print(f"Issues with lists: {has_lists} ({has_lists/len(issues)*100:.2f}%)")
    print(f"Issues with images: {has_images} ({has_images/len(issues)*100:.2f}%)")

# Extract common sections from bodies
print("\nCommon sections in issue bodies:")
section_counter = Counter()
for issue in issues:
    if 'body' in issue and issue['body']:
        body = issue['body']
        lines = body.split('\n')
        for line in lines:
            line = line.strip()
            if line.startswith('## ') or line.startswith('# '):
                section = line.lstrip('#').strip()
                section_counter[section] += 1

print("Top 10 common sections:")
for section, count in section_counter.most_common(10):
    print(f"- {section}: {count} occurrences")

# Print a sample of a well-structured issue
print("\nSample of a well-structured issue:")
well_structured_issues = [issue for issue in issues if 'body' in issue and issue['body'] and '## ' in issue['body']]
if well_structured_issues:
    sample_issue = max(well_structured_issues, key=lambda x: len(x['body']))
    print(f"Title: {sample_issue.get('title', 'N/A')}")
    print(f"Body preview (first 500 chars):\n{sample_issue.get('body', 'N/A')[:500]}...")