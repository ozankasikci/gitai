package main

import (
	"fmt"
	"github.com/ozankasikci/gitai/internal/cmd"
	"github.com/ozankasikci/gitai/internal/config"
	"os"
	"github.com/pterm/pterm"
)

var osExit = os.Exit

func init() {
	if err := config.Setup(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func main() {
	// Get config and provider/model info
	cfg := config.Get()
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

	// Execute root command
	cmd.Execute()
}
