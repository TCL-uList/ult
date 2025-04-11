package release

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Release struct {
	Branch         string
	Assignee       string
	Description    string
	Date           time.Time
	IssueTrackerID int
	Versioning     int
	Version        string
	Bump           int
}

func SaveVersion(db *sql.DB, version string) (int, error) {
	if db == nil {
		return -1, errors.New("database connection is nil")
	}

	if len(version) == 0 {
		return -1, errors.New("version can not be empty")
	}

	query := `
  WITH insert_attempt AS (
    INSERT INTO versions (version)
    VALUES ($1)
    ON CONFLICT (version) DO NOTHING
    RETURNING id
  )
  SELECT id FROM insert_attempt
  UNION ALL
  SELECT id FROM versions WHERE version = $1
  LIMIT 1;
  `

	var id int
	err := db.QueryRow(query, version).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("failed to create version: %w", err)
	}

	return id, nil
}

func SaveRelease(db *sql.DB, release *Release) error {
	if db == nil {
		return errors.New("database connection is nil")
	}

	if release == nil {
		return errors.New("release is nil")
	}

	if release.Branch == "" {
		return errors.New("branch is required")
	}

	if release.Assignee == "" {
		return errors.New("assignee is required")
	}

	if release.Version == "" {
		return errors.New("version is required")
	}

	// Insert the release into the database
	query := `
	INSERT INTO releases (
		branch,
		assignee,
		description,
		date,
		issue_tracker_id,
		version_id,
		bump
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	var id int
	err := db.QueryRow(
		query,
		release.Branch,
		release.Assignee,
		release.Description,
		release.Date,
		release.IssueTrackerID,
		release.Versioning,
		release.Version,
		release.Bump,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to create release: %w", err)
	}

	return nil
}

func FetchReleases() []*Release {
	return []*Release{}
}

func FetchLatestRelease(db *sql.DB) (*Release, error) {
	var branch string
	var assignee string
	var description string
	var date string
	var issueTrackerId int
	var version string
	var bump int

	query := `
  SELECT 
      r.branch, 
      r.assignee, 
      r.description, 
      r.date, 
      r.issue_tracker_id, 
      v.version,
      r.bump 
  FROM 
      releases r
  JOIN 
      versions v ON r.version_id = v.id
  ORDER BY 
      r.bump DESC 
  LIMIT 1;
  `
	err := db.QueryRow(query).Scan(&branch, &assignee, &description, &date, &issueTrackerId, &version, &bump)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}

	dateTime, err := time.Parse("2006-01-02T15:04:05Z0700", date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date from database query result: %w", err)
	}

	release := &Release{Branch: branch,
		Assignee:       assignee,
		Description:    description,
		Date:           dateTime,
		IssueTrackerID: issueTrackerId,
		Version:        version,
		Bump:           bump,
	}
	return release, nil
}
