package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"Go9jaJobs/internal/config"

	"github.com/stretchr/testify/assert"
)

// mockHandler creates a simple handler for testing middleware
func mockHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})
}

func TestLoggingMiddleware(t *testing.T) {
	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create the middleware with our mock handler
	handler := LoggingMiddleware(mockHandler())

	// Serve the request through the middleware
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	assert.Equal(t, "test response", rr.Body.String())
}

func TestAPIKeyAuthMiddleware(t *testing.T) {
	// Create a test config with a API key
	testConfig := &config.Config{
		APIKey: "test-api-key",
	}

	// Create a middleware with our mock handler
	handler := APIKeyAuthMiddleware(testConfig)(mockHandler())

	// Test 1: Valid API key
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	req.Header.Set("X-API-Key", "test-api-key")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "test response", rr.Body.String())

	// Test 2: Missing API key
	req, err = http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	// Test 3: Invalid API key
	req, err = http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	req.Header.Set("X-API-Key", "wrong-api-key")
	rr = httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create the middleware with our mock handler
	handler := SecurityHeadersMiddleware(mockHandler())

	// Serve the request through the middleware
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check security headers
	headers := rr.Header()
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))

	// Check the response body
	assert.Equal(t, "test response", rr.Body.String())
}

func TestCORSMiddleware(t *testing.T) {
	// Test 1: With allowed origins as wildcard
	allowedOrigins := []string{"*"}

	// Create a test request
	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	req.Header.Set("Origin", "https://example.com")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create the middleware with our mock handler
	handler := CORSMiddleware(allowedOrigins)(mockHandler())

	// Serve the request through the middleware
	handler.ServeHTTP(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check CORS headers
	headers := rr.Header()
	assert.Equal(t, "https://example.com", headers.Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, OPTIONS", headers.Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Accept, Content-Type, Authorization, X-API-Key, X-Timestamp, X-Signature", headers.Get("Access-Control-Allow-Headers"))

	// Test 2: With specific allowed origins
	allowedOrigins = []string{"https://example.com", "https://subdomain.example.com"}

	// Create a test request with a matching origin
	req, err = http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	req.Header.Set("Origin", "https://example.com")

	// Create a ResponseRecorder to record the response
	rr = httptest.NewRecorder()

	// Create the middleware with our mock handler
	handler = CORSMiddleware(allowedOrigins)(mockHandler())

	// Serve the request through the middleware
	handler.ServeHTTP(rr, req)

	// Check the CORS headers
	headers = rr.Header()
	assert.Equal(t, "https://example.com", headers.Get("Access-Control-Allow-Origin"))

	// Test 3: With non-matching origin
	req, err = http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	req.Header.Set("Origin", "https://different-site.com")

	// Create a ResponseRecorder to record the response
	rr = httptest.NewRecorder()

	// Create the middleware with our mock handler
	handler = CORSMiddleware(allowedOrigins)(mockHandler())

	// Serve the request through the middleware
	handler.ServeHTTP(rr, req)

	// Check the CORS headers - should not have the origin in response
	headers = rr.Header()
	assert.NotEqual(t, "https://different-site.com", headers.Get("Access-Control-Allow-Origin"))
}
