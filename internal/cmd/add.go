package cmd

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/spf13/cobra"
	"github.com/ozankasikci/gitai/internal/git"
	"github.com/charmbracelet/lipgloss"
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
}

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

	case tea.KeyMsg:
		if m.loading {
			return m, nil // Ignore key presses while loading
		}

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
			m.loading = true
			currentFile := m.choices[m.cursor].Path
			go func() {
				if m.choices[m.cursor].IsStaged {
					git.RestoreStaged(currentFile)
				} else {
					git.StageFile(currentFile)
				}
				m.choices[m.cursor].IsStaged = !m.choices[m.cursor].IsStaged
				m.loading = false
			}()
			return m, m.spinner.Tick
		case "enter":
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
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	return nil
} 