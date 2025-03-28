package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// setupMockDB sets up a mock database for testing
func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err, "Failed to create mock database")
	return db, mock
}

func TestStatusCheck(t *testing.T) {
	// Create a new request
	req, err := http.NewRequest("GET", "/api/status", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a handler with any database (not used in this test)
	db, _ := setupMockDB(t)
	defer db.Close()

	handler := NewHandler(db)

	// Call the handler function directly
	handler.StatusCheck(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check the response fields
	assert.Equal(t, "ok", response["status"])
	assert.NotEmpty(t, response["timestamp"])
	assert.Equal(t, "API is running", response["message"])
}

func TestGetAllJobs(t *testing.T) {
	// Create a new request
	req, err := http.NewRequest("GET", "/api/jobs", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Setup mock DB and expectations
	db, mock := setupMockDB(t)
	defer db.Close()

	// Define the columns that will be returned
	columns := []string{
		"id", "job_id", "title", "company", "company_url", "company_logo",
		"location", "description", "url", "salary", "posted_at",
		"job_type", "is_remote", "source",
	}

	// Setup mock query expectations
	rows := sqlmock.NewRows(columns).
		AddRow(
			"job-uuid-1", "job-id-1", "Golang Developer", "Company A",
			"https://companya.com", "https://companya.com/logo.png",
			"Lagos, Nigeria", "Description for job 1", "https://companya.com/jobs/1",
			"$80K-$100K", time.Now(), "Full-time", true, "indeed",
		).
		AddRow(
			"job-uuid-2", "job-id-2", "Senior Go Engineer", "Company B",
			"https://companyb.com", "https://companyb.com/logo.png",
			"Remote", "Description for job 2", "https://companyb.com/jobs/2",
			"$100K-$120K", time.Now(), "Contract", true, "linkedin",
		)

	mock.ExpectQuery("^SELECT (.+) FROM jobs ORDER BY posted_at DESC$").WillReturnRows(rows)

	// Create handler and call the function
	handler := NewHandler(db)
	handler.GetAllJobs(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Parse the response
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check the response structure
	assert.Equal(t, true, response["success"])
	assert.Equal(t, float64(2), response["count"]) // JSON unmarshals numbers as float64

	// Check the data
	data, ok := response["data"].([]interface{})
	assert.True(t, ok, "data should be an array")
	assert.Equal(t, 2, len(data))

	// Verify first job data
	job1, ok := data[0].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Golang Developer", job1["title"])
	assert.Equal(t, "Company A", job1["company"])
	assert.Equal(t, "https://companya.com/logo.png", job1["company_logo"])

	// Verify second job data
	job2, ok := data[1].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Senior Go Engineer", job2["title"])
	assert.Equal(t, "Company B", job2["company"])
	assert.Equal(t, "https://companyb.com/logo.png", job2["company_logo"])

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAllJobsDBError(t *testing.T) {
	// Create a new request
	req, err := http.NewRequest("GET", "/api/jobs", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Setup mock DB and expectations
	db, mock := setupMockDB(t)
	defer db.Close()

	// Setup mock query to return an error
	mock.ExpectQuery("^SELECT (.+) FROM jobs ORDER BY posted_at DESC$").
		WillReturnError(sql.ErrConnDone)

	// Create handler and call the function
	handler := NewHandler(db)
	handler.GetAllJobs(rr, req)

	// Check the status code should be 500 for internal server error
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Verify that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSetupRoutes(t *testing.T) {
	// Create a mock DB and handler
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock: %v", err)
	}
	defer mockDB.Close()

	handler := NewHandler(mockDB)
	router := handler.SetupRoutes()

	// Define the routes we expect to exist
	expectedRoutes := []struct {
		path   string
		method string
	}{
		{"/api/status", "GET"},
		{"/api/jobs", "GET"},
	}

	// Create a test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Test each route
	for _, route := range expectedRoutes {
		t.Run(fmt.Sprintf("%s %s", route.method, route.path), func(t *testing.T) {
			// Create a request
			req, err := http.NewRequest(route.method, server.URL+route.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Send the request
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			// We don't care about the actual response for this test, just that the route exists
			assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}
