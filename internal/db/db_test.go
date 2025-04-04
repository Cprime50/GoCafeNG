package db

// import (
// 	"context"
// 	"database/sql"
// 	"testing"
// 	"time"

// 	"Go9jaJobs/internal/models"

// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/google/uuid"
// 	_ "github.com/mattn/go-sqlite3"
// 	"github.com/stretchr/testify/assert"
// )

// // setupTestDB creates a new mock database for testing
// func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
// 	db, mock, err := sqlmock.New()
// 	assert.NoError(t, err)
// 	return db, mock
// }

// // createTestJob creates a test job for testing
// func createTestJob() models.Job {
// 	return models.Job{
// 		ID:          uuid.New().String(),
// 		JobID:       "test-job-id",
// 		Title:       "Golang Developer",
// 		Company:     "Test Company",
// 		CompanyURL:  "https://example.com",
// 		CompanyLogo: "https://example.com/logo.png",
// 		Country:     "Nigeria",
// 		State:       "Lagos",
// 		Description: "Test job description",
// 		URL:         "https://example.com/jobs/1",
// 		Salary:      "$80K-$100K",
// 		PostedAt:    time.Now().Add(-24 * time.Hour),
// 		JobType:     "Full-time",
// 		IsRemote:    true,
// 		Source:      "jsearch",
// 		DateGotten:  time.Now(),
// 		ExpDate:     time.Now().Add(30 * 24 * time.Hour),
// 		Location:    "Lagos, Nigeria",
// 		RawData:     "{\"test\": \"data\"}",
// 	}
// }

// // createNonGoJob creates a test job that is not Go-related
// func createNonGoJob() models.Job {
// 	job := createTestJob()
// 	job.Title = "Python Developer"
// 	job.Description = "Experience with Django required"
// 	return job
// }

// func TestInitDB(t *testing.T) {
// 	// Create an in-memory SQLite database for testing
// 	db, err := sql.Open("sqlite3", ":memory:")
// 	assert.NoError(t, err)
// 	defer db.Close()

// 	// Create the jobs table
// 	_, err = db.Exec(`
// 	CREATE TABLE IF NOT EXISTS jobs (
// 		id TEXT PRIMARY KEY,
// 		job_id TEXT NOT NULL,
// 		title TEXT NOT NULL,
// 		company TEXT NOT NULL,
// 		company_url TEXT,
// 		company_logo TEXT,
// 		country TEXT,
// 		state TEXT,
// 		description TEXT,
// 		url TEXT,
// 		salary TEXT,
// 		posted_at TIMESTAMP,
// 		job_type TEXT,
// 		is_remote BOOLEAN,
// 		source TEXT NOT NULL,
// 		employment_type TEXT,
// 		exp_date TIMESTAMP,
// 		date_gotten TIMESTAMP,
// 		location TEXT,
// 		raw_data TEXT,
// 		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// 		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// 	)`)
// 	assert.NoError(t, err)

// 	// Create the job_sync_logs table
// 	_, err = db.Exec(`
// 	CREATE TABLE IF NOT EXISTS job_sync_logs (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		api_name TEXT NOT NULL,
// 		sync_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// 		job_count INTEGER,
// 		status TEXT,
// 		error_message TEXT
// 	)`)
// 	assert.NoError(t, err)

// 	// Test inserting a job
// 	testJob := createTestJob()
// 	_, err = db.Exec(`
// 	INSERT INTO jobs (
// 		id, job_id, title, company, company_url, company_logo,
// 		country, state, description, url, salary, posted_at,
// 		job_type, is_remote, source, employment_type, date_gotten, exp_date, location, raw_data, created_at, updated_at
// 	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
// 		testJob.ID, testJob.JobID, testJob.Title, testJob.Company, testJob.CompanyURL, testJob.CompanyLogo,
// 		testJob.Country, testJob.State, testJob.Description, testJob.URL, testJob.Salary, testJob.PostedAt,
// 		testJob.JobType, testJob.IsRemote, testJob.Source, "", testJob.DateGotten, testJob.ExpDate, testJob.Location, testJob.RawData, time.Now(), time.Now())
// 	assert.NoError(t, err)

// 	// Test inserting a job sync log
// 	_, err = db.Exec(`
// 	INSERT INTO job_sync_logs (api_name, job_count, status, error_message)
// 	VALUES (?, ?, ?, ?)`,
// 		"test-api", 10, "success", "")
// 	assert.NoError(t, err)
// }

// func TestIsDuplicateJob(t *testing.T) {
// 	db, mock := setupTestDB(t)
// 	defer db.Close()

// 	ctx := context.Background()
// 	testJob := createTestJob()

// 	// Test with no jobs in DB (should not be a duplicate)
// 	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM jobs").
// 		WithArgs(testJob.Title, testJob.Company, testJob.PostedAt).
// 		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

// 	isDup, err := IsDuplicateJob(ctx, db, testJob)
// 	assert.NoError(t, err)
// 	assert.False(t, isDup)

// 	// Create a new mock for the next test to avoid expectation conflicts
// 	db2, mock2 := setupTestDB(t)
// 	defer db2.Close()

// 	// Test with a duplicate job
// 	mock2.ExpectQuery("SELECT COUNT\\(\\*\\) FROM jobs").
// 		WithArgs(testJob.Title, testJob.Company, testJob.PostedAt).
// 		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

// 	isDup, err = IsDuplicateJob(ctx, db2, testJob)
// 	assert.NoError(t, err)
// 	assert.True(t, isDup)

// 	// Verify all expectations were met
// 	assert.NoError(t, mock.ExpectationsWereMet())
// 	assert.NoError(t, mock2.ExpectationsWereMet())
// }

// func TestDBLogAPISync(t *testing.T) {
// 	// Create an in-memory SQLite database for testing
// 	db, err := sql.Open("sqlite3", ":memory:")
// 	assert.NoError(t, err)
// 	defer db.Close()

// 	// Create the job_sync_logs table
// 	_, err = db.Exec(`
// 	CREATE TABLE IF NOT EXISTS job_sync_logs (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		api_name TEXT NOT NULL,
// 		sync_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// 		job_count INTEGER,
// 		status TEXT,
// 		error_message TEXT
// 	)`)
// 	assert.NoError(t, err)

// 	// Test with valid parameters
// 	apiName := "TestAPI"
// 	jobCount := 10
// 	status := "Success"
// 	errorMsg := ""

// 	// Call the function
// 	LogAPISync(db, apiName, jobCount, status, errorMsg)

// 	// Verify the log was created
// 	var count int
// 	err = db.QueryRow("SELECT COUNT(*) FROM job_sync_logs WHERE api_name = ?", apiName).Scan(&count)
// 	assert.NoError(t, err)
// 	assert.Equal(t, 1, count)

// 	// Test with an error
// 	apiName = "ErrorAPI"
// 	jobCount = 0
// 	status = "Failed"
// 	errorMsg = "Connection timeout"

// 	// Call the function
// 	LogAPISync(db, apiName, jobCount, status, errorMsg)

// 	// Verify the log was created
// 	err = db.QueryRow("SELECT COUNT(*) FROM job_sync_logs WHERE api_name = ?", apiName).Scan(&count)
// 	assert.NoError(t, err)
// 	assert.Equal(t, 1, count)

// 	// Test with nil database
// 	LogAPISync(nil, apiName, jobCount, status, errorMsg)
// 	// This should just log an error and not panic
// }

// func TestDBSaveJobSyncLog(t *testing.T) {
// 	// Create an in-memory SQLite database for testing
// 	db, err := sql.Open("sqlite3", ":memory:")
// 	assert.NoError(t, err)
// 	defer db.Close()

// 	// Create the job_sync_logs table
// 	_, err = db.Exec(`
// 	CREATE TABLE IF NOT EXISTS job_sync_logs (
// 		id INTEGER PRIMARY KEY AUTOINCREMENT,
// 		timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
// 		message TEXT
// 	)`)
// 	assert.NoError(t, err)

// 	message := "Test sync log message"

// 	// Call the function
// 	err = SaveJobSyncLog(db, message)
// 	assert.NoError(t, err)

// 	// Verify the log was created
// 	var count int
// 	err = db.QueryRow("SELECT COUNT(*) FROM job_sync_logs WHERE message = ?", message).Scan(&count)
// 	assert.NoError(t, err)
// 	assert.Equal(t, 1, count)

// 	// Test with database error - close the database to simulate an error
// 	db.Close()
// 	err = SaveJobSyncLog(db, message)
// 	assert.Error(t, err)
// }

// func TestJobFetchCompanyLogo(t *testing.T) {
// 	t.Skip("Skipping test as it requires API_KEY environment variable")

// 	// This is more of an integration test that would call an external API
// 	// For unit testing, we'll just verify the function doesn't panic with empty inputs
// 	logo := FetchCompanyLogo("", "")
// 	assert.Empty(t, logo)

// 	// Test with invalid URL
// 	logo = FetchCompanyLogo("not-a-url", "test-token")
// 	assert.Empty(t, logo)
// }
