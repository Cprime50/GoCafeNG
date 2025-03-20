package db

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"Go9jaJobs/internal/models"
)

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
func SaveJobsToDB(ctx context.Context, db *sql.DB, jobs []models.Job) (int, error) {
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
