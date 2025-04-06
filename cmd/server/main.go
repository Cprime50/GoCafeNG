package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Go9jaJobs/internal/api"
	"Go9jaJobs/internal/config"
	"Go9jaJobs/internal/db"
	"Go9jaJobs/internal/fetcher"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	if cfg.APIKey == "" {
		log.Fatal("API Key must be set in configuration")
	}

	// Connect to PostgreSQL
	postgresDB, err := db.InitDB(cfg.DBConnStr)
	if err != nil {
		log.Fatal("Failed to connect to Postgres:", err)
	}
	log.Println("Connected to Postgres successfully")
	defer postgresDB.Close()

	// Create job fetcher
	jobFetcher := fetcher.NewJobFetcher(cfg)

	// Initialize API handlers
	apiHandler := api.NewHandler(postgresDB, jobFetcher)

	// Set up routes
	router := apiHandler.SetupRoutes(cfg) // Use SetupRoutes function

	// Create HTTP server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	serverAddress := ":" + port
	server := &http.Server{
		Addr:         serverAddress,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start job scheduler with persistent job schedule info
	//scheduler := services.StartJobScheduler(postgresDB, postgresDB, jobFetcher)

	// Start the server in a goroutine
	go func() {
		host := "localhost"
		if os.Getenv("HOST") != "" {
			host = os.Getenv("HOST")
		}

		url := fmt.Sprintf("http://%s:%s", host, port)
		log.Printf("======================================================")
		log.Printf("  Go9jaJobs API is now running at: \033[1;36m%s\033[0m", url)
		log.Printf("  Status endpoint: \033[1;36m%s/status\033[0m", url)
		log.Printf("  Jobs endpoint: \033[1;36m%s/api/jobs\033[0m", url)
		log.Printf("======================================================")
		log.Printf("  Remember to include X-API-Key, X-Timestamp, and X-Signature headers")
		log.Printf("  in all API requests for proper authentication.")
		log.Printf("======================================================")

		log.Printf("Server listening on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received, shutting down gracefully...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	//scheduler.Stop()
	log.Println("Server gracefully shut down, exiting.")
}
