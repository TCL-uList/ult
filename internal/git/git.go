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
	logger = slog.New(slog.Default().Handler().WithGroup("git"))
)

// execCommand is a helper function that executes an external command and returns its output.
// It logs any errors that occur during execution.
func execCommand(name string, args ...string) ([]byte, error) {
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
	logger.Info("Staging all changes")
	if _, err := execCommand("git", "add", "."); err != nil {
		return fmt.Errorf("staging changes: %w", err)
	}

	logger.Info("Creating commit")
	if _, err := execCommand("git", "commit", "-m", "chore: [AUTO] build and deploy script changes"); err != nil {
		return fmt.Errorf("creating commit: %w", err)
	}

	logger.Info("Changes committed successfully")
	return nil
}

// CreateTag creates a Git tag at the current HEAD commit.
// The tag is created with the --force flag, which will move the tag if it already exists.
// Returns an error if the tag creation fails.
func CreateTag(tag string) error {
	logger.Info("Creating tag", "tag", tag)

	if _, err := execCommand("git", "tag", tag, "--force"); err != nil {
		return fmt.Errorf("creating tag %q: %w", tag, err)
	}

	logger.Info("Tag created successfully", "tag", tag)
	return nil
}

// PushChanges pushes committed changes and tags to the remote repository.
// It pushes the current branch to the 'origin' remote and then pushes all tags
// with the --force flag to override any existing tags.
// Returns an error if either push operation fails.
func PushChanges() error {
	logger.Info("Pushing changes to origin")
	if _, err := execCommand("git", "push", "--set-upstream", "origin"); err != nil {
		return fmt.Errorf("pushing to origin: %w", err)
	}

	logger.Info("Pushing tags")
	if _, err := execCommand("git", "push", "--tags", "--force"); err != nil {
		return fmt.Errorf("pushing tags: %w", err)
	}

	logger.Info("Changes pushed successfully")
	return nil
}

// GetCurrentBranch returns the name of the current Git branch.
// It uses 'git branch --show-current' to obtain the branch name.
// The branch parameter is preserved for compatibility but is not used.
// Returns the branch name and nil on success, or an empty string and an error on failure.
func GetCurrentBranch(branch string) (string, error) {
	logger.Info("Getting current branch")

	output, err := execCommand("git", "branch", "--show-current")
	if err != nil {
		return "", fmt.Errorf("getting current branch: %w", err)
	}

	currentBranch := strings.TrimSpace(string(output))
	logger.Info("Current branch identified", "branch", currentBranch)
	return currentBranch, nil
}
