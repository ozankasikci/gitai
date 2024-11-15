package git

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const githubGitignoreBaseURL = "https://raw.githubusercontent.com/github/gitignore/main/%s.gitignore"

// GenerateGitignore generates a .gitignore file from the specified templates
func GenerateGitignore(templates []string) error {
	content := new(strings.Builder)
	
	for _, template := range templates {
		templateContent, err := fetchTemplate(template)
		if err != nil {
			return fmt.Errorf("failed to fetch template %s: %w", template, err)
		}
		
		content.WriteString(fmt.Sprintf("### %s ###\n", template))
		content.WriteString(templateContent)
		content.WriteString("\n\n")
	}

	// Check if .gitignore already exists
	if _, err := os.Stat(".gitignore"); err == nil {
		// Backup existing file
		if err := os.Rename(".gitignore", ".gitignore.backup"); err != nil {
			return fmt.Errorf("failed to backup existing .gitignore: %w", err)
		}
	}

	// Write new .gitignore
	if err := os.WriteFile(".gitignore", []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}

func fetchTemplate(template string) (string, error) {
	// Capitalize first letter for GitHub URL
	template = strings.Title(strings.ToLower(template))
	url := fmt.Sprintf(githubGitignoreBaseURL, template)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("template not found: %s", template)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
} 