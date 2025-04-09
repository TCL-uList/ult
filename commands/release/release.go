// Package version_command provides functionality for managing version numbers
// in pubspec.yaml files, with options for committing and tagging changes.
package release_command

import (
	"github.com/urfave/cli/v3"
	"ulist.app/ult/commands/release/bump"
)

const (
	flagBump = "bump"
)

var Cmd = cli.Command{
	Name:  "release",
	Usage: "increment the app version number",
	Commands: []*cli.Command{
		&bump.Cmd,
	},
}
