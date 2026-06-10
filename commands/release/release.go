// Package version_command provides functionality for managing version numbers
// in pubspec.yaml files, with options for committing and tagging changes.
package release_command

import (
	"github.com/urfave/cli/v3"
	"ulist.app/ult/commands/release/bump"
	"ulist.app/ult/commands/release/create"
	"ulist.app/ult/commands/release/deploy"
	"ulist.app/ult/commands/release/list"
	"ulist.app/ult/commands/release/set_version"
)

const (
	flagBump = "bump"
)

var Cmd = cli.Command{
	Name:  "release",
	Usage: "manage app releases: bump version, create release records, and deploy builds",
	Commands: []*cli.Command{
		&bump.Cmd,
		&create.Cmd,
		&deploy.Cmd,
		&list.Cmd,
		&set_version.Cmd,
	},
}
