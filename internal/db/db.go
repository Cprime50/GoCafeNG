package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
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
		company_logo TEXT,
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

	// Create job_sync_logs table if it doesn't exist
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS job_sync_logs (
		id SERIAL PRIMARY KEY,
		api_name TEXT NOT NULL,
		sync_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		job_count INTEGER,
		status TEXT,
		error_message TEXT
	)`)

	if err != nil {
		log.Printf("Error creating table job_sync_logs: %v", err)
		return nil, err
	}

	return db, nil
}

// LogAPISync logs API sync attempts in PostgreSQL
func LogAPISync(db *sql.DB, apiName string, jobCount int, status string, errorMsg string) {
	if db == nil {
		log.Println("Error logging API sync: PostgreSQL database connection is nil")
		return
	}

	// Insert the log
	_, err := db.Exec(
		"INSERT INTO job_sync_logs (api_name, job_count, status, error_message) VALUES ($1, $2, $3, $4)",
		apiName, jobCount, status, errorMsg,
	)
	if err != nil {
		log.Println("Error inserting into job_sync_logs:", err)
	} else {
		log.Printf("Successfully logged API sync for %s", apiName)
	}
}

// JobSyncLog represents a log entry for job synchronization
type JobSyncLog struct {
	ID        int       `db:"id"`
	Timestamp time.Time `db:"timestamp"`
	Message   string    `db:"message"`
}

// SaveJobSyncLog saves a job synchronization log entry to the database
func SaveJobSyncLog(db *sql.DB, message string) error {
	query := `INSERT INTO job_sync_logs (timestamp, message) VALUES ($1, $2)`
	_, err := db.Exec(query, time.Now(), message)
	if err != nil {
		log.Printf("Failed to save job sync log: %v", err)
		return err
	}
	return nil
}
