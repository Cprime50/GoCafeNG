package config

import (
	"crypto/rand"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds API keys and settings
type Config struct {
	RapidAPIKey        string
	ApifyAPIKey        string
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
		RapidAPIKey:    os.Getenv("RAPID_API_KEY"),
		ApifyAPIKey:    os.Getenv("APIFY_API_KEY"),
		Mode:           os.Getenv("MODE"),
		Port:           os.Getenv("PORT"),
		DBConnStr:      os.Getenv("POSTGRES_CONNECTION"),
		APIKey:         os.Getenv("API_KEY"),
		AllowedOrigins: parseAllowedOrigins(os.Getenv("ALLOWED_ORIGINS")),
	}

	if config.Port == "" {
		config.Port = "8080"
	}

	// Warn if secrets are missing
	if config.APIKey == "" {
		secret := "go9jajobs_api_key_" + randomString(32)
		log.Fatal("⚠️  API_KEY not set. Generating a strong secret...", secret)

	}

	// Warn if secrets are missing
	if config.Mode == "" {
		log.Fatal("⚠️  MODE not set. Defaulting to 'dev'")
		config.Mode = "dev"
	}

	return config, nil
}

func parseAllowedOrigins(origins string) []string {
	if origins == "" {
		return []string{"*"}
	}
	return strings.Split(origins, ",")
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			b[i] = 'a' // fallback character
			continue
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
