// Package git provides utilities for performing common Git operations
// such as committing changes, creating tags, and pushing to remote repositories.
package git

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

var (
	logger = slog.New(slog.Default().Handler().WithGroup("git command"))
)

// execCommand is a helper function that executes an external command and returns its output.
// It logs any errors that occur during execution.
func execCommand(name string, args ...string) ([]byte, error) {
	logger.Debug("Executing command", "cmd", name, "args", args)

	cmd := exec.Command(name, args...)
	bytes, err := cmd.Output()
	if err != nil {
		logger.Debug("Command execution failed",
			"command", name,
			"args", args,
			"stdout", string(bytes),
			"error", err)
		return []byte{}, err
	}

	return bytes, nil
}

// CommitChanges stages all modified files and commits them with a standard message.
// It stages files using 'git add .' and creates a commit with an automated message.
// Returns an error if either the staging or commit operations fail.
func CommitChanges() error {
	logger.Info("Staging all changes in repo")
	if _, err := execCommand("git", "add", "."); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	logger.Info("Commiting changes")
	if _, err := execCommand("git", "commit", "-m", "chore: [AUTO] build and deploy script changes"); err != nil {
		logger.Error("Failed commit changes", "error", err)
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	return nil
}

// CreateTag creates a Git tag at the current HEAD commit.
// The tag is created with the --force flag, which will move the tag if it already exists.
// Returns an error if the tag creation fails.
func CreateTag(tag string) error {
	logger.Info("Starting tagging commit", "info", tag)

	if _, err := execCommand("git", "tag", tag, "--force"); err != nil {
		logger.Error("Failed create git tag", "error", err)
		return fmt.Errorf("failed to create git tag: %w", err)
	}

	logger.Info("Successfully created tag", "info", tag)
	return nil
}

// PushChanges pushes committed changes and tags to the remote repository.
// It pushes the current branch to the 'origin' remote and then pushes all tags
// with the --force flag to override any existing tags.
// Returns an error if either push operation fails.
func PushChanges() error {
	logger.Info("Starting pushing changes to remote")
	if _, err := execCommand("git", "push", "--set-upstream", "origin"); err != nil {
		logger.Error("Failed to push changes", "error", err)
		return fmt.Errorf("failed to push changes: %w", err)
	}

	logger.Info("Pushing tags")
	if _, err := execCommand("git", "push", "--tags", "--force"); err != nil {
		logger.Error("Failed to force push tags", "error", err)
		return fmt.Errorf("failed to force push tags: %w", err)
	}

	logger.Info("Successfully pushed changes to remote")
	return nil
}

// GetCurrentBranch returns the name of the current Git branch.
// It uses 'git branch --show-current' to obtain the branch name.
// The branch parameter is unused and should be removed.
// Returns the branch name and nil on success, or an empty string and an error on failure.
func GetCurrentBranch(branch string) (string, error) {
	logger.Info("Starting fetching current branch name")

	bytes, err := execCommand("git", "branch", "--show-current")
	if err != nil {
		logger.Error("Failed get current branch name", "error", err)
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	currentBranch := strings.TrimSpace(string(bytes))
	logger.Info("Finished comparing branches")
	return currentBranch, nil

}
