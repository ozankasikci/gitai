package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/git"
)

func NewGitignoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gitignore",
		Short: "Gitignore file operations",
		Long:  `Commands for working with .gitignore files`,
	}

	generateCmd := &cobra.Command{
		Use:   "generate [templates...]",
		Short: "Generate .gitignore file",
		Long: `Generate a .gitignore file based on specified templates.
Example: gitai gitignore generate go node python`,
		RunE: runGenerate,
	}

	cmd.AddCommand(generateCmd)
	return cmd
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("at least one template name is required")
	}

	if err := git.GenerateGitignore(args); err != nil {
		return fmt.Errorf("failed to generate .gitignore: %w", err)
	}

	fmt.Printf("Successfully generated .gitignore with templates: %v\n", args)
	return nil
} 