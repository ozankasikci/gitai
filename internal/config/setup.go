package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ozankasikci/gitai/internal/keyring"
)

// Setup initializes the application configuration and environment
func Setup() error {
	// Clear existing configuration
	if err := clearConfig(); err != nil {
		return fmt.Errorf("failed to clear existing configuration: %v", err)
	}

	// First initialize config
	if err := Init(); err != nil {
		return fmt.Errorf("failed to initialize config: %v", err)
	}

	if err := setupConfigDirectory(); err != nil {
		return fmt.Errorf("failed to setup config directory: %v", err)
	}

	if err := loadEnvFiles(); err != nil {
		return fmt.Errorf("failed to load env files: %v", err)
	}

	// Interactive provider selection
	provider, err := selectProvider()
	if err != nil {
		return err
	}

	// Provider-specific setup
	if err := setupProvider(provider); err != nil {
		return err
	}

	return nil
}

func setupConfigDirectory() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	configDir := filepath.Join(home, ".config", "gitai")
	envFile := filepath.Join(configDir, ".env")

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	if _, err := os.Stat(envFile); err == nil {
		if err := os.Chmod(envFile, 0600); err != nil {
			logrus.Errorf("Failed to set .env file permissions: %v", err)
		}
	}

	if err := godotenv.Load(envFile); err != nil {
		logrus.Debugf("No .env file found in config directory: %v", err)
	} else {
		logrus.Debugf("Successfully loaded .env from config directory")
	}

	return nil
}

func loadEnvFiles() error {
	if err := godotenv.Load(); err != nil {
		logrus.Debugf("No .env file found: %v", err)
	}
	return nil
}

func selectProvider() (string, error) {
	result, err := pterm.DefaultInteractiveSelect.
		WithOptions([]string{"Anthropic", "Ollama"}).
		WithDefaultText("Select AI provider:").
		Show()

	if err != nil {
		return "", fmt.Errorf("failed to get provider selection: %v", err)
	}
	return result, nil
}

func setupProvider(provider string) error {
	// Set the provider in lowercase
	cfg := Get()
	switch provider {
	case "Anthropic":
		cfg.LLM.Provider = "anthropic"
		return setupAnthropic()
	case "Ollama":
		cfg.LLM.Provider = "ollama"
		return setupOllama()
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}
}

func setupAnthropic() error {
	// Get API key
	apiKey, err := pterm.DefaultInteractiveTextInput.
		WithMask("*").
		WithDefaultText("Enter Anthropic API key:").
		Show()

	if err != nil {
		return fmt.Errorf("failed to get API key: %v", err)
	}

	// Store API key in keyring
	if err := keyring.StoreAPIKey(keyring.Anthropic, apiKey); err != nil {
		return fmt.Errorf("failed to store API key in keyring: %v", err)
	}

	// Select model
	model, err := pterm.DefaultInteractiveSelect.
		WithOptions([]string{
			"claude-3-5-haiku-latest",
			"claude-3-5-sonnet-latest",
		}).
		WithDefaultText("Select Claude model:").
		Show()

	if err != nil {
		return fmt.Errorf("failed to get model selection: %v", err)
	}

	// Save to config (without API key)
	cfg := Get()
	cfg.LLM.Anthropic.Model = model
	cfg.LLM.Anthropic.MaxTokens = 1000 // Set default max tokens

	// Save config to file
	if err := SaveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

func setupOllama() error {
	// Get Ollama URL
	url, err := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("http://localhost:11434").
		WithDefaultText("Enter Ollama URL:").
		Show()

	if err != nil {
		return fmt.Errorf("failed to get URL: %v", err)
	}

	// Get model name
	model, err := pterm.DefaultInteractiveTextInput.
		WithDefaultValue("llama3.2").
		WithDefaultText("Enter model name:").
		Show()

	if err != nil {
		return fmt.Errorf("failed to get model name: %v", err)
	}

	// Save to config
	cfg := Get()
	cfg.LLM.Ollama.URL = url
	cfg.LLM.Ollama.Model = model

	// Save config to file
	if err := SaveConfig(); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}

// Add this new function to save the config
func SaveConfig() error {
	for key, value := range map[string]interface{}{
		"llm.provider":        cfg.LLM.Provider,
		"llm.anthropic.model": cfg.LLM.Anthropic.Model,
		"llm.ollama.url":      cfg.LLM.Ollama.URL,
		"llm.ollama.model":    cfg.LLM.Ollama.Model,
	} {
		viper.Set(key, value)
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "gitai")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	viper.SetConfigFile(filepath.Join(configDir, "config.yaml"))
	return viper.WriteConfig()
}

// ClearConfig removes existing configuration
func clearConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	configDir := filepath.Join(home, ".config", "gitai")
	configFile := filepath.Join(configDir, "config.yaml")
	envFile := filepath.Join(configDir, ".env")

	// Remove config file if it exists
	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %v", err)
	}

	// Remove env file if it exists
	if err := os.Remove(envFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove env file: %v", err)
	}

	return nil
}
