package llm

import (
	"context"
	"fmt"
	"github.com/anthropics/anthropic-sdk-go/option"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/ozankasikci/gitai/internal/config"
	"github.com/ozankasikci/gitai/internal/logger"
	"github.com/ozankasikci/gitai/internal/keyring"
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
	logger.Infof("Anthropic API key: %s", apiKey)

	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is not configured")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &AnthropicClient{client: client}, nil
}

func (c *AnthropicClient) GenerateCommitSuggestions(changes string) ([]CommitSuggestion, error) {
	prompt := buildPrompt(changes)

	logger.Debugf("\n=== Changes being sent to LLM ===\n%s\n", changes)
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

Summarizes the purpose of the changes.
Highlights any key modifications or additions.
Creates a unified message that captures changes across all modified files.
	
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

func parseResponse(response string) []CommitSuggestion {
	logger.Debugf("\n=== Starting to parse response ===\nFull response text:\n%s\n", response)
	var suggestions []CommitSuggestion
	lines := strings.Split(response, "\n")
	logger.Debugf("Split response into %d lines", len(lines))

	var currentSuggestion *CommitSuggestion
	for i, line := range lines {
		line = strings.TrimSpace(line)
		logger.Debugf("Line %d: '%s'", i+1, line)
		
		if line == "" {
			logger.Debugf("Skipping empty line")
			continue
		}

		// Check if this is a new suggestion line
		if len(line) > 2 && line[0] >= '1' && line[0] <= '9' && line[1] == ' ' {
			logger.Debugf("Found potential suggestion line: %s", line)
			
			// If we have a previous suggestion, add it
			if currentSuggestion != nil {
				logger.Debugf("Adding previous suggestion: %+v", *currentSuggestion)
				suggestions = append(suggestions, *currentSuggestion)
			}

			// Extract message part
			parts := strings.SplitN(line, "-", 2)
			if len(parts) == 2 {
				logger.Debugf("Creating new suggestion with message: %s", parts[1])
				currentSuggestion = &CommitSuggestion{
					Message: strings.TrimSpace(parts[1]),
				}
			} else {
				logger.Debugf("Line doesn't match expected format (no dash found)")
			}
		} else if currentSuggestion != nil && !strings.Contains(line, "Here are three") {
			lowercaseLine := strings.ToLower(line)
			if strings.HasPrefix(lowercaseLine, "explanation:") {
				explanation := strings.TrimSpace(strings.TrimPrefix(line, "Explanation:"))
				logger.Debugf("Adding explanation to current suggestion: %s", explanation)
				currentSuggestion.Explanation = explanation
			} else {
				logger.Debugf("Skipping line: not a suggestion or explanation")
			}
		}
	}

	// Add the last suggestion
	if currentSuggestion != nil {
		logger.Debugf("Adding final suggestion: %+v", *currentSuggestion)
		suggestions = append(suggestions, *currentSuggestion)
	}

	logger.Debugf("\n=== Parsing complete ===\nFound %d suggestions", len(suggestions))
	for i, s := range suggestions {
		logger.Debugf("Suggestion %d: Message='%s', Explanation='%s'", i+1, s.Message, s.Explanation)
	}
	
	return suggestions
}
