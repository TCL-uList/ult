package commit_command

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"ulist.app/ult/internal/core"
	"ulist.app/ult/internal/version"
)

const (
	flagRef        = "ref"
	flagUseVersion = "use-pubspec-version"
)

var (
	logger = slog.Default().WithGroup("tag_command")
)

var Cmd = cli.Command{
	Name:   "tag",
	Usage:  "creates a new tag in the repository that points to the supplied ref",
	Action: run,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagRef,
			Aliases: []string{"r"},
			Usage:   "a commit SHA, another tag name, or branch name to tag",
		},
		&cli.BoolFlag{
			Name:  flagUseVersion,
			Usage: "use version from pubspec file as tag name",
			Value: false,
		},
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	token, err := core.GetToken(cmd)
	if err != nil {
		return err
	}
	projectId, projectIdErr := core.GetProjectID(cmd)
	ref := cmd.String(flagRef)
	tagName := cmd.Args().First()
	useVersionAsTagName := cmd.Bool(flagUseVersion)
	if len(tagName) == 0 && !useVersionAsTagName {
		return fmt.Errorf("a tag name must be provided as positional argument (usage: ult tag v1.0.0) or pass '--%s'", flagUseVersion)
	}
	logger.Info("creating tag",
		"name", tagName,
		"ref", ref,
		"use_pubspec_version", useVersionAsTagName,
		"project", projectId,
	)

	if projectIdErr != nil {
		return err
	}

	if useVersionAsTagName {
		version, err := fetchVersionFromPubspecFile()
		if err != nil {
			return err
		}

		tagName = version.String()
	}

	appRepo, err := gitlab.NewClient(token)
	if err != nil {
		return err
	}

	opt := &gitlab.CreateTagOptions{
		TagName: gitlab.Ptr(tagName),
		Ref:     gitlab.Ptr(ref),
	}
	tag, _, err := appRepo.Tags.CreateTag(projectId, opt)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully tagged ref: %s\n", tag.Commit.ID)

	return nil
}

func fetchVersionFromPubspecFile() (*version.Version, error) {
	contents, err := os.ReadFile("pubspec.yaml")
	if err != nil {
		logger.Error("Failed to read pubspec.yaml", "error", err)
		return nil, fmt.Errorf("reading pubspec.yaml: %w", err)
	}

	lines := strings.Split(string(contents), "\n")
	logger.Debug("Read pubspec.yaml", "lineCount", len(lines))

	version, _, err := version.FetchFromLines(lines)
	if err != nil {
		logger.Error("Failed to parse version", "error", err)
		return version, fmt.Errorf("parsing version: %w", err)
	}

	return version, err
}
