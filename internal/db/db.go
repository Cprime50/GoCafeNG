package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes the PostgreSQL database connection
func InitDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Create jobs table if it doesn't exist
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		job_id TEXT NOT NULL,
		title TEXT NOT NULL,
		company TEXT NOT NULL,
		company_url TEXT,
		country TEXT,
		state TEXT,
		description TEXT,
		url TEXT,
		salary TEXT,
		posted_at TIMESTAMP,
		job_type TEXT,
		is_remote BOOLEAN,
		source TEXT NOT NULL,
		employment_type TEXT,
		exp_date TIMESTAMP,
		date_gotten TIMESTAMP,
		location TEXT,
		raw_data TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)

	if err != nil {
		return nil, err
	}

	return db, nil
}

// TODO chnage this to be logged in excel instead
// InitSQLite initializes an in-memory SQLite database for logging
func InitSQLite() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create job_sync_logs table in memory
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS job_sync_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		api_name TEXT NOT NULL,
		sync_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		job_count INTEGER,
		status TEXT,
		error_message TEXT
	)`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// LogAPISync logs API sync attempts in SQLite
func LogAPISync(sqliteDB *sql.DB, apiName string, jobCount int, status string, errorMsg string) {
	_, err := sqliteDB.Exec(
		"INSERT INTO job_sync_logs (api_name, job_count, status, error_message) VALUES (?, ?, ?, ?)",
		apiName, jobCount, status, errorMsg,
	)
	if err != nil {
		log.Println("Error logging API sync:", err)
	}
}







