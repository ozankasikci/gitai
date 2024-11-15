package git

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

// Helper function to set up a test repository
func setupTestRepo(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "git-test")
	assert.NoError(t, err)

	_, err = git.PlainInit(dir, false)
	assert.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

func TestGetStagedChanges(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	err := os.Chdir(dir)
	assert.NoError(t, err)

	// Create and stage a Go file
	goContent := `package main

func main() {
	println("Hello, World!")
}
`
	err = os.WriteFile("main.go", []byte(goContent), 0644)
	assert.NoError(t, err)

	// Create and stage a test file
	testContent := `package main

import "testing"

func TestHello(t *testing.T) {
	// Test implementation
}
`
	err = os.WriteFile("main_test.go", []byte(testContent), 0644)
	assert.NoError(t, err)

	// Stage the files
	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)
	w, err := repo.Worktree()
	assert.NoError(t, err)
	_, err = w.Add("main.go")
	assert.NoError(t, err)
	_, err = w.Add("main_test.go")
	assert.NoError(t, err)

	// Test GetStagedChanges
	changes, err := GetStagedChanges()
	assert.NoError(t, err)
	assert.Len(t, changes, 2)

	// Create a map for easier testing
	changeMap := make(map[string]StagedChange)
	for _, change := range changes {
		changeMap[change.Path] = change
	}

	// Test main.go properties
	mainFile := changeMap["main.go"]
	assert.Equal(t, "main.go", mainFile.Path)
	assert.Equal(t, "added", mainFile.Status)
	assert.Equal(t, ".go", mainFile.FileType)
	assert.Equal(t, "main", mainFile.Package)
	assert.False(t, mainFile.IsTestFile)
	assert.Contains(t, mainFile.Content, "package main")
	assert.Contains(t, mainFile.Summary, "package main")

	// Test main_test.go properties
	testFile := changeMap["main_test.go"]
	assert.Equal(t, "main_test.go", testFile.Path)
	assert.Equal(t, "added", testFile.Status)
	assert.Equal(t, ".go", testFile.FileType)
	assert.Equal(t, "main", testFile.Package)
	assert.True(t, testFile.IsTestFile)
	assert.Contains(t, testFile.Content, "func TestHello")
}

func TestGetStagedChangesWithLargeContent(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	err := os.Chdir(dir)
	assert.NoError(t, err)

	// Create a large file
	var largeContent strings.Builder
	for i := 0; i < 2000; i++ {
		largeContent.WriteString(fmt.Sprintf("Line %d\n", i))
	}

	err = os.WriteFile("large.txt", []byte(largeContent.String()), 0644)
	assert.NoError(t, err)

	// Stage the file
	repo, err := git.PlainOpen(".")
	assert.NoError(t, err)
	w, err := repo.Worktree()
	assert.NoError(t, err)
	_, err = w.Add("large.txt")
	assert.NoError(t, err)

	// Test GetStagedChanges
	changes, err := GetStagedChanges()
	assert.NoError(t, err)
	assert.Len(t, changes, 1)

	change := changes[0]
	assert.Equal(t, "large.txt", change.Path)
	assert.Equal(t, ".txt", change.FileType)
	assert.True(t, len(change.Content) <= 1000, "Content should be truncated")
	assert.True(t, len(change.Summary) <= 103, "Summary should be <= 100 chars + '...'")
}

func TestDetectGoPackage(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "simple package",
			content: `package main

func main() {}`,
			expected: "main",
		},
		{
			name: "package with comments",
			content: `// Some comment
package utils

func Helper() {}`,
			expected: "utils",
		},
		{
			name: "no package",
			content: `func main() {}`,
			expected: "",
		},
		{
			name: "empty content",
			content: "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detectGoPackage(tc.content)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestGenerateChangeSummary(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "short content",
			content:  "Short summary",
			expected: "Short summary",
		},
		{
			name:     "long content",
			content:  strings.Repeat("a", 150),
			expected: strings.Repeat("a", 100) + "...",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := generateChangeSummary(tc.content)
			assert.Equal(t, tc.expected, result)
		})
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