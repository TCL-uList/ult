package cloudsql

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// ConnectWithConnector connects to Cloud SQL instance
func ConnectWithConnector() (*sql.DB, error) {
	var (
		dbUser = os.Getenv("ULT_DB_USER")
		dbPwd  = os.Getenv("ULT_DB_PASS")
		dbHost = os.Getenv("ULT_DB_HOST")
	)

	if len(dbUser) == 0 {
		panic("ULT_DB_USER variable must not be empty")
	}
	if len(dbPwd) == 0 {
		panic("ULT_DB_PASS variable must not be empty")
	}
	if len(dbHost) == 0 {
		panic("ULT_DB_HOST variable must not be empty")
	}

	databaseURL := fmt.Sprintf("postgresql://%s:%s@%s?sslmode=require", dbUser, dbPwd, dbHost)
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	return db, nil
}

// GenerateTables creates the database table for Release entities.
// It returns an error if the table creation fails.
func GenerateTables(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	createAssigneesTableQuery := `
	CREATE TABLE IF NOT EXISTS assignees (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE
	);`
	fmt.Printf("creating table 'assignees': %s\n\n", createAssigneesTableQuery)
	_, err := db.Exec(createAssigneesTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create assignees table: %w", err)
	}

	// Create the releases table
	createVersionsTableQuery := `
	CREATE TABLE IF NOT EXISTS versions (
		id SERIAL PRIMARY KEY,
		year INTEGER NOT NULL,
		major INTEGER NOT NULL,
		minor INTEGER NOT NULL,
    UNIQUE (year, major, minor)
	);`
	fmt.Printf("creating table 'versions': %s\n\n", createVersionsTableQuery)
	_, err = db.Exec(createVersionsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create versions table: %w", err)
	}

	idx_versions_sort := `CREATE INDEX IF NOT EXISTS idx_versions_sort ON versions (year DESC, major DESC, minor DESC);`
	fmt.Printf("creating index 'idx_versions_sort': %s\n\n", idx_versions_sort)
	_, err = db.Exec(idx_versions_sort)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	createTableReleasesQuery := `
	CREATE TABLE IF NOT EXISTS releases (
		branch TEXT NOT NULL,
		assignee_id INTEGER NOT NULL,
		description TEXT,
		commit TEXT,
		date TIMESTAMPTZ NOT NULL,
		issue_tracker_id INTEGER NOT NULL,
		version_id INTEGER NOT NULL,
		bump INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (version_id, bump),
		FOREIGN KEY (version_id) REFERENCES versions(id),
		FOREIGN KEY (assignee_id) REFERENCES assignees(id)
	);`
	fmt.Printf("creating table 'releases': %s\n\n", createTableReleasesQuery)
	_, err = db.Exec(createTableReleasesQuery)
	if err != nil {
		return fmt.Errorf("failed to create releases table: %w", err)
	}

	idx_releases_bump := `CREATE INDEX IF NOT EXISTS idx_releases_bump ON releases(bump);`
	fmt.Printf("creating index 'idx_releases_bump': %s\n\n", idx_releases_bump)
	_, err = db.Exec(idx_releases_bump)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}
