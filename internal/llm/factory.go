package llm

import (
	"fmt"

	"github.com/ozankasikci/gitai/internal/config"
)

func NewLLMClient() (CommitMessageGenerator, error) {
	cfg := config.Get()
	
	switch cfg.LLM.Provider {
	case "anthropic":
		return NewAnthropicClient()
	case "ollama":
		return NewOllamaClient()
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.LLM.Provider)
	}
} 