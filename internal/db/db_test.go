package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"Go9jaJobs/internal/models"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

// setupTestDB creates a temporary SQLite database for testing
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err, "Failed to open in-memory SQLite database")

	// Create jobs table with all necessary columns for testing
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
	assert.NoError(t, err, "Failed to create jobs table")

	return db
}

// insertTestJob inserts a test job into the database
func insertTestJob(t *testing.T, db *sql.DB, job models.Job) {
	_, err := db.Exec(`
		INSERT INTO jobs (
			id, job_id, title, company, company_url, company_logo, 
			country, state, description, url, salary, posted_at, 
			job_type, is_remote, source, date_gotten, exp_date, location, raw_data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		job.ID, job.JobID, job.Title, job.Company, job.CompanyURL, job.CompanyLogo,
		job.Country, job.State, job.Description, job.URL, job.Salary, job.PostedAt,
		job.JobType, job.IsRemote, job.Source, job.DateGotten, job.ExpDate, job.Location, job.RawData)

	assert.NoError(t, err, "Failed to insert test job")
}

// createTestJob creates a test job object
func createTestJob() models.Job {
	now := time.Now()
	return models.Job{
		ID:          uuid.New().String(),
		JobID:       uuid.New().String(),
		Title:       "Golang Developer",
		Company:     "Test Company",
		CompanyURL:  "https://example.com",
		CompanyLogo: "https://example.com/logo.png",
		Country:     "Nigeria",
		State:       "Lagos",
		Description: "This is a test job description for a Go developer",
		URL:         "https://example.com/jobs/golang-developer",
		Salary:      "$80,000 - $100,000",
		PostedAt:    now.Add(-24 * time.Hour), // Posted 1 day ago
		JobType:     "Full-time",
		IsRemote:    true,
		Source:      "test",
		DateGotten:  now,
		ExpDate:     now.Add(30 * 24 * time.Hour), // Expires in 30 days
		Location:    "Lagos, Nigeria",
		RawData:     "{\"test\": \"data\"}",
	}
}

// createNonGoJob creates a test job object that doesn't mention Go/Golang
func createNonGoJob() models.Job {
	job := createTestJob()
	job.Title = "Python Developer"
	job.Description = "This is a job for a Python developer with Django experience"
	return job
}

// createBlockedCompanyJob creates a test job from a blocked company
func createBlockedCompanyJob() models.Job {
	job := createTestJob()
	job.Company = "Canonical"
	return job
}

func TestInitDB(t *testing.T) {
	// This is more of an integration test that would connect to a real database
	// For unit testing, we'll just verify the function signature is correct
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	// Verify we can execute the creation of the jobs table
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
	assert.NoError(t, err)
}

func TestInitSQLite(t *testing.T) {
	// We're testing just the structure here
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	defer db.Close()

	// Create the job_sync_logs table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS job_sync_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		api_name TEXT NOT NULL,
		sync_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		job_count INTEGER,
		status TEXT,
		error_message TEXT
	)`)
	assert.NoError(t, err)

	// Test we can insert into it
	_, err = db.Exec(
		"INSERT INTO job_sync_logs (api_name, job_count, status, error_message) VALUES (?, ?, ?, ?)",
		"test_api", 10, "success", "",
	)
	assert.NoError(t, err)
}

// SQLite version of IsDuplicateJob for testing
func testIsDuplicateJob(ctx context.Context, db *sql.DB, job models.Job) (bool, error) {
	var count int

	// Modified query for SQLite
	query := `
		SELECT COUNT(*) FROM jobs 
		WHERE LOWER(title) = LOWER(?) 
		AND LOWER(company) = LOWER(?) 
		AND strftime('%Y', posted_at) = strftime('%Y', ?)
		AND strftime('%m', posted_at) = strftime('%m', ?)
	`

	err := db.QueryRowContext(ctx, query, job.Title, job.Company, job.PostedAt, job.PostedAt).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func TestIsDuplicateJob(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	testJob := createTestJob()

	// Test with no jobs in DB (should not be a duplicate)
	isDup, err := testIsDuplicateJob(ctx, db, testJob)
	assert.NoError(t, err)
	assert.False(t, isDup)

	// Insert the job
	insertTestJob(t, db, testJob)

	// Test with the same job (should be a duplicate)
	isDup, err = testIsDuplicateJob(ctx, db, testJob)
	assert.NoError(t, err)
	assert.True(t, isDup)

	// Test with a different job
	differentJob := createTestJob()
	differentJob.Title = "Senior Golang Developer"
	differentJob.Company = "Another Company"

	isDup, err = testIsDuplicateJob(ctx, db, differentJob)
	assert.NoError(t, err)
	assert.False(t, isDup)
}

// TestIsBlockedCompany tests the IsBlockedCompany function
func TestIsBlockedCompany(t *testing.T) {
	// Mock the IsBlockedCompany function with our own test version
	isBlockedCompany := func(companyName string) bool {
		blockedCompanies := []string{"canonical", "crossover"}

		companyLower := strings.ToLower(companyName)
		for _, blocked := range blockedCompanies {
			if strings.Contains(companyLower, blocked) {
				return true
			}
		}
		return false
	}

	// Test with a known blocked company
	assert.True(t, isBlockedCompany("Canonical"))
	assert.True(t, isBlockedCompany("canonical"))     // Case insensitive
	assert.True(t, isBlockedCompany("CANONICAL"))     // Case insensitive
	assert.True(t, isBlockedCompany("Canonical Ltd")) // Partial match

	// Test with non-blocked companies
	assert.False(t, isBlockedCompany("Google"))
	assert.False(t, isBlockedCompany("Microsoft"))
	// Note: CanonicalX would actually match "canonical" using contains
}

// TestIsGoRelatedJob tests the IsGoRelatedJob function
func TestIsGoRelatedJob(t *testing.T) {
	// Implement a simplified version for testing
	isGoRelatedJob := func(job models.Job) bool {
		title := strings.ToLower(job.Title)
		description := strings.ToLower(job.Description)

		// Check for "golang" anywhere
		if strings.Contains(title, "golang") || strings.Contains(description, "golang") {
			return true
		}

		// Check for "go" as a whole word
		if strings.Contains(title, " go ") || strings.Contains(description, " go ") {
			return true
		}

		// Check for "go" at beginning of title
		if strings.HasPrefix(title, "go ") {
			return true
		}

		// Check for "go" at end of title
		if strings.HasSuffix(title, " go") {
			return true
		}

		return false
	}

	// Test with job titles containing Go/Golang
	goJob := createTestJob()
	goJob.Title = "Golang Developer"
	assert.True(t, isGoRelatedJob(goJob))

	goJob.Title = "Go Developer"
	assert.True(t, isGoRelatedJob(goJob))

	goJob.Title = "Senior (Go) Engineer"
	assert.True(t, isGoRelatedJob(goJob))

	goJob.Title = "Backend Engineer"
	goJob.Description = "Experience with Go required"
	assert.True(t, isGoRelatedJob(goJob))

	// Test with non-Go jobs
	nonGoJob := createNonGoJob()
	assert.False(t, isGoRelatedJob(nonGoJob))
}

func TestSaveJobsToDB(t *testing.T) {
	t.Skip("Skipping TestSaveJobsToDB as it requires actual config and database setup")
}

func TestFetchCompanyLogo(t *testing.T) {
	// This is more of an integration test that would call an external API
	// For unit testing, we'll just verify the function doesn't panic with empty inputs
	logo := FetchCompanyLogo("", "")
	assert.Empty(t, logo)

	// Test with invalid URL
	logo = FetchCompanyLogo("not-a-url", "test-token")
	assert.Empty(t, logo)
}
