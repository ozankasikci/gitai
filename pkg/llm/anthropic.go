package llm

import (
	"context"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go/option"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/ozankasikci/commit-ai/pkg/logger"
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
	
	logger.Debug.Printf("Changes being sent to LLM:\n%s", changes)
	logger.Debug.Printf("Full prompt being sent to LLM:\n%s", prompt)

	msg, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.Model("claude-3-5-sonnet-20241022")),
		MaxTokens: anthropic.F(int64(1024)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		logger.Error.Printf("Error from LLM: %v", err)
		return nil, fmt.Errorf("failed to generate commit message: %w", err)
	}

	var responseText string
	for _, content := range msg.Content {
		if content.Type == "text" {
			responseText = content.Text
			logger.Debug.Printf("Response from LLM:\n%s", responseText)
			break
		}
	}

	return parseResponse(responseText), nil
}

func buildPrompt(changes string) string {
	return fmt.Sprintf(`Generate 3 different commit messages for the following changes following these strict git commit best practices:

1. Use imperative mood ("Add" not "Added" or "Adds")
2. First line should be 50 chars or less
3. First line should be capitalized
4. No period at the end of the first line
5. Leave second line blank
6. Wrap subsequent lines at 72 characters
7. Use the body to explain what and why vs. how

Optionally use the Conventional Commits format (type(scope): description) if the change fits one of these types:
- feat: new feature
- fix: bug fix
- docs: documentation only
- style: formatting, missing semi colons, etc
- refactor: code change that neither fixes a bug nor adds a feature
- test: adding missing tests
- chore: maintain

If the change doesn't fit these types, write a direct descriptive message without a type prefix.

Changes:
%s

Format each suggestion as:
<number> - <commit message>
Explanation: <why this message is appropriate, focusing on motivation and impact>`, changes)
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
