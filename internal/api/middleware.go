package api

import (
	"crypto/subtle"
	"log"
	"net/http"
	"time"

	"Go9jaJobs/internal/config"
)

// CORSMiddleware for CORS handling
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, ao := range allowedOrigins {
				if ao == "*" || origin == ao {
					allowed = true
					break
				}
			}
			if allowed && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(allowedOrigins) > 0 && allowedOrigins[0] != "*" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigins[0])
			} else {
				http.Error(w, "CORS Forbidden", http.StatusForbidden)
				return
			}

			w.Header().Set("Access-Control-Max-Age", "600")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization, X-API-Key, X-Timestamp, X-Signature")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
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