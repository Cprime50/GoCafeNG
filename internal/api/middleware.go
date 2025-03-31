package api

import (
	"crypto/subtle"
	"log"
	"net/http"
	"time"

	"Go9jaJobs/internal/config"

	"github.com/rs/cors"
)

// CORSMiddleware applies CORS settings and blocks unauthorized origins
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if the request origin is in the allowed list
			allowed := false
			for _, ao := range allowedOrigins {
				if ao == "*" || origin == ao {
					allowed = true
					break
				}
			}

			// If the origin is not allowed, block the request
			if !allowed {
				http.Error(w, "CORS Forbidden", http.StatusForbidden)
				return
			}

			// Apply CORS settings
			c := cors.New(cors.Options{
				AllowedOrigins:   allowedOrigins,
				AllowCredentials: true,
				AllowedMethods:   []string{"GET", "OPTIONS"},
				AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization", "X-API-Key", "X-Timestamp", "X-Signature"},
				MaxAge:           600, // Cache preflight requests for 10 minutes
			})

			// Apply CORS middleware and pass request to next handler
			c.Handler(next).ServeHTTP(w, r)
		})
	}
}

func APIKeyAuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = r.URL.Query().Get("api_key")
			}

			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(cfg.APIKey)) != 1 {
				log.Printf("[AUTH FAIL] %s %s from %s - Invalid API Key attempt", r.Method, r.URL.Path, r.RemoteAddr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// If auth passes, log it
			log.Printf("[AUTH SUCCESS] %s %s from %s - API Key Validated", r.Method, r.URL.Path, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s from %s - %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

// SecurityHeadersMiddleware sets basic security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// TODO might add rate limiti later if needed
