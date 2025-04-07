package secrets

import (
	"fmt"
	"io"
	"os"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"ulist.app/ult/internal/tokens"
)

func FetchAll(projectId string) ([]*gitlab.SecureFile, *gitlab.Client, error) {
	appRepo, err := gitlab.NewClient(tokens.AppApi)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create client: %v", err)
	}

	opt := &gitlab.ListProjectSecureFilesOptions{}
	files, _, err := appRepo.SecureFiles.ListProjectSecureFiles(projectId, opt)
	if err != nil {
		return nil, nil, fmt.Errorf("Error fetching secure files from gitlab: %v", err)
	}

	return files, appRepo, nil
}

func Delete(client *gitlab.Client, id int, targetName string, projectId string) error {
	resp, err := client.SecureFiles.RemoveSecureFile(projectId, id)
	if err != nil {
		return fmt.Errorf("Not able to delete the secure file (%s): %v", targetName, err)
	}

	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Not able to delete the secure file named (%s). Response status code: (%d).\n Not able to read response body: %v", targetName, resp.StatusCode, err)
		}
		body := string(bytes)
		return fmt.Errorf("Not able to delete the secure file named (%s). Response status code: (%d).\n Body: %s", targetName, resp.StatusCode, body)
	}

	return nil
}

func Create(client *gitlab.Client, path string, projectId string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Not able to open the given secrets archive (%s): %v", path, err)
	}

	opt := &gitlab.CreateSecureFileOptions{
		Name: gitlab.Ptr(".secrets.tar.gz"),
	}
	_, resp, err := client.SecureFiles.CreateSecureFile(projectId, file, opt)
	if err != nil {
		return fmt.Errorf("API error when creating secure file: %v", err)
	}

	if resp.StatusCode >= 300 {
		defer resp.Body.Close()
		bytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Not able to upload the secure file (%s). Response status code: (%d).\n Not able to read response body: %v", path, resp.StatusCode, err)
		}
		body := string(bytes)
		return fmt.Errorf("Not able to upload the secure file (%s). Response status code: (%d).\n Body: %s", path, resp.StatusCode, body)
	}

	return nil
}

func GetSecureFileId(files []*gitlab.SecureFile, targetFileName string) (bool, int) {
	var foundFile = false
	var id int

	for _, file := range files {
		if file.Name != targetFileName {
			continue
		}

		foundFile = true
		id = file.ID
	}

	if !foundFile {
		return false, -1
	}

	return true, id
}
