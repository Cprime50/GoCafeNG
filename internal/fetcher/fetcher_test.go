package fetcher

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"Go9jaJobs/internal/config"
	"Go9jaJobs/internal/models"

	"github.com/stretchr/testify/assert"
)

// setupTestServer creates a test HTTP server that returns mocked API responses
func setupTestServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()

	// Register all the handlers
	for pattern, handler := range handlers {
		mux.HandleFunc(pattern, handler)
	}

	// Create a test server
	server := httptest.NewServer(mux)

	// Create cache directory if it doesn't exist
	os.MkdirAll("api_response_cache", 0755)

	return server
}

// createMockConfig creates a config with the test server URL
func createMockConfig(serverURL string) *config.Config {
	return &config.Config{
		RapidAPIKey:      "test-rapid-api-key",
		ApifyAPIKey:      "test-apify-api-key",
		BrandFetchAPIKey: "test-logo-api-token",
		Mode:             "dev",
	}
}

// mockJSearchResponse returns a handler that serves a mock JSearch API response
func mockJSearchResponse(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check query params
		query := r.URL.Query().Get("query")
		assert.Equal(t, "golang jobs in nigeria", query)

		// Return mocked response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.JSEARCHResponse{
			Data: []struct {
				ID             string    `json:"job_id"`
				JobTitle       string    `json:"job_title"`
				EmployerName   string    `json:"employer_name"`
				CompanyURL     string    `json:"employer_website"`
				EmployerLogo   string    `json:"employer_logo"`
				JobLocation    string    `json:"job_location"`
				JobDescription string    `json:"job_description"`
				JobApplyLink   string    `json:"job_apply_link"`
				JobSalary      string    `json:"job_salary"`
				JobPostedAt    time.Time `json:"job_posted_at_datetime_utc"`
				JobType        string    `json:"job_employment_type"`
				JobIsRemote    bool      `json:"job_is_remote"`
			}{
				{
					ID:             "job123",
					JobTitle:       "Golang Developer",
					EmployerName:   "Test Company",
					CompanyURL:     "https://testcompany.com",
					EmployerLogo:   "https://testcompany.com/logo.png",
					JobLocation:    "Lagos, Nigeria",
					JobDescription: "We need a Go developer to work on exciting projects",
					JobApplyLink:   "https://testcompany.com/apply",
					JobSalary:      "$50K-$70K",
					JobPostedAt:    time.Now().Add(-24 * time.Hour),
					JobType:        "Full-time",
					JobIsRemote:    true,
				},
			},
		})
	}
}

// mockLinkedInResponse returns a handler that serves a mock LinkedIn API response
func mockLinkedInResponse(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		assert.Equal(t, "linkedin-job-search-api.p.rapidapi.com", r.Header.Get("x-rapidapi-host"))
		assert.Equal(t, "test-rapid-api-key", r.Header.Get("x-rapidapi-key"))

		// Return mocked response
		w.Header().Set("Content-Type", "application/json")
		respData := []map[string]interface{}{
			{
				"id":                "linkedin-job-1",
				"title":             "Senior Golang Developer",
				"organization":      "LinkedIn Company",
				"organization_url":  "https://linkedin-company.com",
				"organization_logo": "https://linkedin-company.com/logo.png",
				"date_posted":       time.Now().Format(time.RFC3339),
				"locations_derived": []string{"Lagos, Nigeria"},
				"url":               "https://linkedin.com/jobs/1",
				"description":       "Looking for a Go expert to join our team",
			},
		}
		json.NewEncoder(w).Encode(respData)
	}
}

// mockIndeedResponse returns a handler that serves a mock Indeed API response
func mockIndeedResponse(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check method and content type
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Check request body
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		assert.NoError(t, err)

		assert.Equal(t, "NG", payload["country"])
		assert.Equal(t, "golang", payload["position"])

		// Return mocked response
		w.Header().Set("Content-Type", "application/json")

		// Create a single Indeed job response
		indeedJob := map[string]interface{}{
			"salary":            "$60K-$80K",
			"postedAt":          "2 days ago",
			"positionName":      "Golang Backend Engineer",
			"jobType":           []string{"Full-time"},
			"company":           "Indeed Company",
			"location":          "Lagos, Nigeria",
			"url":               "https://indeed.com/job/123",
			"id":                "indeed-job-123",
			"scrapedAt":         time.Now().Format(time.RFC3339),
			"postingDateParsed": time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			"description":       "We are looking for a Go developer with 3+ years of experience",
			"searchInput": map[string]string{
				"position": "golang",
				"country":  "NG",
			},
			"isExpired": false,
			"companyInfo": map[string]interface{}{
				"indeedUrl":   "https://indeed.com/company/indeed-company",
				"companyLogo": "https://indeed-company.com/logo.png",
			},
		}

		// Put it in an array and encode
		indeedResp := []map[string]interface{}{indeedJob}
		json.NewEncoder(w).Encode(indeedResp)
	}
}

// mockApifyLinkedInResponse returns a handler that serves a mock Apify LinkedIn API response
func mockApifyLinkedInResponse(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check method and content type
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Check request body
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		defer r.Body.Close()

		var payload map[string]interface{}
		err = json.Unmarshal(body, &payload)
		assert.NoError(t, err)

		urls, ok := payload["urls"].([]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, urls)
		assert.True(t, strings.Contains(urls[0].(string), "golang"))

		// Return mocked response
		w.Header().Set("Content-Type", "application/json")

		// Create sample response data as a map
		apifyJob := map[string]interface{}{
			"id":                 "apify-linkedin-job-1",
			"trackingId":         "tracking-1",
			"refId":              "ref-1",
			"link":               "https://linkedin.com/jobs/1234",
			"title":              "Golang Developer at Apify",
			"companyName":        "Apify Company",
			"companyLinkedinUrl": "https://linkedin.com/company/apify",
			"companyLogo":        "https://apify.com/logo.png",
			"location":           "Lagos, Nigeria",
			"salaryInfo":         []string{"$70K-$90K"},
			"postedAt":           time.Now().Format("2006-01-02"),
			"descriptionHtml":    "<p>We need Go developers</p>",
			"applicantsCount":    "25",
			"applyUrl":           "https://linkedin.com/jobs/1234/apply",
			"descriptionText":    "We need Go developers with experience in microservices",
			"employmentType":     "Full-time",
		}

		// Put it in an array and encode
		apifyResp := []map[string]interface{}{apifyJob}

		// Encode and send response
		json.NewEncoder(w).Encode(apifyResp)
	}
}

// Define a mock transport to redirect requests to our test server
type mockTransport struct {
	URL    string
	Client *http.Client
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Replace the request URL with our test server URL, but keep the query and headers
	req.URL, _ = url.Parse(m.URL + "?" + req.URL.RawQuery)
	return m.Client.Transport.RoundTrip(req)
}

func TestNewJobFetcher(t *testing.T) {
	// Test creating a new job fetcher
	cfg := createMockConfig("")
	fetcher := NewJobFetcher(cfg)

	assert.NotNil(t, fetcher)
	assert.NotNil(t, fetcher.client)
	assert.Equal(t, cfg, fetcher.Config)

	// Check that client timeout is set appropriately
	assert.Equal(t, 180*time.Second, fetcher.client.Timeout)

	// Check that cache directory was created
	_, err := os.Stat("api_response_cache")
	assert.NoError(t, err)
}

func TestFetchJSearchJobs(t *testing.T) {
	// Setup test server
	handlers := map[string]http.HandlerFunc{
		"/jsearch": mockJSearchResponse(t),
	}
	server := setupTestServer(t, handlers)
	defer server.Close()

	// Create config and fetcher
	cfg := createMockConfig(server.URL)
	fetcher := NewJobFetcher(cfg)

	// Override client to use test server
	fetcher.client = server.Client()

	// Set the server URL for the test
	apiURL := server.URL + "/jsearch"

	// Test fetching jobs
	ctx := context.Background()

	// Create a custom request directly to the test server
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	assert.NoError(t, err)

	q := req.URL.Query()
	q.Add("query", "golang jobs in nigeria")
	q.Add("page", "1")
	q.Add("num_pages", "3")
	q.Add("country", "ng")
	req.URL.RawQuery = q.Encode()

	req.Header.Add("x-rapidapi-host", "jsearch.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", fetcher.Config.RapidAPIKey)

	// Save the original client for restoration
	originalClient := fetcher.client

	// Override client to use our custom test client
	fetcher.client = &http.Client{
		Transport: &mockTransport{URL: apiURL, Client: server.Client()},
	}

	// Test fetching jobs
	jobs, err := fetcher.FetchJSearchJobs(ctx)

	// Restore the original client
	fetcher.client = originalClient

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)

	job := jobs[0]
	assert.Equal(t, "Golang Developer", job.Title)
	assert.Equal(t, "Test Company", job.Company)
	assert.Equal(t, "https://testcompany.com/logo.png", job.CompanyLogo)
	assert.Equal(t, "Lagos, Nigeria", job.Location)
	assert.Equal(t, "jsearch", job.Source)
	assert.True(t, job.IsRemote)

	// Check that cache file was created
	cachePath := filepath.Join("api_response_cache", "jsearch_response.json")
	_, err = os.Stat(cachePath)
	assert.NoError(t, err)

	// Clean up
	os.Remove(cachePath)
}

func TestFetchLinkedInJobs(t *testing.T) {
	// Setup test server
	handlers := map[string]http.HandlerFunc{
		"/linkedin/active-jb-7d": mockLinkedInResponse(t),
	}
	server := setupTestServer(t, handlers)
	defer server.Close()

	// Create config and fetcher
	cfg := createMockConfig(server.URL)
	fetcher := NewJobFetcher(cfg)

	// Override client to use test server
	fetcher.client = server.Client()

	// Set dev mode to use test server URL
	fetcher.Config.Mode = "dev"

	// Create a custom client that redirects to our test server
	testClient := &http.Client{
		Transport: &mockTransport{URL: server.URL + "/linkedin/active-jb-7d", Client: server.Client()},
	}

	// Save original client and replace with test client
	originalClient := fetcher.client
	fetcher.client = testClient

	// Test fetching jobs
	ctx := context.Background()
	jobs, err := fetcher.FetchLinkedInJobs(ctx)

	// Restore original client
	fetcher.client = originalClient

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)

	job := jobs[0]
	assert.Equal(t, "Senior Golang Developer", job.Title)
	assert.Equal(t, "LinkedIn Company", job.Company)
	assert.Equal(t, "https://linkedin-company.com/logo.png", job.CompanyLogo)
	assert.Equal(t, "Lagos, Nigeria", job.Location)
	assert.Equal(t, "linkedin", job.Source)

	// Check that cache file was created
	cachePath := filepath.Join("api_response_cache", "linkedin_response.json")
	_, err = os.Stat(cachePath)
	assert.NoError(t, err)

	// Clean up
	os.Remove(cachePath)
}

func TestContainsAny(t *testing.T) {
	// Test function should find substrings
	assert.True(t, containsAny("this is a test with remote work", []string{"remote"}))
	assert.True(t, containsAny("REMOTE WORK OPPORTUNITY", []string{"remote"}))
	assert.True(t, containsAny("work from home opportunity", []string{"work from home"}))
	assert.True(t, containsAny("This is a WFH position", []string{"wfh"}))

	// Test function should not find non-existing substrings
	assert.False(t, containsAny("on-site work only", []string{"remote", "work from home", "wfh"}))
	assert.False(t, containsAny("", []string{"remote"}))
	assert.False(t, containsAny("remotework", []string{"remote work"})) // Not matching the exact substring
}

func TestFetchIndeedJobs(t *testing.T) {
	// Setup test server
	handlers := map[string]http.HandlerFunc{
		"/apify/indeed/run-sync-get-dataset-items": mockIndeedResponse(t),
	}
	server := setupTestServer(t, handlers)
	defer server.Close()

	// Create config and fetcher
	cfg := createMockConfig(server.URL)
	fetcher := NewJobFetcher(cfg)

	// Override client to use test server
	fetcher.client = server.Client()

	// Set dev mode to use test server URL
	fetcher.Config.Mode = "dev"

	// Create a custom client that redirects to our test server
	testClient := &http.Client{
		Transport: &mockTransport{URL: server.URL + "/apify/indeed/run-sync-get-dataset-items", Client: server.Client()},
	}

	// Save original client and replace with test client
	originalClient := fetcher.client
	fetcher.client = testClient

	// Test fetching jobs
	ctx := context.Background()
	jobs, err := fetcher.FetchIndeedJobs(ctx)

	// Restore original client
	fetcher.client = originalClient

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)

	job := jobs[0]
	assert.Equal(t, "Golang Backend Engineer", job.Title)
	assert.Equal(t, "Indeed Company", job.Company)
	assert.Equal(t, "https://indeed-company.com/logo.png", job.CompanyLogo)
	assert.Equal(t, "Lagos, Nigeria", job.Location)
	assert.Equal(t, "$60K-$80K", job.Salary)
	assert.Equal(t, "Full-time", job.JobType)
	assert.Equal(t, "apify indeed", job.Source)

	// Check that cache file was created
	cachePath := filepath.Join("api_response_cache", "indeed_response.json")
	_, err = os.Stat(cachePath)
	assert.NoError(t, err)

	// Clean up
	os.Remove(cachePath)
}

func TestFetchApifyLinkedInJobs(t *testing.T) {
	// Setup test server
	handlers := map[string]http.HandlerFunc{
		"/apify/linkedin/run-sync-get-dataset-items": mockApifyLinkedInResponse(t),
	}
	server := setupTestServer(t, handlers)
	defer server.Close()

	// Create config and fetcher
	cfg := createMockConfig(server.URL)
	fetcher := NewJobFetcher(cfg)

	// Override client to use test server
	fetcher.client = server.Client()

	// Set dev mode to use test server URL
	fetcher.Config.Mode = "dev"

	// Create a custom client that redirects to our test server
	testClient := &http.Client{
		Transport: &mockTransport{URL: server.URL + "/apify/linkedin/run-sync-get-dataset-items", Client: server.Client()},
	}

	// Save original client and replace with test client
	originalClient := fetcher.client
	fetcher.client = testClient

	// Test fetching jobs
	ctx := context.Background()
	jobs, err := fetcher.FetchApifyLinkedInJobs(ctx)

	// Restore original client
	fetcher.client = originalClient

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)

	job := jobs[0]
	assert.Equal(t, "Golang Developer at Apify", job.Title)
	assert.Equal(t, "Apify Company", job.Company)
	assert.Equal(t, "https://apify.com/logo.png", job.CompanyLogo)
	assert.Equal(t, "Lagos, Nigeria", job.Location)
	assert.Equal(t, "Full-time", job.JobType)
	assert.Equal(t, "apify linkedin", job.Source)
	assert.Contains(t, job.Description, "Go developers with experience")

	// Check that cache file was created
	cachePath := filepath.Join("api_response_cache", "apify_linkedin_response.json")
	_, err = os.Stat(cachePath)
	assert.NoError(t, err)

	// Clean up
	os.Remove(cachePath)
}
