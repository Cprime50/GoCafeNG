package api

import (
	"Go9jaJobs/internal/config"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Handler holds dependencies for API handlers
type Handler struct {
	DB *sql.DB
}

// NewHandler creates a new Handler instance
func NewHandler(sqlDB *sql.DB) *Handler {
	return &Handler{
		DB: sqlDB,
	}
}

func (h *Handler) SetupRoutes(cfg *config.Config) *mux.Router {
	r := mux.NewRouter()

	// Public route - No authentication middleware
	r.HandleFunc("/status", h.StatusCheck).Methods("GET")

	// Create protected subrouter
	protected := r.PathPrefix("/api").Subrouter()
	
	// Apply middleware chain to the protected subrouter
	protected.Use(LoggingMiddleware)
	protected.Use(APIKeyAuthMiddleware(cfg))
	protected.Use(SecurityHeadersMiddleware)
	protected.Use(CORSMiddleware(cfg.AllowedOrigins))
	
	// Add protected routes to the subrouter with middleware already applied
	protected.HandleFunc("/jobs", h.GetAllJobs).Methods("GET")

	return r
}

// StatusCheck returns a simple API status
func (h *Handler) StatusCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   "API is running",
	}

	json.NewEncoder(w).Encode(response)
}

// GetAllJobs returns all jobs from the database
func (h *Handler) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Query all jobs from the database
	rows, err := h.DB.Query(`
		SELECT 
			id, job_id, title, company, company_url, company_logo, location, description,
			url, salary, posted_at, job_type, is_remote, source
		FROM jobs
		ORDER BY posted_at DESC
	`)

	if err != nil {
		log.Printf("Error querying jobs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parse results
	var jobs []map[string]interface{}
	for rows.Next() {
		var (
			id          string
			jobID       string
			title       string
			company     string
			companyURL  sql.NullString
			companyLogo sql.NullString
			location    sql.NullString
			description sql.NullString
			url         sql.NullString
			salary      sql.NullString
			postedAt    time.Time
			jobType     sql.NullString
			isRemote    bool
			source      string
		)

		err := rows.Scan(
			&id, &jobID, &title, &company, &companyURL, &companyLogo, &location, &description,
			&url, &salary, &postedAt, &jobType, &isRemote, &source,
		)

		if err != nil {
			log.Printf("Error scanning job row: %v", err)
			continue
		}

		// Convert to a map to handle null values cleanly
		job := map[string]interface{}{
			"id":        id,
			"job_id":    jobID,
			"title":     title,
			"company":   company,
			"is_remote": isRemote,
			"source":    source,
			"posted_at": postedAt.Format(time.RFC3339),
		}

		// Add nullable fields only if they have values
		if companyURL.Valid {
			job["company_url"] = companyURL.String
		}
		if companyLogo.Valid {
			job["company_logo"] = companyLogo.String
		}
		if location.Valid {
			job["location"] = location.String
		}
		if description.Valid {
			job["description"] = description.String
		}
		if url.Valid {
			job["url"] = url.String
		}
		if salary.Valid {
			job["salary"] = salary.String
		}
		if jobType.Valid {
			job["job_type"] = jobType.String
		}

		jobs = append(jobs, job)
	}

	// Return JSON response
	response := map[string]interface{}{
		"success": true,
		"count":   len(jobs),
		"data":    jobs,
	}

	json.NewEncoder(w).Encode(response)
}
