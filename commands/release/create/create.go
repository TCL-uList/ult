package create

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/urfave/cli/v3"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"ulist.app/ult/internal/assignee"
	cloudsql "ulist.app/ult/internal/cloud_sql"
	"ulist.app/ult/internal/git"
	"ulist.app/ult/internal/release"
	"ulist.app/ult/internal/utils"
	"ulist.app/ult/internal/version"
)

const (
	flagLatest         = "latest"
	flagFromCommit     = "from-commit"
	flagIssueTrackerID = "issue"
	flagBranch         = "branch"
	flagVersion        = "version"
	flagApi            = "api"
)

var (
	logger = slog.Default().WithGroup("bump_command")
)

var fromCommitCmd = cli.Command{
	Name:   "from-commit",
	Usage:  "create a release from a commit sha256 hash",
	Action: runFromCommit,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    flagLatest,
			Usage:   "use the latest commit in current repository",
			Aliases: []string{"l"},
		},
		&cli.StringFlag{
			Name:     flagBranch,
			Usage:    "name of the branch that generated this release (required)",
			Aliases:  []string{"b"},
			Required: true,
		},
		&cli.IntFlag{
			Name:     flagIssueTrackerID,
			Usage:    "issue tracker id related with this release (required)",
			Aliases:  []string{"i"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    flagVersion,
			Usage:   "the version of this release (if ommited, will try to fetch from pubspec.yaml file)",
			Aliases: []string{"v"},
		},
		&cli.BoolFlag{
			Name:    flagApi,
			Usage:   "fetch information from gitlab api",
			Aliases: []string{"a"},
		},
	},
}

var Cmd = cli.Command{
	Name:   "create",
	Usage:  "create a new release",
	Action: run,
	Flags:  []cli.Flag{},
	Commands: []*cli.Command{
		&fromCommitCmd,
	},
}

func runFromCommit(ctx context.Context, cmd *cli.Command) error {
	useLatest := cmd.Bool(flagLatest)

	hash := cmd.Args().First()
	if len(hash) == 0 && !useLatest {
		return errors.New("a commit 'hash' must be passesed as positional argument or the flag '--latest'")
	}

	var err error
	branch := cmd.String(flagBranch)
	issueTrackerID := cmd.Int(flagIssueTrackerID)
	versionStr := cmd.String(flagVersion)
	api := cmd.Bool(flagApi)
	version, err := version.Parse(versionStr)
	if err != nil {
		return err
	}

	var commit *git.Commit
	if api {
		projectId, err := utils.GetOrError("project-id", cmd)
		if err != nil {
			return err
		}
		token, err := utils.GetOrError("token", cmd)
		if err != nil {
			return err
		}

		repo, err := gitlab.NewClient(token)
		if err != nil {
			return err
		}

		gitlabCommit, _, err := repo.Commits.GetCommit(projectId, hash, &gitlab.GetCommitOptions{})
		if err != nil {
			return err
		}

		commit = &git.Commit{
			Assignee: assignee.Assignee{Name: gitlabCommit.AuthorName, Email: gitlabCommit.AuthorEmail},
			Hash:     hash,
			Date:     *gitlabCommit.CommittedDate,
			Message:  gitlabCommit.Message,
		}
	} else {
		commit, err = getCommitFromGitCommand(hash)
		if err != nil {
			return err
		}
	}

	releaseEn := &release.Release{
		Branch:         branch,
		Assignee:       commit.Assignee,
		Description:    commit.Message,
		Commit:         commit.Hash,
		IssueTrackerID: int(issueTrackerID),
		Version:        *version,
		Date:           time.Now(),
	}

	db, err := cloudsql.ConnectWithConnector()
	if err != nil {
		return fmt.Errorf("not able to connect with database to create release: %w", err)
	}

	err = release.SaveRelease(db, releaseEn)
	if err != nil {
		return err
	}

	logger.Info("Created release successfully")
	return nil
}

func getCommitFromGitCommand(hash string) (*git.Commit, error) {
	var err error
	var commit *git.Commit

	if len(hash) == 0 {
		commit, err = git.GetLatestCommitInfo()
		if err != nil {
			return nil, err
		}
	} else {
		commit, err = git.GetCommitInfo(hash)
		if err != nil {
			return nil, err
		}
	}

	return commit, nil
}

func run(ctx context.Context, cmd *cli.Command) error {
	return errors.New("must implement create release accepting all parameters as arguments")
}
