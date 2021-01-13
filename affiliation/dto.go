package affiliation

import "time"

// Identity ...
type Identity struct {
	ID           string    `json:"id"`
	LastModified time.Time `json:"last_modified"`
	Name         string    `json:"name"`
	Source       string    `json:"source"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	UUID         string    `json:"uuid"`
}

// AffiliationsResponse ...
type AffiliationsResponse struct {
	Code         string      `json:"Code"`
	Message      string      `json:"Message"`
	Enrollments  interface{} `json:"enrollments"`
	Identities   interface{} `json:"identities"`
	LastModified time.Time   `json:"last_modified"`
	Profile      struct {
		UUID string `json:"uuid"`
	} `json:"profile"`
	UUID string `json:"uuid"`
}

// Enrollment ...
type Enrollment struct {
	End          time.Time `json:"end"`
	ID           int       `json:"id"`
	Organization struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"organization"`
	OrganizationID int       `json:"organization_id"`
	Role           string    `json:"role"`
	Start          time.Time `json:"start"`
	UUID           string    `json:"uuid"`
}

// EnrollmentsResponse
type EnrollmentsResponse struct {
	Enrollments []Enrollment `json:"enrollments"`
	Scope string `json:"scope"`
	User  string `json:"user"`
	UUID  string `json:"uuid"`
}