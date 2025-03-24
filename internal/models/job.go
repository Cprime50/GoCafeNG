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
		ID             string    `json:"job_id"`
		JobTitle       string    `json:"job_title"`
		EmployerName   string    `json:"employer_name"`
		CompanyURL     string    `json:"employer_webiste"`
		JobLocation    string    `json:"job_location"`
		JobDescription string    `json:"job_description"`
		JobApplyLink   string    `json:"job_apply_link"`
		JobSalary      string    `json:"job_salary"`
		JobPostedAt    time.Time `json:"job_posted_at_datetime_utc"`
		JobType        string    `json:"job_employment_type"`
		JobIsRemote    bool      `json:"job_is_remote"`
	} `json:"data"`
}

// LinkedInResponse represents the response from the LinkedIn API
type LinkedInResponse struct {
	Data []struct {
		ID           string   `json:"id"`
		Title        string   `json:"title"`
		Company      string   `json:"organization"`
		CompanyURL   string   `json:"organization_url"`
		LocationData []string `json:"locations_derived"`
		JobType      []string `json:"employment_type"`
		URL          string   `json:"url"`
		PostedDate   string   `json:"date_posted"`
		IsRemote     bool     `json:"remote_derived"`
	} `json:"data"`
}

// MiscresIndeedResponse represents the response from the Indeed API via Apify
type MiscresIndeedResponse []struct {
	ID           string    `json:"id"`
	PositionName string    `json:"positionName"`
	Company      string    `json:"company"`
	Location     string    `json:"location"`
	Description  string    `json:"description"`
	URL          string    `json:"url"`
	Salary       string    `json:"salary"`
	ScrapedAt    time.Time `json:"scraped_at"`
	JobType      []string  `json:"jobType"`
	PostedAt     time.Time `json:"postingDateParsed"`
}

// Apify LinkediN Response
type ApifyLinkedInResponse []struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	CompanyName    string    `json:"companyName"`
	CompanyUrl     string    `json:"companyWebsite"`
	Location       string    `json:"location"`
	SalaryInfo     []string  `json:"salaryInfo"`
	Description    string    `json:"descriptionText"`
	Link           string    `json:"link"`
	SeniorityLevel string    `json:"seniorityLevel"`
	EmploymentType string    `json:"employmentType"`
	PostedAt       time.Time `json:"postedAt"`
	ScrapedAt      time.Time `json:"scraped_at"`
}
