import google.generativeai as genai
import os
from github import Github

api_key = os.environ.get('GEMINI_API_KEY')
pr_number = int(os.environ.get('PR_NUMBER'))
repo_name = os.environ.get('GITHUB_REPOSITORY')
diff = os.environ.get('INPUT_DIFF') or ""
diff = diff.replace("DIFF<<EOF", "").replace("EOF", "").strip()

genai.configure(api_key=api_key)

model = genai.GenerativeModel('gemini-pro')

prompt = f"""
Review the following code diff and provide feedback. Point out potential issues,
suggest improvements, and highlight good practices. Keep the review concise.

```diff
{diff}
