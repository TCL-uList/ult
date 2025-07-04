package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	backend_command "ulist.app/ult/commands/backend"
	commit_command "ulist.app/ult/commands/commit"
	release_command "ulist.app/ult/commands/release"
	secrets_command "ulist.app/ult/commands/secrets"
	tag_command "ulist.app/ult/commands/tag"
	"ulist.app/ult/internal/core"
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
			fmt.Println(version, commit)
			return nil
		},
	}

	commands := []*cli.Command{
		&release_command.Cmd,
		&secrets_command.Cmd,
		&commit_command.Cmd,
		&tag_command.Cmd,
		&versionCmd,
	}

	// only dev builds can access the backend command
	if commit == "none" {
		commands = append(commands, &backend_command.Cmd)
	}

	cmd := &cli.Command{
		Name: "ult",
		Usage: "uList Command Line Tools (ult)\n\n" +
			"The official CLI for managing uList application versions and deployments.\n" +
			"Automates version bumping, release tagging, and deployment workflows.",
		Commands: commands,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    core.VerboseFlag,
				Aliases: []string{"v"},
				Usage:   "show logging messages",
			},
			&cli.StringFlag{
				Name:  core.TokenFlag,
				Usage: "token that will be used on http requests",
			},
			&cli.StringFlag{
				Name:    core.ProjectIDFlag,
				Aliases: []string{"id"},
				Usage:   "The ID or URL-encoded path of the project",
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
