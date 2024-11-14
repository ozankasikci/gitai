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
	assert.Len(t, changes, 1)
	assert.Equal(t, "test.txt", changes[0].Path)
	assert.Equal(t, "added", changes[0].Status)
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

	for name, content := range files {
		err := os.WriteFile(name, []byte(content), 0644)
		assert.NoError(t, err)
	}

	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)
	w, err := repo.Worktree()
	assert.NoError(t, err)

	// Stage all files
	for name := range files {
		_, err = w.Add(name)
		assert.NoError(t, err)
	}

	// Modify one file after staging
	err = os.WriteFile("modified.txt", []byte("modified content"), 0644)
	assert.NoError(t, err)
	_, err = w.Add("modified.txt")
	assert.NoError(t, err)

	changes, err := GetStagedChanges()
	assert.NoError(t, err)
	assert.Len(t, changes, 2)
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