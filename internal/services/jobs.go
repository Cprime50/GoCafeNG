package services

import (
	"context"
	"database/sql"
	"log"
	"time"

	"Go9jaJobs/internal/db"
	"Go9jaJobs/internal/fetcher"
)

// FetchAndSaveJSearch fetches and saves JSearch jobs
func FetchAndSaveJSearch(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB) {
	log.Println("Fetching JSearch jobs...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchJSearchJobs(ctx)
	if err != nil {
		log.Println("Error fetching JSearch jobs:", err)
		db.LogAPISync(postgresDB, "JSearch", 0, "Failed", err.Error())
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving JSearch jobs:", err)
		db.LogAPISync(postgresDB, "JSearch", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d JSearch jobs", count)
	}
}

// FetchAndSaveIndeed fetches and saves Indeed jobs
func FetchAndSaveIndeed(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB) {
	log.Println("Fetching Indeed jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchIndeedJobs(ctx)
	if err != nil {
		log.Println("Error fetching Indeed jobs:", err)
		db.LogAPISync(postgresDB, "Indeed", 0, "Failed", err.Error())
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving Indeed jobs:", err)
		db.LogAPISync(postgresDB, "Indeed", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d Indeed jobs", count)
	}
}

// FetchAndSaveLinkedIn fetches and saves LinkedIn jobs
func FetchAndSaveLinkedIn(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB) {
	log.Println("Fetching LinkedIn jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchLinkedInJobs(ctx)
	if err != nil {
		log.Println("Error fetching LinkedIn jobs:", err)
		db.LogAPISync(postgresDB, "LinkedIn", 0, "Failed", err.Error())
		return
	} else {
		log.Println("Successfully fetched LinkedIn jobs")
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving LinkedIn jobs:", err)
		db.LogAPISync(postgresDB, "LinkedIn", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d LinkedIn jobs", count)
	}
}
func FetchAndSaveApifyLinkedIn(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB) {
	log.Println("Fetching LinkedIn jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchApifyLinkedInJobs(ctx)
	if err != nil {
		log.Println("Error fetching apifyLinkedIn jobs:", err)
		db.LogAPISync(postgresDB, "apifyLinkedIn", 0, "Failed", err.Error())
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving apifyLinkedIn jobs:", err)
		db.LogAPISync(postgresDB, "apifyLinkedIn", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d apifyLinkedIn jobs", count)
	}
}

//No longer neeeded as i will be using github actions to run the job
// // StartJobScheduler runs job fetching on scheduled intervals using gocron
//
