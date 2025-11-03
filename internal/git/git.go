package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureGitConfig sets Git user.name and user.email
// Uses GITHUB_ACTOR environment variable in GitHub Actions environment
func ensureGitConfig(repoPath string) error {
	// Check user.name
	cmd := exec.Command("git", "config", "--local", "user.name")
	cmd.Dir = repoPath
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	userNameSet := err == nil && strings.TrimSpace(stdout.String()) != ""

	if !userNameSet {
		// Set user.name if not configured
		var userName string
		if actor := os.Getenv("GITHUB_ACTOR"); actor != "" {
			userName = actor
		} else {
			// Fallback: GitHub Actions default
			userName = "github-actions[bot]"
		}

		cmd := exec.Command("git", "config", "--local", "user.name", userName)
		cmd.Dir = repoPath
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to set user.name: %w\nstderr: %s", err, stderr.String())
		}
	}

	// Check user.email
	cmd = exec.Command("git", "config", "--local", "user.email")
	cmd.Dir = repoPath
	stdout.Reset()
	stderr.Reset()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	userEmailSet := err == nil && strings.TrimSpace(stdout.String()) != ""

	if !userEmailSet {
		// Set user.email if not configured
		var userEmail string
		if actor := os.Getenv("GITHUB_ACTOR"); actor != "" {
			// GitHub Actions no-reply email address format
			userEmail = fmt.Sprintf("%s@users.noreply.github.com", actor)
		} else {
			// Fallback: GitHub Actions default
			userEmail = "github-actions[bot]@users.noreply.github.com"
		}

		cmd := exec.Command("git", "config", "--local", "user.email", userEmail)
		cmd.Dir = repoPath
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to set user.email: %w\nstderr: %s", err, stderr.String())
		}
	}

	return nil
}

// DetectChanges detects changed files
//
// Preconditions:
// - repoPath is a valid Git repository path
//
// Postconditions:
// - Returns a list of changed files
// - Returns error if an error occurred
//
// Invariants:
// - Returned file paths are relative paths
func DetectChanges(repoPath string) ([]string, error) {
	// Execute git status --porcelain to get changed files
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to execute git status: %w\nstderr: %s", err, stderr.String())
	}

	// Parse output and create file list
	output := stdout.String()
	if output == "" {
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var files []string

	for _, line := range lines {
		// git status --porcelain output format: " M file.txt" or "MM file.txt"
		// The part after the status code is the file path
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			// Second field is the file path
			file := strings.Join(fields[1:], " ")
			// Convert absolute path to relative path
			if filepath.IsAbs(file) {
				relPath, err := filepath.Rel(repoPath, file)
				if err == nil {
					file = relPath
				}
			}
			files = append(files, file)
		}
	}

	return files, nil
}

// HasChanges checks if there are changes in the repository
//
// Preconditions:
// - repoPath is a valid Git repository path
//
// Postconditions:
// - Returns true if there are changes, false otherwise
// - Returns error if an error occurred
//
// Invariants:
// - Git repository must be initialized
func HasChanges(repoPath string) (bool, error) {
	files, err := DetectChanges(repoPath)
	if err != nil {
		return false, err
	}

	return len(files) > 0, nil
}

// Commit commits changes
//
// Preconditions:
// - repoPath is a valid Git repository path
// - message is a valid commit message
// - files is a list of files to commit (if empty, all changes are committed)
//
// Postconditions:
// - Specified files are committed
// - Commit message is applied
//
// Invariants:
// - Git repository must be initialized
func Commit(repoPath, message string, files []string) error {
	if message == "" {
		return fmt.Errorf("commit message is empty")
	}

	// Check if Git user.name and user.email are set, and set them if not
	// Uses GITHUB_ACTOR in GitHub Actions environment
	err := ensureGitConfig(repoPath)
	if err != nil {
		return fmt.Errorf("failed to verify Git configuration: %w", err)
	}

	// Stage all changes if no files are specified
	if len(files) == 0 {
		cmd := exec.Command("git", "add", "-A")
		cmd.Dir = repoPath

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to execute git add: %w\nstderr: %s", err, stderr.String())
		}
	} else {
		// Stage specified files
		args := append([]string{"add"}, files...)
		cmd := exec.Command("git", args...)
		cmd.Dir = repoPath

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to execute git add: %w\nstderr: %s", err, stderr.String())
		}
	}

	// Execute commit
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		// Don't treat as error if there are no changes to commit (already committed)
		if strings.Contains(stderrStr, "nothing to commit") || strings.Contains(stderrStr, "nothing added to commit") {
			return nil
		}
		return fmt.Errorf("failed to execute git commit: %w\nstderr: %s", err, stderrStr)
	}

	return nil
}

// SetRemoteURLWithToken sets token in remote URL
//
// Preconditions:
// - repoPath is a valid Git repository path
// - remote is the remote name (optional, default is "origin")
// - token is a GitHub Personal Access Token
//
// Postconditions:
// - Remote URL is updated to include token
//
// Invariants:
// - Remote repository must be configured
func SetRemoteURLWithToken(repoPath, remote, token string) error {
	if remote == "" {
		remote = "origin"
	}

	if token == "" {
		return fmt.Errorf("token is not set")
	}

	// Get current remote URL
	cmd := exec.Command("git", "remote", "get-url", remote)
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w\nstderr: %s", err, stderr.String())
	}

	currentURL := strings.TrimSpace(stdout.String())

	// Build HTTPS format URL
	// Convert https://github.com/owner/repo.git to https://TOKEN@github.com/owner/repo.git
	var newURL string
	if strings.HasPrefix(currentURL, "https://") {
		// Already starts with https://
		if strings.Contains(currentURL, "@") {
			// Replace if token is already included
			// https://oldtoken@github.com/owner/repo.git -> https://newtoken@github.com/owner/repo.git
			parts := strings.SplitN(currentURL, "@", 2)
			if len(parts) == 2 {
				// parts[1] is the github.com/owner/repo.git part
				newURL = fmt.Sprintf("https://%s@%s", token, parts[1])
			} else {
				return fmt.Errorf("failed to parse remote URL: %s", currentURL)
			}
		} else {
			// Add token if not included
			// https://github.com/owner/repo.git -> https://TOKEN@github.com/owner/repo.git
			newURL = strings.Replace(currentURL, "https://", fmt.Sprintf("https://%s@", token), 1)
		}
	} else if strings.HasPrefix(currentURL, "git@") {
		// Convert SSH format to HTTPS format and add token
		// git@github.com:owner/repo.git -> https://TOKEN@github.com/owner/repo.git
		newURL = strings.Replace(currentURL, "git@github.com:", fmt.Sprintf("https://%s@github.com/", token), 1)
	} else {
		return fmt.Errorf("unsupported remote URL format: %s", currentURL)
	}

	// Set remote URL
	cmd = exec.Command("git", "remote", "set-url", remote, newURL)
	cmd.Dir = repoPath

	stderr.Reset()
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to set remote URL: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}

// Push pushes to remote repository
//
// Preconditions:
// - repoPath is a valid Git repository path
// - remote is the remote name (optional, default is "origin")
// - branch is the branch name (optional, default is current branch)
// - token is a GitHub Personal Access Token (optional, if not set, URL won't include token)
//
// Postconditions:
// - Changes are pushed to remote repository
//
// Invariants:
// - Remote repository must be configured
func Push(repoPath, remote, branch, token string) error {
	if remote == "" {
		remote = "origin"
	}

	// Set token in remote URL if token is set
	if token != "" {
		err := SetRemoteURLWithToken(repoPath, remote, token)
		if err != nil {
			return fmt.Errorf("failed to set token in remote URL: %w", err)
		}
	}

	// Get current branch if branch is not specified
	if branch == "" {
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = repoPath

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to get current branch: %w\nstderr: %s", err, stderr.String())
		}

		branch = strings.TrimSpace(stdout.String())
	}

	// Execute push
	cmd := exec.Command("git", "push", remote, branch)
	cmd.Dir = repoPath

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute git push: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}

// CommitAndPush commits and pushes changes (batch operation)
//
// Preconditions:
// - repoPath is a valid Git repository path
// - message is a valid commit message
// - files is a list of files to commit (if empty, all changes are committed)
// - remote is the remote name (optional)
// - branch is the branch name (optional)
// - token is a GitHub Personal Access Token (optional)
//
// Postconditions:
// - Changes are committed and pushed to remote repository
//
// Invariants:
// - Commit and push are executed sequentially
func CommitAndPush(repoPath, message string, files []string, remote, branch, token string) error {
	// Check if there are changes
	hasChanges, err := HasChanges(repoPath)
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}

	if !hasChanges {
		return nil // Do nothing if there are no changes
	}

	// Commit
	err = Commit(repoPath, message, files)
	if err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Push
	err = Push(repoPath, remote, branch, token)
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// IsGitRepository checks if the specified path is a Git repository
//
// Preconditions:
// - repoPath is a valid directory path
//
// Postconditions:
// - Returns true if it's a Git repository, false otherwise
//
// Invariants:
// - Uses git rev-parse command for verification (more reliable)
// - Also checks for .git directory/file existence as fallback
func IsGitRepository(repoPath string) bool {
	// First check with git rev-parse command (more reliable method)
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = repoPath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		// If git rev-parse succeeds, it's a Git repository
		return true
	}

	// If git rev-parse fails, check for .git existence as fallback
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return false
	}

	gitDir := filepath.Join(absPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}

	// Consider it a Git repository if .git is a directory, file, or symbolic link
	// (in shallow clones, .git can be a file)
	return info.IsDir() || (info.Mode()&os.ModeSymlink != 0) || (info.Mode().IsRegular())
}

// GetCurrentBranch gets the current branch name
//
// Preconditions:
// - repoPath is a valid Git repository path
//
// Postconditions:
// - Returns the current branch name
//
// Invariants:
// - Git repository must be initialized
func GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w\nstderr: %s", err, stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}
