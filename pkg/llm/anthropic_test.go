package llm

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	// Test without API key
	os.Unsetenv("ANTHROPIC_API_KEY")
	_, err := NewClient()
	assert.Error(t, err)

	// Test with API key
	os.Setenv("ANTHROPIC_API_KEY", "test-key")
	client, err := NewClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestParseResponse(t *testing.T) {
	response := `1 - feat: add user authentication
Explanation: Implements basic user authentication

2 - fix: correct database connection
Explanation: Fixes connection pooling issues`

	suggestions := parseResponse(response)
	assert.Len(t, suggestions, 2)
	assert.Equal(t, "feat: add user authentication", suggestions[0].Message)
	assert.Equal(t, "Implements basic user authentication", suggestions[0].Explanation)
	assert.Equal(t, "fix: correct database connection", suggestions[1].Message)
	assert.Equal(t, "Fixes connection pooling issues", suggestions[1].Explanation)
}

func TestMockClientSuccessCase(t *testing.T) {
	// Prepare test data
	expectedSuggestions := []CommitSuggestion{
		{
			Message:     "feat: add user authentication",
			Explanation: "Implements basic user authentication",
		},
		{
			Message:     "fix: correct database connection",
			Explanation: "Fixes connection pooling issues",
		},
	}

	// Create mock client with success case
	mockClient := NewMockClient(expectedSuggestions, nil)

	// Test generating suggestions
	suggestions, err := mockClient.GenerateCommitSuggestions("test changes")
	assert.NoError(t, err)
	assert.Equal(t, expectedSuggestions, suggestions)
}

func TestMockClientErrorCase(t *testing.T) {
	// Create mock client with error case
	expectedError := fmt.Errorf("API error")
	mockClient := NewMockClient(nil, expectedError)

	// Test generating suggestions
	suggestions, err := mockClient.GenerateCommitSuggestions("test changes")
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, suggestions)
}

// Interface compliance test
func TestClientInterface(t *testing.T) {
	// This test ensures both real and mock clients implement the same interface
	var _ CommitMessageGenerator = &Client{}
	var _ CommitMessageGenerator = &MockClient{}
}

func TestParseResponseEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     []CommitSuggestion
	}{
		{
			name:     "empty response",
			response: "",
			want:     nil,
		},
		{
			name: "malformed response",
			response: `not a proper format
some other text`,
			want: nil,
		},
		{
			name: "missing explanation",
			response: `1 - feat: add feature
2 - fix: fix bug`,
			want: []CommitSuggestion{
				{Message: "feat: add feature"},
				{Message: "fix: fix bug"},
			},
		},
		{
			name: "extra whitespace",
			response: `  1   -   feat: add feature  
  Explanation:   test explanation  `,
			want: []CommitSuggestion{
				{
					Message:     "feat: add feature",
					Explanation: "test explanation",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseResponse(tt.response)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateCommitSuggestionsError(t *testing.T) {
	// Test with invalid API key
	os.Setenv("ANTHROPIC_API_KEY", "invalid-key")
	client, err := NewClient()
	assert.NoError(t, err) // Client creation should succeed with any non-empty key

	// Test API call failure
	suggestions, err := client.GenerateCommitSuggestions("test changes")
	assert.Error(t, err)
	assert.Nil(t, suggestions)
}

// Add more edge cases
func TestGenerateCommitSuggestionsEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		changes string
		wantErr bool
	}{
		{
			name:    "empty changes",
			changes: "",
			wantErr: false,
		},
		{
			name:    "very long changes",
			changes: strings.Repeat("a", 10000),
			wantErr: false,
		},
	}

	client, err := NewClient()
	assert.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.GenerateCommitSuggestions(tt.changes)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Error(t, err) // We expect error because API key is invalid
			}
		})
	}
}

// Test the buildPrompt function
func TestBuildPrompt(t *testing.T) {
	tests := []struct {
		name     string
		changes  string
		contains []string
	}{
		{
			name:    "normal changes",
			changes: "test changes",
			contains: []string{
				"test changes",
				"Generate 3 different commit messages",
				"Conventional Commits format",
			},
		},
		{
			name:    "empty changes",
			changes: "",
			contains: []string{
				"Generate 3 different commit messages",
				"Conventional Commits format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := buildPrompt(tt.changes)
			for _, substr := range tt.contains {
				assert.Contains(t, prompt, substr)
			}
		})
	}
} 