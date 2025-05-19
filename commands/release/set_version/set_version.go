package set_version

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/urfave/cli/v3"
	"ulist.app/ult/internal/version"
)

const (
	flagPath = "path"
)

var (
	logger = slog.Default().WithGroup("set_version_command")
)

// Cmd defines the version command for CLI
var Cmd = cli.Command{
	Name:   "set-version",
	Usage:  "set a specific version in pubspec file",
	Action: run,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  flagPath,
			Usage: "path of the pubspec.yaml file",
		},
	},
}

func run(ctx context.Context, cmd *cli.Command) error {
	versionArg := cmd.Args().First()
	if len(versionArg) == 0 {
		return errors.New("you need to provide a version as positional argument (usage: ult release set-version 2000.100.10+01)")
	}
	newVersion, err := version.Parse(versionArg)
	if err != nil {
		return err
	}

	pubspecPath := cmd.String(flagPath)
	if len(pubspecPath) == 0 {
		pubspecPath = "pubspec.yaml"
	}

	logger.Debug("Starting set version command",
		"version", newVersion,
		"path", pubspecPath,
	)

	contents, err := os.ReadFile(pubspecPath)
	if err != nil {
		return fmt.Errorf("failed to read pubspec.yaml: %w", err)
	}

	lines := strings.Split(string(contents), "\n")

	_, idx, err := version.FetchFromLines(lines)
	if err != nil {
		return fmt.Errorf("failed to find version string in pubspec file: %w", err)
	}

	// Update version in pubspec lines
	lines[idx] = fmt.Sprintf("version: %s", newVersion)

	// Write back to file with original permissions
	if err := writeFile(pubspecPath, lines); err != nil {
		return err
	}

	fmt.Printf("updated pubspec.yaml with version: %s\n", newVersion)
	return nil
}

// writeFile writes the provided lines back to the specified file path,
// preserving the original file permissions. It returns an error if
// getting file info or writing to the file fails.
func writeFile(path string, lines []string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}

	content := strings.Join(lines, "\n")
	err = os.WriteFile(path, []byte(content), fileInfo.Mode())
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}
