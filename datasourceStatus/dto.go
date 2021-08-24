package datasourcecache

import "time"

// Status ...
type Status struct {
	ProjectSlug         string      `json:"project_slug"`
	Datasource          string      `json:"datasource"`
	Endpoint            string      `json:"endpoint"`
	UpdatedAt           time.Time   `json:"updated_at"`
	CreatedAt           time.Time   `json:"created_at"`
	ErrorMessage        string      `json:"error_message"`
	Status              int         `json:"status"`
	LastSuccessfulEvent interface{} `json:"last_successful_event"`
}


// TopHits result
type TopHits struct {
	Hits Hits `json:"hits"`
}

// Hits result
type Hits struct {
	Hits []NestedHits `json:"hits"`
}

// NestedHits is the actual hit data
type NestedHits struct {
	ID     string `json:"_id"`
	Source Status `json:"_source"`
}
