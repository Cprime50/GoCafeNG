package models

import (
	"time"
)

// Job represents a job posting
type Job struct {
	ID             string    `json:"id"`
	JobID          string    `json:"job_id"`
	Title          string    `json:"title"`
	Company        string    `json:"company"`
	CompanyURL     string    `json:"company_url"`
	Country        string    `json:"country"`
	State          string    `json:"state"`
	Description    string    `json:"description"`
	URL            string    `json:"url"`
	Source         string    `json:"source"`
	IsRemote       bool      `json:"is_remote"`
	EmploymentType string    `json:"employment_type"`
	PostedAt       time.Time `json:"posted_at"`
	DateGotten     time.Time `json:"date_gotten"`
	ExpDate        time.Time `json:"exp_date"`
	Salary         string    `json:"salary"`
	Location       string    `json:"location"`
	JobType        string    `json:"job_type"`
	RawData        string    `json:"raw_data"`
}

// JSEARCHResponse represents the response from the JSearch API
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

// LinkedInResponse represents the response from the LinkedIn API
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

// MiscresIndeedResponse represents the response from the Indeed API via Apify
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


