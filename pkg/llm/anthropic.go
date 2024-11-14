package llm

import (
	"context"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go/option"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

type CommitSuggestion struct {
	Message     string
	Explanation string
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

	msg, err := c.client.Messages.New(context.Background(), anthropic.MessageNewParams{
		Model:     anthropic.F(anthropic.Model("claude-3-5-sonnet-20241022")),
		MaxTokens: anthropic.F(int64(1024)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Get the text content from the response
	var responseText string
	for _, content := range msg.Content {
		if content.Type == "text" {
			responseText = content.Text
			break
		}
	}

	return parseResponse(responseText), nil
}

func buildPrompt(changes string) string {
	return fmt.Sprintf(`Generate 3 different commit messages for the following changes. 
Follow the Conventional Commits format (type(scope): description).
Each suggestion should be on a new line, prefixed with a number and a dash (e.g., "1 - feat:").
After each suggestion, add a brief explanation of why this message is appropriate.

Changes:
%s

Format each suggestion as:
<number> - <commit message>
Explanation: <why this message is appropriate>`, changes)
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
				Message: strings.TrimSpace(line[3:]), // Remove the "N - " prefix
			}
		} else if strings.HasPrefix(strings.ToLower(line), "explanation:") {
			if currentSuggestion != nil {
				currentSuggestion.Explanation = strings.TrimSpace(line[11:])
			}
		}
	}

	// Add the last suggestion if exists
	if currentSuggestion != nil {
		suggestions = append(suggestions, *currentSuggestion)
	}

	return suggestions
}
