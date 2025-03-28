package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupEnvVars sets environment variables for testing
func setupEnvVars(t *testing.T) {
	// Set required environment variables
	t.Setenv("RAPID_API_KEY", "test-rapid-api-key")
	t.Setenv("APIFY_API_KEY", "test-apify-api-key")
	t.Setenv("API_TOKEN_LOGO", "test-logo-api-token")
	t.Setenv("MODE", "dev")
	t.Setenv("PORT", "8080")
	t.Setenv("API_KEY", "test-api-key")
	t.Setenv("ALLOWED_ORIGINS", "https://example.com,https://app.example.com")
	t.Setenv("POSTGRES_CONNECTION_LOCAL", "postgres://localhost:5432/testdb")
	t.Setenv("POSTGRES_CONNECTION_PROD", "postgres://prod:5432/proddb")
}

// clearEnvVars clears environment variables to prevent test interference
func clearEnvVars(t *testing.T) {
	vars := []string{
		"RAPID_API_KEY",
		"APIFY_API_KEY",
		"API_TOKEN_LOGO",
		"MODE",
		"PORT",
		"API_KEY",
		"ALLOWED_ORIGINS",
		"POSTGRES_CONNECTION_LOCAL",
		"POSTGRES_CONNECTION_PROD",
	}

	for _, v := range vars {
		t.Setenv(v, "")
	}
}

func TestLoadConfig(t *testing.T) {
	// Set up environment variables for the test
	setupEnvVars(t)
	defer clearEnvVars(t)

	// Load the configuration
	cfg, err := LoadConfig()

	// Check that no error occurred
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check that the values were loaded correctly
	assert.Equal(t, "test-rapid-api-key", cfg.RapidAPIKey)
	assert.Equal(t, "test-apify-api-key", cfg.ApifyAPIKey)
	assert.Equal(t, "test-logo-api-token", cfg.BrandFetchAPIKey)
	assert.Equal(t, "dev", cfg.Mode)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "test-api-key", cfg.APIKey)
	assert.Equal(t, "postgres://localhost:5432/testdb", cfg.DBConnStr)

	// Test AllowedOrigins parsing
	expectedOrigins := []string{"https://example.com", "https://app.example.com"}
	assert.Equal(t, expectedOrigins, cfg.AllowedOrigins)
}

func TestLoadConfigMissingEnv(t *testing.T) {
	// Clear all environment variables
	clearEnvVars(t)

	// Set minimal required variables
	t.Setenv("API_KEY", "test-api-key")

	// Load the config
	cfg, err := LoadConfig()

	// Check that no error occurred even with missing variables
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check default values
	assert.Equal(t, "dev", cfg.Mode)
	assert.Equal(t, "8080", cfg.Port)

	// Check that allowed origins defaults to wildcard
	assert.Equal(t, []string{"*"}, cfg.AllowedOrigins)
}

func TestLoadConfigProdMode(t *testing.T) {
	// Set up environment variables in production mode
	setupEnvVars(t)
	t.Setenv("MODE", "production")
	defer clearEnvVars(t)

	// Load the configuration
	cfg, err := LoadConfig()

	// Check that no error occurred
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Should use production database connection string
	assert.Equal(t, "production", cfg.Mode)
	assert.Equal(t, "postgres://prod:5432/proddb", cfg.DBConnStr)
}

func TestParseAllowedOrigins(t *testing.T) {
	// Test with multiple origins
	origins := parseAllowedOrigins("https://example.com,https://app.example.com")
	assert.Equal(t, []string{"https://example.com", "https://app.example.com"}, origins)

	// Test with single origin
	origins = parseAllowedOrigins("https://example.com")
	assert.Equal(t, []string{"https://example.com"}, origins)

	// Test with empty string
	origins = parseAllowedOrigins("")
	assert.Equal(t, []string{"*"}, origins)
}
