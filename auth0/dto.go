package auth0

import "time"

// AuthToken Struct
type AuthToken struct {
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

// Resp struct
type Resp struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// ESTokenSchema for AuthToken
type ESTokenSchema struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string  `json:"_index"`
			Type   string  `json:"_type"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				Name  string `json:"name"`
				Token string `json:"token"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

// LastActionSchema ...
type LastActionSchema struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index  string  `json:"_index"`
			Type   string  `json:"_type"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				Date time.Time `json:"date"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

const (
	lastTokenDate         = "last-token-date"
	lastAuth0TokenRequest = "last-auth0-token-request-"
	auth0TokenCache       = "auth0-token-cache-"
	tokenDoc              = "token"
)

// RefreshResult ...
type RefreshResult string

const(
	// RefreshError ...
	RefreshError RefreshResult = "error refreshing auth0 token"
	// RefreshSuccessful ...
	RefreshSuccessful RefreshResult = "token refreshed successfully"
	// NotExpireSoon ...
	NotExpireSoon RefreshResult = "token will not expire soon"
)
