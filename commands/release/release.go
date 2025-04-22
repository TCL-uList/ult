// Package version_command provides functionality for managing version numbers
// in pubspec.yaml files, with options for committing and tagging changes.
package release_command

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	"ulist.app/ult/commands/release/bump"
	"ulist.app/ult/commands/release/create"
	"ulist.app/ult/internal/google"
)

const (
	flagBump = "bump"
)

var Cmd = cli.Command{
	Name:   "release",
	Usage:  "increment the app version number",
	Action: run,
	Commands: []*cli.Command{
		&bump.Cmd,
		&create.Cmd,
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	contents, err := os.ReadFile(".secrets/prod/GoogleApplicationCredentials-ulist.json")
	if err != nil {
		return err
	}
	latestBuild, err := google.GetVersionFromLatestRelease(contents, "app.ulist")
	if err != nil {
		return err
	}

	println(latestBuild)
	return err
}
