// Package list provides the command for listing releases that are
// currently live on the Play Store production track.
package list

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"ulist.app/ult/internal/playstore"
)

const (
	flagCredentialsPath = "credentials"
)

var Cmd = cli.Command{
	Name:   "list",
	Usage:  "list all releases in the Play Store production track",
	Action: run,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     flagCredentialsPath,
			Usage:    "path to the Google Play Store service account credentials JSON file",
			Required: true,
		},
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	credentials, err := os.ReadFile(cmd.String(flagCredentialsPath))
	if err != nil {
		return fmt.Errorf("reading credentials file: %w", err)
	}

	releases, err := playstore.GetProductionReleases(credentials, "app.ulist")
	if err != nil {
		return fmt.Errorf("fetching production releases for app.ulist: %w", err)
	}
	if len(releases) == 0 {
		return errors.New("no releases found in the production track")
	}

	for _, release := range releases {
		codes := make([]string, 0, len(release.VersionCodes))
		for _, code := range release.VersionCodes {
			codes = append(codes, fmt.Sprint(code))
		}
		fmt.Printf("Latest PROD Release: %s+%s\n",
			release.Name, strings.Join(codes, ", "))
	}

	return nil
}
