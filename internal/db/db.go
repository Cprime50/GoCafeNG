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

	// Initialize the job scheduler persistence table
	err = InitScheduleTable(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// TODO chnage this to be logged in excel instead
// InitSQLite initializes a file-based SQLite database for logging
func InitSQLite() (*sql.DB, error) {
	// Use a file-based database instead of in-memory to ensure persistence across connections
	db, err := sql.Open("sqlite3", "./job_logs.db")
	if err != nil {
		return nil, err
	}

	// Ensure connection is valid
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Create job_sync_logs table in the file database
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
		log.Printf("Error creating table job_sync_logs: %v", err)
		return nil, err
	}

	// Verify table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='job_sync_logs'").Scan(&tableName)
	if err != nil {
		log.Printf("Table verification failed: %v", err)
		return nil, err
	}

	log.Println("SQLite initialized successfully with job_sync_logs table")
	return db, nil
}

// LogAPISync logs API sync attempts in SQLite
func LogAPISync(sqliteDB *sql.DB, apiName string, jobCount int, status string, errorMsg string) {
	if sqliteDB == nil {
		log.Println("Error logging API sync: SQLite database connection is nil")
		return
	}

	// Verify the table exists before insertion
	var tableName string
	err := sqliteDB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='job_sync_logs'").Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("Error logging API sync: job_sync_logs table does not exist")
			// Try to recreate the table
			_, err = sqliteDB.Exec(`
			CREATE TABLE IF NOT EXISTS job_sync_logs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				api_name TEXT NOT NULL,
				sync_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				job_count INTEGER,
				status TEXT,
				error_message TEXT
			)`)
			if err != nil {
				log.Println("Error recreating job_sync_logs table:", err)
				return
			}
			log.Println("Recreated job_sync_logs table successfully")
		} else {
			log.Println("Error verifying job_sync_logs table:", err)
			return
		}
	}

	// Insert the log
	_, err = sqliteDB.Exec(
		"INSERT INTO job_sync_logs (api_name, job_count, status, error_message) VALUES (?, ?, ?, ?)",
		apiName, jobCount, status, errorMsg,
	)
	if err != nil {
		log.Println("Error inserting into job_sync_logs:", err)
	} else {
		log.Printf("Successfully logged API sync for %s", apiName)
	}
}
