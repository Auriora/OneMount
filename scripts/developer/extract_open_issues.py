import json

# Path to the JSON file
json_file_path = 'data/github_issues.json'

# Read the JSON file
with open(json_file_path, 'r') as file:
    issues_data = json.load(file)

# Print the structure of the first issue to understand the format
if issues_data and len(issues_data) > 0:
    first_issue = issues_data[0]
    print("First issue structure:")
    for key, value in first_issue.items():
        if isinstance(value, (str, int, bool)) or value is None:
            print(f"  {key}: {value}")
        else:
            print(f"  {key}: {type(value)}")

    # Check if 'state' field exists
    if 'state' in first_issue:
        print(f"\nState field exists with value: {first_issue['state']}")
    else:
        print("\nState field does not exist in the first issue")

    # Print the first few issues and their states if available
    print("\nFirst 5 issues:")
    for i, issue in enumerate(issues_data[:5]):
        state = issue.get('state', 'N/A')
        title = issue.get('title', 'N/A')
        number = issue.get('number', 'N/A')
        print(f"  Issue #{number}: {title} (State: {state})")
else:
    print("No issues found in the JSON file")

# Filter open issues (case insensitive)
open_issues = [issue for issue in issues_data if issue.get('state', '').upper() == 'OPEN']

# Create a mapping of task descriptions to issue numbers
task_to_issue = {}
issue_details = {}

print("\nOpen issues found:")
for issue in open_issues:
    issue_number = issue.get('number')
    issue_title = issue.get('title', '')
    issue_body = issue.get('body', '')

    # Print issue details for debugging
    print(f"Issue #{issue_number}: {issue_title}")

    # Store issue details for later use
    issue_details[issue_number] = {
        'title': issue_title,
        'body': issue_body
    }

    # Extract task description from title
    task_description = issue_title

    # Add to mapping
    task_to_issue[task_description] = issue_number

# Print the mapping for reference
print("\nTask to Issue Mapping:")
for task, issue_number in task_to_issue.items():
    print(f"Task: {task} -> Issue #{issue_number}")

print(f"\nTotal open issues found: {len(open_issues)}")

# Now let's try to match these issues with tasks in the test-implementation-execution-plan.md
# This is a simple approach that looks for keywords in the issue title/body that match task descriptions
# A more sophisticated approach would be to use NLP or other techniques

# Define keywords for each phase/task in the test-implementation-execution-plan.md
task_keywords = {
    "Implement Enhanced Resource Management": ["resource management", "filesystem resource", "cleanup"],
    "Implement Signal Handling": ["signal handling", "cleanup", "interrupted"],
    "Fix Upload API Race Condition": ["upload api", "race condition", "waitforupload"],
    "Implement Basic TestFramework Structure": ["testframework", "framework structure"],
    "Implement File Utilities": ["file utilities", "file creation", "verification"],
    "Implement Asynchronous Utilities": ["asynchronous utilities", "waiting", "retrying", "timeouts"],
    "Enhance Graph API Test Fixtures": ["graph api", "test fixtures", "driveitem fixtures"],
    "Set Up Basic Mock Providers": ["mock provider", "mockgraphprovider", "mockfilesystemprovider"],
    "Implement Graph API Mocks with Recording": ["graph api mocks", "recording", "request/response"],
    "Implement Filesystem Mocks": ["filesystem mocks", "mock filesystem"],
    "Add Network Condition Simulation": ["network condition", "latency", "bandwidth", "connection"],
    "Implement Specialized Framework Extensions": ["specialized framework", "graphtestframework"],
    "Implement Environment Validation": ["environment validation", "environmentvalidator"],
    "Implement Enhanced Network Simulation": ["network simulation", "intermittent connections"],
    "Set Up Integration Test Environment": ["integration test", "test environment"],
    "Implement Scenario-Based Testing": ["scenario-based", "test scenarios"],
    "Add Basic Performance Benchmarking": ["performance benchmarking", "performance metrics"],
    "Implement Dmelfa Generator": ["dmelfa generator", "test files", "dna sequence"],
    "Integrate with TestFramework": ["integrate", "testframework", "registration"],
    "Merge Old Test Cases": ["merge", "test cases", "old tests"],
    "Add Advanced Coverage Reporting": ["coverage reporting", "coverage analysis"],
    "Implement Load Testing": ["load testing", "load test scenarios"],
    "Add Performance Metrics Collection": ["performance metrics", "performance trending"],
    "Implement Test Type-Specific Frameworks": ["test type", "unit testing", "integration testing"],
    "Create Comprehensive Documentation": ["documentation", "test utilities"],
    "Implement Enhanced Timeout Management": ["timeout management", "timeoutstrategy"],
    "Implement Flexible Authentication Handling": ["authentication handling", "authenticationprovider"],
    "Create Example Tests": ["example tests", "usage", "utilities"],
    "Create Test Framework Documentation": ["test framework documentation", "api documentation"],
    "Create Test Writing Guidelines": ["test writing guidelines", "best practices"],
    "Create Training Materials": ["training materials", "tutorials", "exercises"]
}

# Try to match issues to tasks
print("\nMatching issues to tasks in test-implementation-execution-plan.md:")
matched_tasks = {}

for task_name, keywords in task_keywords.items():
    for issue_num, details in issue_details.items():
        title = details['title'].lower()
        body = details['body'].lower()

        # Check if any keyword is in the title or body
        if any(keyword.lower() in title or keyword.lower() in body for keyword in keywords):
            if task_name not in matched_tasks:
                matched_tasks[task_name] = []
            matched_tasks[task_name].append(issue_num)
            print(f"Task '{task_name}' matched with Issue #{issue_num}: {details['title']}")

# Print summary of matches
print("\nSummary of matches:")
for task_name, issue_nums in matched_tasks.items():
    print(f"Task: {task_name} -> Issues: {', '.join([f'#{num}' for num in issue_nums])}")

# Print tasks without matches
unmatched_tasks = set(task_keywords.keys()) - set(matched_tasks.keys())
if unmatched_tasks:
    print("\nTasks without matching issues:")
    for task in unmatched_tasks:
        print(f"- {task}")

# Print issues without matches
matched_issue_nums = set()
for issues in matched_tasks.values():
    matched_issue_nums.update(issues)

unmatched_issues = set(issue_details.keys()) - matched_issue_nums
if unmatched_issues:
    print("\nIssues without matching tasks:")
    for issue_num in unmatched_issues:
        print(f"- Issue #{issue_num}: {issue_details[issue_num]['title']}")
