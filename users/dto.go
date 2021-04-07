package users

// User ...
type User struct {
	ID       string `json:"ID"`
	Name     string `json:"Name"`
	Username string `json:"Username"`
	Email    string `json:"Link"`
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
