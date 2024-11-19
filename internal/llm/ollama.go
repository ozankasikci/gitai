package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ozankasikci/gitai/internal/config"
	"github.com/ozankasikci/gitai/internal/logger"
)

type OllamaClient struct {
	baseURL string
	model   string
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func NewOllamaClient() (*OllamaClient, error) {
	cfg := config.Get()
	if cfg.LLM.Ollama.URL == "" {
		return nil, fmt.Errorf("Ollama URL is not configured")
	}

	logger.Debugf("Ollama URL: %s", cfg.LLM.Ollama.URL)
	return &OllamaClient{
		baseURL: cfg.LLM.Ollama.URL,
		model:   cfg.LLM.Ollama.Model,
	}, nil
}

func (c *OllamaClient) GenerateCommitSuggestions(changes string) ([]CommitSuggestion, error) {
	if changes == "" {
		return nil, fmt.Errorf("no changes provided to generate suggestions")
	}

	// Add debug logging for the input changes
	logger.Debugf("Input changes to generate suggestions: %s", changes)

	prompt := buildPrompt(changes)
	logger.Debugf("Generated prompt: %s", prompt)

	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logger.Errorf("Failed to marshal request: %v", err)
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	logger.Debugf("Sending request to Ollama URL: %s", c.baseURL+"/api/generate")
	logger.Debugf("Request payload: %s", string(jsonData))

	resp, err := http.Post(c.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Errorf("Failed to send request to Ollama: %v", err)
		return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
	}
	defer resp.Body.Close()

	logger.Debugf("Response status code: %d", resp.StatusCode)

	// Add debug logging for raw response
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	logger.Debugf("Raw Ollama response: %s", string(rawBody))

	var ollamaResp ollamaResponse
	if err := json.Unmarshal(rawBody, &ollamaResp); err != nil {
		logger.Errorf("Failed to decode response: %v", err)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if ollamaResp.Response == "" {
		logger.Errorf("Received empty response from Ollama")
		return nil, fmt.Errorf("empty response from Ollama")
	}

	logger.Debugf("\n=== Response from Ollama ===\n%s\n", ollamaResp.Response)

	return parseResponse(ollamaResp.Response), nil
}
