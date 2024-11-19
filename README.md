# git-ai

git-ai is an AI-powered Git assistant that helps automate and enhance your Git workflow. It uses AI to analyze your code changes and generate meaningful commit messages, making version control more efficient and consistent.

Currently, it supports the following providers:
- Anthropic Claude
- Ollama

## Features

- Interactive file staging
- AI-powered commit message generation
- Support for multiple AI providers (Anthropic Claude and Ollama)
- Conventional Commits format support
- Local and global Git config integration

## Installation

### Pre-built Binaries

#### macOS Universal Binary (Intel & Apple Silicon)
```bash
curl -LO https://github.com/ozankasikci/gitai/releases/latest/download/gitai-darwin-universal
chmod +x gitai-darwin-universal
sudo mv gitai-darwin-universal /usr/local/bin/gitai
```

#### macOS Intel (x86_64)
```bash
curl -LO https://github.com/ozankasikci/gitai/releases/latest/download/gitai-darwin-x86_64
chmod +x gitai-darwin-x86_64
sudo mv gitai-darwin-x86_64 /usr/local/bin/gitai
```

#### macOS Apple Silicon (M1/M2)
```bash
curl -LO https://github.com/ozankasikci/gitai/releases/latest/download/gitai-darwin-arm64
chmod +x gitai-darwin-arm64
sudo mv gitai-darwin-arm64 /usr/local/bin/gitai
```

### Build from source

Alternatively, you can build from source:

```bash
go install github.com/ozankasikci/gitai@latest
```

After downloading, make it executable and move it to your PATH:

## Configuration

Before using git-ai, you'll need to configure it:

```bash
gitai config setup
```

This will guide you through:
1. Selecting an AI provider (Anthropic or Ollama)
2. Setting up provider-specific settings
3. Configuring API keys if needed

## Commands

### `gitai add`

Interactive file staging command that allows you to:
- View all changed files with their status
- Stage/unstage files individually using space
- Toggle all files using 'a'
- Confirm selection with enter

### `gitai commit`

Generates AI-powered commit messages based on your staged changes:
- Analyzes all staged files and their changes
- Generates multiple commit message suggestions
- Follows conventional commits format
- Allows selecting from suggestions or entering custom message

### `gitai auto`

Combines `add` and `commit` commands for a streamlined workflow:
1. Opens interactive staging interface
2. After staging files, automatically proceeds to commit message generation
3. Allows selecting or entering commit message

### `gitai config`

Manages git-ai configuration:
- `gitai config setup`: Interactive configuration wizard
- `gitai config show`: Display current configuration

## Examples

### Interactive Staging

```bash
$ gitai add
Use space to stage/unstage, 'a' to toggle all, enter to finish

> [x] internal/git/changes.go (modified)
  [ ] internal/cmd/commit.go (modified)
  [ ] README.md (added)

(press q to quit)
```

### Commit Message Generation

```bash
$ gitai commit
Generating commit suggestions...

1. feat: Add git-ai core functionality with AI commit message generation
Explanation: Implements the main functionality for AI-powered Git operations including file staging and commit message generation

2. refactor: Restructure git operations and improve error handling
Explanation: Enhances the codebase organization and error handling in git operations

3. feat(git): Implement interactive staging and commit workflow
Explanation: Adds user-friendly interface for staging files and generating commit messages

Select a commit message (1-3), 0 to cancel, or type your own message:
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details
