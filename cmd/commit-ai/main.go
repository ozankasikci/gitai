package main

import (
	"flag"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/ozankasikci/commit-ai/pkg/git"
	"github.com/ozankasikci/commit-ai/pkg/llm"
	"github.com/ozankasikci/commit-ai/pkg/logger"
)

var osExit = os.Exit

func init() {
	// Initialize logger with default settings
	logger.Init(false)
}

func main() {
	// Parse flags
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Re-initialize logger with verbose setting if needed
	if *verbose {
		logger.Init(*verbose)
	}

	logger.Infof("Starting commit-ai...")
	logger.Debugf("Verbose mode: %v", *verbose)

	if err := godotenv.Load(); err != nil {
		logger.Errorf("Warning: Error loading .env file: %v", err)
	} else {
		logger.Debugf("Successfully loaded .env file")
	}

	changes, err := git.GetStagedChanges()
	if err != nil {
		logger.Errorf("Failed to get staged changes: %v", err)
		os.Exit(1)
	}

	if len(changes) == 0 {
		logger.Infoln("No staged changes found")
		os.Exit(0)
	}

	fmt.Printf("→ Found %d staged files:\n", len(changes))
	for _, change := range changes {
		fmt.Printf("  %s (%s)\n", change.Path, change.Status)
	}

	// Get detailed content of changes
	content, err := git.GetStagedContent()
	if err != nil {
		logger.Error.Fatalf("Failed to get staged content: %v", err)
	}

	// Initialize Anthropic client
	client, err := llm.NewClient()
	if err != nil {
		logger.Error.Fatalf("Error initializing Anthropic client: %v", err)
	}

	fmt.Println("\n→ Generating commit message suggestions...")
	suggestions, err := client.GenerateCommitSuggestions(content)
	if err != nil {
		logger.Error.Fatalf("Error generating commit messages: %v", err)
	}

	// Display suggestions
	fmt.Println("\nSuggested commit messages:")
	for i, suggestion := range suggestions {
		fmt.Printf("\n%d - %s\n", i+1, suggestion.Message)
		fmt.Printf("   Explanation: %s\n", suggestion.Explanation)
	}

	// Get user choice with input validation
	var commitMessage string
	for {
		fmt.Printf("\nSelect a message (1-%d) or 'e' to edit, 'q' to quit: ", len(suggestions))
		reader := bufio.NewReader(os.Stdin)
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "q" {
			fmt.Println("Operation cancelled")
			return
		}

		if choice == "e" {
			fmt.Print("Enter your commit message: ")
			message, _ := reader.ReadString('\n')
			commitMessage = strings.TrimSpace(message)
			break
		}

		if index, err := strconv.Atoi(choice); err == nil && index >= 1 && index <= len(suggestions) {
			commitMessage = suggestions[index-1].Message
			break
		}

		fmt.Println("Invalid selection, please try again")
	}

	// Confirm commit
	fmt.Printf("\nCommit with message:\n%s\n\nProceed? (y/n): ", commitMessage)
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm == "y" {
		if err := git.CommitChanges(commitMessage); err != nil {
			logger.Error.Fatalf("Error committing changes: %v", err)
		}
		fmt.Println("Changes committed successfully!")
	} else {
		fmt.Println("Commit cancelled")
	}
}

func generateCommitMessages(generator llm.CommitMessageGenerator, changes string) ([]llm.CommitSuggestion, error) {
	return generator.GenerateCommitSuggestions(changes)
} 