package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"ulist.app/ult/commands/release"
	"ulist.app/ult/commands/secrets"
)

var (
	// version is set dynamically at build
	version = "undefined"
)

func main() {
	versionCmd := cli.Command{
		Name:  "version",
		Usage: "show version information for ult",
		Action: func(ctx context.Context, c *cli.Command) error {
			println(version)
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
			&versionCmd,
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "show logging messages",
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
