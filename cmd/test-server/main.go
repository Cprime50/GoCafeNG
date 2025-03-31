package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// Config holds application settings
type Config struct {
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	config := &Config{
		Port: os.Getenv("PORT"),
	}

	if config.Port == "" {
		config.Port = "8081" // Default port for the test API
	}

	return config, nil
}

// JSEARCHResponse simulates the response from the JSearch API
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

// LinkedInResponse simulates the response from the LinkedIn API
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

// MiscresIndeedResponse simulates the response from the Indeed API via Apify
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

// generateJSearchResponse generates a mock JSearch API response
func generateJSearchResponse() JSEARCHResponse {
	numItems := rand.Intn(10) + 5 // Generate between 5-14 items
	data := make([]struct {
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
	}, numItems)

	jobTypes := []string{"Full-time", "Part-time", "Contract", "Temporary", "Internship"}
	companies := []string{"TechCorp", "CodeNation", "DevHive", "ByteBuilders", "AlgoSystems", "CloudScape"}
	locations := []string{"Lagos, Nigeria", "Abuja, Nigeria", "Port Harcourt, Nigeria", "Ibadan, Nigeria", "Kano, Nigeria"}

	for i := 0; i < numItems; i++ {
		now := time.Now()
		postedAt := now.Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour) // Random date within last 30 days

		description := `We are looking for a Golang developer to join our team. Requirements:
		- Proficient in Go programming
		- Experience with RESTful APIs
		- Knowledge of SQL databases
		- Good understanding of cloud services
		- Experience with containerization technologies
		
		The ideal candidate will have 2+ years of experience working with Go in production environments.
		`

		data[i] = struct {
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
		}{
			JobTitle:       fmt.Sprintf("Golang Developer %d", i+1),
			EmployerName:   companies[rand.Intn(len(companies))],
			JobLocation:    locations[rand.Intn(len(locations))],
			JobDescription: description,
			JobApplyLink:   fmt.Sprintf("https://example.com/jobs/%d", i+1),
			JobSalary:      fmt.Sprintf("$%d,000 - $%d,000", 60+rand.Intn(40), 100+rand.Intn(50)),
			JobPostedAt:    postedAt,
			JobType:        jobTypes[rand.Intn(len(jobTypes))],
			JobIsRemote:    rand.Intn(2) == 1,
			Source:         "jsearch-test",
		}
	}

	return JSEARCHResponse{
		Data: data,
	}
}

// generateLinkedInResponse generates a mock LinkedIn API response
func generateLinkedInResponse() LinkedInResponse {
	numItems := rand.Intn(10) + 5 // Generate between 5-14 items
	data := make([]struct {
		ID           string   `json:"id"`
		Title        string   `json:"title"`
		Company      string   `json:"company"`
		LocationData []string `json:"location_data"`
		URL          string   `json:"url"`
		PostedDate   string   `json:"posted_date"`
		IsRemote     bool     `json:"is_remote"`
	}, numItems)

	companies := []string{"LinkedIn Corp", "Nigerian Tech", "AfroDevs", "CodeAfrica", "GoExperts", "Backend Masters"}
	locations := []string{"Lagos", "Abuja", "Port Harcourt", "Kaduna", "Enugu"}

	for i := 0; i < numItems; i++ {
		now := time.Now()
		daysAgo := rand.Intn(7) // Random date within last week
		postedAt := now.Add(-time.Duration(daysAgo) * 24 * time.Hour)
		postedDateStr := postedAt.Format("2006-01-02T15:04:05")

		locationData := []string{
			fmt.Sprintf("%s, Nigeria", locations[rand.Intn(len(locations))]),
		}

		data[i] = struct {
			ID           string   `json:"id"`
			Title        string   `json:"title"`
			Company      string   `json:"company"`
			LocationData []string `json:"location_data"`
			URL          string   `json:"url"`
			PostedDate   string   `json:"posted_date"`
			IsRemote     bool     `json:"is_remote"`
		}{
			ID:           fmt.Sprintf("linkedin-%s", uuid.New().String()),
			Title:        fmt.Sprintf("Senior Golang Engineer %d", i+1),
			Company:      companies[rand.Intn(len(companies))],
			LocationData: locationData,
			URL:          fmt.Sprintf("https://linkedin.com/jobs/%d", i+1),
			PostedDate:   postedDateStr,
			IsRemote:     rand.Intn(2) == 1,
		}
	}

	return LinkedInResponse{
		Data: data,
	}
}

// // generateIndeedResponse generates a mock Indeed API response
func generateIndeedResponse() MiscresIndeedResponse {
	numItems := rand.Intn(10) + 5 // Generate between 5-14 items
	data := make(MiscresIndeedResponse, numItems)

	companies := []string{"Indeed Tech", "Go Solutions", "Nigerian Dev Agency", "AfricaCode", "TechNaija"}
	locations := []string{"Lagos, Nigeria", "Abuja, Nigeria", "Remote, Nigeria", "Ibadan, Nigeria"}
	jobTypes := [][]string{
		{"Full-time"},
		{"Part-time"},
		{"Contract"},
		{"Full-time", "Remote"},
		{"Contract", "Remote"},
	}

	for i := 0; i < numItems; i++ {
		now := time.Now()
		daysAgo := rand.Intn(14) // Random date within last 2 weeks
		scrapedAt := now.Add(-time.Duration(daysAgo) * 24 * time.Hour)

		description := fmt.Sprintf(`
We are looking for an experienced Golang developer to build and maintain 
backend services for our growing tech company in Nigeria.

Required Skills:
- 2+ years Go programming experience
- Experience with RESTful APIs and microservices
- Knowledge of SQL and NoSQL databases
- Strong understanding of concurrency patterns
- Experience with Docker and Kubernetes

Salary: Competitive
Location: %s
`, locations[rand.Intn(len(locations))])

		data[i] = struct {
			ID           string    `json:"id"`
			PositionName string    `json:"position_name"`
			Company      string    `json:"company"`
			Location     string    `json:"location"`
			Description  string    `json:"description"`
			URL          string    `json:"url"`
			Salary       string    `json:"salary"`
			ScrapedAt    time.Time `json:"scraped_at"`
			JobType      []string  `json:"job_type"`
		}{
			ID:           fmt.Sprintf("indeed-%s", uuid.New().String()),
			PositionName: fmt.Sprintf("Golang Backend Engineer %d", i+1),
			Company:      companies[rand.Intn(len(companies))],
			Location:     locations[rand.Intn(len(locations))],
			Description:  description,
			URL:          fmt.Sprintf("https://indeed.com/viewjob?jk=%s", uuid.New().String()),
			Salary:       fmt.Sprintf("₦%d,000,000 - ₦%d,000,000 per year", 3+rand.Intn(3), 7+rand.Intn(5)),
			ScrapedAt:    scrapedAt,
			JobType:      jobTypes[rand.Intn(len(jobTypes))],
		}
	}

	return data
}

// Handle JSearch API requests
func handleJSearch(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query().Get("query")
	page := r.URL.Query().Get("page")
	country := r.URL.Query().Get("country")

	// Simulate API key validation
	apiKey := r.Header.Get("x-rapidapi-key")
	if apiKey == "" {
		http.Error(w, "Unauthorized: Missing API key", http.StatusUnauthorized)
		return
	}

	// Generate response
	response := generateJSearchResponse()

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Processed JSearch request: query=%s, country=%s, page=%s", query, country, page)
}

// Handle LinkedIn API requests
func handleLinkedIn(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	titleFilter := r.URL.Query().Get("title_filter")
	locationFilter := r.URL.Query().Get("location_filter")

	// Simulate API key validation
	apiKey := r.Header.Get("x-rapidapi-key")
	if apiKey == "" {
		http.Error(w, "Unauthorized: Missing API key", http.StatusUnauthorized)
		return
	}

	// Generate response
	response := generateLinkedInResponse()

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Processed LinkedIn request: title=%s, location=%s", titleFilter, locationFilter)
}

// Handle Indeed API requests via Apify
func handleIndeed(w http.ResponseWriter, r *http.Request) {
	// Simulate token validation
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
		return
	}

	// Generate response
	response := generateIndeedResponse()

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Printf("Processed Indeed request")
}

// Main function
func main() {
	// Initialize random seed

	// Create router
	router := mux.NewRouter()

	// Set up routes to simulate the original APIs
	router.HandleFunc("/jsearch/search", handleJSearch).Methods("GET")
	router.HandleFunc("/linkedin/active-jb-24h", handleLinkedIn).Methods("GET")
	router.HandleFunc("/apify/acts/hMvNSpz3JnHgl5jkh/runs", handleIndeed).Methods("POST")

	// Start server
	port := ":8081"
	log.Printf("Starting test job API server on port %s", port)
	log.Printf("JSearch API: http://localhost:%s/jsearch/search", port)
	log.Printf("LinkedIn API: http://localhost:%s/linkedin/active-jb-24h", port)
	log.Printf("Indeed API: http://localhost:%s/apify/indeed/runs", port)

	log.Fatal(http.ListenAndServe(port, router))
}
