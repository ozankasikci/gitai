package llm

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    
    "github.com/ozankasikci/gitai/internal/config"
    "github.com/ozankasikci/gitai/internal/logger"
)

type OllamaClient struct {
    baseURL string
    model   string
}

type ollamaRequest struct {
    Model    string `json:"model"`
    Prompt   string `json:"prompt"`
    Stream   bool   `json:"stream"`
}

type ollamaResponse struct {
    Response string `json:"response"`
}

func NewOllamaClient() (*OllamaClient, error) {
    cfg := config.Get()
    if cfg.LLM.OllamaURL == "" {
        return nil, fmt.Errorf("Ollama URL is not configured")
    }

    return &OllamaClient{
        baseURL: cfg.LLM.OllamaURL,
        model:   cfg.LLM.Model,
    }, nil
}

func (c *OllamaClient) GenerateCommitSuggestions(changes string) ([]CommitSuggestion, error) {
    prompt := buildPrompt(changes)
    
    reqBody := ollamaRequest{
        Model:  c.model,
        Prompt: prompt,
        Stream: false,
    }
    
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    resp, err := http.Post(c.baseURL+"/api/generate", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to send request to Ollama: %w", err)
    }
    defer resp.Body.Close()

    var ollamaResp ollamaResponse
    if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    logger.Debugf("\n=== Response from Ollama ===\n%s\n", ollamaResp.Response)
    
    return parseResponse(ollamaResp.Response), nil
} 