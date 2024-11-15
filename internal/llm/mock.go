package llm

type MockClient struct {
	suggestions []CommitSuggestion
	err        error
}

func NewMockClient(suggestions []CommitSuggestion, err error) *MockClient {
	return &MockClient{
		suggestions: suggestions,
		err:        err,
	}
}

func (m *MockClient) GenerateCommitSuggestions(changes string) ([]CommitSuggestion, error) {
	return m.suggestions, m.err
} 