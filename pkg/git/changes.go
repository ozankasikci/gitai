package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type StagedChange struct {
	Path     string
	Status   string
	FileType string
	Content  string
	Summary  string
	IsTestFile bool
	Package  string
}

// GetStagedChanges returns a list of files that are staged for commit
func GetStagedChanges() ([]StagedChange, error) {
	// Open repository in current directory
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get the working tree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get the status of the worktree
	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var changes []StagedChange
	for path, fileStatus := range status {
		// Only include files that are actually staged
		if fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked {
			change := StagedChange{
				Path:   path,
				Status: statusToString(fileStatus.Staging),
				FileType: filepath.Ext(path),
			}
			
			// Get content based on file status
			content, err := getFileContent(worktree, path)
			if err == nil {
				change.Content = content
				change.Summary = generateChangeSummary(content)
			}
			
			// Basic file analysis
			change.IsTestFile = strings.Contains(path, "_test.")
			if strings.HasSuffix(path, ".go") {
				change.Package = detectGoPackage(change.Content)
			}
			
			changes = append(changes, change)
		}
	}

	return changes, nil
}

// statusToString converts git status code to string
func statusToString(status git.StatusCode) string {
	switch status {
	case git.Unmodified:
		return "unmodified"
	case git.Added:
		return "added"
	case git.Modified:
		return "modified"
	case git.Deleted:
		return "deleted"
	case git.Renamed:
		return "renamed"
	case git.Copied:
		return "copied"
	default:
		return "unknown"
	}
}

// GetStagedContent returns a summary of the staged changes
func GetStagedContent() (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("failed to open git repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	var summary bytes.Buffer
	for path, fileStatus := range status {
		if fileStatus.Staging == git.Unmodified {
			continue
		}

		summary.WriteString(fmt.Sprintf("File: %s\n", path))
		summary.WriteString(fmt.Sprintf("Status: %s\n", statusToString(fileStatus.Staging)))

		// Handle different types of changes
		switch fileStatus.Staging {
		case git.Added:
			// For new files, show the entire content
			f, err := worktree.Filesystem.Open(path)
			if err != nil {
				continue
			}
			content, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				continue
			}
			summary.WriteString("New file content (first 500 chars):\n")
			if len(content) > 500 {
				summary.Write(content[:500])
				summary.WriteString("...[truncated]")
			} else {
				summary.Write(content)
			}
			summary.WriteString("\n")

		case git.Modified:
			// For modified files, show current content
			f, err := worktree.Filesystem.Open(path)
			if err != nil {
				continue
			}
			content, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				continue
			}
			summary.WriteString("Modified file content (first 500 chars):\n")
			if len(content) > 500 {
				summary.Write(content[:500])
				summary.WriteString("...[truncated]")
			} else {
				summary.Write(content)
			}
			summary.WriteString("\n")
		}
		summary.WriteString("\n---\n")
	}

	return summary.String(), nil
}

// CommitChanges commits the staged changes with the given message
func CommitChanges(message string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	config, err := GetGitConfig()
	if err != nil {
		return fmt.Errorf("failed to get git config: %w", err)
	}

	if config.Name == "" || config.Email == "" {
		return fmt.Errorf("git user.name and user.email must be set")
	}

	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  config.Name,
			Email: config.Email,
			When:  time.Now(),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

// Add these new functions
func GetGitConfig() (*GitConfig, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	config, err := repo.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get git config: %w", err)
	}

	// Get user config
	name := config.User.Name
	email := config.User.Email

	// If not set in local config, try global config
	if name == "" || email == "" {
		globalConfig, err := readGlobalGitConfig()
		if err == nil {
			if name == "" {
				name = globalConfig.Name
			}
			if email == "" {
				email = globalConfig.Email
			}
		}
	}

	return &GitConfig{
		Name:  name,
		Email: email,
	}, nil
}

type GitConfig struct {
	Name  string
	Email string
}

func readGlobalGitConfig() (*GitConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Try to read global git config
	configPath := filepath.Join(home, ".gitconfig")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return &GitConfig{}, nil // Return empty config if file doesn't exist
	}

	// Parse the config file
	name, email := parseGitConfig(string(data))
	return &GitConfig{
		Name:  name,
		Email: email,
	}, nil
}

func parseGitConfig(content string) (name, email string) {
	lines := strings.Split(content, "\n")
	inUserSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "[user]" {
			inUserSection = true
			continue
		} else if strings.HasPrefix(line, "[") {
			inUserSection = false
			continue
		}

		if inUserSection {
			if strings.HasPrefix(line, "name = ") {
				name = strings.TrimPrefix(line, "name = ")
				name = strings.Trim(name, "\"")
			} else if strings.HasPrefix(line, "email = ") {
				email = strings.TrimPrefix(line, "email = ")
				email = strings.Trim(email, "\"")
			}
		}
	}

	return name, email
}

// Helper functions for content analysis
func getFileContent(w *git.Worktree, path string) (string, error) {
	file, err := w.Filesystem.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read first 1000 bytes as preview
	preview := make([]byte, 1000)
	n, _ := file.Read(preview)
	return string(preview[:n]), nil
}

func generateChangeSummary(content string) string {
	// For now, just return first 100 chars if content is longer
	if len(content) > 100 {
		return content[:100] + "..."
	}
	return content
}

func detectGoPackage(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
} 