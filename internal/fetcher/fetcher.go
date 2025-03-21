package fetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"Go9jaJobs/internal/config"
	"Go9jaJobs/internal/models"

	"github.com/google/uuid"
)

// JobFetcher fetches job data from various APIs
type JobFetcher struct {
	client *http.Client
	Config *config.Config
}

// NewJobFetcher creates a new JobFetcher instance
func NewJobFetcher(config *config.Config) *JobFetcher {
	return &JobFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Config: config,
	}
}

// FetchJSearchJobs fetches jobs from the JSearch API
func (jf *JobFetcher) FetchJSearchJobs(ctx context.Context) ([]models.Job, error) {
	apiKey := jf.Config.RapidAPIKey
	// Create request with context for cancellation support

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/jsearch/search", nil)
	if err != nil {
		return nil, err
	}
	// req, err := http.NewRequestWithContext(ctx, "GET", "https://jsearch.p.rapidapi.com/search", nil)
	// if err != nil {
	// 	return nil, err
	// }

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

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/linkedin/active-jb-24h", nil)
	if err != nil {
		return nil, err
	}

	// req, err := http.NewRequestWithContext(ctx, "GET", "https://linkedin-job-search-api.p.rapidapi.com/active-jb-7d", nil)
	// if err != nil {
	// 	return nil, err
	// }

	q := req.URL.Query()
	q.Add("title_filter", "golang")
	q.Add("location_filter", "nigeria")
	req.URL.RawQuery = q.Encode()

	req.Header.Add("x-rapidapi-host", "linkedin-job-search-api.p.rapidapi.com")
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

	var linkedinResp models.LinkedInResponse
	if err := json.Unmarshal(body, &linkedinResp); err != nil {
		return nil, err
	}

	jobs := make([]models.Job, len(linkedinResp.Data))
	now := time.Now()

	for i, item := range linkedinResp.Data {
		location := "Nigeria"
		if len(item.LocationData) > 0 {
			location = item.LocationData[0]
		}

		postedAt, _ := time.Parse("2006-01-02T15:04:05", item.PostedDate)

		jobs[i] = models.Job{
			ID:         uuid.New().String(),
			JobID:      item.ID,
			Title:      item.Title,
			Company:    item.Company,
			Location:   location,
			URL:        item.URL,
			PostedAt:   postedAt,
			IsRemote:   item.IsRemote,
			Source:     "linkedin",
			RawData:    string(body),
			DateGotten: now,
			ExpDate:    now.AddDate(0, 1, 0), // Expires in 1 month
		}
	}

	return jobs, nil
}

// FetchIndeedJobs fetches jobs from the Indeed API via Apify
func (jf *JobFetcher) FetchIndeedJobs(ctx context.Context) ([]models.Job, error) {
	apifyToken := jf.Config.ApifyAPIKey

	// Prepare request payload
	payload := map[string]interface{}{
		"country":              "NG",
		"followApplyRedirects": false,
		"maxItems":             50,
		"parseCompanyDetails":  true,
		"position":             "golang",
		"saveOnlyUniqueItems":  true,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.apify.com/v2/acts/hMvNSpz3JnHgl5jkh/runs?token=%s", apifyToken)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}

	resp, err := jf.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var indeedResp models.MiscresIndeedResponse
	if err := json.Unmarshal(body, &indeedResp); err != nil {
		return nil, err
	}

	now := time.Now()
	jobs := make([]models.Job, len(indeedResp))
	for i, item := range indeedResp {
		jobType := ""
		if len(item.JobType) > 0 {
			jobType = item.JobType[0]
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
			PostedAt:    item.ScrapedAt,
			JobType:     jobType,
			IsRemote:    containsAny(item.Description, []string{"remote", "work from home", "wfh"}),
			Source:      "indeed",
			RawData:     string(body),
			DateGotten:  now,
			ExpDate:     now.AddDate(0, 1, 0), // Expires in 1 month
		}
	}
	return jobs, nil
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


