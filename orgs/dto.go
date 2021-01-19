package orgs

// Organization ...
type Organization struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	Link    string `json:"Link"`
	LogoURL string `json:"LogoURL"`
}

// SearchOrganizationResponse ...
type SearchOrganizationResponse struct {
	Data     []Organization `json:"Data"`
	Metadata struct {
		Offset    int `json:"Offset"`
		PageSize  int `json:"PageSize"`
		TotalSize int `json:"TotalSize"`
	} `json:"Metadata"`
}
