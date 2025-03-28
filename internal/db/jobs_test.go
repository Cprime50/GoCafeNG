package db

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestInitScheduleTable(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Set expectations
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS job_schedule_info").
		WillReturnResult(sqlmock.NewResult(0, 0))

	// Call the function
	err = InitScheduleTable(db)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetJobScheduleInfo(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Test data
	apiName := "TestAPI"
	mockTime := time.Now().UTC()
	lastRunTime := mockTime.Add(-24 * time.Hour)
	nextRunTime := mockTime.Add(12 * time.Hour)

	// Set expectations
	rows := sqlmock.NewRows([]string{
		"api_name", "last_run_time", "next_run_time", "interval_hours",
		"status", "last_run_count", "last_error_msg",
	}).AddRow(
		apiName, lastRunTime, nextRunTime, 12,
		"Success", 100, "",
	)

	mock.ExpectQuery("SELECT (.+) FROM job_schedule_info WHERE api_name = ?").
		WithArgs(apiName).
		WillReturnRows(rows)

	// Call the function
	info, err := GetJobScheduleInfo(db, apiName)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, apiName, info.ApiName)
	assert.Equal(t, lastRunTime.Unix(), info.LastRunTime.Unix())
	assert.Equal(t, nextRunTime.Unix(), info.NextRunTime.Unix())
	assert.Equal(t, 12, info.IntervalHours)
	assert.Equal(t, "Success", info.Status)
	assert.Equal(t, 100, info.LastRunCount)
	assert.Equal(t, "", info.LastErrorMsg)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetJobScheduleInfoNotFound(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Test data
	apiName := "NonexistentAPI"

	// Set expectations for no rows
	mock.ExpectQuery("SELECT (.+) FROM job_schedule_info WHERE api_name = ?").
		WithArgs(apiName).
		WillReturnRows(sqlmock.NewRows([]string{}))

	// Call the function
	info, err := GetJobScheduleInfo(db, apiName)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, info)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpsertJobScheduleInfo(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Test data
	now := time.Now().UTC()
	info := JobScheduleInfo{
		ApiName:       "TestAPI",
		LastRunTime:   now.Add(-24 * time.Hour),
		NextRunTime:   now.Add(12 * time.Hour),
		IntervalHours: 12,
		Status:        "Success",
		LastRunCount:  100,
		LastErrorMsg:  "",
	}

	// Set expectations
	mock.ExpectExec("INSERT INTO job_schedule_info").
		WithArgs(
			info.ApiName, info.LastRunTime, info.NextRunTime, info.IntervalHours,
			info.Status, info.LastRunCount, info.LastErrorMsg,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the function
	err = UpdatesJobScheduleInfo(db, info)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllJobScheduleInfo(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Test data
	mockTime := time.Now().UTC()
	api1LastRun := mockTime.Add(-24 * time.Hour)
	api1NextRun := mockTime.Add(12 * time.Hour)
	api2LastRun := mockTime.Add(-12 * time.Hour)
	api2NextRun := mockTime.Add(24 * time.Hour)

	// Set expectations
	rows := sqlmock.NewRows([]string{
		"api_name", "last_run_time", "next_run_time", "interval_hours",
		"status", "last_run_count", "last_error_msg",
	}).
		AddRow("API1", api1LastRun, api1NextRun, 12, "Success", 100, "").
		AddRow("API2", api2LastRun, api2NextRun, 24, "Failed", 0, "Error message")

	mock.ExpectQuery("SELECT (.+) FROM job_schedule_info").
		WillReturnRows(rows)

	// Call the function
	infos, err := GetAllJobScheduleInfo(db)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, infos, 2)

	// Check first API
	assert.Equal(t, "API1", infos[0].ApiName)
	assert.Equal(t, api1LastRun.Unix(), infos[0].LastRunTime.Unix())
	assert.Equal(t, api1NextRun.Unix(), infos[0].NextRunTime.Unix())
	assert.Equal(t, 12, infos[0].IntervalHours)
	assert.Equal(t, "Success", infos[0].Status)
	assert.Equal(t, 100, infos[0].LastRunCount)
	assert.Equal(t, "", infos[0].LastErrorMsg)

	// Check second API
	assert.Equal(t, "API2", infos[1].ApiName)
	assert.Equal(t, api2LastRun.Unix(), infos[1].LastRunTime.Unix())
	assert.Equal(t, api2NextRun.Unix(), infos[1].NextRunTime.Unix())
	assert.Equal(t, 24, infos[1].IntervalHours)
	assert.Equal(t, "Failed", infos[1].Status)
	assert.Equal(t, 0, infos[1].LastRunCount)
	assert.Equal(t, "Error message", infos[1].LastErrorMsg)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLogJobRun(t *testing.T) {
	// Create a mock database connection
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer db.Close()

	// Test data
	apiName := "TestAPI"
	status := "Success"
	jobCount := 100
	errorMsg := ""
	intervalHours := 12

	// Set expectations for UpsertJobScheduleInfo (updated by LogJobRun)
	mock.ExpectExec("INSERT INTO job_schedule_info").
		WithArgs(
			apiName, sqlmock.AnyArg(), sqlmock.AnyArg(), intervalHours,
			status, jobCount, errorMsg,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the function
	LogJobRun(db, apiName, status, jobCount, errorMsg, intervalHours)

	// Assert
	assert.NoError(t, mock.ExpectationsWereMet())
}
