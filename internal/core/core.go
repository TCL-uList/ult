package core

import (
	"errors"

	"github.com/urfave/cli/v3"
)

const (
	VerboseFlag   = "verbose"
	TokenFlag     = "token"
	ProjectIDFlag = "project-id"
)

// GetToken extracts the authentication token from command line arguments.
// It retrieves the token value using the TokenFlag and validates that it is not empty.
// Returns the token string if present, otherwise returns an error with usage instructions.
func GetToken(cmd *cli.Command) (string, error) {
	token := cmd.String(TokenFlag)
	if len(token) == 0 {
		return "", errors.New("Argument 'token' not found but is required.\nUsage: '--token=yourtokenhere'")
	}
	return token, nil
}

// GetProjectID extracts the project ID from command line arguments.
// It retrieves the project ID value using the ProjectIDFlag and validates that it is not empty.
// Returns the project ID string if present, otherwise returns an error with usage instructions.
func GetProjectID(cmd *cli.Command) (string, error) {
	projectID := cmd.String(ProjectIDFlag)
	if len(projectID) == 0 {
		return "", errors.New("Argument 'project-id' not found but is required.\nUsage: '--project-id=0000000' or '--id=0000000'")
	}
	return projectID, nil
}
