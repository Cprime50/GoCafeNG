package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds API keys and settings
type Config struct {
	RapidAPIKey        string
	ApifyAPIKey        string
	BrandFetchAPIKey   string
	Mode               string
	PostgresConnection string
	Port               string
	DBConnStr          string
	APIKey             string
	AllowedOrigins     []string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found")
	}

	config := &Config{
		RapidAPIKey:      os.Getenv("RAPID_API_KEY"),
		ApifyAPIKey:      os.Getenv("APIFY_API_KEY"),
		BrandFetchAPIKey: os.Getenv("API_TOKEN_LOGO"),
		Mode:             os.Getenv("MODE"),
		Port:             os.Getenv("PORT"),
		APIKey:           os.Getenv("API_KEY"),
		AllowedOrigins:   parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS")),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	if config.Mode == "production" {
		config.DBConnStr = os.Getenv("POSTGRES_CONNECTION_PROD")
	} else {
		config.DBConnStr = os.Getenv("POSTGRES_CONNECTION_LOCAL")
	}

	// Warn if secrets are missing
	if config.APIKey == "" {
		log.Fatal("API_KEY not set. Exiting.")

	}

	// Warn if secrets are missing
	if config.Mode == "" {
		log.Println("MODE not set. Defaulting to 'dev'")
		config.Mode = "dev"
	}

	log.Printf("Running in '%s' mode.", config.Mode)

	return config, nil
}

func parseAllowedOrigins(origins string) []string {
	if origins == "" {
		return []string{"*"}
	}
	return strings.Split(origins, ",")
}
