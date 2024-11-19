package config

import (
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
)

func DisplayCurrentConfig() error {
	logrus.Debug("Starting DisplayCurrentConfig")
	
	// Initialize config without setup
	if err := InitWithoutSetup(); err != nil {
		logrus.Errorf("Failed to initialize config: %v", err)
		return err
	}

	cfg := Get()
	logrus.Debugf("Got config: %+v", cfg)
	
	provider, model := cfg.GetProviderAndModel()
	logrus.Debugf("Provider: %s, Model: %s", provider, model)

	if provider == "" {
		logrus.Debug("No provider configured")
		pterm.Warning.Println("No configuration found. Run 'gitai config setup' to configure GitAI.")
		return nil
	}

	pterm.DefaultSection.Println("Current Configuration")

	// Display Provider info
	pterm.Println()
	pterm.FgLightCyan.Println("Provider Settings:")
	pterm.Printf("• Provider: %s\n", provider)
	pterm.Printf("• Model: %s\n", model)

	if provider == "ollama" {
		pterm.Printf("• URL: %s\n", cfg.LLM.Ollama.URL)
	}

	return nil
} 