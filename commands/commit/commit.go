package commit_command

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
	gitlab "gitlab.com/gitlab-org/api/client-go"
	"ulist.app/ult/internal/core"
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
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	token, err := core.GetToken(cmd)
	if err != nil {
		return err
	}
	projectId, err := core.GetProjectID(cmd)
	if err != nil {
		return err
	}
	filePath := cmd.Args().First()
	commitMessage := cmd.String(flagMessage)
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

	hasChanges, err := git.HasChanges(filePath)
	if err != nil {
		return err
	}
	if !hasChanges {
		fmt.Println("\nThe given file has no changes. Skipping commit!")
		return nil
	}

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
