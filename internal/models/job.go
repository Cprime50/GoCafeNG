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
	CompanyLogo    string    `json:"company_logo"`
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
		CompanyURL     string    `json:"employer_website"`
		EmployerLogo   string    `json:"employer_logo"`
		JobLocation    string    `json:"job_location"`
		JobDescription string    `json:"job_description"`
		JobApplyLink   string    `json:"job_apply_link"`
		JobSalary      string    `json:"job_salary"`
		JobPostedAt    time.Time `json:"job_posted_at_datetime_utc"`
		JobType        string    `json:"job_employment_type"`
		JobIsRemote    bool      `json:"job_is_remote"`
	} `json:"data"`
}

// LinkedInResponse represents the response from the LinkedIn API via RapidAPI
type LinkedInResponse struct {
	Data []struct {
		ID                                  string   `json:"id"`
		DatePosted                          string   `json:"date_posted"`
		DateCreated                         string   `json:"date_created"`
		Title                               string   `json:"title"`
		Organization                        string   `json:"organization"`
		OrganizationURL                     string   `json:"organization_url"`
		DateValidthrough                    string   `json:"date_validthrough"`
		LocationsRaw                        string   `json:"locations_raw,omitempty"`
		LocationType                        *string  `json:"location_type"`
		LocationRequirements                *string  `json:"location_requirements_raw"`
		SalaryRaw                           *string  `json:"salary_raw"`
		EmploymentType                      []string `json:"employment_type"`
		URL                                 string   `json:"url"`
		SourceType                          string   `json:"source_type"`
		Source                              string   `json:"source"`
		SourceDomain                        string   `json:"source_domain"`
		OrganizationLogo                    string   `json:"organization_logo"`
		CitiesDerived                       []string `json:"cities_derived"`
		RegionsDerived                      []string `json:"regions_derived"`
		CountriesDerived                    []string `json:"countries_derived"`
		LocationsDerived                    []string `json:"locations_derived"`
		TimezonesDerived                    []string `json:"timezones_derived"`
		LatsDerived                         []string `json:"lats_derived"`
		LngsDerived                         []string `json:"lngs_derived"`
		RemoteDerived                       bool     `json:"remote_derived"`
		RecruiterName                       *string  `json:"recruiter_name"`
		RecruiterTitle                      *string  `json:"recruiter_title"`
		RecruiterURL                        *string  `json:"recruiter_url"`
		LinkedinOrgEmployees                int      `json:"linkedin_org_employees"`
		LinkedinOrgURL                      string   `json:"linkedin_org_url"`
		LinkedinOrgSize                     string   `json:"linkedin_org_size"`
		LinkedinOrgSlogan                   string   `json:"linkedin_org_slogan"`
		LinkedinOrgIndustry                 string   `json:"linkedin_org_industry"`
		LinkedinOrgFollowers                int      `json:"linkedin_org_followers"`
		LinkedinOrgHeadquarters             string   `json:"linkedin_org_headquarters"`
		LinkedinOrgType                     string   `json:"linkedin_org_type"`
		LinkedinOrgFoundeddate              string   `json:"linkedin_org_foundeddate"`
		LinkedinOrgSpecialties              []string `json:"linkedin_org_specialties"`
		LinkedinOrgLocations                []string `json:"linkedin_org_locations"`
		LinkedinOrgDescription              string   `json:"linkedin_org_description"`
		LinkedinOrgRecruitmentAgencyDerived bool     `json:"linkedin_org_recruitment_agency_derived"`
	} `json:"data"`
}

// MiscresIndeedResponse represents the response from the Indeed API via Apify
type MiscresIndeedResponse []struct {
	Salary            string   `json:"salary"`
	PostedAt          string   `json:"postedAt"`
	ExternalApplyLink *string  `json:"externalApplyLink"`
	PositionName      string   `json:"positionName"`
	JobType           []string `json:"jobType"`
	Company           string   `json:"company"`
	Location          string   `json:"location"`
	Rating            float64  `json:"rating"`
	ReviewsCount      int      `json:"reviewsCount"`
	URLInput          *string  `json:"urlInput"`
	URL               string   `json:"url"`
	ID                string   `json:"id"`
	ScrapedAt         string   `json:"scrapedAt"`
	PostingDateParsed string   `json:"postingDateParsed"`
	Description       string   `json:"description"`
	DescriptionHTML   string   `json:"descriptionHTML,omitempty"`
	SearchInput       struct {
		Position string `json:"position"`
		Country  string `json:"country"`
	} `json:"searchInput"`
	IsExpired   bool `json:"isExpired"`
	CompanyInfo struct {
		IndeedURL          string   `json:"indeedUrl"`
		URL                *string  `json:"url"`
		CompanyDescription *string  `json:"companyDescription"`
		Rating             *float64 `json:"rating"`
		ReviewCount        *int     `json:"reviewCount"`
		CompanyLogo        *string  `json:"companyLogo"`
	} `json:"companyInfo"`
}

// ApifyLinkedInResponse represents the response from the LinkedIn API via Apify
type ApifyLinkedInResponse []struct {
	ID                 string   `json:"id"`
	TrackingID         string   `json:"trackingId"`
	RefID              string   `json:"refId"`
	Link               string   `json:"link"`
	Title              string   `json:"title"`
	CompanyName        string   `json:"companyName"`
	CompanyLinkedinUrl string   `json:"companyLinkedinUrl"`
	CompanyLogo        string   `json:"companyLogo"`
	Location           string   `json:"location"`
	SalaryInfo         []string `json:"salaryInfo"`
	PostedAt           string   `json:"postedAt"`
	Benefits           []string `json:"benefits"`
	DescriptionHtml    string   `json:"descriptionHtml"`
	ApplicantsCount    string   `json:"applicantsCount"`
	ApplyUrl           string   `json:"applyUrl"`
	DescriptionText    string   `json:"descriptionText"`
	SeniorityLevel     string   `json:"seniorityLevel"`
	EmploymentType     string   `json:"employmentType"`
	JobFunction        string   `json:"jobFunction"`
	Industries         string   `json:"industries"`
	InputUrl           string   `json:"inputUrl"`
	CompanyDescription string   `json:"companyDescription,omitempty"`
	CompanyAddress     struct {
		Type            string `json:"type"`
		StreetAddress   string `json:"streetAddress"`
		AddressLocality string `json:"addressLocality"`
		AddressRegion   string `json:"addressRegion"`
		PostalCode      string `json:"postalCode"`
		AddressCountry  string `json:"addressCountry"`
	} `json:"companyAddress,omitempty"`
	CompanyWebsite        string `json:"companyWebsite,omitempty"`
	CompanySlogan         string `json:"companySlogan,omitempty"`
	CompanyEmployeesCount int    `json:"companyEmployeesCount,omitempty"`
}
