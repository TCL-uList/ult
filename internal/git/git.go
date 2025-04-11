// Package git provides utilities for performing common Git operations
// such as committing changes, creating tags, and pushing to remote repositories.
package git

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	logger = slog.Default().WithGroup("git")
)

// execCommand is a helper function that executes an external command and returns its output.
// It logs any errors that occur during execution.
func execCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)

	logger.Debug("Executing command", "cmd", name, "args", args)

	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		stderr := ""
		if execErr, ok := err.(*exec.ExitError); ok {
			exitErr = execErr
			stderr = string(execErr.Stderr)
		}

		logger.Error("Command execution failed",
			"cmd", name,
			"args", args,
			"stdout", string(output),
			"stderr", stderr,
			"error", err,
			"exitCode", exitErr)
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	logger.Debug("Command executed successfully", "cmd", name, "args", args)
	return output, nil
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

func GetLatestCommitInfo() (*Commit, error) {
	logger.Info("Getting commit information")

	output, err := execCommand("git", "log", "-1")
	if err != nil {
		return nil, fmt.Errorf("getting commit info: %w", err)
	}

	return getCommitInfoFromStdout(string(output))
}

func GetCommitInfo(hash string) (*Commit, error) {
	logger.Info("Getting commit information")

	if len(hash) == 0 {
		return nil, errors.New("Commit hash cannot be empty")
	}

	output, err := execCommand("git", "log", "-1", hash)
	if err != nil {
		return nil, fmt.Errorf("getting commit info: %w", err)
	}

	return getCommitInfoFromStdout(string(output))
}

func getCommitInfoFromStdout(stdout string) (*Commit, error) {
	re := regexp.MustCompile(`^commit (.*).*?\nAuthor: (.*?) <(.*)>.*?\nDate:   (.*?)\n\n(.*)`)
	matches := re.FindStringSubmatch(stdout)

	dateStr := matches[4]
	date, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", dateStr)
	if err != nil {
		return nil, fmt.Errorf("not able to parse date (%s): %w", dateStr, err)
	}

	commit := Commit{
		Hash:        matches[1],
		AuthorName:  matches[2],
		AuthorEmail: matches[3],
		Date:        date,
		Message:     strings.TrimSpace(matches[5]),
	}

	return &commit, nil
}
