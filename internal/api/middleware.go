package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"log"
	"net"
	"net/http"
	"strings"
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

// IPWhitelistMiddleware restricts access to specific IP addresses or ranges
func IPWhitelistMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.AllowedIPs == "" {
				// Allow all if no IPs are specified in the .env
				next.ServeHTTP(w, r)
				return
			}

			allowedIPs := strings.Split(cfg.AllowedIPs, ",")
			clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				log.Printf("Failed to parse client IP: %v", err)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			allowed := false
			for _, allowedIP := range allowedIPs {
				if clientIP == allowedIP {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuthMiddleware with HMAC validation
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

			// HMAC Validation
			timestamp := r.Header.Get("X-Timestamp")
			signature := r.Header.Get("X-Signature")
			if timestamp == "" || signature == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate timestamp (e.g., within 5 minutes)
			timeInt, err := time.Parse(time.RFC3339, timestamp)
			if err != nil || time.Since(timeInt) > 5*time.Minute {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Generate HMAC
			mac := hmac.New(sha256.New, []byte(cfg.APIKey))
			mac.Write([]byte(timestamp))
			expectedMAC := mac.Sum(nil)
			expectedSignature := hex.EncodeToString(expectedMAC)

			if subtle.ConstantTimeCompare([]byte(signature), []byte(expectedSignature)) != 1 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			log.Printf("[AUTH SUCCESS] %s %s from %s - API Key and HMAC Validated", r.Method, r.URL.Path, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuthSimpleMiddleware is a simpler version of API key authentication middleware
// that does not use timestamps or HMAC signatures. This is specifically for endpoints
// like /jobs/sync where simpler authentication is required.
func APIKeyAuthSimpleMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				apiKey = r.URL.Query().Get("api_key")
			}

			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(cfg.CronAPIKey)) != 1 {
				log.Printf("[AUTH FAIL] %s %s from %s - Invalid API Key attempt", r.Method, r.URL.Path, r.RemoteAddr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

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
