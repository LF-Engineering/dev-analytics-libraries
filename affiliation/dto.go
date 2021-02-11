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
	Code         int         `json:"Code"`
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

// EnrollmentsResponse ...
type EnrollmentsResponse struct {
	Enrollments []Enrollment `json:"enrollments"`
	Scope       string       `json:"scope"`
	User        string       `json:"user"`
	UUID        string       `json:"uuid"`
}

// ProfileResponse ...
type ProfileResponse struct {
	Enrollments  []Enrollment `json:"enrollments"`
	Identities   []Identity   `json:"identities"`
	LastModified time.Time    `json:"last_modified"`
	Profile      Profile      `json:"profile"`
	UUID         string       `json:"uuid"`
}

// IdentityData ...
type IdentityData struct {
	Email        *string    `json:"email,omitempty"`
	ID           string     `json:"id,omitempty"`
	LastModified *time.Time `json:"last_modified,omitempty"`
	Name         *string    `json:"name,omitempty"`
	Source       string     `json:"source,omitempty"`
	Username     *string    `json:"username,omitempty"`
	UUID         *string    `json:"uuid,omitempty"`
}

// UniqueIdentityFullProfile ...
type UniqueIdentityFullProfile struct {
	Enrollments []*Enrollments  `json:"enrollments"`
	Identities  []*IdentityData `json:"identities"`
	Profile     *Profile        `json:"profile,omitempty"`
	UUID        string          `json:"uuid,omitempty"`
}

// Enrollments ...
type Enrollments struct {
	Organization *Organization `json:"organization,omitempty"`
	End          time.Time     `json:"end"`
	ID           int           `json:"id"`
	Start        time.Time     `json:"start"`
}

// Organization ...
type Organization struct {
	Name string `json:"name,omitempty"`
}

// Profile ...
type Profile struct {
	Email     *string `json:"email,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	GenderAcc *int64  `json:"gender_acc,omitempty"`
	IsBot     *int64  `json:"is_bot,omitempty"`
	Name      *string `json:"name,omitempty"`
	UUID      string  `json:"uuid,omitempty"`
}

// AffIdentity contains affiliation user Identity
type AffIdentity struct {
	ID            *string `json:"id"`
	UUID          *string
	Name          string
	Username      string
	Email         string
	Domain        string
	Gender        *string  `json:"gender"`
	GenderACC     *int64   `json:"gender_acc"`
	OrgName       *string  `json:"org_name"`
	IsBot         *int64   `json:"is_bot"`
	MultiOrgNames []string `json:"multi_org_names"`
}
