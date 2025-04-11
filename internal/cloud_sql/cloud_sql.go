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
		dbName = "ep-bold-mud-ac6h4c01-pooler.sa-east-1.aws.neon.tech/neondb"
	)

	databaseURL := fmt.Sprintf("postgresql://%s:%s@%s?sslmode=require", dbUser, dbPwd, dbName)
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

	// Create the releases table
	createVersionsTableQuery := `
	CREATE TABLE IF NOT EXISTS versions (
		id SERIAL PRIMARY KEY,
		version TEXT NOT NULL UNIQUE
	);`
	_, err := db.Exec(createVersionsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create versions table: %w", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_versions_version ON versions(version);`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS releases (
		branch TEXT NOT NULL,
		assignee TEXT NOT NULL,
		description TEXT,
		date TIMESTAMPTZ NOT NULL,
		issue_tracker_id INTEGER NOT NULL,
		version_id INTEGER NOT NULL,
		bump INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (version_id, bump),
		FOREIGN KEY (version_id) REFERENCES versions(id)
	);`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create releases table: %w", err)
	}
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_releases_bump ON releases(bump);`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}
