package main

import (
	"github.com/joho/godotenv"
	"github.com/ozankasikci/gitai/internal/cmd"
	"github.com/ozankasikci/gitai/internal/logger"
	"github.com/ozankasikci/gitai/internal/config"
	"os"
	"path/filepath"
	"fmt"
)

var osExit = os.Exit

func init() {
	// First initialize config
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	logger.InitDefault()
	// Then initialize logger
	cfg := config.Get()
    logger.UpdateConfig(cfg.Logger.Verbose)

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Errorf("Failed to get user home directory: %v", err)
		return
	}

	configDir := filepath.Join(home, ".config", "gitai")
	envFile := filepath.Join(configDir, ".env")

	if err := os.MkdirAll(configDir, 0700); err != nil {
		logger.Errorf("Failed to create config directory: %v", err)
		return
	}

	if _, err := os.Stat(envFile); err == nil {
		if err := os.Chmod(envFile, 0600); err != nil {
			logger.Errorf("Failed to set .env file permissions: %v", err)
		}
	}

	if err := godotenv.Load(envFile); err != nil {
		logger.Debugf("No .env file found in config directory: %v", err)
	} else {
		logger.Debugf("Successfully loaded .env from config directory")
	}
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logger.Debugf("No .env file found: %v", err)
	}

	// Execute root command
	cmd.Execute()
}
