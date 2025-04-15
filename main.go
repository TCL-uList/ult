package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	commit_command "ulist.app/ult/commands/commit"
	"ulist.app/ult/commands/release"
	"ulist.app/ult/commands/secrets"
)

var (
	// set dynamically at build by ldflags parameter
	version string = "dev"
	// set dynamically at build by ldflags parameter
	commit string = "none"
)

func main() {
	versionCmd := cli.Command{
		Name:  "version",
		Usage: "show version information for ult",
		Action: func(ctx context.Context, c *cli.Command) error {
			println(version, commit)
			return nil
		},
	}

	cmd := &cli.Command{
		Name: "ult",
		Usage: "uList Command Line Tools (ult)\n\n" +
			"The official CLI for managing uList application versions and deployments.\n" +
			"Automates version bumping, release tagging, and deployment workflows.",
		Commands: []*cli.Command{
			&release_command.Cmd,
			&secrets_command.Cmd,
			&commit_command.Cmd,
			&versionCmd,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "show logging messages",
			},
			&cli.StringFlag{
				Name:  "token",
				Usage: "token that will be used on http requests",
			},
			&cli.StringFlag{
				Name:    "project-id",
				Aliases: []string{"id"},
				Usage:   "The ID or URL-encoded path of the project",
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
