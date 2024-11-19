package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"github.com/ozankasikci/gitai/internal/logger"
)

// Provider-specific configurations
type AnthropicConfig struct {
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

// Add this new common method
func setupConfigPaths() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	logger.Infof("Setting config name to config")
	logger.Infof("GITAI_ENV: %s", os.Getenv("GITAI_ENV"))
	logger.Infof("Is dev: %t", os.Getenv("GITAI_ENV") == "dev")

	if os.Getenv("GITAI_ENV") == "dev" {
		logger.Infof("Adding configs directory to config path")
		viper.AddConfigPath("./configs")
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/gitai")
	viper.AddConfigPath("$HOME/.config/gitai")
}

func Init() error {
	// Load .env file first
	if err := godotenv.Load(); err != nil {
		logger.Debugf("No .env file found: %v", err)
	}

	setupConfigPaths()

	// Set default values
	viper.SetDefault("llm.anthropic.max_tokens", 1000)
	viper.SetDefault("llm.ollama.max_tokens", 1000)

	// Initialize empty config
	cfg = &Config{}

	// Try to read existing config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Debugln("Config file not found, returning empty config")
			return nil
		}
		logger.Debugf("Failed to read config file: %v", err)
		return fmt.Errorf("failed to read config file: %w", err)
	}

	logger.Debugln("Successfully read config file")

	if err := viper.Unmarshal(cfg); err != nil {
		logger.Debugf("Failed to unmarshal config: %v", err)
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	logger.Debugln("Successfully unmarshaled config")
	return nil
}

func InitWithoutSetup() error {
	logger.Debugln("Starting InitWithoutSetup")
	
	setupConfigPaths()
	
	// Initialize empty config
	cfg = &Config{}

	// Try to read existing config
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Debugln("No config file found, returning empty config")
			return nil
		}
		logger.Errorf("Failed to read config file: %v", err)
		return fmt.Errorf("failed to read config file: %w", err)
	}

	logger.Debugln("Successfully read config file")

	if err := viper.Unmarshal(cfg); err != nil {
		logger.Errorf("Failed to unmarshal config: %v", err)
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	logger.Debugln("Successfully unmarshaled config")
	return nil
}

func validateConfig(cfg *Config) error {
	switch cfg.LLM.Provider {
	case "anthropic":
		if cfg.LLM.Anthropic.Model == "" {
			return fmt.Errorf("Anthropic model is not configured")
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

// Add this new method after GetProviderAndModel()
func (c *Config) IsSetupDone() bool {
	if c == nil {
		return false
	}

	// Check if provider is set
	if c.LLM.Provider == "" {
		return false
	}

	// Check provider-specific required fields
	switch c.LLM.Provider {
	case "anthropic":
		if c.LLM.Anthropic.Model == "" {
			return false
		}
	case "ollama":
		if c.LLM.Ollama.URL == "" || c.LLM.Ollama.Model == "" {
			return false
		}
	default:
		return false
	}

	return true
}
