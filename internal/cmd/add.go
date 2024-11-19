package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/git"
	"github.com/charmbracelet/lipgloss"
	"github.com/ozankasikci/gitai/internal/logger"
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
	spinner  spinner.Model
	loading  bool
	quitting bool
}

type toggleCompleteMsg struct{}

func initialModel(changes []git.FileChange) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

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
		spinner:  s,
		loading:  false,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case toggleCompleteMsg:
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
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
			currentFile := m.choices[m.cursor].Path
			if m.choices[m.cursor].IsStaged {
				git.RestoreStaged(currentFile)
			} else {
				git.StageFile(currentFile)
			}
			m.choices[m.cursor].IsStaged = !m.choices[m.cursor].IsStaged
			return m, nil
		case "enter":
			logger.Infof("Proceeding to add the selected files")
			m.quitting = false
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.loading {
		return fmt.Sprintf("%s Processing...\n", m.spinner.View())
	}

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
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	// Check if the user quit
	if m != nil {
		finalModel := m.(model)
		if finalModel.quitting {
			return fmt.Errorf("user cancelled")
		}

		// Add this section to actually stage the files
		logger.Infof("Proceeding to add the selected files")
		for idx, selected := range finalModel.selected {
			if selected {
				path := finalModel.choices[idx].Path
				if err := git.StageFile(path); err != nil {
					return fmt.Errorf("failed to stage file %s: %w", path, err)
				}
				logger.Debugf("Staged file: %s", path)
			}
		}
	}

	return nil
} 