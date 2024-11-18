package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/git"
	"github.com/ozankasikci/gitai/internal/logger"
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
	logger.Infof("Starting runAuto...")
	
	// First run the add command
	if err := runAdd(cmd, args); err != nil {
		logger.Errorf("Error from runAdd: %v", err)
		if err.Error() == "user cancelled" {
			return nil // Exit cleanly if user pressed 'q'
		}
		return fmt.Errorf("failed to stage files: %w", err)
	}

	logger.Infof("Add command completed, checking staged changes...")

	// Check if any files are staged
	changes, err := git.GetStagedChanges()
	if err != nil {
		logger.Errorf("Error getting staged changes: %v", err)
		return fmt.Errorf("failed to get staged changes: %w", err)
	}

	if len(changes) == 0 {
		logger.Errorf("No files staged for commit")
		return fmt.Errorf("no files staged for commit")
	}

	logger.Infof("Found %d staged changes, proceeding to commit...", len(changes))

	// Then run the commit command
	if err := runCommit(cmd, args); err != nil {
		logger.Errorf("Error from runCommit: %v", err)
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	logger.Infof("Auto command completed successfully")
	return nil
} 