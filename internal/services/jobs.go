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
func FetchAndSaveJSearch(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, sqliteDB *sql.DB) {
	log.Println("Fetching JSearch jobs...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchJSearchJobs(ctx)
	if err != nil {
		log.Println("Error fetching JSearch jobs:", err)
		db.LogAPISync(sqliteDB, "JSearch", 0, "Failed", err.Error())
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving JSearch jobs:", err)
		db.LogAPISync(sqliteDB, "JSearch", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d JSearch jobs", count)
		db.LogAPISync(sqliteDB, "JSearch", count, "Success", "")
	}
}

// FetchAndSaveIndeed fetches and saves Indeed jobs
func FetchAndSaveIndeed(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, sqliteDB *sql.DB) {
	log.Println("Fetching Indeed jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchIndeedJobs(ctx)
	if err != nil {
		log.Println("Error fetching Indeed jobs:", err)
		db.LogAPISync(sqliteDB, "Indeed", 0, "Failed", err.Error())
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving Indeed jobs:", err)
		db.LogAPISync(sqliteDB, "Indeed", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d Indeed jobs", count)
		db.LogAPISync(sqliteDB, "Indeed", count, "Success", "")
	}
}

// FetchAndSaveLinkedIn fetches and saves LinkedIn jobs
func FetchAndSaveLinkedIn(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, sqliteDB *sql.DB) {
	log.Println("Fetching LinkedIn jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchLinkedInJobs(ctx)
	if err != nil {
		log.Println("Error fetching LinkedIn jobs:", err)
		db.LogAPISync(sqliteDB, "LinkedIn", 0, "Failed", err.Error())
		return
	} else {
		log.Println("Successfully fetched LinkedIn jobs")
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving LinkedIn jobs:", err)
		db.LogAPISync(sqliteDB, "LinkedIn", count, "Partial Success", err.Error())
	} else {
		log.Printf("Successfully saved %d LinkedIn jobs", count)
		db.LogAPISync(sqliteDB, "LinkedIn", count, "Success", "")
	}
}

// StartJobScheduler runs job fetching on scheduled intervals using gocron
func StartJobScheduler(postgresDB *sql.DB, sqliteDB *sql.DB, jobFetcher *fetcher.JobFetcher) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.UTC)

	// // Schedule JSearch jobs
	scheduler.Every(12).Hours().StartAt(time.Now().Add(12*time.Hour)).Do(FetchAndSaveJSearch, jobFetcher, postgresDB, sqliteDB)
	//scheduler.Every(1).Minute().StartAt(time.Now().Add(1* time.Minute)).Do(FetchAndSaveJSearch, jobFetcher, postgresDB, sqliteDB)

	// // Schedule Indeed jobs
	// scheduler.Every(48).Hours()StartAt(time.Now().Add(48* time.Hour)).Do(FetchAndSaveIndeed, jobFetcher, postgresDB, sqliteDB)

	// Schedule LinkedIn jobs
	scheduler.Every(48).Hours().StartAt(time.Now().Add(48*time.Hour)).Do(FetchAndSaveLinkedIn, jobFetcher, postgresDB, sqliteDB)
	//scheduler.Every(1).Minute().StartAt(time.Now().Add(1* time.Minute)).Do(FetchAndSaveLinkedIn, jobFetcher, postgresDB, sqliteDB)

	// Start the scheduler asynchronously
	scheduler.StartAsync()

	// Run all jobs immediately on startup
	//scheduler.RunAll()

	return scheduler
}


