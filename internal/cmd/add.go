package cmd

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/git"
)

type fileSelection struct {
	Path     string
	Status   string
	IsStaged bool
}

func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Interactively stage files for commit",
		RunE:  runAdd,
	}
	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	changes, err := git.GetAllChanges()
	if err != nil {
		return fmt.Errorf("failed to get changes: %w", err)
	}

	if len(changes) == 0 {
		fmt.Println("No changes to stage")
		return nil
	}

	selections := make([]fileSelection, len(changes))
	for i, change := range changes {
		selections[i] = fileSelection{
			Path:     change.Path,
			Status:   change.Status,
			IsStaged: change.Staged,
		}
	}

	templates := &promptui.SelectTemplates{
		Active:   "[{{ if .IsStaged }}✓{{ else }} {{ end }}] {{ .Path | cyan }} ({{ .Status | red }})",
		Inactive: "[{{ if .IsStaged }}✓{{ else }} {{ end }}] {{ .Path }} ({{ .Status }})",
		Selected: "[{{ if .IsStaged }}✓{{ else }} {{ end }}] {{ .Path }} ({{ .Status }})",
	}

	prompt := promptui.Select{
		Label:     "Space to stage/unstage, Enter to exit",
		Items:     selections,
		Templates: templates,
		Size:      10,
	}

	for {
		i, _, err := prompt.Run()
		if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
			return nil // Exit on Ctrl+C or Enter
		}
		if err != nil {
			return fmt.Errorf("prompt failed: %w", err)
		}

		// Toggle staging
		file := selections[i].Path
		if err := git.StageFile(file); err != nil {
			return fmt.Errorf("failed to stage %s: %w", file, err)
		}
		
		// Update the staged status
		changes, _ = git.GetAllChanges()
		for j, change := range changes {
			if j < len(selections) {
				selections[j].IsStaged = change.Staged
			}
		}
		prompt.Items = selections
	}
} 