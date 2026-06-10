package backend_command

import (
	"github.com/urfave/cli/v3"
	"ulist.app/ult/commands/backend/setup"
)

var Cmd = cli.Command{
	Name:  "backend",
	Usage: "backend infrastructure utilities (dev only)",
	Commands: []*cli.Command{
		&setup.Cmd,
	},
}
