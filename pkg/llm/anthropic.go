package llm

import (
	"context"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go/option"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/ozankasikci/gitai/pkg/logger"
)

type CommitSuggestion struct {
	Message     string
	Explanation string
}

type CommitMessageGenerator interface {
	GenerateCommitSuggestions(changes string) ([]CommitSuggestion, error)
}

type Client struct {
	client *anthropic.Client
}

func NewClient() (*Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable is not set")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Client{client: client}, nil
}

func (c *Client) GenerateCommitSuggestions(changes string) ([]CommitSuggestion, error) {
	prompt := buildPrompt(changes)

	logger.Debugf("\n=== Changes being sent to LLM ===\n%s\n", changes)
	logger.Debugf("\n=== Full prompt being sent to LLM ===\n%s\n", prompt)

	msg, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.Model("claude-3-5-sonnet-20241022")),
		MaxTokens: anthropic.F(int64(1024)),
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
			break
		}
	}

	suggestions := parseResponse(responseText)
	logger.Debugf("\n=== Parsed Suggestions ===\n")
	for i, suggestion := range suggestions {
		logger.Debugf("Suggestion %d:\nMessage: %s\nExplanation: %s\n",
			i+1, suggestion.Message, suggestion.Explanation)
	}

	return suggestions, nil
}

func buildPrompt(changes string) string {
	return fmt.Sprintf(`
	You are a highly intelligent assistant skilled in understanding code changes. I will provide you with a git diff. Your task is to analyze the changes and generate a concise and descriptive commit message that:

Summarizes the purpose of the changes.
Highlights any key modifications or additions.
	
	Analyze the following git diff and generate 3 different commit messages.

First, carefully analyze the diff:
- Lines starting with '-' show REMOVED content
- Lines starting with '+' show ADDED content
- Context lines (without + or -) show where in the file the change occurs
- Pay attention to the file paths and component names
- For each change, compare the old and new versions to understand what changed

Follow these git commit message rules:
1. Use imperative mood ("Add" not "Added" or "Adds")
2. First line should be 50 chars or less
3. First line should be capitalized
4. No period at the end of the first line
5. Leave second line blank
6. Wrap subsequent lines at 72 characters

Use the appropriate Conventional Commits prefix based on the diff analysis:
- feat: new feature (entirely new functionality)
- fix: bug fix (correcting incorrect behavior)
- docs: documentation only
- style: formatting, missing semi colons, etc
- refactor: code change that neither fixes a bug nor adds a feature
- test: adding missing tests
- chore: maintain

Changes:
%s

Format each suggestion as:
<number> - <commit message>`, changes)
}

func parseResponse(response string) []CommitSuggestion {
	var suggestions []CommitSuggestion
	lines := strings.Split(response, "\n")

	var currentSuggestion *CommitSuggestion
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if this is a new suggestion line (starts with a number)
		if len(line) > 2 && line[0] >= '1' && line[0] <= '9' && line[1] == ' ' {
			if currentSuggestion != nil {
				suggestions = append(suggestions, *currentSuggestion)
			}
			currentSuggestion = &CommitSuggestion{
				Message: strings.TrimSpace(strings.SplitN(line, "-", 2)[1]), // Remove number and dash
			}
		} else if strings.HasPrefix(strings.ToLower(line), "explanation:") {
			if currentSuggestion != nil {
				currentSuggestion.Explanation = strings.TrimSpace(strings.TrimPrefix(line, "Explanation:"))
			}
		}
	}

	// Add the last suggestion if exists
	if currentSuggestion != nil {
		suggestions = append(suggestions, *currentSuggestion)
	}

	return suggestions
}
