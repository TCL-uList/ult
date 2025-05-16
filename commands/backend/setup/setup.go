package setup

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/urfave/cli/v3"
	cloudsql "ulist.app/ult/internal/cloud_sql"
)

var (
	logger = slog.Default().WithGroup("setup_command")
)

var Cmd = cli.Command{
	Name:   "setup",
	Usage:  "setup database tables",
	Action: run,
	Flags:  []cli.Flag{},
}

func run(ctx context.Context, cmd *cli.Command) error {
	db, err := cloudsql.ConnectWithConnector()
	if err != nil {
		return err
	}

	err = cloudsql.GenerateTables(db)
	if err != nil {
		return err
	}

	fmt.Println("Successfully generated all tables")

	return nil
}
