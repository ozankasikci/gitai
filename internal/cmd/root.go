package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/logger"
	"github.com/ozankasikci/gitai/internal/config"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gitai",
	Short: "GitAI - AI-powered Git assistant",
	Long: `GitAI is a command-line tool that uses AI to help with Git operations.
Currently supports generating commit messages based on staged changes.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Only initialize logger, not the full config
		logger.InitDefault()

		// Skip config initialization for config commands
		if cmd.Parent() != nil && cmd.Parent().Name() == "config" {
			return nil
		}

		// Initialize config for other commands
		if err := config.Init(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}
		return nil
	},
}

func init() {
	// Add all subcommands here
	RootCmd.AddCommand(
		NewAddCommand(),
		NewCommitCommand(),
		NewGitignoreCommand(),
		NewAutoCommand(),
		NewConfigCommand(),
	)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
} 