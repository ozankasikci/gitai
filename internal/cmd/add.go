package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/git"
)

type fileSelection struct {
	Path     string
	Status   string
	IsStaged bool
}

type model struct {
	choices  []fileSelection
	cursor   int
	selected map[int]bool
}

func initialModel(changes []git.FileChange) model {
	selections := make([]fileSelection, len(changes))
	for i, change := range changes {
		selections[i] = fileSelection{
			Path:     change.Path,
			Status:   change.Status,
			IsStaged: change.Staged,
		}
	}

	return model{
		choices:  selections,
		selected: make(map[int]bool),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "a":
			// Toggle all files
			allStaged := true
			for _, choice := range m.choices {
				if !choice.IsStaged {
					allStaged = false
					break
				}
			}

			for i := range m.choices {
				if allStaged {
					git.RestoreStaged(m.choices[i].Path)
					m.choices[i].IsStaged = false
				} else {
					git.StageFile(m.choices[i].Path)
					m.choices[i].IsStaged = true
				}
			}
		case " ":
			// Toggle staging for the current cursor position
			currentFile := m.choices[m.cursor].Path
			if m.choices[m.cursor].IsStaged {
				if err := git.RestoreStaged(currentFile); err == nil {
					m.choices[m.cursor].IsStaged = false
				}
			} else {
				if err := git.StageFile(currentFile); err == nil {
					m.choices[m.cursor].IsStaged = true
				}
			}
		case "enter":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Use space to stage/unstage, 'a' to toggle all, enter to finish\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if choice.IsStaged {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s (%s)\n", cursor, checked, choice.Path, choice.Status)
	}

	s += "\n(press q to quit)\n"
	return s
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

	p := tea.NewProgram(initialModel(changes))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
} 