package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/option"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/ozankasikci/gitai/internal/config"
	"github.com/ozankasikci/gitai/internal/keyring"
	"github.com/ozankasikci/gitai/internal/logger"
)

type CommitSuggestion struct {
	Message     string
	Explanation string
}

type CommitMessageGenerator interface {
	GenerateCommitSuggestions(changes string) ([]CommitSuggestion, error)
}

type AnthropicClient struct {
	client *anthropic.Client
}

type SuggestionsMsg struct {
	Suggestions []CommitSuggestion
}

func NewAnthropicClient() (*AnthropicClient, error) {
	// Get API key from keyring
	apiKey, err := keyring.GetAPIKey(keyring.Anthropic)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from keyring: %w", err)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is not configured")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicClient{client: client}, nil
}

func (c *AnthropicClient) GenerateCommitSuggestions(changes string) ([]CommitSuggestion, error) {
	logger.Debugf("\n=== Input changes string ===\nLength: %d\nContent:\n%s\n", len(changes), changes)

	// Format the changes to include both summary and diff content
	formattedChanges := "=== File Changes Summary ===\n"

	// First, extract and format the file status lines
	logger.Debugf("\n=== Processing file status lines ===\n")
	for _, line := range strings.Split(changes, "\n") {
		if strings.Contains(line, "(status:") {
			logger.Debugf("Found status line: %s", line)
			formattedChanges += line + "\n"
		}
	}

	// Then add the full diff content with clear separation
	formattedChanges += "\n=== Git Diff Content ===\n"
	formattedChanges += changes

	prompt := buildPrompt(formattedChanges)

	logger.Debugf("\n=== Final formatted changes ===\n%s\n", formattedChanges)
	logger.Debugf("\n=== Full prompt being sent to LLM ===\n%s\n", prompt)

	cfg := config.Get()
	msg, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.Model(cfg.LLM.Anthropic.Model)),
		MaxTokens: anthropic.F(cfg.LLM.Anthropic.MaxTokens),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		logger.Errorf("Error from LLM: %v", err)
		return nil, fmt.Errorf("failed to generate commit message: %w", err)
	}

	var responseText string
	for _, content := range msg.Content {
		if content.Type == "text" {
			responseText = content.Text
			logger.Debugf("\n=== Response from LLM ===\n%s\n", responseText)
			logger.Debugf("\n=== Raw LLM Response ===\n%#v\n", responseText)
			break
		}
	}

	if responseText == "" {
		logger.Errorf("No text content found in LLM response")
		return nil, fmt.Errorf("no text content in response")
	}

	suggestions := parseResponse(responseText)
	for i, suggestion := range suggestions {
		logger.Debugf("Suggestion %d:\nMessage: %s\nExplanation: %s\n",
			i+1, suggestion.Message, suggestion.Explanation)
	}

	return suggestions, nil
}

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
