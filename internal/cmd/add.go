package cmd

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/git"
	"strings"
)

func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Interactively stage files for commit",
			RunE:  runAdd,
	}
	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Get all changed files (both staged and unstaged)
	changes, err := git.GetAllChanges()
	if err != nil {
		return fmt.Errorf("failed to get changes: %w", err)
	}

	if len(changes) == 0 {
		fmt.Println("No changes to stage")
		return nil
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "\U0001F449 {{ .Path | cyan }} ({{ .Status | red }})",
		Inactive: "  {{ .Path | faint }} ({{ .Status | faint }})",
		Selected: "\U00002705 {{ .Path | green }} ({{ .Status }})",
	}

	searcher := func(input string, index int) bool {
		change := changes[index]
		name := change.Path
		input = input

		return strings.Contains(strings.ToLower(name), strings.ToLower(input))
	}

	prompt := promptui.Select{
		Label:     "Select a file to stage (↑/↓ to navigate, enter to select, ctrl+c to finish)",
		Items:     changes,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	stagedFiles := make(map[string]bool)
	for {
		i, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				break // User pressed Ctrl+C to finish
			}
			return fmt.Errorf("prompt failed: %w", err)
		}

		file := changes[i].Path
		if !stagedFiles[file] {
			if err := git.StageFile(file); err != nil {
				return fmt.Errorf("failed to stage %s: %w", file, err)
			}
			stagedFiles[file] = true
			fmt.Printf("Staged: %s\n", file)
		}
	}

	fmt.Printf("Successfully staged %d files\n", len(stagedFiles))
	return nil
} 