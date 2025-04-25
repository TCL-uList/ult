package appdistribution

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/firebaseappdistribution/v1"
)

type ReleaseInfo struct {
	Name string
}

var (
	logger = log.NewWithOptions(os.Stdout, log.Options{
		Level:           log.DebugLevel,
		ReportTimestamp: true,
	})
)

func CreateReleaseWithFile(ctx context.Context, binary *os.File, appName string, jwt *jwt.Config) (*ReleaseInfo, error) {
	v1 := "https://firebaseappdistribution.googleapis.com/upload/v1"
	uploadURL := fmt.Sprintf("%s/%s/releases:upload", v1, appName)

	// setting up token
	token, err := jwt.TokenSource(ctx).Token()
	if err != nil {
		return nil, fmt.Errorf("Not able to generate token from jwt: %w", err)
	}

	// setting up content type
	logger.Info("extracting extension from file name")
	extension := strings.Split(binary.Name(), ".")[1]
	contentType, err := contentTypeByExtension(extension)
	if err != nil {
		return nil, err
	}
	logger.Debug("got content type by extension",
		"extension", extension,
		"content type", contentType)

	// generating http request
	logger.Info("generating http request")
	uploadCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	fileInfo, err := binary.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	logger.Debug("file size", "bytes", fileInfo.Size())
	progressReader := &ProgressPrinter{Reader: binary, Total: fileInfo.Size()}
	req, err := http.NewRequestWithContext(uploadCtx, "POST", uploadURL, progressReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("X-Goog-Upload-Protocol", "raw")

	// opening a new release by uploading a binary
	logger.Info("requesting to create a new release with given build file")
	client := jwt.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()
	logger.Info("finished uploading file")

	// parsing response to fetch operation name
	var operation ReleaseInfo
	err = json.NewDecoder(resp.Body).Decode(&operation)
	if err != nil {
		return nil, fmt.Errorf("failed to decode release info: %w", err)
	}
	logger.Info("a operation for a new release was created")
	logger.Debug("operation from request", "operation", operation.Name)

	return &operation, nil
}

func PollOperation(name string, service *firebaseappdistribution.Service) (*Release, error) {
	logger.Info("polling operation for most recent release created")
	var resp *firebaseappdistribution.GoogleLongrunningOperation
	operationCall := service.Projects.Apps.Releases.Operations.Get(name)

	logger.Info("checking release operation status, please wait...")
	for {
		printLoadingSpinner()
		response, err := operationCall.Do()
		if err != nil {
			return nil, fmt.Errorf("get request on operation endpoint failed: %w", err)
		}

		resp = response

		if resp.Done {
			cleanLoadingSpinner()
			logger.Info("polling completed, operation has finished")
			break
		}

		time.Sleep(300 * time.Millisecond)
	}

	logger.Info("parsing release information")
	var result UploadReleaseResult
	err := json.Unmarshal(resp.Response, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal operation response: %w", err)
	}

	logger.Debug("finished polling operation",
		"done", resp.Done,
		"release_name", result.Release.Name,
	)
	return &result.Release, nil
}

func AddReleaseNotesToRelease(releaseName string, notes string, service *firebaseappdistribution.Service) error {
	logger.Info("updating release with release notes")
	releaseNotes := firebaseappdistribution.GoogleFirebaseAppdistroV1ReleaseNotes{Text: notes}
	req := firebaseappdistribution.GoogleFirebaseAppdistroV1Release{ReleaseNotes: &releaseNotes}
	_, err := service.Projects.Apps.Releases.Patch(releaseName, &req).Do()
	if err != nil {
		return fmt.Errorf("failed to patch release with release notes: %w", err)
	}

	logger.Info("successfully updated release with given notes")
	return nil
}

func DistributeRelease(releaseName string, groups []string, service *firebaseappdistribution.Service) error {
	logger.Info("distributing the app", "groups", groups)
	distReq := firebaseappdistribution.GoogleFirebaseAppdistroV1DistributeReleaseRequest{
		GroupAliases: groups,
	}
	_, err := service.Projects.Apps.Releases.Distribute(releaseName, &distReq).Do()
	if err != nil {
		return err
	}

	logger.Info("successfully created a new release and distributed")
	return nil
}
