package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/config"
)

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Configure GitAI settings",
		Long:  `Configure GitAI settings including AI provider, API keys, and other options`,
	}

	// Add subcommands
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  `Display the current GitAI configuration settings`,
		RunE:  runShowConfig,
	}

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Run setup wizard",
		Long:  `Run the interactive setup wizard to configure GitAI`,
		RunE:  runSetupConfig,
	}

	cmd.AddCommand(showCmd, setupCmd)
	return cmd
}

func runShowConfig(cmd *cobra.Command, args []string) error {
	return config.DisplayCurrentConfig()
}

func runSetupConfig(cmd *cobra.Command, args []string) error {
	return config.Setup()
}