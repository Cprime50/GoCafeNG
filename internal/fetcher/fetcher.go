package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Go9jaJobs/internal/config"
	"Go9jaJobs/internal/models"

	"github.com/google/uuid"
)

// cacheResponse saves API responses to cache files for debugging
func cacheResponse(filename string, data []byte) {
	// Create a directory for cache files if it doesn't exist
	cacheDir := "api_response_cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return // Silently fail if we can't create the cache directory
	}

	// Full path for the cache file
	filePath := filepath.Join(cacheDir, filename)

	// Write/update the file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Printf("Failed to write cache file %s: %v\n", filePath, err)
	}
}

// JobFetcher fetches job data from various APIs
type JobFetcher struct {
	client *http.Client
	Config *config.Config
}

// NewJobFetcher creates a new JobFetcher instance
func NewJobFetcher(config *config.Config) *JobFetcher {
	// Ensure cache directory exists
	os.MkdirAll("api_response_cache", 0755)

	return &JobFetcher{
		client: &http.Client{
			Timeout: 180 * time.Second, // Increase timeout to 3 minutes
		},
		Config: config,
	}
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrings {
		if strings.Contains(s, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

// FetchJSearchJobs fetches jobs from the JSearch API
func (jf *JobFetcher) FetchJSearchJobs(ctx context.Context) ([]models.Job, error) {
	apiKey := jf.Config.RapidAPIKey

	mode := jf.Config.Mode // Should be "dev" or "production" from .env

	var apiURL string
	if mode == "dev" {
		apiURL = "http://localhost:8081/jsearch/search"
	} else {
		apiURL = "https://jsearch.p.rapidapi.com/search"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", "golang jobs in nigeria")
	q.Add("page", "1")
	q.Add("num_pages", "3")
	q.Add("country", "ng")
	req.URL.RawQuery = q.Encode()

	req.Header.Add("x-rapidapi-host", "jsearch.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", apiKey)

	resp, err := jf.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cache the API response
	cacheResponse("jsearch_response.json", body)

	var jsearchResp models.JSEARCHResponse
	if err := json.Unmarshal(body, &jsearchResp); err != nil {
		return nil, err
	}

	jobs := make([]models.Job, len(jsearchResp.Data))
	for i, item := range jsearchResp.Data {
		now := time.Now()
		jobs[i] = models.Job{
			ID:          uuid.New().String(),
			JobID:       uuid.New().String(),
			Title:       item.JobTitle,
			Company:     item.EmployerName,
			CompanyURL:  item.CompanyURL,
			Location:    item.JobLocation,
			Description: item.JobDescription,
			URL:         item.JobApplyLink,
			Salary:      item.JobSalary,
			PostedAt:    item.JobPostedAt,
			JobType:     item.JobType,
			IsRemote:    item.JobIsRemote,
			Source:      "jsearch",
			RawData:     string(body),
			DateGotten:  now,
			ExpDate:     now.AddDate(0, 1, 0), // Expires in 1 month
		}
	}

	return jobs, nil
}

// FetchLinkedInJobs fetches jobs from LinkedIn API
func (jf *JobFetcher) FetchLinkedInJobs(ctx context.Context) ([]models.Job, error) {
	apiKey := jf.Config.RapidAPIKey

	mode := jf.Config.Mode // "dev" or "production"

	var apiURL string
	if mode == "dev" {
		apiURL = "http://localhost:8081/linkedin/active-jb-7d"
	} else {
		// Updated URL to match the correct endpoint format
		apiURL = "https://linkedin-job-search-api.p.rapidapi.com/active-jb-7d"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	// Update query parameters to match the expected format
	q.Add("limit", "20")
	q.Add("offset", "0")
	q.Add("title_filter", "golang")
	q.Add("location_filter", "nigeria")
	req.URL.RawQuery = q.Encode()

	req.Header.Add("x-rapidapi-host", "linkedin-job-search-api.p.rapidapi.com")
	req.Header.Add("x-rapidapi-key", apiKey)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := jf.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cache the API response
	cacheResponse("linkedin_response.json", body)

	// Try unmarshaling into different structures based on the response format
	// First, try unmarshaling as an array of items
	var jobArray []map[string]interface{}
	if err := json.Unmarshal(body, &jobArray); err == nil && len(jobArray) > 0 {
		// Parsed as an array - process accordingly
		jobs := make([]models.Job, len(jobArray))
		now := time.Now()

		for i, item := range jobArray {
			// Extract relevant fields from the map
			title, _ := item["title"].(string)
			id, _ := item["id"].(string)
			company, _ := item["organization"].(string)
			companyURL, _ := item["organization_url"].(string)
			url, _ := item["url"].(string)

			// Extract description if available
			description := ""
			if desc, ok := item["description"].(string); ok {
				description = desc
			}

			// Construct the Job object
			jobs[i] = models.Job{
				ID:          uuid.New().String(),
				JobID:       id,
				Title:       title,
				Company:     company,
				CompanyURL:  companyURL,
				URL:         url,
				Description: description,
				Location:    "Nigeria", // Default location
				Source:      "linkedin",
				RawData:     string(body),
				DateGotten:  now,
				ExpDate:     now.AddDate(0, 1, 0), // Expires in 1 month
			}

			// Extract optional fields when available
			if datePosted, ok := item["date_posted"].(string); ok {
				postedAt, err := time.Parse("2006-01-02T15:04:05", datePosted)
				if err != nil {
					jobs[i].PostedAt = now
				} else {
					jobs[i].PostedAt = postedAt
				}
			} else {
				jobs[i].PostedAt = now
			}

			// Extract location information
			if locationsArr, ok := item["locations_derived"].([]interface{}); ok && len(locationsArr) > 0 {
				if loc, ok := locationsArr[0].(string); ok {
					jobs[i].Location = loc
				}
			}
		}

		return jobs, nil
	}

	// If array parsing fails, try the original structure
	var linkedinResp models.LinkedInResponse
	err = json.Unmarshal(body, &linkedinResp)

	// If standard unmarshaling fails, try parsing as a raw JSON string
	if err != nil {
		// Check if the response looks like a JSON string (starts with a quote)
		if len(body) > 0 && body[0] == '"' {
			// Remove surrounding quotes and unescape
			var rawJSON string
			if err := json.Unmarshal(body, &rawJSON); err == nil {
				// Now try to parse the inner JSON
				if err := json.Unmarshal([]byte(rawJSON), &linkedinResp); err != nil {
					return nil, fmt.Errorf("failed to parse linkedin response from raw JSON string: %w", err)
				}
			} else {
				return nil, fmt.Errorf("failed to unmarshal linkedin response as JSON string: %w", err)
			}
		} else {
			return nil, fmt.Errorf("json unmarshal error: %w", err)
		}
	}

	// If we still have no data, return an error
	if linkedinResp.Data == nil || len(linkedinResp.Data) == 0 {
		return nil, fmt.Errorf("no data returned from LinkedIn API")
	}

	jobs := make([]models.Job, len(linkedinResp.Data))
	now := time.Now()

	for i, item := range linkedinResp.Data {
		// Get location from locations_derived, countries_derived, or default to Nigeria
		location := "Nigeria"
		if len(item.LocationsDerived) > 0 {
			location = item.LocationsDerived[0]
		} else if len(item.CountriesDerived) > 0 {
			location = item.CountriesDerived[0]
		}

		// Parse posted date
		postedAt, err := time.Parse("2006-01-02T15:04:05", item.DatePosted)
		if err != nil {
			// Try another format
			postedAt, err = time.Parse(time.RFC3339, item.DatePosted)
			if err != nil {
				postedAt = now // Use current time if parsing fails
			}
		}

		// Join employment types if present
		employmentType := ""
		if len(item.EmploymentType) > 0 {
			employmentType = strings.Join(item.EmploymentType, ", ")
		}

		jobs[i] = models.Job{
			ID:          uuid.New().String(),
			JobID:       item.ID,
			Title:       item.Title,
			Company:     item.Organization,
			CompanyURL:  item.OrganizationURL,
			JobType:     employmentType,
			Location:    location,
			URL:         item.URL,
			PostedAt:    postedAt,
			IsRemote:    item.RemoteDerived,
			Source:      "linkedin",
			RawData:     string(body),
			DateGotten:  now,
			ExpDate:     now.AddDate(0, 1, 0),        // Expires in 1 month
			Description: item.LinkedinOrgDescription, // Using org description as job description
		}
	}

	return jobs, nil
}

// FetchIndeedJobs fetches jobs from the Indeed API via Apify
func (jf *JobFetcher) FetchIndeedJobs(ctx context.Context) ([]models.Job, error) {
	apifyToken := jf.Config.ApifyAPIKey

	mode := jf.Config.Mode // "dev" or "production"

	var apiURL string
	if mode == "dev" {
		apiURL = "http://localhost:8081/apify/indeed/run-sync-get-dataset-items?token=random_test_token"
	} else {
		// Use the sync API endpoint that returns results directly
		apiURL = fmt.Sprintf("https://api.apify.com/v2/acts/misceres~indeed-scraper/run-sync-get-dataset-items?token=%s", apifyToken)
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"country":               "NG",
		"followApplyRedirects":  false,
		"maxItems":              20,
		"parseCompanyDetails":   true,
		"position":              "golang",
		"saveOnlyUniqueItems":   true,
		"forceResponseEncoding": "utf-8",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := jf.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cache the API response
	cacheResponse("indeed_response.json", body)

	// Check for error response first
	var errorResp []map[string]interface{}
	if json.Unmarshal(body, &errorResp) == nil && len(errorResp) > 0 {
		if errMsg, ok := errorResp[0]["error"].(string); ok {
			fmt.Printf("Indeed API error: %s\n", errMsg)
			// Return empty jobs array if we get an error from the API
			return []models.Job{}, nil
		}
	}

	var indeedResp models.MiscresIndeedResponse
	if err := json.Unmarshal(body, &indeedResp); err != nil {
		// Log detailed error
		fmt.Printf("Failed to unmarshal Indeed response: %v\nResponse body: %s\n", err, string(body))
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	if len(indeedResp) == 0 {
		// Just return an empty array instead of an error
		return []models.Job{}, nil
	}

	now := time.Now()
	jobs := make([]models.Job, len(indeedResp))

	for i, item := range indeedResp {
		jobType := ""
		if len(item.JobType) > 0 {
			jobType = item.JobType[0]
		}

		// Parse the posting date
		postedAt, err := time.Parse(time.RFC3339, item.PostingDateParsed)
		if err != nil {
			// Use the scrapedAt time if posting date parsing fails
			postedAt, err = time.Parse(time.RFC3339, item.ScrapedAt)
			if err != nil {
				// Fallback to current time if both date parsings fail
				postedAt = now
			}
		}

		jobs[i] = models.Job{
			ID:          uuid.New().String(),
			JobID:       item.ID,
			Title:       item.PositionName,
			Company:     item.Company,
			Location:    item.Location,
			Description: item.Description,
			URL:         item.URL,
			Salary:      item.Salary,
			PostedAt:    postedAt,
			JobType:     jobType,
			IsRemote:    containsAny(item.Description, []string{"remote", "work from home", "wfh"}),
			Source:      "apify indeed",
			RawData:     string(body),
			DateGotten:  now,
			ExpDate:     now.AddDate(0, 1, 0), // Expires in 1 month
		}
	}

	return jobs, nil
}

// Fetch from apify linkedin
func (jf *JobFetcher) FetchApifyLinkedInJobs(ctx context.Context) ([]models.Job, error) {
	apifyToken := jf.Config.ApifyAPIKey

	mode := jf.Config.Mode

	var apiURL string
	if mode == "dev" {
		apiURL = "http://localhost:8081/apify/linkedin/run-sync-get-dataset-items?token=random_test_token"
	} else {
		// Use the sync API endpoint that returns results directly
		apiURL = fmt.Sprintf("https://api.apify.com/v2/acts/curious_coder~linkedin-jobs-scraper/run-sync-get-dataset-items?token=%s", apifyToken)
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"urls":                  []string{"https://www.linkedin.com/jobs/search/?distance=25&geoId=105365761&keywords=golang"},
		"scrapeCompany":         true,
		"forceResponseEncoding": "utf-8",
		"maxItems":              20,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := jf.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Cache the API response
	cacheResponse("apify_linkedin_response.json", body)

	// Try to unmarshal as ApifyLinkedInResponse (array of jobs)
	var linkedInResp models.ApifyLinkedInResponse
	if err := json.Unmarshal(body, &linkedInResp); err != nil {
		// Log detailed error
		fmt.Printf("Failed to unmarshal Apify LinkedIn response: %v\nResponse body: %s\n", err, string(body))

		// Check if response contains error message
		var errorResp []map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil && len(errorResp) > 0 {
			if errMsg, ok := errorResp[0]["error"].(string); ok {
				return nil, fmt.Errorf("Apify LinkedIn API error: %s", errMsg)
			}
		}

		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}

	if len(linkedInResp) == 0 {
		return nil, fmt.Errorf("no data returned from Apify LinkedIn API")
	}

	now := time.Now()
	jobs := make([]models.Job, len(linkedInResp))

	for i, item := range linkedInResp {
		salary := ""
		if len(item.SalaryInfo) > 0 {
			salary = item.SalaryInfo[0]
		}

		// Parse posted date (format is likely YYYY-MM-DD)
		postedAt, err := time.Parse("2006-01-02", item.PostedAt)
		if err != nil {
			postedAt = now // Use current time if parsing fails
		}

		// Get the company website (either from CompanyWebsite or extract from LinkedIn URL)
		companyURL := item.CompanyWebsite
		if companyURL == "" && item.CompanyLinkedinUrl != "" {
			companyURL = item.CompanyLinkedinUrl
		}

		jobs[i] = models.Job{
			ID:          uuid.New().String(),
			JobID:       item.ID,
			Title:       item.Title,
			Company:     item.CompanyName,
			CompanyURL:  companyURL,
			Location:    item.Location,
			Description: item.DescriptionText,
			URL:         item.Link,
			Salary:      salary,
			JobType:     item.EmploymentType,
			IsRemote:    containsAny(item.DescriptionText, []string{"remote", "work from home", "wfh"}),
			Source:      "apify linkedin",
			PostedAt:    postedAt,
			RawData:     string(body),
			DateGotten:  now,
			ExpDate:     now.AddDate(0, 1, 0), // Expires in 1 month
		}
	}

	return jobs, nil
}
