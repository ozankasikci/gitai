package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/ozankasikci/gitai/internal/git"
	"github.com/ozankasikci/gitai/internal/llm"
	"github.com/ozankasikci/gitai/internal/logger"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate and apply commit messages",
	Long: `Generate commit messages using AI based on your staged changes.
The messages will follow conventional commits format and best practices.`,
	RunE: runCommit,
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Get staged changes
	changes, err := git.GetStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to get staged changes: %w", err)
	}

	if len(changes) == 0 {
		return fmt.Errorf("no staged changes found. Use 'git add' to stage changes")
	}

	// Get the content of staged changes
	content, err := git.GetStagedContent()
	if err != nil {
		return fmt.Errorf("failed to get staged content: %w", err)
	}

	// Initialize LLM client
	client, err := llm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	// Generate commit suggestions
	suggestions, err := client.GenerateCommitSuggestions(content)
	if err != nil {
		return fmt.Errorf("failed to generate commit suggestions: %w", err)
	}

	// Display suggestions
	fmt.Println("\nGenerated commit message suggestions:")
	for i, suggestion := range suggestions {
		fmt.Printf("\n%d. %s\n", i+1, suggestion.Message)
		if suggestion.Explanation != "" {
			fmt.Printf("   Explanation: %s\n", suggestion.Explanation)
		}
	}

	// Prompt user to select a message
	fmt.Printf("\nSelect a commit message (1-%d), 0 to cancel, or type your own message: ", len(suggestions))
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("failed to read input")
	}
	input := scanner.Text()

	var selectedMessage string
	selection, err := strconv.Atoi(input)
	if err != nil {
		// If input isn't a number, use it as the commit message
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

	// Apply the commit message
	err = git.CommitChanges(selectedMessage)
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	logger.Infof("Successfully committed changes with message: %s", selectedMessage)
	return nil
}

func init() {
	RootCmd.AddCommand(commitCmd)
}

var CommitCmd = commitCmd 