package git

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type StagedChange struct {
	Path     string
	Status   string
	FileType string
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
		// Check if file is staged
		if fileStatus.Staging != git.Unmodified {
			change := StagedChange{
				Path:   path,
				Status: statusToString(fileStatus.Staging),
			}
			
			// Get file type based on extension
			change.FileType = filepath.Ext(path)
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

	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "commit-ai",
			Email: "commit-ai@local",
			When:  time.Now(),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
} 