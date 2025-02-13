#!/usr/bin/env python3
import os
import json
import requests
from google import genai


def gather_markdown_files(file_paths):
    combined = ""
    for file in file_paths:
        if os.path.exists(file):
            try:
                with open(file, 'r') as f:
                    content = f.read()
                combined += f"\n\n---\nContent of {file}:\n{content}\n"
            except Exception as e:
                print(f"Error reading {file}: {e}")
        else:
            print(f"File {file} not found, skipping.")
    return combined


def post_comment_to_pr(repository, pr_number, comment_body, github_token):
    url = f"https://api.github.com/repos/{repository}/issues/{pr_number}/comments"
    headers = {
        "Authorization": f"token {github_token}",
        "Content-Type": "application/json",
        "Accept": "application/vnd.github.v3+json"
    }
    payload = {"body": comment_body}
    response = requests.post(url, headers=headers, json=payload)
    if response.status_code in (200, 201):
        print("Successfully posted comment to PR.")
    else:
        print(f"Failed to post comment to PR (status code {response.status_code}): {response.text}")


def main():
    # Define the markdown files to be gathered
    # file_paths = ["README.md", "docs/api-guidelines.md", "docs/api-changes.md"]  
    file_paths = ["README.md"]
    docs = gather_markdown_files(file_paths)

    # Retrieve environment variables for PR data and Gemini API credentials
    pr_number = os.environ.get("PR_NUMBER")
    pr_title = os.environ.get("PR_TITLE")
    pr_body = os.environ.get("PR_BODY")
    gemini_api_key = os.environ.get("GEMINI_API_KEY")  # Now used with google-genai

    # Retrieve GitHub info for posting a comment
    github_token = os.environ.get("GITHUB_TOKEN")
    repository = os.environ.get("GITHUB_REPOSITORY")

    if not all([pr_number, pr_title, pr_body, gemini_api_key]):
        print("Error: One or more required environment variables (PR_NUMBER, PR_TITLE, PR_BODY, GEMINI_API_KEY) are missing.")
        return
    
    # TODO: Add logic to check the diff of the PR and the actual code changes

    # Build a content string for Gemini API
    contents = (
        f"PR Number: {pr_number}\n"
        f"Title: {pr_title}\n"
        f"Body: {pr_body}\n"
        f"Documentation:\n{docs}"
    )

    try:
        print("Sending payload to Gemini API using google-genai client...")
        client = genai.Client(api_key=gemini_api_key)
        response = client.models.generate_content(
            model="gemini-2.0-flash",
            contents=contents
        )
        print(f"Gemini API response: {response.text}")

        # If GitHub credentials are provided, post a comment with the Gemini API response
        if github_token and repository:
            comment_body = (
                "Gemini API Review Result:\n\n"
                f"{response.text}"
            )
            post_comment_to_pr(repository, pr_number, comment_body, github_token)
        else:
            print("GitHub token or repository environment variable is missing; skipping posting comment on PR.")
    except Exception as e:
        print(f"An error occurred while calling Gemini API: {e}")


if __name__ == "__main__":
    main() 