package users

// EmailData ...
type EmailData struct {
	ID           bool   `json:"ID"`
	Active       bool   `json:"Active"`
	EmailAddress string `json:"EmailAddress"`
	IsDeleted    bool   `json:"IsDeleted"`
	IsPrimary    bool   `json:"IsPrimary"`
	IsVerified   bool   `json:"IsVerified"`
}

// User ...
type User struct {
	ID       string      `json:"ID"`
	Name     string      `json:"Name"`
	Username string      `json:"Username"`
	Emails   []EmailData `json:"Emails"`
	Email    string      `json:"-"` // primary email
}

// ListUsersResponse ...
type ListUsersResponse struct {
	Data     []User `json:"Data"`
	Metadata struct {
		Offset    int `json:"Offset"`
		PageSize  int `json:"PageSize"`
		TotalSize int `json:"TotalSize"`
	} `json:"Metadata"`
}
