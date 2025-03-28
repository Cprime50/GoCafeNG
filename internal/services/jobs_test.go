package services

import (
	"testing"
	"time"

	"Go9jaJobs/internal/config"
	"Go9jaJobs/internal/db"
	"Go9jaJobs/internal/fetcher"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// Mock job fetcher for testing


func TestStartJobScheduler(t *testing.T) {
	// Create a mock database connection
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	// Test data - job schedule info for APIs
	now := time.Now().UTC()
	jsearchInfo := db.JobScheduleInfo{
		ApiName:       "JSearch",
		LastRunTime:   now.Add(-6 * time.Hour),
		NextRunTime:   now.Add(6 * time.Hour),
		IntervalHours: 12,
		Status:        "Success",
		LastRunCount:  100,
	}

	indeedInfo := db.JobScheduleInfo{
		ApiName:       "Indeed",
		LastRunTime:   now.Add(-12 * time.Hour),
		NextRunTime:   now.Add(12 * time.Hour),
		IntervalHours: 24,
		Status:        "Success",
		LastRunCount:  200,
	}

	// Set expectations for GetAllJobScheduleInfo
	rows := sqlmock.NewRows([]string{
		"api_name", "last_run_time", "next_run_time", "interval_hours",
		"status", "last_run_count", "last_error_msg",
	}).
		AddRow(jsearchInfo.ApiName, jsearchInfo.LastRunTime, jsearchInfo.NextRunTime,
			jsearchInfo.IntervalHours, jsearchInfo.Status, jsearchInfo.LastRunCount, "").
		AddRow(indeedInfo.ApiName, indeedInfo.LastRunTime, indeedInfo.NextRunTime,
			indeedInfo.IntervalHours, indeedInfo.Status, indeedInfo.LastRunCount, "")

	mock.ExpectQuery("SELECT (.+) FROM job_schedule_info").
		WillReturnRows(rows)

	// Create a mock job fetcher
	cfg := &config.Config{}
	jobFetcher := fetcher.NewJobFetcher(cfg)

	// Call the function
	scheduler := StartJobScheduler(mockDB, jobFetcher)
	defer scheduler.Stop()

	// Assert
	assert.NotNil(t, scheduler)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestStopJobScheduler(t *testing.T) {
	// Create a mock database connection
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	// Create a job scheduler
	cfg := &config.Config{}
	jobFetcher := fetcher.NewJobFetcher(cfg)

	// Mock empty job schedule info - this simplifies the test
	mock.ExpectQuery("SELECT (.+) FROM job_schedule_info").
		WillReturnRows(sqlmock.NewRows([]string{}))

	// Create a scheduler with a simple job
	scheduler := StartJobScheduler(mockDB, jobFetcher)

	// Rest the expectations - this is a fresh slate for StopJobScheduler
	mock.ExpectationsWereMet()

	// Since we won't have any jobs in our mock scheduler (due to empty data above),
	// the StopJobScheduler function won't try to query and update the database

	// Call the function
	StopJobScheduler(scheduler, mockDB)

	// No assertions needed for mock expectations since we're not expecting any database calls
}

func TestFetchAndSaveJobs(t *testing.T) {
	// Skip this test because it requires environment variables and API configuration
	t.Skip("Skipping TestFetchAndSaveJobs because it requires actual API configuration")
}

// TestSchedulerPersistenceAcrossRestart verifies that the job scheduler
// correctly restores its state from the database after a server restart
func TestSchedulerPersistenceAcrossRestart(t *testing.T) {
	// Create a mock database connection
	mockDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	// Create a mock job fetcher
	cfg := &config.Config{}
	jobFetcher := fetcher.NewJobFetcher(cfg)

	// Set up the current time
	now := time.Now().UTC()

	// Create custom job schedule info for testing
	jsearchInfo := db.JobScheduleInfo{
		ApiName:       "JSearch",
		LastRunTime:   now.Add(-6 * time.Hour), // Last ran 6 hours ago
		NextRunTime:   now.Add(6 * time.Hour),  // Next run in 6 hours
		IntervalHours: 12,
		Status:        "Success",
		LastRunCount:  45,
	}

	// Set up the mock to return our test data
	mock.ExpectQuery("SELECT api_name, last_run_time, next_run_time, interval_hours, status, last_run_count, last_error_msg FROM job_schedule_info ORDER BY api_name").
		WillReturnRows(sqlmock.NewRows([]string{
			"api_name", "last_run_time", "next_run_time", "interval_hours",
			"status", "last_run_count", "last_error_msg",
		}).AddRow(
			jsearchInfo.ApiName, jsearchInfo.LastRunTime, jsearchInfo.NextRunTime,
			jsearchInfo.IntervalHours, jsearchInfo.Status, jsearchInfo.LastRunCount, "",
		))

	// Allow any other database interactions
	mock.MatchExpectationsInOrder(false)

	// These allow for any number of these SQL operations
	for i := 0; i < 10; i++ {
		mock.ExpectExec("INSERT INTO job_schedule_info").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectQuery("SELECT (.+) FROM job_schedule_info WHERE api_name").WillReturnRows(sqlmock.NewRows([]string{}))
	}

	// Start the scheduler
	scheduler := StartJobScheduler(mockDB, jobFetcher)
	defer scheduler.Stop()

	// Let it run briefly to initialize
	time.Sleep(200 * time.Millisecond)

	// Get all jobs
	jobs := scheduler.Jobs()

	// Find a job that matches our JSearch schedule
	var foundMatchingJob bool
	for _, job := range jobs {
		timeDiff := job.NextRun().Sub(jsearchInfo.NextRunTime).Abs()
		if timeDiff < 2*time.Second {
			// Found a job with a schedule matching our JSearch job
			foundMatchingJob = true
			break
		}
	}

	// Verify we found a matching job
	assert.True(t, foundMatchingJob, "Should have found a job with schedule matching JSearch from database")

	// --------------------------------------------------
	// SIMULATE SERVER RESTART
	// --------------------------------------------------

	// Stop the scheduler
	StopJobScheduler(scheduler, mockDB)

	// Create a new mock database to simulate a server restart
	restartDB, restartMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("Failed to create restart mock: %v", err)
	}
	defer restartDB.Close()

	// Create updated data that simulates what would be in the database after shutdown
	updatedJsearchInfo := db.JobScheduleInfo{
		ApiName:       "JSearch",
		LastRunTime:   now,                     // Just ran
		NextRunTime:   now.Add(12 * time.Hour), // Next run in 12 hours
		IntervalHours: 12,
		Status:        "Success",
		LastRunCount:  50,
	}

	// Set up the mock to return our updated test data on restart
	restartMock.ExpectQuery("SELECT api_name, last_run_time, next_run_time, interval_hours, status, last_run_count, last_error_msg FROM job_schedule_info ORDER BY api_name").
		WillReturnRows(sqlmock.NewRows([]string{
			"api_name", "last_run_time", "next_run_time", "interval_hours",
			"status", "last_run_count", "last_error_msg",
		}).AddRow(
			updatedJsearchInfo.ApiName, updatedJsearchInfo.LastRunTime, updatedJsearchInfo.NextRunTime,
			updatedJsearchInfo.IntervalHours, updatedJsearchInfo.Status, updatedJsearchInfo.LastRunCount, "",
		))

	// Allow any other database interactions
	restartMock.MatchExpectationsInOrder(false)

	// These allow for any number of these SQL operations
	for i := 0; i < 10; i++ {
		restartMock.ExpectExec("INSERT INTO job_schedule_info").WillReturnResult(sqlmock.NewResult(1, 1))
		restartMock.ExpectQuery("SELECT (.+) FROM job_schedule_info WHERE api_name").WillReturnRows(sqlmock.NewRows([]string{}))
	}

	// Restart the scheduler
	restartScheduler := StartJobScheduler(restartDB, jobFetcher)
	defer restartScheduler.Stop()

	// Let it run briefly to initialize
	time.Sleep(200 * time.Millisecond)

	// Get all jobs after restart
	restartJobs := restartScheduler.Jobs()

	// Find a job that matches our updated JSearch schedule
	foundMatchingJob = false
	for _, job := range restartJobs {
		timeDiff := job.NextRun().Sub(updatedJsearchInfo.NextRunTime).Abs()
		if timeDiff < 2*time.Second {
			// Found a job with a schedule matching our updated JSearch job
			foundMatchingJob = true
			break
		}
	}

	// Verify we found a matching job with the updated schedule
	assert.True(t, foundMatchingJob, "Should have found a job with schedule matching updated JSearch from database after restart")
}
