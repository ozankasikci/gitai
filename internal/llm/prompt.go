package llm

import (
	"fmt"
)

func buildPrompt(changes string) string {
	return fmt.Sprintf(`
You are a highly intelligent assistant skilled in understanding code changes. I will provide you with a git diff. Your task is to analyze the changes and generate a concise and descriptive commit message that:

Summarizes the purpose of ALL changes across ALL files.
Creates a unified message that captures the overall intent of the changes.
If there are multiple types of changes, use the most significant one as the primary message.

Pay special attention to:
- Added files (new functionality or features)
- Modified files (improvements or fixes)
- Deleted files (cleanup or removals)

When multiple files are changed:
- Look for patterns across the changes
- Identify the primary purpose of the changes
- Consider if changes are related (e.g., refactoring across files)

Analyze the following git diff and generate 3 different commit messages.

Format each suggestion exactly like this example:
1 - Add user authentication
Explanation: Implements basic user authentication

2 - Fix database connection issues
Explanation: Fixes connection pooling issues

Follow these git commit message rules:
1. Use imperative mood ("Add" not "Added" or "Adds")
2. First line should be 50 chars or less
3. First line should be capitalized
4. No period at the end of the first line

Optionally, you can use these Conventional Commits prefixes if appropriate:
- feat: new feature
- fix: bug fix
- docs: documentation only
- style: formatting
- refactor: code change that neither fixes a bug nor adds a feature
- test: adding missing tests
- chore: maintain

Changes:
%s

Remember to format each suggestion exactly like the example above.
`, changes)
} 