package set_version

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"ulist.app/ult/internal/git"
	"ulist.app/ult/internal/version"
)

const (
	flagPath    = "path"
	flagFromTag = "from-tag"
	flagApi     = "api"
	flagHash    = "hash"
)

var (
	logger = slog.Default().WithGroup("set_version_command")
)

// Cmd defines the version command for CLI
var Cmd = cli.Command{
	Name:   "set-version",
	Usage:  "set a specific version in pubspec file",
	Action: run,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  flagPath,
			Usage: "path of the pubspec.yaml file",
		},
		&cli.BoolFlag{
			Name:  flagFromTag,
			Usage: "whether to use the current commit tag as version",
		},
		&cli.BoolFlag{
			Name:  flagApi,
			Usage: "whether to use the gitlab api to fetch the tag. defaults to false (git cli)",
		},
		&cli.StringFlag{
			Name:  flagHash,
			Usage: "commit hash from which to fetch the tag (used with --from-tag)",
		},
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	token := cmd.String("token")
	projectId := cmd.String("project-id")
	versionStr := cmd.Args().First()
	useTagAsVersion := cmd.Bool(flagFromTag)
	fetchTagFromGitlab := cmd.Bool(flagApi)
	commitHash := cmd.String(flagHash)
	if len(commitHash) == 0 {
		commitHash = "HEAD"
	}
	pubspecPath := cmd.String(flagPath)
	if len(pubspecPath) == 0 {
		pubspecPath = "pubspec.yaml"
	}

	logger.Debug("Starting set version command",
		"version", versionStr,
		"use_tag", useTagAsVersion,
		"path", pubspecPath,
		"hash", commitHash,
	)

	if len(versionStr) == 0 {
		if !useTagAsVersion {
			return fmt.Errorf("you need to provide version as positional argument (usage: ult release set-version 2000.100.10+01) or use --%s\n", flagFromTag)
		}

		var tag string
		var err error
		if fetchTagFromGitlab {
			tag, err = tagFromGitlab(commitHash, token, projectId)
			if err != nil {
				return err
			}
		} else {
			tag, err = tagFromGit(commitHash)
			if err != nil {
				return err
			}
		}

		if len(tag) == 0 {
			return fmt.Errorf("no tag was found for the given commit hash (%s)", commitHash)
		}

		if strings.Contains(tag, "+") {
			return fmt.Errorf("%s is an invalid QA version tag, it can't contain '+' sign", tag)
		}

		versionStr = strings.Replace(tag, "QA-v", "", 1)
		versionStr = strings.TrimSpace(versionStr)
		// add build part (+00) so version module can parse the string
		versionStr = fmt.Sprintf("%s+00", versionStr)
	}

	newVersion, err := version.Parse(versionStr)
	if err != nil {
		return err
	}

	contents, err := os.ReadFile(pubspecPath)
	if err != nil {
		return fmt.Errorf("failed to read pubspec.yaml: %w", err)
	}

	lines := strings.Split(string(contents), "\n")

	_, idx, err := version.FetchFromLines(lines)
	if err != nil {
		return fmt.Errorf("failed to find version string in pubspec file: %w", err)
	}

	// Update version in pubspec lines
	lines[idx] = fmt.Sprintf("version: %s", newVersion)

	// Write back to file with original permissions
	if err := writeFile(pubspecPath, lines); err != nil {
		return err
	}

	fmt.Printf("updated pubspec.yaml with version: %s\n", newVersion)
	return nil
}

// writeFile writes the provided lines back to the specified file path,
// preserving the original file permissions. It returns an error if
// getting file info or writing to the file fails.
func writeFile(path string, lines []string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}

	content := strings.Join(lines, "\n")
	err = os.WriteFile(path, []byte(content), fileInfo.Mode())
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

func tagFromGit(hash string) (string, error) {
	tag, err := git.GetTagFromCommit(hash)
	if err != nil {
		return tag, err
	}

	return tag, nil
}

func tagFromGitlab(hash, token, projectId string) (string, error) {
	var tag string

	appRepo, err := gitlab.NewClient(token)
	if err != nil {
		return tag, err
	}
	opt := gitlab.ListTagsOptions{
		Search: gitlab.Ptr("^QA-v"),
	}
	tags, _, err := appRepo.Tags.ListTags(projectId, &opt)

	for _, tag := range tags {
		if !strings.Contains(tag.Commit.ID, hash) {
			continue
		}

		return tag.Name, nil
	}

	return tag, nil
}
