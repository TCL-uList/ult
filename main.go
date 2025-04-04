package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"ulist.app/ult/commands/release"
)

func main() {
	cmd := &cli.Command{
		Name: "ult",
		Usage: "uList Command Line Tools (ult)\n\n" +
			"The official CLI for managing uList application versions and deployments.\n" +
			"Automates version bumping, release tagging, and deployment workflows.",
		Commands: []*cli.Command{
			&release_command.Cmd,
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
