package version_command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
	"ulist.app/ult/internal/git"
)

// Command flag constants
const (
	flagBump        = "bump"
	flagFetch       = "fetch"
	flagNoCommitTag = "no-commit-tag"
	flagNoPush      = "no-push"
)

// Bump option constants
const (
	optBuild = "build"
	optPatch = "patch"
	optMinor = "minor"
	optMajor = "major"
)

var (
	errInvalidBumpOption = errors.New("bump option must be one of: build, patch, minor, or major")
	errVersionFormat     = errors.New("invalid version format, expected: \"version: 2020.100.01+01\"")

	logger = slog.New(slog.Default().Handler().WithGroup("version command"))
)

// Cmd defines the version command for CLI
var Cmd = cli.Command{
	Name:   "version",
	Usage:  "increment the app version number",
	Action: run,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  flagFetch,
			Usage: "fetch latest build number from Play Store before incrementing",
		},
		&cli.StringFlag{
			Name:  flagBump,
			Usage: "increment the version in pubspec choosing one of: build, patch, minor or major",
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

	contents, err := os.ReadFile(pubspecPath)
	if err != nil {
		logger.Error("Failed to read pubspec.yaml", "error", err)
		return fmt.Errorf("reading pubspec.yaml: %w", err)
	}

	lines := strings.Split(string(contents), "\n")
	logger.Debug("Read pubspec.yaml", "lines", len(lines))

	version, idx, err := findVersionInPubspec(lines)
	if err != nil {
		logger.Error("Failed to parse version", "error", err)
		return fmt.Errorf("parsing version: %w", err)
	}

	logger.Info("Current version", "version", version)

	if err := bumpVersion(version, cmd); err != nil {
		return err
	}

	logger.Info("New version", "version", version)

	// Update version in pubspec lines
	lines[idx] = fmt.Sprintf("version: %s", version)

	// Write back to file with original permissions
	if err := writeFile(pubspecPath, lines); err != nil {
		return err
	}

	logger.Info("Successfully updated pubspec.yaml with new version")

	shouldCommitAndTag := !cmd.Bool(flagNoCommitTag)
	if shouldCommitAndTag {
		// assert we are on a release branch if bumping for release version
		if cmd.Bool(flagFetch) {
			if err := assertIsReleaseBranch(*version); err != nil {
				return err
			}
		}

		if err := git.CommitChanges(); err != nil {
			return err
		}

		if err := git.CreateTag(version.String()); err != nil {
			return err
		}

		// add "latest" tag for release bumps
		if cmd.Bool(flagFetch) {
			if err := git.CreateTag("latest"); err != nil {
				return err
			}
		}

	}

	shouldPush := !cmd.Bool(flagNoPush)
	if shouldPush {
		if err := git.PushChanges(); err != nil {
			return err
		}
	}

	return nil
}

func assertIsReleaseBranch(version Version) error {
	releaseBranch := fmt.Sprintf("release/v%s", version.String())
	currentBranch, err := git.GetCurrentBranch(releaseBranch)
	if err != nil {
		return err
	}

	if currentBranch != releaseBranch {
		logger.Debug("Current branch is not a release branch", "debug", releaseBranch)
		return fmt.Errorf("Should be on release branch %s but is on %s when bumping for a release version", releaseBranch, currentBranch)
	}

	return nil
}

// bumpVersion updates the provided Version struct based on the bump type specified
// in the command line arguments. It supports bumping the build number, patch, minor,
// or major version components according to semantic versioning rules.
// When bumping the build number, it can optionally fetch the latest build number
// from the Play Store first.
func bumpVersion(version *Version, cmd *cli.Command) error {
	switch cmd.String(flagBump) {
	case optBuild:
		if cmd.Bool(flagFetch) {
			latestBuild, err := fetchLatestBuildFromPlayStore()
			if err != nil {
				logger.Error("Failed to fetch latest build", "error", err)
				return fmt.Errorf("fetching latest build: %w", err)
			}
			version.Build = latestBuild + 1
		} else {
			version.Build++
		}
		logger.Info("Incremented build number", "new_build", version.Build)

	case optPatch:
		version.Build = 1
		version.Patch++
		logger.Info("Bumped patch version", "new_patch", version.Patch)

	case optMinor:
		version.Build = 1
		version.Patch = 1
		version.Minor++
		logger.Info("Bumped minor version", "new_minor", version.Minor)

	case optMajor:
		version.Build = 1
		version.Patch = 1
		version.Minor = version.Minor - (version.Minor % 100) + 100
		version.Major++
		logger.Info("Bumped major version", "new_major", version.Major)

	default:
		logger.Error("Invalid bump option", "error", errInvalidBumpOption)
		return errInvalidBumpOption
	}

	return nil
}

// findVersionInPubspec searches through the lines of pubspec.yaml to find the version
// line, parses it, and returns the Version struct along with the line index where it was found.
// Returns an error if the version line is not found or if parsing fails.
func findVersionInPubspec(lines []string) (*Version, int, error) {
	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "version: ") {
			continue
		}

		version, err := parseVersionLine(line)
		if err != nil {
			return nil, -1, err
		}

		return version, i, nil
	}
	return nil, -1, errors.New("version not found in pubspec.yaml")
}

// parseVersionLine parses a version line like "version: 10.20.03+04" into a Version struct.
// It expects the version to be in the format "major.minor.patch+build" where all
// components are integers. Returns an error if the format is invalid or any component
// cannot be parsed as an integer.
func parseVersionLine(line string) (*Version, error) {
	re := regexp.MustCompile(`^version: (\d+)\.(\d+)\.(\d+)\+(\d+)$`)
	matches := re.FindStringSubmatch(line)
	if matches == nil || len(matches) != 5 {
		return nil, fmt.Errorf("%w: got \"%s\"", errVersionFormat, line)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %w", err)
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %w", err)
	}

	build, err := strconv.Atoi(matches[4])
	if err != nil {
		return nil, fmt.Errorf("invalid build number: %w", err)
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Build: build,
	}, nil
}

// writeFile writes the provided lines back to the specified file path,
// preserving the original file permissions. It returns an error if
// getting file info or writing to the file fails.
func writeFile(path string, lines []string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		logger.Error("Failed to get file info", "path", path, "error", err)
		return fmt.Errorf("getting file info: %w", err)
	}

	err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), fileInfo.Mode())
	if err != nil {
		logger.Error("Failed to write file", "path", path, "error", err)
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// fetchLatestBuildFromPlayStore runs a fastlane command to retrieve the
// latest build number from the Google Play Store for the specified app.
// It parses the output and returns the build number as an integer.
// Returns an error if the command fails or the output cannot be parsed.
func fetchLatestBuildFromPlayStore() (int, error) {
	logger.Info("Fetching latest build from Play Store")

	cmd := exec.Command("fastlane", "run", "google_play_track_version_codes",
		"json_key:.secrets/prod/GoogleApplicationCredentials-ulist.json",
		"package_name:app.ulist")

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("running fastlane command: %w", err)
	}

	re := regexp.MustCompile(`Result: \[(\d+)\]`)
	matches := re.FindStringSubmatch(string(output))
	if matches == nil || len(matches) != 2 {
		return 0, fmt.Errorf("unexpected fastlane output: %s", string(output))
	}

	build, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("parsing build number: %w", err)
	}

	return build, nil
}
