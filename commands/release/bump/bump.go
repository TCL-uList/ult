// Package bump provides functionality for managing version numbers
// in pubspec.yaml files, with options for committing and tagging changes.
package bump

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
	cloudsql "ulist.app/ult/internal/cloud_sql"
	"ulist.app/ult/internal/git"
	"ulist.app/ult/internal/release"
	"ulist.app/ult/internal/version"
)

// Command flag constants
const (
	flagFetch           = "fetch"
	flagFetchForRelease = "play-store"
	flagNoCommitTag     = "no-commit-tag"
	flagNoPush          = "no-push"
)

// Bump option constants
const (
	optBuild = "build"
	optPatch = "patch"
	optMinor = "minor"
	optMajor = "major"
)

var (
	logger = slog.Default().WithGroup("bump_command")
)

// Cmd defines the version command for CLI
var Cmd = cli.Command{
	Name:   "bump",
	Usage:  "increment the version in pubspec choosing one of: build, patch, minor or major",
	Action: run,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  flagFetch,
			Usage: "fetch latest build number from external server before incrementing. Skiping this flag will use the local pubspec.yaml version build number",
		},
		&cli.BoolFlag{
			Name:  flagFetchForRelease,
			Usage: "use play store server to fetch",
		},
		&cli.BoolFlag{
			Name:  flagNoCommitTag,
			Usage: "skip creating Git commits and tag for version changes. Use with --no-push for dry-run simulations",
		},
		&cli.BoolFlag{
			Name:  flagNoPush,
			Usage: "skip pushing changes to remote repository. Changes will remain local for verification",
		},
	},
}

// run is the main entry point for the version command that reads the pubspec.yaml file,
// updates the version according to the specified bump type, and writes the changes back.
// It handles fetching the latest build number from the Play Store if requested.
func run(ctx context.Context, cmd *cli.Command) error {
	const pubspecPath = "pubspec.yaml"

	logger.Debug("Starting version command",
		"fetch for qa", cmd.Bool(flagFetchForRelease),
		"noCommitTag", cmd.Bool(flagNoCommitTag),
		"noPush", cmd.Bool(flagNoPush))

	bumpType, err := version.ParseBumpType(cmd.Args().First())
	if err != nil {
		logger.Error("Failed to parse bump type", "error", err)
		return fmt.Errorf("parsing bump type: %w", err)
	}

	contents, err := os.ReadFile(pubspecPath)
	if err != nil {
		logger.Error("Failed to read pubspec.yaml", "error", err)
		return fmt.Errorf("reading pubspec.yaml: %w", err)
	}

	lines := strings.Split(string(contents), "\n")
	logger.Debug("Read pubspec.yaml", "lineCount", len(lines))

	version, idx, err := version.FetchFromLines(lines)
	if err != nil {
		logger.Error("Failed to parse version", "error", err)
		return fmt.Errorf("parsing version: %w", err)
	}

	logger.Info("Found version in pubspec", "version", version, "at line idx", idx)

	var build int
	if cmd.Bool(flagFetch) {
		if cmd.Bool(flagFetchForRelease) {
			build, err = fetchLatestReleaseBuild()
			if err != nil {
				return err
			}
		} else {
			build, err = fetchLatestDevelopmentBuild()
			if err != nil {
				return err
			}
		}
	}
	version.Build = build
	version.Bump(bumpType)

	// Update version in pubspec lines
	lines[idx] = fmt.Sprintf("version: %s", version)

	// Write back to file with original permissions
	if err := writeFile(pubspecPath, lines); err != nil {
		return err
	}

	logger.Info("Updated pubspec.yaml with new version", "newVersion", version)

	shouldCommitAndTag := !cmd.Bool(flagNoCommitTag)
	if shouldCommitAndTag {
		logger.Info("Processing git operations", "commitAndTag", true)

		// assert we are on a release branch if bumping for release version
		if cmd.Bool(flagFetchForRelease) {
			logger.Info("Checking release branch", "version", version)
			if err := assertIsReleaseBranch(*version); err != nil {
				return err
			}
		}

		logger.Info("Committing changes")
		if err := git.CommitChanges(); err != nil {
			return err
		}

		logger.Info("Creating version tag", "tag", version.String())
		if err := git.CreateTag(version.String()); err != nil {
			return err
		}

		// add "latest" tag for release bumps
		if cmd.Bool(flagFetchForRelease) {
			logger.Info("Creating 'latest' tag for release")
			if err := git.CreateTag("latest"); err != nil {
				return err
			}
		}
	} else {
		logger.Info("Skipping git commit and tag", "noCommitTag", true)
	}

	shouldPush := !cmd.Bool(flagNoPush)
	if shouldPush {
		logger.Info("Pushing changes to remote")
		if err := git.PushChanges(); err != nil {
			return err
		}
		logger.Info("Changes pushed successfully")
	} else {
		logger.Info("Skipping git push", "noPush", true)
	}

	logger.Info("Version command completed successfully", "newVersion", version)
	return nil
}

// assertIsReleaseBranch verifies that the current git branch matches the expected release branch
// pattern for the given version.
func assertIsReleaseBranch(version version.Version) error {
	releaseBranch := fmt.Sprintf("release/v%s", version.String())
	logger.Info("Checking release branch", "expected", releaseBranch)

	currentBranch, err := git.GetCurrentBranch()
	if err != nil {
		logger.Error("Failed to get current branch", "error", err)
		return fmt.Errorf("getting current branch: %w", err)
	}

	if currentBranch != releaseBranch {
		logger.Error("Branch mismatch",
			"expected", releaseBranch,
			"actual", currentBranch)
		return fmt.Errorf("branch mismatch, expected: %q, found: %q when bumping for release version", releaseBranch, currentBranch)
	}

	logger.Info("Confirmed on correct release branch", "branch", releaseBranch)
	return nil
}

// writeFile writes the provided lines back to the specified file path,
// preserving the original file permissions. It returns an error if
// getting file info or writing to the file fails.
func writeFile(path string, lines []string) error {
	logger.Debug("Writing in pubspec file", "path", path, "lineCount", len(lines))

	fileInfo, err := os.Stat(path)
	if err != nil {
		logger.Error("Failed to get pubspec file info", "path", path, "error", err)
		return fmt.Errorf("getting file info: %w", err)
	}

	content := strings.Join(lines, "\n")
	err = os.WriteFile(path, []byte(content), fileInfo.Mode())
	if err != nil {
		logger.Error("Failed to write pubspec file", "path", path, "error", err)
		return fmt.Errorf("writing file: %w", err)
	}

	logger.Debug("Pubspec file written successfully", "path", path, "bytes", len(content))
	return nil
}

// fetchLatestReleaseBuild runs a fastlane command to retrieve the
// latest build number from the Google Play Store for the specified app.
// It parses the output and returns the build number as an integer.
// Returns an error if the command fails or the output cannot be parsed.
func fetchLatestReleaseBuild() (int, error) {
	logger.Info("Fetching latest build from Play Store using fastlane")

	cmd := exec.Command("fastlane", "run", "google_play_track_version_codes",
		"json_key:.secrets/prod/GoogleApplicationCredentials-ulist.json",
		"package_name:app.ulist")

	logger.Debug("Executing fastlane command")
	output, err := cmd.Output()
	if err != nil {
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
		}
		logger.Error("Fastlane command failed",
			"stdout", string(output),
			"stderr", stderr,
			"error", err)
		return 0, fmt.Errorf("running fastlane command: %w", err)
	}

	outputStr := string(output)
	logger.Debug("Parsing fastlane output", "output", outputStr)

	re := regexp.MustCompile(`Result: \[(\d+)\]`)
	matches := re.FindStringSubmatch(outputStr)
	if matches == nil || len(matches) != 2 {
		logger.Error("Unexpected fastlane output format", "output", outputStr)
		return 0, fmt.Errorf("unexpected fastlane output: %s", outputStr)
	}

	build, err := strconv.Atoi(matches[1])
	if err != nil {
		logger.Error("Failed to parse build number", "value", matches[1], "error", err)
		return 0, fmt.Errorf("parsing build number: %w", err)
	}

	logger.Debug("Retrieved latest build number with fastlane", "build", build)
	return build, nil
}

func fetchLatestDevelopmentBuild() (int, error) {
	db, err := cloudsql.ConnectWithConnector()
	if err != nil {
		return -1, err
	}

	release, err := release.FetchLatestRelease(db)
	if err != nil {
		return -1, err
	}

	return release.Bump(), nil
}
