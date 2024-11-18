package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	LLM struct {
		Provider   string
		APIKey     string
		OllamaURL  string
		Model      string
		MaxTokens  int64
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
		// Don't return error if .env doesn't exist
		logrus.Debugf("No .env file found: %v", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.gitai")
	viper.AddConfigPath("/etc/gitai")

	// Set defaults
	viper.SetDefault("llm.model", "claude-3-5-haiku-latest")
	viper.SetDefault("llm.maxTokens", int64(1024))
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.verbose", false)

	// Update the default LLM provider configuration
	viper.SetDefault("llm.provider", "ollama")
	viper.SetDefault("llm.ollamaURL", "http://localhost:11434")
	viper.SetDefault("llm.model", "llama2:3.2")
	viper.SetDefault("llm.maxTokens", int64(1024))

	// Environment variables
	viper.SetEnvPrefix("GITAI")
	viper.AutomaticEnv()
	
	// Remove or comment out the Anthropic-specific binding since Ollama is now default
	// viper.BindEnv("llm.apiKey", "ANTHROPIC_API_KEY")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Update the API key validation to only check when using Anthropic
	if cfg.LLM.Provider == "anthropic" && cfg.LLM.APIKey == "" {
		return fmt.Errorf("Anthropic API key is not configured. Set ANTHROPIC_API_KEY environment variable or in .env file")
	}

	return nil
}

func Get() *Config {
	return cfg
} 