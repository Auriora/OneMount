import json

# Load the GitHub issues data
with open('data/github_issues_7MAY25.json', 'r') as f:
    issues = json.load(f)

if issues:
    # Print all keys from the first issue
    print("All fields in the first issue:")
    for key in issues[0].keys():
        print(f"- {key}: {type(issues[0][key])}")
        
    # Print a sample of the first issue in a more readable format
    print("\nSample issue (first one in the file):")
    sample_issue = issues[0]
    for key, value in sample_issue.items():
        if key == 'body':
            print(f"{key}: [truncated for readability]")
        elif isinstance(value, list) and len(value) > 3:
            print(f"{key}: {value[:3]} ... (truncated)")
        else:
            print(f"{key}: {value}")