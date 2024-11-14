package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"commit-ai/pkg/git"
	"commit-ai/pkg/llm"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		log.Fatal("ANTHROPIC_API_KEY is required in .env file")
	}

	// Get staged changes
	changes, err := git.GetStagedChanges()
	if err != nil {
		log.Fatalf("Error getting staged changes: %v", err)
	}

	if len(changes) == 0 {
		fmt.Println("No staged changes found")
		return
	}

	fmt.Printf("→ Found %d staged files:\n", len(changes))
	for _, change := range changes {
		fmt.Printf("  %s (%s)\n", change.Path, change.Status)
	}

	// Get detailed content of changes
	content, err := git.GetStagedContent()
	if err != nil {
		log.Fatalf("Error getting staged content: %v", err)
	}

	// Initialize Anthropic client
	client, err := llm.NewClient()
	if err != nil {
		log.Fatalf("Error initializing Anthropic client: %v", err)
	}

	fmt.Println("\n→ Generating commit message suggestions...")
	suggestions, err := client.GenerateCommitSuggestions(content)
	if err != nil {
		log.Fatalf("Error generating commit messages: %v", err)
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
			log.Fatalf("Error committing changes: %v", err)
		}
		fmt.Println("Changes committed successfully!")
	} else {
		fmt.Println("Commit cancelled")
	}
} 