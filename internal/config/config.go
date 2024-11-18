package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Provider-specific configurations
type AnthropicConfig struct {
	APIKey    string
	Model     string
	MaxTokens int64
}

type OllamaConfig struct {
	URL       string
	Model     string
	MaxTokens int64
}

type Config struct {
	LLM struct {
		Provider string
		// Provider-specific configs
		Anthropic AnthropicConfig
		Ollama    OllamaConfig
	}
	Logger struct {
		Level   string
		Verbose bool
	}
}

var cfg *Config

func Init() error {
	// Load .env file first
	if err := godotenv.Load(); err != nil {
		logrus.Debugf("No .env file found: %v", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/gitai")
	viper.AddConfigPath("$HOME/.config/gitai")
	if os.Getenv("GITAI_ENV") == "dev" {
		viper.AddConfigPath("configs")
	}

	// Set default values
	viper.SetDefault("llm.anthropic.max_tokens", 1000)
	viper.SetDefault("llm.ollama.max_tokens", 1000)

	// Initialize empty config
	cfg = &Config{}

	// Try to read existing config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func validateConfig(cfg *Config) error {
	switch cfg.LLM.Provider {
	case "anthropic":
		if cfg.LLM.Anthropic.APIKey == "" {
			return fmt.Errorf("Anthropic API key is not configured. Set ANTHROPIC_API_KEY environment variable or in .env file")
		}
	case "ollama":
		if cfg.LLM.Ollama.URL == "" {
			return fmt.Errorf("Ollama URL is not configured")
		}
	default:
		return fmt.Errorf("unsupported LLM provider: %s", cfg.LLM.Provider)
	}
	return nil
}

func Get() *Config {
	return cfg
}

// GetProviderAndModel returns the current provider and model as strings
func (c *Config) GetProviderAndModel() (provider, model string) {
	provider = c.LLM.Provider

	switch provider {
	case "anthropic":
		model = c.LLM.Anthropic.Model
	case "ollama":
		model = c.LLM.Ollama.Model
	default:
		model = "unknown"
	}

	return provider, model
}
