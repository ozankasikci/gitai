package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

func setupTestRepo(t *testing.T) (string, func()) {
	// Create a temporary directory for the test repo
	dir, err := os.MkdirTemp("", "git-test-*")
	assert.NoError(t, err)

	// Initialize a new repo
	_, err = git.PlainInit(dir, false)
	assert.NoError(t, err)

	// Create a cleanup function
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestGetStagedChanges(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Change to test directory
	originalDir, _ := os.Getwd()
	err := os.Chdir(dir)
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	// Create and stage a test file
	testFile := filepath.Join(dir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Get the repo and stage the file
	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)
	w, err := repo.Worktree()
	assert.NoError(t, err)
	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	// Test GetStagedChanges
	changes, err := GetStagedChanges()
	assert.NoError(t, err)
	assert.Len(t, changes, 1, "Should only detect one staged file")
	
	// Verify only staged files are included
	for _, change := range changes {
		assert.Equal(t, "test.txt", change.Path)
		assert.Equal(t, "added", change.Status)
	}
}

func TestGetStagedContent(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Change to test directory
	originalDir, _ := os.Getwd()
	err := os.Chdir(dir)
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	// Create and stage a test file
	testFile := filepath.Join(dir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Stage the file
	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)
	w, err := repo.Worktree()
	assert.NoError(t, err)
	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	// Test GetStagedContent
	content, err := GetStagedContent()
	assert.NoError(t, err)
	assert.Contains(t, content, "test content")
}

func TestCommitChanges(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Change to test directory
	originalDir, _ := os.Getwd()
	err := os.Chdir(dir)
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	// Create and stage a test file
	testFile := filepath.Join(dir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Stage the file
	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)
	w, err := repo.Worktree()
	assert.NoError(t, err)
	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	// Test CommitChanges
	err = CommitChanges("test commit")
	assert.NoError(t, err)

	// Verify the commit
	head, err := repo.Head()
	assert.NoError(t, err)
	commit, err := repo.CommitObject(head.Hash())
	assert.NoError(t, err)
	assert.Equal(t, "test commit", commit.Message)
}

func TestGetStagedChangesErrors(t *testing.T) {
	// Test non-git directory
	tempDir, err := os.MkdirTemp("", "non-git-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	changes, err := GetStagedChanges()
	assert.Error(t, err)
	assert.Nil(t, changes)
}

func TestGetStagedChangesMultipleFiles(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	err := os.Chdir(dir)
	assert.NoError(t, err)

	// Create multiple files with different states
	files := map[string]string{
		"added.txt":    "new content",
		"modified.txt": "initial content",
	}

	// Create and stage the initial files
	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)
	w, err := repo.Worktree()
	assert.NoError(t, err)

	// Create and stage files one by one
	for name, content := range files {
		err := os.WriteFile(name, []byte(content), 0644)
		assert.NoError(t, err)
		_, err = w.Add(name)
		assert.NoError(t, err)
	}

	// Create an untracked file that shouldn't be included
	err = os.WriteFile("untracked.txt", []byte("untracked content"), 0644)
	assert.NoError(t, err)
	// Explicitly NOT staging untracked.txt

	changes, err := GetStagedChanges()
	assert.NoError(t, err)
	assert.Len(t, changes, 2, "Should only detect two staged files")
	
	// Verify only staged files are included and have correct status
	expectedChanges := map[string]string{
		"added.txt":    "added",
		"modified.txt": "added",
	}
	
	for _, change := range changes {
		expectedStatus, exists := expectedChanges[change.Path]
		assert.True(t, exists, "Found unexpected file: %s", change.Path)
		assert.Equal(t, expectedStatus, change.Status, "Incorrect status for %s", change.Path)
	}
}

func TestGetStagedContentErrors(t *testing.T) {
	// Test non-git directory
	tempDir, err := os.MkdirTemp("", "non-git-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	content, err := GetStagedContent()
	assert.Error(t, err)
	assert.Empty(t, content)
}

func TestCommitChangesErrors(t *testing.T) {
	// Test non-git directory
	tempDir, err := os.MkdirTemp("", "non-git-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	err = CommitChanges("test commit")
	assert.Error(t, err)
}

func TestGetGitConfig(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	err := os.Chdir(dir)
	assert.NoError(t, err)

	// Set up test git config
	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)

	cfg, err := repo.Config()
	assert.NoError(t, err)

	cfg.User.Name = "Test User"
	cfg.User.Email = "test@example.com"

	err = repo.SetConfig(cfg)
	assert.NoError(t, err)

	// Test getting config
	config, err := GetGitConfig()
	assert.NoError(t, err)
	assert.Equal(t, "Test User", config.Name)
	assert.Equal(t, "test@example.com", config.Email)
}

func TestCommitChangesWithConfig(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	err := os.Chdir(dir)
	assert.NoError(t, err)

	// Set up test git config
	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)

	cfg, err := repo.Config()
	assert.NoError(t, err)

	cfg.User.Name = "Test User"
	cfg.User.Email = "test@example.com"

	err = repo.SetConfig(cfg)
	assert.NoError(t, err)

	// Create and stage a test file
	err = os.WriteFile("test.txt", []byte("test content"), 0644)
	assert.NoError(t, err)

	w, err := repo.Worktree()
	assert.NoError(t, err)
	_, err = w.Add("test.txt")
	assert.NoError(t, err)

	// Test commit with config
	err = CommitChanges("test commit")
	assert.NoError(t, err)

	// Verify commit author
	head, err := repo.Head()
	assert.NoError(t, err)

	commit, err := repo.CommitObject(head.Hash())
	assert.NoError(t, err)

	assert.Equal(t, "Test User", commit.Author.Name)
	assert.Equal(t, "test@example.com", commit.Author.Email)
}

func TestParseGitConfig(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantName string
		wantEmail string
	}{
		{
			name: "simple config",
			content: `[user]
	name = John Doe
	email = john@example.com`,
			wantName: "John Doe",
			wantEmail: "john@example.com",
		},
		{
			name: "quoted values",
			content: `[user]
	name = "John Doe"
	email = "john@example.com"`,
			wantName: "John Doe",
			wantEmail: "john@example.com",
		},
		{
			name: "with other sections",
			content: `[core]
	editor = vim
[user]
	name = John Doe
	email = john@example.com
[alias]
	st = status`,
			wantName: "John Doe",
			wantEmail: "john@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, email := parseGitConfig(tt.content)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.wantEmail, email)
		})
	}
} 