package deploy

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/firebaseappdistribution/v1"
	"google.golang.org/api/option"
	appdistribution "ulist.app/ult/internal/app_distribution"
)

const (
	flagAppId        = "app"
	flagGroups       = "groups"
	flagJsonKey      = "json-key"
	flagReleaseNotes = "release-notes"
)

var (
	// logger = slog.Default().WithGroup("deploy_command")
	logger = log.NewWithOptions(os.Stdout, log.Options{
		Level:           log.DebugLevel,
		ReportTimestamp: true,
	})
)

var Cmd = cli.Command{
	Name:   "deploy",
	Usage:  "upload a release binary and optionally distribute it to testers",
	Action: run,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     flagAppId,
			Usage:    "the app id of your Firebase app",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:     flagGroups,
			Usage:    "a comma-separated list of group aliases to distribute to",
			Required: true,
		},
		&cli.StringFlag{
			Name:     flagJsonKey,
			Usage:    "path to json key",
			Required: true,
		},
		&cli.StringFlag{
			Name:  flagReleaseNotes,
			Usage: "release notes to include",
		},
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	jsonKeyPath := cmd.String(flagJsonKey)
	appID := strings.Trim(cmd.String(flagAppId), "\"")
	groups := cmd.StringSlice(flagGroups)
	notes := cmd.String(flagReleaseNotes)
	verbose := cmd.Bool("verbose")

	logger.Info("Starting deploy command",
		"build file type", cmd.Args().First(),
		"app_id", appID,
		"groups", groups,
		"notes", notes,
	)

	buildFilePath := strings.ToLower(cmd.Args().First())
	if len(buildFilePath) == 0 {
		return errors.New("No deploy build file found. Usage: ult release deploy path/to/build_file.apk")
	}
	buildExtension := strings.Split(buildFilePath, ".")[1]
	if buildExtension != "apk" && buildExtension != "aab" && buildExtension != "ipa" {
		return errors.New("Invalid build file extension, can only be one of the following: apk, aab or ipa\nUsage: ult release deploy path/to/build_file.apk")
	}

	keyContents, err := os.ReadFile(jsonKeyPath)
	if err != nil {
		logger.Error("Failed to read json key file", "error", err)
		return fmt.Errorf("reading json key file: %w", err)
	}

	buildFile, err := os.Open(buildFilePath)
	if err != nil {
		return err
	}
	defer buildFile.Close()

	logger.Info("starting release distribution")

	if len(groups) == 0 {
		return errors.New("groups parameter cannot be empty")
	}

	logger.Info("creating credentials")
	jwt, err := google.JWTConfigFromJSON(keyContents, firebaseappdistribution.CloudPlatformScope)
	if err != nil {
		return err
	}
	opts := []option.ClientOption{
		option.WithHTTPClient(jwt.Client(ctx)),
		option.WithScopes(firebaseappdistribution.CloudPlatformScope),
	}
	if verbose {
		l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
		opts = append(opts, option.WithLogger(l))
	}
	logger.Info("creating firebase app distribution service")
	service, err := firebaseappdistribution.NewService(ctx, opts...)
	if err != nil {
		return err
	}

	logger.Info("configuring app name")
	appNameParts := strings.Split(appID, ":")
	if len(appNameParts) != 4 {
		return fmt.Errorf("invalid app ID format, expected 1:project_number:platform:hash, got %s", appID)
	}
	projectID := string(appNameParts[1])
	app := fmt.Sprintf("projects/%s/apps/%s", projectID, appID)

	operation, err := appdistribution.CreateReleaseWithFile(ctx, buildFile, app, jwt)
	if err != nil {
		return err
	}

	release, err := appdistribution.PollOperation(operation.Name, service)
	if err != nil {
		return err
	}

	if len(notes) > 0 {
		err := appdistribution.AddReleaseNotesToRelease(release.Name, notes, service)
		if err != nil {
			return err
		}
	}

	err = appdistribution.DistributeRelease(release.Name, groups, service)
	if err != nil {
		return err
	}

	logger.Info("successfully created and distributed release")
	return nil
}
