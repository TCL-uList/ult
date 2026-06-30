package secrets_command

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"

	"github.com/urfave/cli/v3"
	"ulist.app/ult/internal/core"
	"ulist.app/ult/internal/secrets"
)

// Command flag constants
const (
	flagId       = "id"
	flagFileName = "name"
	flagArchive  = "archive"
)

const (
	defaultSecretsFileName = ".secrets.tar.gz"
)

var (
	lvl    = new(slog.LevelVar)
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	})).WithGroup("secrets_command")
)

var Cmd = cli.Command{
	Name:   "secrets",
	Usage:  "manage secure files stored in the GitLab project (list, delete, update)",
	Action: listSecureFilesCommand,
	Commands: []*cli.Command{
		{
			Name:   "list",
			Usage:  "list all secrets from given project",
			Action: listSecureFilesCommand,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    flagId,
					Aliases: []string{"i"},
					Usage:   "print only the numeric ID of each matching file",
				},
				&cli.StringFlag{
					Name:        flagFileName,
					Usage:       "filter results to files matching this name",
					DefaultText: ".secrets.tar.gz",
				},
			},
		},
		{
			Name:   "delete",
			Usage:  "delete a secure file from the project",
			Action: deleteSecureFileCommand,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    flagId,
					Aliases: []string{"i"},
					Usage:   "print only the numeric ID of each matching file",
				},
				&cli.StringFlag{
					Name:        flagFileName,
					Usage:       "name of the secure file to delete",
					DefaultText: ".secrets.tar.gz",
				},
			},
		},
		{
			Name:   "update",
			Usage:  "replace the existing secrets archive with a new one",
			Action: updateSecureFileCommand,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        flagArchive,
					Aliases:     []string{"a"},
					Usage:       "path to the .secrets.tar.gz archive to upload",
					HideDefault: true,
				},
			},
		},
	},
}

func setLoggingVerbosity(verbose bool) {
	if verbose {
		lvl.Set(slog.LevelInfo)
	} else {
		lvl.Set(slog.LevelError)
	}
}

func listSecureFilesCommand(ctx context.Context, cmd *cli.Command) error {
	setLoggingVerbosity(cmd.Bool(core.VerboseFlag))
	token, err := core.GetToken(cmd)
	if err != nil {
		return err
	}

	projectId, err := core.GetProjectID(cmd)
	if err != nil {
		return err
	}

	targetName := cmd.String(flagFileName)
	showOnlyId := cmd.Bool(flagId)
	logger.Info("Fetching secure files", "with name", targetName, "show only id", showOnlyId)

	files, _, err := secrets.FetchAll(projectId, token)
	if err != nil {
		return err
	}

	for _, file := range files {
		secrets.PrintSecureFile(file, showOnlyId, targetName)
	}

	return nil
}

func deleteSecureFileCommand(ctx context.Context, cmd *cli.Command) error {
	setLoggingVerbosity(cmd.Bool(core.VerboseFlag))
	token, err := core.GetToken(cmd)
	if err != nil {
		return err
	}

	projectId, err := core.GetProjectID(cmd)
	if err != nil {
		return err
	}

	targetName := cmd.String(flagFileName)
	showOnlyId := cmd.Bool(flagId)
	logger.Info("Fetching secure files", "with name", targetName, "show only id", showOnlyId)

	files, appRepo, err := secrets.FetchAll(projectId, token)
	if err != nil {
		return err
	}

	foundFile, file := secrets.GetSecureFile(files, targetName)
	if !foundFile {
		return fmt.Errorf("No secure file was found with given name: %s", targetName)
	}

	err = secrets.Delete(appRepo, file.ID, targetName, projectId)
	if err != nil {
		return err
	}

	return nil
}

func updateSecureFileCommand(ctx context.Context, cmd *cli.Command) error {
	setLoggingVerbosity(cmd.Bool(core.VerboseFlag))
	token, err := core.GetToken(cmd)
	if err != nil {
		return err
	}

	projectId, err := core.GetProjectID(cmd)
	if err != nil {
		return err
	}

	archivePath := cmd.String(flagArchive)
	if len(archivePath) == 0 {
		return fmt.Errorf("archive path can not be empty, you must pass as argument '--archive=path' or '-a=path'")
	}
	if path.Base(archivePath) != defaultSecretsFileName {
		return fmt.Errorf("the archive must be named '%s'", defaultSecretsFileName)
	}

	path, err := filepath.Abs(archivePath)
	if err != nil {
		return fmt.Errorf("Failed to get absolute path for %s: %v", archivePath, err)
	}
	logger.Info("Using absolute path for archive upload", "path", path)

	files, appRepo, err := secrets.FetchAll(projectId, token)
	if err != nil {
		return err
	}

	logger.Info("Looking for secure file", "name", defaultSecretsFileName)
	foundFile, file := secrets.GetSecureFile(files, defaultSecretsFileName)
	if !foundFile {
		return fmt.Errorf("No secure file was found with given name: %s", path)
	}
	logger.Info("Found secure file", "id", file.ID, "name", defaultSecretsFileName)

	logger.Info("Deleting existing secure file", "id", file.ID)
	err = secrets.Delete(appRepo, file.ID, path, projectId)
	if err != nil {
		return err
	}
	logger.Info("Successfully deleted secure file", "id", file.ID)

	logger.Info("Creating new secure file", "project_id", projectId)
	err = secrets.Create(appRepo, path, projectId)
	if err != nil {
		return err
	}
	logger.Info("Successfully updated secure file")
	return nil
}
