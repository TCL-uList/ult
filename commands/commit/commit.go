package commit_command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"ulist.app/ult/internal/git"
)

const (
	flagBranch    = "branch"
	flagMessage   = "message"
	flagFiles     = "files"
	flagProjectId = "id"
)

var (
	logger = slog.Default().WithGroup("commit_command")
)

var Cmd = cli.Command{
	Name:   "commit",
	Usage:  "commit files to the repo",
	Action: run,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    flagBranch,
			Aliases: []string{"b"},
			Usage:   "name of the branch to commit into",
		},
		&cli.StringFlag{
			Name:    flagMessage,
			Aliases: []string{"m"},
			Usage:   "commit message",
		},
		&cli.StringFlag{
			Name:  flagProjectId,
			Usage: "The ID or URL-encoded path of the project",
		},
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	token := cmd.String("token")
	if len(token) == 0 {
		return errors.New("token cannot be empty - use `--token='your-token'`")
	}

	projectId := cmd.String(flagProjectId)
	if len(projectId) == 0 {
		return fmt.Errorf("Project id cannot be empty, use: --%s='actual-id'", flagProjectId)
	}

	filePath := cmd.Args().First()
	if len(filePath) == 0 {
		return errors.New("file name is required as positional argument: `ult commit filename`")
	}

	commitMessage := cmd.String(flagMessage)
	if len(commitMessage) == 0 {
		return errors.New("commit message is required - use `--message='your message'`")
	}

	branch := cmd.String(flagBranch)
	if len(branch) == 0 {
		currentBranch, err := git.GetCurrentBranch()
		if err != nil {
			return err
		}
		branch = currentBranch
		logger.Info("no branch passed, will use the current one", "branch", branch)
	}
	logger.Info("updating file",
		"file", filePath,
		"project", projectId,
		"branch", branch,
		"commit", commitMessage,
	)

	appRepo, err := gitlab.NewClient(token)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	opt := gitlab.CreateCommitOptions{
		AuthorName:    gitlab.Ptr(git.Name),
		AuthorEmail:   gitlab.Ptr(git.Email),
		CommitMessage: &commitMessage,
		Branch:        &branch,
		Actions: []*gitlab.CommitActionOptions{
			{
				Action:   gitlab.Ptr(gitlab.FileUpdate),
				FilePath: &filePath,
				Content:  gitlab.Ptr(string(content)),
			},
		},
	}

	commit, _, err := appRepo.Commits.CreateCommit(projectId, &opt)
	if err != nil {
		return fmt.Errorf("Error while trying to update %s file in repo: %w", filePath, err)
	}

	fmt.Printf("Successfully commit updated file: %s\n", commit.String())

	return nil
}
