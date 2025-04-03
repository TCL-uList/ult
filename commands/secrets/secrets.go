package secrets_command

import (
	"context"

	"github.com/urfave/cli/v3"
)

// Command flag constants
const (
	flagBump        = "bump"
	flagFetch       = "fetch"
	flagNoCommitTag = "no-commit-tag"
	flagNoPush      = "no-push"
)

// Cmd defines the version command for CLI
var Cmd = cli.Command{
	Name:   "secrets",
	Usage:  "increment the app version number",
	Action: run,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  flagFetch,
			Usage: "fetch latest build number from Play Store before incrementing",
		},
		&cli.StringFlag{
			Name:  flagBump,
			Usage: "increment the version in pubspec choosing one of: build, patch, minor or major",
		},
		&cli.BoolFlag{
			Name:  flagNoCommitTag,
			Usage: "skip creating Git commits and tag for version changes. Use with --no-push for dry-run simulations",
		},
		&cli.BoolFlag{
			Name:  flagNoPush,
			Usage: "skip pushing changes to remote repository. Changes will remain local for verification",
		},
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	return nil
}
