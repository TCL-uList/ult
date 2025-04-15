package utils

import (
	"fmt"

	"github.com/urfave/cli/v3"
)

func GetOrError(key string, cmd *cli.Command) (string, error) {
	value := cmd.String(key)
	if len(value) == 0 {
		return "", fmt.Errorf("%s is an requied argument", key)
	}
	return value, nil
}
