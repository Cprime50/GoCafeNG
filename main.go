package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// Configuration holds API keys and settings
type Config struct {
	RapidAPIKey        string
	ApifyAPIKey        string
	PostgresConnection string
	Port               string
	DBConnStr          string
}

// JobFetcher fetches job data from various APIs
type JobFetcher struct {
	client *http.Client
	Config *Config
}

// NewJobFetcher creates a new JobFetcher instance
func NewJobFetcher(config *Config) *JobFetcher {
	return &JobFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		Config: config,
	}
}

// Job represents a job posting
type Job struct {
	ID             string    `json:"id"`
	JobID          string    `json:"job_id"`
	Title          string    `json:"title"`
	Company        string    `json:"company"`
	CompanyURL     string    `json:"company_url"`
	Country        string    `json:"country"`
	State          string    `json:"state"`
	Description    string    `json:"description"`
	URL            string    `json:"url"`
	Source         string    `json:"source"`
	IsRemote       bool      `json:"is_remote"`
	EmploymentType string    `json:"employment_type"`
	PostedAt       time.Time `json:"posted_at"`
	DateGotten     time.Time `json:"date_gotten"`
	ExpDate        time.Time `json:"exp_date"`
	Salary         string    `json:"salary"`
	Location       string    `json:"location"`
	JobType        string    `json:"job_type"`
	RawData        string    `json:"raw_data"`
}

// JSEARCHResponse represents the response from the JSearch API
type JSEARCHResponse struct {
	Data []struct {
		JobTitle       string    `json:"job_title"`
		EmployerName   string    `json:"employer_name"`
		JobLocation    string    `json:"job_location"`
		JobDescription string    `json:"job_description"`
		JobApplyLink   string    `json:"job_apply_link"`
		JobSalary      string    `json:"job_salary"`
		JobPostedAt    time.Time `json:"job_posted_at"`
		JobType        string    `json:"job_type"`
		JobIsRemote    bool      `json:"job_is_remote"`
		Source         string    `json:"source"`
	} `json:"data"`
}

// LinkedInResponse represents the response from the LinkedIn API
type LinkedInResponse struct {
	Data []struct {
		ID           string   `json:"id"`
		Title        string   `json:"title"`
		Company      string   `json:"company"`
		LocationData []string `json:"location_data"`
		URL          string   `json:"url"`
		PostedDate   string   `json:"posted_date"`
		IsRemote     bool     `json:"is_remote"`
	} `json:"data"`
}

// MiscresIndeedResponse represents the response from the Indeed API via Apify
type MiscresIndeedResponse []struct {
	ID           string    `json:"id"`
	PositionName string    `json:"position_name"`
	Company      string    `json:"company"`
	Location     string    `json:"location"`
	Description  string    `json:"description"`
	URL          string    `json:"url"`
	Salary       string    `json:"salary"`
	ScrapedAt    time.Time `json:"scraped_at"`
	JobType      []string  `json:"job_type"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	config := &Config{
		RapidAPIKey: os.Getenv("RAPID_API_KEY"),
		ApifyAPIKey: os.Getenv("APIFY_API_KEY"),
		Port:        os.Getenv("PORT"),
		DBConnStr:   os.Getenv("POSTGRES_CONNECTION"),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	return config, nil
}

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

	return db, nil
}

// InitSQLite initializes an in-memory SQLite database for logging
func InitSQLite() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create job_sync_logs table in memory
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
		return nil, err
	}

	return db, nil
}

// LogAPISync logs API sync attempts in SQLite
func LogAPISync(sqliteDB *sql.DB, apiName string, jobCount int, status string, errorMsg string) {
	_, err := sqliteDB.Exec(
		"INSERT INTO job_sync_logs (api_name, job_count, status, error_message) VALUES (?, ?, ?, ?)",
		apiName, jobCount, status, errorMsg,
	)
	if err != nil {
		log.Println("Error logging API sync:", err)
	}
}

func IsDuplicateJob(ctx context.Context, db *sql.DB, job Job) (bool, error) {
	var count int

	query := `
		SELECT COUNT(*) FROM jobs 
		WHERE LOWER(title) = LOWER($1) 
		AND LOWER(company) = LOWER($2) 
		AND EXTRACT(YEAR FROM posted_at) = EXTRACT(YEAR FROM $3::TIMESTAMP)
		AND EXTRACT(MONTH FROM posted_at) = EXTRACT(MONTH FROM $3::TIMESTAMP)
	`

	err := db.QueryRowContext(ctx, query, job.Title, job.Company, job.PostedAt).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}



// IsBlockedCompany checks if the company is in the blocked list
func IsBlockedCompany(companyName string) bool {
	blockedCompanies := []string{"canonical"}
	
	companyLower := strings.ToLower(companyName)
	for _, blocked := range blockedCompanies {
		if strings.Contains(companyLower, blocked) {
			return true
		}
	}
	return false
}

// SaveJobsToDB saves the jobs to the database with duplicate and blocked company filtering
func SaveJobsToDB(ctx context.Context, db *sql.DB, jobs []Job) (int, error) {
	// Use context for transaction to support cancelation
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	stmt, err := tx.PrepareContext(ctx, `
	INSERT INTO jobs (id, job_id, title, company, company_url, location, description, url, salary, 
		posted_at, job_type, is_remote, source, raw_data, date_gotten, country, state)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	ON CONFLICT (id) DO UPDATE SET
		title = EXCLUDED.title, 
		company = EXCLUDED.company,
		location = EXCLUDED.location,
		description = EXCLUDED.description,
		url = EXCLUDED.url,
		salary = EXCLUDED.salary,
		posted_at = EXCLUDED.posted_at,
		job_type = EXCLUDED.job_type,
		is_remote = EXCLUDED.is_remote,
		source = EXCLUDED.source,
		raw_data = EXCLUDED.raw_data,
		updated_at = CURRENT_TIMESTAMP
	`)

	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()

	count := 0
	skippedDuplicates := 0
	skippedBlockedCompanies := 0
	
	for _, job := range jobs {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			tx.Rollback()
			return count, ctx.Err()
		default:
		}
		
		// Skip jobs from blocked companies
		if IsBlockedCompany(job.Company) {
			log.Printf("Skipping job from blocked company: %s - %s", job.Company, job.Title)
			skippedBlockedCompanies++
			continue
		}
		
		// Check for duplicates
		isDuplicate, err := IsDuplicateJob(ctx, db, job)
		if err != nil {
			log.Printf("Error checking for duplicate job: %v", err)
			// Continue processing other jobs even if this check fails
		} else if isDuplicate {
			log.Printf("Skipping duplicate job: %s at %s (posted %s)", 
				job.Title, job.Company, job.PostedAt.Format("Jan 2006"))
			skippedDuplicates++
			continue
		}

		_, err = stmt.ExecContext(ctx,
			job.ID,
			job.JobID,
			job.Title,
			job.Company,
			job.CompanyURL,
			job.Location,
			job.Description,
			job.URL,
			job.Salary,
			job.PostedAt,
			job.JobType,
			job.IsRemote,
			job.Source,
			job.RawData,
			job.DateGotten,
			job.Country,
			job.State,
		)

		if err != nil {
			tx.Rollback()
			return count, err
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return count, err
	}

	log.Printf("Jobs processed: %d saved, %d duplicates skipped, %d from blocked companies skipped", 
		count, skippedDuplicates, skippedBlockedCompanies)
	
	return count, nil
}

// FetchJSearchJobs fetches jobs from the JSearch API
func (jf *JobFetcher) FetchJSearchJobs(ctx context.Context) ([]Job, error) {
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

	var jsearchResp JSEARCHResponse
	if err := json.Unmarshal(body, &jsearchResp); err != nil {
		return nil, err
	}

	jobs := make([]Job, len(jsearchResp.Data))
	for i, item := range jsearchResp.Data {
		now := time.Now()
		jobs[i] = Job{
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
func (jf *JobFetcher) FetchLinkedInJobs(ctx context.Context) ([]Job, error) {
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

	var linkedinResp LinkedInResponse
	if err := json.Unmarshal(body, &linkedinResp); err != nil {
		return nil, err
	}

	jobs := make([]Job, len(linkedinResp.Data))
	now := time.Now()

	for i, item := range linkedinResp.Data {
		location := "Nigeria"
		if len(item.LocationData) > 0 {
			location = item.LocationData[0]
		}

		postedAt, _ := time.Parse("2006-01-02T15:04:05", item.PostedDate)

		jobs[i] = Job{
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
func (jf *JobFetcher) FetchIndeedJobs(ctx context.Context) ([]Job, error) {
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

	var indeedResp MiscresIndeedResponse
	if err := json.Unmarshal(body, &indeedResp); err != nil {
		return nil, err
	}

	now := time.Now()
	jobs := make([]Job, len(indeedResp))
	for i, item := range indeedResp {
		jobType := ""
		if len(item.JobType) > 0 {
			jobType = item.JobType[0]
		}

		jobs[i] = Job{
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

// FetchAndSaveJSearch fetches and saves JSearch jobs
func FetchAndSaveJSearch(jobFetcher *JobFetcher, db *sql.DB, sqliteDB *sql.DB) {
	log.Println("Fetching JSearch jobs...")

	// Create a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchJSearchJobs(ctx)
	if err != nil {
		log.Println("Error fetching JSearch jobs:", err)
		LogAPISync(sqliteDB, "JSearch", 0, "Failed", err.Error())
		return
	}

	count, err := SaveJobsToDB(ctx, db, jobs)
	if err != nil {
		log.Println("Error saving JSearch jobs:", err)
		LogAPISync(sqliteDB, "JSearch", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d JSearch jobs", count)
		LogAPISync(sqliteDB, "JSearch", count, "Success", "")
	}
}

// FetchAndSaveIndeed fetches and saves Indeed jobs
func FetchAndSaveIndeed(jobFetcher *JobFetcher, db *sql.DB, sqliteDB *sql.DB) {
	log.Println("Fetching Indeed jobs...")

	// Create a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchIndeedJobs(ctx)
	if err != nil {
		log.Println("Error fetching Indeed jobs:", err)
		LogAPISync(sqliteDB, "Indeed", 0, "Failed", err.Error())
		return
	}

	count, err := SaveJobsToDB(ctx, db, jobs)
	if err != nil {
		log.Println("Error saving Indeed jobs:", err)
		LogAPISync(sqliteDB, "Indeed", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d Indeed jobs", count)
		LogAPISync(sqliteDB, "Indeed", count, "Success", "")
	}
}

// FetchAndSaveLinkedIn fetches and saves LinkedIn jobs
func FetchAndSaveLinkedIn(jobFetcher *JobFetcher, db *sql.DB, sqliteDB *sql.DB) {
	log.Println("Fetching LinkedIn jobs...")

	// Create a timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchLinkedInJobs(ctx)
	if err != nil {
		log.Println("Error fetching LinkedIn jobs:", err)
		LogAPISync(sqliteDB, "LinkedIn", 0, "Failed", err.Error())
		return
	}

	count, err := SaveJobsToDB(ctx, db, jobs)
	if err != nil {
		log.Println("Error saving LinkedIn jobs:", err)
		LogAPISync(sqliteDB, "LinkedIn", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d LinkedIn jobs", count)
		LogAPISync(sqliteDB, "LinkedIn", count, "Success", "")
	}
}

// StartJobScheduler runs job fetching on scheduled intervals using gocron
func StartJobScheduler(postgresDB *sql.DB, sqliteDB *sql.DB, config *Config) *gocron.Scheduler {
	// Create job fetcher
	jobFetcher := NewJobFetcher(config)

	// Create scheduler with UTC timezone
	scheduler := gocron.NewScheduler(time.UTC)

	// // Schedule JSearch jobs every 12 hours
	// scheduler.Every(12).Hours().Do(FetchAndSaveJSearch, jobFetcher, postgresDB, sqliteDB)

	// Schedule JSearch jobs every 12 hours
	scheduler.Every(1).Minute().Do(FetchAndSaveJSearch, jobFetcher, postgresDB, sqliteDB)

	// // Schedule Indeed jobs every 48 hours
	// scheduler.Every(48).Hours().Do(FetchAndSaveIndeed, jobFetcher, postgresDB, sqliteDB)

	// Schedule LinkedIn jobs every 24 hours
	// scheduler.Every(48).Hours().Do(FetchAndSaveLinkedIn, jobFetcher, postgresDB, sqliteDB)

	// Schedule LinkedIn jobs every 24 hours
	scheduler.Every(2).Minute().Do(FetchAndSaveLinkedIn, jobFetcher, postgresDB, sqliteDB)

	// Start the scheduler asynchronously
	scheduler.StartAsync()

	// Run all jobs immediately on startup
	scheduler.RunAll()

	return scheduler
}

// Main function
func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize SQLite for logging
	sqliteDB, err := InitSQLite()
	if err != nil {
		log.Fatal("Failed to initialize SQLite:", err)
	}
	defer sqliteDB.Close()

	// Connect to PostgreSQL
	postgresDB, err := InitDB(config.DBConnStr)
	if err != nil {
		log.Fatal("Failed to connect to Postgres:", err)
	}
	defer postgresDB.Close()

	// Start job scheduler
	scheduler := StartJobScheduler(postgresDB, sqliteDB, config)

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received, shutting down gracefully...")

	// Stop the scheduler
	scheduler.Stop()
	log.Println("Scheduler stopped, exiting.")
}
