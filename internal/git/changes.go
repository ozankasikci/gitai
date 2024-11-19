package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/ozankasikci/gitai/internal/logger"
)

type StagedChange struct {
	Path       string
	Status     string
	FileType   string
	Content    string
	Summary    string
	IsTestFile bool
	Package    string
}

type FileChange struct {
	Path   string
	Status string
	Staged bool
}

// GetStagedChanges returns a list of files that are staged for commit
func GetStagedChanges() ([]StagedChange, error) {
	logger.Infof("Getting staged changes...")

	repo, err := git.PlainOpen(".")
	if err != nil {
		logger.Errorf("Failed to open git repository: %v", err)
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}
	logger.Debugf("Successfully opened git repository")

	worktree, err := repo.Worktree()
	if err != nil {
		logger.Errorf("Failed to get worktree: %v", err)
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}
	logger.Debugf("Successfully got worktree")

	status, err := worktree.Status()
	if err != nil {
		logger.Errorf("Failed to get status: %v", err)
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var changes []StagedChange
	for path, fileStatus := range status {
		if fileStatus.Staging != git.Unmodified && fileStatus.Staging != git.Untracked ||
		   fileStatus.Worktree == git.Deleted {
			logger.Debugf("Found staged file: %s (status: %s)", path, statusToString(fileStatus.Staging))
			change := StagedChange{
				Path:   path,
				Status: statusToString(fileStatus.Staging),
			}
			if fileStatus.Worktree == git.Deleted {
				change.Status = "deleted"
			}
			changes = append(changes, change)
		}
	}

	logger.Infof("Found %d staged changes", len(changes))
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

// FormatChangesForPrompt converts the staged changes into a string format
// that can be used in the LLM prompt
func FormatChangesForPrompt(changes []StagedChange) string {
	var builder strings.Builder
	for _, change := range changes {
		builder.WriteString(fmt.Sprintf("%s (%s)\n", change.Path, change.Status))
	}
	return builder.String()
}

// GetStagedContent returns a summary of the staged changes
func GetStagedContent() (string, error) {
	changes, err := GetStagedChanges()
	if err != nil {
		return "", fmt.Errorf("failed to get staged changes: %w", err)
	}

	var content strings.Builder
	
	// First add the summary of changes
	content.WriteString("=== Changes Summary ===\n\n")
	for _, change := range changes {
		content.WriteString(fmt.Sprintf("%s (status: %s)\n", change.Path, change.Status))
	}

	// Then add the actual diff content
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", fmt.Errorf("failed to open git repository: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = w.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	content.WriteString("\n=== Detailed Changes ===\n")
	
	// Get the actual diff for each staged file
	for _, change := range changes {
		if change.Status == "deleted" {
			content.WriteString(fmt.Sprintf("\nDeleted file: %s\n", change.Path))
			continue
		}

		diff, err := getDiffForFile(repo, change.Path)
		if err != nil {
			return "", fmt.Errorf("failed to get diff for %s: %w", change.Path, err)
		}
		content.WriteString(fmt.Sprintf("\n=== %s ===\n%s\n", change.Path, diff))
	}

	return content.String(), nil
}

func getDiffForFile(repo *git.Repository, path string) (string, error) {
	// Use git command to get diff since go-git's diff functionality is limited
	cmd := exec.Command("git", "diff", "--cached", "--", path)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff for %s: %w", path, err)
	}

	if len(output) == 0 {
		// If no diff (e.g., for newly added files), get the entire content
		cmd = exec.Command("git", "show", ":"+path)
		output, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get content for %s: %w", path, err)
		}
	}

	return string(output), nil
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

// GetAllChanges returns both staged and unstaged changes
func GetAllChanges() ([]FileChange, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var changes []FileChange
	for path, fileStatus := range status {
		var statusStr string
		if fileStatus.Staging == git.Deleted || fileStatus.Worktree == git.Deleted {
			statusStr = "deleted"
		} else {
			statusStr = statusToString(fileStatus.Worktree)
		}

		change := FileChange{
			Path:   path,
			Status: statusStr,
			Staged: fileStatus.Staging != git.Unmodified,
		}
		changes = append(changes, change)
	}

	return changes, nil
}

// StageFile stages a single file
func StageFile(path string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	_, err = w.Add(path)
	return err
}

// RestoreStaged unstages a single file
func RestoreStaged(path string) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}

	_, err = repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Use git command directly for single file unstaging
	cmd := exec.Command("git", "restore", "--staged", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unstage file: %w", err)
	}

	return nil
}
