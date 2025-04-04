package services

import (
	"context"
	"database/sql"
	"log"
	"time"

	"Go9jaJobs/internal/db"
	"Go9jaJobs/internal/fetcher"

	"github.com/go-co-op/gocron"
)

// FetchAndSaveJSearch fetches and saves JSearch jobs
func FetchAndSaveJSearch(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, logDB *sql.DB) {
	log.Println("Fetching JSearch jobs...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchJSearchJobs(ctx)
	if err != nil {
		log.Println("Error fetching JSearch jobs:", err)
		db.LogAPISync(logDB, "JSearch", 0, "Failed", err.Error())
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving JSearch jobs:", err)
		db.LogAPISync(logDB, "JSearch", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d JSearch jobs", count)
	}
}

// FetchAndSaveIndeed fetches and saves Indeed jobs
func FetchAndSaveIndeed(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, logDB *sql.DB) {
	log.Println("Fetching Indeed jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchIndeedJobs(ctx)
	if err != nil {
		log.Println("Error fetching Indeed jobs:", err)
		db.LogAPISync(logDB, "Indeed", 0, "Failed", err.Error())
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving Indeed jobs:", err)
		db.LogAPISync(logDB, "Indeed", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d Indeed jobs", count)
	}
}

// FetchAndSaveLinkedIn fetches and saves LinkedIn jobs
func FetchAndSaveLinkedIn(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, logDB *sql.DB) {
	log.Println("Fetching LinkedIn jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchLinkedInJobs(ctx)
	if err != nil {
		log.Println("Error fetching LinkedIn jobs:", err)
		db.LogAPISync(logDB, "LinkedIn", 0, "Failed", err.Error())
		return
	} else {
		log.Println("Successfully fetched LinkedIn jobs")
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving LinkedIn jobs:", err)
		db.LogAPISync(logDB, "LinkedIn", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d LinkedIn jobs", count)
	}
}
func FetchAndSaveApifyLinkedIn(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, logDB *sql.DB) {
	log.Println("Fetching LinkedIn jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchApifyLinkedInJobs(ctx)
	if err != nil {
		log.Println("Error fetching apifyLinkedIn jobs:", err)
		db.LogAPISync(logDB, "apifyLinkedIn", 0, "Failed", err.Error())
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving apifyLinkedIn jobs:", err)
		db.LogAPISync(logDB, "apifyLinkedIn", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d apifyLinkedIn jobs", count)
	}
}

// StartJobScheduler runs job fetching on scheduled intervals using gocron
func StartJobScheduler(postgresDB *sql.DB, logDB *sql.DB, jobFetcher *fetcher.JobFetcher) *gocron.Scheduler {
	// Load the timezone, default to UTC if it fails
	loc, err := time.LoadLocation("Africa/Lagos")
	if err != nil {
		log.Printf("Failed to load timezone 'Africa/Lagos': %v. Defaulting to UTC.", err)
		loc = time.UTC
	}

	scheduler := gocron.NewScheduler(loc)

	// Schedule JSearch jobs at 8 PM and 9 AM every day
	if _, err := scheduler.Every(1).Day().At("20:00").Do(FetchAndSaveJSearch, jobFetcher, postgresDB, logDB); err != nil {
		log.Println("Failed to schedule JSearch job at 8PM:", err)
	}
	if _, err := scheduler.Every(1).Day().At("09:00").Do(FetchAndSaveJSearch, jobFetcher, postgresDB, logDB); err != nil {
		log.Println("Failed to schedule JSearch job at 9AM:", err)
	}

	// Schedule jobs at 12 PM
	if _, err := scheduler.Every(1).Day().At("12:00").Do(func() {
		// Schedule Indeed, ApifyLinkedIn, and LinkedIn jobs
		FetchAndSaveIndeed(jobFetcher, postgresDB, logDB)
		FetchAndSaveApifyLinkedIn(jobFetcher, postgresDB, logDB)
	}); err != nil {
		log.Println("Failed to schedule 12 PM jobs:", err)
	}

	scheduler.StartAsync()
	return scheduler
}
