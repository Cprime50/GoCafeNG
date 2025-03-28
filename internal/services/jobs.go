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
func FetchAndSaveJSearch(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, intervalHours int) {
	log.Println("Fetching JSearch jobs...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchJSearchJobs(ctx)
	if err != nil {
		log.Println("Error fetching JSearch jobs:", err)
		db.LogJobRun(postgresDB, "JSearch", "Failed", 0, err.Error(), intervalHours)
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving JSearch jobs:", err)
		db.LogJobRun(postgresDB, "JSearch", "Partial Success", count, err.Error(), intervalHours)
	} else {
		log.Printf("Successfully saved %d JSearch jobs", count)
		db.LogJobRun(postgresDB, "JSearch", "Success", count, "", intervalHours)
	}
}

// FetchAndSaveIndeed fetches and saves Indeed jobs
func FetchAndSaveIndeed(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, intervalHours int) {
	log.Println("Fetching Indeed jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchIndeedJobs(ctx)
	if err != nil {
		log.Println("Error fetching Indeed jobs:", err)
		db.LogJobRun(postgresDB, "Indeed", "Failed", 0, err.Error(), intervalHours)
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving Indeed jobs:", err)
		db.LogJobRun(postgresDB, "Indeed", "Partial Success", count, err.Error(), intervalHours)
	} else {
		log.Printf("Successfully saved %d Indeed jobs", count)
		db.LogJobRun(postgresDB, "Indeed", "Success", count, "", intervalHours)
	}
}

// FetchAndSaveLinkedIn fetches and saves LinkedIn jobs
func FetchAndSaveLinkedIn(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, intervalHours int) {
	log.Println("Fetching LinkedIn jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchLinkedInJobs(ctx)
	if err != nil {
		log.Println("Error fetching LinkedIn jobs:", err)
		db.LogJobRun(postgresDB, "LinkedIn", "Failed", 0, err.Error(), intervalHours)
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving LinkedIn jobs:", err)
		db.LogJobRun(postgresDB, "LinkedIn", "Partial Success", count, err.Error(), intervalHours)
	} else {
		log.Printf("Successfully saved %d LinkedIn jobs", count)
		db.LogJobRun(postgresDB, "LinkedIn", "Success", count, "", intervalHours)
	}
}

func FetchAndSaveApifyLinkedIn(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, intervalHours int) {
	log.Println("Fetching ApifyLinkedIn jobs...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	jobs, err := jobFetcher.FetchApifyLinkedInJobs(ctx)
	if err != nil {
		log.Println("Error fetching ApifyLinkedIn jobs:", err)
		db.LogJobRun(postgresDB, "ApifyLinkedIn", "Failed", 0, err.Error(), intervalHours)
		return
	}

	count, err := db.SaveJobsToDB(ctx, postgresDB, jobs)
	if err != nil {
		log.Println("Error saving ApifyLinkedIn jobs:", err)
		db.LogJobRun(postgresDB, "ApifyLinkedIn", "Partial Success", count, err.Error(), intervalHours)
	} else {
		log.Printf("Successfully saved %d ApifyLinkedIn jobs", count)
		db.LogJobRun(postgresDB, "ApifyLinkedIn", "Success", count, "", intervalHours)
	}
}

// StartJobScheduler runs job fetching on scheduled intervals using gocron
func StartJobScheduler(postgresDB *sql.DB, jobFetcher *fetcher.JobFetcher) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.UTC)

	// Create a map of API names to their interval hours for easier management
	jobConfigs := map[string]struct {
		intervalHours int
		fetchFunc     func(jobFetcher *fetcher.JobFetcher, postgresDB *sql.DB, intervalHours int)
	}{
		"JSearch":       {12, FetchAndSaveJSearch},
		"Indeed":        {24, FetchAndSaveIndeed},
		//"LinkedIn":      {24, FetchAndSaveLinkedIn},
		"ApifyLinkedIn": {24, FetchAndSaveApifyLinkedIn},
	}

	scheduleInfos, err := db.GetAllJobScheduleInfo(postgresDB)
	if err != nil {
		log.Printf("Error getting job schedule info: %v", err)
	}


	scheduleMap := make(map[string]db.JobScheduleInfo)
	for _, info := range scheduleInfos {
		scheduleMap[info.ApiName] = info
	}

	now := time.Now()

	// Set up each job with its proper schedule
	for apiName, config := range jobConfigs {
		// Check if we have persistent schedule info
		info, exists := scheduleMap[apiName]

		// Calculate when the next run should be
		var nextRun time.Time
		if exists && !info.NextRunTime.IsZero() {
			// If we have a valid next run time from the DB, use it
			nextRun = info.NextRunTime

			// If next run time is in the past, schedule immediately
			if nextRun.Before(now) {
				nextRun = now.Add(1 * time.Minute) // Run in 1 minute
			}
		} else {
			// For new jobs or missing next run time, start in 1 hour
			nextRun = now.Add(1 * time.Hour)
		}

		// Initialize or update the job info in the database
		if !exists {
			// This is a new entry, create it
			initialInfo := db.JobScheduleInfo{
				ApiName:       apiName,
				IntervalHours: config.intervalHours,
				NextRunTime:   nextRun,
				Status:        "Scheduled",
			}
			if err := db.UpdatesJobScheduleInfo(postgresDB, initialInfo); err != nil {
				log.Printf("Error initializing schedule for %s: %v", apiName, err)
			}
		}

		// Set up the scheduler
		job := scheduler.Every(config.intervalHours).Hours()

		// Use the next run time from our calculations
		job.StartAt(nextRun)

		// Schedule the job with the appropriate function and parameters
		job.Do(config.fetchFunc, jobFetcher, postgresDB, config.intervalHours)

		log.Printf("Scheduled %s job to run every %d hours, starting at %s",
			apiName, config.intervalHours, nextRun.Format(time.RFC3339))
	}

	// Start the scheduler asynchronously
	scheduler.StartAsync()

	return scheduler
}

// StopJobScheduler properly shuts down the job scheduler
func StopJobScheduler(scheduler *gocron.Scheduler, postgresDB *sql.DB) {
	log.Println("Stopping job scheduler...")

	// Get all pending jobs
	pendingJobs := scheduler.Jobs()

	// For each job, calculate and store the next run time
	for _, job := range pendingJobs {
		// Get the job's next run time
		nextRun := job.NextRun()
		tagList := job.Tags()

		// If the job has a tag that matches our API names, update its next run time
		if len(tagList) > 0 {
			apiName := tagList[0]

			// Get current info
			info, err := db.GetJobScheduleInfo(postgresDB, apiName)
			if err != nil || info == nil {
				log.Printf("Warning: Could not get schedule info for %s: %v", apiName, err)
				continue
			}

			// Update next run time
			info.NextRunTime = nextRun

			// Save changes
			if err := db.UpdatesJobScheduleInfo(postgresDB, *info); err != nil {
				log.Printf("Error updating next run time for %s: %v", apiName, err)
			} else {
				log.Printf("Updated next run time for %s to %s", apiName, nextRun.Format(time.RFC3339))
			}
		}
	}

	// Stop the scheduler
	scheduler.Stop()
	log.Println("Job scheduler stopped.")
}
