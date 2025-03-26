package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"Go9jaJobs/internal/models"

	"github.com/google/uuid"
)

// EnsureCompanyDetailsTable ensures the company_details table exists
func EnsureCompanyDetailsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS company_details (
		id TEXT PRIMARY KEY,
		company_id TEXT NOT NULL,
		name TEXT,
		domain TEXT,
		description TEXT,
		logo_url TEXT,
		icon_url TEXT,
		accent_color TEXT,
		industry JSONB,
		links JSONB,
		raw_data TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`)

	if err != nil {
		return fmt.Errorf("failed to create company_details table: %w", err)
	}

	// Create index on company_id for faster lookups
	_, err = db.ExecContext(ctx, `
	CREATE INDEX IF NOT EXISTS idx_company_details_company_id ON company_details(company_id)
	`)

	if err != nil {
		return fmt.Errorf("failed to create index on company_details: %w", err)
	}

	return nil
}

// GetCompanyDetailsByCompanyID retrieves company details by company ID (name or domain)
func GetCompanyDetailsByCompanyID(ctx context.Context, db *sql.DB, companyID string) (*models.CompanyDetails, error) {
	var (
		companyDetails models.CompanyDetails
		industryJSON   sql.NullString
		linksJSON      sql.NullString
		createdAt      time.Time
		updatedAt      time.Time
	)

	query := `
	SELECT id, company_id, name, domain, description, logo_url, icon_url, 
	       accent_color, industry, links, raw_data, created_at, updated_at
	FROM company_details
	WHERE company_id = $1
	ORDER BY updated_at DESC
	LIMIT 1
	`

	err := db.QueryRowContext(ctx, query, companyID).Scan(
		&companyDetails.ID,
		&companyDetails.CompanyID,
		&companyDetails.Name,
		&companyDetails.Domain,
		&companyDetails.Description,
		&companyDetails.LogoURL,
		&companyDetails.IconURL,
		&companyDetails.AccentColor,
		&industryJSON,
		&linksJSON,
		&companyDetails.RawData,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Parse industry JSON
	if industryJSON.Valid {
		if err := json.Unmarshal([]byte(industryJSON.String), &companyDetails.Industry); err != nil {
			log.Printf("Error parsing industry JSON: %v", err)
		}
	}

	// Parse links JSON
	if linksJSON.Valid {
		if err := json.Unmarshal([]byte(linksJSON.String), &companyDetails.Links); err != nil {
			log.Printf("Error parsing links JSON: %v", err)
		}
	}

	companyDetails.CreatedAt = createdAt
	companyDetails.UpdatedAt = updatedAt

	return &companyDetails, nil
}

// SaveCompanyDetails saves company details to the database
func SaveCompanyDetails(ctx context.Context, db *sql.DB, details *models.CompanyDetails) error {
	if details.ID == "" {
		details.ID = uuid.New().String()
	}

	now := time.Now()
	details.CreatedAt = now
	details.UpdatedAt = now

	// Convert industry and links to JSON
	industryJSON, err := json.Marshal(details.Industry)
	if err != nil {
		return fmt.Errorf("error marshaling industry: %w", err)
	}

	linksJSON, err := json.Marshal(details.Links)
	if err != nil {
		return fmt.Errorf("error marshaling links: %w", err)
	}

	query := `
	INSERT INTO company_details (
		id, company_id, name, domain, description, logo_url, icon_url, 
		accent_color, industry, links, raw_data, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	ON CONFLICT (id) DO UPDATE SET
		company_id = EXCLUDED.company_id,
		name = EXCLUDED.name,
		domain = EXCLUDED.domain,
		description = EXCLUDED.description,
		logo_url = EXCLUDED.logo_url,
		icon_url = EXCLUDED.icon_url,
		accent_color = EXCLUDED.accent_color,
		industry = EXCLUDED.industry,
		links = EXCLUDED.links,
		raw_data = EXCLUDED.raw_data,
		updated_at = EXCLUDED.updated_at
	`

	_, err = db.ExecContext(ctx, query,
		details.ID,
		details.CompanyID,
		details.Name,
		details.Domain,
		details.Description,
		details.LogoURL,
		details.IconURL,
		details.AccentColor,
		string(industryJSON),
		string(linksJSON),
		details.RawData,
		details.CreatedAt,
		details.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("error saving company details: %w", err)
	}

	return nil
}

// FetchCompanyDetailsFromAPI fetches company details from BrandFetch API
func FetchCompanyDetailsFromAPI(ctx context.Context, companyName string, companyURL string) (*models.CompanyDetails, error) {
	// Get the domain from company URL or use company name
	domain := ""
	if companyURL != "" {
		parsedURL, err := url.Parse(companyURL)
		if err == nil && parsedURL.Host != "" {
			domain = parsedURL.Host
		} else {
			// Try parsing with a scheme
			parsedURL, err = url.Parse("https://" + companyURL)
			if err == nil && parsedURL.Host != "" {
				domain = parsedURL.Host
			}
		}
	}

	// Fall back to company name if domain extraction failed
	if domain == "" {
		domain = strings.ToLower(companyName)
		// Convert spaces to hyphens and remove illegal characters
		domain = strings.ReplaceAll(domain, " ", "-")
		domain = strings.ReplaceAll(domain, "&", "and")
		domain += ".com" // Add a domain suffix as a guess
	}

	// Remove www. prefix if present
	domain = strings.TrimPrefix(domain, "www.")

	// Remove URL paths and keep only the domain
	if idx := strings.Index(domain, "/"); idx != -1 {
		domain = domain[:idx]
	}

	// Make API request
	apiURL := fmt.Sprintf("https://api.brandfetch.io/v2/brands/%s", domain)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer 4kVJwDBb1fl6th2WhZ24C8tv6tW3x8qbj+N1ERtpfC0=")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making API request: %w", err)
	}
	defer res.Body.Close()

	// Read the response
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d - %s", res.StatusCode, string(body))
	}

	// Parse the response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Create company details
	companyDetails := &models.CompanyDetails{
		ID:        uuid.New().String(),
		CompanyID: strings.ToLower(companyName),
		Domain:    domain,
		RawData:   string(body),
	}

	// Extract basic info
	if name, ok := response["name"].(string); ok {
		companyDetails.Name = name
	} else {
		companyDetails.Name = companyName
	}

	if description, ok := response["description"].(string); ok {
		companyDetails.Description = description
	}

	// Extract color
	if colors, ok := response["colors"].([]interface{}); ok && len(colors) > 0 {
		for _, c := range colors {
			color, ok := c.(map[string]interface{})
			if !ok {
				continue
			}

			if colorType, ok := color["type"].(string); ok && colorType == "accent" {
				if hex, ok := color["hex"].(string); ok {
					companyDetails.AccentColor = hex
					break
				}
			}
		}
	}

	// Extract logo and icon
	if logos, ok := response["logos"].([]interface{}); ok {
		for _, l := range logos {
			logo, ok := l.(map[string]interface{})
			if !ok {
				continue
			}

			logoType, _ := logo["type"].(string)
			formats, ok := logo["formats"].([]interface{})
			if !ok || len(formats) == 0 {
				continue
			}

			format, ok := formats[0].(map[string]interface{})
			if !ok {
				continue
			}

			src, ok := format["src"].(string)
			if !ok {
				continue
			}

			if logoType == "logo" {
				companyDetails.LogoURL = src
			} else if logoType == "icon" {
				companyDetails.IconURL = src
			}
		}
	}

	// Extract links
	if links, ok := response["links"].([]interface{}); ok {
		for _, l := range links {
			link, ok := l.(map[string]interface{})
			if !ok {
				continue
			}

			name, nameOk := link["name"].(string)
			url, urlOk := link["url"].(string)

			if nameOk && urlOk {
				companyDetails.Links = append(companyDetails.Links, models.CompanyLink{
					Name: name,
					URL:  url,
				})
			}
		}
	}

	// Extract industry
	if company, ok := response["company"].(map[string]interface{}); ok {
		if industries, ok := company["industries"].([]interface{}); ok {
			for _, i := range industries {
				industry, ok := i.(map[string]interface{})
				if !ok {
					continue
				}

				if name, ok := industry["name"].(string); ok {
					companyDetails.Industry = append(companyDetails.Industry, name)
				}
			}
		}
	}

	return companyDetails, nil
}

// GetOrFetchCompanyDetails gets company details from the database or fetches from API
func GetOrFetchCompanyDetails(ctx context.Context, db *sql.DB, companyName string, companyURL string) (*models.CompanyDetails, error) {
	// Create a normalized company ID for lookup
	companyID := strings.ToLower(companyName)

	// Try to get from database first
	companyDetails, err := GetCompanyDetailsByCompanyID(ctx, db, companyID)
	if err != nil {
		return nil, fmt.Errorf("error getting company details from DB: %w", err)
	}

	// If found in database, return it
	if companyDetails != nil {
		return companyDetails, nil
	}

	// Not found, fetch from API
	companyDetails, err = FetchCompanyDetailsFromAPI(ctx, companyName, companyURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching company details from API: %w", err)
	}

	// Save to database
	err = SaveCompanyDetails(ctx, db, companyDetails)
	if err != nil {
		log.Printf("Error saving company details: %v", err)
		// Continue even if save fails, we'll return the details we fetched
	}

	return companyDetails, nil
}
