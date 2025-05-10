#!/usr/bin/env python3
import subprocess

def get_github_token_from_cli():
    """
    Get the GitHub authentication token using the GitHub CLI.
    
    Returns:
        str: The GitHub authentication token if successful, None otherwise
    """
    try:
        token_process = subprocess.run(['gh', 'auth', 'token'], capture_output=True, text=True)
        if token_process.returncode != 0:
            print(f"Error getting GitHub token: {token_process.stderr}")
            return None
        
        auth_token = token_process.stdout.strip()
        if not auth_token:
            print("GitHub CLI returned an empty token. Make sure you're authenticated with 'gh auth login'")
            return None
            
        return auth_token
    except Exception as e:
        print(f"Error executing GitHub CLI: {e}")
        return None

if __name__ == "__main__":
    print("Testing GitHub CLI token retrieval...")
    token = get_github_token_from_cli()
    if token:
        print(f"Successfully retrieved token: {token[:4]}{'*' * (len(token) - 8)}{token[-4:]}")
    else:
        print("Failed to retrieve token from GitHub CLI")