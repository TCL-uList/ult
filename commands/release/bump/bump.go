// Package bump provides functionality for managing version numbers
// in pubspec.yaml files, with options for committing and tagging changes.
package bump

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	cloudsql "ulist.app/ult/internal/cloud_sql"
	"ulist.app/ult/internal/playstore"
	"ulist.app/ult/internal/release"
	"ulist.app/ult/internal/version"
)

// Command flag constants
const (
	flagFetch           = "fetch"
	flagFetchForRelease = "play-store"
	flagCredentialsPath = "credentials"
	flagOnce            = "once"
	flagTarget          = "target"
	flagSource          = "source"
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
		&cli.StringFlag{
			Name:  flagCredentialsPath,
			Usage: "path to credentials json file",
		},
		&cli.BoolFlag{
			Name:  flagOnce,
			Usage: "will skip if previous bump is found in commit messages diff between source and target branches",
		},
		&cli.StringFlag{
			Name:  flagSource,
			Usage: "to/source commit SHA or branch name. required when using '--once'",
		},
		&cli.StringFlag{
			Name:  flagTarget,
			Usage: "from/target commit SHA or branch name. required when using '--once'",
		},
	},
}

// run is the main entry point for the version command that reads the pubspec.yaml file,
// updates the version according to the specified bump type, and writes the changes back.
// It handles fetching the latest build number from the Play Store if requested.
func run(ctx context.Context, cmd *cli.Command) error {
	const pubspecPath = "pubspec.yaml"

	logger.Debug("Starting version command",
		"fetch for qa", cmd.Bool(flagFetchForRelease))

	bumpType, err := version.ParseBumpType(cmd.Args().First())
	if err != nil {
		logger.Error("Failed to parse bump type", "error", err)
		return fmt.Errorf("parsing bump type: %w", err)
	}

	if cmd.Bool(flagOnce) {
		target := cmd.String(flagTarget)
		if len(target) == 0 {
			return errors.New("when using '--once' flag a target branch name or commit sha must be provided '--target=target-name-in-remote'")
		}

		source := cmd.String(flagSource)
		if len(source) == 0 {
			return errors.New("when using '--once' flag a source branch name or commit sha must be provided '--source=source-name-in-remote'")
		}

		projectId := cmd.String("project-id")
		if len(projectId) == 0 {
			return errors.New("when using '--once' flag a projectId must be provided '--project-id=000000000'")
		}

		token := cmd.String("token")
		if len(token) == 0 {
			return errors.New("when using '--once' flag a token must be provided '--token=my-token'")
		}

		alreadyBumped, err := didAlreadyBumpGitlabAPI(target, source, projectId, token)
		if err != nil {
			return err
		}

		if alreadyBumped {
			fmt.Println("Previous bump was found. Skipping a new one...")
			return nil
		}

		fmt.Println("No previous bump was found. Creating a new one...")
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

	if cmd.Bool(flagFetch) {
		if cmd.Bool(flagFetchForRelease) {
			secretsPath := cmd.String(flagCredentialsPath)
			latest, err := fetchLatestReleaseBuild(secretsPath)
			if err != nil {
				return err
			}
			version.Build = int(latest)
		} else {
			version, err = fetchLatestDevelopmentVersion()
			if err != nil {
				return err
			}
		}
	}

	version.Bump(bumpType)

	// Update version in pubspec lines
	lines[idx] = fmt.Sprintf("version: %s", version)

	// Write back to file with original permissions
	if err := writeFile(pubspecPath, lines); err != nil {
		return err
	}

	logger.Info("Updated pubspec.yaml with new version", "newVersion", version)
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
func fetchLatestReleaseBuild(path string) (int64, error) {
	logger.Info("Fetching latest build from Play Store")

	contents, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	latest, err := playstore.GetVersionFromLatestRelease(contents, "app.ulist")
	if err != nil {
		return 0, err
	}

	return latest, nil
}

func fetchLatestDevelopmentVersion() (*version.Version, error) {
	db, err := cloudsql.ConnectWithConnector()
	if err != nil {
		return nil, err
	}

	release, err := release.FetchLatestRelease(db)
	if err != nil {
		return nil, err
	}

	return &release.Version, nil
}

// Will search for bump commit in all commits there are on `source` but not on `target` (same as git log --oneline source --not target).
func didAlreadyBumpGitlabAPI(target, source, projectId, token string) (bool, error) {
	appRepo, err := gitlab.NewClient(token)
	if err != nil {
		return false, fmt.Errorf("Failed to create client: %v", err)
	}

	opt := &gitlab.CompareOptions{
		From: gitlab.Ptr(target),
		To:   gitlab.Ptr(source),
	}
	compare, _, err := appRepo.Repositories.Compare(projectId, opt)
	if err != nil {
		return false, fmt.Errorf("Error fetching commit diff from gitlab branches: %v", err)
	}

	for _, commit := range compare.Commits {
		println(commit.Title)
	}

	for _, commit := range compare.Commits {
		if strings.Contains(commit.Title, "bump version [skip ci]") {
			return true, nil
		}
	}

	return false, nil
}
