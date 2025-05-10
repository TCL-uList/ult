package release

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ulist.app/ult/internal/assignee"
	"ulist.app/ult/internal/version"
)

type Release struct {
	Branch         string
	Assignee       assignee.Assignee
	Description    string
	Commit         string
	Date           time.Time
	IssueTrackerID int
	Version        version.Version
}

func (r Release) String() string {
	return fmt.Sprintf(
		"Branch: %s, Assignee: %s, Description: %s, Commit: %s, Date: %s, Issue tracker ID: %d, Version: %s, Bump: %d",
		r.Branch,
		r.Assignee,
		r.Description,
		r.Commit,
		r.Date,
		r.IssueTrackerID,
		r.Version.StringNoBuild(),
		r.Bump(),
	)
}

func (r Release) Bump() int {
	return r.Version.Build
}

func SaveAssignee(db *sql.DB, assignee assignee.Assignee) (int, error) {
	if db == nil {
		return -1, errors.New("database connection is nil")
	}

	query := `
  WITH insert_attempt AS (
    INSERT INTO assignees (name, email)
    VALUES ($1, $2)
    ON CONFLICT DO NOTHING
    RETURNING id
  )
  SELECT id FROM insert_attempt
  UNION ALL
  SELECT id FROM assignees WHERE email = $2
  LIMIT 1;
  `

	var id int
	err := db.QueryRow(query, assignee.Name, assignee.Email).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("failed to create assignee in database: %w", err)
	}

	return id, nil
}

func SaveVersion(db *sql.DB, version version.Version) (int, error) {
	if db == nil {
		return -1, errors.New("database connection is nil")
	}

	query := `
  WITH insert_attempt AS (
    INSERT INTO versions (year, major, minor)
    VALUES ($1, $2, $3)
    ON CONFLICT DO NOTHING
    RETURNING id
  )
  SELECT id FROM insert_attempt
  UNION ALL
  SELECT id FROM versions WHERE year = $1 AND major = $2 AND minor = $3
  LIMIT 1;
  `

	var id int
	err := db.QueryRow(query, version.Year, version.Major, version.Minor).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("failed to create version in database: %w", err)
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

	assigneeID, err := SaveAssignee(db, release.Assignee)
	if err != nil {
		return err
	}

	versionID, err := SaveVersion(db, release.Version)
	if err != nil {
		return err
	}

	// Insert the release into the database
	query := `
	INSERT INTO releases (
		branch,
		assignee_id,
		description,
    commit,
		date,
		issue_tracker_id,
		version_id,
		bump
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	err = db.QueryRow(
		query,
		release.Branch,
		assigneeID,
		release.Description,
		release.Commit,
		release.Date,
		release.IssueTrackerID,
		versionID,
		release.Bump(),
	).Err()
	if err != nil {
		return fmt.Errorf("failed to create release in database: %w", err)
	}

	return nil
}

func FetchReleases() []*Release {
	return []*Release{}
}

func FetchLatestRelease(db *sql.DB) (*Release, error) {
	var branch string
	var assigneeName string
	var assigneeEmail string
	var description string
	var commit string
	var date string
	var issueTrackerId int
	var year int
	var major int
	var minor int
	var bump int

	query := `
  SELECT
      r.branch,
      a.name,
      a.email,
      r.description,
      r.commit,
      r.date,
      r.issue_tracker_id,
      v.year,
      v.major,
      v.minor,
      r.bump
  FROM
      releases r
  JOIN
      versions v ON r.version_id = v.id
  JOIN
      assignees a ON r.assignee_id = a.id
  ORDER BY
      v.year DESC,
      v.major DESC,
      v.minor DESC,
      r.bump DESC
  LIMIT 1;`
	err := db.QueryRow(query).Scan(&branch, &assigneeName, &assigneeEmail, &description, &commit, &date, &issueTrackerId, &year, &major, &minor, &bump)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release from database: %w", err)
	}

	dateTime, err := time.Parse("2006-01-02T15:04:05Z0700", date)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date from database query result: %w", err)
	}

	release := &Release{Branch: branch,
		Assignee:       assignee.Assignee{Name: assigneeName, Email: assigneeEmail},
		Description:    description,
		Commit:         commit,
		Date:           dateTime,
		IssueTrackerID: issueTrackerId,
		Version:        version.Version{Year: year, Major: major, Minor: minor, Build: bump},
	}
	return release, nil
}
