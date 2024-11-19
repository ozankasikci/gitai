package main

import (
	"fmt"
	"os"

	"github.com/ozankasikci/gitai/internal/cmd"
	"github.com/ozankasikci/gitai/internal/config"
	"github.com/ozankasikci/gitai/internal/logger"
	"github.com/pterm/pterm"
)

var osExit = os.Exit

func init() {
	// Remove config setup from init
}

func main() {
	// Initialize logger first
	logger.InitDefault()

	// Set up config paths first
	config.InitWithoutSetup()

	// Check if we're running config setup command
	isConfigSetup := len(os.Args) > 2 && os.Args[1] == "config" && os.Args[2] == "setup"

	// Only run config setup if config doesn't exist and we're not explicitly running setup
	if !isConfigSetup {
		cfg := config.Get()
		if !cfg.IsSetupDone() {
			if err := config.Setup(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		}

		// Get config and provider/model info for display
		cfg = config.Get()
		provider, model := cfg.GetProviderAndModel()

		// Create table data
		tableData := pterm.TableData{
			{"Provider", provider},
			{"Model", model},
		}

		// Render table
		_ = pterm.DefaultTable.
			WithData(tableData).
			WithBoxed(true).
			Render()
	}

	// Execute root command
	cmd.Execute()
}
