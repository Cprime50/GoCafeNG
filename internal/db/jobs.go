package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"Go9jaJobs/internal/config"
	"Go9jaJobs/internal/models"
)

// BrandFetchResponse represents the response from the BrandFetch API
type BrandFetchResponse struct {
	Logos []struct {
		Formats []struct {
			Src    string `json:"src"`
			Format string `json:"format"`
		} `json:"formats"`
		Type string `json:"type"`
	} `json:"logos"`
}

// FetchCompanyLogo fetches a company logo using the BrandFetch API
func FetchCompanyLogo(companyURL, apiToken string) string {
	if companyURL == "" {
		return ""
	}
	// Extract domain from URL
	parsedURL, err := url.Parse(companyURL)
	if err != nil {
		log.Printf("Error parsing URL %s: %v", companyURL, err)
		return ""
	}

	domain := parsedURL.Host
	if domain == "" {
		// If URL doesn't have a scheme, try using the path
		domain = parsedURL.Path
	}

	// Remove www. prefix if present
	domain = strings.TrimPrefix(domain, "www.")

	// Remove any path components
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	if domain == "" {
		return ""
	}

	// Create a request to BrandFetch API
	apiURL := fmt.Sprintf("https://api.brandfetch.io/v2/brands/%s", domain)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("Error creating LogoFetch request for %s: %v", domain, err)
		return ""
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	// Make the request
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching logo for %s: %v", domain, err)
		return ""
	}
	defer res.Body.Close()

	// Check if the request was successful
	if res.StatusCode != http.StatusOK {
		log.Printf("LogoFetch API returned non-200 status for %s: %d", domain, res.StatusCode)
		return ""
	}

	// Read and parse the response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading LogoFetch response for %s: %v", domain, err)
		return ""
	}

	var brandResponse BrandFetchResponse
	if err := json.Unmarshal(body, &brandResponse); err != nil {
		log.Printf("Error parsing LogoFetch response for %s: %v", domain, err)
		return ""
	}

	// Extract the first logo URL
	for _, logo := range brandResponse.Logos {
		if len(logo.Formats) > 0 {
			return logo.Formats[0].Src
		}
	}

	return ""
}

// IsDuplicateJob checks if a job already exists in the database
func IsDuplicateJob(ctx context.Context, db *sql.DB, job models.Job) (bool, error) {
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
	blockedCompanies := []string{"canonical", "crossover"}

	companyLower := strings.ToLower(companyName)
	for _, blocked := range blockedCompanies {
		if strings.Contains(companyLower, blocked) {
			return true
		}
	}
	return false
}

// IsGoRelatedJob checks if a job is Go-related by looking for "go" or "golang" in title or description
func IsGoRelatedJob(job models.Job) bool {
	title := strings.ToLower(job.Title)
	description := strings.ToLower(job.Description)

	// Check for "go" as a whole word with different patterns:
	// - surrounded by spaces: " go "
	// - at beginning: "go "
	// - at end: " go"
	// - with punctuation: "(go)", "[go]", ",go", etc.
	// - with different capitalization: "Go", "GO"

	// Word boundary patterns to check
	goPrefixes := []string{" go ", " go,", " go.", " go:", " go;", " go-", " go/", " go)", " go]", " go}", "(go ", "[go ", "{go "}
	goSuffixes := []string{" go", ",go ", ".go ", ":go ", ";go ", "-go ", "/go ", "(go)", "[go]", "{go}"}
	goStandalone := []string{"(go)", "[go]", "{go}", " go "}

	// Check for the word "go" with various patterns
	for _, pattern := range goPrefixes {
		if strings.Contains(title, pattern) || strings.Contains(description, pattern) {
			return true
		}
	}

	for _, pattern := range goSuffixes {
		if strings.Contains(title, pattern) || strings.Contains(description, pattern) {
			return true
		}
	}

	for _, pattern := range goStandalone {
		if strings.Contains(title, pattern) || strings.Contains(description, pattern) {
			return true
		}
	}

	// Special case: if title starts with "go" or ends with "go"
	if strings.HasPrefix(title, "go ") || strings.HasSuffix(title, " go") {
		return true
	}

	// Check for "golang" anywhere in title or description
	if strings.Contains(title, "golang") || strings.Contains(description, "golang") {
		return true
	}

	return false
}




// SaveJobsToDB saves the jobs to the database with duplicate and blocked company filtering
func SaveJobsToDB(ctx context.Context, db *sql.DB, jobs []models.Job) (int, error) {
	// Get config to access BrandFetch API token
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: Failed to load config for logo fetching: %v", err)
	}

	// Use context for transaction to support cancelation
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	stmt, err := tx.PrepareContext(ctx, `
	INSERT INTO jobs (id, job_id, title, company, company_url, company_logo, location, description, url, salary, 
		posted_at, job_type, is_remote, source, raw_data, date_gotten, country, state)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
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
		company_logo = EXCLUDED.company_logo,
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
	skippedNonGoJobs := 0


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

		// Skip jobs that are not Go-related
		if !IsGoRelatedJob(job) {
			log.Printf("Skipping non-Go related job: %s at %s", job.Title, job.Company)
			skippedNonGoJobs++
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

		// If we have a config and the job doesn't have a logo, try to fetch one
		if  cfg != nil && cfg.Mode != "dev" && cfg.BrandFetchAPIKey != "" && job.CompanyLogo == "" && job.CompanyURL != "" {
			job.CompanyLogo = FetchCompanyLogo(job.CompanyURL, cfg.BrandFetchAPIKey)
			if job.CompanyLogo != "" {
				log.Printf("Fetched logo for %s from BrandFetch", job.Company)
			}
		}

		_, err = stmt.ExecContext(ctx,
			job.ID,
			job.JobID,
			job.Title,
			job.Company,
			job.CompanyURL,
			job.CompanyLogo,
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

	log.Printf("Jobs processed: %d saved, %d duplicates skipped, %d from blocked companies skipped, %d non-Go jobs skipped",
		count, skippedDuplicates, skippedBlockedCompanies, skippedNonGoJobs)

	return count, nil
}






