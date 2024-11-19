package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/ozankasikci/gitai/internal/git"
	"github.com/ozankasikci/gitai/internal/llm"
	"github.com/ozankasikci/gitai/internal/logger"
	"github.com/spf13/cobra"
	"github.com/pterm/pterm"
)

type commitModel struct {
	spinner    spinner.Model
	loading    bool
	err        error
	suggestions []llm.CommitSuggestion
}

func initialCommitModel() commitModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return commitModel{
		spinner: s,
		loading: true,
	}
}

func (m commitModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m commitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case llm.SuggestionsMsg:
		m.loading = false
		m.suggestions = msg.Suggestions
		return m, tea.Quit
	case error:
		m.err = msg
		m.loading = false
		return m, tea.Quit
	}
	return m, nil
}

func (m commitModel) View() string {
	if m.loading {
		return fmt.Sprintf("%s Generating commit suggestions...\n", m.spinner.View())
	}
	return ""
}

func NewCommitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "commit",
		Short: "Generate and apply commit messages",
		Long: `Generate commit messages using AI based on your staged changes.
The messages will follow conventional commits format and best practices.`,
		RunE: runCommit,
	}
}

func runCommit(cmd *cobra.Command, args []string) error {
	changes, err := git.GetStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to get staged changes: %w", err)
	}

	if len(changes) == 0 {
		return fmt.Errorf("no staged changes found. Use 'git add' to stage changes")
	}

	content, err := git.GetStagedContent()
	if err != nil {
		return fmt.Errorf("failed to get staged content: %w", err)
	}
	logger.Debugf("\n=== Staged content from git.GetStagedContent() ===\nLength: %d\nContent:\n%s\n", len(content), content)

	client, err := llm.NewLLMClient()
	if err != nil {
		return fmt.Errorf("failed to create LLM client: %w", err)
	}

	p := tea.NewProgram(initialCommitModel())
	
	// Run LLM in goroutine
	go func() {
		logger.Infof("Starting LLM goroutine")
		logger.Debugf("\n=== Content being sent to GenerateCommitSuggestions ===\n%s\n", content)
		suggestions, err := client.GenerateCommitSuggestions(content)
		if err != nil {
			logger.Errorf("Error in LLM goroutine: %v", err)
			p.Send(err)
			return
		}
		logger.Infof("Successfully generated %d suggestions, sending to UI", len(suggestions))
		p.Send(llm.SuggestionsMsg{Suggestions: suggestions})
	}()

	model, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running program: %w", err)
	}

	m := model.(commitModel)
	if m.err != nil {
		return m.err
	}

	suggestions := m.suggestions

	// Display suggestions
	fmt.Println("\nGenerated commit message suggestions:")
	for i, suggestion := range suggestions {
		fmt.Printf("\n%d. %s\n", i+1, suggestion.Message)
		if suggestion.Explanation != "" {
			pterm.Println()
			pterm.FgLightCyan.Println("Explanation:")
			pterm.FgGray.Println(suggestion.Explanation)
		}
	}

	// Rest of the commit logic remains the same
	fmt.Printf("\nSelect a commit message (1-%d), 0 to cancel, or type your own message: ", len(suggestions))
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read input")
	}
	input := scanner.Text()

	// Handle empty input (just pressing enter)
	if input == "" {
		fmt.Println("Empty input - commit cancelled")
		return nil
	}

	var selectedMessage string
	selection, err := strconv.Atoi(input)
	if err != nil {
		selectedMessage = input
	} else {
		if selection == 0 {
			fmt.Println("Commit cancelled")
			return nil
		}

		if selection < 1 || selection > len(suggestions) {
			return fmt.Errorf("invalid selection: %d", selection)
		}

		selectedMessage = suggestions[selection-1].Message
	}

	if err := git.CommitChanges(selectedMessage); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	logger.Infof("Successfully committed changes with message: %s", selectedMessage)
	return nil
}

var CommitCmd = NewCommitCommand() 
