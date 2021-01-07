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
