package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewAutoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "auto",
		Short: "Automatically stage and commit changes",
		Long:  `Stage files and generate commit message in one command`,
		RunE:  runAuto,
	}
}

func runAuto(cmd *cobra.Command, args []string) error {
	// First run the add command
	if err := runAdd(cmd, args); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Then run the commit command
	if err := runCommit(cmd, args); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
} 