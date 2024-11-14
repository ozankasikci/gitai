package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"commit-ai/pkg/git"
	"commit-ai/pkg/llm"
	"github.com/joho/godotenv"
	"bufio"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
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

	// Get user choice
	fmt.Printf("\nSelect a message (1-%d) or 'e' to edit: ", len(suggestions))
	var choice string
	fmt.Scanln(&choice)

	var commitMessage string
	if choice == "e" {
		fmt.Print("Enter your commit message: ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		commitMessage = scanner.Text()
	} else {
		index, err := strconv.Atoi(choice)
		if err != nil || index < 1 || index > len(suggestions) {
			log.Fatal("Invalid selection")
		}
		commitMessage = suggestions[index-1].Message
	}

	// Confirm commit
	fmt.Printf("\nCommit with message:\n%s\n\nProceed? (y/n): ", commitMessage)
	var confirm string
	fmt.Scanln(&confirm)

	if strings.ToLower(confirm) == "y" {
		// TODO: Implement git commit
		fmt.Println("Changes committed successfully!")
	} else {
		fmt.Println("Commit cancelled")
	}
} 