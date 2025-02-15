import google.generativeai as genai
import os
from github import Github

def get_pr_diff(repo_name, pr_number, github_token):
    """Retrieves and cleans the PR diff using PyGithub."""
    g = Github(github_token)
    repo = g.get_repo(repo_name)
    pr = repo.get_pull(pr_number)
    diff = pr.get_commits().files.patch  # Get diff from the latest commit
    return diff

def generate_gemini_review(diff, api_key):
    """Generates a code review using the Gemini API."""
    genai.configure(api_key=api_key)
    model = genai.GenerativeModel('gemini-pro')

    prompt = f"""
    Review the following code diff and provide feedback. Point out potential issues,
    suggest improvements, and highlight good practices. Keep the review concise.

    ```diff
    {diff}
    ```
    """
    response = model.generate_content(prompt)
    return response.text if response.text else None

def post_github_comment(repo_name, pr_number, comment, github_token):
    """Posts a comment to a GitHub pull request."""
    g = Github(github_token)
    repo = g.get_repo(repo_name)
    pr = repo.get_pull(pr_number)
    pr.create_issue_comment(comment)
    print("Review comment posted successfully.")

def main():
    """Main function to orchestrate the Gemini PR review."""
    api_key = os.environ.get('GEMINI_API_KEY')
    pr_number = int(os.environ.get('PR_NUMBER'))
    repo_name = os.environ.get('GITHUB_REPOSITORY')
    github_token = os.environ.get('GITHUB_TOKEN')

    diff = get_pr_diff(repo_name, pr_number, github_token)  # Get diff directly

    review_comment = generate_gemini_review(diff, api_key)

    if review_comment:
        post_github_comment(repo_name, pr_number, review_comment, github_token)
    else:
        print("Gemini API returned no response.")

if __name__ == "__main__":
    main()
