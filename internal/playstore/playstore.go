package playstore

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

var (
	// logger = log.NewWithOptions(os.Stdout, log.Options{
	// 	Level:           log.DebugLevel,
	// 	ReportTimestamp: true,
	// })

	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
)

// Release is a single release on a Play Store track.
type Release struct {
	Name         string
	Status       string
	VersionCodes []int64
}

// GetProductionReleases fetches all releases currently on the production
// track of the given package. It opens a read-only edit, reads the track
// and aborts the edit before returning.
func GetProductionReleases(credFile []byte, packageName string) ([]Release, error) {
	logger.Info("creating credentials")
	ctx := context.Background()
	jwtConfig, err := google.JWTConfigFromJSON(credFile, androidpublisher.AndroidpublisherScope)
	if err != nil {
		return nil, err
	}
	opt := option.WithHTTPClient(jwtConfig.Client(ctx))
	logger.Info("creating android publisher service")
	service, err := androidpublisher.NewService(ctx, opt)
	if err != nil {
		return nil, err
	}

	logger.Info("opening edit")
	edit, err := service.Edits.Insert(packageName, nil).Do()
	if err != nil {
		return nil, err
	}

	logger.Info("fetching releases from production track")
	track, err := service.Edits.Tracks.Get(packageName, edit.Id, "production").Do()
	if err != nil {
		return nil, err
	}

	releases := make([]Release, 0, len(track.Releases))
	for _, release := range track.Releases {
		releases = append(releases, Release{
			Name:         release.Name,
			Status:       release.Status,
			VersionCodes: release.VersionCodes,
		})
	}

	logger.Info("closing open edit")
	if err := service.Edits.Delete(packageName, edit.Id).Do(); err != nil {
		return releases, err
	}

	return releases, nil
}

func GetVersionFromLatestRelease(credFile []byte, packageName string) (int64, error) {
	var latestCode int64

	logger.Info("creating credentials")
	ctx := context.Background()
	jwtConfig, err := google.JWTConfigFromJSON(credFile, androidpublisher.AndroidpublisherScope)
	if err != nil {
		return latestCode, err
	}
	opt := option.WithHTTPClient(jwtConfig.Client(ctx))
	logger.Info("creating android publisher service")
	service, err := androidpublisher.NewService(ctx, opt)
	if err != nil {
		return latestCode, err
	}

	logger.Info("opening edit")
	editReq := service.Edits.Insert(packageName, nil)
	edit, err := editReq.Do()
	if err != nil {
		return latestCode, err
	}

	logger.Info("fetching releases from production track")
	prodReq := service.Edits.Tracks.Get(packageName, edit.Id, "production")
	prodResp, err := prodReq.Do()
	if err != nil {
		return latestCode, err
	}

	for _, release := range prodResp.Releases {
		for _, code := range release.VersionCodes {
			if code <= latestCode {
				continue
			}
			latestCode = code
		}
	}

	logger.Info("fetching releases from internal testing track")
	internalReq := service.Edits.Tracks.Get(packageName, edit.Id, "internal")
	internalResp, err := internalReq.Do()
	if err != nil {
		return latestCode, err
	}

	for _, release := range internalResp.Releases {
		for _, code := range release.VersionCodes {
			if code <= latestCode {
				continue
			}
			latestCode = code
		}
	}

	logger.Info("finished checking releases version code",
		"latestVersionCode", latestCode)

	logger.Info("closing open edit")
	abortEditReq := service.Edits.Delete(packageName, edit.Id)
	err = abortEditReq.Do()
	if err != nil {
		msg := fmt.Sprintf("latest version code: %d\n", latestCode)
		logger.Error(msg)
		return latestCode, err
	}

	return latestCode, nil
}
