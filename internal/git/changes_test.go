package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepo(t *testing.T) string {
	// Create a temporary directory for the test repo
	tmpDir, err := os.MkdirTemp("", "git-test-*")
	require.NoError(t, err)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	// Set git config for tests
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tmpDir
	err = cmd.Run()
	require.NoError(t, err)

	return tmpDir
}

func createTestFile(t *testing.T, repoPath, filename, content string) {
	filePath := filepath.Join(repoPath, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)
}

func TestGetStagedChanges(t *testing.T) {
	// Setup test repository
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Change to test repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create and stage a test file
	createTestFile(t, tmpDir, "test.txt", "test content")
	cmd := exec.Command("git", "add", "test.txt")
	err = cmd.Run()
	require.NoError(t, err)

	// Test GetStagedChanges
	changes, err := GetStagedChanges()
	assert.NoError(t, err)
	assert.Len(t, changes, 1)
	assert.Equal(t, "test.txt", changes[0].Path)
	assert.Equal(t, "added", changes[0].Status)
}

func TestStageFile(t *testing.T) {
	// Setup test repository
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Change to test repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a test file
	createTestFile(t, tmpDir, "test.txt", "test content")

	// Test StageFile
	err = StageFile("test.txt")
	assert.NoError(t, err)

	// Verify file is staged
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(output), "A  test.txt")
}

func TestRestoreStaged(t *testing.T) {
	// Setup test repository
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Change to test repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create and stage a test file
	createTestFile(t, tmpDir, "test.txt", "test content")
	
	// First commit to create initial state
	cmd := exec.Command("git", "add", "test.txt")
	err = cmd.Run()
	require.NoError(t, err)
	
	cmd = exec.Command("git", "commit", "-m", "initial commit")
	err = cmd.Run()
	require.NoError(t, err)

	// Modify and stage the file
	err = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("modified content"), 0644)
	require.NoError(t, err)
	
	cmd = exec.Command("git", "add", "test.txt")
	err = cmd.Run()
	require.NoError(t, err)

	// Test RestoreStaged
	err = RestoreStaged("test.txt")
	assert.NoError(t, err)

	// Verify file is unstaged but modified
	cmd = exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(output), " M test.txt")
}

func TestGetGitConfig(t *testing.T) {
	// Setup test repository
	tmpDir := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Change to test repo directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Test GetGitConfig
	config, err := GetGitConfig()
	assert.NoError(t, err)
	assert.Equal(t, "Test User", config.Name)
	assert.Equal(t, "test@example.com", config.Email)
} 